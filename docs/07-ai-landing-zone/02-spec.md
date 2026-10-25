# Spec de référence de la landing zone IA

> **Statut : brouillon v0.6 du 2026-10-25, à valider.** Rien n'est déployé. Cette page est la
> source de vérité des décisions et des conventions : en cas de contradiction avec
> `00-index.md`, une vue ou un schéma, c'est elle qui prime. Les phases et leurs vérifications
> sont dans [03-phases.md](03-phases.md).

**Sources.** La page [Stack cible et compromis](01-stack-cible.md) ; le modèle Noodle
(`architecture/cnp/model.yaml`), en particulier les éléments `ov-*` et `hl-*` de l'option B
(agentgateway et kagent), qui sont la version la plus récente ; le doc de synthèse « Landing
zone IA de CNP » ; les décisions de Brian du 2026-10-25. Branche lue :
`feat/ai-landing-zone-diagrams`, commit `eb2bfa5`.

## 1. Objet et périmètre

CNP veut laisser des agents IA travailler sur la plateforme sans qu'ils puissent faire de
dégâts. La landing zone IA est la plateforme actuelle plus une couche de contrôle qui traite
chaque agent comme un workload non fiable.

**Cadre : démonstration ou MVP.** La taille de l'équipe et la haute disponibilité ne sont pas
des contraintes. Les premiers utilisateurs sont les développeurs de CNP et les tenants.
Gemini est ouvert à tous ceux qui ont accès à la démo, pour un faible volume.

| Dans le périmètre | Hors périmètre |
|---|---|
| Gateway IA, identité des agents, isolation, outils MCP, A2A, guardrails, observabilité GenAI, budgets, audit | Le contenu du serveur GPU : seul son contrat est défini (§5) |
| Le cluster on-prem (K3s) | Les clusters EKS et GKE : le modèle prévoit une gateway par cluster, la démo n'en utilise qu'une |
| Les quatre prérequis P0 de la plateforme actuelle (phase 0) | Conformité AI Act, haute disponibilité, exploitation en production |

## 2. Principes

1. **Une seule porte.** Tout appel d'un agent ou d'une app vers un modèle, un outil ou un
   autre agent passe par agentgateway. Aucune clé de provider dans un pod d'agent ou d'app.
2. **Une identité par agent.** Un client Keycloak par agent, des jetons courts, sans refresh.
3. **Refus par défaut.** Réseau en default-deny, autorisation en deny par défaut.
4. **Isolation selon le risque.** runc et seccomp si l'agent n'exécute pas de code, gVisor
   pour le code généré, Kata pour le code non fiable.
5. **Rule of Two.** Un agent cumule au plus deux éléments parmi : données privées, contenu non
   fiable, sortie vers l'extérieur. Au-delà, on coupe en deux agents.
6. **Une panne donne un refus net.** Si une dépendance manque, la réponse est « service
   indisponible », sans mode dégradé.
7. **Standards d'abord.** API compatible OpenAI, MCP, A2A, OpenTelemetry GenAI : les briques
   peuvent changer, les protocoles restent.
8. **Tout vient de Git.** Agents, routes, policies et serveurs MCP sont déclarés dans Git et
   appliqués par Argo CD, comme le reste de la plateforme.

## 3. Registre des décisions

Statuts : **décidé** (par Brian), **hérité** (des pages et du modèle, non rediscuté),
**proposé** (par Claude dans cette spec, à valider). D1 à D7 gardent la numérotation de
`00-index.md`.

| ID | Décision | Statut | Phase | Note |
|---|---|---|---|---|
| D1 | **agentgateway** pour tout le trafic des agents (LLM, MCP, A2A), en GatewayClass `agentgateway` à part. Envoy Gateway reste l'entrée des apps et des utilisateurs. | hérité | 1 | Remplace Agent Router. Plan B au §8. |
| D2 | Tout appel LLM, outil ou agent passe par la gateway ; la clé Gemini n'est montée que dans la gateway. | hérité | 1 | |
| D3 | Un client Keycloak par agent. Seul, l'agent utilise `client_credentials` ; pour un utilisateur, le token exchange (Keycloak 26.2 ou plus). Cible : le jeton du ServiceAccount remplace le secret client (federated client authentication, Keycloak 26.6). Jeton d'accès de **5 minutes, sans refresh**. | proposé | 1 puis 2 | Remplace « ≤ 15 min ». Phase 1 : secret client depuis Vault ; phase 2 : jeton du ServiceAccount. |
| D4 | Kyverno en Enforce à l'admission. Autorisation à la gateway par `AgentgatewayPolicy` : JWT obligatoire, règles CEL sur les claims et par outil, deny par défaut. OPA en `ext_authz` seulement si le CEL ne suffit plus. | hérité | 1 puis 3 | Cedar reste confiné au proxy ToolHive. |
| D5 | Cilium default-deny entre tenants et en egress, egress par FQDN. RuntimeClass `gvisor` et `kata`. agent-sandbox en `v1beta1`. | hérité | 0 puis 2 | |
| D6 | **Gemini par défaut** (niveau gratuit, plusieurs clés de comptes gratuits, équilibrées (D23)). Repli sur le serveur GPU quand Gemini répond 429. Les projets Local n'ont que le serveur GPU. Pas de données clients sur Gemini. | décidé | 1 | Le modèle prévoit une seule clé : à mettre à jour. Réserves au §9. |
| D7 | OpenTelemetry GenAI → Alloy → Tempo pour les traces ; métriques dans **VictoriaMetrics** ; Grafana ; pas de Langfuse (environ 21,5 Gio de RAM au minimum). | hérité | 1 | Arize Phoenix si une interface pensée pour les LLM devient nécessaire. |
| D8 | **kagent** : CRD `Agent` (`kagent.dev/v1alpha2`, type BYO pour une image LangGraph, ADK ou autre) et `ModelConfig` pointé sur la gateway. Chaque agent est publié en A2A avec son Agent Card. | hérité | 2 | À vérifier : `ModelConfig` sans clé dans le pod, RuntimeClass imposable, mémoire. |
| D9 | Outils : un endpoint `/mcp` sur la gateway, devant le serveur MCP de la CMP (stake policy) et des serveurs tiers lancés par **ToolHive**, un conteneur par outil. Registre des serveurs approuvés en JSON dans Git, avec leur provenance Sigstore. | hérité | 2 | |
| D10 | Guardrails : `promptGuard` d'agentgateway (regex qui masque PII et secrets), puis un webhook maison qui demande à Llama Guard 3, sur le serveur GPU, de classer le prompt. Si le classifieur est injoignable, la requête est refusée. | hérité ; refus en cas de panne décidé | 2 | Le webhook de la gateway ne parle pas le format d'Ollama : l'adaptateur est à écrire. |
| D11 | Budget en tokens **par agent et par jour**, appliqué par le rate limit de la gateway (clé : un claim du JWT). Dépassement → 429. Coût en euros = tokens × tarif payant publié de Gemini, même au niveau gratuit (décidé, voir [06-finops-ia.md](06-finops-ia.md)) ; alerte à 80 %. OpenCost pour le coût de l'infra. | hérité | 1 | |
| D12 | Évals : promptfoo dans GitHub Actions à chaque release d'agent ; rien sur le cluster. | hérité | 2 | Licence MIT ; rachat par OpenAI annoncé en mars 2026, à surveiller. |
| D13 | Panne du serveur GPU : réponse « service indisponible ». Gemini continue pour les projets Cloud. | décidé | 1 | |
| D14 | Audit : logs d'accès de la gateway → Alloy → Loki, puis l'archive Object Lock existante (CronJob horaire, bucket RGW, 30 jours). Champs obligatoires : utilisateur, agent, projet, modèle, outil, tokens, décision. Rétention de 30 jours pour le namespace de la gateway. | hérité | 1 | |
| D15 | Plafond d'étapes : limite d'itérations du framework (par exemple `recursion_limit` de LangGraph), posée par le kit d'agent. Le budget en tokens sert de plafond dur. | hérité | 2 | Rien ne l'impose côté plateforme pour un agent au code maison. |
| D16 | Pas de SPIRE ni de SPIFFE. | hérité | — | À reconsidérer si le mTLS entre agents ou la fédération multicloud deviennent nécessaires. |
| D17 | Emplacement du code : les briques de plateforme dans le dépôt **K3s**, sous `k8s/config/<brique>/` (`ai-gateway`, `kagent`, `toolhive`, `agent-sandbox`), Tempo dans `k8s/config/observability/`, les ClusterPolicies dans `k8s/config/kyverno/policies/`. Les ressources par projet et par agent dans le `ProjectRecord` de **cnp-projects**, rendues par **cnp-project-base**. Le code de chaque agent dans son propre dépôt, créé depuis **app-templates**. Les écrans et l'API dans **CMP**. | proposé | 1 | Suit l'organisation actuelle de chaque dépôt. |
| D18 | La CMP déclare les agents, elle ne les construit pas : pas d'éditeur visuel de workflow, le code vit dans le dépôt de l'agent. | proposé | 2 | OpenAI arrête son Agent Builder visuel le 30 novembre 2026. Détails dans [04-cmp-agents.md](04-cmp-agents.md). |

## 4. Conventions et interfaces

Ces noms viennent du modèle Noodle. Les noms de CRD et de champs propres à agentgateway sont
à confirmer pendant l'essai de la phase 1.

### Identité

| Sujet | Convention |
|---|---|
| Realm Keycloak | `3istor` |
| Client d'un agent | `agent-<projet>-<nom>`, confidentiel, avec service account ; membre du groupe `project-<projet>-agents`. Créé par la CMP quand un agent est déclaré. |
| Client de la gateway | `ai-gateway`, sert uniquement d'audience |
| Claims du jeton | `iss` = `realms/3istor` ; `aud` = `ai-gateway` ; `sub` ; `azp` ; `project` (issu du groupe) ; `scope` ; `exp` = émission + 5 min |
| Après un token exchange | `sub` = l'utilisateur, `azp` = le client de l'agent ; les scopes ne s'élargissent jamais. Exchange autorisé seulement d'un client agent vers `ai-gateway` (Client Policy `exchange-guard`). |
| Secret client (phase 1) | Vault `kvv2/projects/<projet>/agents/<nom>`, synchronisé en Secret `agent-oidc` |
| Client d'une app avec accès IA | `app-<projet>-<app>`, confidentiel ; secret dans Vault `kvv2/projects/<projet>/apps/<app>/ai-oidc` ; même jeton de 5 minutes et même audience `ai-gateway` |
| Audiences | `ai-gateway` (gateway), `cmp-api` (API de la CMP), `cmp-mcp` (serveur MCP de la CMP) ; un jeton ne vaut que pour son destinataire |

### Réseau et namespaces

| Sujet | Convention |
|---|---|
| Namespaces | `ai-gateway`, `kagent`, `mcp-tools`, un namespace par agent (§9, point 1) |
| Entrée vers un agent | Utilisateur → Cloudflare → Envoy Gateway (login OIDC) → agentgateway (A2A, jeton de l'utilisateur) → agent. Le listener d'agentgateway n'est jamais exposé directement. |
| Egress de la gateway | `generativelanguage.googleapis.com:443` et l'IP du serveur GPU en `/32`, rien d'autre |
| Egress d'un agent | DNS et agentgateway ; aucun ingress, sauf depuis la gateway pour l'A2A |
| Egress d'un sandbox | DNS seulement ; ingress depuis le routeur de sandbox seulement |
| Clé Gemini | Dans Vault, synchronisée en Secret `gemini-api-key` dans le seul namespace `ai-gateway` |

### Exécution

| Sujet | Convention |
|---|---|
| Agents | Image libre (LangGraph, ADK…), runc et seccomp `RuntimeDefault`, PSS `restricted`, non-root, sans shell ni outil d'exécution de code |
| Sandbox | `SandboxTemplate` `python` (gVisor) et `python-kata` ; `SandboxWarmPool` de 1 pour `python` (D33) ; une `SandboxClaim` par tâche, avec durée de vie |
| Limites | `ResourceQuota` et `LimitRange` par namespace d'agent |
| Admission | ClusterPolicy Kyverno `sandbox-guard` : PSS restricted, pas de token de ServiceAccount monté inutilement, RuntimeClass imposée selon le niveau de l'agent, images signées seulement |

### Données et observabilité

| Sujet | Convention |
|---|---|
| Niveau de données | Chaque projet est `cloud` ou `local`. Cloud : Gemini puis serveur GPU en repli. Local : serveur GPU seulement. Les projets `critical` (classement du doc d'origine) sont toujours Local. |
| Télémétrie | OTLP (gRPC 4317, HTTP 4318) vers Alloy ; `k8sattributes` ajoute `namespace` et `project` ; `gen_ai.input.messages` et `gen_ai.output.messages` sont supprimés sauf opt-in du projet |
| Tableaux | Coût et budget par projet et par agent dans Grafana ; alerte à 80 % du budget, sur les 429 et sur les refus |
| Budget mémoire | Environ 4 Go pour toute la zone IA sur le cluster on-prem, à mesurer (§7) |

### Formats et standards

Toutes les briques parlent des formats publics. Les documents de référence :

| Format | À quoi il sert ici | Référence |
|---|---|---|
| API Chat Completions d'OpenAI | Format des appels de modèle à la gateway (`/v1/chat/completions`) | [Référence de l'API](https://platform.openai.com/docs/api-reference/chat) |
| Function calling d'OpenAI | Le champ `tools` de la requête et `tool_calls` de la réponse | [Guide](https://platform.openai.com/docs/guides/function-calling) |
| API Embeddings d'OpenAI | `/v1/embeddings` pour le RAG | [Référence](https://platform.openai.com/docs/api-reference/embeddings) |
| Compatibilité OpenAI de Gemini | Gemini servi au format OpenAI | [Doc Google](https://ai.google.dev/gemini-api/docs/openai) |
| Compatibilité OpenAI d'Ollama | Le serveur GPU servi au format OpenAI | [Doc Ollama](https://docs.ollama.com/api/openai-compatibility) |
| MCP (Model Context Protocol), version 2026-07-28 | Outils : `tools/list`, `tools/call`, JSON-RPC 2.0 | [Spécification](https://modelcontextprotocol.io/specification/latest), [outils](https://modelcontextprotocol.io/specification/2026-07-28/server/tools) |
| A2A (Agent2Agent), version 1.0.0 | Agent vers agent ; Agent Card | [Spécification](https://a2a-protocol.org/latest/specification/) |
| OpenTelemetry GenAI | Noms des attributs des traces et métriques de tokens (statut Development) | [Dépôt des conventions](https://github.com/open-telemetry/semantic-conventions-genai) |
| JWT, OAuth 2.0 Token Exchange | Jetons des agents, agir pour un utilisateur | [RFC 7519](https://datatracker.ietf.org/doc/html/rfc7519), [RFC 8693](https://datatracker.ietf.org/doc/html/rfc8693) |
| FOCUS | Lignes de coût du FinOps | [Spécification](https://focus.finops.org/focus-specification/) |

## 5. Contrat du serveur GPU

Le serveur GPU est un **élément externe**, sans configuration figée : Ollama, vLLM, ComfyUI
ou autre chose peuvent y tourner. La plateforme ne dépend que de ce contrat.

| # | Service | Interface attendue | Utilisé pour | Phase | Si le service est injoignable |
|---|---|---|---|---|---|
| 1 | Texte | API compatible OpenAI (`/v1/chat/completions`, streaming) | Repli quand Gemini répond 429 ; projets Local | 1 | « service indisponible » pour les projets Local ; les projets Cloud continuent sur Gemini |
| 2 | Embeddings | API compatible OpenAI (`/v1/embeddings`) | RAG d'UC7 | 1 | Le RAG est indisponible |
| 3 | Garde | Llama Guard 3 ou équivalent, appelé par le webhook de D10 | Guardrails | 2 | La requête est refusée |
| 4 | Image | Outil MCP `generate_image` qui appelle l'API du serveur | Agents qui génèrent des images | 2 | L'outil renvoie une erreur claire |

Exigences pour chaque service exposé :

- un reverse proxy avec authentification devant (jeton ou mTLS), car ComfyUI n'a pas
  d'authentification par défaut ;
- accès limité à l'IP de sortie du cluster ;
- aucune interface web accessible aux tenants ;
- HTTPS. WireGuard n'est pas nécessaire si le cluster et le serveur sont sur le même réseau
  local.

### Pile proposée pour le serveur GPU (D34)

Une proposition, pas une obligation : le contrat ci-dessus reste la seule interface. Le choix
privilégie la qualité du texte, utilisé presque tout le temps, et accepte d'attendre pour
l'image, rare. Les tailles sont celles publiées pour les modèles quantifiés, plus le cache de
contexte : **ordres de grandeur, à mesurer avec `nvidia-smi`** (V1.15).

| Rôle | Logiciel | Modèle proposé | VRAM environ | Chargé |
|---|---|---|---|---|
| Texte et appels d'outils | Ollama | `qwen3.5:9b` en **Q8_0** par défaut ; à départager avec `granite4.2:8b` et `gemma4:12b` (D37) | 11 Go + 1 à 2 Go de contexte | en permanence, sauf pendant une image |
| Garde | Ollama | `llama-guard3:1b` | 1 Go | en permanence |
| Embeddings | Ollama | `nomic-embed-text` | 0,3 Go | en permanence |
| Image | ComfyUI | un modèle qui tient sous 13 Go (famille SDXL, ou plus récent) | jusqu'à 13 Go | à la demande |

En temps normal : environ 13 à 14 Go sur 16 (texte, garde, embeddings). Réglages Ollama :
`OLLAMA_MAX_LOADED_MODELS=3`, `OLLAMA_NUM_PARALLEL=2`, `OLLAMA_KEEP_ALIVE=-1`,
`OLLAMA_CONTEXT_LENGTH=8192`, cache KV en `q8_0`.

**Changer de modèle pour une image.** La garde et les embeddings restent chargés (1,3 Go) :
les guardrails continuent de marcher pendant une image.

1. Le serveur MCP `generate_image` prend un verrou GPU sur le serveur.
2. Il demande à Ollama de décharger le modèle de texte (`keep_alive: 0`).
3. ComfyUI charge le modèle d'image, génère l'image, puis libère la mémoire.
4. Le serveur MCP rend le verrou. Le modèle de texte se recharge à la requête suivante.

Compter 10 à 30 secondes de changement dans chaque sens, à mesurer. Pendant le verrou, le
service de texte du serveur GPU se déclare indisponible à la gateway. Les projets Cloud passent
sur Gemini, les projets Local reçoivent « service indisponible, génération d'image en cours ».
Une seule image à la fois. **Ce que ça apprend pour la suite** : des latences réalistes de
changement de modèle, et la mesure de ce qu'un vrai GPU ou un fournisseur payant
éviterait.

**À essayer plus tard.** `gpt-oss:20b` (environ 14 Go) si la garde et les embeddings passent
sur le CPU du serveur : meilleur modèle agentique, mais plus aucune marge de VRAM.

**Raccordement au cluster.** Ollama et ComfyUI n'écoutent que sur `127.0.0.1`. Caddy, seul
exposé, sert HTTPS sur le réseau local (certificat de son autorité interne, distribuée au
cluster) et route `/ollama/` vers Ollama et `/image/` vers ComfyUI. Il exige un jeton
`Authorization: Bearer` (un par consommateur : gateway, webhook de garde, serveur MCP
d'image), stocké dans Vault, et n'accepte que les IP des nœuds du cluster. Le pare-feu de
l'hôte n'ouvre que le port 443 au sous-réseau du cluster. Côté cluster, la règle Cilium
`ai-egress` n'autorise que cette IP en `/32` et le port 443. Pas de WireGuard en réseau local.

## 6. Table de correspondance des codes

Le dossier emploie sept familles de codes. **Cette spec n'emploie qu'une échelle : les phases
0 à 3.** Les autres codes sont des alias.

| Phase de la spec | Cible du modèle | Phases du Panorama | Écarts | Cas d'usage | Zooms du doc d'origine |
|---|---|---|---|---|---|
| **0. Prérequis** | P0 | absente | G1, G8, G17, sauvegarde CloudNativePG | — | — |
| **1. Squelette** | IA-1 | phase 0, Fondations | — | UC7 | gateway, modèles, identité, observabilité |
| **2. Agents et outils** | IA-2 | phases 1 et 2, plus l'A2A de la phase 3 | — | UC3, UC11, UC21 | MCP, runtime, policy |
| **3. Maturité** | IA-3 | phase 3 | — | UC20, UC15, UC17, UC19 | policy (OPA) |

Notes :

- L'option B du modèle place l'A2A et les guardrails en IA-2, plus tôt que le Panorama.
  La spec suit le modèle.
- Les écarts P1 et P2 (signature d'images, scan, Tetragon, OpenCost…) se traitent en
  parallèle et ne bloquent aucune phase. Les cibles P3 et P4 concernent la ségrégation des
  accès et le multicloud.
- Les écarts G1 à G20 gardent leur numéro ; le modèle ne contient pas G11, G12 ni G14.
- UC1 à UC20 sont décrits dans le doc de synthèse, pas dans ce dépôt. **UC21** est proposé
  ici : un assistant façon Grok dans une app de tenant (le clone de Twitter de Brian). Une app
  peut appeler la gateway dès la phase 1 ; la version avec agents est en phase 2, avec un
  agent « lecteur » public sans droits et un agent qui agit.

## 7. Budget de ressources

La cible est d'environ 4 Go pour toute la zone IA. L'estimation d'origine (1,1 à 2,3 Go)
valait pour Agent Router : elle ne compte ni agentgateway, ni kagent, ni ToolHive, ni le warm
pool de sandbox.

Protocole, en phase 1 puis à chaque nouvelle brique :

1. Déployer la brique avec des `requests` et `limits` explicites.
2. La laisser tourner une semaine avec le trafic de démo.
3. Relever la mémoire réellement utilisée par pod (pic et médiane) dans VictoriaMetrics.
4. Consigner le résultat dans le tableau ci-dessous.

| Brique | Phase | Mémoire mesurée | Date |
|---|---|---|---|
| agentgateway (proxy et contrôleur) | 1 | à mesurer | |
| Rate limit (serveur et Redis) | 1 | à mesurer | |
| Tempo | 1 | à mesurer | |
| kagent | 2 | à mesurer | |
| ToolHive (opérateur et un serveur) | 2 | à mesurer | |
| agent-sandbox (contrôleur et warm pool de 1) | 2 | à mesurer | |
| **Total** | | | |

## 8. Plans B

Chaque brique s'introduit après un essai de 1 à 2 jours avec un critère de passage écrit. Les
versions sont figées ; Dependabot ou Renovate propose les mises à jour, qui sont lues avant
d'être acceptées.

| Brique | Risque | Plan B | Basculer si |
|---|---|---|---|
| agentgateway | Projet jeune, porté surtout par Solo.io | Agent Router (suffisant tant qu'il n'y a pas d'A2A ni de guardrails) ou LiteLLM | L'essai ne passe pas V1.3 à V1.8 |
| kagent | Projet CNCF Sandbox, consommation mémoire inconnue | Un Deployment simple avec un SDK A2A | `ModelConfig` ne peut pas pointer vers la gateway, ou la RuntimeClass n'est pas imposable |
| ToolHive | API récente | Un Deployment du serveur MCP derrière la gateway | Un serveur tiers ne démarre pas sous les contraintes de sécurité |
| agent-sandbox | API en `v1beta1` | RuntimeClass gVisor sur un Pod ou un Job | Le contrôleur et son warm pool ne tiennent pas dans le budget mémoire |
| Federated client authentication | Fonction récente de Keycloak 26.6 | Secret client synchronisé depuis Vault | Le jeton du ServiceAccount n'est pas accepté |
| Webhook Llama Guard | Service maison | `promptGuard` seul (regex), en signalant la limite | L'adaptateur ajoute une latence inacceptable |

## 9. Décisions par défaut (anciens points ouverts)

Pour que la phase 1 puisse être implémentée sans décision en cours de route, chaque ancien
point ouvert reçoit une valeur par défaut. Statut **proposé** : Brian les valide ou les change
en une passe ; l'agent d'implémentation les applique telles quelles.

| ID | Sujet | Décision par défaut | Phase |
|---|---|---|---|
| D19 | Namespaces d'agents | Un namespace par agent, nommé `agent-<projet>-<nom>`, avec le label `cnp.3istor.com/project: <projet>`. Il hérite ainsi du default-deny en entrée par projet que `cnp-project-base` pose déjà. | 2 |
| D20 | Route Local | Pas de claim `data_level`. La gateway choisit la route Local quand le claim `project` appartient à la liste des projets `local`, générée depuis les `ProjectRecord` (`spec.ai.dataLevel`). | 1 |
| D21 | Clé du budget | Le claim `azp` (le client de l'agent ou de l'app). Plafond dur par `azp` et par jour à la gateway ; pour le projet, une alerte Grafana sur la somme, sans plafond dur. | 1 |
| D22 | Emplacement des politiques de la gateway | Dans le namespace `ai-gateway`, rendues par un chart `charts/cnp-ai-policies` du dépôt K3s et un ApplicationSet plateforme `cnp-ai-policies` (une Application par `ProjectRecord` qui a un bloc `spec.ai`, AppProject plateforme). Un tenant ne peut pas les modifier. | 1 |
| D23 | Plusieurs clés Gemini | Toutes les clés dans le même groupe de backends, équilibrées par agentgateway (P2C, ce que décrit sa doc) ; sur 429, la requête est rejouée sur une autre clé, puis sur le serveur GPU quand toutes saturent. Clés dans Vault sous `kvv2/platform/ai/gemini/<n>`, synchronisées dans `ai-gateway`. | 1 |
| D24 | Rétention | Traces Tempo : 7 jours. Logs d'accès de la gateway : 30 jours (D14). Prompts et réponses : jamais stockés, sauf opt-in du projet. | 1 |
| D25 | Chat de test de la CMP | Ses tokens comptent dans le budget de l'agent ; pas de budget séparé. | 2 |
| D26 | Premiers templates d'agent | (1) agent kagent déclaratif vide (consignes, modèle, outils), (2) agent de questions sur une documentation (UC7 sous forme d'agent). L'app UC7 de la phase 1 n'est pas un template d'agent (D31). | 2 |
| D27 | Version de Keycloak | Le cluster est en 26.5.5. La phase 1 utilise un secret client depuis Vault ; la montée en 26.6 ou plus est la première tâche de la phase 2, avant la federated client authentication (D3). | 2 |
| D28 | Où tester | Il n'y a qu'un cluster (K3s on-prem). Chaque nouvelle brique est d'abord installée sur un cluster k3d local, sur la machine de l'agent d'implémentation, et ne passe sur le cluster principal qu'une fois ses vérifications réussies. | 1 |
| D29 | Comment s'écrit un agent | Par défaut, un agent **kagent déclaratif** : un CR `Agent` de type `Declarative` (consignes, `ModelConfig`, outils MCP), sans code ; la boucle est fournie par kagent. Pour une logique que le déclaratif ne sait pas exprimer, un agent **BYO** : une image écrite avec LangGraph (une bibliothèque Python qui tourne dans le conteneur de l'agent, quelques centaines de Mo), construite par la CI comme les apps. Dans les deux cas, le CR `Agent` sert à déployer l'agent. À ne pas confondre avec Langfuse, l'outil d'observabilité écarté en D7. | 2 |
| D30 | Tempo | Chart Grafana `tempo` en mode monolithique dans le namespace `observability`, stockage sur volume, 7 jours. | 1 |
| D32 | Cas limites du budget | Le budget est vérifié avant la requête (réservation d'une estimation) et compté après, avec l'usage réel. Une fois épuisé, toute requête reçoit 429 jusqu'au lendemain : pas de contournement. Cas limites acceptés : la dernière requête admise peut dépasser le reste (une réponse en streaming n'est pas coupée) ; des requêtes parallèles peuvent dépasser ensemble ; un prompt refusé par les guardrails a déjà passé le contrôle de budget (le rate limit s'exécute avant `promptGuard`). C'est le mécanisme natif d'agentgateway, comme des autres gateways LLM : la taille d'une réponse n'est connue qu'à la fin, donc on réserve une estimation à l'entrée puis on corrige avec l'usage réel. Bornes : `max_tokens` plafonné à 4 096 par requête et au plus 30 requêtes par minute par agent (limiteur local). Si le service de rate limit est injoignable, la gateway refuse (fail-closed). Jour compté en UTC. Ces trois comportements sont vérifiés à l'essai T1.1 ; si agentgateway ne sait pas en faire un, l'agent d'implémentation le signale. | 1 |
| D33 | Warm pool de sandbox | Un `SandboxWarmPool` de **1** sandbox gVisor (template `python`), limité à 512 Mi, pour éviter d'attendre le démarrage. Sa mémoire est mesurée (§7) ; s'il pèse trop, on le passe à 0 réplica sans changer l'architecture. Pas de warm pool pour Kata. | 2 |
| D34 | Pile du serveur GPU | Profil du §5 : un modèle de texte le plus fort possible en permanence (`qwen3.5:9b` en Q8_0), la garde et les embeddings à côté ; l'image chargée à la demande, en déchargeant le texte. | 1 (texte), 2 (image) |
| D35 | Limite d'images | Au plus 20 appels de `generate_image` par agent et par jour, compteur de la gateway ; valorisés au TCO du serveur GPU. | 2 |
| D36 | Où vit la configuration de l'IA | Deux niveaux : les droits et plafonds dans le `ProjectRecord` (écrit par la CMP), les réglages dans les `deploy/values-*.yaml` de l'app (bloc `ai`, bloc `agent`). Le second ne dépasse jamais le premier ; les règles de la gateway viennent du premier. Tout est accessible depuis un assistant par les outils MCP de la CMP ([04-cmp-agents.md](04-cmp-agents.md) §8). | 1 |
| D37 | Choix du modèle de texte | Le modèle est un réglage, pas une décision figée. Candidats (octobre 2026, tenant sur 16 Go avec la garde et les embeddings) : `qwen3.5:9b` en Q8_0, `granite4.2:8b` (septembre 2026, Apache 2.0, outils et JSON) en Q8_0, `gemma4:12b` en Q4 si son support des outils est confirmé. Les Qwen plus récents (3.6 et 3.8) n'existent qu'en 27B et plus, soit 17 Go et plus : trop gros. Choix par un comparatif sur nos propres cas (T1.14) : 20 appels d'outils d'UC7 et UC3, taux d'appels corrects, latence, VRAM. | 1 |
| D31 | Application UC7 | Une app fullstack du catalogue existant (FastAPI, React, PostgreSQL avec pgvector), nommée `cnp-docs-assistant`, dans un projet `cnp-ai` de niveau demo. Elle indexe cnp-docs et cnp-agent-kit et cite ses sources. | 1 |

**Restent hors du champ de l'agent d'implémentation :**

1. **Conditions de Gemini : risque accepté par Brian le 2026-10-25.** La démo est en France,
   avec le niveau gratuit et plusieurs comptes. À savoir : la France fait partie de l'EEE, et
   les conditions de l'API réservent le niveau payant aux API mises à disposition
   d'utilisateurs de l'EEE ; le contenu du niveau gratuit peut être relu par des humains. Le
   bandeau « n'envoyez pas de données personnelles » reste affiché sur la démo.
2. **Vues refaites pour agentgateway (PR #22, 2026-10-25).** `ai-gateway-internals`,
   `mcp-tools-internals`, `platform-detailed`, `ai-gateway`, `ai-gateway-call`,
   `platform-global`, `models-internals` et `genai-observability-internals` décrivent
   agentgateway d'après sa doc officielle : GatewayClass `agentgateway`, contrôleur (traducteur,
   serveur xDS, Deployer), proxy Rust, `AgentgatewayBackend` (`ai`, `mcp`, `a2a`),
   `AgentgatewayPolicy` (`traffic`, `backend`, `frontend`), `AgentgatewayParameters`,
   `HTTPRoute` ; ce que la doc ne nomme pas est écrit comme un concept. Le budget passe par le
   service de rate limit global (le limiteur local n'a pas d'unité au jour), `ob-prom` est
   VictoriaMetrics, le serveur GPU est l'élément externe du §5, joint en HTTPS sans WireGuard.
   **Reste à corriger dans les schémas** : les vues `platform-ai-*` (éléments `ov-*` et `hl-*`)
   disent encore « via WireGuard » et « une seule clé » pour Gemini ; le modèle doit aussi
   suivre D33 (pas de warm pool), D34 (Llama Guard 1B, réglages Ollama du profil A) et D23
   (clés équilibrées en P2C). **À confirmer à l'essai T1.1** : le format du corps du webhook de
   garde et les noms d'outils quand un backend MCP a plusieurs cibles (les autres points sont
   dans D23 et D32).
3. **`00-index.md`.** Sa table D1 à D7 est remplacée par le §3 de cette page une fois la spec
   validée.
4. **Côté CMP.** Les choix propres à la CMP sont dans [04-cmp-agents.md](04-cmp-agents.md).

## 10. Policies et conformité as code

Chaque règle est un fichier dans Git, relu par PR, testé en CI, appliqué par Argo CD ou
kustomize. Pas de règle cliquée dans une interface.

| Règle | Outil | Où dans Git | Test en CI | Phase |
|---|---|---|---|---|
| Pods restreints (PSS restricted), images taguées, requests et limits | Kyverno | K3s `k8s/config/kyverno/policies/` (existe, en Audit) | `kyverno test` avec un manifeste conforme et un non conforme par règle | 0 |
| Isolation réseau entre tenants et egress | Cilium | `cnp-project-base/templates/network-policy.yaml` (existe, en entrée seulement) | V0.1 et V0.2 | 0 |
| Protection des branches GitOps | GitHub | réglages des dépôts, vérifiés par script | V0.5 | 0 |
| Aucune clé de fournisseur hors de `ai-gateway` : pas de variable `*_API_KEY` connue (Gemini, OpenAI, Anthropic) ni de Secret de fournisseur dans un autre namespace | Kyverno, nouvelle `ai-no-provider-keys` | K3s `k8s/config/kyverno/policies/` | `kyverno test` | 1 |
| La gateway n'est jamais exposée : Services `ClusterIP` seulement dans `ai-gateway`, aucune route publique sauf l'entrée A2A | Kyverno, nouvelle `ai-gateway-internal` | idem | `kyverno test` | 1 |
| Autorisation, budgets, route Local | agentgateway (`AgentgatewayPolicy`, CEL) | K3s `charts/cnp-ai-policies` (D22) | `helm template` puis validation du schéma ; V1.3 à V1.11 | 1 |
| `ProjectRecord` valide (`spec.ai`, `spec.agents`) ; un projet critical n'est jamais Cloud | validateur de cnp-projects | `cnp-projects/schema.json`, `scripts/validate-registry.py` | `make validate` (existe) | 1 |
| Agents qui exécutent du code : RuntimeClass `gvisor` ou `kata` obligatoire ; pas de token de ServiceAccount monté sans besoin | Kyverno `sandbox-guard` | K3s `k8s/config/kyverno/policies/` | `kyverno test` | 2 |
| `ModelConfig` d'un agent : uniquement l'URL de la gateway, jamais celle d'un fournisseur | Kyverno, nouvelle `ai-modelconfig-gateway-only` | idem | `kyverno test` | 2 |
| Ce qu'un agent ou un assistant peut faire sur la plateforme | stake policy de la CMP | `CMP/backend/app/services/agents/policy.yaml` (existe) | tests du backend de la CMP | 2 |
| Rule of Two | CMP, à la création de l'agent | code de la CMP | test : un agent à trois risques est refusé (VC2.2) | 2 |
| Qualité et sécurité des réponses | promptfoo | dépôt de chaque agent | bloque la release sous le seuil | 2 |
| Autorisation par argument d'outil | OPA | à définir en phase 3 | `opa test` | 3 |

**Conformité.** Les preuves sont les logs d'audit (D14), archivés en Object Lock pendant
30 jours, et les traces. La CI vérifie que chaque règle ci-dessus a ses tests. Les preuves AI
Act restent hors périmètre de la démo.

## 11. État constaté du cluster (2026-10-25)

Relevé dans les dépôts K3s, cnp-project-base et CMP, sans accès au cluster. La tâche T0.1
vérifie ces points sur le cluster avant toute modification.

| Composant | Version ou état dans Git | Conséquence |
|---|---|---|
| Keycloak | 26.5.5 (chart codecentric `keycloakx` 7.2.2) | Token exchange standard disponible ; federated client authentication non (D27) |
| Cilium | 1.20.0 | |
| Envoy Gateway | 1.9.0 | Reste l'entrée des apps (D1) |
| Kyverno | chart 3.8.2 ; PSS restricted en **Audit** | G8 : passer en Enforce (T0.3) |
| Default-deny entre tenants | Politique en entrée par projet dans `cnp-project-base`, derrière l'option `networkIsolation` ; egress ouvert | G1 partiellement traité : activer partout et fermer l'egress (T0.2) |
| CloudNativePG | chart 0.29.0, plugin Barman Cloud 0.8.1 installé | Sauvegarde à configurer et tester par base d'app (T0.5) |
| VictoriaMetrics | chart single 0.48.0, vmalert 0.46.0 | Base des métriques GenAI (D7) |
| Alloy | chart 1.11.1 | Récepteur OTLP à ajouter (T1.3) |
| Grafana, Loki | charts 10.5.15 et 7.3.0 | |
| Tempo | absent | À installer (D30) |
| CMP | décode les JWT du realm sans vérifier l'audience | G7 : à fermer en phase 1 (T1.9) |
