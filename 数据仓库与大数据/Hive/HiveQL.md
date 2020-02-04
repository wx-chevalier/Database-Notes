> - [Hive 四种数据导入方式](http://blog.csdn.net/lifuxiangcaohui/article/details/40588929)

# Index:索引

Hive 的数据分为表数据和元数据，表数据是 Hive 中表格(table)具有的数据；而元数据是用来存储表的名字，表的列和分区及其属性，表的属性(是否为外部表等)，表的数据所在目录等。下面分别来介绍。

索引是标准的数据库技术，hive 0.7 版本之后支持索引。Hive 提供有限的索引功能，这不像传统的关系型数据库那样有“键(key)”的概念，用户可以在某些列上创建索引来加速某些操作，给一个表创建的索引数据被保存在另外的表中。Hive 的索引功能现在还相对较晚，提供的选项还较少。但是，索引被设计为可使用内置的可插拔的 java 代码来定制，用户可以扩展这个功能来满足自己的需求。当然不是说有的查询都会受惠于 Hive 索引。用户可以使用 EXPLAIN 语法来分析 HiveQL 语句是否可以使用索引来提升用户查询的性能。像 RDBMS 中的索引一样，需要评估索引创建的是否合理，毕竟，索引需要更多的磁盘空间，并且创建维护索引也会有一定的代价。用户必须要权衡从索引得到的好处和代价。
　　下面说说怎么创建索引：
　　 1、先创建表：

1
2
3
4
hive> create table user( id int, name string)  
 > ROW FORMAT DELIMITED  
 > FIELDS TERMINATED BY '\t' > STORED AS TEXTFILE;
　　 2、导入数据：

1
2
hive> load data local inpath '/export1/tmp/wyp/row.txt' > overwrite into table user;
　　 3、创建索引之前测试

01
02
03
04
05
06
07
08
09
10
11
12
13
14
15
16
17
18
19
20
21
22
23
24
25
hive> select \* from user where id =500000;
Total MapReduce jobs = 1
Launching Job 1 out of 1
Number of reduce tasks is set to 0 since there's no reduce operator
Cannot run job locally: Input Size (= 356888890) is larger than
hive.exec.mode.local.auto.inputbytes.max (= 134217728)
Starting Job = job_1384246387966_0247, Tracking URL =

http://l-datalogm1.data.cn1:9981/proxy/application_1384246387966_0247/

Kill Command=/home/q/hadoop/bin/hadoop job -kill job_1384246387966_0247
Hadoop job information for Stage-1: number of mappers:2; number of reducers:0
2013-11-13 15:09:53,336 Stage-1 map = 0%, reduce = 0%
2013-11-13 15:09:59,500 Stage-1 map=50%,reduce=0%, Cumulative CPU 2.0 sec
2013-11-13 15:10:00,531 Stage-1 map=100%,reduce=0%, Cumulative CPU 5.63 sec
2013-11-13 15:10:01,560 Stage-1 map=100%,reduce=0%, Cumulative CPU 5.63 sec
MapReduce Total cumulative CPU time: 5 seconds 630 msec
Ended Job = job_1384246387966_0247
MapReduce Jobs Launched:
Job 0: Map: 2 Cumulative CPU: 5.63 sec  
HDFS Read: 361084006 HDFS Write: 357 SUCCESS
Total MapReduce CPU Time Spent: 5 seconds 630 msec
OK
500000 wyp.
Time taken: 14.107 seconds, Fetched: 1 row(s)
一共用了 14.107s
　　 4、对 user 创建索引

01
02
03
04
05
06
07
08
09
10
11
12
hive> create index user_index on table user(id) > as 'org.apache.hadoop.hive.ql.index.compact.CompactIndexHandler' > with deferred rebuild > IN TABLE user_index_table;
hive> alter index user_index on user rebuild;
hive> select \* from user_index_table limit 5;
0 hdfs://mycluster/user/hive/warehouse/table02/000000_0 [0]
1 hdfs://mycluster/user/hive/warehouse/table02/000000_0 [352]
2 hdfs://mycluster/user/hive/warehouse/table02/000000_0 [704]
3 hdfs://mycluster/user/hive/warehouse/table02/000000_0 [1056]
4 hdfs://mycluster/user/hive/warehouse/table02/000000_0 [1408]
Time taken: 0.244 seconds, Fetched: 5 row(s)
这样就对 user 表创建好了一个索引。

在 Hive 创建索引还存在 bug：如果表格的模式信息来自 SerDe，Hive 将不能创建索引：
hive> CREATE INDEX employees_index > ON TABLE employees (country) > AS 'org.apache.hadoop.hive.ql.index.compact.CompactIndexHandler' > WITH DEFERRED REBUILD > IDXPROPERTIES ('creator' = 'me','created_at' = 'some_time') > IN TABLE employees_index_table > COMMENT 'Employees indexed by country and name.';
FAILED: Error in metadata: java.lang.RuntimeException: \
Check the index columns, they should appear in the table being indexed.
FAILED: Execution Error, return code 1 from \
org.apache.hadoop.hive.ql.exec.DDLTask
这个 bug 发生在 Hive0.10.0、0.10.1、0.11.0，在 Hive0.12.0 已经修复了，详情请参见：https://issues.apache.org/jira/browse/HIVE-4251

> - [Hive 查询进阶](http://blog.csdn.net/lifuxiangcaohui/article/details/41548433)
> - [Hive 中分组取前 N 个值](http://blog.csdn.net/lifuxiangcaohui/article/details/41548667)
> - [某个 Hive 查询实例，理清 Hive 的应用思路](http://www.360doc.com/content/14/0107/20/15109633_343417196.shtml)

# Insert

1、insert into 语句
hive> insert into table userinfos2 select id,age,name from userinfos;

2、insert overwrite 语句
hive> insert overwrite table userinfos2 select id,age,name from userinfos;
insert overwrite 会覆盖已经存在的数据，我们假设要插入的数据和已经存在的 N 条数据一样，那么插入后只会保留一条数据；
insert into 只是简单的 copy 插入，不做重复性校验，如果插入前有 N 条数据和要插入的数据一样，那么插入后会有 N+1 条数据；

在 Hive0.8 开始支持 Insert into 语句，它的作用是在一个表格里面追加数据。

标准语法语法如下：

1
2
3
4
5
6
7
8
9
用法一：
INSERT OVERWRITE TABLE tablename1 [PARTITION \
(partcol1=val1, partcol2=val2 ...) [IF NOT EXISTS]] \
select_statement1 FROM from_statement;

用法二：
INSERT INTO TABLE tablename1 [PARTITION \
(partcol1=val1, partcol2=val2 ...)] \
select_statement1 FROM from_statement;
注意：上面语句由于太长了，为了页面显示美观，用’\'符号换行了。
举例：

1
2
hive> insert into table cite
　　> select \* from tt;
这样就会将 tt 表格里面的数据追加到 cite 表格里面。并且在 cite 数据存放目录生成了一个新的数据文件，这个新文件是经过处理的，列之间的分割是 cite 表格的列分割符，而不是 tt 表格列的分隔符。
　　(1)、如果两个表格的维度不一样，将会插入错误：

1
2
3
4
5
6
hive> insert into table cite > select \* from cite_standby;

FAILED: SemanticException [Error 10044]: Line 1:18 Cannot insert into
target table because column number/types are different 'cite':
Table insclause-0 has 2 columns, but query has 1 columns.
从上面错误提示看出，查询的表格 cite_standby 只有一列，而目标表格(也就是需要插入数据的表格)有 2 列，由于列的数目不一样，导致了上面的语句不能成功运行，我们需要保证查询结果列的数目和需要插入数据表格的列数目一致，这样才行。
　　(2)、在用 extended 关键字创建的表格上插入数据将会影响到其它的表格的数据，因为他们共享一份数据文件。
　　(3)、如果查询出来的数据类型和插入表格对应的列数据类型不一致，将会进行转换，但是不能保证转换一定成功，比如如果查询出来的数据类型为 int，插入表格对应的列类型为 string，可以通过转换将 int 类型转换为 string 类型；但是如果查询出来的数据类型为 string，插入表格对应的列类型为 int，转换过程可能出现错误，因为字母就不可以转换为 int，转换失败的数据将会为 NULL。
　　(4)、可以将一个表查询出来的数据插入到原表中：

1
2
hive> insert into table cite  
　　> select \* from cite;
　　结果就是相当于复制了一份 cite 表格中的数据。
　　(5)、和 insert overwrite 的区别：

1
2
hive> insert overwrite table cite  
　　> select \* from tt;
　　上面的语句将会用 tt 表格查询到的数据覆盖 cite 表格已经存在的数据。
