# SQL

SQL 是 Structrued Query Language 的缩写，即结构化查询语言。它是负责与 ANSI(美国国家标准学会)维护的数据库交互的标准。SQL 作为关系数据库的标准语言，它已被众多商用 DBMS 产品所采用，使得它已成为关系数据库领域中一个主流语言，不仅包含数据查询功能，还包括插入、删除、更新和数据定义功能。最为重要的 SQL92 版本的详细标准可以查看[这里](http://www.contrib.andrew.cmu.edu/~shadow/sql/sql1992.txt)，或者在 [Wiki](https://en.wikipedia.org/wiki/SQL) 上查看 SQL 标准的变化。SQL 语言的功能包括查询、操纵、定义和控制，通常会将 SQL 语句分为以下几类：

- 数据定义语言（DDL）：DDL 包括用于创建表，删除表或创建和删除数据库的其他方面的命令。

- 数据操作语言（DML）：DML 包括用于查询和修改数据库的命令。它包括用于查询数据库的 select 语句，以及用于修改数据库的 insert，update 和 delete 语句。

作为查询语言，与普通编程语言相比，它还处于业务上层；SQL 最终会转化为关系代数执行，但关系代数会遵循一些等价的转换规律，比如交换律、结合律、过滤条件拆分等等，通过预估每一步的时间开销，将 SQL 执行顺序重新组合，可以提高执行效率。如果有多个 SQL 同时执行，还可以整合成一个或多个新的 SQL，合并重复的查询请求；在数据驱动商业的今天，SQL 依然是数据查询最通用的解决方案。

# 背景分析

## SQL 简史

SQL 的开发历史如下：

- 1970 E.F. Codd publishes [Definition of Relational Model](https://www.w3resource.com/sql/sql-basic/codd-12-rule-relation.php)
- 1975 Initial version of SQL Implemented (D. Chamberlin)
- IBM experimental version: System R (1977) w/revised SQL
- IBM commercial versions: SQL/DS and DB2 (the early 1980s)
- Oracle introduces commercial version before IBM's SQL/DS
- INGRES 1981 & 85
- ShareBase 1982 & 86
- Data General (1984)
- Sybase (1986)
- by 1992 over 100 SQL products

其标准颁布的更迭如下：

- SEQUEL/Original SQL - 1974
- SQL/86 - Ratification and acceptance of a formal SQL standard by ANSI (American National Standards Institute) and ISO (International Standards Organization).
- SQL/92 - Major revision (ISO 9075), Entry Level SQL-92 adopted as FIPS 127-2.
- SQL/99 - Added regular expression matching, recursive queries (e.g. transitive closure), triggers, support for procedural and control-of-flow statements, non-scalar types, and some object-oriented features (e.g. structured types).
- SQL/2003 - Introduced XML-related features (SQL/XML), Window functions, Auto generation.
- SQL/2006 - Lots of XML Support for XQuery, an XML-SQL interface standard.
- SQL/2008 - Adds INSTEAD OF triggers, TRUNCATE statement.

目前主要有两种 SQL 版本：

- T-SQL 是 SQL 语言的一种版本，且只能在 SQL SERVER 上使用。它是 ANSI SQL 的加强版语言、提供了标准的 SQL 命令。另外，T-SQL 还对 SQL 做了许多补允，提供了数据库脚本语言，即类似 C、Basic 和 Pascal 的基本功能，如变量说明、流控制语言、功能函数等。

- PL-SQL(Procedural Language-SQL)是一种增加了过程化概念的 SQL 语言，是 Oracle 对 SQL 的扩充。与标准 SQL 语言相同，PL-SQL 也是 Oracle 客户端工具(如 SQL Plus、Developer/2000 等)访问服务器的操作语言。它有标准 SQL 所没有的特征：变量(包括预先定义的和自定义的)；控制结构(如 IF-THEN-ELSE 等流控制语句)；自定义的存储过程和函数；对象类型等。由于 P/L-SQL 融合了 SQL 语言的灵活性和过程化的概念，使得 P/L-SQL 成为了一种功能强大的结构化语言，可以设计复杂的应用。

## 声明式语言

SQL 是一种声明式(Declarative)的编程语言，相比一般的编程语言描述的是程序执行的过程，这类编程语言则是描述问题或者需要的结果本身。具体的执行步骤则交由程序自己决定。

- 从技术的角度来说，通过对用户输入的查询进行优化，实现更优的执行步骤规划数据库可以实现更快的执行和更少的 IO 消耗。从而节约资源提高性能。

- 从使用的角度，SQL 作为一种可以被非相关技术人员快速入手的编程语言，其主要优点就在于即使用户因并不了解数据库内部的实现细节而写出来十分糟糕的查询语句，只要表达的意思准确清楚，数据库就可以在一定程度上将其转化为合理的执行方案高效的返回结果，极大的降低了使用门槛。

# 关系代数

SQL 作为一项图灵奖级别的发明，其重要意义不单单是发明了一种可以用作数据查询的语言，更重要的一点是发明了关系代数(Relation Algebra)这一工具，使得计算机理解和处理查询的语义更加方便。SQL 查询语句的优化也是基于关系代数这一模型。所谓关系代数，是 SQL 从语句到执行计划的一种中间表示。首先它不是单纯的抽象语法树(AST)，而是一种经过进一步处理得到的中间表示(可以类比一般编程语言的 IR)。

关系代数是一种抽象的查询语言，它用对关系的运算来表达查询。任何一种运算都是将一定的运算符作用于一定的运算对象上，得到预期的结果。所以运算对象、运算符、运算结果是运算的三大要素。按运算符的不同分为传统的集合运算和专门的关系运算两类：

- 传统的集合运算包括：并（∪）、差（−）、交（∩）、笛卡尔积（×）。
- 专门的关系运算包括：选择（σ）、投影（π）、连接（⋈）、除运算（÷）。

[![image.png](https://i.postimg.cc/dVjs2b9F/image.png)](https://postimg.cc/0zr1xH3X)
