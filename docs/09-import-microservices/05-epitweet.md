# Epitweet sur la CNP

Epitweet est la plateforme de démonstration choisie le 02/11 pour CNP-158 : un clone de Twitter
en microservices. Cette page dit ce qu'il contient, ce que la CNP doit ajouter pour le faire
tourner sans modifier son code, et ce que ça coûte en mémoire.

![Epitweet déployé dans un projet CNP](../01-architecture/diagrams/import-epitweet.light.png#only-light)
![Epitweet déployé dans un projet CNP](../01-architecture/diagrams/import-epitweet.dark.png#only-dark)

## Ce que contient le repo

Source : un repo GitHub, où est poussé le contenu de la branche `tooling/test-frontend` du
repo d'origine. La branche `main` d'origine n'a ni Dockerfile ni front. Le repo de
démonstration contient aussi `seed.sh`, `k8s/frontend.yaml` et `frontend/DEPLOY.md`.

| Composant | Techno | Port | Dépend de | Devient sur la CNP |
| --- | --- | --- | --- | --- |
| frontend | React + Vite, servi par nginx | 80 | les 6 services via `proxy_pass /api/*` | app `service`, seule app publique |
| user-service | Quarkus (Java 21) | 8081 | MongoDB, Redis | app `service` |
| post-service | Quarkus | 8082 | MongoDB, Redis, social-graph-service | app `service` |
| like-service | Quarkus | 8083 | MongoDB, Redis, social-graph-service | app `service` |
| social-graph-service | Quarkus | 8084 | Neo4j, MongoDB, Redis | app `service` |
| timeline-service | Quarkus | 8085 | MongoDB, Redis, social-graph-service, user-service | app `service` |
| search-service | Quarkus | 8086 | Elasticsearch, Redis | app `service` |
| MongoDB 7 | base documents | 27017 | | app `backing`, StatefulSet + volume Ceph |
| Redis 7 | pub/sub des événements | 6379 | | app `backing` |
| Neo4j 5 | graphe social | 7687 | | app `backing`, StatefulSet + volume Ceph |
| Elasticsearch 8 | recherche | 9200 | | app `backing`, StatefulSet + volume Ceph |

Les services échangent des événements par Redis pub/sub (`user.created`, `post.created`,
`like.like`…). Il n'y a ni Kafka ni RabbitMQ : aucun broker lourd à prévoir.

## Ce que la CNP ajoute, sans toucher au code

| Problème | Solution CNP | Pourquoi c'est mieux qu'une modification du code |
| --- | --- | --- |
| Le front appelle `user-service:8081`, la base `mongodb:27017` : noms courts du compose | Un lien `dependsOn` crée dans le namespace appelant un Service ExternalName qui porte le nom d'origine | Le code et la config nginx restent ceux du repo |
| Les Dockerfiles copient un `target/quarkus-app` déjà compilé, et chaque pom dépend du pom parent | Dockerfile réparé, généré côté plateforme : le reactor Maven est construit une fois pour produire les 6 images, JRE 21, utilisateur non root | Prototype : `post-service` construit en 35 s, tourne avec 67 Mio sous une limite de 256 Mio |
| Mots de passe en dur (`neo4j/password`, `admin:admin`) | Secrets générés dans Vault, injectés par VaultSecret, variables du consommateur réécrites | Aucun secret du repo ne part en production |
| Aucun service n'expose de route de santé (`/q/health` absent des poms) | Sondes TCP | Pas de redémarrages en boucle sur un 404 |
| `api-gateway` est décrit dans `k8s/` sans code source | Signalé dans le panneau « non supporté » de l'aperçu, écarté | L'aperçu dit honnêtement ce qui n'est pas importé |
| Elasticsearch demande `vm.max_map_count` | `node.store.allow_mmap=false` dans le preset | Pas de réglage noyau sur les nœuds |
| Kubernetes injecte `NEO4J_PORT_7687_TCP_PORT` et voisines pour le Service `neo4j`, que la validation stricte de Neo4j refuse | `NEO4J_server_config_strict__validation_enabled=false` dans le preset (CMP#227) | Le nom d'origine `neo4j` reste celui du Service |

## Budget mémoire

Valeurs déployées le 14/11 (cnp-projects `apps/epitweet/*/values.yaml`). Epitweet tient en
réglant chaque composant :

| Composant | Request | Limite | Réglage |
| --- | --- | --- | --- |
| 6 services Quarkus | 6 × 192 Mio | 6 × 384 Mio | `-XX:MaxRAMPercentage=70`, SerialGC |
| MongoDB | 256 Mio | 512 Mio | |
| Neo4j | 768 Mio | 1 Gio | heap 512 Mio, page cache 128 Mio |
| Elasticsearch | 1 Gio | 1,25 Gio | heap 512 Mio |
| Redis | 64 Mio | 256 Mio | |
| frontend nginx | 32 Mio | 64 Mio | |
| **Total** | **environ 3,2 Gio** | | |

C'est juste. Environ 4 Gio pour epitweet sont acceptés depuis le 13/11. Le 14/11, le cluster
réservait 95 % des requests mémoire du worker pour la moitié réellement utilisée, et
Elasticsearch ne trouvait pas de place. Un premier lot de requests a été aligné sur l'usage
mesuré (K3s#197), la suite est CNP-272. L'aperçu d'import
affiche cette jauge avant l'application, et l'offhours endort le projet la nuit. Si ça ne passe pas, tout est prêt pour un projet sur cloud public, mais rien n'est
lancé sans l'accord de Brian.

## Recette du 14/11

Epitweet est importé dans le projet `epitweet`, on-prem, et servi sur
`frontend-epitweet.3istor.com`. Vérifié de bout en bout depuis l'URL publique :

- comptes utilisateurs, stockés dans MongoDB ;
- abonnements, dans le graphe Neo4j ;
- posts et likes ;
- fil d'actualité, alimenté par les événements Redis ;
- recherche, sur Elasticsearch.

Ce que ce premier import a montré, et qui reste à faire :

- le front nginx démarrait avant ses services et le run passait en échec : ordre de
  déploiement à suivre les liens (CNP-269) ;
- les échecs de déploiement remontaient sans cause ni message (CNP-270) ;
- le premier groupe de build envoyait environ 1 Go d'images et a calé sur une coupure réseau
  (CNP-271) ;
- les requests du cluster bloquaient le déploiement (CNP-272).
