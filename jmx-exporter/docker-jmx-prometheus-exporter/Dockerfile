FROM centos:centos7

RUN set -eux; \
    yum install -y openssh-server  java-11-openjdk java-11-openjdk-devel curl tcpdump lsof net-tools ca-certificates dirmngr gosu gnupg netcat wget gpg ; \
    JAVA_HOME=/usr/lib/jvm/java-11-openjdk-amd64 ; \
    export PATH=$JAVA_HOME/bin:$PATH ; \
    export JAVA_HOME ; \
    export JRE_HOME ; \
    export PATH

ENV VERSION 0.17.0
ENV JAR jmx_prometheus_httpserver-$VERSION-jar-with-dependencies.jar

COPY dumb-init /usr/local/bin/dumb-init 
RUN  chmod +x /usr/local/bin/dumb-init

RUN mkdir -p /opt/jmx_exporter
COPY jmx_prometheus_httpserver-0.17.0.jar /opt/jmx_exporter/$JAR 
RUN  chmod +x /opt/jmx_exporter/$JAR
COPY start.sh /opt/jmx_exporter/
COPY config.yml /opt/jmx_exporter/

CMD ["usr/local/bin/dumb-init", "/opt/jmx_exporter/start.sh"]

