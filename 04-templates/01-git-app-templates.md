# Git App Templates Architecture

To ensure consistency and rapid onboarding, CNP relies on "Git App Templates". These are barebones GitHub repositories containing the boilerplate code for specific tech stacks (e.g., FastAPI, Next.js, ...). 

When a developer requests a new application, the CMP clones the selected template to bootstrap the new project.

## 1. Standard Template Structure
Every Git App Template follows a strict, standardized directory structure. This ensures that the CI/CD pipeline and ArgoCD know exactly where to find the application code and configuration.

```text
template-fastapi-react/
├── .github/
│   └── workflows/
│       └── ci.yml             # Standardized CI/CD pipeline
├── deploy/
│   └── values.yaml            # The GitOps Configuration file for ArgoCD
├── src/                       # Application Source Code
│   ├── main.py
│   └── requirements.txt
├── Dockerfile                 # Multi-stage optimized Dockerfile
├── docker-compose.yml         # For local developer testing
├── .gitignore
└── README.md
```

## 2. Directory Deep Dive

### A. `.github/workflows/ci.yml`
This file is immutable for the developer. It contains the standard CNP pipeline:
1. Triggers on pull requests (Lint & Test) and pushes to `main` (Build & Push).
2. Authenticates with GitHub Container Registry (GHCR) using the native `${{ secrets.GITHUB_TOKEN }}`.
3. Builds the `Dockerfile` and pushes the image tagged with the commit SHA (e.g., `sha-a1b2c3d`).

### B. `src/` and `Dockerfile`
- The `src/` folder contains the actual business logic. 
- The `Dockerfile` must expose the application on a standard port (usually `80` or `8080`) and run without root privileges to comply with Kubernetes security policies (Kyverno).
- A `docker-compose.yml` is provided so developers can run `docker-compose up` on their laptops and get a replica of the environment (including a local database if needed).

### C. `deploy/values.yaml`
This is the **contract** between the developer and the platform. 
It overrides the defaults of the underlying Generic Helm Chart. When the CMP provisions the app, it customizes this file (e.g., injecting the `appName` and `project`).