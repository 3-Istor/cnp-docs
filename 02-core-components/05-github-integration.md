# GitHub Integration: Dedicated GitHub App Model

## 1. Why GitHub Apps?
To comply with the principle of least privilege, CNP rejects the use of global Personal Access Tokens (PATs). Instead, it implements a decentralized **GitHub App** integration model. 

GitHub Apps offer several security and operational benefits:
- **Granular Permissions:** Strictly limited access (e.g., read/write only to contents and administration).
- **Organization Support:** Can be installed on both personal accounts and enterprise organizations.
- **Short-lived Tokens:** Tokens generated on-the-fly expire after 1 hour, minimizing leakage risks.
- **Native ArgoCD Integration:** ArgoCD natively supports authenticating private repositories using GitHub App credentials.

---

## 2. GitHub App Permissions Matrix
The CNP GitHub App (e.g., `CNP-Portal`) requires the following scopes:

| Scope | Permission | Purpose |
|---|---|---|
| **Repository: Administration** | `Read & Write` | Creating new repositories, configuring branch protection |
| **Repository: Contents** | `Read & Write` | Bootstrapping the code template and updating `values.yaml` |
| **Repository: Metadata** | `Read-only` | Accessing repository basic information (Required by GitHub) |
| **Repository: Webhooks** | `Read & Write` | Subscribing to push events to notify CMP of developer updates |

---

## 3. The User Onboarding Flow (Linking Accounts)

```text
[ Developer ] --1. Click "Link GitHub"--> [ CMP UI ]
                                             |
                                        2. Redirect
                                             |
                                             v
[ GitHub Auth Page ] <-----------------------+
    |
    |--3. Install App on User/Org Account--> [ Select Org/User ]
    |
[ CMP Backend ] <---4. Callback with installation_id --- [ GitHub ]
    |
    +--5. Store installation_id in DB linked to Project
```

1. **Initiation:** In the CMP under Project Settings, an Admin clicks "Link GitHub Account/Organization".
2. **Redirect:** The CMP redirects the user to the GitHub App installation URL: `https://github.com/apps/cnp-portal/installations/new`.
3. **Target Selection:** The user selects which personal account or GitHub Organization they want to install the app on, and selects "All repositories" or "Only select repositories".
4. **Callback:** After installation, GitHub redirects the user back to the CMP with an `installation_id` in the query parameters.
5. **Storage:** The CMP Backend stores this `installation_id` in the Project database. This ID will be used for all future operations within this Project.

---

## 4. Token Generation Workflow (As-Needed)
When the CMP Backend needs to interact with GitHub (e.g., to create a repo during provisioning):

1. The CMP Backend reads the `installation_id` linked to the Project.
2. It generates a JSON Web Token (JWT) signed with the **CNP-Portal Private Key** (stored securely in Vault).
3. It sends a request to GitHub's API: `POST /app/installations/{installation_id}/access_tokens` using the JWT.
4. GitHub returns a **short-lived installation access token** (valid for 60 minutes).
5. The CMP Backend uses this token to authenticate its API calls or passes it to Terraform to bootstrap the repository.

---

## 5. ArgoCD Private Repository Access
To allow ArgoCD to sync the newly created private repository, we configure ArgoCD with the GitHub App credentials directly.

```yaml
# Configured via Terraform during Project Bootstrapping
apiVersion: v1
kind: Secret
metadata:
  name: private-repo-creds
  namespace: argocd
  labels:
    argocd.argoproj.io/secret-type: repository
stringData:
  type: git
  url: https://github.com/my-user-org/my-new-app.git
  githubAppID: "123456"               # CNP GitHub App ID
  githubAppInstallationID: "98765432" # Captured during onboarding
  githubAppPrivateKey: |              # SECURELY fetched from Vault
    -----BEGIN RSA PRIVATE KEY-----
    ...
    -----END RSA PRIVATE KEY-----
```
*Note: Using this approach, ArgoCD automatically handles generating and rotating its own short-lived tokens to pull the code. No static credentials ever touch the cluster.*