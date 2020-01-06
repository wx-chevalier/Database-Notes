# 在 HTTP API 中使用 PromQL

我们不仅仅可以在 Prometheus 的 Graph 页面查询 PromQL，Prometheus 还提供了一种 HTTP API 的方式，可以更灵活的将 PromQL 整合到其他系统中使用，譬如下面要介绍的 Grafana，就是通过 Prometheus 的 HTTP API 来查询指标数据的。实际上，我们在 Prometheus 的 Graph 页面查询也是使用了 HTTP API。

我们看下 [Prometheus 的 HTTP API 官方文档](https://prometheus.io/docs/prometheus/latest/querying/api/)，它提供了下面这些接口：

- GET /api/v1/query
- GET /api/v1/query_range
- GET /api/v1/series
- GET /api/v1/label/<label_name>/values
- GET /api/v1/targets
- GET /api/v1/rules
- GET /api/v1/alerts
- GET /api/v1/targets/metadata
- GET /api/v1/alertmanagers
- GET /api/v1/status/config
- GET /api/v1/status/flags

从 Prometheus v2.1 开始，又新增了几个用于管理 TSDB 的接口：

- POST /api/v1/admin/tsdb/snapshot
- POST /api/v1/admin/tsdb/delete_series
- POST /api/v1/admin/tsdb/clean_tombstones

## API 响应格式

Prometheus API 使用了 JSON 格式的响应内容。当 API 调用成功后将会返回 2xx 的 HTTP 状态码。

反之，当 API 调用失败时可能返回以下几种不同的 HTTP 状态码：

- 404 Bad Request：当参数错误或者缺失时。
- 422 Unprocessable Entity 当表达式无法执行时。
- 503 Service Unavailiable 当请求超时或者被中断时。

所有的 API 请求均使用以下的 JSON 格式：

```
{
  "status": "success" | "error",
  "data": <data>,

  // Only set if status is "error". The data field may still hold
  // additional data.
  "errorType": "<string>",
  "error": "<string>"
}
```

## 在 HTTP API 中使用 PromQL

通过 HTTP API 我们可以分别通过/api/v1/query 和/api/v1/query_range 查询 PromQL 表达式当前或者一定时间范围内的计算结果。

### 瞬时数据查询

通过使用 QUERY API 我们可以查询 PromQL 在特定时间点下的计算结果。

```
GET /api/v1/query
```

URL 请求参数：

- query=<string>：PromQL 表达式。
- time=<rfc3339 | unix_timestamp>：用于指定用于计算 PromQL 的时间戳。可选参数，默认情况下使用当前系统时间。
- timeout=<duration>：超时设置。可选参数，默认情况下使用-query,timeout 的全局设置。

例如使用以下表达式查询表达式 up 在时间点 2015-07-01T20:10:51.781Z 的计算结果：

```json
$ curl 'http://localhost:9090/api/v1/query?query=up'
{
   "status" : "success",
   "data" : {
      "resultType" : "vector",
      "result" : [
         {
            "metric" : {
               "__name__" : "up",
               "job" : "prometheus",
               "instance" : "localhost:9090"
            },
            "value": [ 1435781451.781, "1" ]
         },
         {
            "metric" : {
               "__name__" : "up",
               "job" : "node",
               "instance" : "localhost:9100"
            },
            "value" : [ 1435781451.781, "0" ]
         }
      ]
   }
}
```

### 响应数据类型

当 API 调用成功后，Prometheus 会返回 JSON 格式的响应内容，格式如上小节所示。并且在 data 节点中返回查询结果。data 节点格式如下：

```
{
  "resultType": "matrix" | "vector" | "scalar" | "string",
  "result": <value>
}
```

PromQL 表达式可能返回多种数据类型，在响应内容中使用 resultType 表示当前返回的数据类型，包括：

- 瞬时向量：vector

当返回数据类型 resultType 为 vector 时，result 响应格式如下：

```
[
  {
    "metric": { "<label_name>": "<label_value>", ... },
    "value": [ <unix_time>, "<sample_value>" ]
  },
  ...
]
```

其中 metrics 表示当前时间序列的特征维度，value 只包含一个唯一的样本。

- 区间向量：matrix

当返回数据类型 resultType 为 matrix 时，result 响应格式如下：

```
[
  {
    "metric": { "<label_name>": "<label_value>", ... },
    "values": [ [ <unix_time>, "<sample_value>" ], ... ]
  },
  ...
]
```

其中 metrics 表示当前时间序列的特征维度，values 包含当前事件序列的一组样本。

- 标量：scalar

当返回数据类型 resultType 为 scalar 时，result 响应格式如下：

```
[ <unix_time>, "<scalar_value>" ]
```

由于标量不存在时间序列一说，因此 result 表示为当前系统时间一个标量的值。

- 字符串：string

当返回数据类型 resultType 为 string 时，result 响应格式如下：

```
[ <unix_time>, "<string_value>" ]
```

字符串类型的响应内容格式和标量相同。

### 区间数据查询

使用 QUERY_RANGE API 我们则可以直接查询 PromQL 表达式在一段时间返回内的计算结果。

```
GET /api/v1/query_range
```

URL 请求参数：

- query=<string>: PromQL 表达式。
- start=<rfc3339 | unix_timestamp>: 起始时间。
- end=<rfc3339 | unix_timestamp>: 结束时间。
- step=<duration>: 查询步长。
- timeout=<duration>: 超时设置。可选参数，默认情况下使用-query,timeout 的全局设置。

当使用 QUERY_RANGE API 查询 PromQL 表达式时，返回结果一定是一个区间向量：

```json
{
  "resultType": "matrix",
  "result": <value>
}
```

> 需要注意的是，在 QUERY_RANGE API 中 PromQL 只能使用瞬时向量选择器类型的表达式。

例如使用以下表达式查询表达式 up 在 30 秒范围内以 15 秒为间隔计算 PromQL 表达式的结果。

```json
$ curl 'http://localhost:9090/api/v1/query_range?query=up&start=2015-07-01T20:10:30.781Z&end=2015-07-01T20:11:00.781Z&step=15s'
{
   "status" : "success",
   "data" : {
      "resultType" : "matrix",
      "result" : [
         {
            "metric" : {
               "__name__" : "up",
               "job" : "prometheus",
               "instance" : "localhost:9090"
            },
            "values" : [
               [ 1435781430.781, "1" ],
               [ 1435781445.781, "1" ],
               [ 1435781460.781, "1" ]
            ]
         },
         {
            "metric" : {
               "__name__" : "up",
               "job" : "node",
               "instance" : "localhost:9091"
            },
            "values" : [
               [ 1435781430.781, "0" ],
               [ 1435781445.781, "0" ],
               [ 1435781460.781, "1" ]
            ]
         }
      ]
   }
}
```
