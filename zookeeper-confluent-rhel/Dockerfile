FROM centos:centos7

ARG DOCKER_UPSTREAM_REGISTRY
ARG DOCKER_UPSTREAM_TAG=ubi8-latest



ARG PROJECT_VERSION
ARG ARTIFACT_ID
ARG GIT_COMMIT

LABEL maintainer="partner-support@confluent.io"
LABEL vendor="Confluent"
LABEL version=$GIT_COMMIT
LABEL release=$PROJECT_VERSION
LABEL name=$ARTIFACT_ID
LABEL summary="ZooKeeper is a centralized service for maintaining configuration information, naming, providing distributed synchronization, and providing group services."
LABEL description="ZooKeeper is a centralized service for maintaining configuration information, naming, providing distributed synchronization, and providing group services."
LABEL io.confluent.docker=true
LABEL io.confluent.docker.git.id=$GIT_COMMIT
ARG BUILD_NUMBER=-1
LABEL io.confluent.docker.build.number=$BUILD_NUMBER
LABEL io.confluent.docker.git.repo="confluentinc/kafka-images"

ARG CONFLUENT_VERSION
ARG CONFLUENT_PACKAGES_REPO
ARG CONFLUENT_PLATFORM_LABEL

EXPOSE 2181 2888 3888

ENV COMPONENT=zookeeper
WORKDIR /
USER root

RUN yum update -y && \
yum install -y wget && \
yum install -y java-1.8.0-openjdk java-1.8.0-openjdk-devel python* && \
yum clean all

RUN groupadd -r appuser && useradd -r -g appuser  appuser 
RUN echo "===> Installing ${COMPONENT}..." 
ADD confluent-kafka-7.2.1-1.noarch.rpm /confluent-kafka-7.2.1-1.noarch.rpm 
RUN rpm -ivh confluent-kafka-7.2.1-1.noarch.rpm \
    && echo "===> Setting up ${COMPONENT} dirs" \
    && mkdir -p /var/lib/${COMPONENT}/data /var/lib/${COMPONENT}/log /etc/${COMPONENT}/secrets \
    && chown -R appuser:root /etc/kafka /var/log/kafka /var/log/confluent /var/lib/kafka /var/lib/zookeeper /etc/${COMPONENT}/secrets \
    && chmod -R ug+w /etc/kafka /var/lib/${COMPONENT}/data /var/lib/${COMPONENT}/log /etc/${COMPONENT}/secrets /var/log/kafka

VOLUME ["/var/lib/${COMPONENT}/data", "/var/lib/${COMPONENT}/log", "/etc/${COMPONENT}/secrets"]

COPY include/etc/confluent/docker /etc/confluent/docker
ENV PATH="${PATH}:/etc/confluent/docker"
RUN chown -R appuser:appuser  /etc/confluent/docker

USER appuser

CMD ["/etc/confluent/docker/run"]