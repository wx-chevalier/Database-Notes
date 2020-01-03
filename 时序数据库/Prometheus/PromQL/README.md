# PromQL

PromQL 即是 Prometheus Query Language，虽然它以 QL 结尾，却并非 SQL 那样的语言。标签是 PromQL 的关键部分，您不仅可以使用它们进行任意聚合，还可以将不同的指标结合在一起，以针对它们进行算术运算。从预测到日期以及数学功能，您可以使用多种功能。PromQL 定义了三类查询：

- 瞬时数据 (Instant vector): 包含一组时序，每个时序只有一个点，例如：http_requests_total

- 区间数据 (Range vector): 包含一组时序，每个时序有多个点，例：http_requests_total[5m]

- 纯量数据 (Scalar): 纯量只有一个数字，没有时序，例如：count(http_requests_total)

## 语法速览

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
