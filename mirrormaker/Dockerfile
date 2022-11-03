FROM centos:centos7

ARG kafka_version=3.2.0
ARG scala_version=2.13

ENV KAFKA_VERSION "3.2.0"
ENV KAFKA_MM_CFG_ROOT "/etc/mirror-maker"
ENV KAFKA_ROOT "/opt/kafka"
ENV SCALA_VERSION "2.13"
ENV KAFKA_HEAP_OPTS "-Xms1024M -Xmx1024M"
ENV KAFKA_MM_WHITELIST ".*"
ENV KAFKA_MM_NUMSTREAMS "1"
ENV JMXPORT "9998"


RUN yum install -y  java-1.8.0-openjdk.x86_64 jq curl netcat wget  ; \
    export JRE_HOME=/usr/lib/jvm/jre ; \
    export JAVA_HOME=/usr/lib/jvm/jre-1.8.0-openjdk ; \
    PATH=$PATH:$JRE_HOME:$JAVA_HOME

ENV KAFKA_VERSION=$kafka_version \
     SCALA_VERSION=$scala_version \
     KAFKA_HOME=/opt/kafka

ENV LANG="C.UTF-8"     

ENV KAFKA_VERSION=$kafka_version \
    SCALA_VERSION=$scala_version \
    KAFKA_HOME=/opt/kafka
    
RUN  groupadd kafka ; \
     useradd  -u 1000 -g kafka -m -s /bin/bash kafka
 
RUN mkdir ${KAFKA_ROOT}
RUN mkdir ${KAFKA_MM_CFG_ROOT}

ENV PATH=${PATH}:${KAFKA_HOME}/bin

COPY kafka_${SCALA_VERSION}-${KAFKA_VERSION}.tgz  /tmp2/

RUN set -eux ; \
    tar xfz /tmp2/kafka_${SCALA_VERSION}-${KAFKA_VERSION}.tgz -C /opt ; \
    chown -hLR kafka:kafka /opt ; \
    chown kafka:kafka /usr/bin/*.sh ; \
    rm /tmp2/kafka_${SCALA_VERSION}-${KAFKA_VERSION}.tgz ; \
    ln -s /opt/kafka_${SCALA_VERSION}-${KAFKA_VERSION} ${KAFKA_HOME} ; \
    rm -rf /tmp2 ; \
    rm -rf /var/lib/apt/lists/*


ADD ./run.sh ${KAFKA_ROOT}/
RUN chmod +x ${KAFKA_ROOT}/run.sh
RUN chown kafka:kafka -R ${KAFKA_ROOT}

CMD ["su", "-c", "${KAFKA_ROOT}/run.sh", "kafka"]