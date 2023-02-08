![InfluxDB 架构图](https://s2.ax1x.com/2019/11/20/MWuNZQ.md.png)

# InfluxDB

InfluxDB 是一个由 InfluxData 开发的开源时序型数据库。它由 Go 写成，着力于高性能地查询与存储时序型数据。InfluxDB 被广泛应用于存储系统的监控数据，它还是没有额外依赖的开源时序数据库，用于记录 metrics、events，进行数据分析。

# 背景特性

## 存储引擎的演进

尽管 InfluxDB 自发布以来历时三年多，其存储引擎的技术架构已经做过几次重大的改动, 以下将简要介绍一下 InfluxDB 的存储引擎演进的过程。并且 InfluxDB 的集群版已在 0.12 版就不再开源。

- 版本 0.9.0 之前：基于 LevelDB 的 LSMTree 方案

- 版本 0.9.0 ～ 0.9.4：基于 BoltDB 的 mmap COW B+tree 方案

- 版本 0.9.5 ～ 1.2：基于自研的 WAL + TSMFile 方案（TSMFile 方案是 0.9.6 版本正式启用，0.9.5 只是提供了原型）

- 版本 1.3 ～至今：基于自研的 WAL + TSMFile + TSIFile 方案

InfluxDB 的存储引擎先后尝试过包括 LevelDB，BoltDB 在内的多种方案。但是对于 InfluxDB 的下述诉求终不能完美地支持：

- 时序数据在降采样后会存在大批量的数据删除 => LevelDB 的 LSMTree 删除代价过高。

- 单机环境存放大量数据时不能占用过多文件句柄 => LevelDB 会随着时间增长产生大量小文件。

- 大数据场景下写吞吐量要跟得上 => BoltDB 的 B+tree 写操作吞吐量成瓶颈

- 存储需具备良好的压缩性能 => BoltDB 不支持压缩

基于上述痛点，InfluxDB 团队决定自己做一个存储引擎的实现。
