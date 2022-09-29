FROM centos:centos7
ENV ZK_USER=zookeeper \
ZK_DATA_DIR=/var/lib/zookeeper/data \
ZK_DATA_LOG_DIR=/var/lib/zookeeper/log \
ZK_LOG_DIR=/var/log/zookeeper \
ZOOKEEPER_VERSION=3.8.0


ARG ZK_DIST=apache-zookeeper-3.8.0-bin

COPY apache-zookeeper-${ZOOKEEPER_VERSION}-bin.tar.gz /

RUN set -x \
    && yum install -y openssh-server  java-11-openjdk java-11-openjdk-devel ca-certificates dirmngr gosu gnupg netcat nc wget gpg  \
    && tar -xzf "$ZK_DIST.tar.gz" -C /opt \
    && rm -r "$ZK_DIST.tar.gz" \
    && ln -s /opt/$ZK_DIST /opt/zookeeper 
    
	

RUN set -eux ; \
     export JAVA_HOME=/usr/lib/jvm/java-11-openjdk-amd64 ; \
     export PATH=$JAVA_HOME/bin:$PATH ; \
     export JAVA_HOME ; \
     export JRE_HOME ; \
     export PATH=$PATH:/usr/bin:.

#Copy configuration generator script to bin
COPY scripts /opt/zookeeper/bin/

# Create a user for the zookeeper process and configure file system ownership 
# for nessecary directories and symlink the distribution as a user executable
RUN set -x \
	&& useradd $ZK_USER \
    && [ `id -u $ZK_USER` -eq 1000 ] \
    && [ `id -g $ZK_USER` -eq 1000 ] \
    && mkdir -p $ZK_DATA_DIR $ZK_DATA_LOG_DIR $ZK_LOG_DIR /usr/share/zookeeper /tmp/zookeeper /usr/etc/ \
	&& chown -R "$ZK_USER:$ZK_USER" /opt/$ZK_DIST $ZK_DATA_DIR $ZK_LOG_DIR $ZK_DATA_LOG_DIR /tmp/zookeeper \
	&& ln -s /opt/zookeeper/conf/ /usr/etc/zookeeper \
	&& ln -s /opt/zookeeper/bin/* /usr/bin \
	&& ln -s /opt/zookeeper/$ZK_DIST.jar /usr/share/zookeeper/ \
	&& ln -s /opt/zookeeper/lib/* /usr/share/zookeeper 

