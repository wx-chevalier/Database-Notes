#!/bin/bash
docker-compose up -d
wait_mysql() {
    local PORT=$1
    while :
    do
        sleep 3
        if mysql -P ${PORT} -e 'select version()' &>/dev/null; then
            break
        fi
    done
}

for PORT in 3306 3307 3308; do
    echo "##########################"
    echo "### MySQL at port ${PORT}"
    echo "##########################"
    wait_mysql $PORT
    export TEST_DSN="root:@tcp(127.1:${PORT})/sakila?parseTime=true"
    go test -v ./...
done

