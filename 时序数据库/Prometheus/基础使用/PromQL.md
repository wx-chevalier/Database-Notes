# PromQL

Promethues 的查询语言是 PromQL，语法解析 AST，执行计划和数据聚合是由 PromQL 完成，fanout 模块会向本地和远端同时下发查询数据，PTSDB 负责本地数据的检索。PTSDB 实现了定义的 Adpator，包括 Select, LabelNames, LabelValues 和 Querier.

PromQL 定义了三类查询：

- 瞬时数据 (Instant vector): 包含一组时序，每个时序只有一个点，例如：http_requests_total

- 区间数据 (Range vector): 包含一组时序，每个时序有多个点，例：http_requests_total[5m]

- 纯量数据 (Scalar): 纯量只有一个数字，没有时序，例如：count(http_requests_total)

一些典型的查询包括：

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
