# 告警和通知

作为一个监控系统，最重要的功能，还是应该能及时发现系统问题，并及时通知给系统负责人，这就是 Alerting（告警）。Prometheus 的告警功能被分成两部分：一个是告警规则的配置和检测，并将告警发送给 Alertmanager，另一个是 Alertmanager，它负责管理这些告警，去除重复数据，分组，并路由到对应的接收方式，发出报警。常见的接收方式有：Email、PagerDuty、HipChat、Slack、OpsGenie、WebHook 等。
