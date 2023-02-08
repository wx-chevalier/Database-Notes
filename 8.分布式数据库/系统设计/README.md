# 系统设计

分布式数据库我们关注：

- 数据如何在机器上分布；
- 数据副本如何保持一致性；
- 如何支持 SQL；
- 分布式事务如何实现；

> 推荐阅读《[DistributedSystem-Series](https://github.com/wx-chevalier/DistributedSystem-Series?q=)》中相关内容。

# 数据分布

NewSQL 和 NoSQL 的数据分布是类似的，他们都认为所有数据不适合存放在一台机器上，必须分片存储。因此需要考虑：

- 如何划分分片？
- 如何定位特定的数据？

分片主要有两种方法：哈希或范围：

- 哈希分片：将某个关键字通过哈希函数计算得到一个哈希值，根据哈希值来判断数据应该存储的位置。这样做的优点是易于定位数据，只需要运行一下哈希函数就能够知道数据存储在哪台机器；但缺点也十分明显，由于哈希函数是随机的，数据将无法支持范围查询。
- 范围分片：指按照某个范围划分数据存储的位置，举个最简单的例子，按照首字母从 A-Z 分为 26 个分区，这样的分片方式对于范围查询非常有用；缺点是通常需要对关键字进行查询才知道数据处于哪个节点，这看起来会造成一些性能损耗，但由于范围很少会改变，很容易将范围信息缓存起来。

例如下图所示，我们按照关键字划分为三个范围：[a 开头，h 开头)、[h 开头，p 开头)、[p 开头，无穷）。

![首字母范围划分](https://ngte-superbed.oss-cn-beijing.aliyuncs.com/superbed/2021/08/09/6110b97f5132923bf87e5c68.jpg)

如下图所示，这样进行范围查询效率会更高：

![更高效查询](https://ngte-superbed.oss-cn-beijing.aliyuncs.com/superbed/2021/08/09/6110b99e5132923bf87e9a84.jpg)

我们关心的最后一个问题是，当某个分片的数据过大，超过我们所设的阈值时，如何扩展分片？由于有一个中间层进行转换，这也很容易进行，只需要在现有的范围中选取某个点，然后将该范围一分为二，便得到两个分区。如下图所示，当 p-z 的数据量超过阈值，为了避免负载压力，我们拆分该范围。

![存储拆分](https://pic.imgdb.cn/item/6110b9ef5132923bf87f3fa4.jpg)

显然，这里有一个取舍(trade-off)，如果范围阈值设置得很大，那么在机器之间移动数据会很慢，也很难快速恢复某个故障机器的数据；但如果范围阈值设置得很小，中间转换层可能会增长得非常快，增加查询的开销，同时数据也会频繁拆分。一般范围阈值选择 64 MB 到 128 MB，Cockroachdb 使用 64MB 大小，TiDB 默认阈值为 96 MB 大小。

# 数据一致性

一个带有`分布式`三个字的系统当然需要容忍错误，为了避免一台机器挂掉后数据彻底丢失，通常会将数据复制到多台机器上冗余存储。但分布式系统中请求会丢失、机器会宕机、网络会延迟，因此我们需要某种方式知道冗余的副本中哪些数据是最新的。

最常见的复制数据方式是主从同步（或者直接复制冷备数据），主节点将更新操作同步到从节点。但这样存在潜在的数据不一致问题，同步更新操作丢失了怎么办？从节点恰好写入失败了怎么办？有时这些错误甚至会永久损坏数据，需要数据库管理员介入。保持一致性常常会以性能为代价(以后我们会讨论)，因此，大部分 NoSQL 只保证最终一致性，并通过一些冲突处理方案来解决数据不一致。

现有著名的复制数据的算法是我们经常听到的 Paxos、Raft、Zab 或 Viewstamped Replication 等算法。其中 Raft 诞生后便席卷了分布式共识算法领域，它只需要超过半数的节点写入成功，即认为本次写操作成功，并返回结果给客户端。发生故障时，Raft 算法可以重新选举领导者，只要少于半数的节点发生故障，Raft 就能正常工作。

Raft 算法可以满足可靠复制数据，同时系统能够容忍不超过半数的节点故障。在分布式数据库中，一个分片使用一个共识组(consensus group)复制数据，具体的 Raft 共识组称为 Raft 组(Raft group)，Paxos 共识组称为 Paxos 组(Paxos group)。从 TiDB 官网中找来一张图，TiDB 将一个分片称为一个 Region，如图中有三个 Raft 组，用来复制三个 Region 的数据。

![TiDB Scale-out](https://ngte-superbed.oss-cn-beijing.aliyuncs.com/superbed/2021/08/09/6110cd745132923bf8aa9f79.jpg)

# SQL 表数据 KV 化存储

解决了 KV 存储以后，我们还要想办法用 KV 结构来存储表结构。通常，增删查改可以抽象成如下 5 个 KV 操作（也许可以再多些，但基本就是这些）。

```sh
Get(key)
Put(key, value)
ConditionalPut(key, value, exp)
Scan(startKey, endKey)
Del(key)
```

我们讨论的是 OLTP 类分布式数据库都是行存。我们以 CockroachDB 举例，一个表通常包含行和列，可以将一个表转换成如下结构：

```sh
/<table>/<index>/<key>/<column> -> Value
```

为了可读性使用斜杠来分割字段。`/<index>/<key>/` 这部分表示需要每个表必须有一个主键。这样看不大直观，举个例子，对于以下建表语句：

```sql
CREATE TABLE test (
    id      INTEGER PRIMARY KEY,
    name    VARCHAR,
    price   FLOAT,
);
```

转换成 KV 存储如图所示：

![关系型转化为 KV 存储](https://ngte-superbed.oss-cn-beijing.aliyuncs.com/superbed/2021/08/09/6110cdf05132923bf8abac4e.jpg)

当然，这样的存储方式会将 float 等类型通通转换为 string 类型。除此之外，数据库通常会创建一些非主键索引，主要分为两类：唯一索引、非唯一索引。唯一索引比较简单，由于值唯一，我们可以通过如下映射：

```sh
/<table>/<index>/<key> -> Value
```

![索引键值映射](https://ngte-superbed.oss-cn-beijing.aliyuncs.com/superbed/2021/08/09/6110ce195132923bf8ac1933.jpg)

非唯一索引和主键类似，只不过其值为空。如图所示：

![非唯一索引](https://ngte-superbed.oss-cn-beijing.aliyuncs.com/superbed/2021/08/09/6110ce4d5132923bf8ac9cb1.jpg)

上述表数据 KV 化规则已经有些陈旧，CockroachDB 最新的映射规则参阅《Structured data encoding in CockroachDB SQL》。但其中的思想是相似的。当然，表数据 KV 化并不只有这种方式，TiDB 则按照如下规则进行映射：

![TiDB 映射规则](https://ngte-superbed.oss-cn-beijing.aliyuncs.com/superbed/2021/08/09/6110cfa45132923bf8af755f.jpg)

# 分布式事务

当我们谈论事务时，永远离不开 ACID。分布式事务中最难保证的是原子性和隔离性。在分布式系统中，原子性需要原子提交协议来实现，例如两阶段提交；而隔离性可以通过两阶段锁或多版本并发控制(MVCC)来实现不同的隔离级别。

分布式数据库们都实现了 MVCC，Google Spanner 设计了 TrueTime 来实现，但 TrueTime 并不开源；TiDB 则基于 Google Percolator 来实现。C
