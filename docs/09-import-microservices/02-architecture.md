# Architecture

Le code reste chez l'utilisateur ; tout ce qui sert à déployer vit côté plateforme. Seule la CI de
build lit le repo source, en lecture seule, à un commit figé. Argo CD ne lit jamais le repo de
l'utilisateur. Tout ce qui est décrit ici est **prévu** et en cours d'implémentation, sauf mention contraire.

![Ce que l'import ajoute à la plateforme](../01-architecture/diagrams/import-overview.light.png#only-light)
![Ce que l'import ajoute à la plateforme](../01-architecture/diagrams/import-overview.dark.png#only-dark)

## Ce qui est ajouté, composant par composant

| Composant | Dépôt | Ce qui change | Ticket |
| --- | --- | --- | --- |
| App GitHub **CNP-Portal** | GitHub org 3-Istor | App existante. Permissions réduites à ce que la plateforme utilise, événement `push`. Installée par l'utilisateur sur les seuls repos qu'il choisit | CNP-241 |
| Jetons d'installation utilisateur | CMP | Toujours demandés avec `contents:read`, `metadata:read` et un seul repo | CNP-241 |
| Liaison des comptes | CMP | OAuth avec PKCE, state à usage unique, jetons chiffrés, preuve de l'installation | CNP-241 |
| Analyseur | CMP (module Python) | Lit le repo à un SHA figé et produit un plan d'import | CNP-237 |
| Tables d'import | CMP | `import_sources`, `import_services`, `import_builds`, unicité par projet | CNP-241 |
| Workflow de build central | extern-project-analyzer | Jobs `prepare`, `build`, `publish` ; images `ghcr.io/3-istor/imports/<projet>/<app>` par digest | CNP-238 |
| Écriture du registre | CMP | Un seul commit atomique par application ou par build, sous verrou | CNP-239 |
| Types `service` et `backing` | cnp-projects, K3s, infra-templates | Schéma, ApplicationSet, chart (`workload`, StatefulSet). Rendu identique pour les apps existantes, vérifié avant chaque merge | CNP-239 |
| Liens `dependsOn` | cnp-projects, infra-templates, cnp-project-base | Alias ExternalName dans le namespace appelant ; plus tard, règles réseau par lien | CNP-240 |
| Catalogue de bases | infra-templates | MongoDB, Redis, Neo4j, Elasticsearch épinglés par digest, non root, secrets Vault générés | CNP-239 |
| Métriques Hubble | K3s (Cilium, Alloy) | Flux par workload et verdict, dans VictoriaMetrics | CNP-242 |
| Topologie | CMP | API et onglet du projet, canevas React Flow | CNP-242 |
| Outils MCP | CMP, cnp-agent-kit | `analyze_repo`, `get_topology`, `import_repo`, `link_apps`, `unlink_apps` | CNP-243 |
| Rebuild | CMP | Webhook signé de CNP-Portal, rattrapage au réveil | CNP-244 |

## D'où la topologie tire ses données

![Sources de la vue topologie](../01-architecture/diagrams/topology-data.light.png#only-light)
![Sources de la vue topologie](../01-architecture/diagrams/topology-data.dark.png#only-dark)

| Donnée | Source | Fraîcheur |
| --- | --- | --- |
| Nœuds (apps, bases, exposition publique) | registre cnp-projects | à chaque commit |
| Santé, synchro | Argo CD | 30 s |
| Sommeil offhours | OffhoursSchedule du projet | à chaque bascule |
| Arêtes déclarées | `dependsOn` du registre | à chaque commit |
| Arêtes observées et refusées | métriques de flux Hubble (contexte workload, verdict) dans VictoriaMetrics | 1 min, avec historique |

Une arête a quatre états :

- déclarée et observée : trait plein ;
- déclarée sans trafic : pointillé gris ;
- observée sans être déclarée : ambre, avec un bouton « Déclarer ce lien » ;
- refusée par une politique : rouge.

Hubble Relay viendra ensuite pour un tiroir de flux en direct.

## Décisions

Le brouillon v1 a été confronté au code par une revue contradictoire. Chaque ligne indique ce qui a
été tranché et pourquoi.

| # | Décision | Alternative écartée | Raison |
| --- | --- | --- | --- |
| D1 | Valeurs des apps importées dans cnp-projects, écrites par le CMP. Argo lit seulement cnp-projects et infra-templates | Valeurs dans le repo de l'utilisateur (modèle de Joe) | Argo n'a besoin d'aucun accès aux repos utilisateurs. Pas de clé d'app GitHub dans Argo. L'utilisateur règle ses ressources depuis le CMP |
| D2 | Un commit Git Data API par application ou par build, sous verrou et avec reprise. Une app n'entre dans le registre qu'avec son premier digest | Un commit par fichier, comme aujourd'hui | Dix écritures parallèles donnent des conflits 409, et Argo générerait des apps sans valeurs |
| D3 | Une seule app, CNP-Portal, installée par l'utilisateur sur ses repos. Ses permissions sont réduites au besoin réel. Pour une installation utilisateur, le CMP restreint chaque jeton à `contents:read`, `metadata:read` et un repo (`POST /app/installations/{id}/access_tokens`) | Une app dédiée CNP-Import en lecture seule | Une seule app à gérer. Une app n'a qu'un jeu de permissions pour toutes ses installations : l'utilisateur accorde donc plus que ce qui est utilisé. Ce risque résiduel est accepté le 13/11 |
| D4 | Analyseur en Python, l'outil de l'équipe. Railpack seulement en CLI dans la CI, pour les dossiers sans Dockerfile | Réécriture en Go avec Railpack en bibliothèque | Railpack se trompe sur Quarkus (fast-jar, pom parent), il n'est pas en 1.0, et une nouvelle langue coûte à une équipe de six |
| D5 | Modèle interne en Pydantic, de même forme que Score (CNCF) | Adopter la chaîne d'outils Score | score-k8s produit des manifests bruts qui contournent notre chart. L'export `score.yaml` reste possible plus tard |
| D6 | Le chart infra-templates gagne un bloc `workload` et un mode StatefulSet. Diff vérifié vide sur les apps existantes | Un second chart | Un dossier `charts/` serait lu comme une dépendance Helm, et on perdrait offhours, gatus et le secret de registre |
| D7 | Alias ExternalName rendus par le chart de l'appelant ; règles réseau par lien dans la politique de projet, en entrée seulement | Une CiliumNetworkPolicy par app | L'AppProject interdit les objets Cilium aux tenants, et les règles Cilium s'additionnent : une règle par app ne restreindrait rien |
| D8 | Arêtes observées depuis les métriques Hubble dans VictoriaMetrics | Collecteur gRPC Hubble Relay dans le CMP | Relay n'est pas déployé. Le CMP dort de 01:00 à 07:00 et perdrait des flux. Les métriques gardent l'historique |
| D9 | Un credential de registre par projet. Terraform seulement pour le DNS des apps publiques et les secrets générés | Le module Terraform d'app pour chaque service | Onze exécutions par import, chacune créant un repo GitHub inutile |
| D10 | Jeton de CI émis dans le workflow, jamais en entrée ; callback authentifié par OIDC GitHub | Jeton passé en `workflow_dispatch` | Les entrées sont lisibles dans le run |
| D11 | Actions épinglées par SHA complet, Trivy installé avec vérification de hash | `trivy-action` par tag | 76 des 77 tags de `trivy-action` ont été réécrits par un attaquant le 19/03/2026 (CVE-2026-33634) |
| D12 | Webhook GitHub plus rattrapage au réveil | Webhook seul | GitHub ne rejoue pas un webhook perdu pendant que le CMP dort |
