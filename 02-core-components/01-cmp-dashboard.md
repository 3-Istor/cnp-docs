# Cloud Management Platform (CMP): Core Dashboard & API

The CMP is the developer-facing IDP (Internal Developer Portal). It supports a **Multi-Provider Hybrid Architecture**, allowing developers to deploy legacy IaaS stacks (AWS ASG + OpenStack DB) alongside modern Cloud-Native GitOps stacks (Kubernetes + ArgoCD) from a single unified interface.

---

## 1. Catalog & UX Restructuring

To prevent confusing the user, the CMP Frontend separates the offerings into distinct catalogs.

### A. The Catalog View
- **Category 1: Infrastructure as a Service (IaaS)**
  - Templates combining Raw VMs, AWS Auto Scaling Groups, and OpenStack nodes (The legacy templates).
- **Category 2: Platform as a Service (K8s / GitOps)**
  - Templates deploying standardized Microservices via ArgoCD to the K3s cluster.

### B. User Onboarding (GitHub Link)
Located in `AccountPage.tsx`. To use the PaaS catalog, users must link their GitHub account.
1. User clicks **"Link GitHub Account"**.
2. Redirected to the GitHub App installation page.
3. Upon callback, the FastAPI backend receives the `installation_id`.
4. **Storage:** The backend updates the user's profile in the **Keycloak API**, saving `github_installation_id` as a custom user attribute. This avoids creating complex user tables in SQLite.

---

## 2. Data Model Extension (FastAPI / SQLAlchemy)

We preserve the existing `deployments` table to avoid breaking the legacy AWS/OpenStack workflows, but we extend it to support the new Multi-Tenant and GitOps paradigms.

### Extended `deployments` Table
- `id` (Integer / Primary Key)
- `name` (String, unique)
- `project_id` (String - References the Keycloak Project Group)
- `template_id` (String)
- `status` (Enum: `PENDING`, `DEPLOYING`, `RUNNING`, `DEGRADED`, `FAILED`, `DELETED`)
- **Discriminator:**
  - `provider_type` (Enum: `LEGACY_HYBRID`, `KUBERNETES`)
- **Legacy Hybrid Fields (Preserved):**
  - `os_vm_db1_id`, `aws_asg_name`, etc.
- **New Kubernetes Fields:**
  - `github_repo_url` (String - Link to the developer's GitOps repo)
  - `argocd_app_name` (String - The CRD name in ArgoCD)
  - `k8s_namespace` (String - The target namespace)

---

## 3. Backend Orchestration (The Saga Pattern)

The `saga_orchestrator.py` is updated to route the deployment workflow based on the template's `provider_type`.

### Path A: Legacy Hybrid Workflow
Maintains the existing logic:
1. Terraform provisions OpenStack VMs (Stateful).
2. Terraform provisions AWS ASG (Stateless).
3. Rollbacks if AWS fails.

### Path B: Kubernetes GitOps Workflow
Executes the new Day-0 provisioning:
1. **GitHub Auth:** Fetch the user's `github_installation_id` from Keycloak, generate a dynamic GitHub App token.
2. **Terraform Bootstrap:** Create the GitHub Repo, setup Branch Protection, inject base `values.yaml`.
3. **Vault & K8s:** Create Vault Path, configure ArgoCD App CRD.
4. ArgoCD takes over the actual pod deployment (Day-1).

---

## 4. Multi-Provider Health Monitoring

The `monitoring_service.py` is refactored using a Strategy Pattern to handle different infrastructure backends.

### A. Global Infrastructure Health
Continues to poll the base components:
- AWS VPN Gateways (Boto3).
- OpenStack Hypervisors (OpenStackSDK).
- Kubernetes Core Components (ArgoCD API Health, Cilium Status).

### B. Application Health (`AppHealthResponse`)
When the frontend requests the health of a specific deployment, the backend checks the `provider_type`:

**If `LEGACY_HYBRID`:**
- Connects to AWS (ASG Target Groups) and OpenStack (VM Status) as currently implemented.

**If `KUBERNETES`:**
- The backend makes a REST API call to the **ArgoCD Server API**.
- `GET /api/v1/applications/{argocd_app_name}`
- Maps ArgoCD statuses to CMP statuses:
  - ArgoCD `Healthy` & `Synced` ➔ CMP `healthy`
  - ArgoCD `Progressing` ➔ CMP `deploying`
  - ArgoCD `Degraded` or `OutOfSync` ➔ CMP `degraded`

---

## 5. Dashboard Capabilities

### A. Project Dashboards
Users can create "Projects" in the CMP.
- Creating a Project invokes the Keycloak API to create a `project-<name>-admin` and `project-<name>-member` group.
- The UI queries Keycloak directly to display the members of the project.

### B. Dynamic Deployment Cards
The `DeploymentCard.tsx` adapts its UI based on the `provider_type`:
- **Legacy Hybrid Apps:** Show outputs for ALB URLs and direct OpenStack IP addresses.
- **Kubernetes Apps:** Show quick-action buttons for:
  - `[ View in ArgoCD ]`
  - `[ Open GitHub Repo ]`
  - `[ Manage Secrets in Vault ]`