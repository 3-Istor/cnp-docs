# CMP MCP integration (CNP Agent Kit)

AI assistants use CNP through a remote MCP server served by the CMP.

```mermaid
sequenceDiagram
    participant A as Assistant (Claude, ChatGPT, Codex...)
    participant C as CMP (OAuth server + /mcp)
    participant K as Keycloak
    A->>C: POST /mcp (no token)
    C-->>A: 401, resource_metadata
    A->>C: discovery, POST /register (redirect URI allowlist)
    A->>C: /authorize (PKCE), in the user's browser
    C->>K: usual login (OIDC)
    C-->>A: consent page in the CMP, then code
    A->>C: POST /token
    C-->>A: CMP token (user's Keycloak sub, 15 min, refresh 30 days)
    A->>C: tools/call with Bearer
    C->>C: stake policy, then REST API in-process (RBAC + audit)
```

## Design choices

| Choice | Why |
|---|---|
| The CMP is the OAuth server of its MCP; Keycloak stays the identity provider behind the login | Assistants only talk to the platform; no Keycloak client per assistant, no Keycloak upgrade (CIMD) |
| Opaque tokens, stored hashed, tied to a grant | Revoking a grant in `/account` cuts the assistant at its next call |
| Tools call the CMP REST API in-process with the assistant's token | RBAC, validation and the audit journal apply unchanged; the journal records the assistant (`via`) |
| MCP mounted in the CMP backend, stateless Streamable HTTP | No extra pod on a memory-tight cluster |
| Stake policy (`backend/app/services/agents/policy.yaml`) | Guardrails proportional to what is at stake: allow, ask the user, pull request, approval in the CMP, refused |
| Guides served as `skill://` resources and by `get_guide`, from `3-Istor/cnp-agent-kit` | One source for the Claude Code plugin and for assistants without plugins (progressive disclosure) |

## Stake levels

| Action | sandbox | demo | standard | critical |
|---|---|---|---|---|
| Read | allow | allow | allow | allow |
| Create an app | allow | allow | allow | approval in the CMP |
| Push application code | allow | allow | allow | pull request |
| Change CI, Dockerfile, `deploy/` | refused | ask | pull request | pull request |
| Release, promote, retry, skip | allow | allow | ask | approval in the CMP |
| Abort a rollout | allow | allow | allow | ask |
| Delete an app | ask | ask | CMP only | CMP only |

Abort and retry are always allowed while an alert fires on the app. A project
owner can open an autonomous window of 1 to 8 hours on a standard project.

## Code

| What | Where |
|---|---|
| Tools, resources, prompts | `CMP/backend/app/mcp_server.py` |
| OAuth provider and token store | `CMP/backend/app/services/agents/provider.py`, `store.py` |
| Policy | `CMP/backend/app/services/agents/policy.yaml`, `policy.py` |
| Consent, approval, settings API | `CMP/backend/app/routers/agents.py` |
| Pages | `CMP/frontend/src/app/connect/`, `connect-ai/`, `components/agents/` |
| Route | `K3s/k8s/config/arcl-cmp/mcp-route.yaml` |
| Guides, plugin, doctrine | `3-Istor/cnp-agent-kit` |
