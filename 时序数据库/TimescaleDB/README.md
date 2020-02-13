# TimescaleDB

TimescaleDB 是 Timescale Inc.(成立于 2015 年)开发的一款号称兼容全 SQL 的时序数据库。它的本质是一个基于 PostgreSQL（以下简称 PG）的扩展（Extension），主打的卖点如下：

- 全 SQL 支持

- 背靠 PostgreSQL 的高可靠性

- 时序数据的高写入性能

同样的，TimescaleDB 开源了其单机版本，而集群版本仅在商业化方案中提供。

# 背景特性

## 数据模型

由于 TimescaleDB 的根基还是 PG，因此它的数据模型与 NoSQL 的时序数据库(如我们的阿里时序时空 TSDB，InfluxDB 等)截然不同。在 NoSQL 的时序数据库中，数据模型通常如下所示，即一条数据中既包括了时间戳以及采集的数据，还包括设备的元数据（通常以 Tagset 体现）。数据模型如下所示：

![NoSQL 时序模型](https://s2.ax1x.com/2019/11/24/MLj0yT.png)

但是在 TimescaleDB 中，数据模型必须以一个二维表的形式呈现，这就需要用户结合自己使用时序数据的业务场景，自行设计定义二维表。在 TimescaleDB 的官方文档中，对于如何设计时序数据的数据表，给出了两个范式：Narrow Table、Wide Table。

所谓的 Narrow Table 就是将 metric 分开记录，一行记录只包含一个 metricValue - timestamp。举例如下：

![单值模型](https://s2.ax1x.com/2019/11/24/MLjsw4.png)

而所谓的 Wide Table 就是以时间戳为轴线，将同一设备的多个 metric 记录在同一行，至于设备一些属性（元数据）则只是作为记录的辅助数据，甚至可直接记录在别的表（之后需要时通过 JOIN 语句进行查询）：

![多值模型](https://s2.ax1x.com/2019/11/24/MLvPts.png)

基本上可以认为：Narrow Table 对应的就是单值模型，而 Wide Table 对应的就是多值模型。由于采用的是传统数据库的关系表的模型，所以 TimescaleDB 的 metric 值必然是强类型的，它的类型可以是 PostgreSQL 中的 数值类型，字符串类型 等。

## 特性

TimescaleDB 在 PostgreSQL 的基础之上做了一系列扩展，主要涵盖以下方面：

- 时序数据表的透明自动分区特性

- 提供了若干面向时序数据应用场景的特殊 SQL 接口

- 针对时序数据的写入和查询对 PostgreSQL 的 Planner 进行扩展

- 面向时序数据表的定制化并行查询

其中 3 和 4 都是在 PostgreSQL 的现有机制上进行的面向时序数据场景的微创新。1 和 2 则是 TimescaleDB 的核心功能。综上所述，由于 TimescaleDB 完全基于 PostgreSQL 构建而成，因此它具有若干与生俱来的优势：

- 100%继承 PostgreSQL 的生态。且由于完整支持 SQL，对于未接触过时序数据的初学者反而更有吸引力

- 由于 PostgreSQL 的品质值得信赖，因此 TimescaleDB 在质量和稳定性上拥有品牌优势

- 强 ACID 支持

当然，它的短板也是显而易见的：

- 由于只是 PostgreSQL 的一个 Extension，因此它不能从内核/存储层面针对时序数据库的使用场景进行极致优化。

- 当前的产品架构来看仍然是一个单机库，不能发挥分布式技术的优势。而且数据虽然自动分区，但是由于时间戳决定分区，因此很容易形成 I/O 热点。

- 在功能层面，面向时序数据库场景的特性还比较有限。目前更像是一个 传统 OLTP 数据库 + 部分时序特性。

不管怎样，TimescaleDB 也算是面向时序数据库从另一个角度发起的尝试。在当前时序数据库仍然处于新兴事物的阶段，它未来的发展方向也是值得我们关注并借鉴的。
