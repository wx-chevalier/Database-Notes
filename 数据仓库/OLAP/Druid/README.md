# Druid

Druid 单词来源于西方古罗马的神话人物，中文常常翻译成德鲁伊。Druid 是一个分布式的支持实时分析的数据存储系统（Data Store）。美国广告技术公司 MetaMarkets 于 2011 年创建了 Druid 项目，并且于 2012 年晚期开源了 Druid 项目。Druid 设计之初的想法就是为分析而生，它在处理数据的规模、数据处理的实时性方面，比传统的 OLAP 系统有了显著的性能改进，而且拥抱主流的开源生态，包括 Hadoop 等。多年以来，Druid 一直是非常活跃的开源项目。

# 背景分析

## 设计原则

在设计之初，开发人员确定了三个设计原则（Design Principle）。

- 快速查询（Fast Query）：部分数据的聚合（Partial Aggregate）+内存化（In-emory）+索引（Index）。

- 水平扩展能力（Horizontal Scalability）：分布式数据（Distributed Data）+ 并行化查询（Parallelizable Query）。

- 实时分析（Realtime Analytics）：不可变的过去，只追加的未来（Immutable Past，Append-Only Future）。

### 快速查询（Fast Query）

对于数据分析场景，大部分情况下，我们只关心一定粒度聚合的数据，而非每一行原始数据的细节情况。因此，数据聚合粒度可以是 1 分钟、5 分钟、1 小时或 1 天等。部分数据聚合（Partial Aggregate）给 Druid 争取了很大的性能优化空间。

数据内存化也是提高查询速度的杀手锏。内存和硬盘的访问速度相差近百倍，但内存的大小是非常有限的，因此在内存使用方面要精细设计，比如 Druid 里面使用了 Bitmap 和各种压缩技术。

另外，为了支持 Drill-Down 某些维度，Druid 维护了一些倒排索引。这种方式可以加快 AND 和 OR 等计算操作。

### 水平扩展能力（Horizontal Scalability）

Druid 查询性能在很大程度上依赖于内存的优化使用。数据可以分布在多个节点的内存中，因此当数据增长的时候，可以通过简单增加机器的方式进行扩容。为了保持平衡，Druid 按照时间范围把聚合数据进行分区处理。对于高基数的维度，只按照时间切分有时候是不够的（Druid 的每个 Segment 不超过 2000 万行），故 Druid 还支持对 Segment 进一步分区。

历史 Segment 数据可以保存在深度存储系统中，存储系统可以是本地磁盘、HDFS 或远程的云服务。如果某些节点出现故障，则可借助 Zookeeper 协调其他节点重新构造数据。

Druid 的查询模块能够感知和处理集群的状态变化，查询总是在有效的集群架构中进行。集群上的查询可以进行灵活的水平扩展。Druid 内置提供了一些容易并行化的聚合操作，例如 Count、Mean、Variance 和其他查询统计。对于一些无法并行化的操作，例如 Median，Druid 暂时不提供支持。在支持直方图（Histogram）方面，Druid 也是通过一些近似计算的方法进行支持，以保证 Druid 整体的查询性能，这些近似计算方法还包括 HyperLoglog、DataSketches 的一些基数计算。

### 实时分析（Realtime Analytics）

Druid 提供了包含基于时间维度数据的存储服务，并且任何一行数据都是历史真实发生的事件，因此在设计之初就约定事件一但进入系统，就不能再改变。

对于历史数据 Druid 以 Segment 数据文件的方式组织，并且将它们存储到深度存储系统中，例如文件系统或亚马逊的 S3 等。当需要查询这些数据的时候，Druid 再从深度存储系统中将它们装载到内存供查询使用。
