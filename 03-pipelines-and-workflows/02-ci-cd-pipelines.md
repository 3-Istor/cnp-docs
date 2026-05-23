# Continuous Integration & Delivery (CI/CD)

CNP provides a seamless "code-to-cluster" experience. Developers only need to write application code and push to GitHub. The platform handles testing, building, and delivering the application to the Kubernetes cluster automatically.

## 1. The CI/CD Architecture Overview

The pipeline leverages three primary components:
1. **GitHub Actions (CI):** Handles linting, testing, and building the container images.
2. **GitHub Container Registry (GHCR):** Stores the built Docker images.
3. **ArgoCD Image Updater (CD):** Continuously monitors GHCR for new image tags and automatically updates the Kubernetes deployment manifests in Git.

---

## 2. Continuous Integration (GitHub Actions)

When the CMP provisions a new application from a "Git App Template", it includes a pre-configured `.github/workflows/ci.yml` file.

### A. The Branching Strategy
- `main` branch pushes: Trigger **Production** builds.
- `staging` branch pushes: Trigger **Staging** builds.
- Pull Requests: Trigger **Testing & Linting** (No Docker build/push).

### B. The CI Pipeline Stages
1. **Test Phase:** 
   - Spins up necessary services (e.g., PostgreSQL via `docker-compose`).
   - Runs Linters (e.g., `eslint`, `black`, `pylint`).
   - Executes Unit Tests (e.g., `pytest`, `jest`).
2. **Build Phase:** 
   - Uses Docker Buildx for optimized caching.
   - Builds the Docker image(s) for the application.
3. **Push Phase:**
   - Tags the image using the short commit SHA (e.g., `ghcr.io/my-org/my-app:sha-8f3a1b4`).
   - Pushes the image to the GitHub Container Registry.

*Security Note: The GitHub Action uses the native `${{ secrets.GITHUB_TOKEN }}` to push to GHCR. No manual Docker credentials need to be stored or managed by the developers.*

---

## 3. Continuous Delivery (ArgoCD Image Updater)

Traditionally, after a CI pipeline builds a new image, it must commit the new image tag (e.g., `sha-8f3a1b4`) back into the Git repository containing the Helm `values.yaml`. 
To avoid complex CI scripts and token management, CNP utilizes the **ArgoCD Image Updater**.

### A. How It Works
When Terraform registers the application in ArgoCD, it also deploys an `ImageUpdater` CRD alongside it (e.g., `app-updater.yaml`).

The Image Updater continuously polls the GHCR registry. When it detects a new Docker image matching the allowed pattern (e.g., `regexp:^sha-[a-f0-9]+$`), it automatically instructs ArgoCD to deploy the new image.

### B. The Write-Back Mechanism (True GitOps)
To maintain Git as the Single Source of Truth (SSOT), ArgoCD Image Updater is configured with the `write-back-method: git`. 

When a new image is deployed, ArgoCD uses the GitHub App credentials (provisioned by the CMP) to commit the new image tag back to the application's `values.yaml` file in the repository.

**ArgoCD Image Updater Configuration Example:**
```yaml
apiVersion: argocd-image-updater.argoproj.io/v1alpha1
kind: ImageUpdater
metadata:
  name: app-updater
  namespace: argocd
spec:
  applicationRefs:
    - namePattern: "<app-name>"
      images:
        - alias: "main-container"
          imageName: "ghcr.io/<org>/<app-name>"
          manifestTargets:
            helm:
              name: "image.repository"
              tag: "image.tag"
          commonUpdateSettings:
            updateStrategy: "newest-build"
            allowTags: "regexp:^sha-[a-f0-9]+$"
      writeBackConfig:
        method: "git:secret:argocd/private-repo-creds"
        gitConfig:
          branch: "main"
          writeBackTarget: "helmvalues:/deploy/values.yaml"
```

---

## 4. End-to-End Developer Flow

1. Developer commits a feature and runs `git push origin main`.
2. GitHub Actions intercepts the push, runs tests, and builds `ghcr.io/my-org/my-app:sha-a1b2c3d`.
3. ArgoCD Image Updater detects the new `sha-a1b2c3d` tag in GHCR.
4. ArgoCD pulls the image, deploys it to the K8s cluster, and monitors health.
5. ArgoCD commits `image.tag: "sha-a1b2c3d"` back to the `deploy/values.yaml` file in the developer's Git repository.
6. The CMP Dashboard (polling ArgoCD or via webhooks) updates the Application UI to show the new deployed version and commit hash.