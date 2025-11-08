# Milvus

Milvus 是一个具有分布式系统原生后端的矢量数据库。它是专门为处理索引、存储和查询 10 亿规模的矢量数据而设计的。Milvus 使用多层和多种类型的工作节点，以实现易于扩展的设计。除了使用多个单一用途的节点，Milvus 还使用分段数据，以提高索引的效率。Milvus 使用 512MB 的数据段，这些数据段在填充后不会被改变，并对其进行并行查询，以提供整个行业的最低延迟。

![High-level overview of Milvus's architecture](https://ngte-superbed.oss-cn-beijing.aliyuncs.com/item/architecture_diagram_c2acfbe310.png)
