# 线上运维

本节主要讨论 MySQL 日常使用中存在的异常案例。

# 线上 DDL 导致的数据丢失

user 表增加两个列、两个 UK 索引、两个普通索引，SQL 如下：

```sql
ALTER TABLE `user`
ADD COLUMN `user_account_id` bigint NOT NULL DEFAULT 0 COMMENT '用户账户ID 全局',
ADD COLUMN `state_code` int NOT NULL DEFAULT 86 COMMENT '国家地区编码',
ADD UNIQUE KEY `uk_crew_mobile` (`crew_id`,`state_code`,`user_mobile`,`deletion_flag`),
ADD UNIQUE KEY `uk_user_account_id` (`crew_id`,`user_account_id`,`deletion_flag`),
ADD KEY `idx_user_account_id` (`user_account_id`);
```

执行结果发现现有数据库中违反规则的数据都被删除了。在 binlog 中可以发现如下的踪迹：

```sh
#191025 15:00:40 server id 199514850 end_log_pos 290 Query thread_id=2513873 exec_time=1 error_code=0

SET TIMESTAMP=1571986840/*!*/;

/* query from idb-toolkit */ /* rename-8883761-2513873 */RENAME TABLE `crew`.`user` to `crew`.`tp_8883761_del_user`, `crew`.`tp_8883761_ogt_user` to `crew`.`user`
```

早在 mysql 5.6 推出以前，执行 DDL ALTER TABLE 变更会锁表，如果数据量很大的情况下，会直接导致业务不可用。mysql5.6 开始，引入 onlineDDL，并不会直接执行变更语句 alter table user add xx，大致流程为：

- 创建临时表 create table tp_xxxx_ogt_user
- 临时表执行结构变更 alter table xxx add xxx
- 把原表中数据导入到临时表
- 变更表名 rename table user to del_user, tp_xxxx_ogt_user to user
- 删除原表

可以猜到在执行步骤 3 的时候，会出现 Duplicate entry ‘xxxx’ for key 异常，整个过程原本应该是在一个事务，步骤 3 执行失败操作会整体回滚。在无锁结构变更模式下，会优先判断 DDL 是否锁表，如果不锁表将使用原生 DDL。上面的 SQL 探测结果为可以使用原生，但执行过程中报错。

然而原生的 onlineDDL 存在一个执行限制：变更过程中产生的 DML 如果有回滚操作，将会整体失败，详看 https://dev.mysql.com/doc/refman/5.6/en/innodb-online-ddl-limitations.html 当这个限制生效时，抛出的错误也是 Duplicate entry for key。无锁结构变更为了提高变更成功率，在容错能力中单独处理了这个错误：当原生 onlineDDL 报 Duplicate entry 错误，同时 SQL 本身不包含 UK 的增加，此时切换到无锁结构变更模式即可完成变更。

问题就出在判断 SQL 是否有 add UK 的操作上。SQL 解析模块的 BUG 导致检测 SQL 是否包含 add UK 操作有误，从而走到了无锁结构变更上。无锁结构变更在添加 UK 时丢弃重复数据是已知问题，该问题无法避免，只能提前检测和绕开。解析模块顺序解析 ALTER 语句中的索引类型，以上 SQL 在解析到 uk_crew_mobile 时已标记为 add uk 操作，但继续解析时在 idx_user_account_id 被判定为非 UK 变更，导致整体误判。

根本原因是原本 SQL 解析到 add uk_crew_mobile 后，继续解析 idx_user_account_id 被误判为无锁变更，忽略了 Duplicate entry 错误，最终导致违反约束的数据没有插入成功，在表象上来看就是数据被 “删除”。
