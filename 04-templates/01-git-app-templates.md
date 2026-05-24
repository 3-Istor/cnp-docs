# Git App Templates Architecture

To ensure consistency and rapid onboarding, CNP relies on standardized "Git App Templates" (e.g., `template-react-fastapi`, `template-html-css`). 

When a developer requests a new application via the CMP, the platform clones the chosen template to bootstrap the new project repository.

## 1. Standardized Directory Structure
Every Git App Template MUST follow a strict directory structure. This contract ensures that local development, CI/CD pipelines, and ArgoCD synchronization work flawlessly out of the box.

```text
cloud-native-[stack]-template/
├── .github/
│   └── workflows/
│       └── ci.yml             # Standard CI/CD pipeline (Build & Push to GHCR)
├── deploy/
│   └── values.yaml            # The GitOps Configuration file for ArgoCD
├── src/                       # Application Source Code
│   ├── backend/               # (If applicable)
│   └── frontend/              # (If applicable)
├── Dockerfile                 # Multi-stage optimized Dockerfile (or multiple)
├── docker-compose.yml         # For local developer testing
├── .gitignore
└── README.md
```

## 2. Component Deep Dive

### A. Application Code & Local Dev (`src/`, `Dockerfile`, `docker-compose.yml`)
- **Isolation:** Frontend and Backend code should reside in their respective folders.
- **Containerization:** The `Dockerfile` must expose the application on a standard port (usually `8000` or `3000`) and run as a non-root user to comply with Kyverno security policies.
- **Local Dev:** The `docker-compose.yml` provides a seamless local experience. It spins up the application components and any mocked external services (e.g., a local PostgreSQL container).

### B. The CI/CD Pipeline (`.github/workflows/ci.yml`)
The GitHub Actions workflow is standard across all templates.
1. **Trigger:** Runs on `push` to `main` or Pull Requests.
2. **Build:** Uses Docker Buildx to build the image(s).
3. **Push:** Pushes the image to the GitHub Container Registry (`ghcr.io`) tagged with the commit SHA (`sha-<hash>`) and `latest`.

### C. The GitOps Contract (`deploy/values.yaml`)
This is the **most critical file**. It is the only infrastructure file the developer interacts with. It overrides the defaults of the underlying Generic Helm Chart.

```yaml
# deploy/values.yaml
# -- Number of application pods
replicaCount: 1

# -- Container Image (Tag is managed by ArgoCD Image Updater)
image:
  repository: ghcr.io/3-istor/my-app
  tag: "latest"

# -- Compute Resources
resources:
  requests:
    cpu: "100m"
    memory: "128Mi"
  limits:
    cpu: "500m"
    memory: "512Mi"

# -- Networking & Edge Security
ingress:
  enabled: true
  hostname: "my-app.3istor.com"
  sso_protected: false

# -- Secrets Injection
secrets:
  enabled: true
  # Injected by CMP during repo creation
  vaultPath: "kvv2/projects/my-project/my-app" 
```