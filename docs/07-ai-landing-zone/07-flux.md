# Les flux, pas à pas

> **Statut : v1.1 du 2026-11-22** (v1.0 validée par Brian le 2026-10-25 ; révisée pour l'API
> Anthropic seule et le code mode). Cette page décrit ce qui se passe, dans l'ordre,
> quand un agent travaille : qui parle à qui, quels contrôles s'appliquent, où ils tournent,
> ce qui est compté pour le FinOps. Les décisions sont dans [02-spec.md](02-spec.md).

## 1. Qui est qui, et où ça tourne

| Acteur | Où il tourne | Ce qu'il fait |
|---|---|---|
| **Agent** | Un pod dans le namespace de l'agent (phase 2) ; en phase 1, une app | La boucle de travail : envoie les messages au modèle, exécute les outils que le modèle demande, recommence jusqu'à la réponse finale. Il n'a aucune clé et ne parle qu'à la gateway. |
| **agentgateway** | Namespace `ai-gateway` | La seule porte : vérifie le jeton, les droits et le budget, applique les guardrails, choisit le backend, compte les tokens, émet traces et logs. Elle route trois sortes de trafic : modèles, outils (MCP) et autres agents (A2A). |
| **Service de rate limit et Redis** | Namespace `ai-gateway` | Tient les compteurs de budget par agent et par jour. |
| **Modèle** | API Anthropic (Internet), élément externe (§5 de la spec) | Produit le texte et les demandes d'outils. Il ne voit jamais l'agent directement. |
| **Serveurs d'outils (MCP)** | Serveur MCP de la CMP ; serveurs tiers hébergés par kmcp ou ToolHive (D39) dans `mcp-tools` ; serveurs MCP d'une app (par exemple `epitweet-mcp`) | Exécutent les outils : lire des logs, ouvrir une PR, lire un fil, répondre. |
| **Sandbox** | Pod gVisor (ou Kata) créé à la demande dans le namespace de l'agent (phase 2) | Exécute le code que le modèle a écrit (`run_python`, code mode) ; ne joint que le DNS et la route MCP de la gateway. |
| **Keycloak** | Déjà en place sur le cluster | Donne à l'agent un jeton de 5 minutes. |
| **Alloy, Tempo, VictoriaMetrics, Loki, Grafana** | Namespace `observability` | Traces, métriques de tokens, logs d'audit, tableaux. |

## 2. Un appel de modèle

C'est le flux de base, déjà en phase 1. Chaque étape dit ce qui se passe quand elle échoue.

| # | Étape | Où | Si ça échoue |
|---|---|---|---|
| 1 | L'agent demande un jeton à Keycloak (`client_credentials`, ou token exchange s'il agit pour un utilisateur) | Keycloak | Pas de jeton, pas d'appel |
| 2 | L'agent envoie `POST /v1/chat/completions` à la gateway, au format OpenAI, avec son jeton | agentgateway | — |
| 3 | Vérification du jeton : signature, émetteur, audience `ai-gateway`, expiration | agentgateway (`jwtAuthentication`) | 401 |
| 4 | Autorisation : le projet du jeton a-t-il droit à ce modèle ? | agentgateway (`authorization`, règles CEL) | 403 |
| 5 | Budget : reste-t-il des tokens à cet agent aujourd'hui ? La gateway réserve une estimation. | agentgateway + service de rate limit | 429 « budget épuisé » |
| 6 | Guardrails d'entrée : la regex masque les secrets et les données personnelles. Pas de classifieur (D10). | agentgateway (`promptGuard`) | Passage masqué |
| 7 | Backend : l'API Anthropic, seul backend. Pas de repli. | agentgateway | API injoignable ou plafond de dépenses atteint → 503 « service indisponible » |
| 8 | La gateway ajoute la clé Anthropic (l'agent ne l'a jamais), traduit la requête si elle arrive au format OpenAI (à vérifier à l'essai T1.1), puis l'envoie | agentgateway → API Anthropic | — |
| 9 | La réponse revient, en streaming si demandé | — | — |
| 10 | Guardrails de sortie : masquage des secrets et données personnelles dans la réponse | agentgateway (`promptGuard`, côté réponse) | Passage masqué |
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
   (`tool_calls` au format OpenAI, `tool_use` au format Anthropic) et non dans le texte : le nom de l'outil et ses arguments
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

**Recherche web.** L'outil serveur d'Anthropic `web_search_20260209` s'exécute chez
Anthropic, pas chez l'agent : la gateway ne voit que la requête et la réponse. Il n'est
autorisé que pour les agents qui l'ont déclaré (D41).

## 4. Générer une image

Supprimé le 2026-11-22 : pas de GPU, donc pas de génération d'images (D35). Un graphique se
produit par le code mode (§5).

## 5. Code mode : exécuter du code (phase 2)

Le modèle écrit un programme qui appelle ses outils, au lieu d'enchaîner les appels d'outils
un par un (D38). Le code tourne dans le sandbox ; ses appels d'outils repassent par la
gateway.

1. Le modèle demande l'outil de plateforme `run_python` avec son code.
2. L'agent crée une `SandboxClaim` dans son namespace (il n'a que ce droit sur l'API
   Kubernetes). Le contrôleur agent-sandbox démarre un pod depuis le `SandboxTemplate` :
   gVisor par défaut (un sandbox attend dans le warm pool, D33), Kata pour le niveau élevé
   (quelques secondes de démarrage, D42).
3. L'agent obtient par token exchange un jeton restreint pour le sandbox : outils en lecture
   seule parmi ceux de l'agent, audience de la route MCP de la gateway, 5 minutes. Le moyen de
   le remettre au sandbox est à vérifier à l'essai.
4. L'agent envoie le code au routeur de sandbox. Dans le sandbox, les outils autorisés sont
   un module Python : `import tools` puis `tools.search_posts(...)`. Chaque appel est un
   `tools/call` sur `/mcp` de la gateway, avec ce jeton : même policy par outil, même budget,
   même trace qu'un appel direct (§3).
5. Le sandbox ne joint que le DNS et la route MCP de la gateway. Il n'a pas la clé Anthropic
   ni le jeton de l'agent. Un outil d'écriture est refusé à la gateway.
6. La sortie (texte) revient à l'agent ; les fichiers produits (PNG, CSV) reviennent comme
   artefacts de la tâche. Le sandbox est détruit à la fin de la tâche ou à l'expiration de sa
   durée de vie.
7. Comptage : les appels d'outils du code sont comptés par la gateway ; CPU et mémoire du
   sandbox, dans son namespace, sont vus par OpenCost.

**Écarté** : `code_execution` et le programmatic tool calling d'Anthropic, où le code et les
appels d'outils tournent dans le sandbox d'Anthropic, hors de notre policy et de nos traces.

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
| Guardrails regex | gateway | 2 | Masquage |
| Outils autorisés par agent | gateway (CEL sur le nom de l'outil) | 2 | Outil absent de la liste, appel refusé |
| Stake policy | serveur MCP de la CMP | 2 | allow, ask, PR, approve, deny |
| Plafond d'étapes | runtime de l'agent | 2 | Fin de la boucle |
| Isolation du code (`run_python`) | gVisor ou Kata, NetworkPolicy (DNS et route MCP de la gateway) | 2 | Le code ne voit ni le nœud, ni le réseau, ni les outils d'écriture |
| Rule of Two | CMP et validateur, à la déclaration de l'agent | 2 | Création bloquée si trois risques |
| Plafond de dépenses | console Anthropic | 1 | « service indisponible » pour tous |
| Autorisation par argument d'outil | OPA | 3 | 403 |

## 9. Ce qui est compté pour le FinOps

| Quoi | Compté par | Unité | Valorisé à |
|---|---|---|---|
| Appels de modèle | gateway | tokens en entrée, en sortie et lus en cache, par agent, projet, modèle | Tarif publié d'Anthropic, en euros ([06-finops-ia.md](06-finops-ia.md)) |
| Appels d'outils, code mode compris | gateway | appels, par agent et par outil | Non valorisés (compris dans l'infra) |
| Sandbox, gateway, agents | OpenCost | CPU, mémoire par namespace | Tarif on-prem du plan FinOps |
