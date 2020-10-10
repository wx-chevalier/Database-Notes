# Derived from official mysql image (our base image)
FROM mysql:5.7

# SET ENV
ENV MYSQL_ROOT_PASSWORD roottoor
ENV MYSQL_ROOT_HOST %
ENV TZ 'Asia/Shanghai'

# Add a database
ENV MYSQL_DATABASE test

# Add the content of the sql-scripts/ directory to your image
# All scripts in docker-entrypoint-initdb.d/ are automatically
# executed during container startup
COPY ./docker-entrypoint-initdb.d/ /docker-entrypoint-initdb.d/