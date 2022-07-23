1. Build image with the following:

```
❯ docker build --platform x86_64 -t elasticsearch:latest .


[+] Building 1.4s (25/25) FINISHED
 => [internal] load build definition from Dockerfile                                                                                                                                                                    0.0s
 => => transferring dockerfile: 7.62kB                                                                                                                                                                                  0.0s
 => [internal] load .dockerignore                                                                                                                                                                                       0.0s
 => => transferring context: 2B                                                                                                                                                                                         0.0s
 => [internal] load metadata for docker.io/library/ubuntu:20.04                                                                                                                                                         1.3s
 => [auth] library/ubuntu:pull token for registry-1.docker.io                                                                                                                                                           0.0s
 => [internal] load build context                                                                                                                                                                                       0.0s
 => => transferring context: 220B                                                                                                                                                                                       0.0s
 => [builder  1/10] FROM docker.io/library/ubuntu:20.04@sha256:fd92c36d3cb9b1d027c4d2a72c6bf0125da82425fc2ca37c414d4f010180dc19                                                                                         0.0s
 => CACHED [stage-1  2/10] RUN yes no | dpkg-reconfigure dash &&     for iter in 1 2 3 4 5 6 7 8 9 10; do       export DEBIAN_FRONTEND=noninteractive &&       apt-get update &&       apt-get upgrade -y &&       apt  0.0s
 => CACHED [stage-1  3/10] RUN groupadd -g 1000 elasticsearch &&     adduser --uid 1000 --gid 1000 --home /usr/share/elasticsearch elasticsearch &&     adduser elasticsearch root &&     chown -R 0:0 /usr/share/elas  0.0s
 => CACHED [stage-1  4/10] WORKDIR /usr/share/elasticsearch                                                                                                                                                             0.0s
 => CACHED [builder  2/10] RUN for iter in 1 2 3 4 5 6 7 8 9 10; do       apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y curl  &&       exit_code=0 && break ||         exit_code=$? && echo "apt  0.0s
 => CACHED [builder  3/10] RUN set -eux ;     tini_bin="" ;     case "$(arch)" in         aarch64) tini_bin='tini-arm64' ;;         x86_64)  tini_bin='tini-amd64' ;;         *) echo >&2 ; echo >&2 "Unsupported arch  0.0s
 => CACHED [builder  4/10] RUN mkdir /usr/share/elasticsearch                                                                                                                                                           0.0s
 => CACHED [builder  5/10] WORKDIR /usr/share/elasticsearch                                                                                                                                                             0.0s
 => CACHED [builder  6/10] RUN curl --retry 10 -S -L --output /tmp/elasticsearch.tar.gz https://github.com/willsc/Elastic-build/releases/download/v1.0/elasticsearch-7.17.4-linux-$(arch).tar.gz                        0.0s
 => CACHED [builder  7/10] RUN tar -zxf /tmp/elasticsearch.tar.gz --strip-components=1                                                                                                                                  0.0s
 => CACHED [builder  8/10] COPY config/elasticsearch.yml config/                                                                                                                                                        0.0s
 => CACHED [builder  9/10] COPY config/log4j2.properties config/log4j2.docker.properties                                                                                                                                0.0s
 => CACHED [builder 10/10] RUN sed -i -e 's/ES_DISTRIBUTION_TYPE=tar/ES_DISTRIBUTION_TYPE=docker/' bin/elasticsearch-env &&     mkdir data &&     mv config/log4j2.properties config/log4j2.file.properties &&     mv   0.0s
 => CACHED [stage-1  5/10] COPY --from=builder --chown=0:0 /usr/share/elasticsearch /usr/share/elasticsearch                                                                                                            0.0s
 => CACHED [stage-1  6/10] COPY --from=builder --chown=0:0 /bin/tini /bin/tini                                                                                                                                          0.0s
 => CACHED [stage-1  7/10] COPY bin/docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh                                                                                                                            0.0s
 => CACHED [stage-1  8/10] RUN chmod g=u /etc/passwd &&     chmod 0555 /usr/local/bin/docker-entrypoint.sh &&     find / -xdev -perm -4000 -exec chmod ug-s {} + &&     chmod 0775 /usr/share/elasticsearch &&     cho  0.0s
 => CACHED [stage-1  9/10] COPY bin/docker-openjdk /etc/ca-certificates/update.d/docker-openjdk                                                                                                                         0.0s
 => CACHED [stage-1 10/10] RUN /etc/ca-certificates/update.d/docker-openjdk                                                                                                                                             0.0s
 => exporting to image                                                                                                                                                                                                  0.0s
 => => exporting layers                                                                                                                                                                                                 0.0s
 => => writing image sha256:cba9ab0b1239560f4c3d0f48516eb7508d02b221d62be917d40990a2cc3de790                                                                                                                            0.0s
 => => naming to docker.io/library/elasticsearch:latest
```


2. Run container locally to test it.

```
docker run -p 127.0.0.1:9200:9200 -p 127.0.0.1:9300:9300 -e "discovery.type=single-node" elasticsearch:latest


{"type": "server", "timestamp": "2022-07-23T13:05:31,259Z", "level": "INFO", "component": "o.e.n.Node", "cluster.name": "docker-cluster", "node.name": "db16cb8649b8", "message": "version[7.17.4], pid[8], build[default/docker/79878662c54c886ae89206c685d9f1051a9d6411/2022-05-18T18:04:20.964345128Z], OS[Linux/5.10.76-linuxkit/amd64], JVM[Oracle Corporation/OpenJDK 64-Bit Server VM/18.0.1.1/18.0.1.1+2-6]" }
{"type": "server", "timestamp": "2022-07-23T13:05:31,265Z", "level": "INFO", "component": "o.e.n.Node", "cluster.name": "docker-cluster", "node.name": "db16cb8649b8", "message": "JVM home [/usr/share/elasticsearch/jdk], using bundled JDK [true]" }
{"type": "server", "timestamp": "2022-07-23T13:05:31,265Z", "level": "INFO", "component": "o.e.n.Node", "cluster.name": "docker-cluster", "node.name": "db16cb8649b8", "message": "JVM arguments [-Xshare:auto, -Des.networkaddress.cache.ttl=60, -Des.networkaddress.cache.negative.ttl=10, -XX:+AlwaysPreTouch, -Xss1m, -Djava.awt.headless=true, -Dfile.encoding=UTF-8, -Djna.nosys=true, -XX:-OmitStackTraceInFastThrow, -XX:+ShowCodeDetailsInExceptionMessages, -Dio.netty.noUnsafe=true, -Dio.netty.noKeySetOptimization=true, -Dio.netty.recycler.maxCapacityPerThread=0, -Dio.netty.allocator.numDirectArenas=0, -Dlog4j.shutdownHookEnabled=false, -Dlog4j2.disable.jmx=true, -Dlog4j2.formatMsgNoLookups=true, -Djava.locale.providers=SPI,COMPAT, --add-opens=java.base/java.io=ALL-UNNAMED, -Djava.security.manager=allow, -XX:+UseG1GC, -Djava.io.tmpdir=/tmp/elasticsearch-726754027194583673, -XX:+HeapDumpOnOutOfMemoryError, -XX:+ExitOnOutOfMemoryError, -XX:HeapDumpPath=data, -XX:ErrorFile=logs/hs_err_pid%p.log, -Xlog:gc*,gc+age=trace,safepoint:file=logs/gc.log:utctime,pid,tags:filecount=32,filesize=64m, -Des.cgroups.hierarchy.override=/, -Xms7375m, -Xmx7375m, -XX:MaxDirectMemorySize=3867148288, -XX:G1HeapRegionSize=4m, -XX:InitiatingHeapOccupancyPercent=30, -XX:G1ReservePercent=15, -Des.path.home=/usr/share/elasticsearch, -Des.path.conf=/usr/share/elasticsearch/config, -Des.distribution.flavor=default, -Des.distribution.type=docker, -Des.bundled_jdk=true]" }
{"type": "server", "timestamp": "2022-07-23T13:05:33,453Z", "level": "INFO", "component": "o.e.p.PluginsService", "cluster.name": "docker-cluster", "node.name": "db16cb8649b8", "message": "loaded module [aggs-matrix-stats]" }
{"type": "server", "timestamp": "2022-07-23T13:05:33,453Z", "level": "INFO", "component": "o.e.p.PluginsService", "cluster.name": "docker-cluster", "node.name": "db16cb8649b8", "message": "loaded module [analysis-common]" }
{"type": "server", "timestamp": "2022-07-23T13:05:33,453Z", "level": "INFO", "component": "o.e.p.PluginsService", "cluster.name": "docker-cluster", "node.name": "db16cb8649b8", "message": "loaded module [constant-keyword]" }
{"type": "server", "timestamp": "2022-07-23T13:05:33,454Z", "level": "INFO", "component": "o.e.p.PluginsService", "cluster.name": "docker-cluster", "node.name": "db16cb8649b8", "message": "loaded module [frozen-indices]" }

```

3. Push image to repository.
