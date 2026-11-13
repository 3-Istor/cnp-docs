# Réversibilité et portabilité (Data Act, B1)

Ce que la plateforme détient pour un projet, ce qu'un export en emporte, dans quel
format, et comment un client récupère ses données puis quitte la CNP. Cette page
répond au règlement (UE) 2023/2854 (« Data Act »), applicable depuis le
12 septembre 2025, articles 23 à 30 sur le changement de fournisseur de services
de traitement de données.

!!! info "État au 13/11/2026"
    Export et exercice de restauration : CMP
    [#216](https://github.com/3-Istor/CMP/pull/216), câblage plateforme : K3s
    [#184](https://github.com/3-Istor/K3s/pull/184). Les délais contractuels
    marqués **[Décision]** sont des propositions, à valider par le responsable
    du service avant d'être repris dans les conditions d'utilisation.

## En bref

| Exigence | Article | Réponse CNP |
|---|---|---|
| Liste exhaustive des données exportables | 25(2)(e) | [Inventaire](#inventaire) ci-dessous, et `manifest.json` de chaque export |
| Données internes exclues, avec justification | 25(2)(f) | [Ce qui n'est pas exporté](#ce-qui-nest-pas-exporte) |
| Format structuré, couramment utilisé, lisible par machine | 30(5) | JSON, JSON Lines, SQL, bundle git, objets bruts. Voir [Formats](#formats) |
| Préavis de 2 mois au plus | 25(2)(d) | **[Décision]** 1 mois proposé |
| Transition de 30 jours calendaires au plus | 25(2)(a) | Service maintenu, exports à volonté |
| Récupération pendant 30 jours au moins après la transition | 25(2)(g) | Archive conservée 30 jours, liens renouvelables |
| Effacement complet ensuite | 25(2)(h) | Suppression du projet, purge des archives |
| Pas de frais de changement à partir du 12/01/2027 | 29 | Export gratuit dès aujourd'hui |

## Qui peut exporter

- Le **propriétaire** du projet seulement, celui qui l'a créé. Un admin de projet ou un
  admin plateforme ne peut pas : l'export emporte les bases, les objets et les journaux,
  c'est la divulgation la plus large que la CMP puisse faire.
- Avec une **passkey utilisée dans les 5 dernières minutes** (step-up Keycloak LoA 2).
  Les assistants IA sont refusés, et il n'existe pas d'outil MCP d'export : le lien de
  téléchargement donne accès à toutes ces données à quiconque le détient.
- Chaque demande d'export et chaque lien émis sont inscrits au journal d'audit
  (`project.export`, `project.export.link`).

## Inventaire

Tout ce que la plateforme détient pour un projet. « Client » : données fournies ou
générées par l'utilisation du service pour ce projet, exportées. « Plateforme » :
fonctionnement interne, non exportées, justifiées plus bas. La même liste est dans le
code (`backend/app/services/project_export/inventory.py`) ; un test échoue si une
nouvelle table de la CMP n'y est pas classée.

| Donnée | Source | À qui | Exporté | Fichier dans l'archive | Format |
|---|---|---|---|---|---|
| Projet, apps, environnements, promotions, vérifications, releases | Base CMP | Client | Oui | `cmp/<table>.jsonl` | JSON Lines |
| Journal d'audit des actions sur le projet | Base CMP `audit_events` | Client | Oui | `cmp/audit_events.jsonl` | JSON Lines |
| Activité (Argo CD, Loki) | Base CMP `activity_events` | Client | Oui | `cmp/activity_events.jsonl` | JSON Lines |
| FinOps : budgets, alertes, lignes de coût, centres de coût, économies, recommandations | Base CMP | Client | Oui | `cmp/<table>.jsonl` | JSON Lines |
| Sécurité : constats, scans, exceptions, historique | Base CMP | Client | Oui | `cmp/security_*.jsonl` | JSON Lines |
| Revues d'accès du projet | Base CMP `access_review_items` | Client | Oui | `cmp/access_review_items.jsonl` | JSON Lines |
| Réglages assistants IA et confirmations du projet | Base CMP | Client | Oui | `cmp/project_agent_settings.jsonl`, `cmp/agent_confirmations.jsonl` | JSON Lines |
| Dépôts GitOps des apps (code, `deploy/values-*.yaml`) | GitHub, organisation 3-Istor | Client | Oui, toutes les branches et tags | `git/<org>__<repo>.bundle` | git bundle |
| Dépôts hors organisation 3-Istor | GitHub du client | Client | Listés seulement : le client les détient déjà | `git/repositories.json` | JSON |
| Applications Argo CD (sources, fichiers de valeurs, révisions) | Argo CD | Client | Oui | `gitops/applications/<app>.json` | JSON (manifeste Kubernetes) |
| Objets Kubernetes des namespaces du projet (Deployments, Services, ConfigMaps, HTTPRoutes, CNPG Clusters, VaultSecrets...) | Cluster | Client | Oui, nettoyés pour être réappliqués ailleurs | `kubernetes/<ns>/<Kind>/<nom>.json` | JSON (manifeste Kubernetes) |
| Noms et clés des Secrets Kubernetes | Cluster | Plateforme | Noms seulement | `kubernetes/<ns>/secret-names.json` | JSON |
| Bases PostgreSQL (CNPG) | Cluster CNPG de chaque app | Client | Oui, dump logique + nombre de lignes par table | `databases/<ns>/<cluster>.sql` | SQL `pg_dump` texte |
| Buckets du projet (ObjectBucketClaims Ceph RGW) | Ceph RGW | Client | Oui, objet par objet | `objects/<ns>/<bucket>/<clé>` | Octets bruts |
| Secrets du projet (Vault `project-<nom>`) | Vault | Client | Noms toujours, valeurs sur demande explicite | `secrets/names.json`, `secrets/values.json` | JSON |
| Realms Keycloak du projet (`<projet>`, `<projet>-staging`) | Keycloak | Client | Configuration sans secrets, utilisateurs sans mots de passe | `identity/<realm>/realm.json`, `users.jsonl` | JSON, JSON Lines |
| Membres du projet et leur rôle | Groupes Keycloak de la plateforme | Client | Oui | `identity/members.json` | JSON |
| Journaux applicatifs | Loki, namespaces exacts du projet | Client | Oui, ce que Loki conserve (14 jours) | `logs/<projet>.jsonl` | JSON Lines |
| Images de conteneurs | GHCR | Client | Références avec digest, pas les couches | `images.json` | JSON |
| Métriques | Prometheus | Client | Non, sur demande | | |
| Volumes de StatefulSet (ex. MongoDB) | Cluster | Client | Non, copie à la demande | | |

### Secrets : noms par défaut, valeurs sur demande

Les valeurs des secrets du projet sont des données du client : il les a saisies. Elles
ne partent pourtant dans l'archive que si le propriétaire coche l'option
(`include_secret_values`), pour trois raisons :

1. un lien présigné est un porteur : quiconque le reçoit lit l'archive, secrets compris ;
2. une archive circule (poste, ticket, stockage partagé) bien plus que Vault ;
3. le nouvel hébergeur doit de toute façon émettre ses propres identifiants
   (base, registre, OIDC) : seuls les secrets métier du client ont une valeur hors CNP.

L'option demande la même passkey que l'export et figure dans le journal d'audit.

## Formats

L'archive est un `.tar.gz`. `manifest.json`, à sa racine, décrit chaque fichier :

```json
{
  "schema": "cnp.project-export/v1",
  "export_id": "3f1c...",
  "project": "shop",
  "created_at": "2026-11-13T10:00:00Z",
  "include_secret_values": false,
  "items": [
    {
      "path": "databases/shop-web/shop-web-backend-db.sql",
      "type": "postgres.dump",
      "source": "cnpg:shop-web/shop-web-backend-db",
      "format": "application/sql",
      "sha256": "9b1e...",
      "size": 48213,
      "metadata": {"database": "shop", "row_counts": {"public.orders": 1000}}
    }
  ],
  "not_exported": [{"data": "...", "owner": "platform", "reason": "..."}],
  "problems": []
}
```

| `type` | Contenu | Relire avec |
|---|---|---|
| `cmp.table` | Une ligne JSON par enregistrement, `metadata.columns`, `metadata.rows` | `jq`, tout langage |
| `kubernetes.object`, `gitops.application` | Manifeste Kubernetes, sans `status`, UID, `resourceVersion`, IP de Service | `kubectl apply -f` |
| `git.bundle` | Toutes les refs, `metadata.head` | `git clone <fichier>.bundle` |
| `postgres.dump` | SQL texte sans propriétaire ni droits, `metadata.row_counts` | `psql -f` dans n'importe quel PostgreSQL ≥ version source |
| `object` | Contenu brut, `metadata.bucket` et `metadata.key` | `aws s3 cp`, `mc cp` |
| `vault.secret-names`, `vault.secret-values` | `{chemin: [clés]}` ou `{chemin: {clé: valeur}}` | `vault kv put` |
| `keycloak.realm`, `keycloak.users` | Export partiel Keycloak, secrets masqués | Import de realm Keycloak |
| `loki.logs` | `{time, labels, line}` par ligne | `jq`, import Loki |

Les manifestes Kubernetes sont en JSON et non en YAML : `kubectl` lit le YAML 1.1, où un
`on` ou un `no` non cité devient un booléen. L'exercice de restauration est tombé
dessus.

Chaque fichier de l'archive est un élément du manifeste et réciproquement ; le SHA-256
de l'archive entière est donné avec le lien.

`problems` liste ce qui n'a pas pu être collecté (source injoignable). L'export est
alors `partial` et non `ready` : il reste téléchargeable, et le propriétaire en relance
un autre.

## Procédure

### Depuis la CMP (API)

```bash
# 1. Lancer l'export (propriétaire, passkey récente)
curl -X POST https://cmp.3istor.com/api/projects/shop/export \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"include_secret_values": false}'
# -> 202 {"id": "3f1c...", "status": "pending", ...}

# 2. Suivre : pending, running, puis ready ou partial
curl https://cmp.3istor.com/api/projects/shop/exports/3f1c... -H "Authorization: Bearer $TOKEN"

# 3. Obtenir un lien (passkey récente) et télécharger
curl -X POST https://cmp.3istor.com/api/projects/shop/exports/3f1c.../link \
  -H "Authorization: Bearer $TOKEN"
# -> {"url": "https://s3.3istor.com/...", "expires_in": 604800, "sha256": "..."}
curl -o export.tar.gz "$URL" && sha256sum export.tar.gz
```

Un seul export à la fois par projet (409 sinon).

### Pourquoi le lien dure 7 jours et l'archive 30

Une URL présignée SigV4 ne peut pas dépasser 7 jours : AWS, Ceph RGW et Garage la
refusent au-delà (Garage répond `X-Amz-Expires may not exceed a week`, vérifié). La
période de récupération de 30 jours porte donc sur l'archive, pas sur le lien : le
propriétaire demande un nouveau lien quand il veut pendant les 30 jours, et aucun lien
ne survit à l'archive.

### Restaurer ailleurs

`scripts/export-drill/restore.sh` (dépôt CMP) restaure une archive et prouve qu'elle est
complète. Il ne dépend pas du code de la CMP : `jq`, `sha256sum`, `git`, `kubectl`,
`psql` et la CLI `aws` suffisent, le client peut s'en servir tel quel.

```bash
TARGET_PG_URL=postgresql://postgres:...@db.example:5432/postgres \
TARGET_S3_ENDPOINT=https://s3.example AWS_ACCESS_KEY_ID=... AWS_SECRET_ACCESS_KEY=... \
  scripts/export-drill/restore.sh export.tar.gz <sha256>
```

Le PostgreSQL cible doit être au moins de la version de `pg_dump` qui a fait l'export
(18 aujourd'hui, celle des clusters CNPG). Un dump fait par un `pg_dump` plus récent
contient des réglages qu'un serveur plus ancien refuse : l'exercice l'a montré avec un
dump 17 restauré dans un PostgreSQL 16.

1. Vérifie le SHA-256 de l'archive, refuse un chemin absolu ou contenant `..`.
2. Vérifie chaque fichier contre le manifeste, dans les deux sens.
3. Applique les namespaces puis les objets Kubernetes au contexte `kubectl` courant ;
   un type que le cluster ne sert pas (CRD CNPG, Gateway API...) est signalé, pas en
   échec : installer ces opérateurs fait partie du nouvel hébergement.
4. Restaure chaque dump dans une base neuve et compare le nombre de lignes de chaque
   table à `row_counts`.
5. Copie chaque objet dans un bucket `restored-<bucket>` et relit son SHA-256.
6. Clone chaque bundle git et compare `HEAD`.
7. Écrit `drill-report.json`, code de sortie 1 au moindre écart.

## Exercice de restauration automatisé

Le job GitHub Actions [`Export restore drill`](https://github.com/3-Istor/CMP/actions/workflows/export-restore-drill.yml)
est la preuve que l'export se restaure. Il tourne sur chaque PR qui touche l'export,
sur `main` et chaque lundi, sans aucun secret de la plateforme :

```mermaid
flowchart LR
  subgraph runner[Runner GitHub]
    seed[Projet de test<br/>CMP SQLite, faux cluster,<br/>faux Vault, Keycloak, Loki]
    src[(PostgreSQL 18<br/>source)]
    garage[(Garage v2.2.0<br/>buckets source + exports)]
    exp[Code d'export de la CMP<br/>pg_dump, S3, git réels]
    dst[(PostgreSQL 18<br/>cible)]
    kind[[kind vierge]]
    restore[restore.sh]
  end
  seed --> exp
  src --> exp
  garage --> exp
  exp -- archive --> garage
  garage -- lien présigné --> restore
  restore --> kind
  restore --> dst
  restore -- objets --> garage
```

1. `run.sh` démarre Garage et deux PostgreSQL 18, les versions de la plateforme, avec des
   identifiants tirés au hasard à chaque exécution.
2. `make_fixture.py` remplit un projet de test (250 clients, 1000 commandes, 3 objets,
   un dépôt git à deux branches) et lance le vrai code d'export ; seuls la base CMP,
   le cluster, Vault, Keycloak et Loki sont simulés.
3. L'archive est téléchargée par son lien présigné, comme le ferait le client, puis
   restaurée par `restore.sh` dans un cluster kind vierge, le PostgreSQL cible et
   Garage.
4. Une archive altérée (une ligne ajoutée au dump) doit faire échouer `restore.sh`,
   sinon le job échoue : les contrôles ne sont pas décoratifs.

Le manifeste et `drill-report.json` sont gardés en artefact du job. Les adaptateurs
simulés (API Kubernetes, Vault, Keycloak, Loki) ont leurs propres tests unitaires ; ils
ne se vérifient en réel qu'avec un export d'un projet sandbox en production.

## Calendrier de sortie

```mermaid
gantt
  dateFormat  YYYY-MM-DD
  axisFormat  J%j
  section Client
  Préavis (≤ 2 mois, 1 mois proposé)        :a1, 2027-01-01, 30d
  Transition (≤ 30 jours)                   :a2, after a1, 30d
  Récupération (≥ 30 jours)                 :a3, after a2, 30d
  section Plateforme
  Service normal, exports à volonté         :b1, 2027-01-01, 60d
  Export final, projet gelé                 :milestone, after a2, 0d
  Effacement                                :milestone, after a3, 0d
```

| Étape | Durée | Ce que fait la plateforme |
|---|---|---|
| Demande de sortie | | Le propriétaire la notifie (ticket ou courriel). Les exports restent possibles à tout moment, sans attendre. |
| Préavis | **[Décision]** 1 mois proposé, 2 mois maximum (art. 25(2)(d)) | Service normal. |
| Transition | 30 jours calendaires maximum (art. 25(2)(a)) | Service maintenu, assistance raisonnable, exports à volonté. Si une migration est techniquement impossible dans ce délai, le fournisseur le notifie sous 14 jours ouvrables et propose une transition d'au plus 7 mois (art. 25(4)). |
| Fin de transition | | Export final avec la dernière version des données ; le projet passe en lecture seule. |
| Récupération | 30 jours calendaires (art. 25(2)(g)) | L'archive finale est conservée 30 jours ; nouveaux liens à la demande. Un accès en lecture aux images GHCR est accordé pendant cette période **[Décision]**. |
| Effacement | À l'issue de la récupération (art. 25(2)(h)) | Suppression des apps puis du projet par la CMP (namespaces et bases des apps, groupes Keycloak, politique Vault, AppProject Argo CD, réglages GitHub, état Terraform), purge automatique des archives. Les sauvegardes CNPG expirent sous 7 jours (`retentionPolicy: 7d`), les journaux Loki sous 14 jours. Le journal d'audit de la CMP suit sa propre durée (`AUDIT_RETENTION_DAYS`). |

## Ce qui n'est pas exporté

| Donnée | À qui | Pourquoi |
|---|---|---|
| Couches des images de conteneurs (GHCR) | Client | Reconstruites depuis les dépôts exportés, ou tirées pendant la récupération. `images.json` donne chaque référence avec son digest. |
| Sauvegardes physiques CNPG (base backups et WAL dans Ceph RGW) | Plateforme | Copie des mêmes données dans un format propre à CloudNativePG ; le dump logique porte le contenu dans un format ouvert. |
| Valeurs des Secrets Kubernetes et identifiants de la plateforme | Plateforme | Mots de passe de base, accès registre, secrets clients OIDC : générés par la plateforme, inutiles ailleurs. Leurs noms sont exportés. |
| Métriques Prometheus | Client | Mesures dérivées, régénérées par tout nouvel hébergement. Export au format texte Prometheus sur demande pendant la transition. |
| Volumes de StatefulSet | Client | Pas encore collectés automatiquement : copie du volume à la demande pendant la transition, ou outil de dump de l'application. |
| État Terraform du projet | Plateforme | Description des ressources partagées de la plateforme (politiques Vault, groupes Keycloak, réglages GitHub), avec des identifiants. |
| Journaux de la plateforme hors du tenant du projet | Plateforme | Ingress, opérateurs, plan de contrôle partagés ; les actions sur le projet sont dans le journal d'audit exporté. |
| Notifications, préférences, liaison GitHub personnelle | Personnes | Données de chaque utilisateur, pas du projet : demande individuelle (RGPD art. 15 et 20). |
| Grants, clients et jetons des assistants IA | Personnes / plateforme | Identifiants de connexion ; les actions des assistants sont dans le journal d'audit exporté. |
| Taux de change, intensité carbone, clôtures mensuelles, états internes FinOps et des collecteurs | Plateforme | Données publiques ou tenue de comptes interne. |

## Reste à faire avant la production

- **Bucket et clé Garage** `cnp-project-exports`, valeurs dans Vault
  (K3s [#184](https://github.com/3-Istor/K3s/pull/184)). Sans elles, l'export répond 503.
- **Dumps CNPG** : la CMP doit joindre `<cluster>-rw:5432` dans les namespaces des
  projets. C'est un élargissement de ce que la CMP peut atteindre, à décider
  **[Décision]** ; l'alternative est un Job `pg_dump` lancé dans le namespace du projet.
- **Buckets des projets** : la découverte passe par les ObjectBucketClaims ; aucun projet
  n'en a hors sauvegardes CNPG au 13/11/2026.
- **Volumes de StatefulSet et métriques** : procédure manuelle tant qu'ils ne sont pas
  collectés.
- **Interface CMP** : l'API existe, pas encore de bouton dans le portail.
- **Effacement complet** : vérifier, et compléter si besoin, que la suppression retire
  aussi les dépôts des apps dans l'organisation 3-Istor, les paquets GHCR, le montage
  Vault `project-<nom>` et les realms Keycloak du projet. Non vérifié au 13/11/2026.
