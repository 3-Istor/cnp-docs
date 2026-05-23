# Execution Roadmap: Transitioning to the CNP PaaS Model

This roadmap outlines the step-by-step transition from the legacy Hybrid IaaS broker to the modern Kubernetes GitOps Platform. It is designed to be executed sequentially, allowing safe iteration and testing at each phase.

---

## Phase 1: K3s Cluster & Platform Foundations (Operations)
*Goal: Prepare the Kubernetes cluster to receive dynamic applications from the CMP.*

*   [ ] **1. Deploy Cilium & Envoy Gateway:** Ensure network isolation and edge routing are active.
*   [ ] **2. Deploy Keycloak:** Configure the `3istor` realm, custom RBAC wrapper flows, and base groups (`infra`, `member`).
*   [ ] **3. Deploy Vault & VSO:** Ensure the Vault Secrets Operator is running and OIDC auth is configured with Keycloak.
*   [ ] **4. Deploy ArgoCD & Image Updater:** Configure ArgoCD SSO with Keycloak and setup the global `argocd-rbac-cm`.
*   [ ] **5. GitHub App Creation:** 
    *   Manually create the `CNP-Portal` GitHub App in the GitHub Organization.
    *   Generate the Private Key and Webhook Secret.
    *   Store these credentials securely in Vault.

---

## Phase 2: Helm & Terraform Artifacts (Engineering)
*Goal: Create the reusable templates that the CMP will trigger.*

*   [ ] **1. Build the Generic Helm Chart:** 
    *   Create a single Helm Chart that supports conditional Deployments, Ingress/HTTPRoute (with Envoy Auth), and CloudNativePG.
    *   Publish this chart to your internal Helm Registry or GitHub (GHCR).
*   [ ] **2. Build the Git App Templates:** 
    *   Create template repositories (e.g., `template-fastapi-react`).
    *   Include the standard `.github/workflows/ci.yml` and the default `deploy/values.yaml`.
*   [ ] **3. Create the Day-0 Terraform Module:**
    *   Write the Terraform code that takes dynamic variables (GitHub Token, Project Name, App Name).
    *   *Resources:* `github_repository`, `github_repository_file`, `vault_generic_secret`, `argocd_application`.

---

## Phase 3: CMP Backend Refactoring (FastAPI)
*Goal: Adapt the Python backend to support Multi-Provider deployments without breaking legacy features.*
*(Perfect for AWS Kiro Agent)*

*   [ ] **1. Database Schema Extension:** 
    *   Update `models/deployment.py`. Add `provider_type`, `github_repo_url`, `argocd_app_name`, `k8s_namespace`.
    *   Create an Alembic migration script.
*   [ ] **2. Keycloak Integration Update:** 
    *   Update `account.py` to handle the GitHub App OAuth callback and store the `github_installation_id` as a Keycloak user attribute.
*   [ ] **3. Strategy Pattern for Orchestration:** 
    *   Refactor `saga_orchestrator.py`. If `provider_type == KUBERNETES`, route to the new Terraform Day-0 execution flow (using dynamic S3 keys for the micro-state).
*   [ ] **4. Strategy Pattern for Monitoring:** 
    *   Refactor `monitoring_service.py`. If `provider_type == KUBERNETES`, fetch health metrics from the ArgoCD REST API instead of AWS/OpenStack.

---

## Phase 4: CMP Frontend Refactoring (Next.js)
*Goal: Update the UX to support Projects, the new Catalog, and GitHub Linking.*
*(Perfect for AWS Kiro Agent)*

*   [ ] **1. Account Page:** Add the "Link GitHub Account" button and handle the OAuth redirect flow.
*   [ ] **2. Dual Catalog Views:** Update `page.tsx` and `CatalogGrid.tsx` to separate Legacy IaaS templates from Modern K8s templates.
*   [ ] **3. Deploy Modal:** Update the deployment form to prompt for the target "Project" (fetching the user's allowed Keycloak groups).
*   [ ] **4. Dynamic Deployment Cards:** Update `DeploymentCard.tsx` and `DeploymentHealthPanel.tsx` to render different UI elements (e.g., "Open in ArgoCD" buttons) based on the `provider_type`.

---

## Phase 5: Day-2 Operations & GitOps Write-Back (API)
*Goal: Allow developers to tweak their infrastructure (Replicas, SSO) directly from the CMP Dashboard.*

*   [ ] **1. GitOps API Endpoints:** Create a new FastAPI router (`/api/applications/{id}/config`).
*   [ ] **2. GitHub API Integration:** Write a service that uses the GitHub App Token to fetch the `deploy/values.yaml` from the user's repository, parse it, modify a key (e.g., `replicaCount`), and commit the change back to the `main` branch.
*   [ ] **3. Frontend Toggles:** Build the UI controls in the Application Dashboard to trigger these API endpoints.