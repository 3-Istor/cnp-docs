# FastAPI OpenAPI Optimization & MCP Server Integration

This guide provides a comprehensive tutorial on optimizing your FastAPI OpenAPI schema (Swagger UI) for OIDC Keycloak security and building a dedicated **Model Context Protocol (MCP)** server. The MCP server acts as an AI-native bridge, allowing AI coding assistants (such as Cursor or Claude Desktop) to dynamically query platform documentation, check container status, and execute deployment workflows.

---

## 1. Enhancing FastAPI OpenAPI (Swagger UI)

FastAPI automatically generates an OpenAPI schema. To make the interactive `/docs` UI functional in a secured production environment, you must register your Keycloak realm as an OAuth2 security scheme.

### A. Implementing OAuth2 Authorization Code Flow (`app/main.py`)

Modify your FastAPI initialization code to register Keycloak OIDC inside the Swagger UI:

```python
# app/main.py extract
from fastapi import FastAPI
from fastapi.openapi.utils import get_openapi
from fastapi.security import OAuth2AuthorizationCodeBearer

# Define Keycloak OAuth2 Flow
oauth2_scheme = OAuth2AuthorizationCodeBearer(
    authorizationUrl="https://auth.3istor.com/realms/3istor/protocol/openid-connect/auth",
    tokenUrl="https://auth.3istor.com/realms/3istor/protocol/openid-connect/token",
    scopes={
        "openid": "Required for authentication",
        "profile": "Access user profile metadata",
        "groups": "Project boundary mappings"
    }
)

app = FastAPI(title="CNP Portal API")

# Custom OpenAPI Schema Generator
def custom_openapi():
    if app.openapi_schema:
        return app.openapi_schema

    openapi_schema = get_openapi(
        title="Cloud Native Platform API",
        version="1.0.0",
        description="Core API for provisioning and managing GitOps applications.",
        routes=app.routes,
    )

    # Inject OIDC Security Scheme
    openapi_schema["components"]["securitySchemes"] = {
        "KeycloakOAuth2": {
            "type": "oauth2",
            "flows": {
                "authorizationCode": {
                    "authorizationUrl": "https://auth.3istor.com/realms/3istor/protocol/openid-connect/auth",
                    "tokenUrl": "https://auth.3istor.com/realms/3istor/protocol/openid-connect/token",
                    "scopes": {
                        "openid": "Required for authentication",
                        "profile": "Access user profile metadata",
                        "groups": "Project boundary mappings"
                    }
                }
            }
        }
    }

    # Enforce global security on all non-public endpoints
    for path in openapi_schema["paths"].values():
        for method in path.values():
            if "security" not in method:
                method["security"] = [{"KeycloakOAuth2": []}]

    app.openapi_schema = openapi_schema
    return app.openapi_schema

app.openapi = custom_openapi
```

Now, developers visiting `https://cmp.3istor.com/docs` can click **Authorize**, complete the Keycloak SSO flow, and test API endpoints directly from their browsers.

---

## 2. Building the CNP MCP Server (Python)

The **Model Context Protocol (MCP)** allows local AI clients to consume resources (data/docs) and tools (API execution). We implement this using the official Python `mcp` SDK.

### A. Installing the Dependencies

Inside your `backend/` directory, add the required package:

```bash
poetry add mcp
```

### B. Creating the MCP Server (`backend/app/mcp_server.py`)

Create a new file to act as your AI-native gateway. It exposes **Resources** (read-only architectural documentation) and **Tools** (actions to orchestrate deployments):

```python
# app/mcp_server.py
from pathlib import Path
import httpx
from mcp.server.fastmcp import FastMCP

# Initialize FastMCP Server
mcp = FastMCP("CNP Portal")

DOCS_DIR = Path(__file__).parent.parent.parent / ".kiro" / "steering" / "docs"
CMP_API_URL = "http://localhost:8000/api"

# ═══════════════════════════════════════════════════════════════════════════
# 1. MCP RESOURCES (Allows AI to read project documentation)
# ═══════════════════════════════════════════════════════════════════════════

@mcp.resource("docs://{category}/{filename}")
def get_documentation(category: str, filename: str) -> str:
    """Retrieve specific architectural documentation files.

    Example topics: '01-architecture/02-tenancy-and-isolation.md'
    """
    file_path = DOCS_DIR / category / f"{filename}.md"

    if not file_path.exists():
        return f"Error: Documentation file '{category}/{filename}' not found."

    with open(file_path, "r", encoding="utf-8") as f:
        return f.read()

# ═══════════════════════════════════════════════════════════════════════════
# 2. MCP TOOLS (Allows AI to trigger actions and deployments)
# ═══════════════════════════════════════════════════════════════════════════

@mcp.tool()
async def list_active_deployments(token: str) -> str:
    """List all current application deployments registered in the portal."""
    headers = {"Authorization": f"Bearer {token}"}

    async with httpx.AsyncClient() as client:
        response = await client.get(f"{CMP_API_URL}/deployments", headers=headers)

        if response.status_code != 200:
            return f"Error: Failed to fetch deployments (HTTP {response.status_code})"

        return response.text

@mcp.tool()
async def deploy_new_app(
    token: str,
    name: str,
    project_name: str,
    template_id: str = "k3s-gitops-app",
    github_installation_id: str = "98765432"
) -> str:
    """Triggers Day-0 provisioning of a new Kubernetes GitOps application."""
    headers = {
        "Authorization": f"Bearer {token}",
        "Content-Type": "application/json"
    }

    payload = {
        "name": name,
        "template_id": template_id,
        "provider_type": "kubernetes",
        "project_id": project_name,
        "app_config": {
            "template_repo_name": "template-html-css",
            "app_type": "static",
            "github_owner": "3-Istor",
            "project_name": project_name,
            "github_installation_id": github_installation_id
        }
    }

    async with httpx.AsyncClient() as client:
        response = await client.post(
            f"{CMP_API_URL}/deployments",
            headers=headers,
            json=payload
        )
        return response.text

if __name__ == "__main__":
    # Runs the MCP server over standard input/output (stdio) for local IDE link
    mcp.run()
```

---

## 3. Registering the MCP Server in Your IDE (Cursor / Claude)

To make your local AI agent aware of your documentation and provisioning tools, register the MCP server in your local client configurations.

### A. Cursor IDE Setup

1. Open Cursor and go to **Settings** → **Features** → **MCP**.
2. Click **+ New MCP Server**.
3. Fill out the fields:
   - **Name**: `CNP Portal`
   - **Type**: `command`
   - **Command**: `poetry run python app/mcp_server.py` (Make sure your cwd matches `backend/`)

### B. Claude Desktop Setup

Add the configuration to your local config file:

- MacOS: `~/Library/Application Support/Claude/claude_desktop_config.json`
- Windows: `%APPDATA%\Claude\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "cnp-portal": {
      "command": "poetry",
      "args": [
        "run",
        "python",
        "/absolute/path/to/arcl-cmp/backend/app/mcp_server.py"
      ],
      "env": {
        "KEYCLOAK_URL": "https://auth.3istor.com",
        "VAULT_URL": "https://vault.3istor.com"
      }
    }
  }
}
```

---

## 🦾 Testing the AI Integration Loop

Once registered, you can ask your AI Agent (such as Cursor's Composer or Claude):

> _"Read the documentation on tenancy and explain how Vault Secrets are isolated in our cluster. If it looks correct, deploy a new static HTML app named 'billing-web' inside the project 'test-alpha'."_

The AI agent will sequentially:

1. **Call Resource**: `docs://01-architecture/02-tenancy-and-isolation` to read your dynamic Vault role configurations.
2. **Explain**: Synthesize the isolation boundaries.
3. **Call Tool**: `deploy_new_app` using your active OIDC token to trigger the backend provisioning SAGA, compiling the infrastructure in real-time.

---

_Back to [CMP Backend API Specs](01-cmp-deployment-api.md) or return to the [Project Overview](../README.md)._
