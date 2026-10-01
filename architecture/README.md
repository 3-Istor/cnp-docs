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
| `multicloud` | What stays on the on-prem hub, what goes with each project to its target cloud, which paths cross the VPN | Multicloud index |
| `access-cmp-diff`, `access-dev-diff`, `access-admin-diff`, `access-root-diff` | What the segregation plan changes for one identity | Access segregation |
| `access-tenant-secrets-current`, `-target` | Who reaches one project's secrets, today and after the plan | Access segregation |
| `platform-overview` lenses `cmp-*`, `dev-*` | The platform map lit for one identity, today and after the plan | Access segregation |

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
`target`. Access views and lenses need noodle with access views (phase 6b,
TheGostsniperfr/Noodle#53). Topology views list their elements explicitly, so the
access elements never enter them.
