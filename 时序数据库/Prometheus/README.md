# 基于 Prometheus 的线上应用监控

Prometheus 是一套开源的监控、报警、时间序列数据库的组合；它的灵感来自谷歌的 Borgmon，一个实时的时间序列监控系统，Borgmon 使用这些时间序列数据来识别问题并发出警报。Prometheus 最初由前谷歌 SRE Matt T. Proud 开发，并转为一个研究项目。在 Proud 加入 SoundCloud 之后，他与另一位工程师 Julius Volz 合作开发了 Prometheus。后来其他开发人员陆续加入了这个项目，并在 SoundCloud 内部继续开发，最终于 2015 年 1 月公开发布。

起始是由 SoundCloud 公司开发的。随着发展，越来越多公司和组织接受采用 Prometheus，社区也十分活跃，他们便将它独立成开源项目，并且有公司来运作。google SRE 的书内也曾提到跟他们 BorgMon 监控系统相似的实现是 Prometheus。现在最常见的 Kubernetes 容器管理系统中，通常会搭配 Prometheus 进行监控。

Prometheus 的优势在于其易于安装使用，外部依赖较少；并且直接按照分布式、微服务架构模式进行设计，支持服务自动化发现与代码集成。Prometheus 能够自定义多维度的数据模型，内置强大的查询语句，搭配其丰富的社区扩展，能够轻松实现数据可视化。

# 背景特性

## 关键功能

- 多维数据模型：metric，labels

- 灵活的查询语言：PromQL， 在同一个查询语句，可以对多个 metrics 进行乘法、加法、连接、取分数位等操作。

- 可独立部署，拆箱即用，不依赖分布式存储

- 通过 Http pull 的采集方式

- 通过 push gateway 来做 push 方式的兼容

- 通过静态配置或服务发现获取监控项

- 支持图表和 dashboard 等多种方式

## 优劣对比

同 InfluxDB 相比, 在场景方面：PTSDB 适合数值型的时序数据。不适合日志型时序数据和用于计费的指标统计。InfluxDB 面向的是通用时序平台，包括日志监控等场景。而 Prometheus 更侧重于指标方案。两个系统之间有非常多的相似之处，包括采集，存储，报警，展示等等

- Influxdata 的组合有：telegraf+Influxdb+Kapacitor+Chronograf
- Promethues 的组合有：exporter+prometheus server+AlertManager+Grafana
- 采集端 prometheus 主推拉的模式，同时通过 push gateway 支持推的模式。influxdata 的采集工具 telegraf 则主打推的方式。
- 存储方面二者在基本思想上相通，关键点上有差异包括：时间线的索引，乱序的处理等等。
- 数据模型上 Influxdb 支持多值模型，String 类型等，更丰富一些。
- Kapacitor 是一个混合了 prometheus 的数据处理,存储规则,报警规则以及告警通知功能的工具.而 AlertManager 进一步提供了分组,去重等等。
- influxdb 之前提供的 cluster 模式被移除了，现在只保留了基于 relay 的高可用，集群模式作为商业版本的特性发布。prometheus 提供了一种很有特色的 cluster 模式，通过多层次的 proxy 来聚合多个 prometheus 节点实现扩展。
  同时开放了 remote storage，因此二者又相互融合，Influxdb 作为 prometheus 的远端存储。

OpenTSDB 的数据模型与 Prometheus 几乎相同，查询语言上 PromQL 更简洁，OpenTSDB 功能更丰富。OpenTSDB 依赖的是 Hadoop 生态,Prometheus 成长于 Kubernetes 生态。

# 架构组件

![Prometheus 生态系统](https://i.postimg.cc/g0SDCRhK/image.png)

Prometheus 由两个部分组成，一个是监控报警系统，另一个是自带的时序数据库（TSDB）。上图是 Prometheus 整体架构图，左侧是各种符合 Prometheus 数据格式的 exporter，除此之外为了支持推动数据类型的 Agent，可以通过 Pushgateway 组件，将 Push 转化为 Pull。Prometheus 甚至可以从其它的 Prometheus 获取数据，组建联邦集群。

Prometheus 的基本原理是通过 HTTP 周期性抓取被监控组件的状态，任意组件只要提供对应的 HTTP 接口并且符合 Prometheus 定义的数据格式，就可以接入 Prometheus 监控。

![组件连接](https://s2.ax1x.com/2019/11/20/MWIR7q.png)

上侧是服务发现，Prometheus 支持监控对象的自动发现机制，从而可以动态获取监控对象。图片中间是 Prometheus Server，Retrieval 模块定时拉取数据，并通过 Storage 模块保存数据。PromQL 为 Prometheus 提供的查询语法，PromQL 模块通过解析语法树，调用 Storage 模块查询接口获取监控数据。图片右侧是告警和页面展现，Prometheus 将告警推送到 alertmanger，然后通过 alertmanger 对告警进行处理并执行相应动作。数据展现除了 Prometheus 自带的 WebUI，还可以通过 Grafana 等组件查询 Prometheus 监控数据。

![组成模块](https://s2.ax1x.com/2019/11/20/MW5ixH.png)

Prometheus 的整体技术架构可以分为几个重要模块：

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

# 链接

- https://mp.weixin.qq.com/s/0vZLCZBPFfOMNqubpQUrbg
- https://mp.weixin.qq.com/s/0vZLCZBPFfOMNqubpQUrbg
- https://mp.weixin.qq.com/s/ijx8zIUmpmBB6akh8Ru0zw Prometheus 基础知识介绍
