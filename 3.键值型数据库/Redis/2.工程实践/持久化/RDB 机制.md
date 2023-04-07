# RDB 机制

RDB 机制是以指定的时间间隔将 Redis 中的数据生成快照并保存到硬盘中，它更适合于定时备份数据的应用场景。可以通过手动或者自动的方式来触发 RDB 机制：

### 2.1 手动触发

可以通过以下两种方式来手动触发 RDB 机制：

- **save** ：save 命令会阻塞当前 Redis 服务，直到 RDB 备份过程完成，在这个时间内，客户端的所有查询都会被阻塞；
- **bgsave** ：Redis 进程会 fork 出一个子进程，阻塞只会发生在 fork 阶段，之后持久化的操作则由子进程来完成。

### 2.2 自动触发

除了手动使用命令触发外，在某些场景下也会自动触发 Redis 的 RDB 机制：

- 在 `redis.conf` 中配置了 `save m n` ，表示如果在 m 秒内存在了 n 次修改操作时，则自动触发 `bgsave`;
- 如果从节点执行全量复制操作，则主节点自动执行 `bgsave`，并将生成的 RDB 文件发送给从节点；
- 执行 `debug reload` 命令重新加载 Redis 时，会触发 `save` 操作；
- 执行 `shutdown` 命令时候，如果没有启用 AOF 持久化则默认采用 `bgsave ` 进行持久化。

### 2.3 相关配置

**1. 文件目录**

RDB 文件默认保存在 Redis 的工作目录下，默认文件名为 `dump.rdb`，可以通过静态或动态方式修改：

- 静态配置：通过修改 `redis.conf` 中的工作目录 `dir` 和数据库存储文件名 `dbfilename` 两个配置；

- 动态修改：通过在命令行中执行以下命令：

  ```shell
  config set dir{newDir}
  config set dbfilename{newFileName}
  ```

**2. 压缩算法**

Redis 默认采用 LZF 算法对生成的 RDB 文件做压缩处理， 这样可以减少占用空间和网络传输的数据量，但是压缩过程会耗费 CPU 的计算资源， 你可以按照实际情况，选择是否启用。可以通过修改 `redis.conf` 中的 `rdbcompression` 配置或使用以下命令来进行动态修改：

```shell
config set rdbcompression{yes|no}
```
