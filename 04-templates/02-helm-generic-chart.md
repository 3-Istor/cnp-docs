# Central Generic Helm Chart Design

Rather than maintaining custom Kubernetes manifests for every single application, CNP relies on a single **Generic Helm Chart** (also known as a Library or Umbrella Chart).

ArgoCD points to this central Helm Chart and merges the specific `deploy/values.yaml` from the application's repository.

## 1. Chart Architecture & Structure
The Generic Chart contains templates for every possible Kubernetes resource an application might need. Resources are rendered conditionally based on the `values.yaml`.

```text
cnp-generic-chart/
├── Chart.yaml
├── values.yaml                     # Default fallback values
└── templates/
    ├── _helpers.tpl                # Naming conventions & labels
    ├── deployment.yaml             # K8s Deployment
    ├── service.yaml                # K8s Service
    ├── ingress-httproute.yaml      # Envoy Gateway HTTPRoute
    ├── envoy-securitypolicy.yaml   # Edge SSO Auth (Conditional)
    ├── vaultsecret.yaml            # Ricoberger VaultSecret (Conditional)
    ├── database-cnpg.yaml          # CloudNativePG Cluster (Conditional)
    └── servicemonitor.yaml         # Prometheus Monitoring (Conditional)
```

---

## 2. Conditional Rendering Logic

The power of the Generic Chart lies in Go template `if` statements.

### Example: Envoy Gateway & SSO Integration
Instead of writing complex Envoy Gateway CRDs, the developer simply sets `sso_protected: true` in their `values.yaml`. The Helm Chart handles the translation:

```yaml
# templates/envoy-securitypolicy.yaml
{{- if and .Values.ingress.enabled .Values.ingress.sso_protected }}
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: {{ include "cnp.fullname" . }}-security
  namespace: {{ .Release.Namespace }}
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      name: {{ include "cnp.fullname" . }}-route
  oidc:
    provider:
      issuer: "https://auth.3istor.com/realms/3istor"
      # ... OIDC config mapping ...
    clientID: "3-istor-openid"
    clientSecret:
      name: {{ include "cnp.fullname" . }}-envoy-auth-secret
{{- end }}
```

### Example: Vault Secrets Injection
If the application needs secrets (Database passwords, API keys), the chart generates the VSO custom resource.

```yaml
# templates/vaultsecret.yaml
{{- if .Values.secrets.enabled }}
apiVersion: ricoberger.de/v1alpha1
kind: VaultSecret
metadata:
  name: {{ include "cnp.fullname" . }}-secrets
spec:
  # The path is injected by Terraform during app creation
  path: {{ .Values.secrets.vaultPath }}
  vaultRole: {{ include "cnp.fullname" . }}-role
  type: Opaque
  keys:
    - database-password
    - api-key
{{- end }}
```

---

## 3. Database Provisioning via CloudNativePG (CNPG)
When an application requires a database, the Generic Chart does not rely on heavy subcharts. Instead, it conditionally renders a native **CloudNativePG Cluster** resource. 

This model is multi-cloud ready, allowing cross-cluster streaming replication and S3-backed continuous archiving (via Garage S3 or AWS S3).

```yaml
# templates/database-cnpg.yaml
{{- if and .Values.database.enabled (eq .Values.database.type "postgres") }}
apiVersion: postgresql.cnpg.io/v1
kind: Cluster
metadata:
  name: {{ include "cnp.fullname" . }}-db
  namespace: {{ .Release.Namespace }}
spec:
  instances: {{ .Values.database.instances | default 3 }} # Default 3 for High Availability (Primary + 2 Replicas)

  # Storage configuration using local-path or fast-ssd
  storage:
    size: {{ .Values.database.storageSize | default "2Gi" }}
    storageClass: {{ .Values.database.storageClass | default "local-path" }}

  # Multi-cloud continuous backup & archiving configuration
  backup:
    barmanObjectStore:
      destinationPath: "s3://{{ .Values.database.backupBucket }}"
      endpointURL: "https://s3.3istor.com" # Points to Garage S3 / AWS S3
      s3Credentials:
        name: {{ include "cnp.fullname" . }}-s3-creds
      wal:
        compression: gzip
{{- end }}
```

**Developer Configuration (`deploy/values.yaml`):**
To request a resilient database, the developer only needs to add this simple block:
```yaml
database:
  enabled: true
  type: "postgres"
  instances: 3
  storageSize: "10Gi"
  backupBucket: "project-backups"
```