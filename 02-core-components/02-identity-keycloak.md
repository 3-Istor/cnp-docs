# Identity & Access Management: Keycloak

Keycloak acts as the Central Identity Provider (IdP) for the entire CNP ecosystem. It secures the developer portal (CMP), infrastructure tools (Vault, ArgoCD), and the deployed end-user applications via Envoy Gateway.

## 1. The Realm Configuration (`3istor`)
The core realm is configured via Terraform (`auth_flow.tf` and `main.tf`). It enforces modern security standards:
- Short-lived Access Tokens (1h).
- Strict Password Policies.
- Mandatory Multi-Factor Authentication (MFA).

## 2. The Custom Authentication Flow
CNP utilizes a highly customized browser authentication flow (`Browser-3-istor-Login-v4`). This flow wraps the standard login process with advanced security conditions:

```text
[ Browser Login Flow ]
  ├── 1. Login Wrapper
  │     ├── Cookie (SSO transparent login)
  │     ├── Kerberos / Spnego (AD Integration)
  │     └── Forms Wrapper (Fallback)
  │           ├── Username/Password
  │           └── Mandatory OTP (Forces user to setup Authenticator App)
  │
  └── 2. RBAC Deny Wrapper
        ├── Condition: User must possess the `openid_client_access` role.
        └── Action: Deny Access Authenticator (Blocks unauthorized users).
```
*Note: This architecture ensures that even if an attacker obtains valid credentials and an OTP, they cannot access the platform without being explicitly authorized by an Administrator (assigned to the `member` group).*

## 3. Project Scopes and Group Mappings
In addition to the global base groups (`infra`, `member`, `k3s-admin`, `k3s-dev`, `k3s-view`), the CNP platform dynamically generates Project-specific groups via Terraform:

- `project-<project-name>-admin`: Can manage the project inside the CMP (add/remove members, delete the project).
- `project-<project-name>-member`: Can view the project, deploy applications, and view metrics.

These group memberships are mapped into the OIDC JWT token via the `groups` client scope, which is subsequently consumed by ArgoCD, Vault, and Envoy Gateway for downstream RBAC.