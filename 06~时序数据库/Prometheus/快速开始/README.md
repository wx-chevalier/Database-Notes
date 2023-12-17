# 部署配置

Prometheus 生态系统包含了几个关键的组件：Prometheus server、Pushgateway、Alertmanager、Web UI 等。

![架构组件](https://s2.ax1x.com/2020/01/03/lU802Q.png)

上侧是服务发现，Prometheus 支持监控对象的自动发现机制，从而可以动态获取监控对象。图片中间是 Prometheus Server，Retrieval 模块定时拉取数据，并通过 Storage 模块保存数据。PromQL 为 Prometheus 提供的查询语法，PromQL 模块通过解析语法树，调用 Storage 模块查询接口获取监控数据。图片右侧是告警和页面展现，Prometheus 将告警推送到 alertmanger，然后通过 alertmanger 对告警进行处理并执行相应动作。数据展现除了 Prometheus 自带的 WebUI，还可以通过 Grafana 等组件查询 Prometheus 监控数据。

![组成模块](https://s2.ax1x.com/2019/11/20/MW5ixH.png)

具体而言，Prometheus 的整体技术架构可以分为几个重要模块：

- Main function：作为入口承担着各个组件的启动，连接，管理。以 Actor-Like 的模式协调组件的运行

- Configuration：配置项的解析，验证，加载

- Scrape discovery manager：服务发现管理器同抓取服务器通过同步 channel 通信，当配置改变时需要重启服务生效。

- Scrape manager：抓取指标并发送到存储组件

- Storage：

  - Fanout Storage：存储的代理抽象层，屏蔽底层 local storage 和 remote storage 细节，samples 向下双写，合并读取。
  - Remote Storage：Remote Storage 创建了一个 Queue 管理器，基于负载轮流发送，读取客户端 merge 来自远端的数据。
  - Local Storage：基于本地磁盘的轻量级时序数据库。

- PromQL engine：查询表达式解析为抽象语法树和可执行查询，以 Lazy Load 的方式加载数据。

- Rule manager：告警规则管理

- Notifier：通知派发管理器

- Notifier discovery：通知服务发现

- Web UI and API：内嵌的管控界面，可运行查询表达式解析，结果展示。
