FROM centos:centos7
ENV ZOOKEEPER_VERSION 3.8.0
ENV ZOO_CONF_DIR=/conf \
    ZOO_DATA_DIR=/data \
    ZOO_DATA_LOG_DIR=/datalog \
    ZOO_LOG_DIR=/logs \
    ZOO_TICK_TIME=2000 \
    ZOO_INIT_LIMIT=5 \
    ZOO_SYNC_LIMIT=2 \
    ZOO_AUTOPURGE_PURGEINTERVAL=0 \
    ZOO_AUTOPURGE_SNAPRETAINCOUNT=3 \
    ZOO_MAX_CLIENT_CNXNS=60 \
    ZOO_STANDALONE_ENABLED=true \
    ZOO_ADMINSERVER_ENABLED=true

# Add a user with an explicit UID/GID and create necessary directories
RUN set -eux; \
    groupadd -r zookeeper --gid=1000; \
    useradd -r -g zookeeper --uid=1000 zookeeper; \
    mkdir -p "$ZOO_DATA_LOG_DIR" "$ZOO_DATA_DIR" "$ZOO_CONF_DIR" "$ZOO_LOG_DIR"; \
    chown zookeeper:zookeeper "$ZOO_DATA_LOG_DIR" "$ZOO_DATA_DIR" "$ZOO_CONF_DIR" "$ZOO_LOG_DIR"

# Install required packges
RUN set -eux; \
    yum install -y openssh-server  java-11-openjdk java-11-openjdk-devel ca-certificates dirmngr gosu gnupg netcat wget gpg ; \
    JAVA_HOME=/usr/lib/jvm/java-11-openjdk-amd64 ; \
    export PATH=$JAVA_HOME/bin:$PATH ; \
    export JAVA_HOME ; \
    export JRE_HOME ; \
    export PATH



COPY gosu /usr/local/bin/gosu
COPY gosu.asc /usr/local/bin/gosu.asc
RUN gpg --verify /usr/local/bin/gosu.asc ; \
    rm /usr/local/bin/gosu.asc ; \
    rm -r /root/.gnupg/ ;\
    chmod +x /usr/local/bin/gosu

ARG SHORT_DISTRO_NAME=zookeeper-3.8.0
ARG DISTRO_NAME=apache-zookeeper-3.8.0-bin

# Download Apache Zookeeper, verify its PGP signature, untar and clean up
COPY apache-zookeeper-${ZOOKEEPER_VERSION}-bin.tar.gz /
RUN set -eux; \
    tar -zxf "$DISTRO_NAME.tar.gz"; \
    mv "$DISTRO_NAME/conf/"* "$ZOO_CONF_DIR"; \
    rm -rf "$DISTRO_NAME.tar.gz"; \
    chown -R zookeeper:zookeeper "/$DISTRO_NAME"

WORKDIR $DISTRO_NAME
VOLUME ["$ZOO_DATA_DIR", "$ZOO_DATA_LOG_DIR", "$ZOO_LOG_DIR"]

EXPOSE 2181 2888 3888 8080

ENV PATH=$PATH:/$DISTRO_NAME/bin \
    ZOOCFGDIR=$ZOO_CONF_DIR

COPY docker-entrypoint.sh /
ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["zkServer.sh", "start-foreground"] 
