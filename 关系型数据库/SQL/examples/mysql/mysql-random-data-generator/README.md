# Random data generator for MySQL
[![Build Status](https://travis-ci.org/Percona-Lab/mysql_random_data_load.svg?branch=master)](https://travis-ci.org/Percona-Lab/mysql_random_data_load)

Many times in my job I need to generate random data for a specific table in order to reproduce an issue.  
After writing many random generators for every table, I decided to write a random data generator, able to get the table structure and generate random data for it.  
Plase take into consideration that this is the first version and it doesn't support all field types yet!  

**NOTICE**  
This is an early stage project.  

## Supported fields:

|Field type|Generated values|
|----------|----------------|
|tinyint|0 ~ 0xFF|
|smallint|0 ~ 0XFFFF|
|mediumint|0 ~ 0xFFFFFF|
|int - integer|0 ~ 0xFFFFFFFF|
|bigint|0 ~ 0xFFFFFFFFFFFFFFFF|
|float|0 ~ 1e8|
|decimal(m,n)|0 ~ 10^(m-n)|
|double|0 ~ 1000|
|char(n)|up to n random chars|
|varchar(n)|up to n random chars|
|date|NOW() - 1 year ~ NOW()|
|datetime|NOW() - 1 year ~ NOW()|
|timestamp|NOW() - 1 year ~ NOW()|
|time|00:00:00 ~ 23:59:59|
|year|Current year - 1 ~ current year|
|tinyblob|up to 100 chars random paragraph|
|tinytext|up to 100 chars random paragraph|
|blob|up to 100 chars random paragraph|
|text|up to 100 chars random paragraph|
|mediumblob|up to 100 chars random paragraph|
|mediumtext|up to 100 chars random paragraph|
|longblob|up to 100 chars random paragraph|
|longtext|up to 100 chars random paragraph|
|varbinary|up to 100 chars random paragraph|
|enum|A random item from the valid items list|
|set|A random item from the valid items list|

### How strings are generated

- If field size < 10 the program generates a random "first name"
- If the field size > 10 and < 30 the program generates a random "full name"
- If the field size > 30 the program generates a "lorem ipsum" paragraph having up to 100 chars.
 
The program can detect if a field accepts NULLs and if it does, it will generate NULLs ramdomly (~ 10 % of the values).

## Usage
`mysql_random_data_load <database> <table> <number of rows> [options...]`

## Options
|Option|Description|
|------|-----------|
|--bulk-size|Number of rows per INSERT statement (Default: 1000)|
|--debug|Show some debug information|
|--fk-samples-factor|Percentage used to get random samples for foreign keys fields. Default 0.3|
|--host|Host name/ip|
|--max-fk-samples|Maximum number of samples for fields having foreign keys constarints. Default: 100|
|--max-retries|Maximum number of rows to retry in case of errors. See duplicated keys. Deafult: 100|
|--no-progressbar|Skip showing the progress bar. Default: false|
|--password|Password|
|--port|Port number|
|--Print|Print queries to the standard output instead of inserting them into the db|
|--user|Username|
|--version|Show version and exit|

## Foreign keys support
If a field has Foreign Keys constraints, `random-data-load` will get up to `--max-fk-samples` random samples from the referenced tables in order to insert valid values for the field.  
The number of samples to get follows this rules:  
**1.** Get the aproximate number of rows in the referenced table using the `rows` field in:  
```
EXPLAIN SELECT COUNT(*) FROM <referenced schema>.<referenced table>
```
**1.1** If the number of rows is less than `max-fk-samples`, all rows are retrieved from the referenced table using this query: 
```
SELECT <referenced field> FROM <referenced schema>.<referenced table>
```
**1.2** If the number of rows is greater than `max-fk-samples`, samples are retrieved from the referenced table using this query:  
```
SELECT <referenced field> FROM <referenced schema>.<referenced table> WHERE RAND() <= <fk-samples-factor> LIMIT <max-fk-samples>
```

### Example
```
CREATE DATABASE IF NOT EXISTS test;

CREATE TABLE `test`.`t3` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `tcol01` tinyint(4) DEFAULT NULL,
  `tcol02` smallint(6) DEFAULT NULL,
  `tcol03` mediumint(9) DEFAULT NULL,
  `tcol04` int(11) DEFAULT NULL,
  `tcol05` bigint(20) DEFAULT NULL,
  `tcol06` float DEFAULT NULL,
  `tcol07` double DEFAULT NULL,
  `tcol08` decimal(10,2) DEFAULT NULL,
  `tcol09` date DEFAULT NULL,
  `tcol10` datetime DEFAULT NULL,
  `tcol11` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `tcol12` time DEFAULT NULL,
  `tcol13` year(4) DEFAULT NULL,
  `tcol14` varchar(100) DEFAULT NULL,
  `tcol15` char(2) DEFAULT NULL,
  `tcol16` blob,
  `tcol17` text,
  `tcol18` mediumtext,
  `tcol19` mediumblob,
  `tcol20` longblob,
  `tcol21` longtext,
  `tcol22` mediumtext,
  `tcol23` varchar(3) DEFAULT NULL,
  `tcol24` varbinary(10) DEFAULT NULL,
  `tcol25` enum('a','b','c') DEFAULT NULL,
  `tcol26` set('red','green','blue') DEFAULT NULL,
  `tcol27` float(5,3) DEFAULT NULL,
  `tcol28` double(4,2) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB;
```
To generate 100K random rows, just run:
```
mysql_random-data-load test t3 100000 --user=root --password=root
```
```
mysql> select * from t3 limit 1\G
*************************** 1. row ***************************
    id: 1
tcol01: 10
tcol02: 173
tcol03: 1700
tcol04: 13498
tcol05: 33239373
tcol06: 44846.4
tcol07: 5300.23
tcol08: 11360967.75
tcol09: 2017-09-04
tcol10: 2016-11-02 23:11:25
tcol11: 2017-03-03 08:11:40
tcol12: 03:19:39
tcol13: 2017
tcol14: repellat maxime nostrum provident maiores ut quo voluptas.
tcol15: Th
tcol16: Walter
tcol17: quo repellat accusamus quidem odi
tcol18: esse laboriosam nobis libero aut dolores e
tcol19: Carlos Willia
tcol20: et nostrum iusto ipsa sunt recusa
tcol21: a accusantium laboriosam voluptas facilis.
tcol22: laudantium quo unde molestiae consequatur magnam.
tcol23: Pet
tcol24: Richard
tcol25: c
tcol26: green
tcol27: 47.430
tcol28: 6.12
1 row in set (0.00 sec)
```

## How to download the precompiled binaries

There are binaries available for each version for Linux and Darwin. You can find compiled binaries for each version in the releases tab:

https://github.com/Percona-Lab/mysql_random_data_load/releases

## To do
- [ ] Add suport for all data types.
- [X] Add supporrt for foreign keys.
- [ ] Support config files to override default values/ranges.
- [ ] Support custom functions via LUA plugins.

## Version history

#### 0.1.10
- Fixed connection parameters for MySQL 5.7 (set driver's AllowNativePasswords: true)

#### 0.1.9
- Added support for bunary and varbinary columns
- By default, read connection params from ${HOME}/.my.cnf

#### 0.1.8 
- Fixed error for triggers created with MySQL 5.6
- Added Travis-CI
- Code clean up

#### 0.1.7 
- Support for MySQL 8.0
- Added --print parameter 
- Added --version parameter
- Removed qps parameter

#### 0.1.6 
- Improved generation speed (up to 50% faster)
- Improved support for TokuDB (Thanks Agustin Gallego)
- Code refactored
- Improved debug logging
- Added Query Per Seconds support (experimental)

#### 0.1.5 
- Fixed handling of NULL collation for index parser

#### 0.1.4
- Fixed handling of time columns
- Improved support of GENERATED columns

#### 0.1.3
- Fixed handling of nulls

#### 0.1.2
- New table parser able to retrieve all the information for fields, indexes and foreign keys constraints.
- Support for foreign keys constraints
- Added some tests

#### 0.1.1
- Fixed random data generation

#### 0.1.0
- Initial version
