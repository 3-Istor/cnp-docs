# Architecture diagrams

The CNP diagrams are generated with [noodle](https://github.com/TheGostsniperfr/Noodle),
not drawn. `cnp/` is one system: `model.yaml` says what exists, each file in `views/` is
one diagram, and `layouts/` places the topology views.

| View | Question | Published in |
|---|---|---|
| `platform-overview` | What runs on the cluster, and which path do a request, a new project, a deployment and a secret take? | System overview |
| `app-runtime` | Seen from inside one app: who reaches it, what it reaches, what should stay out, what the platform hands it | Tenancy and isolation |
| `runtime` | How a request reaches a tenant app, where it is authenticated, what keeps tenants apart | Network topology |
| `oidc-login` | The login when there is no session | Network topology |
| `multicloud` | What runs where between the on-prem hub, EKS and GKE, and which links cross between them? | Multicloud index, Multicloud schemas |
| `multicloud-links` | Which links cross between the on-prem hub and the public clouds, who opens them, and what allows them? | Multicloud schemas |
| `multicloud-policies` | What may reach a project's workloads on each cluster, and what enforces it? | Multicloud schemas |
| `cloud-up` | What happens, in order, when an operator runs `cnp.sh up` for a public cloud, and what survives `cnp.sh down`? | Multicloud schemas |
| `agent-kit` | How does an AI assistant reach CNP, what checks its calls, and where do its changes land? | CMP MCP integration |
| `agent-connect` | What happens the first time an assistant calls CNP, until its first tool call runs with the user's rights | CMP MCP integration |
| `access-matrix` | Who may act on what, and what the segregation plan changes, phase by phase | Access segregation |

To change a diagram, edit the model (facts), the view (what is shown) or the layout
(where), then rebuild every view, dark and light, into `docs/01-architecture/diagrams/`:

```bash
./architecture/build-diagrams.sh
```

It needs the `noodle` CLI (Claude Code noodle plugin, or `nix run github:TheGostsniperfr/Noodle`),
`drawio` for the PNG export and `pngquant` to compress it (`PNG_QUALITY`, default `80-95`;
`off` keeps draw.io's export). noodle refuses to build a view while boxes, labels or arrows
collide. Look at the PNG before opening the PR: the lint catches geometry, not readability.

Every fact comes from the code (K3s, infra-templates, cnp-project-base, CMP,
app-templates). Where the docs say otherwise, the diagram draws the code and adds a
`Gx` badge with an entry in its *Gaps* card.

Project logos not bundled with noodle live in `.noodle/icons/`.

Access: `memberships` and `grants` in the model say who may act on what. `deprecated`
marks what the segregation plan removes, `planned` what it adds, with the phase as
`target`. The `access-matrix` view needs noodle with matrix views (phase 6b,
TheGostsniperfr/Noodle#54). Topology views list their elements explicitly, so the
access elements never enter them.
