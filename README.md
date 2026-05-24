# Cloud Native Platform (CNP) ⚡ - Documentation

## 📖 Project Overview

The Cloud Native Platform (CNP) is a modern **Internal Developer Portal (IDP)** designed to completely abstract infrastructure complexity. It empowers developers to provision, manage, and monitor containerized applications on Kubernetes through standardized "Golden Paths".

Built on top of a highly robust stack (Next.js, FastAPI, Kubernetes, ArgoCD, Terraform, Vault, and Keycloak), CNP shifts the paradigm from traditional IaaS (VM deployment) to modern PaaS/CaaS (Platform/Container as a Service).

## 🎯 Core Objectives

1. **Developer Autonomy:** Self-service provisioning of full-stack applications via a visual dashboard (CMP) or a GitOps configuration file (`cnp.yaml`).
2. **Strict Multi-Tenancy:** A hierarchical `Project > Application` model ensuring absolute isolation (Network, Secrets, RBAC) between different teams.
3. **End-to-End Automation:** Seamless integration from GitHub repository creation, CI/CD bootstrapping, Kubernetes deployment (Helm + ArgoCD), to DNS and Secret injection.

## 📚 Documentation Directory

This repository contains the architectural decisions, workflows, and specifications of the CNP ecosystem. It is structured for both human engineers and AI assistants to quickly grasp the system design.

- **[01. Architecture & Topologies](./01-architecture/)**: System overviews, Tenancy models, and Networking.
- **[02. Core Components](./02-core-components/)**: Deep dives into the CMP, Keycloak, Vault, and ArgoCD configurations.
- **[03. Pipelines & Workflows](./03-pipelines-and-workflows/)**: Sequential flows for app provisioning, CI/CD, and the `cnp.yaml` GitOps cycle.
- **[04. Templates Design](./04-templates/)**: Specifications for Git App Templates, Generic Helm Charts, and Terraform wrappers.
- **[05. CMP Backend API](./05-cmp-backend-api/)**: ⭐ **NEW** - API specifications and integration guides for the Cloud Management Platform.

## 📋 Project Documentation

- **[CHANGELOG.md](../CHANGELOG.md)**: Version history and release notes
- **[README_ROADMAP.md](./README_ROADMAP.md)**: Implementation roadmap and phase tracking
- **[PHASE3_TEAM_HANDOFF.md](../PHASE3_TEAM_HANDOFF.md)**: Team handoff document for Phase 3
- **[PHASE3_SUMMARY.md](../PHASE3_SUMMARY.md)**: Quick summary of Phase 3 changes

```
cnp-docs/
├── README.md                               # Project pitch and documentation index
├── README_ROADMAP.md                       # Implementation roadmap
├── 01-architecture/
│   ├── 01-system-overview.md               # Global architecture (Frontend, Backend, K8s, Git)
│   ├── 02-tenancy-and-isolation.md         # Project > App model (RBAC, Network, Vault)
│   └── 03-network-topology.md              # Envoy Gateway, Cilium, Ingress, DNS
├── 02-core-components/
│   ├── 01-cmp-dashboard.md                 # Frontend/Backend specifications
│   ├── 02-identity-keycloak.md            # SSO flows, Groups, OIDC
│   ├── 03-secrets-vault.md                # Secrets management, VSO
│   └── 04-gitops-argocd.md               # Cluster synchronization
├── 03-pipelines-and-workflows/
│   ├── 01-app-provisioning-flow.md       # App creation sequence (Terraform -> Git -> ArgoCD)
│   ├── 02-ci-cd-pipelines.md             # GitHub Actions (Build, Push)
│   └── 03-developer-cnp-yaml.md          # cnp.yaml spec and translation
├── 04-templates/
│   ├── 01-git-app-templates.md           # Application repository structure
│   ├── 02-helm-umbrella-chart.md         # Generic Helm chart
│   └── 03-terraform-provisioner.md       # Terraform for infra and GitHub initialization
└── 05-cmp-backend-api/                    # ⭐ NEW - CMP API Documentation
    ├── 01-cmp-deployment-api.md           # Deployment API specification
    ├── 02-cmp-phase3-changes.md           # Phase 3 changes summary
    └── 03-cmp-frontend-integration.md     # Frontend integration guide
```
