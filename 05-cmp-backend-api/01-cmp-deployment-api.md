# Deployment API Specification

## Overview

The Deployment API manages application deployments across multiple infrastructure providers. As of Phase 3, it supports both legacy hybrid (OpenStack + AWS) and modern Kubernetes (GitOps) deployments.

---

## Data Model

### Deployment Object

```json
{
  "id": 1,
  "name": "my-app",
  "template_id": "kubernetes-fastapi",
  "status": "running",
  "step_message": "✅ Running - ArgoCD syncing from https://github.com/...",

  // Provider discriminator (NEW in Phase 3)
  "provider_type": "kubernetes", // or "legacy_hybrid"

  // Multi-tenancy (NEW in Phase 3)
  "project_id": "project-alpha",

  // Kubernetes-specific fields (NEW in Phase 3)
  "github_repo_url": "https://github.com/user/my-app",
  "argocd_app_name": "project-alpha-my-app",
  "k8s_namespace": "project-alpha-my-app",

  // Common fields
  "terraform_outputs": "{...}", // JSON string
  "terraform_state_path": "cmp/projects/project-alpha/my-app.tfstate",
  "resource_count": 5,
  "app_config": "{...}", // JSON string

  // Template metadata
  "template_name": "FastAPI + React",
  "template_icon": "🚀",
  "template_category": "paas",

  // Timestamps
  "created_at": "2026-05-24T10:00:00Z",
  "updated_at": "2026-05-24T10:05:00Z"
}
```

### Provider Types

- **`legacy_hybrid`**: OpenStack VMs + AWS Auto Scaling Groups (legacy SAGA pattern)
- **`kubernetes`**: GitHub + Terraform + ArgoCD (modern GitOps pattern)

### Deployment Status

- `pending`: Initial state, waiting to start
- `initializing`: Setting up prerequisites
- `planning`: Terraform planning phase
- `deploying`: Active deployment in progress
- `running`: Successfully deployed and operational
- `degraded`: Deployed but experiencing issues
- `failed`: Deployment failed
- `deleting`: Deletion in progress
- `deleted`: Successfully deleted

---

## API Endpoints

### Create Deployment

**POST** `/api/deployments`

Creates a new deployment. The behavior depends on the `provider_type`.

**Request Body (Kubernetes)**:

```json
{
  "name": "my-app",
  "template_id": "kubernetes-fastapi",
  "provider_type": "kubernetes",
  "app_config": {
    "project_name": "project-alpha",
    "github_installation_id": "12345678",
    "replica_count": 2,
    "sso_protected": false
  }
}
```

**Request Body (Legacy Hybrid)**:

```json
{
  "name": "my-app",
  "template_id": "hybrid-web-db",
  "provider_type": "legacy_hybrid",
  "app_config": {
    "instance_type": "t3.micro",
    "db_size": "small"
  }
}
```

**Response**: `201 Created`

```json
{
  "id": 1,
  "name": "my-app",
  "status": "pending",
  "provider_type": "kubernetes",
  "created_at": "2026-05-24T10:00:00Z"
}
```

---

### Get Deployment

**GET** `/api/deployments/{id}`

Retrieves a single deployment by ID.

**Response**: `200 OK`

```json
{
  "id": 1,
  "name": "my-app",
  "status": "running",
  "provider_type": "kubernetes",
  "github_repo_url": "https://github.com/user/my-app",
  "argocd_app_name": "project-alpha-my-app",
  "k8s_namespace": "project-alpha-my-app",
  ...
}
```

---

### List Deployments

**GET** `/api/deployments`

Lists all deployments. Supports filtering by provider type.

**Query Parameters**:

- `provider_type` (optional): Filter by `kubernetes` or `legacy_hybrid`
- `project_id` (optional): Filter by project (Kubernetes only)
- `status` (optional): Filter by status

**Response**: `200 OK`

```json
{
  "deployments": [
    {
      "id": 1,
      "name": "my-app",
      "provider_type": "kubernetes",
      "status": "running",
      ...
    }
  ],
  "total": 1
}
```

---

### Delete Deployment

**DELETE** `/api/deployments/{id}`

Deletes a deployment and all associated resources.

For Kubernetes deployments:

- Executes `terraform destroy` on the micro-state
- Removes GitHub repository (archives it by default)
- Removes Kubernetes namespace
- Removes Vault secrets
- Removes ArgoCD Application

**Response**: `202 Accepted`

```json
{
  "message": "Deletion started",
  "deployment_id": 1
}
```

---

## Frontend Integration Guide

### Detecting Provider Type

When rendering deployment cards, check the `provider_type` field:

```typescript
if (deployment.provider_type === "kubernetes") {
  // Show Kubernetes-specific UI
  // - GitHub repo link
  // - ArgoCD app link
  // - Namespace info
} else {
  // Show legacy hybrid UI
  // - AWS ALB URL
  // - OpenStack VM IPs
}
```

### GitHub Installation ID

For Kubernetes deployments, the frontend must provide the `github_installation_id` in the `app_config`. This is obtained from:

1. User links their GitHub account (OAuth flow)
2. GitHub redirects back with `installation_id`
3. Store in Keycloak user profile
4. Fetch when creating deployment

**Example Flow**:

```typescript
// 1. User clicks "Link GitHub"
window.location.href = "https://github.com/apps/cnp-portal/installations/new";

// 2. GitHub redirects back
// URL: /api/github/callback?installation_id=12345678

// 3. Store in user profile
await updateKeycloakProfile({ github_installation_id: "12345678" });

// 4. Use when creating deployment
const deployment = await createDeployment({
  name: "my-app",
  provider_type: "kubernetes",
  app_config: {
    github_installation_id: user.github_installation_id,
    ...
  }
});
```

### Deployment Status Polling

Poll the deployment status every 3-5 seconds while `status` is `deploying`:

```typescript
const pollStatus = async (deploymentId: number) => {
  const interval = setInterval(async () => {
    const deployment = await fetchDeployment(deploymentId);

    if (deployment.status === "running" || deployment.status === "failed") {
      clearInterval(interval);
    }

    updateUI(deployment);
  }, 3000);
};
```

### Action Buttons by Provider Type

**Kubernetes Deployments**:

- 🐙 View in ArgoCD
- 📦 Open GitHub Repo
- 🔒 Manage Secrets (Vault)
- 📊 View Logs (K8s)

**Legacy Hybrid Deployments**:

- 🌐 Open Application (ALB URL)
- ☁️ View AWS Console
- 🖥️ View OpenStack Dashboard

---

## Error Handling

### Common Error Responses

**400 Bad Request**:

```json
{
  "error": "Invalid provider_type",
  "message": "provider_type must be 'kubernetes' or 'legacy_hybrid'"
}
```

**404 Not Found**:

```json
{
  "error": "Deployment not found",
  "deployment_id": 999
}
```

**500 Internal Server Error**:

```json
{
  "error": "Deployment failed",
  "message": "GitHub App authentication failed: Invalid installation ID",
  "deployment_id": 1
}
```

### Deployment Failure States

When `status` is `failed`, check `step_message` for details:

```json
{
  "status": "failed",
  "step_message": "❌ GitHub App authentication failed: Invalid installation ID"
}
```

Common failure reasons:

- **Kubernetes**: Invalid GitHub installation ID, Terraform errors, ArgoCD sync failures
- **Legacy Hybrid**: OpenStack quota exceeded, AWS capacity issues, network errors

---

## Migration Notes (Phase 3)

### Breaking Changes

**None**. The API is fully backward compatible.

### New Fields

All new fields are optional and only populated for Kubernetes deployments:

- `provider_type` (defaults to `legacy_hybrid`)
- `project_id`
- `github_repo_url`
- `argocd_app_name`
- `k8s_namespace`

### Frontend Changes Required

1. **Catalog View**: Separate IaaS and PaaS templates
2. **Account Page**: Add "Link GitHub Account" button
3. **Deployment Cards**: Conditional rendering based on `provider_type`
4. **Create Form**: Add project selection for Kubernetes deployments

---

## Examples

### Create Kubernetes Deployment (Full Example)

```bash
curl -X POST http://localhost:8000/api/deployments \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "my-fastapi-app",
    "template_id": "kubernetes-fastapi",
    "provider_type": "kubernetes",
    "app_config": {
      "project_name": "project-alpha",
      "github_installation_id": "12345678",
      "replica_count": 3,
      "sso_protected": true
    }
  }'
```

### Get Deployment with Kubernetes Fields

```bash
curl http://localhost:8000/api/deployments/1 \
  -H "Authorization: Bearer $TOKEN"
```

**Response**:

```json
{
  "id": 1,
  "name": "my-fastapi-app",
  "template_id": "kubernetes-fastapi",
  "status": "running",
  "step_message": "✅ Running - ArgoCD syncing from https://github.com/user/my-fastapi-app",
  "provider_type": "kubernetes",
  "project_id": "project-alpha",
  "github_repo_url": "https://github.com/user/my-fastapi-app",
  "argocd_app_name": "project-alpha-my-fastapi-app",
  "k8s_namespace": "project-alpha-my-fastapi-app",
  "terraform_state_path": "cmp/projects/project-alpha/my-fastapi-app.tfstate",
  "created_at": "2026-05-24T10:00:00Z",
  "updated_at": "2026-05-24T10:05:00Z"
}
```
