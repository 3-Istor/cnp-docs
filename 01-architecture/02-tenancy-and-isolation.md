# Multi-Tenancy & Isolation Model

CNP implements strict multi-tenancy by utilizing native Kubernetes constructs (Namespaces, RBAC), advanced CNI features (Cilium Network Policies), and Identity & Access Management (Keycloak + Vault).

## 1. Hierarchy Definitions
* **Project:** A logical grouping representing a Team or a Business Domain. It acts as the ultimate security boundary.
* **Application:** A deployable unit (or stack of microservices) belonging to a Project.

**Mapping to Kubernetes:**
- 1 Project = 1 ArgoCD `AppProject` + 1 Keycloak Parent Group + 1 Vault Policy.
- 1 Application = 1 Kubernetes `Namespace` + 1 ArgoCD `Application` + 1 Vault Kubernetes Auth Role.

---

## 2. Security Boundaries

### A. Network Isolation (Cilium CNI)
- **Namespace Design:** Application namespaces are named using the pattern `<project_name>-<app_name>`. They are labeled with `project: <project_name>`.
- **Policy Enforcement:** A Cilium `CiliumClusterwideNetworkPolicy` enforces a "Default Deny" posture for cross-project communication. 
- **Intra-Project Communication:** Pods can freely resolve and communicate with other pods located in namespaces sharing the same `project` label. Communication with external projects requires explicit Envoy Gateway routing.

### B. Access Control & Edge Security (Keycloak & Envoy Gateway)
- **Role-Based Access Control:** Every Project provisions two Keycloak groups: `project-<name>-admins` and `project-<name>-members`.
- **Edge OIDC via Envoy:** 
  Applications can be protected without writing any authentication code. The CNP platform provisions an Envoy `SecurityPolicy` (via the `secure-route` Helm chart) linked to the application's `HTTPRoute`. 
  Envoy intercepts requests, validates the session against Keycloak, and checks that the user's JWT contains the specific project group claim before allowing traffic into the cluster.

### C. Secrets Isolation (HashiCorp Vault & VSO)
- **Storage Strategy:** Vault uses the KV V2 engine (`kvv2/`). Each project gets a dedicated path: `kvv2/data/projects/<project-name>/`.
- **Pod Consumption:** The cluster utilizes the `ricoberger.de/v1alpha1` Vault Secrets Operator. 
  Each application namespace receives a dedicated ServiceAccount. Terraform provisions a specific `vault_kubernetes_auth_backend_role` bound strictly to this ServiceAccount, ensuring that Pods in Application A cannot request Vault to decrypt secrets belonging to Application B.