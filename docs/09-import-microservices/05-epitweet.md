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

## Budget mémoire

Le cluster on-prem a environ 4 Gio de requests libres (deux nœuds, 20 Go et 8 Go, requests à
environ 85 %). Epitweet tient en réglant chaque composant :

| Composant | Request | Limite | Réglage |
| --- | --- | --- | --- |
| 6 services Quarkus | 6 × 192 Mio | 6 × 384 Mio | `-XX:MaxRAMPercentage=70`, SerialGC |
| MongoDB | 384 Mio | 768 Mio | cache WiredTiger 0,25 Go |
| Neo4j | 768 Mio | 1 Gio | heap 512 Mio, page cache 128 Mio |
| Elasticsearch | 1 Gio | 1,25 Gio | heap 512 Mio |
| Redis | 64 Mio | 128 Mio | |
| frontend nginx | 32 Mio | 64 Mio | |
| **Total** | **environ 3,4 Gio** | | |

C'est juste. Environ 4 Gio pour epitweet sont acceptés depuis le 13/11. L'aperçu d'import
affiche cette jauge avant l'application, et l'offhours endort le projet la nuit. Si ça ne passe pas, tout est prêt pour un projet sur cloud public, mais rien n'est
lancé sans l'accord de Brian.

## Ce qui reste à valider

- La recette fonctionnelle : créer un compte, poster, suivre, liker et chercher depuis l'URL
  publique, avec les arêtes qui apparaissent sur la carte du projet.
