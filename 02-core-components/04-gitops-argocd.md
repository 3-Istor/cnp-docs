# GitOps Delivery: ArgoCD Configuration

ArgoCD is the core Continuous Delivery (CD) engine of the CNP platform. It continuously reconciles the desired state defined in Git with the live state in the K3s cluster. 

Given CNP's strict Multi-Tenant architecture, ArgoCD is configured to enforce Role-Based Access Control (RBAC) synchronized with Keycloak and project isolation using `AppProject` boundaries.

---

## 1. Single Sign-On (SSO) Integration

ArgoCD delegates authentication to Keycloak. Developers log in to the ArgoCD UI using the exact same credentials they use for the CMP dashboard.

### A. OIDC Configuration (`argocd-cm`)
ArgoCD is configured to connect to the Keycloak `3istor` realm. It requests the `groups` scope, which maps the user's project affiliations into the JWT token.

```yaml
# k8s/config/argocd/patch-argocd-cm.yaml
oidc.config: |
  name: 3Istor ID
  issuer: https://auth.3istor.com/realms/3istor
  clientID: argocd
  # The secret is injected by Vault Secrets Operator
  clientSecret: $oidc.keycloak.clientSecret
  requestedScopes: ["openid", "profile", "email", "groups"]
```

### B. Global RBAC Policies (`argocd-rbac-cm`)
ArgoCD maps the Keycloak groups to specific internal roles. By default, access is denied unless explicitly granted.

```yaml
# k8s/config/argocd/patch-argocd-rbac-cm.yaml
policy.default: "role:denied"
scopes: "[groups]"
policy.csv: |
  # Global Infrastructure Admins (Full cluster access)
  g, infra, role:admin
  g, k3s-admin, role:admin
  
  # Platform-level read-only access
  g, k3s-view, role:readonly
```
*Note: Project-specific permissions are not defined here globally. They are dynamically defined per `AppProject`.*

---

## 2. Multi-Tenancy Boundaries (`AppProject`)

By default, ArgoCD places all `Application` CRDs into a default project. In CNP, we isolate teams using ArgoCD's native `AppProject` Custom Resource.

When Terraform provisions a new Project in the CMP, it automatically creates an `AppProject` inside ArgoCD.

### The Dynamic AppProject Specification
The `AppProject` acts as a sandbox. It defines:
1. **Destinations:** Which Kubernetes Namespaces this project is allowed to deploy to.
2. **Sources:** Which GitHub repositories it is allowed to pull from.
3. **RBAC:** Which Keycloak groups are allowed to manage these applications.

**Example: Terraform-generated AppProject for "Project Alpha"**
```yaml
apiVersion: argoproj.io/v1alpha1
kind: AppProject
metadata:
  name: project-alpha
  namespace: argocd
spec:
  # 1. Allowed Source Repositories
  # Project Alpha can only deploy code from its specific GitHub organization
  sourceRepos:
    - "https://github.com/my-org/project-alpha-*"

  # 2. Allowed Cluster Destinations
  # Apps in Project Alpha can only be deployed to namespaces starting with project-alpha-
  destinations:
    - server: "https://kubernetes.default.svc"
      namespace: "project-alpha-*"

  # 3. Project-Level RBAC mapping to Keycloak Groups
  roles:
    # Project Admins: Can view, sync, and override applications
    - name: alpha-admins
      description: Admins of Project Alpha
      policies:
        - p, proj:project-alpha:alpha-admins, applications, get, project-alpha/*, allow
        - p, proj:project-alpha:alpha-admins, applications, sync, project-alpha/*, allow
        - p, proj:project-alpha:alpha-admins, applications, override, project-alpha/*, allow
      groups:
        - project-alpha-admins # Matches the Keycloak Group

    # Project Members: Can view and sync applications
    - name: alpha-members
      description: Members of Project Alpha
      policies:
        - p, proj:project-alpha:alpha-members, applications, get, project-alpha/*, allow
        - p, proj:project-alpha:alpha-members, applications, sync, project-alpha/*, allow
      groups:
        - project-alpha-members # Matches the Keycloak Group
```

---

## 3. Application Lifecycle Management

When Terraform creates a new Application (e.g., `app-frontend` inside `project-alpha`), it generates an `Application` CRD assigned to the isolated `AppProject`.

### The Application CRD
```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: project-alpha-app-frontend
  namespace: argocd
  annotations:
    # Essential when using Kyverno to prevent continuous OutOfSync loops 
    # caused by mutating webhooks injecting default values.
    argocd.argoproj.io/compare-options: ServerSideDiff=true,IncludeMutationWebhook=true
spec:
  project: project-alpha  # Maps to the AppProject sandbox
  source:
    repoURL: https://github.com/my-org/project-alpha-app-frontend.git
    targetRevision: HEAD
    path: .
    helm:
      # Renders the Central Generic Helm Chart dynamically
      valueFiles:
        - deploy/values.yaml
  destination:
    server: https://kubernetes.default.svc
    namespace: project-alpha-app-frontend
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
      - ServerSideApply=true
      - CreateNamespace=true
```

---

## 4. Kyverno Mutual Exclusion Best Practices
As noted in the `metadata.annotations`, the cluster enforces policies via Kyverno (e.g., requiring explicit memory limits). 

When ArgoCD attempts to apply a manifest, Kyverno mutating webhooks may inject missing defaults (e.g., setting a `runAsUser` if omitted). By default, ArgoCD detects this injected value as a drift from Git and attempts to revert it, causing an infinite sync loop.

CNP completely mitigates this by enabling **Server-Side Apply** and **ServerSideDiff**:
1. `ServerSideApply=true`: ArgoCD lets the Kubernetes API handle the merge resolution.
2. `ServerSideDiff=true`: ArgoCD computes the drift against the server's post-mutation state rather than the dry-run client state.