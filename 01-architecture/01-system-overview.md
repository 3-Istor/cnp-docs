# System Overview: Cloud Native Platform (CNP)

## 1. Architectural Paradigm
The CNP relies on a decoupled, asynchronous, and GitOps-driven architecture. Instead of the Cloud Management Platform (CMP) directly interfacing with Kubernetes API servers to deploy pods, it acts as an **Orchestrator of Orchestrators**. 

The Single Source of Truth (SSOT) is always Git.

---

## 2. Global Component Interaction

The platform is structured into four functional planes, ensuring clear segregation of duties:

```mermaid
flowchart TD
    subgraph Control_Plane [Control Plane - Next.js & FastAPI]
        CMP[CMP IDP Portal] <-->|DB State| SQLite[(arcl.db)]
        CMP -->|Enqueue Tasks| SAGA[Saga Orchestrator]
    end

    subgraph Provisioning_Plane [Provisioning Plane - Terraform]
        SAGA -->|1. Run Day-0 Bootstrapping| TF[Terraform Runner]
        TF -->|Create Private Repo| GH_API[GitHub App API]
        TF -->|Generate Project Groups| KC[Keycloak Admin API]
        TF -->|Setup Secrets Path| Vault[HashiCorp Vault]
        TF -->|Deploy App CRD| Argo[ArgoCD API]
    end

    subgraph Delivery_Plane [Delivery Plane - GitOps]
        Argo <-->|Continuous Reconciliation| GH_Repo[(GitHub Private Repo)]
        Argo -->|Deploys Standard Chart| K3s[K3s Cluster Workloads]
    end

    subgraph Security_Plane [Security & Network Plane - Native]
        K3s -->|Network Isolation| Cilium[Cilium CNI]
        K3s -->|Edge Routing & OIDC| Envoy[Envoy Gateway]
        K3s -->|Secret Syncing| VSO[Vault Secrets Operator]
        VSO <-->|Fetch Decrypted Values| Vault
    end

    %% Styles
    classDef default fill:#f9f9f9,stroke:#333,stroke-width:1px;
    classDef control fill:#eef2ff,stroke:#4f46e5,stroke-width:2px;
    classDef delivery fill:#f0fdf4,stroke:#16a34a,stroke-width:2px;
    classDef security fill:#fff1f2,stroke:#e11d48,stroke-width:2px;
    
    class Control_Plane,CMP,SAGA,SQLite control;
    class Delivery_Plane,Argo,K3s delivery;
    class Security_Plane,Cilium,Envoy,VSO,Vault,KC security;
```

---

### A. The Control Plane (CMP Backend & Frontend)
* **Role**: The user-facing portal (IDP).
* **Stack**: Next.js (Frontend), FastAPI (Backend), SQLite/PostgreSQL, Boto3/OpenStackSDK.
* **Responsibility**: Captures user intent (e.g., "Deploy application X into Project Y"). It manages local metadata, coordinates async background tasks using a strict Saga pattern, and hosts the real-time health-polling dashboard.

### B. The Provisioning Plane (Terraform)
* **Role**: Infrastructure and Bootstrapping (Day-0).
* **Responsibility**: 
  1. Clones the requested Git App Template into a newly generated private GitHub repository.
  2. Creates Vault KV paths (`kvv2/projects/<project_name>/<app_name>`).
  3. Configures Keycloak (generates Tenant Realms, Client IDs, and group mappings).
  4. Configures Cloudflare DNS and dedicated application micro-tunnels.
  5. Deploys the ArgoCD `Application` Custom Resource (CR) and repository access secrets.

### C. The Delivery Plane (ArgoCD & Helm)
* **Role**: Continuous Deployment & Reconciliation (Day-1).
* **Responsibility**: ArgoCD monitors the application's private GitHub repository, reads the `deploy/values.yaml` file, and compiles it against the centralized generic Helm chart. Drift detection and automated syncs are handled natively.

### D. The Security & Network Plane (K8s Native)
* **Cilium CNI**: Enforces network policies (`CiliumClusterwideNetworkPolicy`) to isolate tenant namespaces.
* **Envoy Gateway**: Terminates incoming TLS, routes external traffic using standard Gateway API `HTTPRoute` resources, and enforces Keycloak OIDC at the edge via a `SecurityPolicy`.
* **Vault Secrets Operator (VSO)**: Automatically synchronizes Vault secrets into native K8s `Secret` objects using service account tokens.

---
**Next Step**: Continue to [Multi-Tenancy & Isolation Model](02-tenancy-and-isolation.md).
