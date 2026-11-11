# Access segregation

Who may act on what on the platform, today and once the segregation plan is done. Every
identity reaches only what it needs: root credentials stay offline, people get their
rights from Keycloak groups, each automation has its own narrow identity, and a project
is a closed box across Vault, Keycloak, Argo CD and Kubernetes.

The plan runs in four phases, shown on every change in the matrix:

| Phase | What changes |
|---|---|
| **P1** Root accounts, audit trail | Vault root token revoked; Keycloak master admin, a second GitHub org owner, Cloudflare super admin and cloud roots kept offline for break-glass only; platform admins use kubectl over OIDC instead of a committed kubeconfig. Vault, K3s, Keycloak and Argo CD audit logs shipped to an append-only store that only break-glass can delete |
| **P2** CMP rights | The CMP backend leaves the root token and the master admin for a scoped Vault role, a Keycloak client and scoped Cloudflare tokens |
| **P3** Tenant isolation, GitOps | A project's members read its secrets and admins write them; kubectl rights per project; the global read roles go. Platform admins reach tenant secrets only through just-in-time access brokered by the CMP portal: requested there, approved by a second platform admin, expiring on its own, audited. The GitOps repos take changes through reviewed pull requests only |
| **P4** Cloud accounts | AWS organization with management, log-archive and workloads accounts; Terraform state behind a role assumed for one hour through Keycloak federation (`AssumeRoleWithWebIdentity`, like `cnp-finops-reader`) |

Where it stands on 2026-11-11:

| Phase | Done | Still to do |
|---|---|---|
| **P1** | Vault root token revoked, generated again from the unseal keys for break-glass; kubectl over OIDC (`k3s-admin`), the committed kubeconfig is gone; Vault, K3s, Keycloak and Argo CD audit logs in Loki with an Object Lock archive | Keycloak realm roles for `infra` and the master admin kept offline; a second GitHub owner kept offline (one owner today, `G23`); Cloudflare super admin |
| **P2** | The CMP logs into Vault as Kubernetes auth role `cmp-backend` (1-hour token) | Removing the static `VAULT_TOKEN`, kept in the CMP values as a rollback path until a create and delete test passes (`G15`); Keycloak client `cmp-provisioner` in place of the master admin; scoped Cloudflare tokens |
| **P3** | Members read and admins write their project's secrets; kubectl per project; no global Argo CD read for members | Just-in-time access and `vault_data_admin`; GitHub ruleset on `main`; `k3s-dev` and `k3s-view` removed; the auditor's Vault, Keycloak, kubectl and audit log reads |
| **P4** | Nothing | AWS organization (none yet: one account); Terraform state role |

The CMP portal follows the same identities: a Keycloak group gives the platform and the
portal rights alike, and the portal reads the token's `groups` claim. Its own changes run
in five lots ([CNP-228](https://3istors.atlassian.net/browse/CNP-228)):

| Lot | What changes |
|---|---|
| **L1** Authorization holes | The API refuses tokens issued to another realm client; catalog resync and the list of closed months leave the signed-in users |
| **L2** Keycloak groups | `security-auditor` and `finops` groups, demo accounts per role, a dedicated `cnp-portal` client |
| **L3** Roles in the portal | Platform admin from the `infra` group instead of a username list; auditor reads every project and approves critical CVE exceptions; FinOps reads every project's costs, exports and closes the month |
| **L4** Grafana and Argo CD | The Grafana platform org stops taking every new user as Viewer; the auditor reads Argo CD |
| **L5** Cleanup | `FINOPS_ADMIN_USERS` and `FINOPS_FINANCE_USERS` leave the CMP values |

| Identity | Keycloak group | In the portal |
|---|---|---|
| Platform admin | `infra` | Every project, platform views, security policy, catalog |
| Auditor | `security-auditor` | Reads every project except logs and secrets; approves critical CVE exceptions |
| FinOps | `finops` | Every project's costs, exports, month close |
| Project owner | the project's creator, in its admins | Admin rights plus the budget |
| Project admin | `<project>-admins` | Deploys, configures, manages members, approves CVE exceptions below critical |
| Project developer | `<project>-members` | Operates apps, reads logs, requests CVE exceptions |
| Project guest | `<project>-guests` | Sees apps, releases and costs |

OpenStack stays out of scope: it hosts the cluster and keeps one admin account to deploy it.

## Target

![CNP access matrix, target](diagrams/access-matrix-target.light.png#only-light)
![CNP access matrix, target](diagrams/access-matrix-target.dark.png#only-dark)

The end state only: who can touch what once every phase and lot is done.

## What changes

![CNP access matrix](diagrams/access-matrix.light.png#only-light)
![CNP access matrix](diagrams/access-matrix.dark.png#only-dark)

Read a row for everything an identity can do, a column for who can touch a resource. The
fill is the level after the plan; the frame and its tag show what the plan changes and in
which phase or lot. A workload column touched by another project's identity is a leak.

## What the plan does not remove

- **GitOps is admin.** Argo CD's controller can apply anything anywhere, so whoever can merge into the GitOps repos can deploy anywhere. The plan guards those repos (reviewed pull requests, CODEOWNERS); it does not shrink the controller.
- **The CMP writes `project-*` policies**, so it could grant itself any project's secrets. This is inherent to a provisioning tool: an accepted risk, to watch in the Vault audit log.

## Keeping it true

The memberships and grants live in [`architecture/cnp/model.yaml`](https://github.com/3-Istor/cnp-docs/tree/main/architecture/cnp/model.yaml)
next to the rest of the platform: `deprecated` marks what the plan removes, `planned` what
it adds, and `target` the phase. When a phase is done, delete its `deprecated` entries and
drop `status` and `target` from its `planned` ones; the matrix then shows the new state with
fewer changes, until none is left.

Facts not verified yet are badged in the model: the policy of the CMP's static Vault
token (`G15`) and the scope of its Cloudflare token (`G16`); who holds Cloudflare super
admin is not checked either and is drawn as the platform admin. Checked on 2026-11-11:
the GitHub org gives every member write on every repo, does not require 2FA, has no
ruleset and protects `main` on CMP only (`G17`), and has a single owner (`G23`).
