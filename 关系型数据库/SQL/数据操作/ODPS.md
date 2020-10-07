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

Partitioned by 指定表的分区字段，目前支持 Tinyint、Smallint、Int、Bigint、Varchar 和 String 类型。分区值不允许有双字节字符（如中文），必须是以英文字母 a-z，A-Z 开始后可跟字母数字，名称的长度不超过 128 字节。当利用分区字段对表进行分区时，新增分区、更新分区内数据和读取分区数据均不需要做全表扫描，可以提高处理效率。

在 create table…as select…语句中，如果在 select 子句中使用常量作为列的值，建议指定列的名字；否则创建的表 sale_detail_ctas3 的第四、五列类似于 `_c5`、`_c6`。

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
