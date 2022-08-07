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
LABEL summary="Confluent platform Kafka."
LABEL io.confluent.docker=true
LABEL io.confluent.docker.git.id=$GIT_COMMIT
ARG BUILD_NUMBER=-1
LABEL io.confluent.docker.build.number=$BUILD_NUMBER
LABEL io.confluent.docker.git.repo="confluentinc/kafka-images"

ARG CONFLUENT_VERSION
ARG CONFLUENT_PACKAGES_REPO
ARG CONFLUENT_PLATFORM_LABEL

# allow arg override of required env params
ARG KAFKA_ZOOKEEPER_CONNECT
ENV KAFKA_ZOOKEEPER_CONNECT=${KAFKA_ZOOKEEPER_CONNECT}
ARG KAFKA_ADVERTISED_LISTENERS
ENV KAFKA_ADVERTISED_LISTENERS=${KAFKA_ADVERTISED_LISTENERS}

ENV COMPONENT=kafka

# primary
EXPOSE 9092
WORKDIR /
USER root
RUN groupadd -r appuser && useradd -r -g appuser  appuser 
RUN echo "===> Installing ${COMPONENT}..." \
    && echo "===> Adding confluent repository...${CONFLUENT_PACKAGES_REPO}" 
ADD confluent-kafka-7.2.1-1.noarch.rpm /confluent-kafka-7.2.1-1.noarch.rpm 
RUN rpm -ivh  confluent-kafka-7.2.1-1.noarch.rpm \
    && echo "===> Setting up ${COMPONENT} dirs" \
    && mkdir -p /var/lib/${COMPONENT}/data /etc/${COMPONENT}/secrets \
    && chown appuser:root -R /etc/kafka /var/log/kafka /var/log/confluent /var/lib/kafka /var/lib/zookeeper /etc/${COMPONENT}/secrets /var/lib/${COMPONENT} /etc/${COMPONENT} \
    && chmod -R ug+w /etc/kafka /var/log/kafka /var/log/confluent /var/lib/kafka /var/lib/zookeeper /var/lib/${COMPONENT} /etc/${COMPONENT}/secrets /etc/${COMPONENT}

VOLUME ["/var/lib/${COMPONENT}/data", "/etc/${COMPONENT}/secrets"]

COPY include/etc/confluent/docker /etc/confluent/docker
RUN chown -R appuser:appuser  /etc/confluent/docker

USER appuser

CMD ["/etc/confluent/docker/run"]
