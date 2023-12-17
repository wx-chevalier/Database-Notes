# Redis 集群模式

## 一、集群模式介绍

Redis Cluster 是 Redis 官方提供的分布式实现，在 Redis 3.0 版本正式推出，通过集群模式可以扩展单机的性能瓶颈，同时也可以通过横向扩展来实现扩容。此外，Redis 集群模式还提供了副本迁移机制，用于保证数据的安全和提高集群的容错能力，从而实现高可用。

### 1.1 数据分区

Redis Cluster 采用虚拟槽进行分区，槽是集群内数据管理和迁移的基本单位。所有的键根据哈希函数映射到 16384 个整数槽内，每个节点负责维护一部分槽及槽上的数据，计算公式如下：

```shell
HASH_SLOT = CRC16(key) mod 16384
```

假设现在有一个 6 个节点的集群，分别有 3 个 Master 点和 3 个 Slave 节点，槽会尽量均匀的分布在所有 Master 节点上。数据经过散列后存储在指定的 Master 节点上，之后 Slave 节点会进行对应的复制操作。这里再次说明一下槽只是一个虚拟的概念，并不是数据存放的实际载体。

<div align="center"> <img src="https://gitee.com/heibaiying/Full-Stack-Notes/raw/master/pictures/redis-集群架构.png"/> </div>

### 1.2 节点通讯

在 Redis 分布式架构中，每个节点都存储有整个集群所有节点的元数据信息，这是通过 P2P 的 Gossip 协议来实现的。集群中的每个节点都会单独开辟一个 TCP 通道，用于节点之间彼此通信，通信端口号在基础端口上加 10000；每个节点定期通过特定的规则选择部分节点发送 ping 消息，接收到 ping 信息的节点用 pong 消息作为响应，通过一段时间的彼此通信，最终所有节点都会达到一致的状态，每个节点都会知道整个集群全部节点的状态信息，从而到达集群状态同步的目的。

<div align="center"> <img src="https://gitee.com/heibaiying/Full-Stack-Notes/raw/master/pictures/redis节点通讯.png"/> </div>

### 1.3 请求路由

#### 1. 请求重定向

在集群模式下，Redis 接收到命令时会先计算键对应的槽，然后根据槽找出对应的目标节点，如果目标节点就是此时所在的节点，则直接进行处理，否则返回 MOVED 重定向消息给客户端，通知客户端去正确的节点上执行操作。

#### 2. Smart 客户端

Redis 的大多数客户端都是 Smart 客户端，Smart 客户端会在内部缓存槽与节点之间的映射关系，从而在本机就可以查找到正确的节点，这样可以保证 IO 效率的最大化。如果客户端还接收到 MOVED 重定向的消息，则代表客户端内部的缓存已经失效，此时客户端会去重新获取映射关系然后刷新本地缓存。

#### 3. ASK 重定向

当集群处于扩容阶段时，此时槽上的数据可能正在从源节点迁移到目标节点，在这个期间可能出现一部分数据在源节点，而另一部分在目标节点情况。此时如果源节点接收到命令并判断出键对象不存在，说明其可能存在于目标节点上，这时会返回给客户端 ASK 重定向异常。

ASK 重定向与 MOVED 重定向的区别在于：收到 ASK 重定向时说明集群正在进行数据迁移，客户端无法知道什么时候迁移完成，因此只是临时性的重定向，客户端不会更新映射缓存。但是 MOVED 重定向说明键对应的槽已经明确迁移到新的节点，因此需要更新映射缓存。

### 1.4 故障发现

由于 Redis 集群的节点间都保持着定时通讯，某个节点向另外一个节点发送 ping 消息，如果正常接受到 pong 消息，此时会更新与该节点最后一次的通讯时间记录，如果之后无法正常接受到 pong 消息，并且判断当前时间与最后一次通讯的时间超过 `cluster-node-timeout` ，此时会对该节点做出主观下线的判断。

当做出主观下线判断后，节点会把这个判断在集群内传播，通过 Gossip 消息传播，集群内节点不断收集到故障节点的下线报告。当半数以上持有槽的主节点都标记某个节点是主观下线时，触发客观下线流程。这里需要注意的是只有持有槽主节点才有权利做出主观下线的判断，因为集群模式下只有处理槽的主节点才负责读写请求和维护槽等关键信息，而从节点只进行主节点数据和状态信息的复制。

### 1.5 故障恢复

#### 1. 资格检查

每个从节点都要检查最后与主节点断线时间，判断是否有资格替换故障的主节点。如果从节点与主节点断线时间超过 `cluster-node-time*cluster-slave-validity-factor`，则当前从节点不具备故障转移资格。这两个参数可以在 `redis.conf` 中进行修改，默认值分别为 15000 和 10。

#### 2. 准备选举

当从节点符合故障转移资格后，更新触发故障选举的时间，只有到达该时间后才能执行后续流程。在这一过程中，Redis 会比较每个符合资格的从节点的复制偏移量，然后让复制偏移量大（即数据更加完整）的节点优先发起选举。

#### 3. 选举投票

从节点每次发起投票时都会自增集群的全局配置纪元，全局配置纪元是一个只增不减的整数。之后会在集群内广播选举消息，只有持有槽的主节点才会处理故障选举消息，并且每个持有槽的主节点在一个配置纪元内只有唯一的一张选票。假设集群内有 N 个持有槽的主节点，当某个从节点获得 N/2+1 张选票则代表选举成功。如果在开始投票之后的 `cluster-node-timeout*2` 时间内没有从节点获取足够数量的投票，则本次选举作废，从节点会对配置纪元自增并发起下一轮投票，直到选举成功为止。

#### 4. 替换主节点

当从节点收集到足够的选票之后，就会触发替换主节点操作：

- 当前从节点取消复制变为主节点。
- 执行 clusterDelSlot 操作撤销原主节点负责的槽，并执行 clusterAddSlot 把这些槽委派给自己。
- 向集群广播自己的 pong 消息，通知集群内的其他节点自己已经成为新的主节点。

## 二、集群模式搭建

### 2.1 节点配置

拷贝 6 份 `redis.conf`，分别命名为 `redis-6479.conf` ~ `redis-6484.conf`，需要修改的配置项如下：

```shell
# redis-6479.conf
port 6479
# 以守护进程的方式启动
daemonize yes
# 当Redis以守护进程方式运行时，Redis会把pid写入该文件
pidfile /var/run/redis_6479.pid
logfile 6479.log
dbfilename dump-6479.rdb
dir /home/redis/data/
# 开启集群模式
cluster-enabled yes
# 节点超时时间，单位毫秒
cluster-node-timeout 15000
# 集群内部配置文件
cluster-config-file nodes-6479.conf


# redis-6480.conf
port 6480
daemonize yes
pidfile /var/run/redis_6480.pid
logfile 6480.log
dbfilename dump-6480.rdb
dir /home/redis/data/
cluster-enabled yes
cluster-node-timeout 15000
cluster-config-file nodes-6480.conf

..... 其他配置类似，修改所有用到端口号的地方
```

### 2.2 启动集群

启动所有 Redis 节点，启动后使用 `ps -ef | grep redis` 查看进程，输出应如下：

<div align="center"> <img src="https://gitee.com/heibaiying/Full-Stack-Notes/raw/master/pictures/redis-cluster-ps-ef.png"/> </div>
接着需要使用以下命令创建集群，集群节点之间会开始进行通讯，并完成槽的分配：

```shell
redis-cli --cluster create 127.0.0.1:6479 127.0.0.1:6480 127.0.0.1:6481 \
127.0.0.1:6482 127.0.0.1:6483  127.0.0.1:6484 --cluster-replicas 1
```

执行后输出如下：M 开头的表示持有槽的主节点，S 开头的表示从节点，每个节点都有一个唯一的 ID。最后一句输出表示所有的槽都已经分配到主节点上，此时代表集群搭建成功。

<div align="center"> <img src="https://gitee.com/heibaiying/Full-Stack-Notes/raw/master/pictures/redis-cluster-create.png"/> </div>

### 2.3 集群完整性校验

集群完整性指所有的槽都分配到存活的主节点上，只要 16384 个槽中有一个没有分配给节点则表示集群不完整。可以使用以下命令进行检测，check 命令只需要给出集群中任意一个节点的地址就可以完成整个集群的检查工作：

```shell
redis-cli --cluster check 127.0.0.1:6479
```

### 2.4 关于版本差异的说明

如果你使用的是 Redis 5，可以和上面的示例一样，直接使用嵌入到 redis-cli 中的 Redis Cluster 命令来创建和管理集群。

如果你使用的是 Redis 3 或 4，则需要使用 redis-trib.rb 工具，此时需要预先准备 Ruby 环境和安装 redis gem。

## 三、集群伸缩

Redis 集群提供了灵活的节点扩容和缩容方案，可以在不影响集群对外服务的情况下，进行动态伸缩。

### 3.1 集群扩容

这里准备两个新的节点 6485 和 6486，配置和其他节点一致，配置完成后进行启动。集群扩容的命令为 `add-node`，第一个参数为需要加入的新节点，第二个参数为集群中任意节点，用于发现集群：

```shell
redis-cli --cluster add-node 127.0.0.1:6485 127.0.0.1:6479
```

成功加入集群后，可以使用 `cluster nodes` 命令查看集群情况。不做任何特殊指定，默认加入集群的节点都是主节点，但是集群并不会为分配任何槽。如下图所示，其他 master 节点后面都有对应的槽的位置信息，但新加入的 6485 节点则没有，由于没有负责的槽，所以该节点此时不能进行任何读写操作：

```shell
redis-cli -h 127.0.0.1 -p 6479 cluster nodes
```

<div align="center"> <img src="https://gitee.com/heibaiying/Full-Stack-Notes/raw/master/pictures/redis-cluster-nodes.png"/> </div>

想要让新加入的节点能够进行读写操作，可以使用 `reshard` 命令为其分配槽，这里我们将其他三个主节点上的槽迁移一部分到 6485 节点上，这里一共迁移 4096 个槽，即 16384 除以 4 。`cluster-from` 用于指明槽的源节点，可以为多个，`cluster-to` 为槽的目标节点，`cluster-slots` 为需要迁移的槽的总数：

```shell
redis-cli --cluster reshard 127.0.0.1:6479 \
--cluster-from fd35b17ace0f15314ed3b3d4f8ff4da08e11b89d,ebd0425db25b8bcf843fee9826755848e23a895a,98a175734db4a106ae676dc403f39b2783640789 \
--cluster-to 819f867afd1da1acfb1a528d3efa91cffb02ba97 \
--cluster-slots 4096 --cluster-yes
```

迁移后，再次使用 `cluster nodes` 命令可以查看到此时 6485 上已经有其他三个主节点上迁移过来的槽：

<div align="center"> <img src="https://gitee.com/heibaiying/Full-Stack-Notes/raw/master/pictures/redis-cluster-nodes2.png"/> </div>

为保证高可用，可以为新加入的主节点添加从节点，命令如下。`add-node` 接收两个参数，第一个为需要添加的从节点，第二个参数为集群内任意节点，用于发现集群。`cluster-master-id` 参数用于指明作为哪个主节点的从节点，如果不加这个参数，则自动分配给从节点较少的主节点：

```shell
redis-cli --cluster add-node 127.0.0.1:6486 127.0.0.1:6479 --cluster-slave \
--cluster-master-id 819f867afd1da1acfb1a528d3efa91cffb02ba97
```

### 3.2 集群缩容

集群缩容的命令如下：第一个参数为集群内任意节点，用于发现集群；第二个参数为需要删除的节点：

```
redis-cli --cluster del-node 127.0.0.1:6479 `<node-id>`
```

需要注意的是待删除的主节点上必须为空，如果不为空则需要将它上面的槽和数据迁移到其他节点上，和扩容时一样，可以使用 `reshard` 命令来完成数据迁移。

## 参考资料

1. 付磊，张益军 . 《Redis 开发与运维》. 机械工业出版社 . 2017-3-1
2. 官方文档：[Redis cluster tutorial](https://redis.io/topics/cluster-tutorial)
