# Terraform App Provisioner & State Management

The Terraform App Provisioner is the automation engine utilized by the CMP Backend during the **Day-0 (Bootstrapping)** phase of an application. 

Its role is to interact with external APIs (GitHub, Keycloak, Vault, Kubernetes) to prepare the environment before ArgoCD takes over.

---

## 1. The Micro-State Pattern (S3 State Isolation)

To prevent risk, latency, and state-locking issues associated with a single monolithic state file, CNP implements a **Micro-State Pattern**. 

Every application deployment managed by the CMP receives its own isolated `.tfstate` file stored in S3.

```text
3-istor-tf-infra-aws (S3 Bucket)
├── infra/
│   └── k3s-master/
│       └── terraform.tfstate      # Static Base Infra State (Managed manually)
└── cmp/
    └── projects/
        └── project-alpha/
            ├── app-frontend.tfstate # Dynamic App State (Managed by CMP)
            └── app-backend.tfstate  # Dynamic App State (Managed by CMP)
```

### Dynamic Backend Initialization
When the CMP Backend triggers a deployment, it does not use a hardcoded `key` in the backend configuration. Instead, it initializes Terraform dynamically using `-backend-config` CLI flags.

**The Initialization Command executed by the CMP Backend:**
```bash
terraform init \
  -backend-config="bucket=3-istor-tf-infra-aws" \
  -backend-config="key=cmp/projects/${PROJECT_NAME}/${APP_NAME}.tfstate" \
  -backend-config="region=eu-west-3" \
  -backend-config="encrypt=true" \
  -backend-config="dynamodb_table=terraform-state-lock" \
  -reconfigure
```
*Note: This ensures that the state files are completely isolated. Deploying `app-frontend` will never lock or interfere with `app-backend` or the base infrastructure.*

---

## 2. Provisioner Responsibilities (Day-0)

When `terraform apply` is executed with the dynamic state, the provisioner executes the following resources sequentially:

### Step 1: GitHub Repository Creation
- Creates a private repository under the user's linked organization or personal account using the temporary installation token from the CNP GitHub App.
- Pushes the initial boilerplate code from the template.
- Configures branch protection rules on `main` (disallowing direct force pushes).

### Step 2: Keycloak Client & Group Mapping
- Generates Keycloak groups if they do not exist: `project-<project-name>-members`.
- Maps these groups to the authentication scopes for downstream SSO validation.

### Step 3: Vault Engines & Paths Configuration
- Creates the Vault KV path: `kvv2/projects/<project-name>/<app-name>/`.
- Populates the path with auto-generated sensitive values (e.g., random database passwords, JWT signing keys).
- Creates the Vault Kubernetes Auth Role bound to the application's K8s ServiceAccount.

### Step 4: ArgoCD Application Registration
- Provisions a Kubernetes Secret in the `argocd` namespace containing the GitHub App credentials so ArgoCD can fetch the private repository.
- Deploys the ArgoCD `Application` Custom Resource pointing to the newly created GitHub repository.

---

## 3. Day-2 Deletion & Soft-Delete Strategy

When a developer requests the deletion of an application:
1. The CMP Backend executes `terraform destroy` pointing to the application's micro-state.
2. Terraform destroys the K8s namespace, Keycloak clients, and ArgoCD application mappings.
3. **Repository Preservation:** To prevent accidental loss of code history, the GitHub repository is **archived (read-only)** rather than permanently deleted, unless explicitly overridden by an administrator.
4. The micro-state file in S3 is removed or moved to an archive path.