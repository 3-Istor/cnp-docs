# Access segregation

Who may act on what on the platform, today and once the segregation plan is done. Every
identity reaches only what it needs: root credentials stay offline, people get their
rights from Keycloak groups, each automation has its own narrow identity, and a project
is a closed box across Vault, Keycloak, Argo CD and Kubernetes.

The plan runs in four phases. They are the `target` of every planned grant in the model:

| Phase | What changes |
|---|---|
| **P1** Root accounts | Vault root token revoked, Keycloak master admin and cloud roots kept offline for break-glass only |
| **P2** CMP rights | The CMP backend leaves the root token and the master admin for a scoped Vault role, a Keycloak client and scoped Cloudflare tokens |
| **P3** Tenant isolation | A project's members read its secrets and admins write them; Argo CD and kubectl rights per project; the global read roles go |
| **P4** Cloud accounts | Terraform state behind a role assumed for one hour; cloud root accounts with MFA and no key |

OpenStack stays out of scope: it hosts the cluster and keeps one admin account to deploy it.

## The CMP backend

What the plan takes away (dotted) and adds (dashed, with its phase):

![What the plan changes for the CMP](diagrams/access-cmp-diff.light.png#only-light)
![What the plan changes for the CMP](diagrams/access-cmp-diff.dark.png#only-dark)

The same platform map, lit for the CMP, today and after the plan:

![Platform lit for the CMP today](diagrams/platform-overview.cmp-current.light.png#only-light)
![Platform lit for the CMP today](diagrams/platform-overview.cmp-current.dark.png#only-dark)

![Platform lit for the CMP after the plan](diagrams/platform-overview.cmp-target.light.png#only-light)
![Platform lit for the CMP after the plan](diagrams/platform-overview.cmp-target.dark.png#only-dark)

## A tenant developer

![What the plan changes for a tenant developer](diagrams/access-dev-diff.light.png#only-light)
![What the plan changes for a tenant developer](diagrams/access-dev-diff.dark.png#only-dark)

![Platform lit for a tenant developer after the plan](diagrams/platform-overview.dev-target.light.png#only-light)
![Platform lit for a tenant developer after the plan](diagrams/platform-overview.dev-target.dark.png#only-dark)

## One project's secrets

Who reaches `project-<a>/` today, then after the plan:

![Who reaches a tenant's secrets today](diagrams/access-tenant-secrets-current.light.png#only-light)
![Who reaches a tenant's secrets today](diagrams/access-tenant-secrets-current.dark.png#only-dark)

![Who reaches a tenant's secrets after the plan](diagrams/access-tenant-secrets-target.light.png#only-light)
![Who reaches a tenant's secrets after the plan](diagrams/access-tenant-secrets-target.dark.png#only-dark)

The platform admin and the root accounts have their own views:
[`access-admin-diff`](diagrams/access-admin-diff.light.png) and
[`access-root-diff`](diagrams/access-root-diff.light.png).

## Keeping it true

The grants live in [`architecture/cnp/model.yaml`](https://github.com/3-Istor/cnp-docs/tree/main/architecture/cnp/model.yaml)
next to the rest of the platform. When a phase is done, delete its `deprecated` grants and
drop `status` and `target` from its `planned` ones: the current state catches up with the
target and the diff views empty out. Badges `G15` and `G16` mark two facts not checked
yet: the policy of the CMP's Vault token and the scope of its Cloudflare token.
