# Alerting rules

告警规则语法如下：

```yml
# The name of the alert. Must be a valid metric name.
alert: <string>

# The PromQL expression to evaluate. Every evaluation cycle this is
# evaluated at the current time, and all resultant time series become
# pending/firing alerts.
expr: <string>

# Alerts are considered firing once they have been returned for this long.
# Alerts which have not yet fired for long enough are considered pending.
[ for: <duration> | default = 0s ]

# Labels to add or overwrite for each alert.
labels:
  [ <labelname>: <tmpl_string> ]

# Annotations to add to each alert.
annotations:
  [ <labelname>: <tmpl_string> ]
```

告警规则的例子为：

```yml
groups:
  - name: example
    rules:
      - alert: HighErrorRate
        expr: job:request_latency_seconds:mean5m{job="myjob"} > 0.5
        for: 10m
        labels:
          severity: page
        annotations:
          summary: High request latency

groups:
- name: example
  rules:

    # Alert for any instance that is unreachable for >5 minutes.
    - alert: InstanceDown
        expr: up == 0
        for: 5m
        labels:
        severity: page
        annotations:
        summary: "Instance {{ $labels.instance }} down"
        description: "{{ $labels.instance }} of job {{ $labels.job }} has been down for more than 5 minutes."

    # Alert for any instance that has a median request latency >1s.
    - alert: APIHighRequestLatency
        expr: api_http_request_latencies_second{quantile="0.5"} > 1
        for: 10m
        annotations:
        summary: "High request latency on {{ $labels.instance }}"
        description: "{{ $labels.instance }} has a median request latency above 1s (current value: {{ $value }}s)"
```

这个规则文件里，包含了两条告警规则：`InstanceDown` 和 `APIHighRequestLatency`。顾名思义，InstanceDown 表示当实例宕机时（up === 0）触发告警，APIHighRequestLatency 表示有一半的 API 请求延迟大于 1s 时（api_http_request_latencies_second{quantile="0.5"} > 1）触发告警。配置好后，需要重启下 Prometheus server，然后访问 `http://localhost:9090/rules` 可以看到刚刚配置的规则。
