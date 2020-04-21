# 数据库设计

- Preserve everything we know when we split our information into tables.

- Reduce the redundancy of how we store it. Regarding the second point: do we only want to reduce redundancy because of a storage size problem? No, we do it mainly because having redundant data leads you to inconsistency problems if you don’t update all the fields that represent the same information when an update occurs

# Data Definition Language | 数据定义

DDL 包含 CREATE, ALTER, DROP 等常见的数据定义语句，这里可以查阅[完整的表结构 SQL](https://gist.github.com/wx-chevalier/ebd1ceb919a68e428e7901f7fc766f02)。

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
