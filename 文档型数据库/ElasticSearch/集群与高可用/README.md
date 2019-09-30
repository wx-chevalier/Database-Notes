# ElasticSearch 集群与高可用

# Cluster & HA: 集群与高可用

## Replica: 主从架构

Elasticsearch can be run in an HA configuration after the initial stack comes up. The first node needs to register as healthy before scaling it out. After the initial Elasticsearch member is healthy, then it can be scaled.

```sh
$ sysctl -w vm.max_map_count=262144

# 持久化配置信息
$ echo 'vm.max_map_count=262144' >> /etc/sysctl.conf
```

```sh
# 将服务部署到 Swarm 集群中
$ docker stack deploy -c $(pwd)/docker-compose.yml elk

# Find the Elasticsearch service ID
$ docker service ls

# Scale out the service to include more replicas:
$ docker service update --replicas=3 <replica_id>
```

## 多集群

# 链接

- https://logz.io/blog/elasticsearch-cluster-tutorial/
