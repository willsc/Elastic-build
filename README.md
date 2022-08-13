# Elastic-build

```
version: '2.2'
services:
  es01:
    image: docker.elastic.co/elasticsearch/elasticsearch:7.17.4
    container_name: es01
    environment:
      - node.name=es01
      - cluster.name=es-docker-cluster
      - discovery.seed_hosts=es02,es03
      - cluster.initial_master_nodes=es01,es02,es03
      - bootstrap.memory_lock=true
      - "ES_JAVA_OPTS=-Xms512m -Xmx512m"
    ulimits:
      memlock:
        soft: -1
        hard: -1
    volumes:
      - data01:/usr/share/elasticsearch/data
    ports:
      - 9200:9200
    networks:
      - elastic
  es02:
    image: docker.elastic.co/elasticsearch/elasticsearch:7.17.4
    container_name: es02
    environment:
      - node.name=es02
      - cluster.name=es-docker-cluster
      - discovery.seed_hosts=es01,es03
      - cluster.initial_master_nodes=es01,es02,es03
      - bootstrap.memory_lock=true
      - "ES_JAVA_OPTS=-Xms512m -Xmx512m"
    ulimits:
      memlock:
        soft: -1
        hard: -1
    volumes:
      - data02:/usr/share/elasticsearch/data
    networks:
      - elastic
  es03:
    image: docker.elastic.co/elasticsearch/elasticsearch:7.17.4
    container_name: es03
    environment:
      - node.name=es03
      - cluster.name=es-docker-cluster
      - discovery.seed_hosts=es01,es02
      - cluster.initial_master_nodes=es01,es02,es03
      - bootstrap.memory_lock=true
      - "ES_JAVA_OPTS=-Xms512m -Xmx512m"
    ulimits:
      memlock:
        soft: -1
        hard: -1
    volumes:
      - data03:/usr/share/elasticsearch/data
    networks:
      - elastic
  kibana:
      image: docker.elastic.co/kibana/kibana:7.17.4
      container_name: kibana
      ports:
        - 5601:5601
      environment:
        ELASTICSEARCH_URL: http://es01:9200
        ELASTICSEARCH_HOSTS: '["http://es01:9200","http://es02:9200","http://es03:9200"]'
      networks:
        - elastic
volumes:
  data01:
    driver: local
  data02:
    driver: local
  data03:
    driver: local

networks:
  elastic:
    driver: bridge


```


Deployment via Helm chart:

```

1. Dry run the helm chart 
helm install elasticsearch-master --dry-run  ./elasticsearch  -f ./elasticsearch/master-values.yaml -n hornet
NAME: elasticsearch-master
LAST DEPLOYED: Sat Aug 13 13:07:50 2022
NAMESPACE: hornet
STATUS: pending-install
REVISION: 1
HOOKS:
---
# Source: elasticsearch/templates/test/test-elasticsearch-health.yaml
apiVersion: v1
kind: Pod
metadata:
  name: "elasticsearch-master-zmhbe-test"
  annotations:
    "helm.sh/hook": test
    "helm.sh/hook-delete-policy": hook-succeeded
spec:
  securityContext:
    fsGroup: 1000
    runAsUser: 1000
  containers:
  - name: "elasticsearch-master-zsimg-test"
    image: "docker.elastic.co/elasticsearch/elasticsearch:7.17.4"
    imagePullPolicy: "IfNotPresent"
    command:
      - "sh"
      - "-c"
      - |
        #!/usr/bin/env bash -e
        curl -XGET --fail 'elasticsearchuat-master:9200/_cluster/health?wait_for_status=green&timeout=1s'
  restartPolicy: Never
MANIFEST:
---
# Source: elasticsearch/templates/poddisruptionbudget.yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: "elasticsearchuat-master-pdb"
spec:
  maxUnavailable: 1
  selector:
    matchLabels:
      app: "elasticsearchuat-master"
---
# Source: elasticsearch/templates/service.yaml
kind: Service
apiVersion: v1
metadata:
  name: elasticsearchuat-master
  labels:
    heritage: "Helm"
    release: "elasticsearch-master"
    chart: "elasticsearch"
    app: "elasticsearchuat-master"
  annotations:
    {}
spec:
  type: ClusterIP
  selector:
    release: "elasticsearch-master"
    chart: "elasticsearch"
    app: "elasticsearchuat-master"
  publishNotReadyAddresses: false
  ports:
  - name: http
    protocol: TCP
    port: 9200
  - name: transport
    protocol: TCP
    port: 9300
---
# Source: elasticsearch/templates/service.yaml
kind: Service
apiVersion: v1
metadata:
  name: elasticsearchuat-master-headless
  labels:
    heritage: "Helm"
    release: "elasticsearch-master"
    chart: "elasticsearch"
    app: "elasticsearchuat-master"
  annotations:
    service.alpha.kubernetes.io/tolerate-unready-endpoints: "true"
spec:
  clusterIP: None # This is needed for statefulset hostnames like elasticsearch-0 to resolve
  # Create endpoints also if the related pod isn't ready
  publishNotReadyAddresses: true
  selector:
    app: "elasticsearchuat-master"
  ports:
  - name: http
    port: 9200
  - name: transport
    port: 9300
---
# Source: elasticsearch/templates/statefulset.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: elasticsearchuat-master
  labels:
    heritage: "Helm"
    release: "elasticsearch-master"
    chart: "elasticsearch"
    app: "elasticsearchuat-master"
  annotations:
    esMajorVersion: "7"
spec:
  serviceName: elasticsearchuat-master-headless
  selector:
    matchLabels:
      app: "elasticsearchuat-master"
  replicas: 3
  podManagementPolicy: Parallel
  updateStrategy:
    type: RollingUpdate
  volumeClaimTemplates:
  - metadata:
      name: elasticsearchuat-master
    spec:
      accessModes:
      - ReadWriteOnce
      resources:
        requests:
          storage: 10Gi
  template:
    metadata:
      name: "elasticsearchuat-master"
      labels:
        release: "elasticsearch-master"
        chart: "elasticsearch"
        app: "elasticsearchuat-master"
      annotations:


    spec:
      securityContext:
        fsGroup: 1000
        runAsUser: 1000
      automountServiceAccountToken: true
      affinity:
        podAntiAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
          - labelSelector:
              matchExpressions:
              - key: app
                operator: In
                values:
                - "elasticsearchuat-master"
            topologyKey: kubernetes.io/hostname
      terminationGracePeriodSeconds: 120
      volumes:
      enableServiceLinks: true
      containers:
      - name: "elasticsearch"
        securityContext:
          capabilities:
            drop:
            - ALL
          runAsNonRoot: true
          runAsUser: 1000
        image: "docker.elastic.co/elasticsearch/elasticsearch:7.17.4"
        imagePullPolicy: "IfNotPresent"
        readinessProbe:
          exec:
            command:
              - bash
              - -c
              - |
                set -e
                # If the node is starting up wait for the cluster to be ready (request params: "wait_for_status=green&timeout=1s" )
                # Once it has started only check that the node itself is responding
                START_FILE=/tmp/.es_start_file


                # Disable nss cache to avoid filling dentry cache when calling curl
                # This is required with Elasticsearch Docker using nss < 3.52
                export NSS_SDB_USE_CACHE=no


                http () {
                  local path="${1}"
                  local args="${2}"
                  set -- -XGET -s


                  if [ "$args" != "" ]; then
                    set -- "$@" $args
                  fi


                  if [ -n "${ELASTIC_PASSWORD}" ]; then
                    set -- "$@" -u "elastic:${ELASTIC_PASSWORD}"
                  fi


                  curl --output /dev/null -k "$@" "http://127.0.0.1:9200${path}"
                }


                if [ -f "${START_FILE}" ]; then
                  echo 'Elasticsearch is already running, lets check the node is healthy'
                  HTTP_CODE=$(http "/" "-w %{http_code}")
                  RC=$?
                  if [[ ${RC} -ne 0 ]]; then
                    echo "curl --output /dev/null -k -XGET -s -w '%{http_code}' \${BASIC_AUTH} http://127.0.0.1:9200/ failed with RC ${RC}"
                    exit ${RC}
                  fi
                  # ready if HTTP code 200, 503 is tolerable if ES version is 6.x
                  if [[ ${HTTP_CODE} == "200" ]]; then
                    exit 0
                  elif [[ ${HTTP_CODE} == "503" && "7" == "6" ]]; then
                    exit 0
                  else
                    echo "curl --output /dev/null -k -XGET -s -w '%{http_code}' \${BASIC_AUTH} http://127.0.0.1:9200/ failed with HTTP code ${HTTP_CODE}"
                    exit 1
                  fi


                else
                  echo 'Waiting for elasticsearch cluster to become ready (request params: "wait_for_status=green&timeout=1s" )'
                  if http "/_cluster/health?wait_for_status=green&timeout=1s" "--fail" ; then
                    touch ${START_FILE}
                    exit 0
                  else
                    echo 'Cluster is not yet ready (request params: "wait_for_status=green&timeout=1s" )'
                    exit 1
                  fi
                fi
          failureThreshold: 3
          initialDelaySeconds: 10
          periodSeconds: 10
          successThreshold: 3
          timeoutSeconds: 5
        ports:
        - name: http
          containerPort: 9200
        - name: transport
          containerPort: 9300
        resources:
          limits:
            cpu: 1000m
            memory: 2Gi
          requests:
            cpu: 1000m
            memory: 2Gi
        env:
          - name: node.name
            valueFrom:
              fieldRef:
                fieldPath: metadata.name
          - name: cluster.initial_master_nodes
            value: "elasticsearchuat-master-0,elasticsearchuat-master-1,elasticsearchuat-master-2,"
          - name: discovery.seed_hosts
            value: "elasticsearchuat-master-headless"
          - name: cluster.name
            value: "elasticsearchuat"
          - name: network.host
            value: "0.0.0.0"
          - name: cluster.deprecation_indexing.enabled
            value: "false"
          - name: ES_JAVA_OPTS
            value: "-Xmx1g -Xms1g"
          - name: node.data
            value: "false"
          - name: node.ingest
            value: "false"
          - name: node.master
            value: "true"
          - name: node.ml
            value: "false"
          - name: node.remote_cluster_client
            value: "false"
        volumeMounts:
          - name: "elasticsearchuat-master"
            mountPath: /usr/share/elasticsearch/data


NOTES:
1. Watch all cluster members come up.
  $ kubectl get pods --namespace=hornet -l app=elasticsearchuat-master -w2. Test cluster health using Helm test.
  $ helm --namespace=hornet test elasticsearch-master

2. Deploy the chart 
❯ helm install elasticsearch-master   ./elasticsearch  -f ./elasticsearch/master-values.yaml -n hornet
NAME: elasticsearch-master
LAST DEPLOYED: Sat Aug 13 13:09:21 2022
NAMESPACE: hornet
STATUS: deployed
REVISION: 1
NOTES:
1. Watch all cluster members come up.
  $ kubectl get pods --namespace=hornet -l app=elasticsearchuat-master -w2. Test cluster health using Helm test.
  $ helm --namespace=hornet test elasticsearch-master

The deployment  should look like the following :

k get svc -n hornet
NAME                               TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)             AGE
elasticsearchuat-master            ClusterIP   172.20.61.147   <none>        9200/TCP,9300/TCP   93s
elasticsearchuat-master-headless   ClusterIP   None            <none>        9200/TCP,9300/TCP   93s
❯ k get pods -n hornet
NAME                        READY   STATUS    RESTARTS   AGE
elasticsearchuat-master-0   1/1     Running   0          102s
elasticsearchuat-master-1   1/1     Running   0          102s
elasticsearchuat-master-2   1/1     Running   0          102s

Cluster should have formed with the following node count.
❯ k exec -it elasticsearchuat-master-0 -n hornet -- sh
sh-5.0$  curl http://localhost:9200/_cluster/health?pretty
{
  "cluster_name" : "elasticsearchuat",
  "status" : "green",
  "timed_out" : false,
  "number_of_nodes" : 3,
  "number_of_data_nodes" : 0,
  "active_primary_shards" : 0,
  "active_shards" : 0,
  "relocating_shards" : 0,
  "initializing_shards" : 0,
  "unassigned_shards" : 0,
  "delayed_unassigned_shards" : 0,
  "number_of_pending_tasks" : 0,
  "number_of_in_flight_fetch" : 0,
  "task_max_waiting_in_queue_millis" : 0,
  "active_shards_percent_as_number" : 100.0
}

3. Configure and deploy data nodes
Dry run the helm chart for the data nodes.
❯ helm install elasticsearch-data --dry-run  ./elasticsearch  -f ./elasticsearch/data-values.yaml -n hornet
NAME: elasticsearch-data
LAST DEPLOYED: Sat Aug 13 13:29:00 2022
NAMESPACE: hornet
STATUS: pending-install
REVISION: 1
HOOKS:
---
# Source: elasticsearch/templates/test/test-elasticsearch-health.yaml
apiVersion: v1
kind: Pod
metadata:
  name: "elasticsearch-data-wxtqt-test"
  annotations:
    "helm.sh/hook": test
    "helm.sh/hook-delete-policy": hook-succeeded
spec:
  securityContext:
    fsGroup: 1000
    runAsUser: 1000
  containers:
  - name: "elasticsearch-data-ysiwi-test"
    image: "docker.elastic.co/elasticsearch/elasticsearch:7.17.4"
    imagePullPolicy: "IfNotPresent"
    command:
      - "sh"
      - "-c"
      - |
        #!/usr/bin/env bash -e
        curl -XGET --fail 'elasticsearchuat-data:9200/_cluster/health?wait_for_status=green&timeout=1s'
  restartPolicy: Never
MANIFEST:
---
# Source: elasticsearch/templates/poddisruptionbudget.yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: "elasticsearchuat-data-pdb"
spec:
  maxUnavailable: 1
  selector:
    matchLabels:
      app: "elasticsearchuat-data"
---
# Source: elasticsearch/templates/service.yaml
kind: Service
apiVersion: v1
metadata:
  name: elasticsearchuat-data
  labels:
    heritage: "Helm"
    release: "elasticsearch-data"
    chart: "elasticsearch"
    app: "elasticsearchuat-data"
  annotations:
    {}
spec:
  type: ClusterIP
  selector:
    release: "elasticsearch-data"
    chart: "elasticsearch"
    app: "elasticsearchuat-data"
  publishNotReadyAddresses: false
  ports:
  - name: http
    protocol: TCP
    port: 9200
  - name: transport
    protocol: TCP
    port: 9300
---
# Source: elasticsearch/templates/service.yaml
kind: Service
apiVersion: v1
metadata:
  name: elasticsearchuat-data-headless
  labels:
    heritage: "Helm"
    release: "elasticsearch-data"
    chart: "elasticsearch"
    app: "elasticsearchuat-data"
  annotations:
    service.alpha.kubernetes.io/tolerate-unready-endpoints: "true"
spec:
  clusterIP: None # This is needed for statefulset hostnames like elasticsearch-0 to resolve
  # Create endpoints also if the related pod isn't ready
  publishNotReadyAddresses: true
  selector:
    app: "elasticsearchuat-data"
  ports:
  - name: http
    port: 9200
  - name: transport
    port: 9300
---
# Source: elasticsearch/templates/statefulset.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: elasticsearchuat-data
  labels:
    heritage: "Helm"
    release: "elasticsearch-data"
    chart: "elasticsearch"
    app: "elasticsearchuat-data"
  annotations:
    esMajorVersion: "7"
spec:
  serviceName: elasticsearchuat-data-headless
  selector:
    matchLabels:
      app: "elasticsearchuat-data"
  replicas: 3
  podManagementPolicy: Parallel
  updateStrategy:
    type: RollingUpdate
  volumeClaimTemplates:
  - metadata:
      name: elasticsearchuat-data
    spec:
      accessModes:
      - ReadWriteOnce
      resources:
        requests:
          storage: 30Gi
  template:
    metadata:
      name: "elasticsearchuat-data"
      labels:
        release: "elasticsearch-data"
        chart: "elasticsearch"
        app: "elasticsearchuat-data"
      annotations:


    spec:
      securityContext:
        fsGroup: 1000
        runAsUser: 1000
      automountServiceAccountToken: true
      affinity:
        podAntiAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
          - labelSelector:
              matchExpressions:
              - key: app
                operator: In
                values:
                - "elasticsearchuat-data"
            topologyKey: kubernetes.io/hostname
      terminationGracePeriodSeconds: 120
      volumes:
      enableServiceLinks: true
      containers:
      - name: "elasticsearch"
        securityContext:
          capabilities:
            drop:
            - ALL
          runAsNonRoot: true
          runAsUser: 1000
        image: "docker.elastic.co/elasticsearch/elasticsearch:7.17.4"
        imagePullPolicy: "IfNotPresent"
        readinessProbe:
          exec:
            command:
              - bash
              - -c
              - |
                set -e
                # If the node is starting up wait for the cluster to be ready (request params: "wait_for_status=green&timeout=1s" )
                # Once it has started only check that the node itself is responding
                START_FILE=/tmp/.es_start_file


                # Disable nss cache to avoid filling dentry cache when calling curl
                # This is required with Elasticsearch Docker using nss < 3.52
                export NSS_SDB_USE_CACHE=no


                http () {
                  local path="${1}"
                  local args="${2}"
                  set -- -XGET -s


                  if [ "$args" != "" ]; then
                    set -- "$@" $args
                  fi


                  if [ -n "${ELASTIC_PASSWORD}" ]; then
                    set -- "$@" -u "elastic:${ELASTIC_PASSWORD}"
                  fi


                  curl --output /dev/null -k "$@" "http://127.0.0.1:9200${path}"
                }


                if [ -f "${START_FILE}" ]; then
                  echo 'Elasticsearch is already running, lets check the node is healthy'
                  HTTP_CODE=$(http "/" "-w %{http_code}")
                  RC=$?
                  if [[ ${RC} -ne 0 ]]; then
                    echo "curl --output /dev/null -k -XGET -s -w '%{http_code}' \${BASIC_AUTH} http://127.0.0.1:9200/ failed with RC ${RC}"
                    exit ${RC}
                  fi
                  # ready if HTTP code 200, 503 is tolerable if ES version is 6.x
                  if [[ ${HTTP_CODE} == "200" ]]; then
                    exit 0
                  elif [[ ${HTTP_CODE} == "503" && "7" == "6" ]]; then
                    exit 0
                  else
                    echo "curl --output /dev/null -k -XGET -s -w '%{http_code}' \${BASIC_AUTH} http://127.0.0.1:9200/ failed with HTTP code ${HTTP_CODE}"
                    exit 1
                  fi


                else
                  echo 'Waiting for elasticsearch cluster to become ready (request params: "wait_for_status=green&timeout=1s" )'
                  if http "/_cluster/health?wait_for_status=green&timeout=1s" "--fail" ; then
                    touch ${START_FILE}
                    exit 0
                  else
                    echo 'Cluster is not yet ready (request params: "wait_for_status=green&timeout=1s" )'
                    exit 1
                  fi
                fi
          failureThreshold: 3
          initialDelaySeconds: 10
          periodSeconds: 10
          successThreshold: 3
          timeoutSeconds: 5
        ports:
        - name: http
          containerPort: 9200
        - name: transport
          containerPort: 9300
        resources:
          limits:
            cpu: 1000m
            memory: 2Gi
          requests:
            cpu: 1000m
            memory: 2Gi
        env:
          - name: node.name
            valueFrom:
              fieldRef:
                fieldPath: metadata.name
          - name: discovery.seed_hosts
            value: "elasticsearchuat-master-headless"
          - name: cluster.name
            value: "elasticsearchuat"
          - name: network.host
            value: "0.0.0.0"
          - name: cluster.deprecation_indexing.enabled
            value: "false"
          - name: ES_JAVA_OPTS
            value: "-Xmx1g -Xms1g"
          - name: node.data
            value: "true"
          - name: node.ingest
            value: "false"
          - name: node.master
            value: "false"
          - name: node.ml
            value: "false"
          - name: node.remote_cluster_client
            value: "false"
        volumeMounts:
          - name: "elasticsearchuat-data"
            mountPath: /usr/share/elasticsearch/data


NOTES:
1. Watch all cluster members come up.
  $ kubectl get pods --namespace=hornet -l app=elasticsearchuat-data -w2. Test cluster health using Helm test.
  $ helm --namespace=hornet test elasticsearch-data
4. Deploy the data nodes
❯ helm install elasticsearch-data  ./elasticsearch  -f ./elasticsearch/data-values.yaml -n hornet
NAME: elasticsearch-data
LAST DEPLOYED: Sat Aug 13 13:22:19 2022
NAMESPACE: hornet
STATUS: deployed
REVISION: 1
NOTES:
1. Watch all cluster members come up.
  $ kubectl get pods --namespace=hornet -l app=elasticsearchuat-data -w2. Test cluster health using Helm test.
  $ helm --namespace=hornet test elasticsearch-data

5. Deploy of the data nodes should be as shown:
❯ k get pod -n hornet
NAME                        READY   STATUS    RESTARTS   AGE
elasticsearchuat-data-0     1/1     Running   0          2m13s
elasticsearchuat-data-1     1/1     Running   0          2m13s
elasticsearchuat-data-2     1/1     Running   0          2m13s
elasticsearchuat-master-0   1/1     Running   0          23m
elasticsearchuat-master-1   1/1     Running   0          23m
elasticsearchuat-master-2   1/1     Running   0          23m
sh-5.0$  curl http://localhost:9200/_cluster/health?pretty
{
  "cluster_name" : "elasticsearchuat",
  "status" : "green",
  "timed_out" : false,
  "number_of_nodes" : 6,
  "number_of_data_nodes" : 3,
  "active_primary_shards" : 1,
  "active_shards" : 2,
  "relocating_shards" : 0,
  "initializing_shards" : 0,
  "unassigned_shards" : 0,
  "delayed_unassigned_shards" : 0,
  "number_of_pending_tasks" : 0,
  "number_of_in_flight_fetch" : 0,
  "task_max_waiting_in_queue_millis" : 0,
  "active_shards_percent_as_number" : 100.0

6. Deploy client tier :
❯ helm install elasticsearch-client --dry-run  ./elasticsearch  -f ./elasticsearch/client-values.yaml -n hornet
NAME: elasticsearch-client
LAST DEPLOYED: Sat Aug 13 13:44:43 2022
NAMESPACE: hornet
STATUS: pending-install
REVISION: 1
HOOKS:
---
# Source: elasticsearch/templates/test/test-elasticsearch-health.yaml
apiVersion: v1
kind: Pod
metadata:
  name: "elasticsearch-client-gmhfz-test"
  annotations:
    "helm.sh/hook": test
    "helm.sh/hook-delete-policy": hook-succeeded
spec:
  securityContext:
    fsGroup: 1000
    runAsUser: 1000
  containers:
  - name: "elasticsearch-client-wrrfp-test"
    image: "docker.elastic.co/elasticsearch/elasticsearch:7.17.4"
    imagePullPolicy: "IfNotPresent"
    command:
      - "sh"
      - "-c"
      - |
        #!/usr/bin/env bash -e
        curl -XGET --fail 'elasticsearchuat-client:9200/_cluster/health?wait_for_status=green&timeout=1s'
  restartPolicy: Never
MANIFEST:
---
# Source: elasticsearch/templates/poddisruptionbudget.yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: "elasticsearchuat-client-pdb"
spec:
  maxUnavailable: 1
  selector:
    matchLabels:
      app: "elasticsearchuat-client"
---
# Source: elasticsearch/templates/service.yaml
kind: Service
apiVersion: v1
metadata:
  name: elasticsearchuat-client
  labels:
    heritage: "Helm"
    release: "elasticsearch-client"
    chart: "elasticsearch"
    app: "elasticsearchuat-client"
  annotations:
    {}
spec:
  type: LoadBalancer
  selector:
    release: "elasticsearch-client"
    chart: "elasticsearch"
    app: "elasticsearchuat-client"
  publishNotReadyAddresses: false
  ports:
  - name: http
    protocol: TCP
    port: 9200
  - name: transport
    protocol: TCP
    port: 9300
---
# Source: elasticsearch/templates/service.yaml
kind: Service
apiVersion: v1
metadata:
  name: elasticsearchuat-client-headless
  labels:
    heritage: "Helm"
    release: "elasticsearch-client"
    chart: "elasticsearch"
    app: "elasticsearchuat-client"
  annotations:
    service.alpha.kubernetes.io/tolerate-unready-endpoints: "true"
spec:
  clusterIP: None # This is needed for statefulset hostnames like elasticsearch-0 to resolve
  # Create endpoints also if the related pod isn't ready
  publishNotReadyAddresses: true
  selector:
    app: "elasticsearchuat-client"
  ports:
  - name: http
    port: 9200
  - name: transport
    port: 9300
---
# Source: elasticsearch/templates/statefulset.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: elasticsearchuat-client
  labels:
    heritage: "Helm"
    release: "elasticsearch-client"
    chart: "elasticsearch"
    app: "elasticsearchuat-client"
  annotations:
    esMajorVersion: "7"
spec:
  serviceName: elasticsearchuat-client-headless
  selector:
    matchLabels:
      app: "elasticsearchuat-client"
  replicas: 3
  podManagementPolicy: Parallel
  updateStrategy:
    type: RollingUpdate
  volumeClaimTemplates:
  - metadata:
      name: elasticsearchuat-client
    spec:
      accessModes:
      - ReadWriteOnce
      resources:
        requests:
          storage: 30Gi
  template:
    metadata:
      name: "elasticsearchuat-client"
      labels:
        release: "elasticsearch-client"
        chart: "elasticsearch"
        app: "elasticsearchuat-client"
      annotations:


    spec:
      securityContext:
        fsGroup: 1000
        runAsUser: 1000
      automountServiceAccountToken: true
      affinity:
        podAntiAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
          - labelSelector:
              matchExpressions:
              - key: app
                operator: In
                values:
                - "elasticsearchuat-client"
            topologyKey: kubernetes.io/hostname
      terminationGracePeriodSeconds: 120
      volumes:
      enableServiceLinks: true
      containers:
      - name: "elasticsearch"
        securityContext:
          capabilities:
            drop:
            - ALL
          runAsNonRoot: true
          runAsUser: 1000
        image: "docker.elastic.co/elasticsearch/elasticsearch:7.17.4"
        imagePullPolicy: "IfNotPresent"
        readinessProbe:
          exec:
            command:
              - bash
              - -c
              - |
                set -e
                # If the node is starting up wait for the cluster to be ready (request params: "wait_for_status=green&timeout=1s" )
                # Once it has started only check that the node itself is responding
                START_FILE=/tmp/.es_start_file


                # Disable nss cache to avoid filling dentry cache when calling curl
                # This is required with Elasticsearch Docker using nss < 3.52
                export NSS_SDB_USE_CACHE=no


                http () {
                  local path="${1}"
                  local args="${2}"
                  set -- -XGET -s


                  if [ "$args" != "" ]; then
                    set -- "$@" $args
                  fi


                  if [ -n "${ELASTIC_PASSWORD}" ]; then
                    set -- "$@" -u "elastic:${ELASTIC_PASSWORD}"
                  fi


                  curl --output /dev/null -k "$@" "http://127.0.0.1:9200${path}"
                }


                if [ -f "${START_FILE}" ]; then
                  echo 'Elasticsearch is already running, lets check the node is healthy'
                  HTTP_CODE=$(http "/" "-w %{http_code}")
                  RC=$?
                  if [[ ${RC} -ne 0 ]]; then
                    echo "curl --output /dev/null -k -XGET -s -w '%{http_code}' \${BASIC_AUTH} http://127.0.0.1:9200/ failed with RC ${RC}"
                    exit ${RC}
                  fi
                  # ready if HTTP code 200, 503 is tolerable if ES version is 6.x
                  if [[ ${HTTP_CODE} == "200" ]]; then
                    exit 0
                  elif [[ ${HTTP_CODE} == "503" && "7" == "6" ]]; then
                    exit 0
                  else
                    echo "curl --output /dev/null -k -XGET -s -w '%{http_code}' \${BASIC_AUTH} http://127.0.0.1:9200/ failed with HTTP code ${HTTP_CODE}"
                    exit 1
                  fi


                else
                  echo 'Waiting for elasticsearch cluster to become ready (request params: "wait_for_status=green&timeout=1s" )'
                  if http "/_cluster/health?wait_for_status=green&timeout=1s" "--fail" ; then
                    touch ${START_FILE}
                    exit 0
                  else
                    echo 'Cluster is not yet ready (request params: "wait_for_status=green&timeout=1s" )'
                    exit 1
                  fi
                fi
          failureThreshold: 3
          initialDelaySeconds: 10
          periodSeconds: 10
          successThreshold: 3
          timeoutSeconds: 5
        ports:
        - name: http
          containerPort: 9200
        - name: transport
          containerPort: 9300
        resources:
          limits:
            cpu: 1000m
            memory: 2Gi
          requests:
            cpu: 1000m
            memory: 2Gi
        env:
          - name: node.name
            valueFrom:
              fieldRef:
                fieldPath: metadata.name
          - name: discovery.seed_hosts
            value: "elasticsearchuat-master-headless"
          - name: cluster.name
            value: "elasticsearchuat"
          - name: network.host
            value: "0.0.0.0"
          - name: cluster.deprecation_indexing.enabled
            value: "false"
          - name: ES_JAVA_OPTS
            value: "-Xmx1g -Xms1g"
          - name: node.data
            value: "false"
          - name: node.ingest
            value: "false"
          - name: node.master
            value: "false"
          - name: node.ml
            value: "false"
          - name: node.remote_cluster_client
            value: "false"
        volumeMounts:
          - name: "elasticsearchuat-client"
            mountPath: /usr/share/elasticsearch/data


NOTES:
1. Watch all cluster members come up.
  $ kubectl get pods --namespace=hornet -l app=elasticsearchuat-client -w2. Test cluster health using Helm test.
  $ helm --namespace=hornet test elasticsearch-client


Disable the persistence for the clients.

❯ helm install elasticsearch-client  ./elasticsearch  -f ./elasticsearch/client-values.yaml -n hornet
NAME: elasticsearch-client
LAST DEPLOYED: Sat Aug 13 13:46:18 2022
NAMESPACE: hornet
STATUS: deployed
REVISION: 1
NOTES:
1. Watch all cluster members come up.
  $ kubectl get pods --namespace=hornet -l app=elasticsearchuat-client -w2. Test cluster health using Helm test.
  $ helm --namespace=hornet test elasticsearch-client
❯ k get pods -n hornet
NAME                        READY   STATUS    RESTARTS   AGE
elasticsearchuat-client-0   1/1     Running   0          2m1s
elasticsearchuat-client-1   1/1     Running   0          2m1s
elasticsearchuat-client-2   1/1     Running   0          2m1s
elasticsearchuat-data-0     1/1     Running   0          17m
elasticsearchuat-data-1     1/1     Running   0          17m
elasticsearchuat-data-2     1/1     Running   0          17m
elasticsearchuat-master-0   1/1     Running   0          38m
elasticsearchuat-master-1   1/1     Running   0          38m
elasticsearchuat-master-2   1/1     Running   0          38m

7. Deploy kabana.

❯ helm install elasticsearch-kibana --dry-run ./kibana -f ./kibana/values.yaml -n hornet
NAME: elasticsearch-kibana
LAST DEPLOYED: Sat Aug 13 14:00:12 2022
NAMESPACE: hornet
STATUS: pending-install
REVISION: 1
TEST SUITE: None
HOOKS:
MANIFEST:
---
# Source: kibana/templates/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: elasticsearch-kibana-kibana
  labels:
    app: kibana
    release: "elasticsearch-kibana"
    heritage: Helm
spec:
  type: LoadBalancer
  ports:
    - port: 5601
      protocol: TCP
      name: http
      targetPort: 5601
  selector:
    app: kibana
    release: "elasticsearch-kibana"
---
# Source: kibana/templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: elasticsearch-kibana-kibana
  labels:
    app: kibana
    release: "elasticsearch-kibana"
    heritage: Helm
spec:
  replicas: 1
  strategy:
    type: Recreate
  selector:
    matchLabels:
      app: kibana
      release: "elasticsearch-kibana"
  template:
    metadata:
      labels:
        app: kibana
        release: "elasticsearch-kibana"
      annotations:


    spec:
      automountServiceAccountToken: true
      securityContext:
        fsGroup: 1000
      volumes:
      containers:
      - name: kibana
        securityContext:
          capabilities:
            drop:
            - ALL
          runAsNonRoot: true
          runAsUser: 1000
        image: "docker.elastic.co/kibana/kibana:7.17.3"
        imagePullPolicy: "IfNotPresent"
        env:
          - name: ELASTICSEARCH_HOSTS
            value: "http://elasticsearchuat-master:9200"
          - name: SERVER_HOST
            value: "0.0.0.0"
          - name: NODE_OPTIONS
            value: --max-old-space-size=1800
        readinessProbe:
          failureThreshold: 3
          initialDelaySeconds: 10
          periodSeconds: 10
          successThreshold: 3
          timeoutSeconds: 5
          exec:
            command:
              - bash
              - -c
              - |
                #!/usr/bin/env bash -e


                # Disable nss cache to avoid filling dentry cache when calling curl
                # This is required with Kibana Docker using nss < 3.52
                export NSS_SDB_USE_CACHE=no


                http () {
                    local path="${1}"
                    set -- -XGET -s --fail -L


                    if [ -n "${ELASTICSEARCH_USERNAME}" ] && [ -n "${ELASTICSEARCH_PASSWORD}" ]; then
                      set -- "$@" -u "${ELASTICSEARCH_USERNAME}:${ELASTICSEARCH_PASSWORD}"
                    fi


                    STATUS=$(curl --output /dev/null --write-out "%{http_code}" -k "$@" "http://localhost:5601${path}")
                    if [[ "${STATUS}" -eq 200 ]]; then
                      exit 0
                    fi


                    echo "Error: Got HTTP code ${STATUS} but expected a 200"
                    exit 1
                }


                http "/app/kibana"
        ports:
        - containerPort: 5601
        resources:
          limits:
            cpu: 1000m
            memory: 2Gi
          requests:
            cpu: 1000m
            memory: 2Gi
        volumeMounts:

Deploy the kibana.
❯ helm install elasticsearch-kibana  ./kibana -f ./kibana/values.yaml -n hornet
NAME: elasticsearch-kibana
LAST DEPLOYED: Sat Aug 13 14:02:36 2022
NAMESPACE: hornet
STATUS: deployed
REVISION: 1
TEST SUITE: None

❯ k get pods -n hornet
NAME                                           READY   STATUS    RESTARTS   AGE
elasticsearch-kibana-kibana-58c7b97d5b-xzhl9   1/1     Running   0          87s
elasticsearchuat-client-0                      1/1     Running   0          17m
elasticsearchuat-client-1                      1/1     Running   0          17m
elasticsearchuat-client-2                      1/1     Running   0          17m
elasticsearchuat-data-0                        1/1     Running   0          33m
elasticsearchuat-data-1                        1/1     Running   0          33m
elasticsearchuat-data-2                        1/1     Running   0          33m
elasticsearchuat-master-0                      1/1     Running   0          54m
elasticsearchuat-master-1                      1/1     Running   0          54m
elasticsearchuat-master-2                      1/1     Running   0          54m

```
