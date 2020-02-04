# 数据库设计

# Data Definition Language | 数据定义

DDL 包含 CREATE, ALTER, DROP 等常见的数据定义语句

[完整的表结构 SQL](https://gist.github.com/wx-chevalier/ebd1ceb919a68e428e7901f7fc766f02)

```sql
CREATE TABLE `product` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT COMMENT '_ID，内部自增编号',
  `code` varchar(6) DEFAULT NULL,
  `name` varchar(15) DEFAULT NULL,
  `category` varchar(15) DEFAULT NULL,
  `price` decimal(4,2) DEFAULT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_product_in_category` (`name`,`category`) USING BTREE,
  KEY `code` (`code`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4
```

## 表与索引规约

参考阿里的 [p3c](https://github.com/alibaba/p3c) 规范。

### 命名

数据库字段名的修改代价很大，因为无法进行预发布，所以字段名称需要慎重考虑。表名、字段名必须使用小写字母或数字，禁止出现数字开头，禁止两个下划线中间只出现数字。MySQL 在 Windows 下不区分大小写，但在 Linux 下默认是区分大小写；因此，数据库名、表名、字段名，都不允许出现任何大写字母，避免节外生枝。

表名应该仅仅表示表里面的实体内容，不应该表示实体数量，对应于 DO 类名也是单数形式，不使用复数名词，符合表达习惯。单表行数超过 500 万行或者单表容量超过 2GB，才推荐进行分库分表。

### 字段

表必备三字段：id, gmt_create, gmt_modified。其中 id 必为主键，类型为 unsigned bigint、单表时自增、步长为 1。gmt_create, gmt_modified 的类型均为 datetime 类型，前者现在时表示主动创建，后者过去分词表示被动更新。

任何字段如果为非负数，必须是 unsigned。小数类型为 decimal，禁止使用 float 和 double。float 和 double 在存储的时候，存在精度损失的问题，很可能在值的比较时，得到不正确的结果。如果存储的数据范围超过 decimal 的范围，建议将数据拆成整数和小数分开存储。

表达是与否概念的字段，必须使用 is_xxx 的方式命名，数据类型是 unsigned tinyint(1 表示是，0 表示否)。但是 POJO 类的布尔属性不能加 is，要求在 resultMap 中进行字段与属性之间的映射。

### 索引

主键索引名为 `pk_` 字段名；唯一索引名为 `uk_` 字段名；普通索引名则为 `idx_` 字段名。业务上具有唯一特性的字段，即使是多个字段的组合，也必须建成唯一索引。

在 varchar 字段上建立索引时，必须指定索引长度，没必要对全字段建立索引，根据实际文本区分度决定索引长度即可。索引的长度与区分度是一对矛盾体，一般对字符串类型数据，长度为 20 的索引，区分度会高达 90%以上，可以使用`count(distinct left(列名, 索引长度))/count(*)`的区分度来确定。

### 外键

1、要设置外键的字段不能为主键

2、改建所参考的字段必须为主键

3、两个字段必须具有相同的数据类型和约束

满足这三个条件一般在创建外键的时候就不会报错，而这里报错了 cannot add foreign key constraint 大多数是因为第三个条件不满足。

# Database Design | 数据库设计

![image](https://user-images.githubusercontent.com/5803001/46092422-70455c00-c1e7-11e8-80b8-2b1c8520c4ff.png)

# Max Compute/ODPS

MaxCompute SQL 适用于海量数据（GB、TB、EB 级别），离线批量计算的场合。MaxCompute 作业提交后会有几十秒到数分钟不等的排队调度，所以适合处理跑批作业，一次作业批量处理海量数据，不适合直接对接需要每秒处理几千至数万笔事务的前台业务系统。

## DDL

```sql
CREATE [EXTERNAL] TABLE [IF NOT EXISTS] table_name
[(col_name data_type [COMMENT col_comment], ...)]
[COMMENT table_comment]
[PARTITIONED BY (col_name data_type [COMMENT col_comment], ...)]
[STORED BY StorageHandler] -- 仅限外部表
[WITH SERDEPROPERTIES (Options)] -- 仅限外部表
[LOCATION OSSLocation];-- 仅限外部表
[LIFECYCLE days]
[AS select_statement]

create table if not exists sale_detail
(
	shop_name     string,
	customer_id   string,
	total_price   double
)
partitioned by (sale_date string,region string);
-- 创建一张分区表sale_detail

CREATE TABLE [IF NOT EXISTS] table_name
LIKE existing_table_name

create table sale_detail_ctas1 as
select * from sale_detail;
```

Partitioned by 指定表的分区字段，目前支持 Tinyint、Smallint、 Int、 Bigint、Varchar 和 String 类型。分区值不允许有双字节字符（如中文），必须是以英文字母 a-z，A-Z 开始后可跟字母数字，名称的长度不超过 128 字节。当利用分区字段对表进行分区时，新增分区、更新分区内数据和读取分区数据均不需要做全表扫描，可以提高处理效率。

在 create table…as select…语句中，如果在 select 子句中使用常量作为列的值，建议指定列的名字；否则创建的表 sale_detail_ctas3 的第四、五列类似于\_c5、\_c6。

```sql
--- 删除表
DROP TABLE [IF EXISTS] table_name;

--- 重命名表
ALTER TABLE table_name RENAME TO new_table_name;
```

## Select | 查询

### Join

```sql
--- 左连接
select a.shop_name as ashop, b.shop_name as bshop from shop a
        left outer join sale_detail b on a.shop_name=b.shop_name;
    -- 由于表shop及sale_detail中都有shop_name列，因此需要在select子句中使用别名进行区分。

--- 右连接
select a.shop_name as ashop, b.shop_name as bshop from shop a
        right outer join sale_detail b on a.shop_name=b.shop_name;

--- 全连接
select a.shop_name as ashop, b.shop_name as bshop from shop a
        full outer join sale_detail b on a.shop_name=b.shop_name;
```

连接条件，只允许 and 连接的等值条件。只有在 MAPJOIN 中，可以使用不等值连接或者使用 or 连接多个条件。

### Map Join

当一个大表和一个或多个小表做 Join 时，可以使用 MapJoin，性能比普通的 Join 要快很多。MapJoin 的基本原理为：在小数据量情况下，SQL 会将您指定的小表全部加载到执行 Join 操作的程序的内存中，从而加快 Join 的执行速度。

![image](https://user-images.githubusercontent.com/5803001/47965355-15721080-e081-11e8-8e33-ad18258c6d9f.png)

MapJoin 简单说就是在 Map 阶段将小表读入内存，顺序扫描大表完成 Join；以 Hive MapJoin 的原理图为例，可以看出 MapJoin 分为两个阶段：

- 通过 MapReduce Local Task，将小表读入内存，生成 HashTableFiles 上传至 Distributed Cache 中，这里会对 HashTableFiles 进行压缩。

- MapReduce Job 在 Map 阶段，每个 Mapper 从 Distributed Cache 读取 HashTableFiles 到内存中，顺序扫描大表，在 Map 阶段直接进行 Join，将数据传递给下一个 MapReduce 任务。

```sql
select /* + mapjoin(a) */
        a.shop_name,
        b.customer_id,
        b.total_price
    from shop a join sale_detail b
    on a.shop_name = b.shop_name;
```

left outer join 的左表必须是大表，right outer join 的右表必须是大表，inner join 左表或右表均可以作为大表，full outer join 不能使用 MapJoin。

### Subquery | 子查询

在 from 子句中，子查询可以当作一张表来使用，与其它的表或子查询进行 Join 操作，子查询必须要有别名。

```sql
create table shop as select * from sale_detail;

--- 子查询作为表
select a.shop_name, a.customer_id, a.total_price from
(select * from shop) a join sale_detail on a.shop_name = sale_detail.shop_name;

--- IN SUBQUERY / NOT IN SUBQUERY
SELECT * from mytable1 where id in (select id from mytable2);
--- 等效于
SELECT * from mytable1 a LEFT SEMI JOIN mytable2 b on a.id=b.id;

--- EXISTS SUBQUERY/NOT EXISTS SUBQUERY
SELECT * from mytable1 where not exists (select * from mytable2 where id = mytable1.id);
--- 等效于
SELECT * from mytable1 a LEFT ANTI JOIN mytable2 b on a.id=b.id;

--- SCALAR SUBQUERY
select * from t1 where (select count(*)  from t2 where t1.a = t2.a) > 1;
-- 等效于
select t1.* from t1 left semi join (select a, count(*) from t2 group by a having count(*) > 1) t2 on t1 .a = t2.a;
```

## UDF

```java
package org.alidata.odps.udf.examples;

import com.aliyun.odps.udf.UDF;

public final class Lower extends UDF {

  public String evaluate(String s) {
    if (s == null) {
      return null;
    }
    return s.toLowerCase();
  }
}
```

# 数据建模

每个 DBMS，无论是 NoSQL 还是 SQL，最终，都是把无意义的物理状态（高电压和低电压，或者开和关）和有意义的事物建立映射关系，从而表示数据。我们把这个映射称为物理表示。在更高的层次上，我们使用表、图形和文档等结构来表示关系。理解的关键是逻辑数据模型应该完全忽略这些物理映射问题。逻辑数据模型应该把重点完全放在数据的含义上以及数据如何按照逻辑表示问题域内的数据。但是，在从逻辑模型转移到物理模型时，保留从物理模型到逻辑模型的映射关系以及物理表示设计都变得至关重要了。
