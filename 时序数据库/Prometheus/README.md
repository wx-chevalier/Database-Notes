> 导读：参考《[Kubernetes 实战](https://github.com/wx-chevalier/Cloud-Series)》了解 K8s 中 Prometheus 的应用。

# 基于 Prometheus 的线上应用监控

Prometheus 是一套开源的监控、报警、时间序列数据库的组合，它的灵感来自谷歌的 Borgmon。Prometheus 最初由前谷歌 SRE Matt T. Proud 开发，并转为一个研究项目；在 Proud 加入 SoundCloud 之后，他与另一位工程师 Julius Volz 合作开发了 Prometheus。随着发展，越来越多公司和组织接受采用 Prometheus，社区也十分活跃，他们便将它独立成开源项目，并且有公司来运作。Google SRE 的书内也曾提到跟他们 BorgMon 监控系统相似的实现是 Prometheus。

Prometheus 的优势在于其易于安装使用，外部依赖较少；并且直接按照分布式、微服务架构模式进行设计，支持服务自动化发现与代码集成。Prometheus 能够自定义多维度的数据模型，内置强大的查询语句，搭配其丰富的社区扩展，能够轻松实现数据可视化。而随着容器、云计算、云原生的发展，目前 Prometheus 已经广泛用于 Kubernetes 集群的监控系统中。2015 年 7 月，隶属于 Linux 基金会的 云原生计算基金会（CNCF，Cloud Native Computing Foundation）应运而生。第一个加入 CNCF 的项目是 Google 的 Kubernetes，而 Prometheus 是第二个加入的（2016 年）。

![Prometheus 生态系统](https://i.postimg.cc/g0SDCRhK/image.png)

上图左侧是各种符合 Prometheus 数据格式的 exporter，除此之外为了支持推动数据类型的 Agent，可以通过 Pushgateway 组件，将 Push 转化为 Pull。Prometheus 甚至可以从其它的 Prometheus 获取数据，组建联邦集群。

# 背景特性

## 关键功能

- 多维度数据模型：多维度数据模型和强大的查询语言这两个特性，正是时序数据库所要求的，所以 Prometheus 不仅仅是一个监控系统，同时也是一个时序数据库。

- 方便的部署和维护：纵观比较流行的时序数据库，他们要么组件太多，要么外部依赖繁重，比如：Druid 有 Historical、MiddleManager、Broker、Coordinator、Overlord、Router 一堆的组件，而且还依赖于 ZooKeeper、Deep storage（HDFS 或 S3 等），Metadata store（PostgreSQL 或 MySQL），部署和维护起来成本非常高。而 Prometheus 采用去中心化架构，可以独立部署，不依赖于外部的分布式存储，你可以在几分钟的时间里就可以搭建出一套监控系统。

- 灵活的数据采集：要采集目标的监控数据，首先需要在目标处安装数据采集组件，这被称之为 Exporter，它会在目标处收集监控数据，并暴露出一个 HTTP 接口供 Prometheus 查询，Prometheus 通过 Pull 的方式来采集数据，这和传统的 Push 模式不同。不过 Prometheus 也提供了一种方式来支持 Push 模式，你可以将你的数据推送到 Push Gateway，Prometheus 通过 Pull 的方式从 Push Gateway 获取数据。目前的 Exporter 已经可以采集绝大多数的第三方数据，比如 Docker、HAProxy、StatsD、JMX 等等。

- 强大的查询语言

除了这四大特性，随着 Prometheus 的不断发展，开始支持越来越多的高级特性，比如：服务发现，更丰富的图表展示，使用外部存储，强大的告警规则和多样的通知方式。

# 优劣对比

## InfluxDB

同 InfluxDB 相比, 在场景方面，InfluxDB 面向的是通用时序平台，包括日志监控等场景。而 Prometheus 更侧重于指标方案。PTSDB 适合数值型的时序数据。不适合日志型时序数据和用于计费的指标统计。两个系统之间有非常多的相似之处，包括采集，存储，报警，展示等等：

| 类目     | InfluxDB                                                         | Prometheus                                                                               |
| -------- | ---------------------------------------------------------------- | ---------------------------------------------------------------------------------------- |
| 生态组合 | telegraf+Influxdb+Kapacitor+Chronograf                           | exporter+prometheus server+AlertManager+Grafana                                          |
| 采集端   | 采集工具 telegraf 则主打推的方式                                 | 主推拉的模式，同时通过 push gateway 支持推的模式                                         |
| 存储     | 基本思想上相通，关键点上有差异包括：时间线的索引，乱序的处理等等 |
| 数据模型 | 多值模型                                                         | 单值模型                                                                                 |
| 集群模式 | 只保留了基于 relay 的高可用，集群模式作为商业版本的特性发布      | 提供了一种很有特色的 cluster 模式，通过多层次的 proxy 来聚合多个 prometheus 节点实现扩展 |

## OpenTSDB

OpenTSDB 的数据模型与 Prometheus 几乎相同，查询语言上 PromQL 更简洁，OpenTSDB 功能更丰富。OpenTSDB 依赖的是 Hadoop 生态,Prometheus 成长于 Kubernetes 生态。

# 链接

- https://mp.weixin.qq.com/s/0vZLCZBPFfOMNqubpQUrbg
- https://mp.weixin.qq.com/s/0vZLCZBPFfOMNqubpQUrbg
- https://mp.weixin.qq.com/s/ijx8zIUmpmBB6akh8Ru0zw Prometheus 基础知识介绍
- https://learning.oreilly.com/library/view/prometheus-up/9781492034131/preface01.html#idm46043484494872
