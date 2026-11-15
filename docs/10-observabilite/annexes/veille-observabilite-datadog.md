# Observability for a small self-hosted IDP (CNP): state of the art, November 2026

Research date: 2026-11-13. All "accessed" dates in Sources are 2026-11-13.

Confidence tags: [HIGH] primary source read (docs, GitHub page, vendor release page). [MED] secondary or search-result summary of a primary. [LOW] inferred, from general knowledge, or unverified.
Many pages were read through a summarizing fetch tool, so numbers such as star counts are as displayed at fetch time and may be rounded.
Resource figures (RAM) are qualitative estimates unless a source is cited. I did not find or run benchmarks. Treat all RAM tiers as [LOW].

---

## 0. Executive summary

1. Stack choice is validated. VictoriaMetrics' own docs recommend Grafana as the primary UI for VictoriaStack because cross-signal correlation is most mature there. Perses is the dashboards-as-code option with less mature correlation [HIGH].
2. VictoriaTraces is the risky piece. Releases are 0.x (v0.13.0 on 2026-11-05), docs say the Grafana/Jaeger query path is "not yet stable", and VictoriaMetrics storage has no exemplars. Make traces opt-in per app, not a hard dependency [HIGH].
3. The highest-value, lowest-cost wins for a tiny cluster are: per-app auto-generated golden-signal dashboards, auto-generated SLOs with multi-window burn-rate alerts, deploy annotations, Argo Rollouts canary analysis on the same SLI queries, Discord alerts with runbook links and project routing, and a status page. All of these reuse the same SLI definition.
4. Grafana OnCall OSS is dead (archived). Opsgenie shuts down 2027-04-05. For a student/demo platform, do not build on-call. Discord is the pager, ntfy is optional.
5. Skip or defer: session replay, Kepler on VMs, vmanomaly (Enterprise), Sentry self-hosted (heavy), Keep, DevLake, full LLM observability, Datadog-style incident management.
6. Cheap AI: read-only, triggered only on paging alerts, with deterministic pre-fetched context, and an MCP server for ad hoc investigation. Do not give AI write access (surveys show only 50% of decision-makers accept that).

---

## 1. Datadog feature map to OSS (VictoriaMetrics + Grafana)

Datadog popularity signal. Datadog Q3 2026 earnings (reported early Nov 2026): 61% of customers use 4+ products, 41% use 6+, 16% use 10+; Infrastructure Monitoring ARR above $2B [MED, via earnings summaries, not primary filing]. I found no public per-product usage ranking for 2025-2026, so the order below is my judgment: infra, logs, APM, then dashboards/monitors, then the rest [LOW].

Datadog Bits AI SRE: launched Dec 2025, company claims 100,000+ investigations and 2,000+ customers in the last month; DASH 2026 added autonomous remediation and Bits Agent Builder, with several pieces still in preview per TechTarget [MED, vendor-reported].

RAM tiers: Light < ~200 MB, Medium ~200 MB to 1 GB, Heavy > 1 GB per component, for a cluster of a few nodes. All [LOW] estimates.

| Datadog feature | Best OSS fit for VM + Grafana | Effort | RAM | Worth it on a tiny cluster? |
|---|---|---|---|---|
| Infrastructure monitoring | vmagent/VM operator, node_exporter, kube-state-metrics, Grafana k8s dashboards | Low | Medium | Yes, baseline |
| APM + service map | OTel Operator auto-instrumentation to VictoriaTraces via Jaeger datasource; service map from Grafana node graph, Hubble UI or OBI (see 6) | Medium | Medium | Yes, opt-in per app. VictoriaTraces is pre-1.0 |
| Log management | VictoriaLogs (GA announced 2024-11-13; v1.52.1 is first LTS, backports are Enterprise) | Low | Light to medium | Yes |
| RUM | Grafana Faro Web SDK + Alloy faro.receiver (see 5) | Medium | Light | Yes for React template, Core Web Vitals only |
| Session replay | OpenReplay (self-host) or rrweb | High | Heavy | No (GDPR consent, heavy) |
| Synthetics | Gatus (config as code, status page) and/or blackbox exporter via VM probes; k6 for load | Low | Light | Yes |
| Continuous profiler | Grafana Pyroscope + Alloy `pyroscope.ebpf` (GA) | Medium | Medium | Yes as demo, opt-in. Needs root/host PID or specific capabilities |
| Database monitoring | CNPG built-in metrics + custom `pg_stat_statements` queries + dashboard 20417; pgwatch v5.2.0 (May 2026) as alternative | Low | Light | Yes, high value for the FastAPI+PG template |
| SLOs | Sloth (CRD/CLI/Go lib) or Pyrra (UI) | Low to medium | Light | Yes, core feature (see 3) |
| Incident management | Discord thread + a record in the portal. Grafana IRM is Cloud (free: 3 active users) | Low if minimal | None | Build minimal only |
| On-call | None worth running. Grafana OnCall OSS archived | n/a | n/a | No |
| Watchdog anomalies | vmanomaly is Enterprise-only. OSS: MetricsQL outlier/z-score style rules in vmalert | Low | Light | Only cheap rules. Verify function names in MetricsQL docs [LOW] |
| CI visibility | GitLab CI Pipelines Exporter (Prometheus metrics) or GitLab CI/CD telemetry (Beta, `GITLAB_OBSERVABILITY_EXPORT`); OTel CI/CD semantic conventions | Low to medium | Light | Yes if CI is GitLab. Not verified which CI CNP uses |
| DORA metrics | Compute directly from platform events (git tags, Argo CD syncs, Rollouts aborts, alerts). Apache DevLake is the OSS product but brings its own DB and Grafana | Low (custom) vs medium (DevLake) | Light vs heavy | Yes, custom. DevLake no |
| LLM observability | OpenLIT (Apache-2.0, OTel GenAI conventions) or Langfuse (MIT core, ClickHouse-owned since Jan 2026, heavy stack) | Medium | Heavy for Langfuse | Only if tenant apps call LLMs. OTel GenAI conventions still experimental |
| Cost management | OpenCost (CNCF sandbox) or DIY recording rules from kube-state-metrics requests | Low | Light | Yes for FinOps role, DIY first |
| Security signals | Tetragon or Falco events to VictoriaLogs, Hubble flows, K8s audit logs | Medium | Medium | Partial. Not researched in depth this session [LOW] |
| Error tracking | GlitchTip (about 256-512 MB per third-party blogs) vs self-hosted Sentry (8-16 GB, 40+ containers per third-party blogs; FSL license) | Low (GlitchTip) | Light vs heavy | GlitchTip only if wanted. Sentry no |
| Bits AI / AI SRE | HolmesGPT (CNCF sandbox), MCP servers, custom small-model summarizer (see 7) | Low to medium | Light | Yes, scoped |
| Service catalog + scorecards | Build in CNP portal ("observability readiness score"); Backstage needs plugins; Port/OpsLevel/Cortex are SaaS-first | Low (custom) | None | Yes, cheap and demo-friendly |
| Network performance | Cilium Hubble metrics + UI | Low | Light (already in Cilium) | Yes. L7 metrics need L7 policies |

Key details with sources:
- VictoriaTraces: built on VictoriaLogs, OTLP ingest, Jaeger Query JSON API, default retention 7 days, no built-in per-tenant authorization (use vmauth), vendor claim of up to 3.7x less RAM and 2.6x less CPU than Tempo (no methodology seen, treat as marketing) [HIGH for docs content, LOW for claim]. The docs do not state stability or license. The Grafana querying page says the feature is not yet stable [MED]. Releases: v0.13.0 (2026-11-05), v0.12.1 (2026-10-23), v0.12.0 (2026-09-29), v0.11.1, v0.11.0 (2026-08-14) [HIGH]. No GA announcement found.
- Exemplars: VictoriaMetrics storage does not support them. Metric-to-trace jumps go through Grafana correlations. Trace-to-metrics is supported in the VM datasource from Grafana 12.4.0 [HIGH, VM docs].
- vmanomaly: requires VictoriaMetrics Enterprise license key (trial available) [HIGH].
- OBI (OpenTelemetry eBPF Instrumentation, the donated Grafana Beyla) reached v1.0.0 on 2026-10-30. Config v2 is stable, runtime metrics for Go/Java/.NET/Node/Python are still experimental. The post does not list kernel requirements or overhead [HIGH].
- Pyroscope eBPF: Alloy must run as root in the host PID namespace, or on K8s with specific capabilities and kernel fs mounts. Kernel >= 4.9. Supports native code, HotSpot JVM, Python, Node.js and others [HIGH, Alloy docs]. Grafana Agent reached EOL 2025-11-01, use Alloy [HIGH].
- Coroot (Apache-2.0, eBPF service map, SLO inspections) is a credible all-in-one alternative, but it brings its own ClickHouse-based storage and does not fit the "full VictoriaMetrics" decision [MED]. Community vs Enterprise feature split not verified.
- CNPG ships postgres_exporter metrics but per-query metrics need custom queries (third-party blog) [MED].

Verdict on the "tiny cluster" question: run the always-on trio (VM, VictoriaLogs, Grafana) plus vmalert/Alertmanager and Gatus. Make traces, profiling and RUM opt-in per app so RAM scales with what tenants actually switch on.

---

## 2. What roles actually ask observability for

### 2.1 Survey evidence

| Source (date) | Finding | Confidence |
|---|---|---|
| Grafana Labs 4th Observability Survey, published 2026-03-18, n=1,363, 76 countries, fielded Oct 2025 to Jan 2026 | Top concerns: complexity/overhead 38%, signal-to-noise 34%, cost 31%. Alert fatigue is the biggest obstacle to faster incident response (30%, nearly double next). Cost is top tool-selection criterion 3 years running (65%), then ease of use (49%). 77% say open source or open standards matter, 61% essential or very important. Prometheus investment 77%, OTel 76%, 65% invest in both. OTel use: metrics 57%, traces 50%, logs 48%. Half use observability tools for business metrics. Self-managed teams cite complexity most. 95% want AI to show its reasoning. | HIGH (Grafana press release read) |
| Grafana/ETR "State of Observability 2026", Oct 2026, decision-makers n=250 (US 150, JP/LatAm/W.Europe 50 each) | 85% comfortable with AI read-only investigation, 50% with AI deploying changes. Consolidation very or extremely important to 79%. | MED (press coverage, report not opened) |
| CNCF post 2026-05-06 citing Middleware survey, Feb 2026, n=407 | 46.7% run 2-3 observability tools, only 7.4% a single unified experience. Top challenge: dashboard and alert configuration (54%), then integration complexity (46.4%). 59.5% want AI anomaly detection, 48.3% want human oversight before autonomous remediation. Vendor-affiliated survey (Middleware). | HIGH on content, LOW on neutrality |
| Grafana 3rd survey, 2025-03 | 71% use Prometheus and OTel in some capacity. [STALE, 20 months old] | MED |
| CNCF Annual Survey 2025 (n=628, Sept 2025, published Jan 2026) | 82% of container users run Kubernetes in production. Observability-specific figures not found in what I read. | MED |
| DORA 2025 "State of AI-assisted Software Development" (about 5,000 respondents) | Five metrics grouped as throughput (lead time, deploy frequency, failed deployment recovery time) and instability (change fail rate, deployment rework rate). Tiers replaced by archetypes. AI adoption correlates with higher throughput and higher instability. Grouping differs slightly across secondary sources, check the report. | MED |
| Gartner Magic Quadrant for Observability Platforms, July 2025 | Cost optimization, AI observability, and cost management listed as common feature areas. No 2026 edition found in my searches. [STALE, 16 months] | MED |
| Honeycomb / Charity Majors "observability 2.0" | One source of truth of wide structured events, derive metrics/logs/traces from it, control cost by sampling events. Critiqued as partly marketing ("North Star not silver bullet", Grafana talk summary). No 2026-dated primary source found. | MED |

Interpretation for CNP: the most credible pain points are complexity, noise and cost. A golden path that auto-configures dashboards, SLOs and alerts addresses the top two survey concerns directly (dashboard/alert configuration is the top setup challenge at 54%).

Wide events note: VictoriaLogs can hold wide structured JSON logs per request, which gives a cheap "observability 2.0 lite" (one structured access-log event per request with trace_id, user/tenant, status, duration) without migrating off the metrics/logs/traces trio. [LOW, my inference]

### 2.2 Role to need mapping (mostly inference, tagged [LOW] unless cited)

| Role | What they ask for | What CNP should provide |
|---|---|---|
| Developer (project member) | Is my deploy healthy, why is it slow or failing, logs and traces for one request, error rate after release | Per-app dashboard (RED + saturation), deploy markers, log explorer scoped to the project, trace-to-logs, canary verdict |
| Project owner/admin | SLO status, error budget, who gets alerted, cost of the project, uptime page | SLO burn-rate view, alert routing settings (Discord webhook per project), status page, cost showback |
| Guest | Read-only status | Status page and read-only dashboard |
| Platform admin / SRE | Golden signals and USE for infra, platform component health, tenant noisy-neighbour, alert noise | Platform dashboards, per-tenant resource usage, Alertmanager health, alert volume dashboard (`alertmanager_alerts_limited_total` exists for flood control [HIGH]) |
| Security auditor | Who accessed what, policy violations, runtime events, audit trail | Audit logs (K8s API, Keycloak, Vault) in VictoriaLogs, Hubble drops/policy verdicts, Tetragon/Falco events. Read-only, cross-project, no write rights |
| FinOps | Cost per project and per request, idle/over-provisioned requests, trend | OpenCost or recording rules, requests vs usage ratio, cost per 1k requests (namespace cost / Envoy request count) |
| Product/business owner | Business KPIs, user experience | Custom app metrics guidance, Core Web Vitals (LCP, INP, CLS) per app. Business KPIs need app-level instrumentation. Half of Grafana-survey respondents track business metrics [HIGH] |
| Management (demo audience) | Is the platform reliable and improving | DORA four/five metrics per project, SLO attainment rollup, readiness scorecard |

Core Web Vitals: Faro Web SDK v2 upgraded to web-vitals v5 and removed FID (INP is the current responsiveness metric) [HIGH for the removal, MED for INP replacement being general knowledge].

---

## 3. SLO tooling and auto-generation

### 3.1 Tool status

| Tool | Status (checked) | Fit |
|---|---|---|
| Sloth (slok/sloth) | v0.16.0 on 2026-04-04, adds a `server` command with web UI and K8s transformer plugins. v0.15.0 (2025-10-31) exposes generation as a Go library `pkg/lib`. v0.14.0 adds contrib plugin `validate_victoria_metrics/v1` and allows `github.com/VictoriaMetrics/metricsql` in plugins. v0.13.0 moved to a plugin system with breaking changes. About 2.6k stars. Apache-2.0. Supports OpenSLO v1alpha. Generates multi-window multi-burn-rate page and ticket alerts. [HIGH] | Best fit. Go library fits a Go platform backend; VM plugin exists |
| Pyrra | Now under prometheus-community/pyrra. v0.10.2 on 2026-09-18, v0.10.1 2026-06-25, v0.10.0 2026-04-30 (performance mode, alert severity overrides, Prometheus v3 histogram handling, `pyrra_url` annotation). [HIGH] VictoriaMetrics compatibility not confirmed in release notes [LOW]. | Nicer built-in UI, but VM support unverified. Test before adopting |
| OpenSLO | Vendor-neutral spec started by Nobl9 in 2021. Sloth supports v1alpha only, and an old Sloth doc listed restrictions (30-day window cap, no alerting from OpenSLO). Those limits are from v0.6.0 docs and may be stale [LOW]. | Use as an export format at most, not a source of truth |
| Grafana SLO app | Presented as a Grafana Cloud product. I could not confirm whether it installs on Grafana OSS. | Do not count on it [LOW] |

Conflict note: one aggregator suggested Pyrra YAML is "OpenSLO-compatible" while the repo examples use the native `pyrra.dev` CRD. Not resolved.

### 3.2 Burn-rate alerting parameters (Google SRE Workbook, 99.9% SLO example) [HIGH]

| Severity | Long window | Short window | Burn rate | Budget consumed |
|---|---|---|---|---|
| Page | 1 h | 5 min | 14.4 | 2% |
| Page | 6 h | 30 min | 6 | 5% |
| Ticket | 3 d | 6 h | 1 | 10% |

The workbook says numbers depend on service and baseline traffic.

### 3.3 Auto-generating SLOs per app on CNP (design suggestion, [LOW] unless noted)

- Source of SLIs: Envoy Gateway request metrics per route/backend are the platform-side SLI, so no app instrumentation is required. Exact metric names depend on the Envoy Gateway version and need verification.
- Default objectives per template: availability 99.5% over a short window for demo apps (a 30-day window barely fills on a student platform), latency threshold SLO using a histogram bucket (for example 95% under 500 ms). Let owners edit targets in the portal.
- Generation path: the Go backend calls Sloth as a library (v0.15+) or CLI in CI, emits PrometheusRule/VMRule, commits it to the GitOps repo alongside the app. VM operator converts Prometheus Operator CRDs, which should work with Sloth's output [MED, general knowledge, verify].
- Low-traffic problem: burn-rate alerts misfire on near-zero traffic. Add a minimum request-rate guard to the alert expression, and mark the SLO "insufficient data" in the UI instead of 100%.
- Link to rollouts: use the same SLI query for Argo Rollouts analysis (section 6.4). One definition feeds dashboards, alerts and canaries.
- Scorecard: "has SLO", "has runbook_url on every alert", "has dashboard", "has deploy annotations" feed an observability readiness score in the portal. OpsLevel and Cortex market similar checks (SLO present, alert policy, on-call, runbook, dashboard); Backstage needs plugins for it [MED, vendor sources].

---

## 4. Alerting, on-call, status pages, synthetics

### 4.1 On-call and paging status

| Item | Status | Confidence |
|---|---|---|
| Grafana OnCall OSS | Maintenance mode since 2025-03-11. README says archived 2026-03-24 and Cloud connection features (SMS, phone, push) ended that day. GitHub banner on the repo says archived 2026-06-05, read-only. Dates conflict, outcome is the same. AGPLv3. Recommended path is Grafana Cloud IRM. | HIGH |
| Grafana IRM (Cloud) | Free plan includes IRM for 3 active users, community support. Pro listed from $20 per active IRM user. Base fee details differ across third-party sources. | MED |
| Opsgenie | Atlassian stopped selling 2025-06-04. End of support and shutdown 2027-04-05, APIs shut off and data deleted. From vendor blogs, Atlassian primary page not fetched. | MED |
| Keep (keephq) | Open-source alert management, about 11.9k stars per an aggregator, updated 2026-05-28. YC page says Keep joined Elastic in May 2025 (unverified). AIOps features marked Enterprise-only. License shown as NOASSERTION by one aggregator, check the LICENSE file. | LOW to MED |
| Karma (prymitive/karma) | v0.131 on 2026-05-18, requires Alertmanager >= 0.22, Apache-2.0 | MED |
| ntfy | Self-hosted push. Android works direct. iOS needs `NTFY_UPSTREAM_BASE_URL` pointing at ntfy.sh for APNs relay (content-free poll). Alertmanager needs a template or adapter. | MED |

Recommendation: no on-call product for a demo platform. If a "paging" story is wanted for the demo, Alertmanager to ntfy for mobile push is the cheapest honest option. Do not adopt Keep: it adds a backend, DB and UI for features Alertmanager routing plus a portal alerts page already cover. [LOW, judgment]

### 4.2 Alertmanager practices (from Alertmanager docs, [HIGH])

- Group by deliberate labels (for example `project`, `alertname`) so one notification replaces hundreds.
- Inhibition for root-cause alerts, for example a platform or node outage suppresses tenant pod alerts.
- Silences are created in the Alertmanager UI. No mature OSS "silence from chat" tool was found in this research. Building a small Discord slash command that calls the Alertmanager v2 API is a modest Go task [LOW].
- Do not load-balance Prometheus/vmalert to Alertmanagers. List every instance. Cluster with `--cluster-*` flags.
- `--alerts.per-alertname-limit` caps floods. Monitor `alertmanager_alerts_limited_total` (enable `alert-names-in-metrics` feature flag for the alertname label).
- Discord: native `discord_configs` receiver since Alertmanager v0.25.0 (2022-12-23) [HIGH, PromLabs]. Later option changes not verified. No bridge needed.

### 4.3 Per-project routing and alert-as-code (design, [LOW])

- Every alert rule carries `project`, `app`, `severity`, `runbook_url`. The platform injects `project`/`app` from namespace labels so tenants cannot forge them.
- Route by `project` to a per-project Discord webhook stored in Vault, rendered into the Alertmanager config by the platform. Platform-level alerts go to the admin channel. Security alerts go to a security-auditor channel.
- Severity policy: only `critical` (SLO fast burn) pings; `warning` goes to a quiet channel or digest. This directly targets the alert-fatigue finding (30%).
- Alert rules as code: templates ship default rules (SLO burn, crashloop, 5xx, DB connections) as VMRule objects generated by the platform and delivered by Argo CD. Lint in CI that `runbook_url` exists. Validate with promtool or vmalert tooling [LOW, tools not verified this session].
- Add `keep_firing_for` on flappy rules to cut notification churn (supported in recent Prometheus; verify vmalert support) [LOW].
- A portal "Alerts" tab reading the Alertmanager v2 API filtered by `project` is cleaner for tenants than exposing Karma. Karma is fine for platform admins.

### 4.4 Status pages and synthetics

| Tool | Facts | Fit |
|---|---|---|
| Gatus | 12.8k stars, Apache-2.0, 273 open issues at fetch time. Status page, conditions, Discord alerting, Prometheus metrics (feature list from the fetch tool's notes, not all verified on the page). YAML config is git-reviewable. K8s auto-discovery was not confirmed in core docs, may be a third-party sidecar [LOW]. | Best fit. Generate its config from the app catalog (one endpoint per app/env) |
| Uptime Kuma | UI-first, SQLite. One comparison names v2.5.3 (Aug 2026). Not GitOps friendly. | Skip |
| cState | Not researched this session. | Not assessed |
| blackbox exporter | Standard Prometheus prober, feeds VM via probe CRDs (VMProbe, from operator knowledge) [LOW] | Use for SLO-visible probes, Gatus for the page |
| k6 | Grafana load testing, operator exists | Demo "load test before canary" idea, [LOW, not researched] |
| Checkly OSS | Not found / not researched. | Skip |

Gatus status pages and probe-based SLIs add black-box signals that catch things Envoy metrics cannot (DNS, cert expiry, ingress down).

---

## 5. Frontend RUM

### 5.1 Faro and OpenTelemetry

- Faro Web SDK: 1.2k stars, Apache-2.0, v2 is current (web-vitals v5, FID removed, simplified console and tracing attributes), active repo (1,767 commits). Sends to Grafana Alloy `faro.receiver` (or Grafana Cloud or custom receiver). Alloy forwards logs to Loki and traces to Tempo. Browser-only: no Node/service workers. [HIGH]
- The Alloy docs describe the Faro payload as its own JSON format accepted only by `faro.receiver`. Alloy 1.20.0 (2026-09-25) mentions an OTel/Faro preview [MED, releasebot summary].
- For the VM stack: Faro logs and measurements are log-shaped. Likely path is Alloy faro.receiver to VictoriaLogs (Loki push API compatibility is from my knowledge, verify) and traces via OTLP to VictoriaTraces. Core Web Vitals p75 per app then come from LogsQL stats plus vmalert recording rules to metrics [LOW, unverified pipeline, plan a spike].
- Source maps: Grafana Cloud handles upload; for self-hosted, verify what `faro.receiver` supports before promising readable stack traces [LOW].
- OpenTelemetry JS SDK 3.0 shipped in 2026 with stable core packages. Browser-specific SDK and event instrumentations remain experimental 0.x per a Tessl skill page captured Aug 2026 [LOW to MED]. Prefer Faro today.
- Lean alternative: the `web-vitals` library with `navigator.sendBeacon` to a platform endpoint that records Prometheus histograms by app. Less capable than Faro, but no Alloy dependency and trivial to bake into the React template [LOW, my suggestion].

### 5.2 Session replay and RGPD

- OpenReplay: self-hostable, claims GDPR and SOC 2 for both cloud and self-host, with input masking and sanitization rules [MED, vendor and reseller sources]. It is a multi-component stack, heavy for a tiny cluster [LOW].
- rrweb: MIT library, privacy handling is whatever you build [MED].
- Under ePrivacy plus GDPR, replay usually needs prior consent in the EU, self-hosting does not remove that [MED, Temps blog, a competitor, low neutrality].
- Recommendation: skip session replay. Faro RUM for Core Web Vitals and JS errors with no PII collection and a documented consent/no-cookie stance is enough and safer.

---

## 6. Platform-engineering patterns

### 6.1 Golden path observability by default

- OTel Operator `Instrumentation` CRD plus pod annotations: `inject-java`, `inject-python`, `inject-nodejs`, `inject-dotnet`, `inject-go`, `inject-sdk` [HIGH, opentelemetry.io docs]. Caveats from the docs:
  - Python: only `http/protobuf`, operator >= v0.108.0 sets it automatically. Set `otel-python-platform: "musl"` for Alpine images.
  - Java: many libraries on by default, disable selectively. Quarkus has its own native OTel extension, so the imported Quarkus services should prefer native OTel config over the agent [LOW, general knowledge].
  - Go: eBPF sidecar needs `privileged: true`, root, `OTEL_GO_AUTO_TARGET_EXE`, not multi-container pods.
  - The `Instrumentation` resource must exist before the app deploys or the init container cannot be created.
  - If `spec.exporter.endpoint` is omitted it defaults to localhost:4317.
- Instrumentation CRD is `v1alpha1` [HIGH].
- eBPF: OBI v1.0.0 (2026-10-30) gives language-agnostic RED metrics and traces without code changes, including HTTP/2, gRPC, many DB and messaging protocols. Requires privileged DaemonSet access. Overhead and kernel requirements not stated in the release post [HIGH for what the post says]. Reasonable for the Quarkus/imported apps where agent injection is risky.
- Choose one injection mechanism per language to avoid double instrumentation: OTel Operator for Python/Java/Node, OBI as fallback or for service-map coverage [LOW, judgment].

### 6.2 Dashboards-as-code and per-tenant UI

| Option | Facts | Fit |
|---|---|---|
| Grafana Foundation SDK | Recommended by Grafana docs for Go, TypeScript, Python, Java, PHP. Pre-1.0. A May 2026 engineering note says it is "not yet fully mature", v2 schema not yet merged upstream. A third-party page says v0.0.x tags, pin versions. [MED] | Good for a Go backend generating dashboards. Pin and test against the Grafana version |
| Grafonnet | Marked deprecated, not officially supported [HIGH] | Avoid for new work |
| Perses | CNCF **Sandbox** (accepted 2024-08-29, no later promotion on CNCF project page). 2.5k stars. Go and CUE SDKs, `percli`, K8s-native CRDs via a separate operator. First-class VictoriaLogs plugin from v0.53.0-beta.2. VM docs say its correlation functionality is less mature than Grafana's. A Red Hat Jira ticket calling it "incubating" appears wrong. [HIGH] | Good for GitOps dashboards and a clean embeddable demo, but not a replacement for Grafana correlations today. Project-level RBAC not verified [LOW] |
| Dashboard JSON provisioning via ConfigMaps | Standard Grafana sidecar approach | Simplest. Pair with the SDK to generate JSON |

Multi-tenancy:
- Grafana OSS: Keycloak role mapping works through `role_attribute_path` (admin/editor/viewer). Team sync from IdP groups is Enterprise/Cloud only [HIGH]. Grafana datasource permissions and fine-grained RBAC historically Enterprise, verify for Grafana 13.x [LOW].
- Safest OSS isolation pattern: one Grafana org per project (orgs are OSS) with a datasource that carries a project-scoped token, and the backend enforces the scope. Alternative: single org and per-project query-side enforcement only, with iframe embedding or the portal proxying queries. [LOW, design]
- Data-side enforcement: VictoriaMetrics multitenancy exists only in cluster mode. For single-node, vmauth can inject `extra_label` filters per token. vmgateway (OIDC/JWT tenant claims) is Enterprise. VictoriaTraces and VictoriaLogs use AccountID/ProjectID headers and need vmauth for authorization [HIGH for docs, MED for vmauth being OSS; the search could not confirm its license, my knowledge says it is in the Apache-2.0 repo].
- Embedding dashboards in the portal: use Grafana's anonymous-free options carefully. Prefer a portal-side reverse proxy that sets the org/token, or Grafana `allow_embedding` with per-project service-account tokens [LOW, not researched].

### 6.3 Correlation (metrics to traces to logs)

- VM storage has no exemplars. Use Grafana correlations (metric to logs and metric to traces) and trace-to-metrics (Grafana >= 12.4.0) [HIGH].
- Put `trace_id` in structured logs (OTel auto-instrumentation can inject trace context into logs, Python logs instrumentation is off by default [HIGH]). Then a Grafana derived field or correlation from VictoriaLogs to the Jaeger datasource gives trace-to-logs in one click.
- Derive RED metrics from spans with vmalert recording rules on VictoriaTraces (ITNext article describes recording rules and alerts from spans [MED, secondary]). Service graph support in VictoriaTraces: flags exist but I could not confirm a working service-graph feature [LOW].

### 6.4 Deploy annotations and canary analysis

- Argo CD Notifications has a built-in Grafana service: configure `service.grafana` (`apiUrl`, `apiKey` from a secret), subscribe apps with `notifications.argoproj.io/subscribe.<trigger>.grafana: tag1|tag2`, and write a generic message template [HIGH, Argo CD docs]. Tag with `project` and `app` so dashboards filter annotations per tenant.
- Argo Rollouts analysis [HIGH, official docs]:
  - `AnalysisTemplate` with Prometheus provider, `successCondition: result[0] >= 0.95`, `failureLimit: 3`, `interval: 5m`.
  - Background analysis (`strategy.canary.analysis`, `startingStep`) or inline analysis steps.
  - Failed background run aborts the rollout and sets canary weight to 0, marking it Degraded. Inconclusive pauses.
  - "No new metric integrations are accepted in the core controller", new ones are plugins. Providers documented: Prometheus, Datadog, Web.
- VictoriaMetrics works with the Prometheus provider if you point `address` at VM's Prometheus-compatible API [MED, inference, no VM-specific example found]. Test the path prefix for single-node vs vmselect.
- Gotchas for tiny traffic [LOW, general knowledge]: an empty result or NaN on zero requests makes the analysis fail or error. Guard the query and use `default()`/length checks. For demo canaries, generate synthetic traffic (k6 or Gatus probes) during analysis so the verdict is meaningful.
- Reuse the SLO SLI query as the canary gate so "burn-rate dashboard", "alert" and "auto-rollback" are one story.

---

## 7. Cheap AI that is actually useful

### 7.1 Tooling status

| Tool | Facts | Confidence |
|---|---|---|
| VictoriaMetrics MCP server | `VictoriaMetrics/mcp-victoriametrics`, Apache-2.0, 245 stars, 446 commits at fetch time. Tools: query, query_range, alerts, rules, tenants (cluster), series, labels, tsdb_status, top_queries, documentation search. Modes: stdio, SSE, streamable HTTP. No explicit read-only flag, it exposes read APIs. A public playground at play-mcp.victoriametrics.com is mentioned by a directory. | HIGH |
| VictoriaLogs MCP server | Exists, image `ghcr.io/victoriametrics-community/mcp-victorialogs`, uses `VL_INSTANCE_ENTRYPOINT`, nearly all read-only VL APIs. Directories disagree on the owning org (`VictoriaMetrics` vs `VictoriaMetrics-Community`). Verify the repo before pinning. | MED |
| Grafana MCP | `grafana/mcp-grafana`, Apache-2.0, 3.7k stars. Tools for dashboards, Prometheus, Loki, Pyroscope, Tempo, alerting, incidents, annotations, and more. `--disable-write` makes it read-only. VictoriaMetrics support not stated, likely only via a Prometheus-type datasource [LOW]. | HIGH |
| HolmesGPT | CNCF Sandbox since Oct 2025, Apache-2.0, created by Robusta, any OpenAI-compatible LLM. A third-party aggregator says v0.35.0 on 2026-07-01 (unverified). Loki toolset not confirmed. | MED |
| K8sGPT | CNCF Sandbox since Dec 2023. Rule-based analyzers, LLM only for explanations, works with Ollama. | MED |

### 7.2 Ideas ranked by value per token

1. Alert enrichment summary (best ratio). On a `critical` alert only, de-duplicated per alert group and rate-limited, a Go service pre-fetches deterministic context (recent Argo CD deploys, pod restarts and last termination reason, top error log lines from VictoriaLogs for the last 15 minutes, the SLO burn state, link to dashboard and runbook) and sends about 2-4k tokens to a small model for a 5-line summary posted into the Discord thread. Output is cached per fingerprint. Cost: with a small model this is on the order of fractions of a cent per alert [LOW, check current pricing, no price source read]. Latency does not matter.
2. Postmortem draft from a structured timeline. Inputs: alert firing/resolved times, Argo deploy events, rollout aborts, silences, Discord thread. One call per resolved incident. Keep human editing in the loop. [LOW]
3. Ask-the-platform chat via MCP. Run the VM and VL MCP servers read-only and behind vmauth tokens scoped to the caller's project. Logs and repository content are untrusted data, a prompt-injection path: never let the model act on instructions found in logs (the CNP MCP guidance itself says this). [MED]
4. Natural language to MetricsQL/LogsQL in the portal. Constrain: provide the app's metric and label catalog as context, validate the generated query against the backend with limits, always show the query. Treat as a learning aid, not a source of truth. [LOW]
5. Skip: autonomous remediation, anomaly detection by LLM, per-log-line AI. The Grafana 2026 survey says 95% want visible reasoning and only 50% trust AI deploying changes [HIGH on stat, MED on population].

Cost controls: trigger-based not polling, budget cap per day, a small model by default, escalate to a larger model only on user request, store summaries to avoid recomputation, log token usage as a metric on the platform dashboard (a good FinOps demo itself).

---

## 8. Demo "wow" features vs gimmicks (2026)

Strong demo value, low or medium cost:
1. Deploy with a canary, SLO-driven auto-rollback, live: deploy marker on the dashboard, AnalysisRun failing, rollout aborted, Discord message with AI summary and runbook link. Best end-to-end story.
2. Auto-generated per-app dashboard and SLO burn-rate panel created at app creation, zero config.
3. One-click trace-to-logs and log-to-trace (needs trace_id in logs and Grafana correlations).
4. Live service map: Hubble UI is free with Cilium (L7 needs L7 policies, `labelsContext` for namespace labels [MED]). Grafana node graph from trace metrics as a second option.
5. Flamegraph via Pyroscope with Alloy eBPF, shown for a deliberately slow endpoint.
6. Cost per project and cost per 1k requests, plus requests-vs-usage waste for the FinOps role.
7. Core Web Vitals per React app, per release.
8. Auto-generated public status page per project (Gatus).
9. Observability readiness scorecard in the portal.
10. DORA panel computed from platform events.

Gimmicks or low return for this context:
- Kepler/carbon on VMs. Kepler was re-architected in 2026 (no eBPF, reads /proc and /sys, needs RAPL or a power meter). Bare-metal CPU attribution is the priority, VM support relies on an estimation model that "needs to be (re-)trained", accuracy numbers are the authors' own, still CNCF Sandbox as far as I found. [HIGH for the CNCF post content]. Show it only if nodes are bare metal.
- Session replay (see 5.2).
- AI anomaly detection (vmanomaly is Enterprise).
- Full LLM observability, unless tenant apps use LLMs.
- 3D or exotic topology visuals without data quality behind them.
- Running DevLake, Sentry, ClickHouse-based stacks, Keep, or Backstage solely for the demo. They multiply RAM and ops for little extra signal.

---

## 9. Suggested phasing (proposal)

Phase 1, platform baseline: VM + VictoriaLogs + Grafana + vmalert + Alertmanager with `project` routing and Discord receivers. Templates emit RED dashboards (Envoy metrics), standard alert rules with `runbook_url`. Deploy annotations via Argo CD Notifications.
Phase 2, reliability story: Sloth-generated SLOs and burn-rate alerts, Argo Rollouts AnalysisTemplates reusing the SLIs, Gatus status pages, portal alerts tab and scorecard.
Phase 3, depth, opt-in: OTel Operator annotations and VictoriaTraces (flagged experimental), Faro in React template, Pyroscope, CNPG query metrics, OpenCost or DIY cost, DORA from platform events.
Phase 4, cheap AI: alert summary on Discord, read-only MCP, postmortem draft.

---

## 10. Gaps, conflicts and things I could not verify

- No primary per-product Datadog usage ranking. Ranking in section 1 is judgment.
- Grafana OnCall archive date conflict: README 2026-03-24 vs GitHub banner 2026-06-05.
- VictoriaTraces: no stability statement found, no GA announcement, service graph support unconfirmed. Docs call the Grafana query path not stable.
- Pyrra with VictoriaMetrics: unconfirmed. Sloth has a VM validation plugin, but not an end-to-end vmalert guide found.
- Grafana SLO app on OSS: unconfirmed.
- Grafana OSS datasource permissions and org_mapping availability in 13.x: not verified.
- Perses project-level RBAC and operator details: not verified.
- Faro to VictoriaLogs/VictoriaTraces pipeline and source map support: not verified, plan a spike.
- Gatus Kubernetes auto-discovery: not confirmed in core docs.
- Security signals (Tetragon/Falco), k6, Checkly, cState, Grafana IRM beyond pricing: not researched in depth.
- Resource (RAM/CPU) figures: no benchmark read. VictoriaMetrics' "3.7x less RAM than Tempo" is a vendor claim without methodology.
- Several Grafana-survey statistics for the Oct 2026 ETR report come from press coverage, not the report.
- Third-party blog posts (oneuptime.com, dev.to, aggregators) were used only as secondary pointers.
- Datadog Q3 2026 numbers come from earnings summaries, not the filing.

---

## Sources (all accessed 2026-11-13; "fetched" = page opened, "search" = relied on search-result summary)

Surveys and trend sources
- https://grafana.com/press/2026/03/18/grafana-labs-4th-annual-observability-survey-reveals-a-field-at-a-crossroads-ai-economics-complexity-and-the-enduring-power-of-open-source/ (fetched)
- https://grafana.com/press/2026/10/19/survey-ai-adoption-is-outpacing-trust-as-enterprises-push-toward-autonomous-operations/ (search)
- https://grafana.com/blog/observability-survey-AI-2026 (search)
- https://grafana.com/blog/observability-survey-takeaways (search, 2025 survey)
- https://www.cncf.io/blog/2026/05/06/the-tools-are-ready-so-why-are-most-cloud-native-teams-still-running-three-observability-stacks/ (fetched)
- https://www.cncf.io/blog/2026/01/20/kubernetes-fuels-ai-growth-organizational-culture-remains-the-decisive-factor/ (search)
- https://redmonk.com/rstephens/2025/12/18/dora2025/ (search)
- https://services.google.com/fh/files/misc/2025_state_of_ai_assisted_software_development.pdf (search)
- https://www.gartner.com/doc/6688834 (search)
- https://www.honeycomb.io/blog/why-observability-2-0-gamechanger (search)
- https://charity.wtf/2024/11/ (search)
- https://grafana.com/events/obsessions/2025/boston/are-we-ready-for-observability-2-0-pragmatic-observability-at-simplisafe/ (search)
- https://www.marketbeat.com/instant-alerts/transcript-datadog-q3-earnings-call-highlights-2026-11-04/ (search)
- https://finsee.ai/earnings/ddog/2026/q3/en/ (search)
- https://www.techtarget.com/it-infrastructure/opinion/DASH-2026-Datadog-goes-from-watching-code-to-writing-it (search)
- https://www.helpnetsecurity.com/?p=350215 (search, Bits AI SRE)

VictoriaMetrics ecosystem
- https://github.com/VictoriaMetrics/VictoriaTraces/releases (fetched)
- https://docs.victoriametrics.com/victoriatraces/ (fetched)
- https://docs.victoriametrics.com/victoriatraces/querying/grafana/ (search)
- https://docs.victoriametrics.com/victoriastack/ui/ (fetched)
- https://docs.victoriametrics.com/victorialogs/integrations/perses/ (search)
- https://docs.victoriametrics.com/victoriametrics/integrations/grafana/datasource/ (fetched)
- https://docs.victoriametrics.com/anomaly-detection/quickstart (search)
- https://docs.victoriametrics.com/vmauth (search)
- https://docs.victoriametrics.com/victoriastack/multitenancy/ (search)
- https://docs.victoriametrics.com/victoriastack/access-control/ (search)
- https://newreleases.io/project/github/VictoriaMetrics/VictoriaLogs/release/v1.50.0 (search, VictoriaLogs release stream)
- https://victoriametrics.com/blog/our-latest-updates-across-the-victoriametrics-observability-ecosystem (search)
- https://itnext.io/victoriatraces-recording-rules-metrics-and-alerts-from-trace-spans-abd814e8d34c (search)
- https://victoriametrics.com/blog/victoriametrics-openlit-agents-observability/ (search)
- https://github.com/VictoriaMetrics/mcp-victoriametrics (fetched)
- https://glama.ai/mcp/servers/@VictoriaMetrics-Community/mcp-victorialogs (search)

SLO, alerting, on-call
- https://github.com/slok/sloth and https://github.com/slok/sloth/releases (fetched)
- https://github.com/pyrra-dev/pyrra (fetch failed) and https://github.com/pyrra-dev/pyrra/releases (fetched, redirects to prometheus-community/pyrra)
- https://sre.google/workbook/alerting-on-slos/ (fetched)
- https://grafana.com/products/cloud/slo/ (search)
- https://github.com/grafana/oncall (fetched)
- https://grafana.com/blog/grafana-oncall-maintenance-mode/ (search)
- https://grafana.com/products/cloud/irm (search)
- https://grafana.com/pricing/?pg=oncall (search)
- https://hyperping.com/blog/opsgenie-end-of-life (search)
- https://alertops.com/blogs/opsgenie-end-of-life/ (search)
- https://github.com/keephq/keep (search)
- https://www.ycombinator.com/companies/keep (search)
- https://github.com/prymitive/karma (search)
- https://prometheus.io/docs/alerting/latest/alertmanager/ (fetched)
- https://promlabs.com/blog/2022/12/23/sending-prometheus-alerts-to-discord-with-alertmanager-v0-25-0 (search)
- https://github.com/TwiN/gatus (fetched)
- https://www.bitdoze.com/uptime-kuma-vs-gatus/ (search)
- https://selfhosted-sh.pages.dev/ntfy-push-notifications (search)

Frontend, instrumentation, eBPF
- https://github.com/grafana/faro-web-sdk (fetched)
- https://grafana.com/docs/alloy/next/reference/components/faro/faro.receiver/ (search)
- https://releasebot.io/updates/grafana/grafana-alloy (search)
- https://opentelemetry.io/blog/2026/otel-js-sdk-3-0/ (search)
- https://tessl.io/registry/skills/github/ollygarden/opentelemetry-agent-skills/otel-browser (search)
- https://opentelemetry.io/docs/platforms/kubernetes/operator/automatic/ (fetched)
- https://github.com/open-telemetry/opentelemetry-operator (fetched, little content)
- https://opentelemetry.io/blog/2026/obi-v1-release/ (fetched)
- https://grafana.com/blog/opentelemetry-ebpf-instrumentation-beyla-donation/ (search)
- https://grafana.com/docs/alloy/latest/reference/components/pyroscope/pyroscope.ebpf/ (search)
- https://temps.sh/blog/can-you-self-host-session-replay-2026 (search, competitor source)
- https://oneuptime.com/blog/post/2026-08-12-session-replay-without-leaking-pii/view (search)

Platform patterns
- https://argo-rollouts.readthedocs.io/en/stable/features/analysis/ (fetched)
- https://argo-cd.readthedocs.io/en/stable/operator-manual/notifications/services/grafana/ (search)
- https://github.com/perses/perses (fetched)
- https://www.cncf.io/projects/perses/ (fetched)
- https://grafana.com/docs/grafana/v13.0/as-code/observability-as-code/ (search)
- https://grafana.com/docs/grafana/latest/setup-grafana/configure-access/configure-authentication/keycloak/ (search)
- https://github.com/grafana/mcp-grafana (fetched)
- https://www.cncf.io/blog/2026/06/30/kepler-re-architected-improved-power-accuracy-and-a-community-call-to-action/ (fetched)
- https://grafana.com/orgs/isovalent/dashboards (search, Hubble dashboards)
- https://dev.co/observability/open-source/coroot and https://www.helpnetsecurity.com/2026/02/23/coroot-open-source-observability-apm-tool/ (search)
- https://devlake.apache.org/docs/DORA (search)
- https://about.gitlab.com/blog/how-to-build-ci-cd-observability-at-scale/ and https://docs.gitlab.com/operations/observability/ci_cd/ (search)
- https://cloudnative-pg.io/documentation/1.23/monitoring (search) and https://grafana.com/grafana/dashboards/20417-cloudnativepg/ (search)
- https://lucaberton.com/blog/postgresql-monitoring-per-query-duration-gap/ (search)
- https://dev.to/selfhostingsh/glitchtip-vs-sentry-206o and https://1vps-o5g.pages.dev/self-hosted-app/sentry (search, third-party RAM figures)
- https://www.cncf.io/blog/2026/01/07/holmesgpt-agentic-troubleshooting-built-for-the-cloud-native-era/ (search)
- https://k8sgpt.ai (search)
- https://futureagi.com/blog/best-open-source-llm-observability-2026 and https://dev.to/gabrielanhaia/opentelemetry-genai-semantic-conventions-your-llm-traces-should-look-like-this-in-2026-3ff6 (search, third-party)
- https://docs.port.io/solutions/engineering-intelligence/use-cases/reliability/measure-reliability and https://www.opslevel.com/use-cases/site-reliability-engineering (search, vendor)
