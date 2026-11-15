# VictoriaMetrics ecosystem, state of the art as of 2026-11-13

Target: small k3s platform, 2 nodes, 28 GB RAM, 8 vCPU, Ceph RBD plus Ceph RGW S3. About 70 pods, up to 200k active series, 5 to 20 GB logs/day.
Confidence tags: [HIGH] primary source read in full or in relevant part, [MED] secondary or vendor benchmark, [LOW] inferred or estimate.
Data older than 12 months is flagged [STALE].
All GitHub figures below were read from the GitHub API or release pages on 2026-11-13.

## 0. Executive conclusions

1. Use the operator, installed through victoria-metrics-k8s-stack 0.98.1 (appVersion v1.156.0), with VMSingle, not VMCluster. Your multitenancy is already label based through vmauth extra_filters, so VMCluster brings no isolation gain at this scale. [HIGH]
2. VictoriaLogs single-node is production ready (GA since v1.0.0, current v1.54.0 released 2026-11-13). It cannot use S3 yet and cannot do per-stream or per-tenant retention. Both are real regressions against your Loki setup. [HIGH]
3. VictoriaTraces is still pre-GA (v0.13.0, 2026-11-05). Its roadmap says the data structure is not yet frozen. Usable on a tiny cluster, but treat it as beta. [HIGH]
4. Replace Alloy with vmagent for metrics and vlagent for logs. Zero-code traces and RED metrics: OBI v1.0.0 (stable, 2026-10-30). [HIGH]
5. Enterprise-only items that matter to you: retention filters, downsampling, vmanomaly, multi-tenant vmalert and VMRule tenant field, vmbackupmanager, vmauth IP filters. Everything else you need is open source. [HIGH]
6. Realistic RAM for the whole target stack is about 4 to 6 GB, against roughly 1.8 GB for Alloy alone today. See section 8. Most figures are estimates and must be measured. [LOW]

## 1. Operator vs Helm charts, k8s-stack, VMSingle vs VMCluster, tenancy, OSS vs Enterprise

### 1.1 Operator status
- Operator latest v0.75.2 (released 28 Oct 2026). v0.75.0 released 25 Sep 2026. Helm chart victoria-metrics-operator 0.68.2 (2026-10-28). Repo 600 stars, Apache-2.0. [HIGH]
  https://github.com/VictoriaMetrics/operator/releases
- CRDs listed in the docs: VMSingle, VMCluster, VMAgent, VMAlert, VMAlertmanager, VMAlertmanagerConfig, VMAuth, VMUser, VMRule, VMServiceScrape, VMPodScrape, VMNodeScrape, VMStaticScrape, VMScrapeConfig, VMProbe, VMAnomaly, VLSingle, VLCluster, VLAgent, VTSingle, VTCluster, VTAgent, VMDistributed, plus VLDistributed. [HIGH]
  https://docs.victoriametrics.com/operator/resources/
- New in 2026: VLDistributed (v0.74.0), VTAgent (v0.75.0). [HIGH]
- Breaking or update notes to read before upgrading: v0.75.0 config-reloader argument change rolls many pods once. VMCluster vmstorage serviceSpec that makes the Service non-headless is rejected. VTAgent needs extra RBAC or the operator shuts down. v0.74.0 needs NetworkPolicy permissions in the ClusterRole and disables prometheus-operator validation webhooks by default. [HIGH]
- VMAgent has daemonSetMode since operator v0.55.0, but it supports only VMPodScrape and cannot combine with statefulMode. Not useful for you, a single Deployment replica is enough at 200k series. [HIGH]
  https://github.com/VictoriaMetrics/operator/blob/master/docs/api.md (field daemonSetMode)
- VMAnomaly CRD (since operator v0.60.0) is Enterprise only: "VMAnomaly is enterprise only, license is required for this CRD". [HIGH]
  https://docs.victoriametrics.com/operator/resources/vmanomaly/
- VMAlertmanagerConfig has an enforced namespace matcher: alerts need a namespace label equal to the config's namespace, unless VMAlertmanager spec.disableNamespaceMatcher is true. Useful for per-project routing to Discord. [HIGH]
  https://github.com/VictoriaMetrics/operator/blob/master/docs/resources/vmalertmanagerconfig.md
- VMRule tenant field is Enterprise only. [HIGH]
  https://docs.victoriametrics.com/operator/enterprise/

### 1.2 Operator vs plain charts
- The operator buys you CRD-driven config: VMServiceScrape/VMPodScrape/VMNodeScrape/VMProbe, VMRule per namespace, VMUser per project, VMAlertmanagerConfig per namespace. These match the "project = set of namespaces" model, and prometheus-operator CRDs can be converted automatically (docs list a conversion path). [HIGH]
- Plain charts (victoria-metrics-single 0.51.0, victoria-metrics-cluster 0.55.0, victoria-metrics-agent 0.52.0, victoria-metrics-alert 0.53.0, victoria-metrics-auth 0.46.0, victoria-logs-single 0.13.12, victoria-traces-single 0.1.13) were all published 2026-11-09. They remain maintained. [HIGH]
  https://github.com/VictoriaMetrics/helm-charts/releases
- Operator cost: one extra controller pod. Memory not measured. [LOW]
- Recommendation: operator, because per-project VMUser, VMRule and scrape CRDs are exactly the self-service surface an IDP needs.

### 1.3 victoria-metrics-k8s-stack 0.98.1 contents (read from Chart.yaml and values.yaml)
- appVersion v1.156.0, kubeVersion >= 1.25. Dependencies: victoria-metrics-operator 0.68.*, kube-state-metrics 9.0.*, prometheus-node-exporter 4.60.*, grafana 14.4.* from the grafana-community chart repo. [HIGH]
  https://raw.githubusercontent.com/VictoriaMetrics/helm-charts/master/charts/victoria-metrics-k8s-stack/Chart.yaml
- Enabled by default: vmsingle (retention 1 month, 20Gi RWO PVC), vmagent, vmalert, alertmanager, grafana, defaultRules, defaultDashboards via a sync job, scrape of kubelet, kubeApiServer, kubeControllerManager, kubeEtcd, kubeScheduler. Disabled by default: vmcluster, vmauth, vlsingle, vlcluster, vlagent, vtsingle, vtcluster, kubeProxy. [HIGH]
  https://raw.githubusercontent.com/VictoriaMetrics/helm-charts/master/charts/victoria-metrics-k8s-stack/values.yaml
- The chart can provision Grafana datasources for VictoriaMetrics (prometheus type and victoriametrics-metrics-datasource), VictoriaLogs (victoriametrics-logs-datasource) and VictoriaTraces (jaeger type). The plugins must be listed in grafana.plugins, and Grafana needs internet access to install them. [HIGH]
- Caveats for your setup:
  - You run your own Grafana 12.3 with one org per project. The chart's dashboard sidecar provider is bound to orgid 1, so default dashboards land in org 1 only. Disable the bundled grafana and provision dashboards per org yourself. [LOW, inferred from values.yaml provider.orgid: 1]
  - kubeControllerManager, kubeEtcd and kubeScheduler scrapes are on by default. k3s embeds these in one process and does not expose them as separate pods, so expect down targets and noisy default alerts unless you disable them. [LOW, inferred]
  - kube-state-metrics and node-exporter are bundled, you can drop your standalone copies.

### 1.4 VMSingle vs VMCluster
- Vendor guidance: use single-node below one million data points per second. Single node needs less CPU and RAM than cluster for the same workload, is easier to operate, and supports HA. [HIGH]
  https://docs.victoriametrics.com/victoriametrics/cluster-victoriametrics/ and https://docs.victoriametrics.com/victoriametrics/faq/
- The FAQ lists cluster as preferable mainly for native multitenancy or for workloads beyond what one node handles. It also states multitenancy can be done with labels and vmauth enforced query args on single node. [HIGH]
- Your load is 200k active series, which is about 0.2 percent of the 100M series single-node scale claim in the docs. [HIGH]
- Migration cluster to single: data formats differ. The documented path is to run both in parallel until old data ages out, or use vmctl native export/import. The docs describe single to cluster explicitly. Cluster to single is the reverse via vmctl. [MED]
  https://docs.victoriametrics.com/victoriametrics/faq/
- Recommendation: VMSingle, one replica, PVC on Ceph RBD, 30 to 90 day retention. Backups with vmbackup to RGW S3 (open source, supports any S3 compatible store including Ceph). [HIGH]
  https://docs.victoriametrics.com/victoriametrics/vmbackup/

### 1.5 Multitenancy options
| Option | Needs | Works on | Notes |
|---|---|---|---|
| Label filtering with vmauth extra_filters and extra_label | vmauth user per project | VMSingle and VMCluster | Open source. Your current model. Isolation is only as strong as the label provenance, enforce at both write and read. [HIGH] |
| accountID/projectID tenants | VMCluster | VMCluster only | Single node has no multitenancy. Headers AccountID and ProjectID work since v1.143.0, on by default since v1.150.0. [HIGH] |
| vm_account_id and vm_project_id labels with the multitenant endpoints | VMCluster | VMCluster | Lets one writer or reader span tenants. Restrict access to these endpoints. [HIGH] |
| Per-tenant retention, per-tenant stats, per-tenant rate limits | Enterprise | VMCluster | -retentionFilter with vm_account_id, vmgateway. [HIGH] |

Sources: https://docs.victoriametrics.com/victoriametrics/cluster-victoriametrics/ , https://docs.victoriametrics.com/victoriametrics/vmauth/
- vmauth templating from JWT claims (metrics_extra_filters, metrics_extra_labels, logs_extra_filters, logs_extra_stream_filters) is open source and exists since v1.137.0. OIDC discovery, JWT claim matching and deny_paths are open source. So Grafana per-org datasources can use a JWT or basic-auth user per project. [HIGH]
- Project label at write time: vmagent kubernetes_sd supports attach_metadata namespace true, exposing __meta_kubernetes_namespace_label_<name>. You can stamp a project label from a namespace label during relabeling. [HIGH]
  https://docs.victoriametrics.com/victoriametrics/sd_configs/
- Cardinality control: cardinality explorer in vmui, vmagent -remoteWrite.maxHourlySeries and maxDailySeries, per-scrape series_limit and sample_limit, vmstorage or VMSingle -storage.maxHourlySeries and -storage.maxDailySeries. These limits are global. Per-tenant limits are Enterprise (vmgateway). [HIGH for flags, MED for the per-tenant claim]
  https://docs.victoriametrics.com/victoriametrics/vmagent/ and https://docs.victoriametrics.com/victoriametrics/enterprise/
- Per-project alert rules with one open source vmalert: groups can carry params with extra_label or extra_filters for the project, or run one vmalert per tenant. The group tenant field and -clusterMode are Enterprise. [HIGH]
  https://docs.victoriametrics.com/victoriametrics/vmalert/

### 1.6 Open source vs Enterprise (VictoriaMetrics)
Enterprise only: downsampling, retention filters (multiple retentions), vmbackupmanager, vmanomaly, advanced per-tenant stats, query execution stats, vmgateway, mTLS and auto TLS, vmauth mTLS routing and IP filters, vmalert multitenancy and reading rules from object storage, Kafka and PubSub integration, vmstorage auto discovery, FIPS builds. [HIGH]
https://docs.victoriametrics.com/victoriametrics/enterprise/
Open source and relevant: vmbackup and vmrestore to S3, vmctl, stream aggregation and deduplication, vmalert, vmalert-tool, cardinality explorer, vmauth including JWT and OIDC, vmagent cardinality limiter, relabeling debugger. [HIGH]
Licensing changes: from v1.156.0 (2026-11-09) Enterprise components take an -eula flag and license keys are no longer verified. Production use still requires an enterprise contract. Evaluation use is allowed non-production. [HIGH]
LTS lines are published for Enterprise only. Currently v1.148.x and v1.136.x. Open source users are told to track the latest release. [HIGH]
https://docs.victoriametrics.com/victoriametrics/lts-releases/
Implication: you cannot get VM-native per-tenant retention, so give every project the same retention, or run a second VMSingle for long-lived data and route with vmauth.

### 1.7 Version and bug watchlist
- v1.154.0 carried an OOM risk on long-range series and label queries and connection churn with -search.skipSlowReplicas. Fixed in v1.155.0 and v1.156.0. v1.154.0 also fixed a vmauth token brute-force issue and a vmui XSS. v1.152.0 fixed a vmauth JWT authorization bypass. v1.151.0 fixed an HTTP Basic Auth bypass on suffixed paths. Run at least v1.156.0 when exposing vmauth. [HIGH]
  https://github.com/VictoriaMetrics/VictoriaMetrics/releases
- v1.155.0: /admin/tenants returns only the current day without start and end. v1.150.0: -enableMultitenancyViaHeaders defaults to true. v1.148.1: delete_series is POST only. [HIGH]

## 2. VictoriaLogs

### 2.1 Maturity and versions
- VictoriaLogs repo created 2025-07-06 (split out from the main repo), 2500 stars, Apache-2.0. Latest v1.54.0 (2026-11-13), v1.53.0 (2026-10-01), LTS line v1.52.x (v1.52.2, Enterprise-targeted). [HIGH]
  https://github.com/VictoriaMetrics/VictoriaLogs/releases
- FAQ: "Yes. VictoriaLogs is ready for production use starting from v1.0.0." [HIGH]
  https://docs.victoriametrics.com/victorialogs/faq/
- Release cadence is fast and includes breaking changes within minor versions. Examples: v1.51.0 rejects filter pipes without the filter prefix, v1.53.0 requires POST on /internal/*, mixed-version cluster queries fail during rolling upgrades in several releases. Pin versions and read the changelog on each bump. [HIGH]
  https://docs.victoriametrics.com/victorialogs/changelog/

### 2.2 Single vs cluster
- Docs: use single-node if one host can handle the load. It is simpler and faster on the same hardware. Cluster mode (vlinsert, vlselect, vlstorage, same binary with -storageNode) only for vertical scaling limits. Cluster shards, it does not replicate. [HIGH]
  https://docs.victoriametrics.com/victorialogs/cluster/
- At 5 to 20 GB/day, single node. [HIGH on the guidance, LOW on the exact threshold]

### 2.3 Storage, S3, retention
- Local disk only today. Data lives in per-day partitions under one directory. [HIGH]
  https://docs.victoriametrics.com/victorialogs/
- Object storage: still on the roadmap. Issue #48 open. PR #1155 "introduce object storage support" open. On 2026-10-08 valyala wrote the merge ETA is November 2026. Design: per-day partitions older than -offloadPeriod (minimum one day) move to object storage, recent data stays local, moved data remains queryable. [HIGH]
  https://github.com/VictoriaMetrics/VictoriaLogs/issues/48
  Open bug reports on the PR build: on 2026-11-02 a user reported queries on data older than 7 days crash the process on any S3 error and stop ingestion, and an interrupted offload can cause a restart loop until the day folder is removed by hand. On 2026-11-05 the author said an unavailable object store causes the restart loop because startup treats a failed remote listing as fatal. Do not plan on S3 offload before it ships and stabilises. [HIGH, dates as quoted]
- Backups: snapshots plus rclone to S3 or GCS are documented. Native backup tooling is a roadmap item (issue #123). [HIGH]
  https://raw.githubusercontent.com/VictoriaMetrics/VictoriaLogs/master/docs/victorialogs/Roadmap.md
- Retention: global -retentionPeriod (default 7d) and disk-based -retention.maxDiskSpaceUsageBytes or -retention.maxDiskUsagePercent, mutually exclusive. No per-tenant, per-stream or per-level retention flag exists in the open source flag list or in the Enterprise flag list. [HIGH]
  https://raw.githubusercontent.com/VictoriaMetrics/VictoriaLogs/master/docs/victorialogs/victoria_logs_enterprise_flags.md
- Issue #15 "retentionPeriod for different streams" is closed (last update 2025-12-05). The maintainer position: VM Enterprise retention filters were "a mistake" because of reliability and resource cost. The supported pattern is several independent VictoriaLogs instances with different retentions, fronted by vmauth routing, plus the victoria-logs-multilevel Helm chart (0.2.10 as of 2026-11-02) for querying them together. [HIGH]
  https://github.com/VictoriaMetrics/VictoriaLogs/issues/15
  https://docs.victoriametrics.com/victorialogs/security-and-lb/
- Consequence for you: your Loki per-stream retention rules must become two or three VLSingle instances (for example 7d, 30d, 90d), with the collector or vmauth deciding the destination. Each extra instance costs RAM. [LOW on cost]

### 2.4 Multitenancy
- Tenant is (AccountID, ProjectID), 32-bit each, set by request headers on ingest and query. Thousands of tenants per instance is described as fine. VictoriaLogs does no authorization itself, use vmauth. [HIGH]
- vmauth can force headers per user, route by tenant header, and apply extra_filters and extra_stream_filters query args to restrict access. Both args are documented on the query API. [HIGH]
  https://raw.githubusercontent.com/VictoriaMetrics/VictoriaLogs/master/docs/victorialogs/querying/README.md
- vlagent Kubernetes collector: documented tenant selection is per remoteWrite through -remoteWrite.headers, or the /insert/multitenant/native endpoint for mixed tenants arriving via ingestion protocols. I found no documented per-namespace tenant routing in the Kubernetes collector. Log transformation in vlagent is still a roadmap item (#858). [HIGH for what is documented, LOW for the absence claim]
- Practical design: ship everything to one tenant and enforce project isolation in vmauth with extra_stream_filters on kubernetes.pod_namespace, which is a default stream field. This mirrors your current extra_filters model for metrics. [MED, design inference from docs]
- Per-tenant stats and quotas on ingest rate, stream cardinality and query cost are Enterprise roadmap items, not shipped. [HIGH]

### 2.5 Query language and Grafana
- LogsQL. LogQL conversion guide and playground exist; no migration tool yet (roadmap issue #521). I found no native Loki query API in VictoriaLogs; a third-party Loki-VL-proxy exists but I did not verify it. [HIGH for the guide, LOW for the proxy]
  https://docs.victoriametrics.com/victorialogs/logql-to-logsql/
- Grafana plugin victoriametrics-logs-datasource: v0.34.0 (2026-11-06), Apache-2.0, 204 stars, has a Drilldown view. Grafana's Logs Drilldown for Loki does not apply. [HIGH]
  https://github.com/VictoriaMetrics/victorialogs-datasource
- Rules on logs: vmalert with rule type vlogs, enabled per group or with -rule.defaultRuleType=vlogs. v1.51.0 added limit and offset pipes in stats_query for rules. Rules on logs and metrics need separate vmalert instances or typed groups. [HIGH]
  https://docs.victoriametrics.com/victorialogs/vmalert/

### 2.6 Ingestion protocols and collectors
- HTTP endpoints: Elasticsearch bulk /insert/elasticsearch/_bulk, JSON lines /insert/jsonline, Loki push /insert/loki/api/v1/push, OTLP /insert/opentelemetry/v1/logs. Also Syslog, Filebeat, Fluent Bit, Fluentd, Logstash, Vector, Promtail, Alloy, Telegraf, OTel Collector, Journald, Datadog, Splunk (Splunk added v1.50.0). [HIGH]
  https://docs.victoriametrics.com/victorialogs/data-ingestion/
- vlagent: log collector with Kubernetes pod log discovery (-kubernetesCollector, DaemonSet), file collector by glob, replication, disk buffering. Deployed by the victoria-logs-collector chart (0.3.8) or the VLAgent CRD. No journald collector in vlagent docs I read, use journald forwarding or Fluent Bit for node journals. [HIGH, journald absence LOW]
  https://docs.victoriametrics.com/victorialogs/vlagent/
- vmagent does not collect logs. [HIGH by omission in vmagent docs]

### 2.7 Benchmarks, all vendor or single-author, treat as directional
| Source | Date | Conditions | Result |
|---|---|---|---|
| VictoriaMetrics docs | current | not stated | "up to 30x less RAM, up to 15x less disk than Elasticsearch and Loki" [MED, vendor claim] |
| TrueFoundry author, dev.to mirror of harshit.cloud post | 2025-11-19 | 4 vCPU 8 GiB nodes, about 500 GB over 7 days, 65 MB/s via flog and Vector, block cache off, versions unstated | Loki 6 to 7 GB RAM vs about 1.3 GB, disk 501 GB vs 318 GB, needle query 12 s vs 0.9 s. Internal inconsistencies noted in the post [MED] |
https://dev.to/sachincool/victorialogs-vs-loki-real-world-benchmarking-results-26bg
- Your volume is far below those test loads, so absolute numbers will not transfer. The FAQ advises ingesting 1 to 10 percent of production logs and measuring. [HIGH]

### 2.8 Migration path from Loki
1. Deploy VLSingle (or several, see 2.3) and a vlagent DaemonSet. Dual-ship during the longest retention window, or point Alloy loki.write at the VictoriaLogs Loki endpoint for a quick start.
2. Rewrite dashboards and alerts from LogQL to LogsQL, Loki ruler rules to vmalert vlogs groups.
3. Keep Loki read-only until its retention expires, then remove it and its RGW bucket.
4. Disk: logs compress about 10x or more per the docs, so 5 to 20 GB/day raw is roughly 0.5 to 2 GB/day on disk. 30 days is about 15 to 60 GB on RBD. [MED]
Steps are my synthesis of the cited docs. [LOW]

## 3. VictoriaTraces

- v0.13.0 (2026-11-05). Repo 521 stars, Apache-2.0, created 2025-07-07. Releases are 0.x. Roadmap states GA requires "Finalize the data structure and commit to backward compatibility". [HIGH]
  https://raw.githubusercontent.com/VictoriaMetrics/VictoriaTraces/master/docs/victoriatraces/roadmap.md
- Breaking behaviours already seen: v0.9.4 changed the internal vtselect/vtstorage protocol so mixed versions fail, v0.12.0 replaced the UI and required POST on some internal endpoints. [HIGH]
- Ingestion: OTLP over HTTP at /insert/opentelemetry/v1/traces and OTLP/gRPC. vtagent exists (operator CRD VTAgent). Multitenancy through AccountID and ProjectID headers, auth via vmauth. [HIGH]
  https://docs.victoriametrics.com/victoriatraces/data-ingestion/
- Query: Jaeger Query Service JSON API, Tempo HTTP API marked experimental, TraceQL fuzzy and regex matching appears in the v0.9.1 release notes. Grafana uses the Jaeger datasource. [HIGH]
- Service graph: /select/jaeger/api/dependencies is experimental and needs -servicegraph.enableTask=true. [HIGH]
  https://raw.githubusercontent.com/VictoriaMetrics/VictoriaTraces/master/docs/victoriatraces/querying/README.md
- RED metrics derived from traces: not found in the VictoriaTraces docs. Use OBI span_metrics and service_graph_metrics, or the OTel Collector spanmetrics connector, and send the result to vmagent. [LOW on absence, HIGH that OBI offers it, see section 5]
- Storage: local disk only, retention global (-retentionPeriod default 7d), no S3, same per-day partition design as VictoriaLogs. [HIGH]
- Alerting through vmalert. [HIGH]
- Benchmark [STALE - 2025-06-09], vendor, proof of concept predating VictoriaTraces: at 10k spans/s with 4 CPU and 8 GiB, VictoriaLogs-based tracing used 0.50 vCPU and 1.15 GiB, ClickHouse 0.69 vCPU and 1.12 GiB, Tempo 1.35 vCPU and 4.26 GiB and ran out of memory. Jaeger plus Elasticsearch was unstable above about 5,000 spans/s. Versions not stated. [MED]
  https://victoriametrics.com/blog/dev-note-distributed-tracing-with-victorialogs/

Comparison for a tiny cluster
| | VictoriaTraces 0.13 | Tempo 3.1.1 (2026-10-21) | Jaeger 2.23.0 (2026-11-05) |
|---|---|---|---|
| Maturity | pre-GA | GA, AGPL-3.0 | GA, Apache-2.0 |
| Storage | local disk | object storage, fits your RGW | Badger (dev scale), Elasticsearch, Cassandra, ClickHouse experimental per one third-party source [LOW] |
| Footprint | lowest in vendor test | monolithic mode, no Kafka needed in 3.0 | collector plus query plus a store |
| Query | Jaeger API, Tempo API experimental | TraceQL, TraceQL metrics GA in 3.0 | Jaeger UI |
| Multitenancy | headers via vmauth | native tenant header | limited |
Tempo 3.0 microservices mode requires a Kafka-compatible system, monolithic mode does not. [HIGH]
https://grafana.com/docs/tempo/latest/release-notes/v3-0/
Recommendation: VictoriaTraces if you accept pre-GA and want one vendor stack. Tempo monolithic if you need frozen formats and TraceQL metrics. [LOW, judgement]

## 4. Collection layer

### 4.1 Benchmarks
- Log collectors, VictoriaMetrics benchmark dated 2026-03-20, GCP n2-highcpu-32 single-node kind, about 10k logs/s at 216 bytes, 1 CPU and 1 GiB limits, default Helm values. vlagent v1.48.0, Vector v0.53.0, Alloy v1.13.2, Fluent Bit v4.2.3, OTel Collector v0.146.1. [MED, vendor]
  https://victoriametrics.com/blog/log-collectors-benchmark-2026/
| Collector | CPU cores | Mean memory MiB |
|---|---|---|
| vlagent | 0.062 | 27.9 |
| Fluent Bit | 0.260 | 78.1 |
| Vector | 0.412 | 153.5 |
| OTel Collector | 0.491 | 106.8 |
| Grafana Alloy | 0.578 | 66.4 |
Filebeat and Fluentd dropped logs and were excluded.
- Metrics agents, VictoriaMetrics post [STALE - 2025-02-05]: scraping 1,200 node exporters on 2 vCPU 2 GB instances. vmagent v1.108.0 used 3.2x less CPU and 2.7x less memory than OTel Collector v0.115.0, 1.6x less CPU and 3.0x less memory than Prometheus 3.x agent mode. Alloy not tested. Figures only in charts. [MED]
  https://victoriametrics.com/blog/opentelemetry-prometheus-and-more/
- Alloy official sizing: about 0.4 CPU and 11 GiB RAM per 1M active series, and about 1 CPU and 120 MiB per 1 MiB/s of logs. Applies to its Prometheus pipeline, which keeps a WAL. This explains why a DaemonSet that both scrapes metrics and tails logs reaches 900 MB per node. [MED]
  https://grafana.com/docs/alloy/latest/set-up/estimate-resource-usage/
  A community report of 742 MiB growing over days exists but has no confirmed fix. [LOW]

### 4.2 Best combination for this platform
| Signal | Collector | Reason |
|---|---|---|
| Metrics | vmagent as one VMAgent Deployment | Lowest footprint in vendor tests, native relabeling, stream aggregation, series limits, kubernetes_sd namespace metadata. DaemonSet mode only supports VMPodScrape so avoid it. |
| Logs | vlagent DaemonSet (victoria-logs-collector chart 0.3.8) | 28 MiB and 0.06 CPU at 10k logs/s in vendor test, native Kubernetes metadata, checkpoint file, disk buffer. |
| Traces from SDKs | OTLP direct to VictoriaTraces | OTel Operator Instrumentation CRD injects the SDK and points it at VT. |
| Traces and RED without code change | OBI DaemonSet | See section 5. |
| Gateway | OTel Collector only if you need tail sampling, spanmetrics connector, or attribute processing | Otherwise skip, it costs roughly 100 MiB or more. |
- Fluent Bit is the fallback for node journald or per-namespace tenant routing, at about 78 MiB in the same test. Vector at 154 MiB is the heaviest of the viable options. [MED]
- OTel Operator v0.163.0 (2026-11-11). Auto-instrumentation supports Apache HTTPD, DotNet, Go, Java, Nginx, NodeJS, PHP, Python, Ruby. The Instrumentation CRD is still apiVersion opentelemetry.io/v1alpha1. Injection is by pod or namespace annotation. [HIGH]
  https://github.com/open-telemetry/opentelemetry-operator/blob/main/docs/auto-instrumentation/README.md
- vmagent accepts OTLP metrics at /opentelemetry/v1/metrics. VictoriaLogs and VictoriaTraces accept OTLP natively, so no vendor-specific exporter is required. [HIGH]
- Stream aggregation, deduplication and relabeling are open source in vmagent and VMSingle. [HIGH]
  https://docs.victoriametrics.com/victoriametrics/stream-aggregation/

## 5. eBPF zero-code instrumentation

| Tool | Status 2026-11-13 | Outputs | Fit |
|---|---|---|---|
| OpenTelemetry eBPF Instrumentation (OBI) | v1.0.0 stable 2026-10-30. Repo 639 stars, Apache-2.0 [HIGH] | OTLP traces, RED metrics, runtime metrics, network flow metrics, span metrics, service graph metrics, built-in Prometheus endpoint | Best fit |
| Grafana Beyla | v3.41.0 2026-11-04, v3.42.0 pre-release, 2153 stars [HIGH] | Same family, Grafana distribution | Redundant with OBI |
| Coroot | v1.29.2 2026-11-04, 8018 stars [HIGH] | Own UI, service map, SLOs, profiling | Needs ClickHouse plus a Prometheus-compatible store. Supports VictoriaMetrics as a drop-in. Too many moving parts for 2 nodes [HIGH on requirements, LOW on cost] |
| Pixie | CNCF sandbox, Vizier pre-releases only in 2026 [HIGH] | Own UI | Docs minimum 1 GiB per node, no more than 25 percent of node RAM. Rule out [HIGH] |
| Odigos | Not verified. GitHub API returned 404 for odigos-io/odigos. A third-party page lists v1.31.1 from July 2026 [LOW] | OTel format | Not assessed |
| Cilium Hubble L7 | Cilium v1.20.4 stable [HIGH] | http_requests_total, http_request_duration_seconds, dns, flow, tcp, drop | L7 metrics only for pods with L7 protocol visibility, disabled by default. No traces. Complement, not substitute [HIGH] |

OBI details
- Requirements: Linux 5.8 or newer, BTF, amd64 or arm64, root or specific capabilities. Validated kernels include 5.15, 6.1, 6.6, 6.12, 6.18. [HIGH]
  https://github.com/open-telemetry/opentelemetry-ebpf-instrumentation/blob/v1.0.0/SUPPORT_MATRIX.md
- Languages: Java, .NET, Go, Python, Ruby, Node.js, C, C++, Rust. Protocols include HTTP/S, gRPC, Kafka, PostgreSQL, MySQL, Redis, MongoDB. [HIGH]
  https://opentelemetry.io/docs/zero-code/obi/
- RED metric names in Prometheus form: http_server_request_duration_seconds, http_client_request_duration_seconds. Span metrics and service graph metrics are enabled through the features option. [HIGH, exact feature keys not confirmed]
  https://opentelemetry.io/docs/zero-code/obi/metrics/
- Known v1.0.0 limitations: application metrics can go missing after pod restart until OBI restarts, child spans can start before parents, port-based discovery can pick a helper process, service graph needs the instance label preserved across multiple OBI instances. [HIGH]
  https://github.com/open-telemetry/opentelemetry-ebpf-instrumentation/releases/tag/v1.0.0
- Resource cost: no official numbers. Blog figures range from 40 to 80 MB per node to a 512 MiB limit and are unsourced. Measure on your nodes. [LOW]
- Interaction with Cilium eBPF programs: not researched. Test on one node first. [LOW]
- Output to the VM stack: OTLP traces to VictoriaTraces, Prometheus endpoint scraped by a VMPodScrape. No collector needed. [MED, design inference]

## 6. Continuous profiling

| Option | Status | Cost and fit |
|---|---|---|
| Grafana Pyroscope | v2.4.3 (2026-11-03), 11.7k stars, AGPL-3.0 [HIGH] | Single-binary mode with local filesystem exists but docs say not recommended for production and has no HA. Distributed mode requires object storage. No official small-cluster sizing found. Alloy pyroscope.ebpf figures not found. |
| Parca | v0.30.0 (2026-11-06), 5017 stars, Apache-2.0 [HIGH] | Polar Signals was acquired by Dash0 on 2026-08-17. Both companies say Parca stays open source and maintained. Direction risk. Sizing not found. |
| OTel profiling signal | Public alpha since March 2026, not stable in OTLP, no production backends [HIGH] | Do not build on it yet |
| Coroot | Built-in profiling | Needs the full Coroot stack |
| VictoriaMetrics | No profile store exists [HIGH by absence] | Not applicable |
Sources: https://grafana.com/docs/pyroscope/latest/reference-pyroscope-v2-architecture/deployment-modes/ , https://opentelemetry.io/blog/2026/profiles-alpha , https://blog.polarsignals.com/blog/posts/2026/08/17/polar-signals-is-joining-dash0
Recommendation: defer profiling. If wanted later, a single Pyroscope or Parca pod with its own PVC, accept no HA. [LOW]

## 7. VictoriaMetrics-native tooling

- SLOs: no native SLO engine found. Generate Prometheus-format rules with Sloth (v0.16.0, 2026-04-04) or Pyrra (v0.10.2, 2026-09-18) and wrap them in VMRule. A Pyrra pull request for VictoriaMetrics support from June 2025 exists, merge status not confirmed. [LOW]
  https://github.com/pyrra-dev/pyrra/pull/1472
- vmalert-tool unittest: open source, Prometheus-compatible test format except the field name metricsql_expr_test replacing promql_expr_test. Starts an isolated VictoriaMetrics instance. Fits CI for VMRule. [HIGH]
  https://docs.victoriametrics.com/victoriametrics/vmalert-tool/
- Anomaly detection: vmanomaly is Enterprise only and needs a license. Chart victoria-metrics-anomaly 1.12.21. Open source alternatives: MetricsQL zscore, outliers_mad, outliersk, mad plus vmalert. [HIGH]
  https://docs.victoriametrics.com/anomaly-detection/
- Grafana plugins: victoriametrics-metrics-datasource v1.0.0 (2026-11-02, AGPL-3.0 licence per GitHub API) and victoriametrics-logs-datasource v0.34.0 (Apache-2.0). The metrics plugin is optional since the Prometheus datasource type works, but the plugin gives MetricsQL support and the Drilldown. Check the AGPL licence against your distribution policy. [HIGH]
- Perses has VictoriaLogs and VictoriaMetrics datasource integrations if you ever leave Grafana. [MED]
  https://docs.victoriametrics.com/victorialogs/integrations/perses/
- MCP servers, all Apache-2.0: mcp-victoriametrics v1.21.0 (2026-11-12, 245 stars), mcp-victorialogs v1.9.0 (2026-04-10, 111 stars), mcp-victoriatraces v1.5.0 (2025-12-15, 19 stars). The metrics server exposes read-only APIs, supports stdio, SSE and streamable HTTP, a Helm chart, VM_DEFAULT_TENANT_ID, and header passthrough via MCP_PASSTHROUGH_HEADERS so you can put it behind vmauth per project. [HIGH]
  https://github.com/VictoriaMetrics/mcp-victoriametrics
- OpenCost works against VictoriaMetrics: blog guidance sets prometheus.external to the vmselect or vmsingle URL, based on a feature in OpenCost v1.110.0. Current OpenCost v1.121.4 (2026-11-10). No dedicated OpenCost reference page found. [MED]
  https://opencost.io/blog/victoria-metrics

## 8. RAM and CPU budget

Assumptions: 2 nodes, about 70 pods, 200k active series at peak, 10 GB/day logs at the middle of your range, a few dozen spans per second. Only rows marked with a source have a published anchor. Everything else is my estimate and must be measured with vmagent and process metrics before sizing requests. All [LOW] unless stated.
Requests are what I would start with, limits are where OOM becomes acceptable.

| Component | Replicas | Request RAM | Limit RAM | CPU request | Basis |
|---|---|---|---|---|---|
| VM operator | 1 | 64 Mi | 256 Mi | 20m | Estimate |
| VMSingle | 1 | 768 Mi | 1.5 Gi | 300m | Docs: under 1 KB per active series, keep 50 percent RAM free. 200k series is about 200 MB resident plus caches and queries. [MED anchor] |
| VMAgent | 1 | 256 Mi | 768 Mi | 100m | Estimate. Vendor says it uses less than Prometheus agent, no absolute number. Alloy rule of 11 GiB per 1M series does not apply to vmagent. |
| vmalert | 1 or 2 | 64 Mi | 256 Mi | 20m | Estimate |
| Alertmanager or VMAlertmanager | 1 | 48 Mi | 128 Mi | 10m | Estimate |
| vmauth | 1 | 48 Mi | 128 Mi | 20m | Estimate |
| kube-state-metrics | 1 | 96 Mi | 256 Mi | 20m | Estimate |
| node-exporter | 2 | 24 Mi | 64 Mi | 10m | Estimate |
| OpenCost | 1 | 96 Mi | 256 Mi | 20m | Estimate |
| VictoriaLogs single | 1 | 512 Mi | 1.5 Gi | 250m | Vendor test shows 1.3 GB at 65 MB/s, you ingest about 0.1 MB/s on average. Heavy queries are the risk. [MED anchor] |
| Extra VLSingle for long retention | 1 or 2 | 256 Mi | 768 Mi | 100m | Needed if you keep per-class retention. Estimate |
| vlagent | 2 | 48 Mi | 128 Mi | 25m | Vendor test: 28 MiB and 0.062 CPU at 10k logs/s. Yours is lower per node. [MED anchor] |
| VictoriaTraces single | 1 | 256 Mi | 768 Mi | 100m | Vendor PoC: 1.15 GiB at 10k spans/s, you send a small fraction. [STALE anchor] |
| OBI DaemonSet | 2 | 128 Mi | 384 Mi | 50m | Unverified blog range 40 to 512 MiB |
| Grafana 12.3 | 1 | 256 Mi | 512 Mi | 100m | Existing |
| OTel Collector gateway, optional | 1 | 128 Mi | 256 Mi | 50m | Vendor test 107 MiB at 10k logs/s |
| Pyroscope or Parca, optional | 1 | 512 Mi | 1 Gi | 100m | No official sizing found |

Totals, sum of requests
- Core without optional rows: about 3.2 GiB requested, about 8 GiB of limits. [LOW]
- With optional rows: about 4.0 GiB requested. [LOW]
- Today: Alloy alone is about 1.8 GB across two nodes (900 MB each, your figure), plus VMCluster with three components, Loki SingleBinary, and standalone Alertmanager. [HIGH for Alloy, from your report]
- Expected saving: Alloy replaced by vmagent plus vlagent plus OBI is roughly 1.5 GB lower on the collection side, and VMCluster to VMSingle removes two components. Net saving of 2 to 4 GB is plausible but unmeasured. [LOW]

Disk, at 200k series and 10 GB/day logs
- Metrics: VM compresses to roughly a byte or two per sample. At 200k series and a 30s scrape interval that is about 576M samples per day, so about 0.6 to 1.2 GB/day, 20 to 40 GB for 30 days before dedup. [LOW, estimate]
- Logs: about 0.5 to 2 GB/day on disk at 10x or better compression, 15 to 60 GB at 30 days. [MED, docs compression claim]
- Traces: small at your span rate. [LOW]
- All on Ceph RBD PVCs. VictoriaLogs and VictoriaTraces cannot use RGW until the object storage PR lands. [HIGH]
- Docs advise keeping 20 percent disk free and 50 percent spare RAM and CPU for VictoriaMetrics. On 28 GB total you have room, but the limits column above should not all be reached together.

## 9. Recommended stack

1. victoria-metrics-k8s-stack 0.98.x with operator, VMSingle, VMAgent, VMAlert, VMAlertmanager or your existing Alertmanager, VMAuth. Bundled Grafana off, bundled KSM and node-exporter on.
2. VLSingle plus VLAgent (or the victoria-logs-collector chart). Add a second VLSingle only for the retention classes you cannot drop.
3. VTSingle plus OTLP from SDKs and OBI. Accept pre-GA or fall back to Tempo monolithic.
4. vmauth per project: metrics_extra_filters on the project label, logs extra_stream_filters on pod namespace, trace access by tenant header.
5. vmalert-tool unit tests in CI for VMRule. Sloth or Pyrra output wrapped in VMRule for SLOs.
6. Hubble stays for network flows and DNS. Enable L7 only on selected pods.
7. Skip: VMCluster, vmanomaly, Coroot, Pixie, profiling, until a concrete need shows up.

## 10. Open questions and gaps

- No official RAM figures for vmagent at 200k series, VictoriaTraces, OBI or Pyroscope. Measure.
- Per-namespace tenant routing in vlagent not documented. Confirm by test or accept the single-tenant plus vmauth filter design.
- Whether OBI coexists cleanly with Cilium eBPF on your kernel, not researched.
- Odigos status not verified.
- Whether the k8s-stack default dashboards can be provisioned into several Grafana orgs without custom work, not verified.
- VictoriaLogs S3 offload release date, only a stated ETA of November 2026 with open bug reports.
- Alloy and Loki sizing on your cluster come from your own report, not measured here.

## Sources accessed 2026-11-13

GitHub releases and repo metadata
- https://github.com/VictoriaMetrics/VictoriaMetrics/releases
- https://github.com/VictoriaMetrics/VictoriaLogs/releases
- https://github.com/VictoriaMetrics/VictoriaTraces/releases
- https://github.com/VictoriaMetrics/operator/releases
- https://github.com/VictoriaMetrics/helm-charts/releases
- https://github.com/VictoriaMetrics/VictoriaLogs/issues/48
- https://github.com/VictoriaMetrics/VictoriaLogs/issues/15
- https://github.com/open-telemetry/opentelemetry-ebpf-instrumentation/releases/tag/v1.0.0
- https://github.com/open-telemetry/opentelemetry-ebpf-instrumentation/blob/v1.0.0/SUPPORT_MATRIX.md
- https://github.com/open-telemetry/opentelemetry-operator/blob/main/docs/auto-instrumentation/README.md
- https://github.com/VictoriaMetrics/mcp-victoriametrics
- https://github.com/VictoriaMetrics/victorialogs-datasource
- https://github.com/pyrra-dev/pyrra/pull/1472
- GitHub API for stars, licenses and latest tags of VictoriaMetrics, VictoriaLogs, VictoriaTraces, operator, helm-charts, victorialogs-datasource, victoriametrics-datasource, mcp-victoriametrics, mcp-victorialogs, mcp-victoriatraces, OBI, grafana/beyla, coroot, pixie, parca, pyroscope, otel operator, otel collector, alloy, vector, fluent-bit, cilium, loki, tempo, jaeger, opencost, sloth, pyrra, perses

Raw docs read from raw.githubusercontent.com master branches
- VictoriaMetrics docs: README.md, Cluster-VictoriaMetrics.md, FAQ.md, enterprise.md, vmagent.md, vmauth.md, vmalert.md, vmalert-tool.md, vmbackup.md, sd_configs.md, MetricsQL.md, LTS-releases.md, stream-aggregation/README.md, and the enterprise flag files
- VictoriaLogs docs: README.md, cluster.md, FAQ.md, Roadmap.md, LTS-releases.md, vlagent.md, vmalert.md, security-and-lb.md, querying/README.md, logql-to-logsql.md, victoria_logs_enterprise_flags.md
- VictoriaTraces docs: README.md, roadmap.md, querying/README.md, keyConcepts.md, data-ingestion/README.md
- operator docs: api.md, enterprise.md, resources/vmanomaly.md, resources/vmalertmanagerconfig.md, resources/vmrule.md, backups.md
- helm-charts: victoria-metrics-k8s-stack Chart.yaml and values.yaml

Docs sites
- https://docs.victoriametrics.com/victorialogs/
- https://docs.victoriametrics.com/victorialogs/cluster/
- https://docs.victoriametrics.com/victorialogs/data-ingestion/
- https://docs.victoriametrics.com/victorialogs/changelog/
- https://docs.victoriametrics.com/victorialogs/roadmap/
- https://docs.victoriametrics.com/victorialogs/logsql/
- https://docs.victoriametrics.com/victoriametrics/enterprise/
- https://docs.victoriametrics.com/operator/resources/
- https://docs.victoriametrics.com/helm/victoria-logs-collector/
- https://docs.victoriametrics.com/anomaly-detection/
- https://opentelemetry.io/docs/zero-code/obi/
- https://opentelemetry.io/docs/zero-code/obi/metrics/
- https://docs.cilium.io/en/stable/observability/metrics/
- https://docs.coroot.com/ , https://docs.coroot.com/configuration/prometheus , https://docs.coroot.com/installation/requirements
- https://grafana.com/docs/alloy/latest/set-up/estimate-resource-usage/
- https://grafana.com/docs/pyroscope/latest/reference-pyroscope-v2-architecture/deployment-modes/
- https://www.dash0.com/knowledge/what-is-continuous-profiling

Blogs and benchmarks
- https://victoriametrics.com/blog/log-collectors-benchmark-2026/
- https://victoriametrics.com/blog/opentelemetry-prometheus-and-more/
- https://victoriametrics.com/blog/dev-note-distributed-tracing-with-victorialogs/
- https://dev.to/sachincool/victorialogs-vs-loki-real-world-benchmarking-results-26bg

Search results consulted without full fetch, lower weight
- Tempo 3.0 release notes https://grafana.com/docs/tempo/latest/release-notes/v3-0/
- OpenTelemetry Profiles alpha https://opentelemetry.io/blog/2026/profiles-alpha
- Polar Signals joins Dash0 https://blog.polarsignals.com/blog/posts/2026/08/17/polar-signals-is-joining-dash0
- OpenCost with VictoriaMetrics https://opencost.io/blog/victoria-metrics
- Pixie requirements https://docs.px.dev/installing-pixie/requirements/
- Odigos third-party listings, Jaeger storage comparisons, VM sizing rule of thumb from mirrored docs, Alloy community memory thread
