# TiDB

![image](https://user-images.githubusercontent.com/5803001/51458039-bcb43900-1d8e-11e9-8bb6-9c984ffeafbd.png)

最底层 TiKV 层，是分布式数据库的存储引擎层，不只是用来存取和管理数据，同时也负责执行对数据的并行运算。在 TiKV 之上即是 TiDB 层，为分布式数据库的 SQL 引擎层，处理关系型数据库诸如连接会话管理、权限控制、SQL 解析、优化器优化、执行器等核心功能。此外，还有一个承担集群大脑角色的集中调度器，叫做“PD”，同时整体架构中还会融合一些运维管理工具，包括部署、调度、监控、备份等。

TiDB 可实现自动水平伸缩，强一致性的分布式事务，基于 Raft 算法的多副本复制等重要 NewSQL 特性，并且也满足我行对于高可用、高性能、高可扩展性的要求。TiDB 部署简单，在线弹性扩容不影响业务，异地多活及自动故障恢复保障数据安全，同时兼容 MySQL 协议，使迁移使用成本降到极低。

# HTAP

TiDB 4.0 提供一套完备的 Hybrid transaction/analytical processing (HTAP) 解决方案，那就是 TiDB + TiFlash 。

![HTAP](https://s1.ax1x.com/2020/06/06/ty7Sr6.md.png)

简单来说，我们会在 TiDB 里面处理 OLTP 类型业务，在 TiFlash 里面处理 OLAP 类型业务，相比于传统的 ETL 方案，或者其他的 HTAP 解决方案，我们做了更多：

- 实时的强一致性。在 TiDB 里面更新的数据会实时的同步到 TiFlash，保证 TiFlash 在处理的时候一定能读取到最新的数据。

- TiDB 可以智能判断选择行存或者列存，以应对各种不同的查询场景，无需用户干预。
