FROM centos:centos7
ENV KAFKA_USER=kafka \
KAFKA_DATA_DIR=/var/lib/kafka/data \
KAFKA_HOME=/opt/kafka \
PATH=$PATH:/opt/kafka/bin


ARG KAFKA_DIST=kafka_2.13-3.2.0
COPY ${KAFKA_DIST}.tgz /

RUN set -x \
    && yum install -y openssh-server  java-11-openjdk java-11-openjdk-devel ca-certificates dirmngr gosu gnupg netcat wget gpg  \
    &&  tar -xzf "$KAFKA_DIST.tgz" -C /opt \
    && rm -r "$KAFKA_DIST.tgz"  
    
RUN set -eux ; \
     export JAVA_HOME=/usr/lib/jvm/java-11-openjdk-amd64 ; \
     export PATH=$JAVA_HOME/bin:$PATH ; \
     export JAVA_HOME ; \
     export JRE_HOME ; \
     export PATH   

COPY log4j.properties /opt/$KAFKA_DIST/config/

RUN set -x \
    && ln -s /opt/$KAFKA_DIST $KAFKA_HOME \
    && useradd $KAFKA_USER \
    && [ `id -u $KAFKA_USER` -eq 1000 ] \
    && [ `id -g $KAFKA_USER` -eq 1000 ] \
    && mkdir -p $KAFKA_DATA_DIR \
    && chown -R "$KAFKA_USER:$KAFKA_USER"  /opt/$KAFKA_DIST \
    && chown -R "$KAFKA_USER:$KAFKA_USER"  $KAFKA_DATA_DIR
