# db-mysql

- Create Image | 生成镜像

```sh
$ ./build-image.sh
```

- Deploy Image | 部署镜像

```sh
# 无目录共享运行
$ docker run --rm --name=test-mysql -p 3306:3306 test-mysql

# 自定义配置文件
$ docker run --rm --name=test-mysql -p 3306:3306 -v ./etc:/etc/mysql/conf.d test-mysql

# MAC 下添加特殊目录共享
$ docker run -d --restart always --name=test-mysql  -v ~/Desktop/test/mysql:/var/lib/mysql test-mysql

$ docker run -d --restart always --name=test-mysql -v /var/test/mysql:/var/lib/mysql test-mysql
```

- Test | 测试

```sh
$ docker run --rm -ti --name=mycli \
  --link=test-mysql:mysql \
  diyan/mycli \
  --host=mysql \
  --database=test \
  --user=root \
  --password=roottoor
```
