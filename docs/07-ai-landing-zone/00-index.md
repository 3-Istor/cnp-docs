# Landing zone IA (proposition)

> **Statut : proposition du 2026-10-22, rien n'est déployé.** Tous les éléments ajoutés
> au modèle Noodle portent `status: planned` et une cible : `P0`–`P2` pour les écarts de la
> landing zone actuelle, `IA-1`–`IA-3` pour la landing zone IA.

La landing zone IA se greffe sur la plateforme existante. Les appels des agents vers un
modèle ou un outil passent par une AI gateway (Agent Router, ex-Envoy AI Gateway) posée
sur l'Envoy Gateway déjà en place. La gateway porte l'authentification Keycloak,
l'autorisation par projet, les quotas en tokens, le choix du modèle et les traces.

## Décisions

| # | Décision | Choix proposé |
|---|---|---|
| D1 | AI gateway | Agent Router sur l'Envoy Gateway existant, une gateway par cluster (on-prem, EKS, GKE) |
| D2 | Point de sortie | Tout appel LLM ou outil passe par la gateway ; aucune clé provider dans un pod |
| D3 | Identité | Un client Keycloak par agent, jetons ≤ 15 min, standard token exchange pour agir au nom d'un utilisateur |
| D4 | Policy | Kyverno en Enforce à l'admission ; autorisation native de la `SecurityPolicy` à la gateway, puis OPA pour l'autorisation par outil |
| D5 | Isolation | Default-deny Cilium entre tenants et en egress ; CRD `Sandbox` + gVisor pour les agents qui exécutent du code |
| D6 | Modèles | Gemini (free tier) et Ollama sur le GPU du homelab via WireGuard, avec bascule automatique sur Ollama |
| D7 | Observabilité | OpenTelemetry GenAI vers Alloy puis Tempo ; tokens, coût et latence dans Grafana |

Prérequis issus de la landing zone actuelle (P0) : isolation inter-tenant (G1), contexte
de sécurité des pods (G8), branches GitOps protégées (G17), sauvegarde des bases
CloudNativePG.

## Schémas

### Tech stack : actuel et ajouts

![CNP AI landing zone tech stack](../01-architecture/diagrams/ai-landing-zone.light.png#only-light)
![CNP AI landing zone tech stack](../01-architecture/diagrams/ai-landing-zone.dark.png#only-dark)

### AI gateway

![CNP AI gateway](../01-architecture/diagrams/ai-gateway.light.png#only-light)
![CNP AI gateway](../01-architecture/diagrams/ai-gateway.dark.png#only-dark)

### Un appel de modèle

![A model call through the AI gateway](../01-architecture/diagrams/ai-gateway-call.light.png#only-light)
![A model call through the AI gateway](../01-architecture/diagrams/ai-gateway-call.dark.png#only-dark)

Sources : `architecture/cnp/views/ai-landing-zone.yaml`, `ai-gateway.yaml` et
`ai-gateway-call.yaml`, construites avec `./architecture/build-diagrams.sh`.
