FROM centos:centos7

ENV ZK_USER=zookeeper \
    ZK_DATA_DIR=/var/lib/zookeeper/data \
    ZK_DATA_LOG_DIR=/var/lib/zookeeper/log \
    ZK_LOG_DIR=/var/log/zookeeper \
    JAVA_HOME=/usr/lib/jvm/java-8-openjdk-amd64

RUN set -eux; \
    groupadd -r zookeeper --gid=1000; \
    useradd -r -g zookeeper --uid=1000 zookeeper


ARG ZK_DIST=apache-zookeeper-3.8.0-bin

RUN set -eux; \
    yum install -y openssh-server  java-11-openjdk java-11-openjdk-devel ca-certificates dirmngr gosu gnupg netcat wget gpg ; \
    JAVA_HOME=/usr/lib/jvm/java-11-openjdk-amd64 ; \
    export PATH=$JAVA_HOME/bin:$PATH ; \
    export JAVA_HOME ; \
    export JRE_HOME ; \
    export PATH

COPY $ZK_DIST.tar.gz /

RUN set -x \   
    && tar -xzf "$ZK_DIST.tar.gz" -C /opt \
    && rm -r  "$ZK_DIST.tar.gz"  \
    && ln -s /opt/$ZK_DIST /opt/zookeeper \
    && chown -hLR zookeeper:zookeeper /opt/zookeeper \
    && rm -rf /opt/zookeeper/CHANGES.txt \
    /opt/zookeeper/README.txt \
    /opt/zookeeper/NOTICE.txt \
    /opt/zookeeper/CHANGES.txt \
    /opt/zookeeper/README_packaging.txt \
    /opt/zookeeper/build.xml \
    /opt/zookeeper/config \
    /opt/zookeeper/contrib \
    /opt/zookeeper/dist-maven \
    /opt/zookeeper/docs \
    /opt/zookeeper/ivy.xml \
    /opt/zookeeper/ivysettings.xml \
    /opt/zookeeper/recipes \
    /opt/zookeeper/src \
    /opt/zookeeper/$ZK_DIST.jar.asc \
    /opt/zookeeper/$ZK_DIST.jar.md5 \
    /opt/zookeeper/$ZK_DIST.jar.sha1 

# Copy configuration generator script to bin
COPY zkGenConfig.sh zkOk.sh zkMetrics.sh /opt/zookeeper/bin/

# Create a user for the zookeeper process and configure file system ownership
# for necessary directories and symlink the distribution as a user executable
RUN set -x \    
    && mkdir -p $ZK_DATA_DIR $ZK_DATA_LOG_DIR $ZK_LOG_DIR /usr/share/zookeeper /tmp/zookeeper /usr/etc/ \
    && ln -s /opt/zookeeper/conf/ /usr/etc/zookeeper \
    && ln -s /opt/zookeeper/bin/* /usr/bin \
    && ln -s /opt/zookeeper/$ZK_DIST.jar /usr/share/zookeeper/ \
    && ln -s /opt/zookeeper/lib/* /usr/share/zookeeper \
    && chown -R "$ZK_USER:$ZK_USER" /opt/$ZK_DIST $ZK_DATA_DIR $ZK_LOG_DIR $ZK_DATA_LOG_DIR /tmp/zookeeper 

USER zookeeper
