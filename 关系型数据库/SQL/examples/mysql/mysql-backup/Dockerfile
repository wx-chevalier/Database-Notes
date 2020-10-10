FROM debian:8

MAINTAINER confirm IT solutions, dbarton

#
# Add user.
#

RUN \
    groupadd -g 666 mybackup && \
    useradd -u 666 -g 666 -d /backup -c "MySQL Backup User" mybackup

#
# Install required packages.
#

RUN \
    apt-get -y update && \
    apt-get -y install mydumper && \
    rm -rf /var/lib/apt/lists/*

#
# Install start script.
#

COPY init.sh /init.sh
RUN chmod 750 /init.sh

#
# Set container settings.
#

VOLUME ["/backup"]
WORKDIR /backup

#
# Start process.
#

CMD ["/init.sh"]
