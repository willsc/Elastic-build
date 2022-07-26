FROM centos:centos7

MAINTAINER C Wills

ENV ZOOKEEPER_VERSION 3.8.0

#Download Zookeeper
COPY apache-zookeeper-${ZOOKEEPER_VERSION}-bin.tar.gz /

#Install
RUN yum install -y openssh-server java-1.8.0-openjdk.x86_64 ; \
    rm -r /etc/ssh/ssh*key ; \
    systemctl restart sshd ; \
    export JRE_HOME=/usr/lib/jvm/jre ; \
    export JAVA_HOME=/usr/lib/jvm/jre-1.8.0-openjdk ; \
    PATH=$PATH:$JRE_HOME:$JAVA_HOME
RUN /usr/bin/ssh-keygen -A
RUN mkdir /opt ; \
    chmod 755 /opt
RUN tar -xzf apache-zookeeper-${ZOOKEEPER_VERSION}-bin.tar.gz -C /opt

#Configure
RUN mv /opt/apache-zookeeper-${ZOOKEEPER_VERSION}-bin/conf/zoo_sample.cfg /opt/apache-zookeeper-${ZOOKEEPER_VERSION}-bin/conf/zoo.cfg

ENV JAVA_HOME=/usr/lib/jvm/jre-1.8.0-openjdk
ENV ZK_HOME /opt/apache-zookeeper-${ZOOKEEPER_VERSION}-bin
RUN sed  -i "s|/tmp/zookeeper|$ZK_HOME/data|g" $ZK_HOME/conf/zoo.cfg; mkdir $ZK_HOME/data

ADD start-zk.sh /usr/bin/start-zk.sh 
EXPOSE 2181 2888 3888

WORKDIR /opt/apache-zookeeper-${ZOOKEEPER_VERSION}-bin
VOLUME ["/opt/apache-zookeeper-${ZOOKEEPER_VERSION}-bin/conf", "/opt/apache-zookeeper-${ZOOKEEPER_VERSION}-bin/data"]

CMD /usr/sbin/sshd && bash /usr/bin/start-zk.sh


