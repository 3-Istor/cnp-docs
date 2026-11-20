# Providers : ce que la plateforme fournit

Un provider est un service de données que la plateforme fait tourner pour une app. L'app le
reçoit par des variables d'environnement, sans rien installer. Cette page dit ce qui est fourni
sur `main` le 20/11/2026, ce qui est recommandé et ce qui ne l'est pas. Les sources sont les
blocs `db` et `bucket` du chart infra-templates, le catalogue d'import du CMP et les règles de
l'analyseur.

![Ce qu'une app reçoit pour ses données](../01-architecture/diagrams/providers-catalog.light.png#only-light)
![Ce qu'une app reçoit pour ses données](../01-architecture/diagrams/providers-catalog.dark.png#only-dark)

## Ce qui est fourni

### PostgreSQL managé, CloudNativePG

- Un `Cluster` CNPG par app, PostgreSQL 18.6 épinglé par digest, une instance.
- Volume Ceph RBD quand Rook est présent. Les templates on-prem forcent encore `local-path`.
- Sauvegarde quotidienne à 03:00 vers Ceph RGW, rétention 7 jours. Sur AWS et GCP, elle passe
  par le relais VPN vers le RGW du site.
- Sauvegarde avant chaque migration si `migrations.backupBefore` est actif.
- Restauration depuis le CMP, à un instant donné ou depuis une sauvegarde.
- Pooler PgBouncer en option, mode `transaction` par défaut. Le mode transaction casse `SET`,
  `LISTEN`, les verrous consultatifs et les requêtes préparées : passer alors en `session`.

L'app reçoit `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER` et `DB_PASSWORD`. `DB_HOST` pointe vers
le pooler s'il est actif. Le chart n'injecte pas de `DATABASE_URL`. L'utilisateur et le mot de
passe viennent de Vault par le `VaultSecret` de l'app.

| Origine de l'app | Comment l'obtenir |
| --- | --- |
| Template fullstack | Actif d'office sur le backend : `db.enabled: true` |
| Template static | `db.enabled: true` dans `deploy/values.yaml`, ou depuis le panneau Configuration |
| Import | Un service `postgres` du compose devient une base managée, rattachée au seul service qui l'utilise. Le CMP réécrit ses variables `POSTGRES*`, `PG*`, `DATABASE_*` et `SPRING_DATASOURCE_*` vers `DB_*` |

### Bucket S3 sur Ceph RGW

- Un `ObjectBucketClaim` par app, bucket `cnp-<app>`, classe `ceph-bucket-retain` par défaut.
- Un utilisateur RGW propre au bucket. Rook crée la ConfigMap et le Secret de même nom.
- Interne seulement : le RGW n'est pas exposé sur Internet. Une URL présignée n'est donc pas
  utilisable depuis un navigateur externe.
- Hors site, mode `vault` : les identifiants sont lus dans Vault, à `<app>/bucket`.

L'app reçoit `S3_ENDPOINT`, `S3_BUCKET`, `S3_REGION`, `AWS_ACCESS_KEY_ID` et
`AWS_SECRET_ACCESS_KEY`. Le client doit adresser le bucket en path-style.

On l'obtient avec `bucket.enabled: true` dans le fichier de valeurs. Le CMP n'a pas encore de
bouton pour ça, et un import n'en crée jamais : il le recommande seulement.

### Conteneurs d'appui, en boîte noire

Pour les imports seulement. Le CMP reconnaît ces moteurs dans le compose et les lance tels quels.

| Moteur | Image | Port | Volume | Authentification |
| --- | --- | --- | --- | --- |
| MongoDB | `mongo:8.0` | 27017 | 1 Gio | Aucune |
| Redis | `redis:8.2-alpine` | 6379 | 1 Gio, AOF | Aucune |
| Neo4j | `neo4j:5.26-community` | 7687, 7474 | 1 Gio | Mot de passe généré, dans Vault |
| Elasticsearch | `elasticsearch:8.19.25` | 9200 | 2 Gio | Aucune, sans TLS |

- Images épinglées par digest, StatefulSet, non root, label `cnp.3istor.com/role: backing`.
- `network.allowFrom` n'ouvre leur port qu'aux apps qui les déclarent dans `dependsOn`. Sans
  aucune app liée, la liste est vide et le chart ne pose pas de règle.
- Alertes génériques : redémarrages, OOM, pod non prêt, volume presque plein, CPU bridé.
- **Pas de sauvegarde.** Un volume perdu est une donnée perdue.

Un autre moteur détecté est refusé à l'application du plan : il faut l'exclure. C'est le cas de
MySQL, MariaDB, RabbitMQ, Kafka, MinIO, Cassandra, NATS et d'autres.

## Ce qui est recommandé

L'analyseur lit le repo et propose de passer à un provider managé. Il ne change jamais le plan :
c'est un conseil. Les recommandations s'affichent sur le plan, dans l'onglet Imports du CMP et
dans `get_import` en MCP. Un admin peut en écarter une : elle reste écartée pour les analyses
suivantes du projet.

| Règle | Détecte | Propose | Niveau |
| --- | --- | --- | --- |
| `mongodb-to-postgres` | Un MongoDB | Postgres avec JSONB et index GIN | Conseillé |
| `neo4j-to-postgres` | Un Neo4j | Postgres avec une table de liaison | Conseillé si les requêtes Cypher font un saut, à évaluer sinon |
| `elasticsearch-to-postgres` | Un Elasticsearch | Recherche plein texte Postgres, `tsvector` et GIN | Conseillé si les requêtes sont simples, à évaluer sinon |
| `minio-to-bucket` | Une image MinIO | Bucket S3 | Conseillé |
| `local-storage-to-bucket` | `UPLOAD_DIR`, `MEDIA_ROOT` et variables proches | Bucket S3 | À évaluer |

Chaque recommandation donne la ligne du repo qui la justifie, la mémoire gagnée et l'effort.
La réécriture du code reste à l'équipe ou à son agent, par une PR dans son repo. Le skill
`cnp-migrate-to-provider` de cnp-agent-kit guide cette migration.

## Ce qui n'est pas fourni, et pourquoi

Décisions de Brian du 18/11/2026.

| Pas fourni | Pourquoi |
| --- | --- |
| FerretDB ou une autre couche de compatibilité | On standardise les bases au lieu d'émuler un moteur. Les recommandations poussent vers Postgres |
| D'autres providers : files de messages, cache managé, recherche | Le catalogue s'en tient aux principaux. Valkey, NATS et les files attendent un troisième nœud |
| Sauvegarde des conteneurs d'appui | Ils restent des boîtes noires, sans recette par moteur. La réponse est le provider managé |
| Environnements par PR | Décision du propriétaire : une app a prod et, au choix, staging |
| Tâches planifiées applicatives | Décision du propriétaire : pas de CronJob fourni |
| Envoi d'e-mails, SMTP | Décision du propriétaire : pas de relais fourni |
| Extensions PostgreSQL | Décision du propriétaire : le chart n'en active aucune. Le skill de migration demande de confirmer un besoin comme `pg_trgm` |
| Haute disponibilité Postgres | Le chart fixe `instances: 1` par défaut |
