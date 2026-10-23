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
| `ai-landing-zone` | What is the platform built with today, and what do the landing zone fixes and the AI landing zone add? (proposal, planned) | AI landing zone |
| `platform-global` | What does the whole platform look like, current hub plus the proposed AI landing zone, on one canvas? (proposal, partly planned) | AI landing zone |
| `ai-gateway-internals` | What is the AI gateway made of, and what happens inside it between a request and the answer? (proposal, planned) | AI landing zone |
| `agent-sandbox-internals` | How does an agent get an isolated place to run the code it writes? (proposal, planned) | AI landing zone |
| `agent-sandbox-kata` | The same sandbox with one micro-VM and kernel per sandbox (Kata tier, proposal) | AI landing zone |
| `agent-sandbox-runc` | Agents that never run generated code: standard runtime and what keeps them in check (proposal) | AI landing zone |
| `mcp-tools-internals` | How does an agent's tool call reach a tool server, and who checks it on the way? (proposal) | AI landing zone |
| `agent-identity-internals` | Who is an agent for the platform, how does it get a token, and how does it act for a user? (proposal) | AI landing zone |
| `models-internals` | Where does a model call end up, what limits it, and how does it fall back to the homelab GPU? (proposal) | AI landing zone |
| `genai-observability-internals` | How does one agent call become a trace and token counts per project? (proposal) | AI landing zone |
| `platform-ai-stack` | What runs today and what the AI landing zone adds (option B: agentgateway, kagent): which tool does what, where the isolation boundaries are, how an agent request flows (proposal) | AI landing zone |
| `platform-detailed` | How do all the AI landing zone parts connect to each other and to the hub, at zoom level of detail? (proposal) | AI landing zone |
| `ai-gateway` | How does an agent's call reach a model or a tool, and what checks it on the way? (proposal, planned) | AI landing zone |
| `ai-gateway-call` | What happens, in order, when a hosted agent asks a model and Gemini's free quota is spent? (proposal, planned) | AI landing zone |

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
