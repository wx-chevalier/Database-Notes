# 数据操作

# Data Manipulation Language | 数据操作

DML 包含了 INSERT, UPDATE, DELETE 等常见的数据操作语句。

## Update | 更新

### 存在性更新

我们经常需要处理某个唯一索引时存在则更新，不存在则插入的情况，其基本形式如下：

```sql
INSERT INTO ... ON DUPLICATE KEY UPDATE ...
```

对于多属性索引的更新方式如下：

```sql
/* 创建语句中添加索引描述 */
UNIQUE INDEX `index_var` (`var1`, `var2`, `var3`)

/* 同时更新索引包含的多属性域值 */
INSERT INTO `test_table`
(`var1`, `var2`, `var3`, `value1`, `value2`, `value3`) VALUES
('abcd', 0, 'xyz', 1, 2, 3)
ON DUPLICATE KEY UPDATE `value1` = `value1` + 1 AND
`value2` = `value2` + 2 AND `value3` = `value3` + 3;
```

## DELETE

Delete only the deadline rows:
sql

```sql
DELETE `deadline` FROM `deadline` LEFT JOIN `job` ....
```

Delete the deadline and job rows:

```sql
DELETE `deadline`, `job` FROM `deadline` LEFT JOIN `job` ....
```

Delete only the job rows:

```sqk
DELETE `job` FROM `deadline` LEFT JOIN `job` ....
```

删除某个表中的重复数据：

```sql
DELETE product
FROM
	product
LEFT JOIN (
	SELECT
		count(*) AS cnt,
		id
	FROM
		product
	GROUP BY
		id
) a ON a.id = product.id
WHERE
	a.cnt > 1;
```
