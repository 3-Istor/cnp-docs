# Phases et vérifications

> **Statut : brouillon v0.2 du 2026-10-25, à valider.** Les décisions (D1 à D16) et les
> conventions sont dans [02-spec.md](02-spec.md). Une phase se termine quand toutes ses
> vérifications passent. Les phases 0 et 1 sont détaillées ; les phases 2 et 3 le seront à
> leur entrée, avec les mêmes règles.

**Règles.**

- Une vérification répond **oui ou non**, se lance à la main ou en script, et laisse une trace
  (sortie de commande, capture ou lien de trace) jointe à la PR qui clôt la phase.
- Les valeurs entre chevrons (`<projet>`, `<agent>`) sont à remplacer. Les valeurs marquées
  *seuil proposé* sont des points de départ à valider.
- Une brique nouvelle s'introduit après un essai de 1 à 2 jours avec un critère de passage
  écrit (spec §8). Sa mémoire est mesurée et consignée (spec §7).

## Phase 0 : prérequis de la plateforme

**Objectif.** Fermer les quatre écarts qui bloquent l'arrivée d'agents, indépendamment de
l'IA : un agent compromis ne doit pas pouvoir joindre un autre tenant, tourner en root, ou
déployer partout par un push sur un dépôt GitOps.

**Livrables.**

- CiliumClusterwideNetworkPolicy default-deny par projet, egress default-deny ; Hubble pour
  vérifier.
- Kyverno en Enforce avec Pod Security Standards `restricted` ; chart générique corrigé pour
  rendre `securityContext` et les probes.
- Protection de `main`, CODEOWNERS et revue obligatoire sur K3s, cnp-projects,
  cnp-project-base, app-templates et CMP.
- Sauvegarde CloudNativePG (plugin Barman Cloud vers Garage) avec un test de restauration.

| ID | Écart | Vérification | Attendu |
|---|---|---|---|
| V0.1 | G1 | Depuis un pod du tenant A, ouvrir une connexion vers un Service du tenant B | Échec ; Hubble affiche un verdict `DROPPED` |
| V0.2 | G1 | Depuis un pod de tenant, joindre un hôte Internet non autorisé | Échec |
| V0.3 | G8 | Appliquer un pod root ou privilégié dans un namespace de tenant | Kyverno le refuse |
| V0.4 | G8 | Inspecter les pods rendus par le chart générique | `runAsNonRoot`, `securityContext` et probes présents, aucun pod en root |
| V0.5 | G17 | Sur chacun des cinq dépôts GitOps, tenter un push direct sur `main` | Refusé ; une revue est exigée |
| V0.6 | Sauvegarde | Lancer une sauvegarde, puis restaurer dans un cluster de test | Sauvegarde terminée ; mêmes données après restauration (comptage de lignes sur deux tables) |

**Sortie.** V0.1 à V0.6 passent.

## Phase 1 : squelette de bout en bout (IA-1)

**Objectif.** Prouver le chemin complet sur le cas le plus simple : répondre à des questions
sur la documentation de CNP (UC7), en passant par la gateway, avec identité, budget, trace et
audit. Pas de kagent, de sandbox, d'outils MCP ni de guardrails.

**Livrables.**

1. agentgateway (proxy et contrôleur) dans le namespace `ai-gateway`, listener interne, egress
   limité à Gemini et au serveur GPU.
2. Keycloak : client de l'agent UC7, client `ai-gateway`, mappers `project` et audience, Client
   Policy `exchange-guard`, jetons de 5 minutes ; secret client dans Vault.
3. Clé Gemini dans Vault, montée dans la seule gateway ; backend Gemini par défaut.
4. Backend du serveur GPU dès qu'il répond au contrat (spec §5, services 1 et 2) : repli sur
   429 et route Local.
5. Politique de la gateway : JWT obligatoire, règle CEL sur `project`, deny par défaut ;
   budget en tokens par agent et par jour.
6. Traces OpenTelemetry GenAI vers Alloy et Tempo, métriques dans VictoriaMetrics, tableau
   Grafana par projet et par agent, alerte à 80 % du budget.
7. Logs d'accès de la gateway dans Loki avec les champs de D14, namespace à 30 jours, ajouté à
   l'export de l'archive d'audit.
8. L'agent UC7 : une app du chart générique, dans un namespace verrouillé, avec un RAG
   pgvector sur CloudNativePG, déclenchée depuis la CMP.

| ID | Vérification | Comment | Attendu |
|---|---|---|---|
| V1.1 | Aucune clé de provider dans le pod de l'agent | `kubectl exec <agent> -- env` et volumes montés | Aucune clé Gemini ou de provider |
| V1.2 | Pas d'autre sortie que la gateway | Depuis le pod : `curl -m 5` vers `generativelanguage.googleapis.com`, puis vers `example.com` | Les deux échouent |
| V1.3 | Pas d'appel sans jeton | Appel à la gateway sans en-tête `Authorization` | 401 |
| V1.4 | Pas d'appel hors de son projet | Jeton valide d'un projet non autorisé | 403 |
| V1.5 | Appel nominal | Jeton `client_credentials`, requête de chat en streaming | 200, réponse en flux |
| V1.6 | Durée du jeton | Décoder le JWT ; rejouer un jeton expiré | `exp − iat` = 300 s ; jeton expiré → 401 |
| V1.7 | Budget par agent | Abaisser le budget de l'agent de test, puis le dépasser | 429 au dépassement ; l'appel repasse le lendemain ; un autre agent n'est pas touché |
| V1.8 | Trace | Une requête, puis la trace dans Grafana | Attributs `project`, agent, modèle, tokens en entrée et en sortie, latence ; aucun prompt |
| V1.9 | Audit | Une requête refusée et une acceptée, puis LogQL sur Loki | Les deux lignes ont utilisateur, agent, projet, modèle, tokens et décision |
| V1.10 | Repli | Forcer un 429 de Gemini (quota simulé ou clé invalide) | L'appel bascule sur le serveur GPU et aboutit |
| V1.11 | Projet Local | Projet `local`, egress vers Gemini coupé | La requête aboutit par le serveur GPU ; aucune connexion vers Gemini |
| V1.12 | Panne du serveur GPU | Arrêter le service texte, puis requête d'un projet Local, puis d'un projet Cloud | Local : « service indisponible » en moins de 10 s *(seuil proposé)* ; Cloud : réponse normale |
| V1.13 | Qualité d'UC7 | 10 questions de référence sur cnp-docs | Au moins 8 réponses correctes, avec la bonne source citée *(seuil proposé)* |
| V1.14 | Mémoire | Une semaine de trafic de démo | Résultats consignés dans la spec §7 et comparés aux 4 Go |

**Sortie.** V1.1 à V1.9, V1.13 et V1.14 passent. V1.10 à V1.12 passent dès que le serveur
GPU répond au contrat : elles ne bloquent pas la phase 2, mais aucune route Local n'est
ouverte avant.

## Phase 2 : agents et outils (IA-2)

**Objectif.** Le golden path : un développeur déclare un agent dans la CMP, et la plateforme
fournit l'identité, l'isolation, les outils, les guardrails et les évals, sans ticket à
l'équipe plateforme.

**Contenu**, une brique à la fois, chacune après son essai :

1. **kagent** : CRD `Agent` (BYO) et `ModelConfig` pointé sur la gateway ; la CMP commit
   l'`Agent` dans Git et crée le client Keycloak ; l'agent s'authentifie par son jeton de
   ServiceAccount.
2. **Outils MCP** : `/mcp` sur la gateway, serveur MCP de la CMP, un serveur tiers par
   ToolHive (par exemple GitHub), registre signé dans Git, filtre d'outils par agent.
3. **Guardrails** : `promptGuard` (regex), puis le webhook Llama Guard si le serveur GPU
   répond au service 3 du contrat.
4. **A2A** : un utilisateur s'adresse à un agent par Envoy Gateway puis agentgateway ; un
   agent en délègue à un autre par la gateway.
5. **Sandbox** : agent-sandbox avec gVisor, et Kata pour le code non fiable, au premier cas qui
   exécute du code (UC9 ou UC13) ; Kyverno impose la RuntimeClass.
6. **Évals** : promptfoo dans GitHub Actions à chaque release d'agent.
7. **Outil `generate_image`** si le serveur GPU répond au service 4.
8. **Cas** : UC3 (diagnostic de déploiement), UC11 (revue de PR infra), puis UC21
   (assistant du clone de Twitter).

**Vérifications prévues**, à détailler à l'entrée de la phase :

| ID | Vérification | Attendu |
|---|---|---|
| V2.1 | Un agent déclaré dans la CMP | Il apparaît dans Git, Argo CD le déploie, son client Keycloak existe, son pod n'a aucun secret |
| V2.2 | Appel d'un outil de la CMP par la gateway | La stake policy du projet s'applique (allow, ask, PR, approve, deny) |
| V2.3 | Outil exclu par le filtre de l'agent | Il n'apparaît pas dans la liste vue par l'agent, et son appel est refusé |
| V2.4 | Serveur MCP absent du registre signé | ToolHive refuse de le lancer |
| V2.5 | Secret dans un prompt | Masqué avant d'atteindre le modèle |
| V2.6 | Webhook Llama Guard injoignable | La requête est refusée |
| V2.7 | Délégation A2A | La trace relie les deux agents ; l'agent délégué n'a pas plus de droits que l'appelant |
| V2.8 | Code exécuté dans le sandbox | Il tourne sous gVisor et n'a aucun accès réseau sauf le DNS |
| V2.9 | Agent qui exécute du code déclaré sur runc | Kyverno refuse le déploiement |
| V2.10 | Log de déploiement piégé (UC3) | L'agent n'exécute pas l'instruction injectée ; il n'a aucun droit de release |
| V2.11 | Tweet piégé (UC21) | L'agent lecteur n'a accès à aucune donnée privée ; l'instruction injectée reste sans effet |
| V2.12 | Release d'un agent | promptfoo tourne en CI et bloque la release sous le seuil fixé pour le cas |

## Phase 3 : maturité (IA-3)

**Objectif.** Les capacités avancées, une fois le reste stable.

**Contenu prévu.** OPA en `ext_authz` si les règles CEL ne suffisent plus (autorisation par
argument d'outil) ; bornes de délégation A2A en profondeur et en durée ; red teaming régulier
contre l'OWASP Agentic Top 10 ; modèle de menace par composant ; cas qui cumulent contenu non
fiable et données sensibles (UC15, UC17, UC19), coupés en deux agents ; UC20, gestion
d'incident multi-agents.

**Vérifications prévues**, à détailler à l'entrée de la phase :

| ID | Vérification | Attendu |
|---|---|---|
| V3.1 | Appel d'outil avec un argument interdit | Refusé par OPA, décision dans la trace et l'audit |
| V3.2 | Délégation au-delà de la profondeur ou de la durée autorisée | Refusée |
| V3.3 | Campagne de red teaming | Aucune exfiltration ; les écarts deviennent des tests de non-régression |
