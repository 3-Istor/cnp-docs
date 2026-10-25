# Plan d'implémentation

> **Statut : brouillon v0.1 du 2026-10-25, à valider.** Cette page sert de consignes à
> l'agent qui implémente (par exemple Claude Sonnet). Elle découpe les phases 0 et 1 en
> tâches ; les phases 2 et 3 seront découpées par une passe de spec à leur entrée, pas par
> l'agent d'implémentation.

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
5. le budget mémoire de la zone IA (environ 4 Go) serait dépassé ;
6. une action toucherait des données d'un tenant ou la branche `main` d'un dépôt.

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
| T1.1 | K3s (k3d local) | **Essai d'agentgateway**, 1 à 2 jours. Installer la dernière version stable, épinglée. Relever les noms exacts des ressources (GatewayClass, politique, backends LLM, rate limit, routes), la clé du rate limit par claim, le rejeu sur une autre clé Gemini puis le serveur GPU après un 429 (D23), le comportement si le service de rate limit est injoignable, la possibilité de plafonner `max_tokens` et le comptage d'une requête rejouée (une ou deux fois), et la mémoire (D32). Rapport : une PR sur cnp-docs qui complète la spec §4 avec ces noms et la §7 avec la mémoire. Décision de continuer ou de basculer sur le plan B (spec §8). | Rapport fusionné ; V1.3 à V1.8 passent sur le k3d local |
| T1.2 | K3s | `k8s/config/ai-gateway/` : namespace, chart épinglé, GatewayClass et Gateway avec un listener interne (8080, jamais exposé), CiliumNetworkPolicy `ai-egress` (D23 : Gemini, serveur GPU), VaultSecret des clés Gemini, backends Gemini dans l'ordre puis serveur GPU, rate limit. Ajout à `k8s/kustomization.yaml`. | V1.2, V1.5, V1.10 |
| T1.3 | K3s | Observabilité : Tempo (D30) dans `k8s/config/observability/` ; récepteur OTLP et pipeline GenAI dans `values-alloy.yaml` (`k8s.attributes`, suppression des `gen_ai.*.messages`) ; datasource Tempo et tableau « IA » dans Grafana ; règles vmalert : budget à 80 %, 429, refus. Logs d'accès de la gateway avec les champs de D14, rétention 30 jours, ajout à `audit-archive.yaml`. | V1.8, V1.9 |
| T1.4 | K3s | Keycloak, par le même mécanisme que la configuration actuelle du realm `3istor` : client `ai-gateway` (audience seule), client scope `project` (mapper groupes → `project`), mapper d'audience `ai-gateway`, Client Policy `exchange-guard`, durée de jeton de 5 minutes pour les clients d'agent et d'app IA. Audience `cmp-api` pour la CMP (préparation de T1.9). | V1.3, V1.4, V1.6 |
| T1.5 | K3s | Chart `charts/cnp-ai-policies` et ApplicationSet `cnp-ai-policies` (D22) : pour chaque `ProjectRecord` qui a `spec.ai`, les règles d'autorisation sur `project`, les budgets par `azp`, l'appartenance à la liste Local (D20). | V1.4, V1.7, V1.11 |
| T1.6 | cnp-projects | Schéma : bloc `spec.ai` (`dataLevel` : `cloud` ou `local` ; `dailyTokenBudget` ; `models`) et bloc `ai` par app (`enabled`, `dailyTokenBudget`). Validateur et tests ; un projet `critical` ne peut pas être `cloud`. | `make validate` passe ; VC1.1, VC1.3 |
| T1.7 | CMP (backend) | Routes `GET`/`PUT /api/projects/{projet}/ai`, `POST /api/projects/{projet}/ai/apps/{app}`, `GET /api/projects/{projet}/ai/usage`, `GET /api/ai/backends` (04 §7). Écrit le `ProjectRecord` par la GitHub App ; crée le client Keycloak de l'app ; écrit son secret dans Vault ; ajoute à son `values.yaml` l'URL de la gateway, le modèle et la référence du secret. Usage lu dans VictoriaMetrics. | VC1.1 à VC1.3 ; tests du backend |
| T1.8 | CMP (frontend) | Onglet « IA » du projet, version phase 1 (04 §5) : niveau de données, budget, modèles et état des backends, apps avec accès IA, bouton « Donner l'accès IA ». | Écran conforme à la maquette ; tests du frontend |
| T1.9 | CMP (backend) | Fermer G7 : `decode_platform_token` vérifie l'audience `cmp-api` ; les jetons d'assistant (opaques) ne changent pas. | Un jeton d'audience `ai-gateway` est refusé par l'API de la CMP |
| T1.10 | nouveau dépôt `cnp-docs-assistant` | App UC7 (D31) créée par la CMP depuis le template fullstack : indexation de cnp-docs et cnp-agent-kit dans pgvector, réponses par la gateway avec sources citées, accès IA donné par T1.7. Vérifier que l'image PostgreSQL de CloudNativePG fournit pgvector, sinon le signaler. | V1.1, V1.12, V1.13 |
| T1.13 | CMP, cnp-agent-kit, chart générique | Outils MCP `get_project_ai`, `set_project_ai`, `enable_app_ai`, `get_ai_usage` et action `ai.configure` dans la stake policy ; bloc `ai` des values dans le chart générique ; guide `cnp-ai` dans cnp-agent-kit (04 §8). | VC1.4, VC1.5 |
| T1.14 | serveur GPU, cnp-docs | Comparatif des modèles de texte quand le serveur GPU répond au contrat (D37) : résultat consigné dans la spec. | Modèle retenu consigné |
| T1.12 | K3s | Policies Kyverno `ai-no-provider-keys` et `ai-gateway-internal`, avec leurs tests `kyverno test` en CI (spec §10). | Les tests passent ; un pod avec `GEMINI_API_KEY` hors de `ai-gateway` est refusé |
| T1.11 | cnp-docs | Rapport de phase : sorties de V1.1 à V1.14, mémoire mesurée, écarts. | Rapport fusionné |

**Ordre et parallélisme.** T1.1 d'abord. Puis T1.2, T1.3, T1.4 et T1.6 en parallèle. Puis
T1.5, T1.7, T1.9. Puis T1.8 et T1.10. T1.11 en dernier.

## 4. Versions retenues

À remplir par l'agent d'implémentation, une ligne par brique installée.

| Brique | Version | Date | Tâche |
|---|---|---|---|
| agentgateway | | | T1.1 |
| Tempo | | | T1.3 |

## 5. Phases 2 et 3

Elles seront découpées en tâches par une passe de spec à leur entrée. Ce qui est déjà connu :
montée de Keycloak en 26.6 (D27), kagent, outils MCP et serveur MCP de la CMP avec les jetons
d'audience `cmp-mcp` (04 §9), guardrails, A2A, sandbox, évals, écrans d'agents de la CMP,
templates d'agent (D26), UC3, UC11, UC21.
