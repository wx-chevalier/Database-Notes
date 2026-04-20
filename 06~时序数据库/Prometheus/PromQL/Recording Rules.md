# Recording rules

有一些监控的数据查询时很耗时的，还有一些数据查询所使用的查询语句很繁琐。Recording rules 可以把一些很耗时的查询或者很繁琐的查询进行提前查询好，然后在需要数据的时候就可以很快拉出数据。譬如：

```yml
# 规则分组 rule_group，不论是 recording rules 还是 alerting rules 都要在组里面
groups:
  # groups的名称
  - name: example
    #该组下的规则
    rules:
      [ - <rule> ... ]

# 案例
groups:
  - name: example
    rules:
      - record: job:http_inprogress_requests:sum
        expr: sum(http_inprogress_requests) by (job)

groups:
 - name: example
   rules:
    - record: job:process_cpu_seconds:rate5m
      expr: sum without(instance)(rate(process_cpu_seconds_total[5m]))
    - record: job:process_open_fds:max
      expr: max without(instance)(process_open_fds)
```

## 语法校验

完整的规则语法如下：

```sh
groups:
  [ - <rule_group> ]

<rule_group>的语法
# 规则组名 必须是唯一的
name: <string>

# 规则评估间隔时间
[ interval: <duration> | default = global.evaluation_interval ]

rules:
  [ - <rule> ... ]

<rule>的语法
# 收集的指标名称
record: <string>

# 评估时间
# evaluated at the current time, and the result recorded as a new set of
# time series with the metric name as given by 'record'.
expr: <string>

# Labels to add or overwrite before storing the result.
labels:
  [ <labelname>: <labelvalue> ]
```

要在不启动 Prometheus 进程的情况下快速检查规则文件是否在语法上正确，可以通过安装并运行 Prometheus 的 promtool 命令行工具来校验：

```sh
$ go get github.com/prometheus/prometheus/cmd/promtool
```

# 案例

## 减少基数

如果我们有如下的表达式：

```sh
$ sum without(instance)(rate(process_cpu_seconds_total{job="node"}[5m]))
```

如果直接在 Dashboard 中输入如上的 PromQL，在数据量较大的情况下，极有可能会导致过长的计算时间。我们可以使用如下的记录规则来构建新的指标：

```yml
groups:
  - name: node
    rules:
      - record: job:process_cpu_seconds:rate5m
        expr: >
          sum without(instance)(
            rate(process_cpu_seconds_total{job="node"}[5m])
          )
```

现在，您只需要在渲染仪表板时获取一个时间序列即可，即使正在播放检测标签，也能快速地返回结果，因为您将要处理的时间序列数减少了多少个实例。实际上，您正在以持续资源成本进行交易，而查询的延迟和资源成本却大大降低。由于这种折衷，使用长向量范围的规则通常是不明智的，因为这样的查询往往很昂贵，并且定期运行它们会导致性能问题。您应该尝试将一项工作的所有规则归为一组。这样，它们将具有相同的时间戳，并在您对它们进行进一步的数学运算时避免出现失真。一组中的所有记录规则对于执行都具有相同的查询评估时间，并且所有输出样本也将具有该时间戳。您通常会有基于相同度量标准但具有不同标签集的聚合规则。通过让一个规则使用另一个规则的输出，您可以提高效率，而不是单独计算每个聚合。例如：

```yml
groups:
  - name: node
    rules:
      - record: job_device:node_disk_read_bytes:rate5m
        expr: >
          sum without(instance)(
            rate(node_disk_read_bytes_total{job="node"}[5m])
          )
      - record: job:node_disk_read_bytes:rate5m
        expr: >
          sum without(device)(
            job_device:node_disk_read_bytes:rate5m{job="node"}
          )
```

## 组合 Range Vector 函数

如前所述，不能在产生即时矢量的函数的输出上使用范围矢量函数，譬如 `max_over_time(sum without(instance)(rate(x_total[5m]))[1h])` 这样的操作就是不可行的。我们可以利用 Recording Rules 实现这样的功能：

```yml
groups:
  - name: j_job_rules
    rules:
      - record: job:x:rate5m
        expr: >
          sum without(instance)(
            rate(x_total{job="j"}[5m])
          )
      - record: job:x:max_over_time1h_rate5m
        expr: max_over_time(job:x:rate5m{job="j"}[1h])
```

此方法可与任何范围向量函数一起使用，不仅包括 `_over_time` 函数，而且包括 `predict_linear`，deriv 和 holt_winters。但是，此技术不应与 rate, irate, 或者 increase 一起使用，因为有效的速率表达 `(sum（x_total）[5m])` 每次其组成计数器之一重置或消失时都会产生大量尖峰。
