# Plan d'implémentation

> **Statut : v1.1 du 2026-11-22** (v1.0 validée par Brian le 2026-10-25). Cette page sert de
> consignes à l'agent qui implémente (par exemple Claude Sonnet). Elle découpe les phases 0 et
> 1 en tâches, et liste les tâches de phase 2 déjà connues (§5) ; le reste des phases 2 et 3
> sera découpé par une passe de spec à leur entrée, pas par l'agent d'implémentation.

## 1. Règles pour l'agent d'implémentation

**Ordre de priorité des documents.** [02-spec.md](02-spec.md) prime, puis
[04-cmp-agents.md](04-cmp-agents.md) et [06-finops-ia.md](06-finops-ia.md), puis
[03-phases.md](03-phases.md), puis cette page. Une contradiction entre eux est une raison de
s'arrêter, pas de choisir.

**Ce que l'agent décide seul.** Le détail d'implémentation à l'intérieur d'une tâche : noms
de variables, structure du code, tests, valeurs de `requests` et `limits` dans le budget
mémoire, formulation des messages.

**Quand l'agent s'arrête et remonte le problème.** Dans chacun de ces cas, il ouvre la PR en
brouillon avec une section « Écart » qui décrit le problème, les options et sa recommandation,
puis il attend :

1. une décision de la spec est impossible à appliquer telle quelle (fonction absente, version
   incompatible, ressource qui n'existe pas) ;
2. une vérification V échoue encore après deux tentatives de correction ;
3. il faudrait un composant, un dépôt ou un service qui n'est pas dans la spec ;
4. une modification affaiblirait une protection (ouvrir un egress, monter un secret dans un
   agent, désactiver une policy, élargir un scope) ;
5. le budget mémoire de la zone IA dépasserait la cible d'environ 4 Go (non bloquant depuis
   le 2026-11-22 : l'agent le signale dans sa PR et continue tant que les 6 à 8 Go
   disponibles ne sont pas atteints) ;
6. une action toucherait des données d'un tenant ou la branche `main` d'un dépôt ;
7. une action imposerait un arrêt du cluster ou un redémarrage de VM (par exemple pour Kata,
   D42) : prévenir Brian avant.

**Règles de travail.**

- Une tâche = une branche et une PR par dépôt concerné, au format Conventional Commits, avec
  les sorties des vérifications V de la tâche dans la description.
- Versions exactes partout (charts, images, CRD). Chaque version choisie est consignée dans
  le tableau §4 de cette page dans la même PR.
- Aucun secret dans Git : Vault, puis le vault-secrets-operator déjà en place.
- Tout passe par Git et Argo CD ou kustomize, comme le reste de la plateforme ; pas de
  `kubectl apply` à la main sur le cluster principal, sauf pour les vérifications.
- D'abord un cluster k3d local, puis le cluster principal (D28).
- Chaque brique : mesurer sa mémoire et la consigner dans la spec §7.

## 2. Phase 0 : prérequis

| ID | Dépôt | Quoi | Fait quand |
|---|---|---|---|
| T0.1 | — (lecture) | État des lieux sur le cluster des quatre écarts : G1 (`networkIsolation` activé sur quels projets, egress), G8 (rapports d'audit Kyverno PSS, pods root), G17 (protection de `main` sur K3s, cnp-projects, cnp-project-base, app-templates, CMP), sauvegardes CloudNativePG (`ObjectStore`, `ScheduledBackup`). Rapport dans une PR sur cnp-docs, dans cette page. | Rapport fusionné |
| T0.2 | cnp-project-base | Activer `networkIsolation` par défaut. Ajouter l'egress default-deny par projet : DNS, Services du projet, `envoy-gateway-system`, `observability`, `ai-gateway`, plus les sorties déclarées. Migrer les projets existants. | V0.1, V0.2 |
| T0.3 | K3s, dépôt du chart générique | Corriger le chart générique pour rendre `securityContext` et probes. Passer `pod-security-restricted` de `Audit` à `Enforce` une fois les rapports d'audit vides. | V0.3, V0.4 |
| T0.4 | GitHub | Protection de `main`, revue obligatoire et CODEOWNERS sur les cinq dépôts GitOps. | V0.5 |
| T0.5 | cnp-project-base ou template de base | `ObjectStore` Barman Cloud vers Garage et `ScheduledBackup` quotidien pour chaque base d'app ; un test de restauration documenté. | V0.6 |

## 3. Phase 1 : squelette

| ID | Dépôt | Quoi | Fait quand |
|---|---|---|---|
| T1.1 | K3s (k3d local) | **Essai d'agentgateway**, 1 à 2 jours. Installer la dernière version stable, épinglée. Relever les noms exacts des ressources (GatewayClass, politique, backend Anthropic, rate limit, routes), la clé du rate limit par claim, la traduction du format OpenAI vers l'API Messages et le passage de `output_config.effort` et de l'adaptive thinking, la possibilité d'autoriser ou de retirer un outil serveur d'Anthropic (`web_search`) par agent (D41), le comportement si le service de rate limit est injoignable, la possibilité de plafonner `max_tokens`, et la mémoire (D32). Rapport : une PR sur cnp-docs qui complète la spec §4 et §5 avec ces noms et la §7 avec la mémoire. Décision de continuer ou de basculer sur le plan B (spec §8). | Rapport fusionné ; V1.3 à V1.8 passent sur le k3d local |
| T1.2 | K3s | `k8s/config/ai-gateway/` : namespace, chart épinglé, GatewayClass et Gateway avec un listener interne (8080, jamais exposé), CiliumNetworkPolicy `ai-egress` (`api.anthropic.com:443` seulement), VaultSecret de la clé Anthropic (`kvv2/platform/ai/anthropic`, D23), backend Anthropic (`claude-sonnet-5-5` et `claude-haiku-5-5`), masquage regex `promptGuard`, rate limit. Ajout à `k8s/kustomization.yaml`. Le plafond de dépenses de la console est réglé par Brian. | V1.2, V1.5, V1.10, V1.12 |
| T1.3 | K3s | Observabilité : Tempo (D30) dans `k8s/config/observability/` ; récepteur OTLP et pipeline GenAI dans `values-alloy.yaml` (`k8s.attributes`, suppression des `gen_ai.*.messages`) ; datasource Tempo et tableau « IA » dans Grafana ; règles vmalert : budget à 80 %, 429, refus. Logs d'accès de la gateway avec les champs de D14, rétention 30 jours, ajout à `audit-archive.yaml`. | V1.8, V1.9 |
| T1.4 | K3s | Keycloak, par le même mécanisme que la configuration actuelle du realm `3istor` : client `ai-gateway` (audience seule), client scope `project` (mapper groupes → `project`), mapper d'audience `ai-gateway`, Client Policy `exchange-guard`, durée de jeton de 5 minutes pour les clients d'agent et d'app IA. Audience `cmp-api` pour la CMP (préparation de T1.9). | V1.3, V1.4, V1.6 |
| T1.5 | K3s | Chart `charts/cnp-ai-policies` et ApplicationSet `cnp-ai-policies` (D22) : pour chaque `ProjectRecord` qui a `spec.ai`, les règles d'autorisation sur `project`, les modèles autorisés, les budgets par `azp`. | V1.4, V1.7 |
| T1.6 | cnp-projects | Schéma : bloc `spec.ai` (`dailyTokenBudget` ; `models` ; plus de `dataLevel`, D20) et bloc `ai` par app (`enabled`, `dailyTokenBudget`). Validateur et tests ; un projet `critical` ne peut pas avoir de bloc `spec.ai`. | `make validate` passe ; VC1.1, VC1.3, V1.11 |
| T1.7 | CMP (backend) | Routes `GET`/`PUT /api/projects/{projet}/ai`, `POST /api/projects/{projet}/ai/apps/{app}`, `GET /api/projects/{projet}/ai/usage`, `GET /api/ai/backends` (04 §7). Écrit le `ProjectRecord` par la GitHub App ; crée le client Keycloak de l'app ; écrit son secret dans Vault ; ajoute à son `values.yaml` l'URL de la gateway, le modèle et la référence du secret. Usage lu dans VictoriaMetrics. | VC1.1 à VC1.3 ; tests du backend |
| T1.8 | CMP (frontend) | Onglet « IA » du projet, version phase 1 (04 §5) : niveau de données, budget, modèles et état des backends, apps avec accès IA, bouton « Donner l'accès IA ». | Écran conforme à la maquette ; tests du frontend |
| T1.9 | CMP (backend) | Fermer G7 : `decode_platform_token` vérifie l'audience `cmp-api` ; les jetons d'assistant (opaques) ne changent pas. | Un jeton d'audience `ai-gateway` est refusé par l'API de la CMP |
| T1.10 | nouveau dépôt `cnp-docs-assistant` | App UC7 (D31) créée par la CMP depuis le template fullstack : index plein texte PostgreSQL de cnp-docs et cnp-agent-kit, outils de recherche appelés par le modèle (recherche agentique), réponses par la gateway avec sources citées, accès IA donné par T1.7. | V1.1, V1.13 |
| T1.13 | CMP, cnp-agent-kit, chart générique | Outils MCP `get_project_ai`, `set_project_ai`, `enable_app_ai`, `get_ai_usage` et action `ai.configure` dans la stake policy ; bloc `ai` des values dans le chart générique ; guide `cnp-ai` dans cnp-agent-kit (04 §8). | VC1.4, VC1.5 |
| T1.12 | K3s | Policies Kyverno `ai-no-provider-keys` et `ai-gateway-internal`, avec leurs tests `kyverno test` en CI (spec §10). | Les tests passent ; un pod avec `ANTHROPIC_API_KEY` hors de `ai-gateway` est refusé |
| T1.11 | cnp-docs | Rapport de phase : sorties de V1.1 à V1.14, mémoire mesurée, écarts. | Rapport fusionné |

**Ordre et parallélisme.** T1.1 d'abord. Puis T1.2, T1.3, T1.4 et T1.6 en parallèle. Puis
T1.5, T1.7, T1.9. Puis T1.8 et T1.10. T1.11 en dernier. T1.14 (comparatif des modèles
locaux) est supprimée le 2026-11-22 : il n'y a plus de serveur GPU.

## 4. Versions retenues

À remplir par l'agent d'implémentation, une ligne par brique installée.

| Brique | Version | Date | Tâche |
|---|---|---|---|
| agentgateway | | | T1.1 |
| Tempo | | | T1.3 |

## 5. Phases 2 et 3

Elles seront découpées en tâches par une passe de spec à leur entrée. Ce qui est déjà connu :
montée de Keycloak en 26.6 (D27), kagent, outils MCP et serveur MCP de la CMP avec les jetons
d'audience `cmp-mcp` (04 §9), guardrails (regex seule, D10), A2A, évals, écrans d'agents de la
CMP, templates d'agent (D26), UC3, UC11. Tâches ajoutées le 2026-11-22, à détailler à l'entrée
de la phase 2 :

| ID | Dépôt | Quoi | Fait quand |
|---|---|---|---|
| T2.1 | K3s (k3d local) | **Essai kmcp contre ToolHive** (D39), 1 à 2 jours : héberger le même serveur MCP tiers avec chacun, derrière `/mcp` de la gateway, sous PSS restricted. Relever la mémoire, le registre, la vérification des serveurs. Garder kmcp s'il couvre le besoin. | Choix consigné dans la spec (D39, §7) ; un seul des deux installé |
| T2.2 | K3s | agent-sandbox, `SandboxTemplate` `python` (gVisor) et warm pool de 1 (D33) ; CiliumNetworkPolicy du sandbox : DNS et route MCP de la gateway seulement. | V2.8 |
| T2.3 | infra-templates, cnp-agent-kit, K3s | **Code mode** (D38) : outil `run_python` (une `SandboxClaim` par tâche), module Python des outils MCP autorisés dans l'image du sandbox, token exchange vers un jeton restreint (outils en lecture, 5 minutes, Client Policy Keycloak), retour des fichiers produits en artefacts de tâche. Mécanisme d'injection du jeton et des artefacts : à vérifier à l'essai. | V2.8, V2.13 |
| T2.4 | OpenStack (pae-node-*), K3s | **Kata** (D42) : vérifier si `/dev/kvm` existe dans les VM K3s ; sinon préparer la virtualisation imbriquée sur les hôtes de calcul et le passthrough du CPU. **Prévenir Brian avant tout redémarrage dur de VM** (arrêt du cluster). Puis RuntimeClass `kata` et `SandboxTemplate` `python-kata`, sans warm pool. | V2.15 ; ou plan B du §8 de la spec |
| T2.5 | infra-templates, cnp-projects, CMP | Bloc `agents:` des values (D29) avec le champ `trigger` (D40), rendu en `Agent`, `ModelConfig`, ServiceAccount, NetworkPolicy et `RemoteMCPServer` ; validateur qui refuse un agent qui dépasse `spec.ai` ; assistant de la CMP qui écrit le même YAML par PR. | V2.1, V2.14 ; VC2.1, VC2.7 |
| T2.6 | EpiTweet, cnp-docs | **UC21** ([08-uc21-grok.md](08-uc21-grok.md)) : `epitweet-mcp` (outils publics en lecture et un outil d'écriture), `grok-bridge` (consommateur des événements `post.created` qui mentionnent @grok), agents `grok-reader` et `grok-actor` déclarés dans les values d'EpiTweet, scénario de démo. | V2.11 ; scénario de démo rejoué de bout en bout |

## 6. Prompt de lancement

À donner tel quel à l'agent d'implémentation, une fois ses accès prêts.

```text
Tu implémentes la landing zone IA de CNP. La spec est dans le dépôt 3-Istor/cnp-docs, dossier
docs/07-ai-landing-zone/ : lis dans l'ordre 02-spec.md, 05-plan-implementation.md,
03-phases.md, 07-flux.md, puis 04-cmp-agents.md et 06-finops-ia.md pour les tâches qui les
citent. La spec est validée : tu appliques ses décisions telles quelles.

Commence par la phase 0, tâche T0.1 (état des lieux sur le cluster), puis T0.2 à T0.5, puis
la phase 1 dans l'ordre du §3 de 05-plan-implementation.md.

Règles (05-plan-implementation.md §1) : une tâche = une branche et une PR par dépôt, au format
Conventional Commits, avec les sorties des vérifications V dans la description ; versions
exactes, consignées au §4 ; aucun secret dans Git ; d'abord un cluster k3d local, puis le
cluster principal ; tu ne fusionnes jamais toi-même, c'est Brian qui fusionne.

Arrête-toi et remonte le problème dans une PR en brouillon (section « Écart » : problème,
options, recommandation) dans les cas du §1 : décision inapplicable, vérification qui
échoue deux fois, composant non prévu, protection affaiblie, action sur des données de
tenant ou sur main, arrêt du cluster ou redémarrage de VM. Un dépassement de la cible
mémoire de 4 Go se signale dans la PR sans t'arrêter.

La clé Anthropic est dans Vault sous kvv2/platform/ai/anthropic : ne la lis jamais en clair.
Elle ne doit être montée que dans le namespace ai-gateway. Le plafond de dépenses de la
console Anthropic est réglé par Brian, pas par toi.

À la fin de chaque phase, écris le rapport de phase (T1.11 pour la phase 1) et attends la
validation de Brian avant la phase suivante.
```
