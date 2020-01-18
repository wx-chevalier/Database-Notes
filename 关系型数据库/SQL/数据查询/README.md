# 数据查询

```sql
SELECT
    [ALL | DISTINCT | DISTINCTROW ]
      [HIGH_PRIORITY]
      [STRAIGHT_JOIN]
      [SQL_SMALL_RESULT] [SQL_BIG_RESULT] [SQL_BUFFER_RESULT]
      [SQL_CACHE | SQL_NO_CACHE] [SQL_CALC_FOUND_ROWS]
    select_expr [, select_expr ...]
    [FROM table_references
      [PARTITION partition_list]
    [WHERE where_condition]
    [GROUP BY {col_name | expr | position}
      [ASC | DESC], ... [WITH ROLLUP]]
    [HAVING where_condition]
    [WINDOW window_name AS (window_spec)
        [, window_name AS (window_spec)] ...]
    [ORDER BY {col_name | expr | position}
      [ASC | DESC], ...]
    [LIMIT {[offset,] row_count | row_count OFFSET offset}]
    [INTO OUTFILE 'file_name'
        [CHARACTER SET charset_name]
        export_options
      | INTO DUMPFILE 'file_name'
      | INTO var_name [, var_name]]
    [FOR {UPDATE | SHARE} [OF tbl_name [, tbl_name] ...] [NOWAIT | SKIP LOCKED]
      | LOCK IN SHARE MODE]]
```

# DQL | 数据查询

## Column | 查询列

```sql
CASE expression
    WHEN condition1 THEN result1
    WHEN condition2 THEN result2
   ...
    WHEN conditionN THEN resultN
    ELSE result
END as field_name
```

## Where | 条件查询

```sql
expression IS NOT NULL

SELECT *
FROM contacts
WHERE last_name IS NOT NULL;
```

页面搜索严禁左模糊或者全模糊，如果需要请走搜索引擎来解决；索引文件具有 B-Tree 的最左前缀匹配特性，如果左边的值未确定，那么无法使用此索引。

### 分页查询

```sql
SELECT column1, column2, ...
FROM table_name
LIMIT offset, count;

SELECT *
FROM yourtable
ORDER BY id
LIMIT 100, 20
```

```sql
SELECT *
FROM yourtable
WHERE id > 234374
ORDER BY id
LIMIT 20
```

对于超多分页的场景，利用延迟关联或者子查询优化；MySQL 并不是跳过 offset 行，而是取 offset+N 行，然后返回放弃前 offset 行，返回 N 行，那当 offset 特别大的时候，效率就非常的低下，要么控制返回的总页数，要么对超过特定阈值的页数进行 SQL 改写。

```sql
SELECT a.* FROM 表1 a, (select id from 表1 where 条件 LIMIT 100000,20 ) b where a.id=b.id
```

## 统计查询

不要使用 count(列名)或 count(常量)来替代 count()，count()是 SQL92 定义的标准统计行数的语法，跟数据库无关，跟 NULL 和非 NULL 无关。`count(*)` 会统计值为 NULL 的行，而 count(列名)不会统计此列为 NULL 值的行。

count(distinct col) 计算该列除 NULL 之外的不重复行数，注意 count(distinct col1, col2) 如果其中一列全为 NULL，那么即使另一列有不同的值，也返回为 0。

## Join | 表联接

表联接最常见的即是出现在查询模型中，但是实际的用法绝不会局限在查询模型中。较常见的联接查询包括了以下几种类型：Inner Join  / Outer Join / Full Join / Cross Join。

值得一提的是，超过三个表禁止 join，需要 join 的字段，数据类型必须绝对一致；多表关联查询时，保证被关联的字段需要有索引。

关于 A left join B on condition 的提醒。
ON 条件：用于决定如何从 表 B 中检索行，如果表 B 中没有任何数据匹配 ON 条件，则会额外生成一行全部为 NULL 的外部行。
WHERE 条件：在匹配阶段，where 条件不会被使用到。仅在匹配阶段完成后，where 子句才会被使用。它将从匹配产生的结果中检索过滤。

![image](https://user-images.githubusercontent.com/5803001/51289337-b8a8b400-1a3a-11e9-942e-c48d4e3d80a4.png)

### Inner Join | 内联查询

Inner Join 是最常用的 Join 类型，基于一个或多个公共字段把记录匹配到一起。Inner Join 只返回进行联结字段上匹配的记录。  如：

```sql
select * from Products inner join Categories on Products.categoryID=Categories.CategoryID
```

以上语句，只返回物品表中的种类 ID 与种类表中的 ID 相匹配的记录数。这样的语句就相当于：

```sql
select * from Products, Categories where Products.CategoryID=Categories.CategoryID
```

Inner Join 是在做排除操作，任一行在两个表中不匹配，注定将从结果集中除掉。(我想，相当于两个集合中取其两者的交集，这个交集的条件就是 on 后面的限定)还要注意的是，不仅能对两个表作联结，可以把一个表与其自身进行联结。

## Outer Join | 外联接

Outer Join 包含了 Left Outer Join 与 Right Outer Join. 其实简写可以写成 Left Join 与 Right Join。left join，right join 要理解并区分左表和右表的概念，A 可以看成左表,B 可以看成右表。left join 是以左表为准的.,左表(A)的记录将会全部表示出来,而右表(B)只会显示符合搜索条件的记录(例子中为: A.aID = B.bID).B 表记录不足的地方均为 NULL。right join 和 left join 的结果刚好相反,这次是以右表(B)为基础的,A 表不足的地方用 NULL 填充。

## Full Join | 全连接

Full Join 相当于把 Left 和 Right 联结到一起，告诉数据库要全部包含左右两侧所有的行，相当于做集合中的并集操作。

## Cross Join | 笛卡尔积

与其它的 JOIN 不同在于，它没有 ON 操作符，它将 JOIN 一侧的表中每一条记录与另一侧表中的所有记录联结起来，得到的是两侧表中所有记录的笛卡儿积。

## 子查询

子查询本质上是嵌套进其他 SELECT, UPDATE, INSERT, DELETE 语句的一个被限制的 SELECT 语句,在子查询中，只有下面几个子句可以使用：

- SELECT 子句(必须)
- FROM 子句(必选)
- WHERE 子句(可选)
- GROUP BY(可选)
- HAVING(可选)

子查询也可以嵌套在其他子查询中，子查询也叫内部查询(Inner query)或者内部选择(Inner Select),而包含子查询的查询语句也叫做外部查询(Outter)或者外部选择(Outer Select)。

### 子查询作为数据源使用

当子查询在外部查询的 FROM 子句之后使用时,子查询被当作一个数据源使用,即使这时子查询只返回一个单一值(Scalar)或是一列值(Column)，在这里依然可以看作一个特殊的数据源,即一个二维数据表(Table)。作为数据源使用的子查询很像一个视图(View),只是这个子查询只是临时存在，并不包含在数据库中。

### 子查询作为选择条件使用

作为选择条件的子查询也是子查询相对最复杂的应用。作为选择条件的子查询是那些只返回一列(Column)的子查询，如果作为选择条件使用，即使只返回单个值，也可以看作是只有一行的一列。譬如我们需要查询价格高于某个指定产品的所有其余产品信息：

```sql
SELECT
	*
FROM
	product
WHERE
	price > (
		SELECT
			price
		FROM
			product
		WHERE
			NAME = "产品一"
	)
```

### 子查询作为计算列使用

当子查询作为计算列使用时，只返回单个值(Scalar)，其用在 SELECT 语句之后，作为计算列使用，同样分为相关子查询和无关子查询。

```sql
--- 查询每个类别中价格大于某个值的产品数目
SELECT
	p1.category,
	(
		SELECT
			count(*)
		FROM
			product p2
		WHERE
			p2.category = p1.category
		AND p2.price > 30
	) AS 'Expensive'
FROM
	product p1
GROUP BY
	p1.category;
```

```sql
--- 自连接查询不同等级的数目
SELECT a.distributor_id,
      (SELECT COUNT(*) FROM my_table WHERE level='personal' and distributor_id = a.distributor_id) as personal_count,
      (SELECT COUNT(*) FROM my_table WHERE level='exec' and distributor_id = a.distributor_id) as exec_count,
      (SELECT COUNT(*) FROM my_table WHERE distributor_id = a.distributor_id) as total_count
FROM my_table a ;
```
