# Application Provisioning Workflow

This document details the exact sequence of events when a developer requests a new application via the CMP.

## The Trigger
The user selects a template (e.g., "FastAPI + React Stack") from the CMP Catalog, fills out the basic variables (App Name, Project Name, Replicas), and clicks **Deploy**.

## Phase 1: CMP & Terraform Bootstrapping
The CMP Backend enqueues an asynchronous background task to execute Terraform.

1. **Dynamic Token Retrieval:**
   - The CMP Backend fetches the Project's GitHub `installation_id` from the DB.
   - It requests a short-lived **Installation Access Token** from GitHub using the CNP App Private Key.
   - It injects this temporary token as a Terraform variable.

2. **GitHub Provisioning (Terraform):**
   - Terraform uses the temporary token to create a new Private GitHub repository inside the developer's chosen Organization or User account.
   - It pushes the initial "Git App Template" codebase (including the default `values.yaml` configurations) to the repository.
   - It sets up branch protection on the `main` branch.

3. **Kubernetes Namespace:** Terraform creates the K8s namespace `<project-name>-<app-name>`.

4. **Secret Generation:** Terraform generates initial random credentials (e.g., Database passwords) and stores them in Vault at `kvv2/projects/<project-name>/<app-name>`.

5. **ArgoCD Registration:**
   - Terraform configures ArgoCD to access the private repository. It creates a Kubernetes Secret in the `argocd` namespace containing the GitHub App credentials (App ID, Installation ID, and Private Key).
   - Terraform applies the ArgoCD `Application` Custom Resource (CR) pointing to the repository.

```mermaid
graph TD
    %% --- Nodes ---
    dev["👤 Developer"]
    ui["💻 CMP Frontend - Next.js"]
    api["⚙️ CMP Backend - FastAPI"]
    db[("💾 CMP Database - SQLite")]
    keycloak["🔑 Keycloak - SSO and User Profile"]
    tf["🛠️ Terraform Executor"]
    s3[("📦 S3 Bucket - TF States")]
    gh_app["🐙 GitHub App - CNP-Portal"]
    vault["🔒 HashiCorp Vault - Secrets"]
    cloudflare["🌐 Cloudflare - DNS"]
    argocd["🐙 ArgoCD - GitOps Controller"]
    k3s_ns["☸️ K3s Target Namespace"]

    %% --- Actions & Flows ---
    dev -->|1. Fills form and clicks Deploy| ui
    ui -->|2. Sends POST deployments| api
    
    api -->|3a. Read user profile and installation ID| keycloak
    api -->|3b. Store PENDING status| db
    
    api -->|4a. Sign JWT with Private Key| gh_app
    gh_app -->|4b. Return temporary token| api
    
    api -->|5. Enqueue background task| tf
    tf -->|6. Dynamic init and lock state| s3
    
    tf -->|7a. Create private repo and push code| gh_app
    tf -->|7b. Create path and write secrets| vault
    tf -->|7c. Create CNAME record| cloudflare
    tf -->|7d. Apply ArgoCD App CRD and Repo Secret| k3s_ns

    k3s_ns -->|8a. Exposes App CRD| argocd
    argocd -->|8b. Pulls values yaml via App Creds| gh_app
    argocd -->|8c. Mounts Secrets via K8s Role| vault
    argocd -->|8d. Deploys Generic Chart| k3s_ns
    
    api -->|9a. Queries App Health| argocd
    api -->|9b. Updates status to RUNNING| db
    ui -->|9c. Polls status every 3s| api
```

## Phase 2: GitOps Delivery (ArgoCD)
Once Terraform successfully completes Phase 1, the CMP marks the deployment state as `RUNNING` in its internal database. From this point forward, ArgoCD is fully responsible for the state of the cluster.

1. ArgoCD detects the new `Application` CR.
2. It uses the GitHub App credentials to dynamically authenticate with GitHub and pull the `values.yaml` from the private repository.
3. It fetches the "Generic Microservice Helm Chart" from the central Helm registry.
4. It merges the values and deploys the resulting manifests (Deployments, Services, Ingress, VaultSecret) into the application's namespace.

## Phase 3: Developer Day-2 Operations
- The developer clones their new private repository.
- To change infrastructure settings (e.g., scale from 2 to 5 replicas), the developer edits the `values.yaml` located in the `deploy/` folder of their repo.
- The developer commits and pushes the change to the `main` branch.
- ArgoCD automatically detects the drift, syncs the new replica count, and updates the cluster without any intervention from the CMP or Terraform.