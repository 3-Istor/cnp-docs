# Project Deletion Protection

A project can be locked against deletion. One flag protects the project and all its apps, present and future. Once it is on, only a person can delete anything, and only after turning it off in the CMP: no script, AI assistant or `kubectl` can do it, even by accident. Off by default.

## Turning it on or off

- **Registry**: `spec.deletionProtection: true` in `cnp-projects/registry/projects/<project>.yaml`. It is the only source of truth. The CMP writes it, and removes the key when protection is turned off.
- **CMP**: project "···" menu, "Turn on deletion protection" / "Turn off deletion protection" (turning off asks for confirmation). API: `PUT /api/projects/{name}/deletion-protection` with `{"enabled": true|false}`.

| Caller | Can toggle |
| --- | --- |
| Project admin | yes |
| Platform admin | yes (API) |
| Assistant (AI agent, MCP token) | never, `403` |

## What it blocks

| Path | While protection is on |
| --- | --- |
| CMP UI | "Delete project" and "Delete application" are greyed with "Deletion protection is on" |
| CMP API | `409` on any deletion of the project or of one of its apps |
| MCP | `delete_app` gets the same `409` (it calls the REST route) |
| `kubectl` | Deleting a labelled object is refused, and so is removing or changing the label (only Argo CD may) |
| CNPG | Deleting a `Cluster` in a protected namespace is refused (its PVCs follow it) |
| `cnp-clean.sh` | The project is skipped in every scope (`--project`, `--cloud`, `--all`), in dry-run as with `--yes`: `skipped: <p> (deletion protection)`. The script refuses to run if the registry is unreadable |

### Cluster side

The ApplicationSets and the `cnp-project-base` chart put the label `cnp.3istor.com/deletion-protected: "true"` on:

- the Argo CD Applications `<p>-appproject`, `<p>-bootstrap` and the prod Application of each app;
- both AppProjects;
- the `<p>-system` namespace and the prod namespaces of the apps.

A `ValidatingAdmissionPolicy`, `cnp-deletion-protection`, enforces it (on-prem, EKS and GKE). It refuses:

1. deleting a labelled namespace, Application or AppProject, with no exception;
2. an update that removes or changes the label, unless the caller is Argo CD;
3. deleting a CNPG `Cluster` in a labelled namespace.

A commit that removes a protected app or project from the registry is blocked too: Argo CD's prune is refused. See [GitOps with Argo CD](04-gitops-argocd.md) for the ApplicationSets.

## Turning it off

Turning protection off removes the registry key, then Argo CD removes the labels. Until it has, about a minute, any deletion answers `409` "being turned off on the cluster, try again in a minute". Without this check, a deletion started too early would remove the app from the registry while the label is still there: admission would refuse the prune and the objects would stay stuck.

## Not covered

- Staging (`-stg`) objects: only prod carries the label, since turning staging off prunes the `-stg` objects. Deleting the whole app stays blocked by the `409`.
- A PVC or a Secret deleted by hand inside a protected namespace. Data safety is the job of [database backups](../01-architecture/08-database-backups.md).
- A direct commit to `cnp-projects` that removes the flag, and a manual `terraform destroy` on a state. Both remain the platform team's power.

See also: [cleanup tool](../06-multicloud/05-nettoyage-et-decommissionnement.md) (`cnp-clean.sh`), [MCP integration](../05-cmp-backend-api/11-cmp-mcp-integration.md).
