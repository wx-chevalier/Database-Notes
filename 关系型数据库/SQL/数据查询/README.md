# 数据查询

完整的 SQL 查询语句语法如下：

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

# SELECT

SELECT 语句是我们最常用的用于数据查询的语句：

```sql
SELECT 42;
SELECT 'hello world!';
SELECT 'We are just getting started.';

-- SQL 也支持简单的数学表达式
SELECT 2 + 3;
SELECT 5 * 12;
SELECT 164 / 8;

-- 字符串连接
SELECT 'Dave' || ' is a fantastic SQL instructor';
SELECT DATE 'Mar 03, 2020';
```

## FROM

```sql
SELECT [stuff you want to select] FROM [the table that it is in];

SELECT id, title, artist_id FROM albums;
```

## ORDER BY

默认情况下，结果按存储在数据库中的顺序返回。但是有时您会希望对它们进行不同的排序。您可以在查询结束时使用 ORDER BY 命令执行此操作，如此处的 SQL 模板的扩展版本所示：

```sql
SELECT [stuff you want to select] FROM [the table that it is in] ORDER BY [column you want to order by];
```

例如，以下查询显示了 album_id 排序的所有曲目。尝试按其他列对其进行排序。

```sql
SELECT * FROM tracks ORDER BY album_id;
```

您可以将多个事物列出到 ORDER BY 中，这在重复行很多的情况下很有用。例如，在曲目中，我们可以通过作曲家对所有数据进行排序，然后通过列出这两个排序列来排序歌曲的长度（毫秒）。

```sql
SELECT * FROM tracks ORDER BY composer, milliseconds;
```

默认情况下，事物按升序排序。您可以通过指定 DESC（降序）来选择颠倒顺序。同样，如果您想指定要 ASCending，请使用 ASC。

```sql
SELECT * FROM tracks ORDER BY name DESC;
```
