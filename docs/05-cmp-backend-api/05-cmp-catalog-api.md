# Catalog API Contract

The Catalog API provides access to the available application templates (deployable blueprints) dynamically read and compiled from the centralized Git repository.

---

## 1. List Blueprints (Templates)

Retrieves all active blueprints currently synced to the platform.

* **HTTP Method**: `GET`
* **Path**: `/api/catalog`
* **Response (`200 OK`)**:

```json
[
  {
    "id": "k3s-gitops-app",
    "name": "Kubernetes App (GitOps)",
    "description": "Deploys a containerized application on K3s via ArgoCD.",
    "icon": "🚀",
    "category": "paas",
    "fields": [
      {
        "name": "template_repo_name",
        "label": "Stack Type (Git Template)",
        "type": "select",
        "default": "template-html-css",
        "options": [
          "template-html-css",
          "template-app-webapp-python-fastapi-react"
        ],
        "required": true
      },
      {
        "name": "github_owner",
        "label": "GitHub Organization/User",
        "type": "text",
        "default": "3-Istor",
        "options": null,
        "required": true
      }
    ],
    "image_path": "/static/templates/k3s-gitops-app/icon.png",
    "enabled": true
  }
]
```

---

## 2. Get Blueprint Details

Retrieves schema details for a specific template.

* **HTTP Method**: `GET`
* **Path**: `/api/catalog/{template_id}`
* **Response (`200 OK`)**:

```json
{
  "id": "k3s-gitops-app",
  "name": "Kubernetes App (GitOps)",
  "description": "Deploys a containerized application on K3s via ArgoCD.",
  "icon": "🚀",
  "category": "paas",
  "fields": [
    {
      "name": "template_repo_name",
      "label": "Stack Type (Git Template)",
      "type": "select",
      "default": "template-html-css",
      "options": [
        "template-html-css",
        "template-app-webapp-python-fastapi-react"
      ],
      "required": true
    }
  ],
  "image_path": "/static/templates/k3s-gitops-app/icon.png",
  "enabled": true
}
```

---

## 3. Force Sync Catalog

Instructs the CMP backend to fetch the latest template definitions from the remote `app-templates` Git repository.

* **HTTP Method**: `POST`
* **Path**: `/api/catalog/sync`
* **Response (`200 OK`)**:

```json
{
  "message": "Template repository synced successfully"
}
```

### Example cURL Request:
```bash
curl -X POST "https://cmp.3istor.com/api/catalog/sync" \
  -H "Authorization: Bearer eyJhbGci..."
```

---

## ⚙️ How Template Compilation Works Under the Hood

The catalog is entirely dynamic. The backend does not hardcode template schemas:

1. **Git Cloning**: On startup (and when `POST /sync` is triggered), `template_repository.py` pulls from the remote Git repository into `data/templates/`.
2. **Directory Parsing**: The backend scans every folder inside `data/templates/templates/` for a `manifest.json` file.
3. **Pydantic Compilation**: The `catalog_service.py` parses each valid `manifest.json` into a structured `CatalogTemplate` object, exposing variable schemas and options directly to the Next.js form generator.

```text
data/templates/templates/
├── k3s-gitops-app/
│   ├── manifest.json   ◄── Read & compiled into Pydantic schema
│   ├── main.tf
│   └── outputs.tf
```

---
**Next Step**: Continue to [Infrastructure & Health API Contract](06-cmp-infra-api.md) (or return to the [Project Overview](../index.md)).
