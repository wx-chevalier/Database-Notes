# 持久化备份与迁移

| **持久化方式**  | **原理**                                                           | **特点**                                 | **启用方式** |
| --------------- | ------------------------------------------------------------------ | ---------------------------------------- | ------------ |
| RDB(快照持久化) | 当符合快照条件时，会自动将内存中的所有数据进行快照并且存储到硬盘上 | 针对热数据，记录和恢复效率高，恢复不完整 | 默认启用     |
| AOF(追加持久化) | 将发送到服务端的每一条请求都记录下来,并且保存在硬盘的 AOF 文件中   | 针对所有请求，记录和恢复效率低，恢复完整 | 配置文件     |

Redis 提供了多种不同级别的持久化方式:一种是 RDB,另一种是 AOF。

- RDB 持久化可以在指定的时间间隔内生成数据集的时间点快照(point-in-time snapshot)。

- AOF 持久化记录服务器执行的所有写操作命令，并在服务器启动时，通过重新执行这些命令来还原数据集。AOF 文件中的命令全部以 Redis 协议的格式来保存，新命令会被追加到文件的末尾。Redis 还可以在后台对 AOF 文件进行重写(rewrite)，使得 AOF 文件的体积不会超出保存数据集状态所需的实际大小。Redis 还可以同时使用 AOF 持久化和 RDB 持久化。在这种情况下，当 Redis 重启时，它会优先使用 AOF 文件来还原数据集，因为 AOF 文件保存的数据集通常比 RDB 文件所保存的数据集更完整。你甚至可以关闭持久化功能，让数据只在服务器运行时存在。

一般来说,如果想达到足以媲美 PostgreSQL 的数据安全性，你应该同时使用两种持久化功能。如果你非常关心你的数据,但仍然可以承受数分钟以内的数据丢失，那么你可以只使用 RDB 持久化。有很多用户都只使用 AOF 持久化，但我们并不推荐这种方式: 因为定时生成 RDB 快照(snapshot)非常便于进行数据库备份，并且 RDB 恢复数据集的速度也要比 AOF 恢复的速度要快，除此之外，使用 RDB 还可以避免之前提到的 AOF 程序的 bug。因为以上提到的种种原因，未来我们可能会将 AOF 和 RDB 整合成单个持久化模型。(这是一个长期计划。)

```
# The filename where to dump the DB
dbfilename dump.rdb

# The working directory.
#
# The DB will be written inside this directory, with the filename specified
# above using the 'dbfilename' configuration directive.
#
# Also the Append Only File will be created inside this directory.
#
# Note that you must specify a directory here, not a file name.
dir /var/lib/redis
```

# Links

- https://cubox.pro/c/uACwYC Redis 专题：万字长文详解持久化原理
