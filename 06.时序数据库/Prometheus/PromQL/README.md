# PromQL

Prometheus 通过指标名称（metrics name）以及对应的一组标签（label set）唯一定义一条时间序列。指标名称反映了监控样本的基本标识，而 label 则在这个基本特征上为采集到的数据提供了多种特征维度。用户可以基于这些特征维度过滤，聚合，统计从而产生新的计算后的一条时间序列。

PromQL 即是 Prometheus Query LanGauge，虽然它以 QL 结尾，却并非 SQL 那样的语言。其提供对时间序列数据丰富的查询，聚合以及逻辑运算能力的支持。并且被广泛应用在 Prometheus 的日常应用当中，包括对数据查询、可视化、告警处理当中。PromQL 定义了三类查询：

- 瞬时数据 (Instant vector): 包含一组时序，每个时序只有一个点，例如：http_requests_total

- 区间数据 (Range vector): 包含一组时序，每个时序有多个点，例：http_requests_total[5m]

- 纯量数据 (Scalar): 纯量只有一个数字，没有时序，例如：count(http_requests_total)

依赖于使用场景（例如画图展示或者输出表达式结果），根据用户所写的表达式，只有部分类型是合法的返回结果。例如：瞬时向量类型是唯一可以直接在图表中使用的。

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

# 值过滤：value大于100
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
