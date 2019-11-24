# TSDB

TSDB( Time Series and Spatial-Temporal Database , 简称 TSDB) 是一种集时序时空数据高效读写，压缩存储，实时计算能力为一体的数据库服务。可广泛应用于物联网和互联网领域，实现对设备及业务服务的实时监控，实时预测告警等等。阿里巴巴 TSDB 在集团内部孵化，聚焦于业务逐步经历了 3 个阶段的演进。

- 第一阶段以 HiStore+HBase 双引擎接入 OpenTSDB 体系。
- 第二阶段自研 TSDB 核心引擎(倒排+压缩+流式聚合)。
- 第三阶段形成整体生态体系：边云一体，TSQL，Prometheus 兼容，时空引擎等等。
