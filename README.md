[![image.png](https://i.postimg.cc/j24P0gbr/image.png)](https://postimg.cc/wR37DkMW)

# 深入浅出分布式基础架构-数据库篇

![mindmap](https://i.postimg.cc/k4XKQvh4/Database.png)

# 数据库简史

数据库系统的萌芽出现于 60 年代。当时计算机开始广泛地应用于数据管理，对数据的共享提出了越来越高的要求。传统的文件系统已经不能满足人们的需要。能够统一管理和共享数据的数据库管理系统（DBMS）应运而生。1961 年通用电气公司（General ElectricCo.）的 Charles Bachman 成功地开发出世界上第一个网状 DBMS 也是第一个数据库管理系统—— 集成数据存储（Integrated DataStore IDS），奠定了网状数据库的基础。

1970 年，IBM 的研究员 E.F.Codd 博士在刊物 Communication of the ACM 上发表了一篇名为“A Relational Modelof Data for Large Shared Data Banks”的论文，提出了关系模型的概念，奠定了关系模型的理论基础。1974 年，IBM 的 Ray Boyce 和 DonChamberlin 将 Codd 关系数据库的 12 条准则的数学定义以简单的关键字语法表现出来，里程碑式地提出了 SQL（Structured Query Language）语言。在很长的时间内，关系数据库（如 MySQL 和 Oracle）对于开发任何类型的应用程序都是首选，巨石型架构也是应用程序开发的标准架构。

近年来随着互联网的迅猛发展，产生了庞大的数据，也催生了 NoSQL、NewSQL 等新一代数据库的出现。2009 年 MongoDB 开源，掀开了 NoSQL 的序幕，一时之间 NoSQL 的概念受人追捧，MongoDB 也因为其易用性迅速在社区普及。NoSQL 抛弃了传统关系数据库中的事务和数据一致性，从而在性能上取得了极大提升，并且天然支持分布式集群。

然而，不支持事务始终是 NoSQL 的痛点，让它无法在关键系统中使用。2012 年，Google 发布了 Spanner 论文，从此既支持分布式又支持事务的数据库逐渐诞生，。以 TiDB、蟑螂数据库等为代表的 NewSQL 身兼传统关系数据库和 NoSQL 的优点，开始崭露头角。2014 年亚马逊推出了基于新型 NVME SSD 虚拟存储层的 Aurora，它实现了完全兼容 MySQL 的超大单机数据库，于单点写多点读的主从架构做了进一步的发展，使得事务和存储引擎分离，为数据库架构的发展提供了具有实战意义的已实践用例。另外，各种不同用途的数据库也纷纷诞生并取得了较大的发展，比如用于 LBS 的地理信息数据库，用于监控和物联网的时序数据库，用于知识图谱的图数据库等。

与数据库技术的历史发展类似，数据库的托管方式在过去几十年中也发生了很大变化。在网络发展的早期，每个人都必须在自己的物理服务器上运行数据库，EC2 和 Digital Ocean 使这变得更容易，但仍需要深入的技术理解来手动操作数据库。诸如 Heroku 的 Postgres 服务，AWS RDS 和 Mongo Atlas 等托管服务抽象出了许多复杂的细节，数据库管理变得更加简单，但底层模型仍然相同，需要开发人员提前配置计算容量。最新开发的无服务器数据库使开发人员无需担心基础架构，因为他们的数据库只需根据实际使用情况进行扩展和缩小以匹配当前负载，其中 Aurora Serverless 和 CosmosDB 就是一个突出的例子。

# 数据库的特性

对于数据库的期许往往会包含以下几方面，首先是**易用与灵活**，尽可能可以用贴近业务语言的方式存取数据，而不需要理解太多抽象的语义或者函数；然后是**高性能**，无论存取皆可以迅速完成；然后是**高可用与可扩展**，我们能够根据实际的业务需要快速扩展数据库，提供长期的可用性与数据的安全一致，而不会因为数据的爆炸性增长导致数据库的崩溃。

以 Oracle, MySQL, SQLServer, PostgreSQL 为代表的关系型数据库，以行存储的方式结构化地存储数据。搜索引擎擅长文本查询；与 SQL 数据库中的文本匹配（例如 LIKE）相比，搜索引擎提供了更高的查询功能和更好的开箱即用性能。文档存储提供比传统数据库更好的数据模式适应性；通过将数据存储为单个文档对象（通常表示为 JSON），它们不需要预定义模式。列式存储专门用于单列查询和值聚合，在列式存储中，诸如 SUM 和 AVG 之类的 SQL 操作要快得多，因为同一列的数据在硬盘驱动器上更紧密地存储在一起。而 OceanBase, TiDB, Spanner／F1 等 NewSQL 数据库兼具了 SQL 以及可扩展性，数据被拆分成一个个 Range，分散在不同的服务器中，通过增加服务器就可以一定程度上的线性扩容；其通过 Paxos 或者 Raft 保证多副本之间的一致性，通过 2PC，MVCC 支持不同隔离级别的事务等。

![](https://www.confluent.io/wp-content/uploads/platform_chart_updated.png)

实际上，各个数据库之间也并发泾渭分明，多模异构是指在单个数据库平台中支持非结构化结构化数据在内的多种数据类型。一直以来，传统关系型数据库仅支持表单类型的结构化数据存储和访问能力，而对于层次型对象、图片影像等半结构化与非结构化数据管理无能为力。如今，随着应用类型的多样化和存储成本的降低，单一数据类型已经无法满足许多综合性业务平台的需求。数据库层面的多模异构和非结构化数据管理，将能实现结构化、半结构化和非结构化数据的统一管理，实现非结构化数据的实时访问，大大降低了运维和应用的成本。同时，非关系型数据库在访问模式上也渐渐将 SQL、K/V、文档、宽表、图等分支互相融合，支持除了 SQL 查询语言之外的其他访问模式，大大丰富了过去 NoSQL 数据库单一的设计用途。

本篇即是希望能够概述常见数据库的使用与内部原理，让我们对数据库有更深入的理解。

# About

![default](https://i.postimg.cc/y1QXgJ6f/image.png)

## Copyright | 版权

![License: CC BY-NC-SA 4.0](https://img.shields.io/badge/License-CC%20BY--NC--SA%204.0-lightgrey.svg) ![](https://parg.co/bDm)

笔者所有文章遵循 [知识共享 署名-非商业性使用-禁止演绎 4.0 国际许可协议](https://creativecommons.org/licenses/by-nc-nd/4.0/deed.zh)，欢迎转载，尊重版权。如果觉得本系列对你有所帮助，欢迎给我家布丁买点狗粮(支付宝扫码)~

![](https://i.postimg.cc/y1QXgJ6f/image.png?raw=true)

## Home & More | 延伸阅读

![](https://i.postimg.cc/CMDmg2SQ/image.png)

您可以通过以下导航来在 Gitbook 中阅读笔者的系列文章，涵盖了技术资料归纳、编程语言与理论、Web 与大前端、服务端开发与基础架构、云计算与大数据、数据科学与人工智能、产品设计等多个领域：

- 知识体系：《[Awesome Lists | CS 资料集锦](https://ngte-al.gitbook.io/i/)》、《[Awesome CheatSheets | 速学速查手册](https://ngte-ac.gitbook.io/i/)》、《[Awesome Interviews | 求职面试必备](https://github.com/wx-chevalier/Awesome-Interviews)》、《[Awesome RoadMaps | 程序员进阶指南](https://github.com/wx-chevalier/Awesome-RoadMaps)》、《[Awesome MindMaps | 知识脉络思维脑图](https://github.com/wx-chevalier/Awesome-MindMaps)》、《[Awesome-CS-Books | 开源书籍（.pdf）汇总](https://github.com/wx-chevalier/Awesome-CS-Books)》

- 编程语言：《[编程语言理论](https://ngte-pl.gitbook.io/i/)》、《[Java 实战](https://ngte-pl.gitbook.io/i/java/java)》、《[JavaScript 实战](https://ngte-pl.gitbook.io/i/javascript/javascript)》、《[Go 实战](https://ngte-pl.gitbook.io/i/go/go)》、《[Python 实战](https://ngte-pl.gitbook.io/i/python/python)》、《[Rust 实战](https://ngte-pl.gitbook.io/i/rust/rust)》

- 软件工程、模式与架构：《[编程范式与设计模式](https://ngte-se.gitbook.io/i/)》、《[数据结构与算法](https://ngte-se.gitbook.io/i/)》、《[软件架构设计](https://ngte-se.gitbook.io/i/)》、《[整洁与重构](https://ngte-se.gitbook.io/i/)》、《[研发方式与工具](https://ngte-se.gitbook.io/i/)》

* Web 与大前端：《[现代 Web 开发基础与工程实践](https://ngte-web.gitbook.io/i/)》、《[数据可视化](https://ngte-fe.gitbook.io/i/)》、《[iOS](https://ngte-fe.gitbook.io/i/)》、《[Android](https://ngte-fe.gitbook.io/i/)》、《[混合开发与跨端应用](https://ngte-fe.gitbook.io/i/)》

* 服务端开发实践与工程架构：《[服务端基础](https://ngte-be.gitbook.io/i/)》、《[微服务与云原生](https://ngte-be.gitbook.io/i/)》、《[测试与高可用保障](https://ngte-be.gitbook.io/i/)》、《[DevOps](https://ngte-be.gitbook.io/i/)》、《[Node](https://ngte-be.gitbook.io/i/)》、《[Spring](https://github.com/wx-chevalier/Spring-Series)》、《[信息安全与渗透测试](https://ngte-be.gitbook.io/i/)》

* 分布式基础架构：《[分布式系统](https://ngte-infras.gitbook.io/i/)》、《[分布式计算](https://ngte-infras.gitbook.io/i/)》、《[数据库](https://github.com/wx-chevalier/Database-Series)》、《[网络](https://ngte-infras.gitbook.io/i/)》、《[虚拟化与编排](https://ngte-infras.gitbook.io/i/)》、《[云计算与大数据](https://ngte-infras.gitbook.io/i/)》、《[Linux 与操作系统](https://github.com/wx-chevalier/Linux-Series)》

* 数据科学，人工智能与深度学习：《[数理统计](https://ngte-aidl.gitbook.io/i/)》、《[数据分析](https://ngte-aidl.gitbook.io/i/)》、《[机器学习](https://ngte-aidl.gitbook.io/i/)》、《[深度学习](https://ngte-aidl.gitbook.io/i/)》、《[自然语言处理](https://ngte-aidl.gitbook.io/i/)》、《[工具与工程化](https://ngte-aidl.gitbook.io/i/)》、《[行业应用](https://ngte-aidl.gitbook.io/i/)》

* 产品设计与用户体验：《[产品设计](https://ngte-pd.gitbook.io/i/)》、《[交互体验](https://ngte-pd.gitbook.io/i/)》、《[项目管理](https://ngte-pd.gitbook.io/i/)》

* 行业应用：《[行业迷思](https://github.com/wx-chevalier/Business-Series)》、《[功能域](https://github.com/wx-chevalier/Business-Series)》、《[电子商务](https://github.com/wx-chevalier/Business-Series)》、《[智能制造](https://github.com/wx-chevalier/Business-Series)》

此外，你还可前往 [xCompass](https://wx-chevalier.github.io/home/#/search) 交互式地检索、查找需要的文章/链接/书籍/课程；或者在 [MATRIX 文章与代码索引矩阵](https://github.com/wx-chevalier/Developer-Zero-To-Mastery)中查看文章与项目源代码等更详细的目录导航信息。最后，你也可以关注微信公众号：『**某熊的技术之路**』以获取最新资讯。
