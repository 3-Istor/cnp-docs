# GitHub Repositories Landscape

The CNP ecosystem is distributed across several specialized GitHub repositories to ensure a clear separation of concerns.

## 1. Platform Core
| Repository | Purpose | Status / Action Needed |
| :--- | :--- | :--- |
| **`arcl-cmp`** | The Cloud Management Platform (Next.js + FastAPI). Contains the UI and the Backend API orchestrating Terraform and interacting with Keycloak/ArgoCD. | **WIP:** Needs refactoring to support the `Project > Application` model and GitHub App integration. |
| **`cnp-docs`** | The central architecture and documentation repository. Linked as a Git Submodule in the CMP for AI Agents (Kiro). | **Ready:** Contains the target architecture specs. |
| **`K3s`** | The central GitOps repository for the base infrastructure (ArgoCD "App of Apps", Cert-Manager, Vault, Kyverno, etc.). | **Active:** Manages the cluster foundation. |

## 2. Application Templates (Day-0 Bootstrapping)
These repositories are cloned by the CMP when a developer creates a new application.
| Repository | Purpose | Status / Action Needed |
| :--- | :--- | :--- |
| **`template-app-webapp-python-fastapi-react`** | Boilerplate for a full-stack Python/React app. | **WIP:** Needs the `deploy/values.yaml` folder added. |
| **`template-html-css`** | Boilerplate for a static website. | **WIP:** Needs the `deploy/values.yaml` folder added. |

## 3. Kubernetes Delivery (Day-1 & Day-2)
| Repository | Purpose | Status / Action Needed |
| :--- | :--- | :--- |
| **`infra-templates`** | *Target:* The central Helm repository hosting the `cnp-generic-chart`. ArgoCD will pull this chart to deploy applications. | **Critical Rewrite Needed:** Currently contains raw Kustomize/K8s manifests. Must be converted into a true parameterized Helm Chart. |

## 4. Deprecated / Archive
| Repository | Purpose | Status / Action Needed |
| :--- | :--- | :--- |
| **`app-templates`** | Contains the legacy Terraform templates for OpenStack/AWS (VMs, ASG, Octavia LBs). | **Archive/Reference:** The K8s transition deprecates the need to apply these manually. Terraform logic will move directly into the CMP backend modules. |