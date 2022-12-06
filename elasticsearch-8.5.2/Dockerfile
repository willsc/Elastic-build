FROM centos:centos7

RUN  yum install -y curl ca-certificates curl netcat p11-kit unzip zip openjdk-11* && yum clean all
COPY bin/tini-amd64 /bin/tini
RUN set -eux ; \
    chmod 0555 /bin/tini

RUN mkdir /usr/share/elasticsearch
WORKDIR /usr/share/elasticsearch

COPY bin/elasticsearch-8.5.2-linux-x86_64.tar.gz /tmp/elasticsearch.tar.gz
RUN tar -zxf /tmp/elasticsearch.tar.gz --strip-components=1

# The distribution includes a `config` directory, no need to create it
COPY config/elasticsearch.yml config/
COPY config/log4j2.properties config/log4j2.docker.properties

RUN sed -i -e 's/ES_DISTRIBUTION_TYPE=tar/ES_DISTRIBUTION_TYPE=docker/' bin/elasticsearch-env && \
    mkdir data && \
    mv config/log4j2.properties config/log4j2.file.properties && \
    mv config/log4j2.docker.properties config/log4j2.properties && \
    find . -type d -exec chmod 0555 {} + && \
    find . -type f -exec chmod 0444 {} + && \
    chmod 0555 bin/* jdk/bin/* jdk/lib/jspawnhelper modules/x-pack-ml/platform/linux-*/bin/* && \
    chmod 0775 bin config config/jvm.options.d data logs plugins && \
    find config -type f -exec chmod 0664 {} +

RUN groupadd -g 1000 elasticsearch 
RUN adduser --uid 1000 --gid 1000 --home-dir /usr/share/elasticsearch elasticsearch && \
    chown -R 0:0 /usr/share/elasticsearch 
    

ENV ELASTIC_CONTAINER true

WORKDIR /usr/share/elasticsearch

RUN chown 0:0 /usr/share/elasticsearch /usr/share/elasticsearch
RUN chown 0:0 /bin/tini /bin/tini

ENV PATH /usr/share/elasticsearch/bin:$PATH

COPY bin/docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh



RUN chmod g=u /etc/passwd && \
    chmod 0555 /usr/local/bin/docker-entrypoint.sh && \
    find / -xdev -perm -4000 -exec chmod ug-s {} + && \
    chmod 0775 /usr/share/elasticsearch && \
    chown elasticsearch bin config config/jvm.options.d data logs plugins



EXPOSE 9200 9300


LABEL org.label-schema.build-date="2022-11-17T21:17:54.410437150Z" \
  org.label-schema.license="Elastic-License-2.0" \
  org.label-schema.name="Elasticsearch" \
  org.label-schema.schema-version="1.0" \
  org.label-schema.url="https://www.elastic.co/products/elasticsearch" \
  org.label-schema.usage="https://www.elastic.co/guide/en/elasticsearch/reference/index.html" \
  org.label-schema.vcs-ref="a846182fa16b4ebfcc89aa3c11a11fd5adf3de04" \
  org.label-schema.vcs-url="https://github.com/elastic/elasticsearch" \
  org.label-schema.vendor="Elastic" \
  org.label-schema.version="8.5.2" \
  org.opencontainers.image.created="2022-11-17T21:17:54.410437150Z" \
  org.opencontainers.image.documentation="https://www.elastic.co/guide/en/elasticsearch/reference/index.html" \
  org.opencontainers.image.licenses="Elastic-License-2.0" \
  org.opencontainers.image.revision="a846182fa16b4ebfcc89aa3c11a11fd5adf3de04" \
  org.opencontainers.image.source="https://github.com/elastic/elasticsearch" \
  org.opencontainers.image.title="Elasticsearch" \
  org.opencontainers.image.url="https://www.elastic.co/products/elasticsearch" \
  org.opencontainers.image.vendor="Elastic" \
  org.opencontainers.image.version="8.5.2"

# Our actual entrypoint is `tini`, a minimal but functional init program. It
# calls the entrypoint we provide, while correctly forwarding signals.
ENTRYPOINT ["/bin/tini", "--", "/usr/local/bin/docker-entrypoint.sh"]
# Dummy overridable parameter parsed by entrypoint
CMD ["eswrapper"]

USER elasticsearch:root
