# Exporter

# 收集 MySQL 指标

mysqld_exporter 是 Prometheus 官方提供的一个 exporter，我们首先 下载最新版本 并解压（开箱即用）。mysqld_exporter 需要连接到 mysqld 才能收集它的指标，可以通过两种方式来设置 mysqld 数据源。第一种是通过环境变量 DATA_SOURCE_NAME，这被称为 DSN（数据源名称），它必须符合 DSN 的格式，一个典型的 DSN 格式像这样：user:password@(host:port)/。

```s
$ export DATA_SOURCE_NAME='root:123456@(192.168.0.107:3306)/'
$ ./mysqld_exporter
```

另一种方式是通过配置文件，默认的配置文件是 ~/.my.cnf，或者通过 --config.my-cnf 参数指定：

```s
$ ./mysqld_exporter --config.my-cnf=".my.cnf"

# 配置文件的格式如下
$ cat .my.cnf
[client]
host=localhost
port=3306
user=root
password=123456
```

这里为简单起见，在 mysqld_exporter 中直接使用了 root 连接数据库，在真实环境中，可以为 mysqld_exporter 创建一个单独的用户，并赋予它受限的权限（PROCESS、REPLICATION CLIENT、SELECT），最好还限制它的最大连接数（MAX_USER_CONNECTIONS）。

```s
CREATE USER 'exporter'@'localhost' IDENTIFIED BY 'password' WITH MAX_USER_CONNECTIONS 3;
GRANT PROCESS, REPLICATION CLIENT, SELECT ON *.* TO 'exporter'@'localhost';
```

# 收集 Nginx 指标

官方提供了两种收集 Nginx 指标的方式。第一种是 Nginx metric library，这是一段 Lua 脚本（prometheus.lua），Nginx 需要开启 Lua 支持（libnginx-mod-http-lua 模块）。为方便起见，也可以使用 OpenResty 的 OPM（OpenResty Package Manager）或者 luarocks（The Lua package manager）来安装。第二种是 Nginx VTS exporter，这种方式比第一种要强大的多，安装要更简单，支持的指标也更丰富，它依赖于 nginx-module-vts 模块，vts 模块可以提供大量的 Nginx 指标数据，可以通过 JSON、HTML 等形式查看这些指标。Nginx VTS exporter 就是通过抓取 /status/format/json 接口来将 vts 的数据格式转换为 Prometheus 的格式。不过，在 nginx-module-vts 最新的版本中增加了一个新接口：/status/format/prometheus，这个接口可以直接返回 Prometheus 的格式。

除此之外，还有很多其他的方式来收集 Nginx 的指标，比如：nginx_exporter 通过抓取 Nginx 自带的统计页面 /nginx_status 可以获取一些比较简单的指标（需要开启 ngx_http_stub_status_module 模块）；nginx_request_exporter 通过 syslog 协议 收集并分析 Nginx 的 access log 来统计 HTTP 请求相关的一些指标；nginx-prometheus-shiny-exporter 和 nginx_request_exporter 类似，也是使用 syslog 协议来收集 access log，不过它是使用 Crystal 语言 写的。还有 vovolie/lua-nginx-prometheus 基于 Openresty、Prometheus、Consul、Grafana 实现了针对域名和 Endpoint 级别的流量统计。

# JMX

最后让我们来看下如何收集 Java 应用的指标，Java 应用的指标一般是通过 JMX（Java Management Extensions）来获取的，顾名思义，JMX 是管理 Java 的一种扩展，它可以方便的管理和监控正在运行的 Java 程序。

JMX Exporter 用于收集 JMX 指标，很多使用 Java 的系统，都可以使用它来收集指标，比如：Kafaka、Cassandra 等。首先我们下载 JMX Exporter，然后在运行 Java 程序时通过 -javaagent 参数来加载：

```sh
$ java -javaagent:jmx_prometheus_javaagent-0.3.1.jar=9404:config.yml -jar spring-boot-sample-1.0-SNAPSHOT.jar
```

其中，9404 是 JMX Exporter 暴露指标的端口，`config.yml` 是 JMX Exporter 的配置文件，它的内容可以 [参考 JMX Exporter 的配置说明](https://github.com/prometheus/jmx_exporter#configuration)。然后检查下指标数据是否正确获取：

```s
$ curl http://localhost:9404/metrics
```
