# OpenTSDB

OpenTSDB 是一个分布式、可伸缩的时序数据库，支持高达每秒百万级的写入能力，支持毫秒级精度的数据存储，不需要降精度也可以永久保存数据。其优越的写性能和存储能力，得益于其底层依赖的 HBase，HBase 采用 LSM 树结构存储引擎加上分布式的架构，提供了优越的写入能力，底层依赖的完全水平扩展的 HDFS 提供了优越的存储能力。OpenTSDB 对 HBase 深度依赖，并且根据 HBase 底层存储结构的特性，做了很多巧妙的优化。关于存储的优化，我在这篇文章中有详细的解析。在最新的版本中，还扩展了对 BigTable 和 Cassandra 的支持。

![OpenTSDB 架构](https://s2.ax1x.com/2019/11/24/MOk9M9.png)

如图是 OpenTSDB 的架构，核心组成部分就是 TSD 和 HBase。TSD 是一组无状态的节点，可以任意的扩展，除了依赖 HBase 外没有其他的依赖。TSD 对外暴露 HTTP 和 Telnet 的接口，支持数据的写入和查询。TSD 本身的部署和运维是很简单的，得益于它无状态的设计，不过 HBase 的运维就没那么简单了，这也是扩展支持 BigTable 和 Cassandra 的原因之一吧。
