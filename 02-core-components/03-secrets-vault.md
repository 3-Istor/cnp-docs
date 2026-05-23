# Secrets Management: HashiCorp Vault

CNP utilizes HashiCorp Vault for centralized secret storage, encryption, and dynamic injection. No secrets are ever stored in plain text in GitHub, ArgoCD, or the CMP database.

## 1. Vault Architecture Overview
- **Engine:** KV Version 2 (`kvv2/`).
- **Operators:** `ricoberger` Vault Secrets Operator (VSO) running in the `vault-secrets-operator` namespace.
- **Authentication:** 
  - **Machine-to-Machine:** Kubernetes Service Account JWT Auth.
  - **Human-to-Machine:** OIDC Auth bound to Keycloak.

## 2. Developer Experience (Human Access)
Developers access the Vault UI via OIDC (`Vault` Client in Keycloak). 
When a new Project is provisioned by the CMP, Terraform creates a Vault Policy restricting access to that project's specific path, and binds this policy to the Keycloak Project Group.

**Generated Policy Example:**
```hcl
# project-alpha-policy
path "kvv2/data/projects/alpha/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}
path "kvv2/metadata/projects/alpha/*" {
  capabilities = ["list", "read"]
}
```

## 3. Application Experience (Machine Access)
End-user applications consume secrets via the Vault Secrets Operator (VSO). 
To maintain strict isolation, Terraform automatically provisions a Vault Role for every new Application.

**Workflow:**
1. Terraform creates `vault_kubernetes_auth_backend_role` named `<project-name>-<app-name>-role`.
2. This role is bound strictly to the `vault-secrets-operator` ServiceAccount, but enforcing the `<project-name>-<app-name>` namespace context.
3. ArgoCD deploys a `VaultSecret` Custom Resource (CR) alongside the application:
```yaml
apiVersion: ricoberger.de/v1alpha1
kind: VaultSecret
metadata:
  name: app-database-credentials
  namespace: project-alpha-app-front
spec:
  path: kvv2/projects/alpha/app-front/db
  vaultRole: project-alpha-app-front-role
  type: Opaque
  keys:
    - password
```
4. The VSO authenticates with Vault using the specific role, retrieves the secret, and creates a native Kubernetes `Secret`.
5. The application Pod mounts the native `Secret` as an environment variable or volume.