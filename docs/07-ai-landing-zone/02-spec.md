# Spec de référence de la landing zone IA

> **Statut : v1.1 du 2026-11-22.** v1.0 validée par Brian le 2026-10-25, révisée selon ses
> décisions du 2026-11-22 : un seul fournisseur (l'API Anthropic), plus de serveur GPU, code
> mode, agents déclarés depuis la CMP ou en YAML. Les décisions touchées portent la mention
> « révisé le 2026-11-22 ». Rien n'est déployé. Cette page est la
> source de vérité des décisions et des conventions : en cas de contradiction avec
> `00-index.md`, une vue ou un schéma, c'est elle qui prime. Les phases et leurs vérifications
> sont dans [03-phases.md](03-phases.md).

**Sources.** La page [Stack cible et compromis](01-stack-cible.md) ; le modèle Noodle
(`architecture/cnp/model.yaml`), en particulier les éléments `ov-*` et `hl-*` de l'option B
(agentgateway et kagent), qui sont la version la plus récente ; le doc de synthèse « Landing
zone IA de CNP » ; les décisions de Brian du 2026-10-25 et du 2026-11-22. Branche lue :
`feat/ai-landing-zone-diagrams`, commit `eb2bfa5`.

## 1. Objet et périmètre

CNP veut laisser des agents IA travailler sur la plateforme sans qu'ils puissent faire de
dégâts. La landing zone IA est la plateforme actuelle plus une couche de contrôle qui traite
chaque agent comme un workload non fiable.

**Cadre : démonstration ou MVP.** La taille de l'équipe et la haute disponibilité ne sont pas
des contraintes. Les premiers utilisateurs sont les développeurs de CNP et les tenants.
L'API Anthropic est ouverte à tous ceux qui ont accès à la démo, pour un faible volume, sur un
crédit d'environ 10 $ que Brian charge sur la clé.

| Dans le périmètre | Hors périmètre |
|---|---|
| Gateway IA, identité des agents, isolation, outils MCP, A2A, code mode, guardrails, observabilité GenAI, budgets, audit | Modèles locaux, GPU, génération d'images, embeddings : abandonnés le 2026-11-22 (§5) |
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
7. **Standards d'abord.** API compatible OpenAI ou API Messages d'Anthropic, MCP, A2A,
   OpenTelemetry GenAI : les briques peuvent changer, les protocoles restent.
8. **Tout vient de Git.** Agents, routes, policies et serveurs MCP sont déclarés dans Git et
   appliqués par Argo CD, comme le reste de la plateforme.

## 3. Registre des décisions

Statuts : **décidé** (par Brian), **hérité** (des pages et du modèle, non rediscuté),
**proposé** (par Claude, puis validé par Brian le 2026-10-25 : il n'en reste plus). D1 à D7 gardent la numérotation de
`00-index.md`. Une décision révisée garde son numéro ; une décision abandonnée reste dans le
registre, marquée « sans objet ». Les nouvelles décisions continuent la numérotation (D38 et
suivantes, §9).

| ID | Décision | Statut | Phase | Note |
|---|---|---|---|---|
| D1 | **agentgateway** pour tout le trafic des agents (LLM, MCP, A2A), en GatewayClass `agentgateway` à part. Envoy Gateway reste l'entrée des apps et des utilisateurs. | hérité | 1 | Remplace Agent Router. Plan B au §8. |
| D2 | Tout appel LLM, outil ou agent passe par la gateway ; la clé Anthropic n'est montée que dans la gateway (D23). *Révisé le 2026-11-22.* | hérité | 1 | |
| D3 | Un client Keycloak par agent. Seul, l'agent utilise `client_credentials` ; pour un utilisateur, le token exchange (Keycloak 26.2 ou plus). Cible : le jeton du ServiceAccount remplace le secret client (federated client authentication, Keycloak 26.6). Jeton d'accès de **5 minutes, sans refresh**. | décidé | 1 puis 2 | Remplace « ≤ 15 min ». Phase 1 : secret client depuis Vault ; phase 2 : jeton du ServiceAccount. |
| D4 | Kyverno en Enforce à l'admission. Autorisation à la gateway par `AgentgatewayPolicy` : JWT obligatoire, règles CEL sur les claims et par outil, deny par défaut. OPA en `ext_authz` seulement si le CEL ne suffit plus. | hérité | 1 puis 3 | Cedar seulement si ToolHive est retenu (D39), confiné à son proxy. |
| D5 | Cilium default-deny entre tenants et en egress, egress par FQDN. RuntimeClass `gvisor` et `kata`. agent-sandbox en `v1beta1`. | hérité | 0 puis 2 | |
| D6 | **Un seul fournisseur : l'API Anthropic**, avec une seule clé (D23). Modèle par défaut `claude-sonnet-5-5` avec `output_config.effort: "medium"` et l'adaptive thinking ; `claude-haiku-5-5` autorisé comme modèle économique (agent lecteur, classification). Pas de Gemini, de serveur GPU, d'Ollama, de génération d'images ni d'embeddings locaux. Les prompts partent chez Anthropic : le masquage regex de la gateway (D10) reste actif. *Révisé le 2026-11-22.* | décidé | 1 | Remplace « Gemini par défaut, repli sur le serveur GPU ». Réserves au §9. |
| D7 | OpenTelemetry GenAI → Alloy → Tempo pour les traces ; métriques dans **VictoriaMetrics** ; Grafana ; pas de Langfuse (environ 21,5 Gio de RAM au minimum). | hérité | 1 | Arize Phoenix si une interface pensée pour les LLM devient nécessaire. |
| D8 | **kagent** : CRD `Agent` (`kagent.dev/v1alpha2`, type BYO pour une image LangGraph, ADK ou autre) et `ModelConfig` pointé sur la gateway. Chaque agent est publié en A2A avec son Agent Card. | hérité | 2 | À vérifier : `ModelConfig` sans clé dans le pod, RuntimeClass imposable, mémoire. |
| D9 | Outils : un endpoint `/mcp` sur la gateway, devant le serveur MCP de la CMP (stake policy) et des serveurs tiers hébergés par **kmcp ou ToolHive** (un seul des deux, D39), un conteneur par outil. Registre des serveurs approuvés en JSON dans Git, avec leur provenance Sigstore. *Révisé le 2026-11-22.* | hérité | 2 | |
| D10 | Guardrails : `promptGuard` d'agentgateway seulement (regex intégrée qui masque PII et secrets), plus la **défense par l'architecture** contre l'injection de prompt : Rule of Two, découpage lecteur et acteur, un seul outil d'écriture étroit. Pas de Llama Guard. Plus tard, en option : un webhook qui fait classer le prompt par `claude-haiku-5-5`, seulement si l'essai en montre le besoin. *Révisé le 2026-11-22.* | décidé | 2 | Plus de webhook maison ni de classifieur sur GPU. |
| D11 | Budget en tokens **par agent et par jour**, appliqué par le rate limit de la gateway (clé : un claim du JWT). Dépassement → 429. Coût en euros = tokens × tarif publié d'Anthropic, calculé dans Grafana depuis la table de prix (voir [06-finops-ia.md](06-finops-ia.md)) ; alerte à 80 %. Fusible financier dur : le plafond de dépenses de la console Anthropic (D23). OpenCost pour le coût de l'infra. *Révisé le 2026-11-22.* | hérité | 1 | |
| D12 | Évals : promptfoo dans GitHub Actions à chaque release d'agent ; rien sur le cluster. | hérité | 2 | Licence MIT ; rachat par OpenAI annoncé en mars 2026, à surveiller. |
| D13 | API Anthropic injoignable, ou plafond de dépenses de la console atteint : réponse « service indisponible », sans repli. *Révisé le 2026-11-22.* | décidé | 1 | Il n'y a plus de second backend. |
| D14 | Audit : logs d'accès de la gateway → Alloy → Loki, puis l'archive Object Lock existante (CronJob horaire, bucket RGW, 30 jours). Champs obligatoires : utilisateur, agent, projet, modèle, outil, tokens, décision. Rétention de 30 jours pour le namespace de la gateway. | hérité | 1 | |
| D15 | Plafond d'étapes : limite d'itérations du framework (par exemple `recursion_limit` de LangGraph), posée par le kit d'agent. Le budget en tokens sert de plafond dur. | hérité | 2 | Rien ne l'impose côté plateforme pour un agent au code maison. |
| D16 | Pas de SPIRE ni de SPIFFE. | hérité | — | À reconsidérer si le mTLS entre agents ou la fédération multicloud deviennent nécessaires. |
| D17 | Emplacement du code : les briques de plateforme dans le dépôt **K3s**, sous `k8s/config/<brique>/` (`ai-gateway`, `kagent`, `kmcp` ou `toolhive` selon D39, `agent-sandbox`), Tempo dans `k8s/config/observability/`, les ClusterPolicies dans `k8s/config/kyverno/policies/`. Les ressources par projet et par agent dans le `ProjectRecord` de **cnp-projects**, rendues par **cnp-project-base**. Le code de chaque agent dans son propre dépôt, créé depuis **app-templates**. Les écrans et l'API dans **CMP**. *Révisé le 2026-11-22.* | décidé | 1 | Suit l'organisation actuelle de chaque dépôt. |
| D18 | La CMP déclare les agents, elle ne les construit pas : pas d'éditeur visuel de workflow, le code vit dans le dépôt de l'agent. | décidé | 2 | OpenAI arrête son Agent Builder visuel le 30 novembre 2026. Détails dans [04-cmp-agents.md](04-cmp-agents.md). |

## 4. Conventions et interfaces

Ces noms viennent du modèle Noodle. Les noms de CRD et de champs propres à agentgateway, kagent,
kmcp, ToolHive et agent-sandbox sont à vérifier à l'essai T1.1 (ou à l'essai de la brique en
phase 2).

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
| Egress de la gateway | `api.anthropic.com:443`, rien d'autre |
| Egress d'un agent | DNS et agentgateway ; aucun ingress, sauf depuis la gateway pour l'A2A |
| Egress d'un sandbox | DNS et la route MCP d'`ai-gateway` seulement (code mode, D38) ; ingress depuis le routeur de sandbox seulement |
| Clé Anthropic | Dans Vault `kvv2/platform/ai/anthropic`, synchronisée en Secret `anthropic-api-key` dans le seul namespace `ai-gateway` (D23) |

### Exécution

| Sujet | Convention |
|---|---|
| Agents | Image libre (LangGraph, ADK…), runc et seccomp `RuntimeDefault`, PSS `restricted`, non-root, sans shell ni outil d'exécution de code |
| Sandbox | `SandboxTemplate` `python` (gVisor) et `python-kata` ; `SandboxWarmPool` de 1 pour `python` (D33) ; une `SandboxClaim` par tâche, avec durée de vie ; outil de plateforme `run_python` (D38) |
| Limites | `ResourceQuota` et `LimitRange` par namespace d'agent |
| Admission | ClusterPolicy Kyverno `sandbox-guard` : PSS restricted, pas de token de ServiceAccount monté inutilement, RuntimeClass imposée selon le niveau de l'agent, images signées seulement |

### Données et observabilité

| Sujet | Convention |
|---|---|
| Niveau de données | Plus de niveau Local (D20) : tout appel de modèle part chez Anthropic, après le masquage regex. Un projet `critical` (classement du doc d'origine) ne peut pas activer l'IA. |
| Télémétrie | OTLP (gRPC 4317, HTTP 4318) vers Alloy ; `k8sattributes` ajoute `namespace` et `project` ; `gen_ai.input.messages` et `gen_ai.output.messages` sont supprimés sauf opt-in du projet |
| Tableaux | Coût et budget par projet et par agent dans Grafana ; alerte à 80 % du budget, sur les 429 et sur les refus |
| Budget mémoire | 6 à 8 Go disponibles ; cible frugale d'environ 4 Go pour toute la zone IA, à mesurer (§7). Ce n'est plus un critère bloquant. |

### Formats et standards

Toutes les briques parlent des formats publics. Les documents de référence :

| Format | À quoi il sert ici | Référence |
|---|---|---|
| API Chat Completions d'OpenAI | Format des appels de modèle à la gateway (`/v1/chat/completions`), traduit vers Anthropic par la gateway : à vérifier à l'essai T1.1 | [Référence de l'API](https://platform.openai.com/docs/api-reference/chat) |
| Function calling d'OpenAI | Le champ `tools` de la requête et `tool_calls` de la réponse | [Guide](https://platform.openai.com/docs/guides/function-calling) |
| API Messages d'Anthropic | Format natif du backend Anthropic d'agentgateway (`/v1/messages`), outils serveur (`web_search_20260209`) | [Référence de l'API](https://docs.anthropic.com/en/api/messages) |
| MCP (Model Context Protocol), version 2026-07-28 | Outils : `tools/list`, `tools/call`, JSON-RPC 2.0 | [Spécification](https://modelcontextprotocol.io/specification/latest), [outils](https://modelcontextprotocol.io/specification/2026-07-28/server/tools) |
| A2A (Agent2Agent), version 1.0.0 | Agent vers agent ; Agent Card | [Spécification](https://a2a-protocol.org/latest/specification/) |
| OpenTelemetry GenAI | Noms des attributs des traces et métriques de tokens (statut Development) | [Dépôt des conventions](https://github.com/open-telemetry/semantic-conventions-genai) |
| JWT, OAuth 2.0 Token Exchange | Jetons des agents, agir pour un utilisateur | [RFC 7519](https://datatracker.ietf.org/doc/html/rfc7519), [RFC 8693](https://datatracker.ietf.org/doc/html/rfc8693) |
| FOCUS | Lignes de coût du FinOps | [Spécification](https://focus.finops.org/focus-specification/) |

## 5. Backend de modèles : l'API Anthropic

*Révisé le 2026-11-22.* Le serveur GPU, son contrat, sa pile (Ollama, Llama Guard, embeddings,
ComfyUI) et les tables de modèles locaux sont abandonnés. Il reste un backend externe.

| Sujet | Valeur |
|---|---|
| Fournisseur | API Anthropic, élément externe derrière la gateway, joint en HTTPS (`api.anthropic.com:443`) |
| Clé | Une seule, dans Vault `kvv2/platform/ai/anthropic`, synchronisée dans le seul namespace `ai-gateway`. Aucun pod hors d'`ai-gateway` ne la détient (Kyverno `ai-no-provider-keys`, §10). |
| Backend de la gateway | Le backend Anthropic d'agentgateway. Nom exact de la ressource, traduction du format OpenAI et passage des champs propres à Anthropic (`output_config.effort`, adaptive thinking, outils serveur) : à vérifier à l'essai T1.1. |
| Modèles autorisés | `claude-sonnet-5-5` (défaut, `effort: medium`, adaptive thinking) ; `claude-haiku-5-5` (économique) |
| Prix par million de tokens | Sonnet 5.5 : 2 $ en entrée, 10 $ en sortie, 0,10 $ en lecture de cache. Haiku 5.5 : 0,10 $ en entrée, 0,50 $ en sortie. Table versionnée dans [06-finops-ia.md](06-finops-ia.md). |
| Fusible financier | Plafond de dépenses sur la console Anthropic (environ 10 $ de crédit). La gateway garde les budgets en tokens par `azp` et par jour (D21, D32). |
| Données | Les prompts sortent chez Anthropic. Le masquage regex de la gateway reste actif ; un projet `critical` n'active pas l'IA (D20). |
| Panne ou plafond atteint | « service indisponible » (D13) |
| Outils serveur d'Anthropic | `web_search_20260209` autorisé agent par agent (D41). `code_execution` et le programmatic tool calling ne sont pas utilisés (D38). Savoir si la gateway peut autoriser ou retirer un outil serveur par agent : à vérifier à l'essai T1.1. |

**RAG sans embeddings.** Plus d'index vectoriel : l'agent cherche lui-même (recherche plein
texte PostgreSQL, outils de type grep) et cite ses sources. UC7 suit ce modèle (D31).

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
  ici : un assistant façon Grok dans EpiTweet, l'app de tenant de Brian. Il se construit avec
  les templates 4 et 3 (D26), sans rien de propre à EpiTweet dans la plateforme. Détails dans
  [08-uc21-grok.md](08-uc21-grok.md).

## 7. Budget de ressources

*Révisé le 2026-11-22.* 6 à 8 Go sont disponibles aujourd'hui. La cible reste frugale,
environ 4 Go pour toute la zone IA, mais ce n'est plus un critère bloquant : un dépassement se
signale et se discute, il n'arrête pas une brique. L'estimation d'origine (1,1 à 2,3 Go)
valait pour Agent Router : elle ne compte ni agentgateway, ni kagent, ni kmcp ou ToolHive, ni
le warm pool de sandbox.

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
| kmcp ou ToolHive (contrôleur et un serveur, D39) | 2 | à mesurer | |
| agent-sandbox (contrôleur et warm pool de 1) | 2 | à mesurer | |
| Sandbox Kata (une micro-VM par tâche, sans warm pool) | 2 | à mesurer | |
| **Total** | | | |

## 8. Plans B

Chaque brique s'introduit après un essai de 1 à 2 jours avec un critère de passage écrit. Les
versions sont figées ; Dependabot ou Renovate propose les mises à jour, qui sont lues avant
d'être acceptées.

| Brique | Risque | Plan B | Basculer si |
|---|---|---|---|
| agentgateway | Projet jeune, porté surtout par Solo.io | Agent Router (suffisant tant qu'il n'y a pas d'A2A ni de guardrails) ou LiteLLM | L'essai ne passe pas V1.3 à V1.8, ou le backend Anthropic ne marche pas |
| kagent | Projet CNCF Sandbox, consommation mémoire inconnue | Un Deployment simple avec un SDK A2A | `ModelConfig` ne peut pas pointer vers la gateway, ou la RuntimeClass n'est pas imposable |
| kmcp ou ToolHive (D39) | API récentes | Un Deployment du serveur MCP derrière la gateway | Un serveur tiers ne démarre pas sous les contraintes de sécurité |
| agent-sandbox | API en `v1beta1` | RuntimeClass gVisor sur un Pod ou un Job | Le contrôleur et son warm pool ne tiennent pas dans le budget mémoire |
| Federated client authentication | Fonction récente de Keycloak 26.6 | Secret client synchronisé depuis Vault | Le jeton du ServiceAccount n'est pas accepté |
| Kata (D42) | `/dev/kvm` absent des VM K3s | gVisor pour tous les niveaux, en signalant la limite | La virtualisation imbriquée ne peut pas être activée sans arrêt jugé inacceptable |

## 9. Décisions par défaut (anciens points ouverts)

Pour que la phase 1 puisse être implémentée sans décision en cours de route, chaque ancien
point ouvert reçoit une valeur par défaut. Validées par Brian le 2026-10-25, révisées et complétées le 2026-11-22 ; l'agent d'implémentation les applique telles quelles.

| ID | Sujet | Décision par défaut | Phase |
|---|---|---|---|
| D19 | Namespaces d'agents | Un namespace par agent, nommé `agent-<projet>-<nom>`, avec le label `cnp.3istor.com/project: <projet>`. Il hérite ainsi du default-deny en entrée par projet que `cnp-project-base` pose déjà. | 2 |
| D20 | Niveau de données | Plus de niveau Local ni de route Local : il n'y a plus de modèle local. Un projet `critical` ne peut pas activer l'IA du tout (au lieu d'être forcé en Local). Pas de claim `data_level` ; `spec.ai.dataLevel` disparaît du `ProjectRecord`. *Révisé le 2026-11-22.* | 1 |
| D21 | Clé du budget | Le claim `azp` (le client de l'agent ou de l'app). Plafond dur par `azp` et par jour à la gateway ; pour le projet, une alerte Grafana sur la somme, sans plafond dur. | 1 |
| D22 | Emplacement des politiques de la gateway | Dans le namespace `ai-gateway`, rendues par un chart `charts/cnp-ai-policies` du dépôt K3s et un ApplicationSet plateforme `cnp-ai-policies` (une Application par `ProjectRecord` qui a un bloc `spec.ai`, AppProject plateforme). Un tenant ne peut pas les modifier. | 1 |
| D23 | Clé du fournisseur | Une seule clé Anthropic dans Vault `kvv2/platform/ai/anthropic`, synchronisée en Secret dans le seul namespace `ai-gateway`, utilisée par le backend Anthropic d'agentgateway. Pas de groupe de clés ni de repli. Fusible dur : plafond de dépenses sur la console Anthropic. *Révisé le 2026-11-22.* | 1 |
| D24 | Rétention | Traces Tempo : 7 jours. Logs d'accès de la gateway : 30 jours (D14). Prompts et réponses : jamais stockés, sauf opt-in du projet. | 1 |
| D25 | Chat de test de la CMP | Ses tokens comptent dans le budget de l'agent ; pas de budget séparé. | 2 |
| D26 | Templates d'agent | Six templates génériques, qui couvrent la plupart des besoins réels : (1) **Assistant** déclaratif vide (consignes, modèle, outils du registre) ; (2) **Questions sur une doc** : recherche agentique dans un dépôt ou un bucket, sources citées (UC7 en agent) ; (3) **Analyste de données** en code mode : outils en lecture sur les données de l'app (MCP ou rôle CloudNativePG en lecture seule) et `run_python` ; (4) **Bot d'app** déclenché par événement, en paire lecteur et acteur (A2A) ; (5) **Diagnostic de déploiement** (UC3) : outils MCP de la CMP en lecture, propositions par les approbations de la CMP ; (6) **Revue de PR** (UC11) : lecture GitHub et commentaire. Détails dans [04-cmp-agents.md](04-cmp-agents.md) §5. L'app UC7 de la phase 1 n'est pas un template d'agent (D31). *Révisé le 2026-11-22.* | 2 |
| D27 | Version de Keycloak | Le cluster est en 26.5.5 (à la rédaction). La phase 1 utilise un secret client depuis Vault ; la montée en 26.6 ou plus est la première tâche de la phase 2, avant la federated client authentication (D3). | 2 |
| D28 | Où tester | Il n'y a qu'un cluster (K3s on-prem). Chaque nouvelle brique est d'abord installée sur un cluster k3d local, sur la machine de l'agent d'implémentation, et ne passe sur le cluster principal qu'une fois ses vérifications réussies. | 1 |
| D29 | Comment s'écrit et se déclare un agent | Par défaut, un agent **kagent déclaratif** : un CR `Agent` de type `Declarative` (consignes, `ModelConfig`, outils MCP), sans code ; la boucle est fournie par kagent. Pour une logique que le déclaratif ne sait pas exprimer, un agent **BYO** : une image écrite avec LangGraph, construite par la CI comme les apps. Un agent se déclare **depuis l'assistant de la CMP ou en YAML** dans le dépôt de l'app : un bloc `agents:` dans `deploy/values-<app>.yaml` (agent attaché à une app), ou une app autonome de type agent (`deploy/values-<agent>.yaml`). L'assistant de la CMP écrit ce même YAML par une PR sur le dépôt (comme la parité CNP-304) : la CMP et le YAML ne divergent jamais. infra-templates le rend en CR `Agent` de kagent, `ModelConfig` pointé sur la gateway, ServiceAccount, NetworkPolicy et références `RemoteMCPServer`. À ne pas confondre avec Langfuse, l'outil d'observabilité écarté en D7. *Révisé le 2026-11-22.* | 2 |
| D30 | Tempo | Chart Grafana `tempo` en mode monolithique dans le namespace `observability`, stockage sur volume, 7 jours. | 1 |
| D32 | Cas limites du budget | Le budget est vérifié avant la requête (réservation d'une estimation) et compté après, avec l'usage réel. Une fois épuisé, toute requête reçoit 429 jusqu'au lendemain : pas de contournement. Cas limites acceptés : la dernière requête admise peut dépasser le reste (une réponse en streaming n'est pas coupée) ; des requêtes parallèles peuvent dépasser ensemble ; un prompt refusé par les guardrails a déjà passé le contrôle de budget (le rate limit s'exécute avant `promptGuard`). C'est le mécanisme natif d'agentgateway, comme des autres gateways LLM : la taille d'une réponse n'est connue qu'à la fin, donc on réserve une estimation à l'entrée puis on corrige avec l'usage réel. Bornes : `max_tokens` plafonné à 4 096 par requête et au plus 30 requêtes par minute par agent (limiteur local). Si le service de rate limit est injoignable, la gateway refuse (fail-closed). Jour compté en UTC. Ces trois comportements sont vérifiés à l'essai T1.1 ; si agentgateway ne sait pas en faire un, l'agent d'implémentation le signale. | 1 |
| D33 | Warm pool de sandbox | Un `SandboxWarmPool` de **1** sandbox gVisor (template `python`), limité à 512 Mi, pour éviter d'attendre le démarrage. Sa mémoire est mesurée (§7) ; s'il pèse trop, on le passe à 0 réplica sans changer l'architecture. Pas de warm pool pour Kata. | 2 |
| D34 | Backend de modèles | Le profil du §5 : l'API Anthropic seule. Sans objet pour la pile du serveur GPU, abandonnée. *Révisé le 2026-11-22.* | 1 |
| D35 | Limite d'images | Sans objet : pas de génération d'images (pas de GPU). L'outil `generate_image` est retiré du registre. *Révisé le 2026-11-22.* | 2 |
| D36 | Où vit la configuration de l'IA | Deux niveaux : les droits et plafonds dans le `ProjectRecord`, `spec.ai` (écrit par la CMP) : modèles autorisés, budgets, outils autorisés, isolation minimale ; les réglages dans les `deploy/values-*.yaml` de l'app (bloc `ai`, bloc `agents:`, D29). Le second ne dépasse jamais le premier ; les règles de la gateway viennent du premier. La CMP écrit le second par PR, un développeur ou son assistant aussi. Tout est accessible depuis un assistant par les outils MCP de la CMP ([04-cmp-agents.md](04-cmp-agents.md) §8). *Révisé le 2026-11-22.* | 1 |
| D37 | Choix du modèle | Le modèle est un réglage par agent ou par app, parmi `spec.ai.models` : `claude-sonnet-5-5` par défaut (`effort: medium`, adaptive thinking), `claude-haiku-5-5` pour un agent lecteur ou une classification. Plus de comparatif de modèles locaux (T1.14 supprimée). *Révisé le 2026-11-22.* | 1 |
| D31 | Application UC7 | Une app fullstack du catalogue existant (FastAPI, React, PostgreSQL), nommée `cnp-docs-assistant`, dans un projet `cnp-ai` de niveau demo. Elle indexe cnp-docs et cnp-agent-kit en recherche plein texte PostgreSQL (pas de pgvector ni d'embeddings), le modèle cherche lui-même par outils (recherche agentique) et cite ses sources. *Révisé le 2026-11-22.* | 1 |

Décisions ajoutées le 2026-11-22 :

| ID | Sujet | Décision | Phase |
|---|---|---|---|
| D38 | Code mode | Le modèle écrit du code qui appelle ses outils et l'exécute, au lieu d'enchaîner des appels d'outils un par un. Outil de plateforme `run_python` : une `SandboxClaim` agent-sandbox par tâche, gVisor par défaut, Kata pour le niveau de risque élevé (D42). Egress du sandbox : DNS et la route MCP d'`ai-gateway` seulement. Dans le sandbox, les outils MCP autorisés de l'agent sont importables comme un module Python qui appelle la route MCP de la gateway avec un jeton court restreint (token exchange, outils en lecture seule, durée de vie de 5 minutes) : le code écrit par le modèle passe par la même policy par outil, le même budget et les mêmes traces qu'un appel direct. Les fichiers produits (PNG, CSV) reviennent comme artefacts de la tâche. Alternative écartée : `code_execution` et le programmatic tool calling côté Anthropic, parce que les données et les appels d'outils tourneraient dans le sandbox d'Anthropic, hors de notre policy. Injection du jeton dans le sandbox et artefacts de tâche par agent-sandbox ou kagent : à vérifier à l'essai. | 2 |
| D39 | Hébergement des serveurs MCP | Un seul des deux : **kmcp** de kagent (CRD `MCPServer`) ou **ToolHive**. L'essai tranche : kmcp s'il couvre le besoin, ToolHive seulement s'il apporte une fonction qu'on utilise (registre, vérification des serveurs). Capacités de chacun : à vérifier à l'essai T1.1 (phase 2). | 2 |
| D40 | Déclencheur d'un agent | Un champ `trigger`, pas un template : `chat` (chat de test de la CMP ou API), `event` (webhook ou stream Redis de l'app), `schedule` (cron), `a2a` (appelé par un autre agent). Rendu par infra-templates ; le support de chaque mode par kagent est à vérifier à l'essai. | 2 |
| D41 | Recherche web | Outil serveur d'Anthropic `web_search_20260209`, autorisé agent par agent, avec les badges « sort vers Internet » et « lit du contenu non fiable » dans la Rule of Two. Pas de serveur MCP de recherche à héberger. Si la gateway ne peut pas autoriser ou retirer un outil serveur par agent (essai T1.1), la recherche web reste coupée pour tous. | 2 |
| D42 | Kata | Voulu par Brian pour le niveau de risque élevé du code mode. Il faut `/dev/kvm` dans les VM K3s (invitées nova d'OpenStack sur les pae-node-*) : virtualisation imbriquée sur les hôtes de calcul et passthrough du CPU des VM. L'activer change la configuration de l'hyperviseur et impose un redémarrage dur des VM, donc un arrêt du cluster : vérifier d'abord, prévenir Brian avant tout redémarrage. Tant que Kata manque, le niveau élevé n'est pas proposé (plan B au §8). | 2 |

**Restent hors du champ de l'agent d'implémentation :**

1. **Données envoyées à Anthropic (révisé le 2026-11-22).** Les risques sur les conditions
   du niveau gratuit de Gemini (EEE, relecture humaine, comptes multiples) disparaissent avec
   Gemini. Les prompts partent toujours chez un fournisseur externe : le masquage regex reste
   actif, un projet `critical` n'active pas l'IA, et le bandeau « n'envoyez pas de données
   personnelles » reste affiché sur la démo.
2. **Vues refaites pour agentgateway (PR #22, 2026-10-25).** `ai-gateway-internals`,
   `mcp-tools-internals`, `platform-detailed`, `ai-gateway`, `ai-gateway-call`,
   `platform-global`, `models-internals` et `genai-observability-internals` décrivent
   agentgateway d'après sa doc officielle : GatewayClass `agentgateway`, contrôleur (traducteur,
   serveur xDS, Deployer), proxy Rust, `AgentgatewayBackend` (`ai`, `mcp`, `a2a`),
   `AgentgatewayPolicy` (`traffic`, `backend`, `frontend`), `AgentgatewayParameters`,
   `HTTPRoute` ; ce que la doc ne nomme pas est écrit comme un concept. Le budget passe par le
   service de rate limit global (le limiteur local n'a pas d'unité au jour), `ob-prom` est
   VictoriaMetrics. Les vues de sandbox montrent le warm pool de 1 réplica gVisor à 512 Mi,
   sans pool pour Kata (D33). `ai-gateway-internals` montre les bornes de D32 (`max_tokens` à
   4 096, 30 requêtes par minute et par agent, refus si le service de rate limit est
   injoignable) et l'origine des règles et réglages (D36). Les PNG trop grands
   (`platform-detailed`) sont exportés à l'échelle 1 et le script de build vérifie qu'ils sont
   complets. **Révisé le 2026-11-22** : les vues n'ont plus ni Gemini, ni serveur GPU, ni
   Ollama, ni Llama Guard, ni webhook de garde ; l'API Anthropic est un élément externe
   derrière la gateway, `models-internals` montre ce backend unique. **À vérifier à l'essai
   T1.1** : les noms d'outils quand un backend MCP a plusieurs cibles, et les trois bornes de
   D32 (dessinées d'après la spec, pas d'après un essai).
3. **`00-index.md`.** Fait le 2026-10-25 : l'index renvoie au registre de cette page.
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
| Autorisation, budgets | agentgateway (`AgentgatewayPolicy`, CEL) | K3s `charts/cnp-ai-policies` (D22) | `helm template` puis validation du schéma ; V1.3 à V1.9 | 1 |
| `ProjectRecord` valide (`spec.ai`, agents) ; un projet critical n'active jamais l'IA | validateur de cnp-projects | `cnp-projects/schema.json`, `scripts/validate-registry.py` | `make validate` (existe) | 1 |
| Agents qui exécutent du code (`run_python`, D38) : RuntimeClass `gvisor` ou `kata` obligatoire ; pas de token de ServiceAccount monté sans besoin ; egress limité au DNS et à la route MCP de la gateway | Kyverno `sandbox-guard`, Cilium | K3s `k8s/config/kyverno/policies/` | `kyverno test` | 2 |
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
| Keycloak | 26.5.5 (à la rédaction ; chart codecentric `keycloakx` 7.2.2) | Token exchange standard disponible ; federated client authentication non (D27) |
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
