# La CMP et les agents

> **Statut : brouillon v0.1 du 2026-10-25, à valider.** Cette page dit comment un
> développeur ou un tenant active l'IA et déclare un agent dans la CMP, et ce que la CMP
> écrit en conséquence. Les décisions de plateforme sont dans [02-spec.md](02-spec.md), les
> phases dans [03-phases.md](03-phases.md). Les noms d'API et d'écrans sont des propositions.

## 1. Ce qui existe déjà

La landing zone IA s'appuie sur des mécanismes que la CMP a déjà ; elle les étend plutôt que
d'en créer de nouveaux.

| Existant | Où | Ce que les agents en réutilisent |
|---|---|---|
| Projets = groupes Keycloak du realm `3istor` | `keycloak_service.py` | Le groupe `project-<projet>-agents` et la création des clients d'agent |
| `ProjectRecord` (`cnp.3istor.com/v1alpha1`) avec `spec.apps` et `spec.features`, écrit par la GitHub App de la CMP | `cnp-projects/registry/projects/<projet>.yaml` | Un bloc `spec.ai` et une liste `spec.agents` |
| Chart `cnp-project-base`, rendu par les ApplicationSets d'Argo CD | `cnp-project-base/templates/` | Les ressources par agent (namespace, NetworkPolicy, quotas, `Agent`) |
| Apps créées depuis des templates (dépôt GitHub, CI, release par tag) | `app-templates`, `infra-templates` | Des templates d'agent ; un agent est une app avec un dépôt et des releases |
| Serveur MCP de la CMP, stake policy par niveau de projet (sandbox, demo, standard, critical) | `backend/app/mcp_server.py`, `services/agents/policy.yaml` | Les outils de plateforme et les règles par action |
| Confirmations dans la CMP (approuver ou refuser une action demandée par un assistant) | `routers/agents.py`, `ConfirmationPanel` | Les approbations des actions d'agents hébergés |
| Réglages assistants par projet : niveau, fenêtre autonome | `ProjectAgentPanel` | Le même panneau, étendu aux agents hébergés |
| FinOps : budgets, jauges, recommandations | `/finops`, `BudgetForm`, `BudgetGauge` | Le budget en tokens et le coût par agent |
| Journal d'activité, qui enregistre déjà l'assistant (`via`) | `/activity`, `audit.py` | L'agent comme auteur d'une action |

## 2. Ce que font les autres

Les plateformes du marché convergent vers les mêmes briques que la landing zone de CNP, ce qui
valide le design. Elles diffèrent surtout par l'expérience proposée.

| Plateforme | Ce qu'elle propose | Ce que CNP en retient |
|---|---|---|
| **AWS Bedrock AgentCore** | Runtime, Gateway (MCP, inférence avec repli de modèle, limites de débit), Identity (OAuth au nom de l'utilisateur, portail de consentement), Policy (avec Bedrock Guardrails), Memory, Observability, Evaluations (A/B, batch), Agent Registry ; outils intégrés : navigateur, interpréteur de code, recherche web | La recherche web comme **outil fourni par la plateforme** ; un **écran de consentement** quand un agent agit pour un utilisateur ; les évals comme étape de publication |
| **Google Gemini Enterprise Agent Platform** (avril 2026) | Agent Registry, Agent Gateway (identité et accès), policies, télémétrie, journaux d'audit, Agent Studio (low-code) | **Un registre unique** des agents et de leurs endpoints, consultable par tous |
| **Microsoft Foundry + Agent 365** | Inventaire de tous les agents (y compris ceux non déclarés), identité Entra par agent, activation et désactivation, propriétaire et attestation périodique, approbation d'un admin avant publication, tableaux d'activité | Un **propriétaire** par agent, un **bouton de coupure**, une **revue périodique** des agents à fort impact |
| **OpenAI AgentKit** | Agent Builder (canevas visuel, templates, aperçu, versions, nœuds de guardrails, évals intégrées), Connector Registry, ChatKit (interface de chat à intégrer) | Les **templates**, l'**aperçu avant publication**, les **versions**. Agent Builder et Evals s'arrêtent le 30 novembre 2026 au profit du SDK : **pas d'éditeur visuel de workflow** dans la CMP |
| **kagent** (déjà retenu, D8) | Interface : liste des agents avec chat, modèles, bibliothèque d'outils, serveurs MCP | Le **chat de test** directement sur la page de l'agent |

**Ce qu'on ne copie pas.** Pas d'éditeur visuel de workflow : le code de l'agent vit dans son
dépôt (LangGraph, ADK…), la CMP ne gère que sa déclaration. Pas de mémoire longue gérée par
la plateforme au départ : le RAG par projet suffit pour la démo. Pas de paiements.

## 3. Parcours

Chaque parcours indique sa phase (spec §6).

| # | Parcours | Qui | Phase |
|---|---|---|---|
| C1 | Activer l'IA sur un projet : niveau de données (Cloud ou Local), budget du projet, modèles autorisés | admin du projet | 1 |
| C2 | Donner l'accès aux modèles à une app existante (par exemple le clone de Twitter) | admin du projet | 1 |
| C3 | Créer un agent depuis un template | membre du projet | 2 |
| C4 | Choisir ses outils dans le registre et voir le profil de risque calculé | membre du projet | 2 |
| C5 | Tester l'agent dans un chat, avec le détail de chaque étape | membre du projet | 2 |
| C6 | Publier une version (tag de release, évals en CI, approbation si le projet est critical) | membre du projet | 2 |
| C7 | Suivre l'agent : tokens, coût, erreurs, refus, traces | membre du projet | 1 (app), 2 (agent) |
| C8 | Approuver ou refuser une action en attente | admin du projet | 2 |
| C9 | Suspendre un agent en un clic, puis le réactiver | admin du projet, admin plateforme | 2 |
| C10 | Supprimer un agent | admin du projet | 2 |
| C11 | Inventaire global des agents, registre des serveurs MCP, état des modèles | admin plateforme | 2 |
| C12 | Revue périodique : le propriétaire confirme qu'un agent sert encore et que ses droits sont justes | propriétaire | 3 |

## 4. Écrans

### Projet, nouvel onglet « IA »

- **Carte « Niveau de données ».** Cloud ou Local, avec une phrase sur ce que ça implique
  (« Local : vos données ne quittent pas le homelab ; les réponses sont plus lentes »). Un
  projet critical est forcé en Local.
- **Carte « Budget ».** Jauge des tokens consommés aujourd'hui, coût estimé en euros, seuil
  d'alerte. Reprend `BudgetGauge` et `BudgetForm`.
- **Carte « Modèles ».** Modèles autorisés pour le projet, et l'état de chaque backend
  (Gemini : quota restant ; serveur GPU : en ligne ou « service indisponible »).
- **Liste « Agents et apps avec accès IA ».** Nom, type (agent ou app), version, statut,
  isolation, tokens du jour, dernière activité.
- **Boutons.** « Créer un agent » ; « Donner l'accès IA à une app ».

### Assistant de création d'un agent

Sur le modèle du `DeploymentStepper` des apps, en six étapes :

1. **Template.** Agent vide (LangGraph), agent de questions sur une doc (RAG), agent de
   diagnostic de déploiement (UC3), agent de revue de PR (UC11). Le template crée le dépôt avec
   le code, le Dockerfile et `deploy/`.
2. **Identité.** Nom, description, propriétaire (par défaut le créateur).
3. **Modèle.** Choix parmi les modèles autorisés du projet, et plafond d'étapes par tâche.
4. **Outils.** Liste du registre, chaque outil avec ses badges de risque (« lit du contenu non
   fiable », « écrit », « sort vers Internet », « exécute du code »). Exemples : outils de la
   CMP, GitHub, recherche web, recherche dans la doc du projet, génération d'image.
5. **Risque.** Quatre questions, préremplies à partir des outils choisis : exécute-t-il du
   code ? lit-il du contenu non fiable ? accède-t-il à des données privées ? envoie-t-il vers
   l'extérieur ? La CMP affiche le niveau d'isolation qui en découle et le verdict de la
   Rule of Two. Trois risques cumulés bloquent la création, avec la proposition de couper en
   deux agents.
6. **Résumé.** Ce qui va être créé (dépôt, client, namespace, budget), puis « Créer ». Dans un
   projet critical, la création passe par une pull request ou une approbation, selon la stake
   policy.

### Page d'un agent

- **En-tête.** Nom, statut (actif, suspendu, en erreur), version, badges d'isolation et de
  risque, propriétaire.
- **Boutons.** « Tester », « Publier une version », « Suspendre » ou « Réactiver »,
  « Paramètres », « Supprimer ».
- **Onglet Aperçu.** Tokens du jour face au budget, coût, taux d'erreur, nombre de refus
  (budget, policy, guardrails).
- **Onglet Tester.** Un chat qui passe par la gateway, comme en production, avec le jeton de
  l'utilisateur ; à côté, le détail de chaque étape : appels de modèle, appels d'outils,
  décisions des guardrails, tokens. Les tokens consommés comptent dans le budget de l'agent.
- **Onglet Traces.** Les dernières tâches, avec un lien vers la trace dans Grafana.
- **Onglet Outils et accès.** Outils autorisés ; qui peut appeler l'agent (membres, groupes,
  autres agents en A2A) ; son Agent Card.
- **Onglet Évals.** Résultat du dernier passage de promptfoo, par version.
- **Onglet Journal.** Actions de l'agent, demandes d'approbation, changements de réglages.

### Administration de la plateforme

- **Agents.** Inventaire de tous les agents : projet, propriétaire, isolation, statut,
  dernière activité ; filtre et bouton de coupure.
- **Outils MCP.** Registre des serveurs : provenance, outils exposés, risque ; approuver ou
  retirer un serveur. L'approbation est une PR sur le registre dans Git.
- **Modèles.** Backends, quota Gemini restant, état du serveur GPU, table des prix.

### Écrans existants à étendre

- **FinOps.** Tokens et coût par projet et par agent.
- **Journal d'activité.** Les actions d'un agent hébergé, avec `via: agent-<projet>-<nom>`.
- **Confirmations.** Les demandes des agents hébergés, au même endroit que celles des
  assistants.

## 5. Ce que la CMP écrit

La CMP n'applique rien directement sur le cluster : elle écrit dans Git et dans Keycloak, et
Argo CD fait le reste, comme pour les apps.

| Objet | Où | Phase |
|---|---|---|
| Bloc `spec.ai` : niveau de données, budget du projet, modèles autorisés | `ProjectRecord` dans `cnp-projects` (schéma `schema.json` et validateur à étendre) | 1 |
| Liste `spec.agents[]` : nom, dépôt, template, isolation, outils, budget par jour, plafond d'étapes, appelants autorisés, statut | `ProjectRecord` | 2 |
| Client Keycloak `agent-<projet>-<nom>` dans le groupe `project-<projet>-agents` (ou un client pour une app, en phase 1) | Keycloak, realm `3istor`, par l'API d'administration | 1 |
| Secret client (phase 1 seulement) | Vault `kvv2/projects/<projet>/agents/<nom>` | 1 |
| Namespace, ServiceAccount, NetworkPolicy, `ResourceQuota`, `LimitRange` de l'agent | Rendus par `cnp-project-base` à partir de `spec.agents[]` | 2 |
| CR `Agent` de kagent et son `ModelConfig` | Rendus par `cnp-project-base` | 2 |
| Politique de la gateway pour l'agent : budget, outils autorisés, appelants A2A | À confirmer à l'essai d'agentgateway (namespace `ai-gateway` ou namespace de l'agent) | 1 (budget), 2 (outils, A2A) |
| Dépôt du code de l'agent, à partir d'un template | GitHub, template dans `app-templates` | 2 |

**Suspendre** (C9) : retirer le client Keycloak du groupe et passer `status: suspended`
dans le `ProjectRecord`, ce qui ramène l'agent à zéro réplica. Les jetons déjà émis expirent
en 5 minutes au plus.

## 6. API de la CMP

Routes proposées, sur le modèle des API existantes (`/api/projects/...`) :

| Méthode et route | Rôle | Phase |
|---|---|---|
| `GET`, `PUT /api/projects/{projet}/ai` | Lire ou changer le niveau de données, le budget et les modèles du projet | 1 |
| `POST /api/projects/{projet}/ai/apps/{app}` | Donner l'accès IA à une app (client Keycloak, URL de la gateway dans son `values.yaml`) | 1 |
| `GET`, `POST /api/projects/{projet}/agents` | Lister, créer un agent | 2 |
| `GET`, `PATCH`, `DELETE /api/projects/{projet}/agents/{agent}` | Lire, modifier, supprimer | 2 |
| `POST /api/projects/{projet}/agents/{agent}/suspend` et `/resume` | Couper, réactiver | 2 |
| `POST /api/projects/{projet}/agents/{agent}/test` | Chat de test, par la gateway, avec le jeton de l'utilisateur | 2 |
| `GET /api/projects/{projet}/agents/{agent}/usage` | Tokens, coût, erreurs, refus | 1 (app), 2 |
| `GET /api/admin/agents` | Inventaire global | 2 |
| `GET`, `POST /api/admin/mcp-servers` | Registre des serveurs MCP | 2 |
| `GET /api/ai/backends` | État des backends de modèles | 1 |

Les assistants (claude.ai, ChatGPT…) pourront créer et gérer des agents par de nouveaux outils
du serveur MCP de la CMP, soumis à la stake policy (nouvelles actions `agent.create`,
`agent.publish`, `agent.suspend`).

## 7. Le serveur MCP de la CMP face aux agents hébergés

Deux changements sont nécessaires.

**Accepter les jetons des agents.** Le serveur MCP de la CMP accepte aujourd'hui ses propres
jetons opaques, émis par son OAuth aux assistants. Un agent hébergé arrive par la gateway avec
un JWT Keycloak. Proposition : le serveur MCP accepte aussi un JWT du realm `3istor` dont
l'audience est le serveur MCP de la CMP, obtenu par token exchange ; il lit `azp` pour
l'agent et `sub` pour l'utilisateur, puis applique la stake policy du projet. L'audience à
utiliser et qui fait l'échange (l'agent ou la gateway) sont à trancher (§9).

**Adapter la stake policy.** Pour un assistant, `ask` veut dire « demander à l'utilisateur
dans la conversation ». Un agent hébergé n'a souvent personne dans la conversation : `ask`
devient alors `confirm`, une approbation dans la CMP. Le journal enregistre l'agent comme
`via`.

## 8. Vérifications

| ID | Phase | Vérification | Attendu |
|---|---|---|---|
| VC1.1 | 1 | Activer l'IA sur un projet | `spec.ai` apparaît dans le `ProjectRecord` et passe le validateur |
| VC1.2 | 1 | Donner l'accès IA à une app | Le client Keycloak existe ; l'app appelle la gateway sans aucune clé de provider |
| VC1.3 | 1 | Passer un projet critical en Cloud | Refusé |
| VC2.1 | 2 | Créer un agent depuis un template | Dépôt créé, `spec.agents[]` mis à jour, Argo CD déploie l'agent, client Keycloak présent |
| VC2.2 | 2 | Déclarer un agent qui cumule les trois risques | Création bloquée, avec la proposition de deux agents |
| VC2.3 | 2 | Chat de test | La réponse passe par la gateway ; la trace est visible ; les tokens sont décomptés |
| VC2.4 | 2 | Suspendre un agent | Ses appels échouent en 5 minutes au plus ; il est à zéro réplica |
| VC2.5 | 2 | Un agent hébergé appelle un outil `ask` de la CMP | Une confirmation apparaît dans la CMP ; rien ne s'exécute avant l'approbation |
| VC2.6 | 2 | Action d'un agent | Le journal d'activité l'enregistre avec `via: agent-<projet>-<nom>` |

## 9. Points ouverts

1. **Utilisateurs d'une app de tenant.** Un projet qui active la fonctionnalité `keycloak` a
   son propre realm pour les utilisateurs de ses apps, alors que les clients d'agent sont dans
   le realm `3istor`. Pour
   UC21, un agent qui agit au nom d'un utilisateur du clone de Twitter devra échanger un jeton
   d'un autre realm. À trancher : échange entre realms, ou l'app appelle la gateway avec sa
   propre identité et transmet l'utilisateur en claim.
2. **Audience du serveur MCP de la CMP** et acteur du token exchange (§7).
3. **Emplacement des politiques par agent** de la gateway (§5), à fixer à l'essai.
4. **Chat de test** : faut-il un budget de test séparé du budget de production ?
5. **Templates d'agent** : lesquels livrer en premier (proposition : RAG sur une doc, puis
   diagnostic de déploiement).
