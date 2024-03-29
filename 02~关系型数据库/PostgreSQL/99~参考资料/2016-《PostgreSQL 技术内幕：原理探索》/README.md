> [原文地址](https://pg-internal.vonng.com/#/preface)

# PostgreSQL 技术内幕：原理探索

数据库是信息系统的核心组件，关系型数据库则是数据库皇冠上的明珠，而 PostgreSQL 的 Title 是”世界上最先进的开源关系型数据库“。PostgreSQL 在各行各业，各种场景下都有着广泛应用。但会用，只是”知其然“，知道背后的原理，才能“知其所以然”。对数据库原理，及其具体实现的理解，能让架构师以最小的复杂度代价实现所需的功能，能让程序员以最少的复杂度代价写出更可靠高效的代码，能让 DBA 在遇到疑难杂症时带来精准的直觉与深刻的洞察。

数据库是一个博大精深的领域，存储 I/O 计算无所不包。PostgreSQL 可以视作关系型数据库实现的典范，用 100 万行不到的 C 代码实现了功能如此丰富的软件系统可谓凝练无比。它的每一个功能模块都值得用一本甚至几本书的篇幅去介绍。本书限于篇幅虽无法一一深入所有细节，但它为读者进一步深入理解 PostgreSQL 提供了一副全局的概念地图。读者完全可以顺着各个章节的线索，以点破面，深入挖掘源码背后的设计思路。
