# KeyDB

https://s2.ax1x.com/2019/10/08/ufeL6J.png

https://s2.ax1x.com/2019/10/08/ufeql4.png

For the first charts comparing Redis and KeyDB, the following commands were used:

Memtier: memtier_benchmark -s <'ip of test instance> -p 6379 –hide-histogram --authenticate <'yourpassword> --threads 32 –data-size <size of test ranging 8-16384>

KeyDB: keydb-server --port 6379 --requirepass <'yourpassword> --server-threads 7 --server-thread-affinity true

Redis: redis-server --port 6379 --requirepass <'yourpassword>

For the chart comparing KeyDB ops/sec vs #threads enabled: Memtier: memtier_benchmark -s <'ip of test instance> -p 6379 --hide-histogram --authenticate --threads 32 --data-size 32

KeyDB pinned: keydb-server --port 6379 --requirepass <'yourpassword> --server-threads <#threads used for test> --server-thread-affinity true

KeyDB unpinned: keydb-server --port 6379 --requirepass <'yourpassword> --server-threads <#threads used for test>
