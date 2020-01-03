# PromQL

PromQL 即是 Prometheus Query Language，虽然它以 QL 结尾，却并非 SQL 那样的语言。标签是 PromQL 的关键部分，您不仅可以使用它们进行任意聚合，还可以将不同的指标结合在一起，以针对它们进行算术运算。从预测到日期以及数学功能，您可以使用多种功能。PromQL 定义了三类查询：

- 瞬时数据 (Instant vector): 包含一组时序，每个时序只有一个点，例如：http_requests_total

- 区间数据 (Range vector): 包含一组时序，每个时序有多个点，例：http_requests_total[5m]

- 纯量数据 (Scalar): 纯量只有一个数字，没有时序，例如：count(http_requests_total)

# 语法速览

最简单的 PromQL 就是直接输入指标名称，比如 up 这个指标名称就能够得到所有的实例的运行幸亏：

```sh
up{instance="192.168.0.107:9090",job="prometheus"}    1
up{instance="192.168.0.108:9090",job="prometheus"}    1
up{instance="192.168.0.107:9100",job="server"}    1
up{instance="192.168.0.108:9104",job="mysql"}    0

# 指定某个 label
up{job="prometheus"}
```

这种写法被称为 Instant vector selectors，这里不仅可以使用 = 号，还可以使用 !=、=~、!~，比如下面这样：

```sh
up{job!="prometheus"}

# =~ 是根据正则表达式来匹配，必须符合 RE2 的语法。
up{job=~"server|mysql"}
up{job=~"192\.168\.0\.107.+"}
```

## 典型查询

这里我们列举了一些常见的 PromQL 查询，及其与 SQL 的对比：

```sh
# 查询当前所有数据
http_requests_total
select * from http_requests_total where timestamp between xxxx and xxxx

# 条件查询
http_requests_total{code="200", handler="query"}
select * from http_requests_total where code="200" and handler="query" and timestamp between xxxx and xxxx

# 模糊查询: code 为 2xx 的数据
http_requests_total{code~="20"}
select * from http_requests_total where code like "%20%" and timestamp between xxxx and xxxx

# 值过滤： value大于100
http_requests_total > 100
select * from http_requests_total where value > 100 and timestamp between xxxx and xxxx

# 范围区间查询: 过去 5 分钟数据
http_requests_total[5m]
select * from http_requests_total where timestamp between xxxx-5m and xxxx

# count 查询: 统计当前记录总数
count(http_requests_total)
select count(*) from http_requests_total where timestamp between xxxx and xxxx

# sum 查询：统计当前数据总值
sum(http_requests_total)
select sum(value) from http_requests_total where timestamp between xxxx and xxxx

# top 查询: 查询最靠前的 3 个值
topk(3, http_requests_total)
select * from http_requests_total where timestamp between xxxx and xxxx order by value desc limit 3

# irate查询：速率查询
irate(http_requests_total[5m])
select code, handler, instance, job, method, sum(value)/300 AS value from http_requests_total where timestamp between xxxx and xxxx group by code, handler, instance, job, method;
```

# Gauge

Gauge 是对于状态的快照，通常被用于那些需要求和、求平均以及最小值、最大值的场景。譬如 Node Exporter 导出了 node_filesystem_size_bytes 这个度量，其反映了挂载的文件系统的体积，并且包含了 device, fstype, 以及 mountpoint 标签。我们能够通过如下的方式计算总的文件大小：

```sh
sum without(device, fstype, mountpoint)(node_filesystem_size_bytes)
```

这里的 without 函数会高速 sum 聚合器将有相同 label 的度量值相加，但是忽略 device, fstype, 以及 mountpoint 标签。我们可能会得到如下的结果：

```sh
node_filesystem_free_bytes{device="/dev/sda1",fstype="vfat",
    instance="localhost:9100",job="node",mountpoint="/boot/efi"} 70300672
node_filesystem_free_bytes{device="/dev/sda5",fstype="ext4",
    instance="localhost:9100",job="node",mountpoint="/"} 30791843840
node_filesystem_free_bytes{device="tmpfs",fstype="tmpfs",
    instance="localhost:9100",job="node",mountpoint="/run"} 817094656
node_filesystem_free_bytes{device="tmpfs",fstype="tmpfs",
    instance="localhost:9100",job="node",mountpoint="/run/lock"} 5238784
node_filesystem_free_bytes{device="tmpfs",fstype="tmpfs",
    instance="localhost:9100",job="node",mountpoint="/run/user/1000"} 826912768

# 聚合的结果如下
{instance="localhost:9100",job="node"} 32511390720
```

同样地，我们也可以忽略 instance 标签，即获得总的文件系统的大小：

```sh
sum without(device, fstype, mountpoint, instance)(node_filesystem_size_bytes)
{job="node"} 32511390720
```

您可以对其他聚合使用相同的方法。 max 会告诉您每台计算机上最大的已挂载文件系统的大小：

```sh
max without(device, fstype, mountpoint)(node_filesystem_size_bytes)

{instance="localhost:9100",job="node"} 30792601600

avg without(instance, job)(process_open_fds)
```

# Counter

Counters 跟踪事件的数量或大小，并且您的应用程序在其 metrics 上公开的值是自启动以来的总和。但是，这对于您自己而言并没有多大用处，您真正想知道的是计数器随着时间增长的速度。尽管 increase 和 irate 函数也可以对计数器值进行操作，但这通常使用 rate 函数来完成。

譬如，如果我们要去计算每秒吞吐的网络流量：

```sh
rate(node_network_receive_bytes_total[5m])
```

`[5m]` 意为使用过去五分钟的数据进行计算，即会得到过去五分钟的平均值：

```sh
{device="lo",instance="localhost:9100",job="node"}  1859.389655172414
{device="wlan0",instance="localhost:9100",job="node"} 1314.5034482758622
```

rate 函数的输出值即为 Gauge，我们也就可以应用 Gauge 的聚合函数。如上的用法，我们也可以忽略网卡详情而获得总的容量：

```sh
sum without(device)(rate(node_network_receive_bytes_total[5m]))

# {instance="localhost:9100",job="node"} 3173.8931034482762
```

或者我们也可以指定获取某个网卡的数据：

```sh
sum without(instance)(rate(node_network_receive_bytes_total{device="eth0"}[5m]))

# {device="eth0",job="node"} 3173.8931034482762
```
