# Access segregation

Who may act on what on the platform, today and once the segregation plan is done. Every
identity reaches only what it needs: root credentials stay offline, people get their
rights from Keycloak groups, each automation has its own narrow identity, and a project
is a closed box across Vault, Keycloak, Argo CD and Kubernetes.

The plan runs in four phases, shown on every change in the matrix:

| Phase | What changes |
|---|---|
| **P1** Root accounts | Vault root token revoked, Keycloak master admin and cloud roots kept offline for break-glass only; platform admins use kubectl over OIDC instead of a committed kubeconfig |
| **P2** CMP rights | The CMP backend leaves the root token and the master admin for a scoped Vault role, a Keycloak client and scoped Cloudflare tokens |
| **P3** Tenant isolation | A project's members read its secrets and admins write them; kubectl rights per project; the global read roles go |
| **P4** Cloud accounts | AWS organization with management, log-archive and workloads accounts; Terraform state behind a role assumed for one hour |

OpenStack stays out of scope: it hosts the cluster and keeps one admin account to deploy it.

## Access matrix

![CNP access matrix](diagrams/access-matrix.light.png#only-light)
![CNP access matrix](diagrams/access-matrix.dark.png#only-dark)

Read a row for everything an identity can do, a column for who can touch a resource. The
fill is the level after the plan; the frame and its tag show what the plan changes and in
which phase. A workload column touched by another project's identity is a leak.

## Keeping it true

The memberships and grants live in [`architecture/cnp/model.yaml`](https://github.com/3-Istor/cnp-docs/tree/main/architecture/cnp/model.yaml)
next to the rest of the platform: `deprecated` marks what the plan removes, `planned` what
it adds, and `target` the phase. When a phase is done, delete its `deprecated` entries and
drop `status` and `target` from its `planned` ones; the matrix then shows the new state with
fewer changes, until none is left.

Two facts are not verified yet, badged `G15` and `G16` in the model: the policy of the
CMP's Vault token and the scope of its Cloudflare token.
