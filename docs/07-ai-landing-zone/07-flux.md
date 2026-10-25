# Les flux, pas à pas

> **Statut : v1.0 du 2026-10-25, validée par Brian.** Cette page décrit ce qui se passe, dans l'ordre,
> quand un agent travaille : qui parle à qui, quels contrôles s'appliquent, où ils tournent,
> ce qui est compté pour le FinOps. Les décisions sont dans [02-spec.md](02-spec.md).

## 1. Qui est qui, et où ça tourne

| Acteur | Où il tourne | Ce qu'il fait |
|---|---|---|
| **Agent** | Un pod dans le namespace de l'agent (phase 2) ; en phase 1, une app | La boucle de travail : envoie les messages au modèle, exécute les outils que le modèle demande, recommence jusqu'à la réponse finale. Il n'a aucune clé et ne parle qu'à la gateway. |
| **agentgateway** | Namespace `ai-gateway` | La seule porte : vérifie le jeton, les droits et le budget, applique les guardrails, choisit le backend, compte les tokens, émet traces et logs. Elle route trois sortes de trafic : modèles, outils (MCP) et autres agents (A2A). |
| **Service de rate limit et Redis** | Namespace `ai-gateway` | Tient les compteurs de budget par agent et par jour. |
| **Webhook de guardrail** | Namespace `ai-gateway` (petit service maison, phase 2) | Reçoit le prompt de la gateway et demande à Llama Guard, sur le serveur GPU, s'il est dangereux. |
| **Modèles** | Gemini (Internet) ; serveur GPU (homelab, §5 de la spec) | Produisent du texte, des embeddings, des classifications. Ils ne voient jamais l'agent directement. |
| **Serveurs d'outils (MCP)** | Serveur MCP de la CMP ; serveurs tiers lancés par ToolHive dans `mcp-tools` | Exécutent les outils : lire des logs, ouvrir une PR, chercher sur le web, générer une image. |
| **Sandbox** | Pod gVisor (ou Kata) créé à la demande dans le namespace de l'agent (phase 2) | Exécute le code que l'agent a écrit, sans réseau. |
| **Keycloak** | Déjà en place sur le cluster | Donne à l'agent un jeton de 5 minutes. |
| **Alloy, Tempo, VictoriaMetrics, Loki, Grafana** | Namespace `observability` | Traces, métriques de tokens, logs d'audit, tableaux. |

## 2. Un appel de modèle

C'est le flux de base, déjà en phase 1. Chaque étape dit ce qui se passe quand elle échoue.

| # | Étape | Où | Si ça échoue |
|---|---|---|---|
| 1 | L'agent demande un jeton à Keycloak (`client_credentials`, ou token exchange s'il agit pour un utilisateur) | Keycloak | Pas de jeton, pas d'appel |
| 2 | L'agent envoie `POST /v1/chat/completions` à la gateway, au format OpenAI, avec son jeton | agentgateway | — |
| 3 | Vérification du jeton : signature, émetteur, audience `ai-gateway`, expiration | agentgateway (`jwtAuthentication`) | 401 |
| 4 | Autorisation : le projet du jeton a-t-il droit à ce modèle ? Est-il Local ? | agentgateway (`authorization`, règles CEL) | 403 |
| 5 | Budget : reste-t-il des tokens à cet agent aujourd'hui ? La gateway réserve une estimation. | agentgateway + service de rate limit | 429 « budget épuisé » |
| 6 | Guardrails d'entrée (phase 2) : la regex masque les secrets et les données personnelles ; en phase 2, le webhook demande l'avis de Llama Guard | agentgateway (`promptGuard`) + webhook + serveur GPU | Prompt refusé ; si le webhook est injoignable, refus aussi |
| 7 | Choix du backend : projet Local → serveur GPU seulement. Sinon, une des clés Gemini (équilibrage P2C) ; sur 429 de Google, une autre clé ; si toutes saturent, le serveur GPU | agentgateway | Serveur GPU injoignable → 503 « service indisponible » |
| 8 | La gateway ajoute la clé du fournisseur (l'agent ne l'a jamais), puis envoie la requête | agentgateway → Gemini ou serveur GPU | — |
| 9 | La réponse revient, en streaming si demandé | — | — |
| 10 | Guardrails de sortie (phase 2) : masquage des secrets et données personnelles dans la réponse | agentgateway (`promptGuard`, côté réponse) | Passage masqué |
| 11 | Comptage : la gateway lit l'usage réel (tokens en entrée et en sortie) et remplace l'estimation de l'étape 5 | agentgateway + service de rate limit | — |
| 12 | Trace OpenTelemetry, métriques de tokens, ligne d'audit (utilisateur, agent, projet, modèle, tokens, décision) | Alloy → Tempo, VictoriaMetrics, Loki | — |
| 13 | La réponse arrive à l'agent | — | — |

## 3. Comment un agent utilise un outil

Le modèle n'exécute jamais rien lui-même. Il **demande** un outil ; c'est l'agent qui
l'exécute.

1. **L'agent récupère la liste des outils** auxquels il a droit : `tools/list` sur `/mcp` de la
   gateway, qui ne renvoie que les outils autorisés pour cet agent. Chaque outil a un nom, une
   description et un schéma JSON de ses arguments.
2. **L'agent appelle le modèle** avec les messages et cette liste d'outils (flux §2).
3. **Le modèle répond par une demande d'outil**, dans un champ structuré de la réponse
   (`tool_calls` au format OpenAI) et non dans le texte : le nom de l'outil et ses arguments
   en JSON, par exemple `{"name": "cmp__get_app_logs", "arguments": {"app": "twitter-clone"}}`.
   C'est le *function calling* : il faut un modèle entraîné pour ça.
4. **L'agent exécute l'appel** : `tools/call` sur `/mcp` de la gateway, avec son jeton. La
   gateway vérifie le jeton et que cet outil est autorisé pour cet agent (règle CEL sur le nom
   de l'outil), puis route vers le bon serveur : la CMP, qui applique sa stake policy (allow,
   ask, PR, approve, deny), ou un serveur ToolHive, qui vérifie le jeton à son tour.
5. **Le résultat revient à l'agent**, qui l'ajoute aux messages comme un message `tool`. C'est
   du contenu non fiable (des logs, une page web) : il est traité comme des données, jamais
   comme des instructions.
6. **L'agent rappelle le modèle** avec ce résultat. Le modèle demande un autre outil ou donne
   sa réponse finale. La boucle s'arrête à la réponse, au plafond d'étapes (D15) ou quand le
   budget est épuisé (429).

Dans un agent kagent déclaratif, cette boucle est fournie par le runtime de kagent. Dans un
agent BYO, elle est écrite avec un framework (LangGraph par exemple). Dans l'app UC7 de la
phase 1, elle est dans le code de l'app.

## 4. Générer une image (phase 2)

1. Le modèle demande l'outil `generate_image` (flux §3).
2. L'agent appelle `/mcp` de la gateway ; la gateway vérifie que l'agent a droit à cet outil.
3. Le serveur MCP `generate_image` (lancé par ToolHive dans `mcp-tools`) envoie le workflow
   versionné au service d'image du serveur GPU, par HTTPS avec son propre jeton.
4. Le serveur GPU génère l'image ; le serveur MCP la renvoie à l'agent (lien ou contenu).
   Avant, il décharge le modèle de texte du GPU et prend un verrou ; après, il le rend (spec §5,
   « Changer de modèle pour une image »). Pendant ce temps, le texte des projets Cloud part sur
   Gemini, celui des projets Local attend.
5. Comptage : un appel d'image n'est pas compté en tokens. La gateway limite le nombre
   d'appels par agent et par jour (D35) et le FinOps le valorise au TCO du serveur GPU.

## 5. Exécuter du code (phase 2)

Ce flux ne passe **pas** par la gateway : le code ne sort jamais du sandbox.

1. L'agent crée une `SandboxClaim` dans son namespace (il n'a que ce droit sur l'API
   Kubernetes).
2. Le contrôleur agent-sandbox démarre un pod depuis le `SandboxTemplate` : gVisor, ou Kata
   pour du code non fiable. Un sandbox gVisor attend déjà dans le warm pool (D33) : la tâche
   démarre tout de suite ; un sandbox Kata démarre en quelques secondes.
3. L'agent envoie le code au routeur de sandbox ; le code tourne ; la sortie revient.
4. Le sandbox n'a aucun réseau sauf le DNS, ni secret, ni jeton. Il est détruit à la fin de
   la tâche ou à l'expiration de sa durée de vie.
5. Comptage : CPU et mémoire du sandbox, dans son namespace, sont vus par OpenCost.

## 6. Un agent qui en appelle un autre (A2A, phase 2)

1. L'agent A découvre l'agent B par son Agent Card.
2. A envoie la tâche à la gateway, route A2A de B, avec son jeton (ou un jeton échangé s'il
   agit pour un utilisateur).
3. La gateway vérifie que A a le droit d'appeler B (liste des appelants autorisés de B).
4. B fait son travail : ses propres appels de modèle et d'outils passent par la gateway et
   sont comptés sur **son** budget.
5. La réponse revient à A. La trace relie A et B. B n'a jamais plus de droits que A pour
   cette tâche.

## 7. Un utilisateur parle à un agent

1. L'utilisateur ouvre l'interface (la CMP, ou l'app qui embarque l'agent).
2. Cloudflare, puis Envoy Gateway, qui fait le login OIDC : c'est l'entrée habituelle des apps.
3. Envoy Gateway transmet à agentgateway (route A2A de l'agent) avec le jeton de l'utilisateur.
4. L'agent agit pour cet utilisateur par token exchange (`sub` = l'utilisateur, `azp` =
   l'agent) : il ne peut jamais faire plus que ce que l'utilisateur a le droit de faire.

## 8. Les contrôles, en un tableau

| Contrôle | Où | Phase | Effet |
|---|---|---|---|
| Identité (jeton de 5 minutes) | Keycloak, puis gateway | 1 | 401 sans jeton valide |
| Autorisation par projet et par modèle | gateway (CEL) | 1 | 403 |
| Budget de tokens par agent et par jour | gateway + rate limit | 1 | 429, réinitialisé chaque jour |
| Sortie réseau | Cilium | 0 et 1 | Connexion refusée hors gateway |
| Guardrails regex | gateway | 2 | Masquage ou refus |
| Guardrails Llama Guard | gateway → webhook → serveur GPU | 2 | Refus ; refus aussi si injoignable |
| Outils autorisés par agent | gateway (CEL sur le nom de l'outil) | 2 | Outil absent de la liste, appel refusé |
| Stake policy | serveur MCP de la CMP | 2 | allow, ask, PR, approve, deny |
| Plafond d'étapes | runtime de l'agent | 2 | Fin de la boucle |
| Isolation du code | gVisor ou Kata, NetworkPolicy | 2 | Le code ne voit ni le nœud ni le réseau |
| Rule of Two | CMP, à la création de l'agent | 2 | Création bloquée si trois risques |
| Autorisation par argument d'outil | OPA | 3 | 403 |

## 9. Ce qui est compté pour le FinOps

| Quoi | Compté par | Unité | Valorisé à |
|---|---|---|---|
| Appels de modèle Gemini | gateway | tokens en entrée et en sortie, par agent, projet, modèle | Tarif payant publié de Gemini ([06-finops-ia.md](06-finops-ia.md)) |
| Appels de modèle au serveur GPU | gateway | tokens | TCO on-prem, au prorata |
| Images générées | gateway (nombre d'appels de l'outil) | appels | TCO on-prem |
| Sandbox, gateway, agents | OpenCost | CPU, mémoire par namespace | Tarif on-prem du plan FinOps |
