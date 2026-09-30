---
name: cnp-diagram
description: Create or update a CNP architecture diagram in the house style (Cocoon dark palette, official logos, draw.io output). Use when asked for an architecture, flow, deployment or security diagram of the CNP platform, or when a change to the platform should be reflected in an existing diagram.
---

# CNP architecture diagrams

Diagrams are generated, not drawn. The source of truth is a YAML spec in
`architecture/specs/<id>.yaml`; `tools/archgen` turns it into
`architecture/diagrams/<id>.drawio` and refuses to emit it while the lint finds a
collision. Hand edits in draw.io are for throwaway variants only; a canonical change
goes through the spec so the palette, the IDs and the PR diff stay consistent.

## Workflow

1. **Facts first.** Read the code that implements what the diagram shows (Helm
   templates, Terraform, K8s manifests), not only `docs/`. When code and docs disagree,
   draw the code, add `badge: Gx` on the node and an entry in the "Gaps" card.
2. **One question per diagram.** Write it as the subtitle. Anything that does not help
   answer it goes to "Out of scope" in `meta`.
3. **Write the spec in English** (proper nouns aside), then lint and build both themes:
   ```bash
   B=architecture/diagrams/<id> I=.claude/skills/cnp-diagram/icons
   for t in dark light; do
     go run ./tools/archgen -theme $t -icons $I -o $B.$t.drawio architecture/specs/<id>.yaml
     drawio -x -f png -s 2 --border 20 -o $B.$t.png $B.$t.drawio
   done
   ```
4. **Look at the PNG.** The lint catches geometry, not taste.
5. For the docs, export the editable SVG:
   `drawio -x -f svg --embed-diagram -o $B.$t.drawio.svg $B.$t.drawio`.
   The PNG is the one to paste in chats.

## Semantics

- **An arrow is a connection, drawn from the side that opens it (client) to the side
  that listens (server).** Traffic carried back over a connection opened the other way,
  such as a tunnel, is said in the label ("[2] relayed via tunnel ◀").
- **Draw requests only, never responses.** When a specific response changes the path
  (302, 401, 500, retry), write it as a note next to the component that sends it
  ("Without a session, Envoy replies 302 to auth…"), not as a return arrow.
- **Topology and sequence are separate diagrams.** A structural diagram numbers only the
  nominal path, with one step per connection and each step used once; the lint rejects
  anything else. A flow that reuses connections or goes back and forth (OIDC login,
  token exchange, retries, sagas) gets its own sequence diagram. Its connections stay on
  the topology, unnumbered, with a label pointing to that sequence.
- **Egress leaves from the right or bottom border, ingress enters on the left or top.**
  The lint enforces it. `against_flow: true` is the only exception, for outbound
  connections that run against the reading direction: tunnels, agents calling home, a
  hairpin to a public endpoint. Use it knowingly.
- **`port:` is the listening port**, drawn as a badge straddling the server's border,
  like `TCP 8080` or `UDP 7844`. Client ports are ephemeral and are not drawn.
- **Arrow label: `[step] verb · protocol`**, like `[6] proxy · HTTP`. The port is on the
  badge, so it is not repeated in the label.
- **Numbers `[1]…[n]` are the request path (data plane), in time order. Letters `[A]…`
  are background work (control plane):** tunnels, syncs, operators. Two cards list them.
- **`kind: link` is an object reference** (parentRef, targetRef, envFrom, secret name).
  It is dotted, has an open arrow, carries no step and ignores the side rule.
- **Node text is the C4 triptych:** `title` in bold, `tech` in brackets and italics,
  `desc` as one line about the component's responsibility. Technical specifics go in
  the step and gap cards, not in the box.

## Visual language

| Element | Rule |
|---|---|
| Node `kind` | `frontend` cyan · `backend` emerald (services, routing) · `database` violet (data and secrets) · `cloud` amber (edge, tunnels, providers) · `security` rose (identity, policies) · `bus` orange · `external` slate |
| Node `icon` | Official logo when the box is a product (`cloudflare`, `envoy`, `keycloak`, `vault`, `cnpg`, `argo`, `cilium`, `kubernetes`). Kubernetes resource icon (`k8s-svc`, `k8s-deploy`, `k8s-pod`, `k8s-secret`, `k8s-ns`, `k8s-sa`, `k8s-crd`) when the box is a K8s object. CRDs without an official icon use `k8s-crd`, as Argo CD does. |
| Node `shape` | `box` default · `cylinder` for a datastore · `actor` for a human |
| Zone `kind` | `region` = infra or trust perimeter. `group` = functional category, named by role first ("GATEWAY API", "IDENTITY", "SECRETS"), namespace in `sub`. |
| Zone `icon` | Always set. Environment on regions (`globe` internet, `cloudflare` SaaS, `rack` on-prem, a provider logo for a cloud), product or `k8s-ns` on groups. |
| Zone `color` | Sibling zones get different colours so boundaries read at a glance. |
| Edge `kind` | `flow` connection · `auth` authentication · `tunnel` outbound tunnel set up in advance · `async` background sync · `blocked` must not happen · `link` reference |

Palettes are Tailwind v3: 400 strokes on slate-950 for dark, 600 strokes on 50 fills for
light, both checked for WCAG AA text contrast. Colour is never the only carrier: kind is
also shown by icon, dash pattern and legend. Icons live in `icons/`. Add `<name>.svg`
from the CNCF artwork repo, the Kubernetes community icon set or simple-icons, recoloured
with the brand colour. Add `<name>.dark.svg` / `<name>.light.svg` when one variant
disappears on a background.

## Layout and cable management

- Leave **60–80 px corridors between zones**. Edges travel in corridors, never through
  a zone they do not serve. Inside a zone, 24 px minimum between nodes.
- Every edge has an explicit orthogonal `path` whose first and last points sit on the
  borders of `from` and `to`. No diagonals. Keep port badges clear of the node icon, top-left.
- **Backbones, not spaghetti.** When several edges go the same way, give them parallel
  lanes 15–20 px apart in the same corridor and turn them together, like a PCB bus or
  cable management in a rack. Fan out only at the last bend before each target.
- Put a label on the longest straight segment, or set `label_at` on the drawn path in a
  clear spot. Never on a bend, a box, a zone title, a port badge or another line.
- A perpendicular crossing is fine: draw.io draws a bridge. Two edges running on top of
  each other are not: the lint rejects it.
- Flow reads left to right: actors → edge/SaaS → cluster → data.

## Required parts

Title with a one-line question, scope and out-of-scope lines, version, source path.
Cards under the diagram: `Legend` (`legend: true`, generated from the kinds used),
`Request path · data plane`, `Background · control plane`, and `Gaps between docs and
code` when there are any.

## Stable IDs

IDs are semantic (`envoy-shared-gateway`, `e-4-connector-envoy`) and never renamed for
cosmetic reasons. The PR diff matches elements across versions by ID.
