# Alertmanager

虽然 Prometheus 的 /alerts 页面可以看到所有的告警，但是还差最后一步：触发告警时自动发送通知。这是由 Alertmanager 来完成的，我们首先 下载并安装 Alertmanager，和其他 Prometheus 的组件一样，Alertmanager 也是开箱即用的。Alertmanager 启动后默认可以通过 http://localhost:9093/ 来访问，但是现在还看不到告警，因为我们还没有把 Alertmanager 配置到 Prometheus 中，我们回到 Prometheus 的配置文件 prometheus.yml，添加下面几行：

```sh
alerting:
  alertmanagers:
  - scheme: http
    static_configs:
    - targets:
      - "192.168.0.107:9093"
```

这个配置告诉 Prometheus，当发生告警时，将告警信息发送到 Alertmanager，Alertmanager 的地址为 http://192.168.0.107:9093。也可以使用命名行的方式指定 Alertmanager：

```s
$ ./prometheus -alertmanager.url=http://192.168.0.107:9093
```

# 告警信息推送

默认的配置文件 alertmanager.yml：

```yml
global:
  resolve_timeout: 5m

route:
  group_by: ["alertname"]
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: "web.hook"
receivers:
  - name: "web.hook"
    webhook_configs:
      - url: "http://127.0.0.1:5001/"
inhibit_rules:
  - source_match:
      severity: "critical"
    target_match:
      severity: "warning"
    equal: ["alertname", "dev", "instance"]
```

其中 global 块表示一些全局配置；route 块表示通知路由，可以根据不同的标签将告警通知发送给不同的 receiver，这里没有配置 routes 项，表示所有的告警都发送给下面定义的 web.hook 这个 receiver；如果要配置多个路由，可以参考 这个例子：

```yml
routes:
  - receiver: "database-pager"
    group_wait: 10s
    match_re:
      service: mysql|cassandra

  - receiver: "frontend-pager"
    group_by: [product, environment]
    match:
      team: frontend
```

紧接着，receivers 块表示告警通知的接收方式，每个 receiver 包含一个 name 和一个 xxx_configs，不同的配置代表了不同的接收方式，Alertmanager 内置了下面这些接收方式：

- email_config
- hipchat_config
- pagerduty_config
- pushover_config
- slack_config
- opsgenie_config
- victorops_config
- wechat_configs
- webhook_config
