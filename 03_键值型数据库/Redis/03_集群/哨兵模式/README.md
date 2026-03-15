# Redis Sentinel

## 一、复制

为了解决单点问题，保证数据的安全性，Redis 提供了复制机制，用于满足故障恢复和负载均衡等需求。通过复制机制，Redis 可以通过多个副本保证数据的安全性，从而提供高可用的基础，Redis 的哨兵和集群模式都是在复制基础上实现高可用的。

### 1.1 建立复制关系

想要对两个 Redis 节点建立主从复制关系，可以通过以下三种方式来实现：

- 在从节点的配置文件中配置 `slaveof {masterHost} {masterPort}` 选项；
- 在从节点启动时候加入 `--slaveof {masterHost} {masterPort}` 参数；
- 直接在从节点上执行 `slaveof {masterHost} {masterPort}` 命令。

<div align="center"> <img src="https://gitee.com/heibaiying/Full-Stack-Notes/raw/master/pictures/redis主从复制.png"/> </div>

主从节点复制关系建立后，可以使用 `info replication` 命令查看相关的复制状态。需要注意的是每个从节点只能有一个主节点，但主节点可以同时拥有多个从节点，复制行为是单向的，只能由主节点复制到从节点，因此从节点默认都是只读模式，即 `slave-read-only` 的值默认为 `yes`。

### 1.2 断开复制关系

在启动后，如果想要断开复制关系，可以通过在从节点上执行 `slaveof no one` 命令，此时从节点会断开与主节点的复制关系，但并不会删除原有的复制成功的数据，只是无法再获取主节点上的数据。

通过 slaveof 命令还可以实现切换主节点的操作，命令如下：

```shell
slaveof {newMasterIp} {newMasterPort}
```

需要注意的是，当你从一个主节点切换到另外一个主节点时，该从节点上的原有的数据会被完全清除，然后再执行复制操作，从而保证该从节点上的数据和新主节点上的数据相同。

### 1.3 复制机制缺陷

复制机制最主要的缺陷在于，一旦主节点出现故障，从节点无法自动晋升为主节点，需要通过手动切换来实现，这样无法达到故障的快速转移，因此也不能实现高可用。基于这个原因，就产生了哨兵模式。

## 二、哨兵模式原理

哨兵模式的主要作用在于它能够自动完成故障发现和故障转移，并通知客户端，从而实现高可用。哨兵模式通常由一组 Sentinel 节点和一组（或多组）主从复制节点组成，架构如下：

<div align="center"> <img src="https://gitee.com/heibaiying/Full-Stack-Notes/raw/master/pictures/redis哨兵模式.png"/> </div>

### 2.1 架构说明

#### 1. Sentinel 与 Redis Node

Redis Sentinel 是一个特殊的 Redis 节点。在哨兵模式创建时，需要通过配置指定 Sentinel 与 Redis Master Node 之间的关系，然后 Sentinel 会从主节点上获取所有从节点的信息，之后 Sentinel 会定时向主节点和从节点发送 `info` 命令获取其拓扑结构和状态信息。

#### 2. Sentinel 与 Sentinel

基于 Redis 的发布订阅功能，每个 Sentinel 节点会向主节点的 `__sentinel__：hello` 频道上发送该 Sentinel 节点对于主节点的判断以及当前 Sentinel 节点的信息 ，同时每个 Sentinel 节点也会订阅该频道，来获取其他 Sentinel 节点的信息以及它们对主节点的判断。

#### 3. 心跳机制

通过以上两步所有的 Sentinel 节点以及它们与所有的 Redis 节点之间都已经彼此感知到，之后每个 Sentinel 节点会向主节点、从节点、以及其余 Sentinel 节点定时发送 ping 命令作为心跳检测，来确认这些节点是否可达。

### 2.2 故障转移原理

每个 Sentinel 都会定时进行心跳检查，当发现主节点出现心跳检测超时的情况时，此时认为该主节点已经不可用，这种判定称为主观下线。之后该 Sentinel 节点会通过 `sentinel ismaster-down-by-addr` 命令向其他 Sentinel 节点询问对主节点的判断，当 quorum 个 Sentinel 节点都认为该节点故障时，则执行客观下线，即认为该节点已经不可用。这也同时解释了为什么必须需要一组 Sentinel 节点，因为单个 Sentinel 节点很容易对故障状态做出误判。

> 这里 quorum 的值是我们在哨兵模式搭建时指定的，后文会有说明，通常为 `Sentinel节点总数/2+1`，即半数以上节点做出主观下线判断就可以执行客观下线。

因为故障转移的工作只需要一个 Sentinel 节点来完成，所以 Sentinel 节点之间会再做一次选举工作，基于 Raft 算法选出一个 Sentinel 领导者来进行故障转移的工作。被选举出的 Sentinel 领导者进行故障转移的具体步骤如下：

1. 在从节点列表中选出一个节点作为新的主节点，选择方法如下：
   - 过滤不健康或者不满足要求的节点；
   - 选择 slave-priority（优先级）最高的从节点，如果存在则返回，不存在则继续；
   - 选择复制偏移量最大的从节点 ，如果存在则返回，不存在则继续；
   - 选择 runid 最小的从节点。
2. Sentinel 领导者节点会对选出来的从节点执行 `slaveof no one` 命令让其成为主节点。
3. Sentinel 领导者节点会向剩余的从节点发送命令，让他们从新的主节点上复制数据。
4. Sentinel 领导者会将原来的主节点更新为从节点，并对其进行监控，当其恢复后命令它去复制新的主节点。

## 三、哨兵模式搭建

下面演示在单机上搭建哨兵模式，多机搭建步骤亦同。需要注意的是在实际生产环境中，为了保证高可用，Sentinel 节点需要尽量部署在不同主机上，同时为了保证正常选举，至少需要 3 个 Sentinel 节点。

### 3.1 配置复制集

拷贝三份 `redis.conf`，分别命名为 redis-6379.conf ，redis-6380.conf ，redis-6381.conf ，需要修改的配置项如下：

```shell
# redis-6379.conf
port 6379
daemonize yes   #以守护进程的方式启动
pidfile /var/run/redis_6379.pid  #当Redis以守护进程方式运行时，Redis会把pid写入该文件
logfile 6379.log
dbfilename dump-6379.rdb
dir /home/redis/data/


# redis-6380.conf
port 6380
daemonize yes
pidfile /var/run/redis_6380.pid
logfile 6380.log
dbfilename dump-6380.rdb
dir /home/redis/data/
slaveof 127.0.0.1 6379

# redis-6381.conf
port 6381
daemonize yes
pidfile /var/run/redis_6381.pid
logfile 6381.log
dbfilename dump-6381.rdb
dir /home/redis/data/
slaveof 127.0.0.1 6379
```

### 3.2 配置 Sentinel

拷贝三份 `sentinel.conf` ，分别命名为 sentinel-26379.conf ，sentinel-26380.conf ，sentinel-26381.conf ，配置如下：

```shell
# sentinel-26379.conf
port 26379
daemonize yes
logfile 26379.log
dir /home/redis/data/
sentinel monitor mymaster 127.0.0.1 6379 2
sentinel down-after-milliseconds mymaster 30000
sentinel parallel-syncs mymaster 1
sentinel failover-timeout mymaster 180000

# sentinel-26380.conf
port 26380
daemonize yes
logfile 26380.log
dir /home/redis/data/
sentinel monitor mymaster 127.0.0.1 6379 2
sentinel down-after-milliseconds mymaster 30000
sentinel parallel-syncs mymaster 1
sentinel failover-timeout mymaster 180000

# sentinel-26381.conf
port 26381
daemonize yes
logfile 26381.log
dir /home/redis/data/
sentinel monitor mymaster 127.0.0.1 6379 2
sentinel down-after-milliseconds mymaster 30000
sentinel parallel-syncs mymaster 1
sentinel failover-timeout mymaster 180000
```

### 3.3 启动集群

分别启动三个 Redis 节点，命令如下：

```shell
redis-server redis-6379.conf
redis-server redis-6380.conf
redis-server redis-6381.conf
```

分别启动三个 Sentinel 节点，命令如下：

```shell
redis-sentinel sentinel-26379.conf
redis-sentinel sentinel-26380.conf
redis-sentinel sentinel-26381.conf
```

使用 `ps -ef | grep redis` 命令查看进程，此时输出应该如下：

<div align="center"> <img src="https://gitee.com/heibaiying/Full-Stack-Notes/raw/master/pictures/redis-sentinel-ps-ef.png"/> </div>

可以使用 `info replication` 命令查看 Redis 复制集的状态，此时输出如下。可以看到 6379 节点为 master 节点，并且有两个从节点，分别为 slave0 和 slave1，对应的端口为 6380 和 6381：

<div align="center"> <img src="https://gitee.com/heibaiying/Full-Stack-Notes/raw/master/pictures/redis-info-replication.png"/> </div>

可以使用 `info Sentinel` 命令查看任意 Sentinel 节点的状态，从最后一句输出可以看到 Sentinel 节点已经感知到 6379 的 master 节点，并且也知道它有两个 slaves 节点；同时 Sentinel 节点彼此之间也感知到，共有 3 个 Sentinel 节点：

<div align="center"> <img src="https://gitee.com/heibaiying/Full-Stack-Notes/raw/master/pictures/redis-sentinel-infomation.png"/> </div>

## 参考资料

1. 付磊，张益军 . 《Redis 开发与运维》. 机械工业出版社 . 2017-3-1
2. 官方文档：[Redis Sentinel Documentation](https://redis.io/topics/sentinel)
