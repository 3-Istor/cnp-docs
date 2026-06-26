# Deployment API Specifications

The Deployment API manages the full lifecycle of application instances, supporting initial Day-0 provisioning, real-time log streaming, and Day-2 GitOps configuration write-back.

---

## 1. Create Deployment

Starts the SAGA or Terraform provisioning flow. This is an asynchronous operation returning `202 Accepted` immediately so the UI can initiate polling.

* **HTTP Method**: `POST`
* **Path**: `/api/deployments`
* **Content-Type**: `application/json`
* **Request Headers**:
  * `Authorization: Bearer <JWT>`
* **Payload Schema (`DeploymentCreate`)**:

| Field | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `name` | String | Yes | Name of the application (lowercase, alphanumeric, hyphens). |
| `template_id` | String | Yes | ID of the template in the catalog (e.g., `k3s-gitops-app`). |
| `provider_type` | String | No | Discriminator: `kubernetes` or `legacy_hybrid` (Default: `legacy_hybrid`). |
| `project_id` | String | No | Target project boundary (Keycloak group prefix). Required for Kubernetes. |
| `app_config` | Object | Yes | Map of template-specific configuration variables. |

### Example Request (Kubernetes GitOps)
```bash
curl -X POST "https://cmp.3istor.com/api/deployments" \
  -H "Authorization: Bearer eyJhbGci..." \
  -H "Content-Type: application/json" \
  -d '{
    "name": "frontend",
    "template_id": "k3s-gitops-app",
    "provider_type": "kubernetes",
    "project_id": "alpha",
    "app_config": {
      "template_repo_name": "template-html-css",
      "app_type": "static",
      "github_owner": "3-Istor",
      "project_name": "alpha",
      "github_installation_id": "98765432"
    }
  }'
```

**Expected Response (`202 Accepted`)**:
```json
{
  "id": 42,
  "name": "frontend",
  "template_id": "k3s-gitops-app",
  "template_name": "Kubernetes App (GitOps)",
  "template_icon": "🚀",
  "template_category": "paas",
  "status": "pending",
  "step_message": "Queued...",
  "provider_type": "kubernetes",
  "project_id": "alpha",
  "github_repo_url": null,
  "argocd_app_name": null,
  "k8s_namespace": null,
  "terraform_outputs": null,
  "resource_count": 0,
  "created_at": "2026-06-26T15:49:00Z",
  "updated_at": "2026-06-26T15:49:00Z"
}
```

---

## 2. List & Filter Deployments

Retrieves all active deployments. Supports filtering by project boundary and provider type.

* **HTTP Method**: `GET`
* **Path**: `/api/deployments`
* **Query Parameters**:
  * `provider_type` (Optional): `kubernetes` | `legacy_hybrid`
  * `project_id` (Optional): String filter.
* **Response (`200 OK`)**:

```json
[
  {
    "id": 42,
    "name": "frontend",
    "template_id": "k3s-gitops-app",
    "template_name": "Kubernetes App (GitOps)",
    "template_icon": "🚀",
    "template_category": "paas",
    "status": "running",
    "step_message": "✅ Running - loadbalancer_ip: 192.168.1.215",
    "provider_type": "kubernetes",
    "project_id": "alpha",
    "github_repo_url": "https://github.com/3-Istor/frontend",
    "argocd_app_name": "alpha-frontend",
    "k8s_namespace": "alpha-frontend",
    "terraform_outputs": "{\"github_repo_url\":\"https://github.com/3-Istor/frontend\",\"k8s_namespace\":\"alpha-frontend\",\"argocd_app_name\":\"alpha-frontend\"}",
    "resource_count": 8,
    "created_at": "2026-06-26T15:49:00Z",
    "updated_at": "2026-06-26T15:52:00Z"
  }
]
```

---

## 3. Get Deployment Details

Retrieves the live record of a single deployment.

* **HTTP Method**: `GET`
* **Path**: `/api/deployments/{id}`
* **Response (`200 OK`)**: Matches object schema returned in List API.

---

## 4. Fetch Terraform Outputs

Returns only the parsed, unescaped JSON object containing the outputs declared in the Terraform module.

* **HTTP Method**: `GET`
* **Path**: `/api/deployments/{id}/outputs`
* **Response (`200 OK`)**:

```json
{
  "github_repo_url": "https://github.com/3-Istor/frontend",
  "k8s_namespace": "alpha-frontend",
  "argocd_app_name": "alpha-frontend",
  "vault_path": "kvv2/data/projects/alpha/frontend"
}
```

---

## 5. Stream Real-Time Logs (Server-Sent Events)

Streams stdout/stderr logs from the background Terraform worker in real-time using standard W3C Server-Sent Events (SSE).

* **HTTP Method**: `GET`
* **Path**: `/api/deployments/{id}/logs/stream`
* **Response Headers**:
  * `Content-Type: text/event-stream`
  * `Cache-Control: no-cache`
  * `Connection: keep-alive`

### Example Stream Output
```text
data: [TF] Initializing the backend...
data: [TF] 
data: [TF] Successfully configured the backend "s3"!
data: [TF] 
data: [TF] github_repository.app: Creating...
```

---

## 6. Fetch GitOps Configuration (Day-2)

Pulls the raw, commented `values.yaml` directly from the application's private GitHub repository using the GitHub App token.

* **HTTP Method**: `GET`
* **Path**: `/api/deployments/{id}/config`
* **Required Roles**: `project-<project_id>-members` | `project-<project_id>-admins`
* **Response (`200 OK`)**:

```json
{
  "repo": "3-Istor/frontend",
  "file_path": "deploy/values.yaml",
  "_sha": "25a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4",
  "config": {
    "project_name": "alpha",
    "app_name": "frontend",
    "replicaCount": 2,
    "image": {
      "repository": "ghcr.io/3-istor/frontend",
      "tag": "sha-8f3a1b4"
    },
    "ingress": {
      "enabled": true,
      "hostname": "frontend-alpha.3istor.com",
      "sso_protected": false
    }
  }
}
```

---

## 7. Update GitOps Configuration (Day-2 Write-Back)

Deep-merges new values into the `values.yaml` file in the GitHub repository and triggers an automated commit, preserving existing formatting and comments.

* **HTTP Method**: `PATCH`
* **Path**: `/api/deployments/{id}/config`
* **Required Roles**: `project-<project_id>-admins`
* **Request Payload**: Must include the active `_sha` of the file (to detect stale writes) and the key-value modifications.

### Example Request
```bash
curl -X PATCH "https://cmp.3istor.com/api/deployments/42/config" \
  -H "Authorization: Bearer eyJhbGci..." \
  -H "Content-Type: application/json" \
  -d '{
    "_sha": "25a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4",
    "replicaCount": 4,
    "ingress": {
      "sso_protected": true
    }
  }'
```

**Expected Response (`200 OK`)**:
```json
{
  "message": "Configuration updated successfully. ArgoCD will sync shortly.",
  "repo": "3-Istor/frontend",
  "file_path": "deploy/values.yaml",
  "commit_sha": "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0",
  "changed_keys": [
    "replicaCount",
    "ingress"
  ]
}
```

---

## 8. Delete Deployment

Triggers a non-blocking background `terraform destroy` command to wipe out all cloud resources.

* **HTTP Method**: `DELETE`
* **Path**: `/api/deployments/{id}`
* **Required Roles**: `project-<project_id>-admins` (Only admins can delete)
* **Response (`202 Accepted`)**:

```json
{
  "message": "Deletion started",
  "id": 42
}
```

---
**Next Step**: Continue to [Projects & RBAC API Contract](04-cmp-projects-api.md) (or return to the [Project Overview](../README.md)).
