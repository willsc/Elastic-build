FROM centos:centos7

RUN yum  install -y curl fontconfig fonts-liberation libnss3 libfontconfig1 ca-certificates 
COPY bin/tini-amd64 /bin/tini
RUN set -eux ; \
    chmod 0555 /bin/tini

COPY  bin/kibana-8.5.2-linux-x86_64.tar.gz /tmp/kibana.tar.gz
RUN mkdir /usr/share/kibana
WORKDIR /usr/share/kibana
RUN tar --strip-components=1 -zxf /tmp/kibana.tar.gz
RUN chmod -R g=u /usr/share/kibana


RUN mkdir /usr/share/fonts/local
COPY bin/NotoSansCJK-Regular.ttc /usr/share/fonts/local/NotoSansCJK-Regular.ttc
RUN fc-cache -v

RUN chown 1000:0 /usr/share/kibana /usr/share/kibana
WORKDIR /usr/share/kibana
RUN ln -s /usr/share/kibana /opt/kibana


ENV ELASTIC_CONTAINER true
ENV PATH=/usr/share/kibana/bin:$PATH

# Set some Kibana configuration defaults.
COPY config/kibana.yml /usr/share/kibana/config/kibana.yml
RUN chown 1000:0 /usr/share/kibana/config/kibana.yml

COPY bin/kibana-docker /usr/local/bin/

# Ensure gid 0 write permissions for OpenShift.
RUN chmod g+ws /usr/share/kibana && \
    find /usr/share/kibana -gid 0 -and -not -perm /g+w -exec chmod g+w {} \;

# Remove the suid bit everywhere to mitigate "Stack Clash"
RUN find / -xdev -perm -4000 -exec chmod u-s {} +

# Provide a non-root user to run the process.
RUN groupadd --gid 1000 kibana && \
    useradd --uid 1000 --gid 1000 -G 0 \
      --home-dir /usr/share/kibana --no-create-home \
      kibana


LABEL org.label-schema.build-date="2022-11-16T20:43:16.419Z" \
  org.label-schema.license="Elastic License" \
  org.label-schema.name="Kibana" \
  org.label-schema.schema-version="1.0" \
  org.label-schema.url="https://www.elastic.co/products/kibana" \
  org.label-schema.usage="https://www.elastic.co/guide/en/kibana/reference/index.html" \
  org.label-schema.vcs-ref="d08bc2fc2c62274c16732ca0621964747cc961b8" \
  org.label-schema.vcs-url="https://github.com/elastic/kibana" \
  org.label-schema.vendor="Elastic" \
  org.label-schema.version="8.5.2" \
  org.opencontainers.image.created="2022-11-16T20:43:16.419Z" \
  org.opencontainers.image.documentation="https://www.elastic.co/guide/en/kibana/reference/index.html" \
  org.opencontainers.image.licenses="Elastic License" \
  org.opencontainers.image.revision="d08bc2fc2c62274c16732ca0621964747cc961b8" \
  org.opencontainers.image.source="https://github.com/elastic/kibana" \
  org.opencontainers.image.title="Kibana" \
  org.opencontainers.image.url="https://www.elastic.co/products/kibana" \
  org.opencontainers.image.vendor="Elastic" \
  org.opencontainers.image.version="8.5.2"


ENTRYPOINT ["/bin/tini", "--"]


CMD ["/usr/local/bin/kibana-docker"]


USER kibana      