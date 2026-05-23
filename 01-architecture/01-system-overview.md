# System Overview: Cloud Native Platform (CNP)

## 1. Architectural Paradigm
The CNP relies on a decoupled, asynchronous, and GitOps-driven architecture. 
Instead of the Cloud Management Platform (CMP) directly interfacing with Kubernetes API servers to deploy pods, it acts as an **Orchestrator of Orchestrators**. 

The Single Source of Truth (SSOT) is always Git.

## 2. Global Component Interaction
The system is divided into four functional planes:

### A. The Control Plane (CMP Backend & Frontend)
- **Role:** The user-facing portal (IDP).
- **Stack:** Next.js (Frontend), FastAPI (Backend), SQLite/PostgreSQL, Boto3/OpenStackSDK.
- **Responsibility:** Captures user intent (e.g., "Deploy a React app for Project X"). It stores metadata, triggers Terraform via background tasks, and polls health metrics.

### B. The Provisioning Plane (Terraform)
- **Role:** Infrastructure and Bootstrapping.
- **Responsibility:** 
  1. Clones the requested Git App Template into a new Private GitHub Repository.
  2. Creates Vault Paths (`kvv2/projects/<name>`).
  3. Configures Keycloak (Creates Groups/Roles for the new Project).
  4. Configures Cloudflare DNS records.
  5. Generates the `Application` Custom Resource Definition (CRD) for ArgoCD.
  6. Creates the ArgoCD Repository Secret to allow ArgoCD to read the private GitHub repo.

### C. The Delivery Plane (ArgoCD & Helm)
- **Role:** Continuous Deployment (GitOps).
- **Responsibility:** ArgoCD monitors the application's private GitHub repository. It reads the `values.yaml` provided by the developers and applies the "Generic Microservice Helm Chart" to the Kubernetes cluster. Drift detection is handled natively by ArgoCD.

### D. The Security & Network Plane (K8s Native)
- **Cilium CNI:** Enforces NetworkPolicies to isolate traffic strictly within the boundaries of a Project.
- **Envoy Gateway:** Handles ingress traffic, terminating TLS, and enforcing Keycloak OIDC Authentication at the edge.
- **Vault Secrets Operator (VSO):** Synchronizes Vault secrets into Kubernetes native `Secret` objects, allowing pods to consume them securely.

## 3. High-Level Diagram
```text
[ Developer UI ] ---> (FastAPI CMP) ---> [ Terraform Executor ]
                                                |
                                                |-- 1. Create Repo (GitHub)
                                                |-- 2. Setup SSO (Keycloak)
                                                |-- 3. Setup Secrets (Vault)
                                                |-- 4. Create App CRD (ArgoCD)
                                                
[ ArgoCD ] <--- Monitors --- [ App GitHub Repo (values.yaml) ]
    |
    |-- Applies --> [ Generic Helm Chart ]
    |-- Deploys --> [ Kubernetes Cluster (Namespace) ]
```
