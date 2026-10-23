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

### La landing zone IA en un coup d'œil

Une boîte par brique, quinze flux : le parcours d'une requête (numéros) et la mise en place (lettres). En clair ce que la landing zone IA ajoute, estompé ce qui tourne déjà.

![La landing zone IA en un coup d'œil](../01-architecture/diagrams/platform-ai-simple.light.png#only-light)
![La landing zone IA en un coup d'œil](../01-architecture/diagrams/platform-ai-simple.dark.png#only-dark)

### La plateforme déployée et sa landing zone IA

Ce qui tourne aujourd'hui, estompé, et toute la landing zone IA en clair, dans sa version retenue (option B : agentgateway et kagent). Chaque boîte dit ce que fait l'outil ; les zones sont les frontières d'isolation (namespace, niveau de runtime, réseau). Les numéros suivent une requête d'utilisateur vers un agent, qui appelle un modèle, un outil et un autre agent en A2A ; les lettres montrent ce qui est mis en place avant. Seuls les deux flux refusés par conception sont en rouge.

![The deployed platform and its AI landing zone](../01-architecture/diagrams/platform-ai-stack.light.png#only-light)
![The deployed platform and its AI landing zone](../01-architecture/diagrams/platform-ai-stack.dark.png#only-dark)

### Vue d'ensemble détaillée : tous les zooms sur une seule carte

Le hub on-prem tel qu'il tourne aujourd'hui (estompé) et toute la landing zone IA (en clair), au niveau de détail des zooms : filtres Envoy, contrôleurs, CRDs, clients Keycloak, pipeline Alloy. Les trois niveaux d'isolation du sandbox sont trois nœuds distincts côte à côte (gVisor, Kata, runc standard) : la plateforme peut en proposer un, deux ou les trois, selon que l'agent exécute du code ou non. EKS et GKE restent résumés à leur API server. Grande image : l'ouvrir dans draw.io (`platform-detailed.*.drawio`) pour zoomer.

![Platform overview, detailed](../01-architecture/diagrams/platform-detailed.light.png#only-light)
![Platform overview, detailed](../01-architecture/diagrams/platform-detailed.dark.png#only-dark)

### Carte globale : plateforme actuelle et landing zone IA

Toute la plateforme sur une seule toile : le hub on-prem, la landing zone IA proposée, GitHub, les modèles et les clusters EKS et GKE. Les numéros suivent un utilisateur, une app, un agent et une release ; les lettres, les tâches de fond.

![CNP global map](../01-architecture/diagrams/platform-global.light.png#only-light)
![CNP global map](../01-architecture/diagrams/platform-global.dark.png#only-dark)

### AI gateway : vue détaillée

L'AI gateway ouverte jusqu'à ses CRDs, ses contrôleurs, les filtres d'Envoy et les sous-modules du sidecar ExtProc, avec le chemin d'une requête en 16 étapes.

![AI gateway internals](../01-architecture/diagrams/ai-gateway-internals.light.png#only-light)
![AI gateway internals](../01-architecture/diagrams/ai-gateway-internals.dark.png#only-dark)

### Agent sandbox : vue détaillée

Comment un agent obtient un bac à sable isolé pour exécuter le code qu'il écrit : CRDs agent-sandbox, contrôleurs, garde-fous, et la chaîne kubelet, containerd, runsc jusqu'au noyau gVisor.

![Agent sandbox internals](../01-architecture/diagrams/agent-sandbox-internals.light.png#only-light)
![Agent sandbox internals](../01-architecture/diagrams/agent-sandbox-internals.dark.png#only-dark)

Variante Kata Containers (une micro-VM et un noyau par sandbox) :

![Agent sandbox, Kata tier](../01-architecture/diagrams/agent-sandbox-kata.light.png#only-light)
![Agent sandbox, Kata tier](../01-architecture/diagrams/agent-sandbox-kata.dark.png#only-dark)

Niveau standard (agents sans exécution de code, runc) :

![Agent runtime, standard tier](../01-architecture/diagrams/agent-sandbox-runc.light.png#only-light)
![Agent runtime, standard tier](../01-architecture/diagrams/agent-sandbox-runc.dark.png#only-dark)

### Outils MCP : vue détaillée

Comment un appel d'outil va de l'agent au serveur d'outils : le proxy MCP de la gateway, ToolHive pour les serveurs tiers, le serveur MCP du CMP et sa politique de stake.

![MCP tools internals](../01-architecture/diagrams/mcp-tools-internals.light.png#only-light)
![MCP tools internals](../01-architecture/diagrams/mcp-tools-internals.dark.png#only-dark)

### Identité des agents : vue détaillée

Un client Keycloak par agent, client_credentials quand l'agent agit seul, token exchange quand il agit pour un utilisateur, et ce que contient le jeton.

![Agent identity internals](../01-architecture/diagrams/agent-identity-internals.light.png#only-light)
![Agent identity internals](../01-architecture/diagrams/agent-identity-internals.dark.png#only-dark)

### Modèles : vue détaillée

Gemini en premier (quota gratuit, par projet), repli sur Ollama sur le GPU du homelab via WireGuard, et la route réservée aux projets Local.

![Models internals](../01-architecture/diagrams/models-internals.light.png#only-light)
![Models internals](../01-architecture/diagrams/models-internals.dark.png#only-dark)

### Observabilité GenAI : vue détaillée

De l'appel d'un agent à une trace et des compteurs de tokens par projet : les sources, le pipeline Alloy, Tempo, Prometheus et Grafana.

![GenAI observability internals](../01-architecture/diagrams/genai-observability-internals.light.png#only-light)
![GenAI observability internals](../01-architecture/diagrams/genai-observability-internals.dark.png#only-dark)

### Tech stack : actuel et ajouts

![CNP AI landing zone tech stack](../01-architecture/diagrams/ai-landing-zone.light.png#only-light)
![CNP AI landing zone tech stack](../01-architecture/diagrams/ai-landing-zone.dark.png#only-dark)

### AI gateway

![CNP AI gateway](../01-architecture/diagrams/ai-gateway.light.png#only-light)
![CNP AI gateway](../01-architecture/diagrams/ai-gateway.dark.png#only-dark)

### Un appel de modèle

![A model call through the AI gateway](../01-architecture/diagrams/ai-gateway-call.light.png#only-light)
![A model call through the AI gateway](../01-architecture/diagrams/ai-gateway-call.dark.png#only-dark)

Sources : `architecture/cnp/views/platform-ai-simple.yaml`, `platform-ai-stack.yaml`, `platform-detailed.yaml`, `platform-global.yaml`, `ai-gateway-internals.yaml`, `agent-sandbox-internals.yaml`, `agent-sandbox-kata.yaml`, `agent-sandbox-runc.yaml`, `mcp-tools-internals.yaml`, `agent-identity-internals.yaml`, `models-internals.yaml`, `genai-observability-internals.yaml`, `ai-landing-zone.yaml`, `ai-gateway.yaml` et
`ai-gateway-call.yaml`, construites avec `./architecture/build-diagrams.sh`.
