# PTSDB

PTSDB 的核心包括：倒排索引+窗口存储 Block。数据的写入按照两个小时为一个时间窗口，将两小时内产生的数据存储在一个 Head Block 中，每一个块中包含该时间窗口内的所有样本数据(chunks)，元数据文件(meta.json)以及索引文件(index)。

最新写入数据保存在内存 block 中，2 小时后写入磁盘。后台线程把 2 小时的数据最终合并成更大的数据块，一般的数据库在固定一个内存大小后，系统的写入和读取性能会受限于这个配置的内存大小。而 PTSDB 的内存大小是由最小时间周期，采集周期以及时间线数量来决定的。为防止内存数据丢失，实现 wal 机制。删除记录在独立的 tombstone 文件中。

# 存储引擎

PTSDB 的核心数据结构是 HeadAppender，Appender commit 时 wal 日志编码落盘，同时写入 head block 中。

![Head Appender](https://s2.ax1x.com/2019/11/20/MW5ui8.png)

PTSDB 本地存储使用自定义的文件结构。主要包含：WAL，元数据文件，索引，chunks，tombstones。

![本地文件结构](https://s2.ax1x.com/2019/11/20/MW53ss.png)

## 乱序处理

PTSDB 对于乱序的处理采用了最小时间窗口的方式，指定合法的最小时间戳，小于这一时间戳的数据会丢弃不再处理。
合法最小时间戳取决于当前 head block 里面最早的时间戳和可存储的 chunk 范围。
这种对于数据行为的限定极大的简化了设计的灵活性，对于 compaction 的高效处理以及数据完整性提供了基础。

## 内存的管理

使用 mmap 读取压缩合并后的大文件（不占用太多句柄），
建立进程虚拟地址和文件偏移的映射关系，只有在查询读取对应的位置时才将数据真正读到物理内存。
绕过文件系统 page cache，减少了一次数据拷贝。
查询结束后，对应内存由 Linux 系统根据内存压力情况自动进行回收，在回收之前可用于下一次查询命中。
因此使用 mmap 自动管理查询所需的的内存缓存，具有管理简单，处理高效的优势。

## Compaction

Compaction 主要操作包括合并 block、删除过期数据、重构 chunk 数据。

合并多个 block 成为更大的 block，可以有效减少 block 个，当查询覆盖的时间范围较长时，避免需要合并很多 block 的查询结果。
为提高删除效率，删除时序数据时，会记录删除的位置，只有 block 所有数据都需要删除时，才将 block 整个目录删除。
block 合并的大小也需要进行限制，避免保留了过多已删除空间(额外的空间占用)。
比较好的方法是根据数据保留时长，按百分比（如 10%）计算 block 的最大时长, 当 block 的最小和最大时长超过 2/3blok 范围时，执行 compaction

## 快照

PTSDB 提供了快照备份数据的功能，用户通过 admin/snapshot 协议可以生成快照，快照数据存储于 data/snapshots/-目录。

# 存储格式

## Write Ahead Log

WAL 有 3 种编码格式：时间线，数据点，以及删除点。总体策略是基于文件大小滚动，并且根据最小内存时间执行清除。

- 当日志写入时，以 segment 为单位存储，每个 segment 默认 128M, 记录数大小达到 32KB 页时刷新一次。当剩余空间小于新的记录数大小时，创建新的 Segment。

- 当 compation 时 WAL 基于时间执行清除策略，小于内存中 block 的最小时间的 wal 日志会被删除。

- 重启时，首先打开最新的 Segment，从日志中恢复加载数据到内存。

![WAL 文件结构](https://s2.ax1x.com/2019/11/20/MW5dWF.png)

## 元数据文件

meta.json 文件记录了 Chunks 的具体信息, 比如新的 compactin chunk 来自哪几个小的 chunk。这个 chunk 的统计信息，比如：最小最大时间范围，时间线，数据点个数等等。

compaction 线程根据统计信息判断该 blocks 是否可以做 compact：（maxTime-minTime）占整体压缩时间范围的 50%，删除的时间线数量占总体数量的 5%。

![元数据文件](https://s2.ax1x.com/2019/11/20/MW55yd.png)

## 索引

索引一部分先写入 Head Block 中，随着 compaction 的触发落盘。
索引采用的是倒排的方式，posting list 里面的 id 是局部自增的，作为 reference id 表示时间线。索引 compact 时分为 6 步完成索引的落盘:Symbols->Series->LabelIndex->Posting->OffsetTable->TOC

- Symbols 存储的是 tagk, tagv 按照字母序递增的字符串表。比如**name**,go_gc_duration_seconds, instance, localhost:9090 等等。字符串按照 utf8 统一编码。
- Series 存储了两部分信息，一部分是标签键值对的符号表引用；另外一部分是时间线到数据文件的索引，按照时间窗口切割存储数据块记录的具体位置信息，因此在查询时可以快速跳过大量非查询窗口的记录数据，
  为了节省空间，时间戳范围和数据块的位置信息的存储采用差值编码。
- LabelIndex 存储标签键以及每一个标签键对应的所有标签值，当然具体存储的数据也是符号表里面的引用值。
- Posting 存储倒排的每个 label 对所对应的 posting refid
- OffsetTable 加速查找做的一层映射，将这部分数据加载到内存。OffsetTable 主要关联了 LabelIndex 和 Posting 数据块。TOC 是各个数据块部分的位置偏移量，如果没有数据就可以跳过查找。

![索引文件格式](https://s2.ax1x.com/2019/11/20/MWIpmn.png)

## Chunks

数据点存放在 chunks 目录下，每个 data 默认 512M，数据的编码方式支持 XOR，chunk 按照 refid 来索引，refid 由 segmentid 和文件内部偏移量两个部分组成。

![Chunks 结构](https://s2.ax1x.com/2019/11/20/MWInmR.png)

## Tombstones

记录删除通过 mark 的方式，数据的物理清除发生在 compaction 和 reload 的时候。以时间窗口为单位存储被删除记录的信息。

![Tombstones 结构](https://s2.ax1x.com/2019/11/20/MWIQk6.png)

# 最佳实践

在一般情况下，Prometheus 中存储的每一个样本大概占用 1-2 字节大小。如果需要对 Prometheus Server 的本地磁盘空间做容量规划时，可以通过以下公式计算：
needed*disk_space = retention_time_seconds * ingested*samples_per_second * bytes_per_sample
保留时间(retention_time_seconds)和样本大小(bytes_per_sample)不变的情况下，如果想减少本地磁盘的容量需求，
只能通过减少每秒获取样本数(ingested_samples_per_second)的方式。

因此有两种手段，一是减少时间序列的数量，二是增加采集样本的时间间隔。
考虑到 Prometheus 会对时间序列进行压缩，因此减少时间序列的数量效果更明显。
PTSDB 的限制在于集群和复制。因此当一个 node 宕机时，会导致一定窗口的数据丢失。当然，如果业务要求的数据可靠性不是特别苛刻，本地盘也可以存储几年的持久化数据。当 PTSDB Corruption 时，可以通过移除磁盘目录或者某个时间窗口的目录恢复。
PTSDB 的高可用，集群和历史数据的保存可以借助于外部解决方案，不在本文讨论范围。

历史方案的局限性，PTSDB 在早期采用的是单条时间线一个文件的存储方式。这中方案有非常多的弊端，比如：
Snapshot 的刷盘压力：定期清理文件的负担；低基数和长周期查询查询，需要打开大量文件；时间线膨胀可能导致 inode 耗尽。
