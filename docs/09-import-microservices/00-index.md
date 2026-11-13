# Import de repos, microservices et topologie

**Statut : dossier de conception à valider, rien n'est construit.** Ticket parent
[CNP-158](https://3istors.atlassian.net/browse/CNP-158), sous-tâches CNP-237 à CNP-246.

Une équipe arrive avec son code existant : un repo GitHub ou GitLab, parfois un monorepo de
microservices avec son docker-compose et ses bases. La CNP le lit à un commit figé et propose un
plan d'import sur la carte du projet. Après validation, elle construit chaque service, lance les
bases, relie les services entre eux sans toucher à leur code et montre en continu qui parle à qui.
Le repo de l'utilisateur reste la source de vérité. Les mêmes liens et la même carte servent aux
apps créées par template.

La démonstration est epitweet : 6 services Quarkus, un front React, MongoDB, Redis, Neo4j et
Elasticsearch, importés tels quels.

## Les pages

| Page | Ce qu'on y trouve |
| --- | --- |
| [Valeur pour les utilisateurs](01-valeur.md) | Études 2025-2026, personas, fonctions retenues en v1 et après |
| [Architecture](02-architecture.md) | Ce qui est ajouté composant par composant, sources de la topologie, décisions D1 à D12 |
| [Parcours](03-parcours.md) | Les 7 parcours avec leurs diagrammes de séquence |
| [Rôles et sécurité](04-roles-et-securite.md) | Qui peut faire quoi, identités techniques, décisions D13 à D21 |
| [Epitweet](05-epitweet.md) | Ce que contient epitweet, ce que la CNP ajoute, budget mémoire |
| [Audit du premier import](06-audit-travail-existant.md) | Pourquoi les PRs de CNP-96 ne sont pas fusionnées, ce qu'on en reprend |
| [Plan et recette](07-plan-et-recette.md) | Ordre de livraison, jalons, critères de recette |
| [Audit de sécurité](08-audit-securite.md) | Rapport complet, méthode du skill `security-audit` de Cloudflare |

## Comment ce dossier a été construit

1. Audit du travail existant (six PRs de CNP-96) et cartographie de la plateforme actuelle.
2. Veille sur l'état de l'art technique : Railway, Render, Coolify, Score, Radius, Hubble, Kiali,
   Railpack, chaîne d'approvisionnement. Veille métier en parallèle : études 2025-2026, personas.
3. Brouillon d'architecture en 10 décisions, puis revue contradictoire par un relecteur
   indépendant. Il a confronté chaque décision au code et en a fait réviser huit.
4. Prototypes jetables pour vérifier les hypothèses. Le Dockerfile réparé d'un service Quarkus
   construit en 35 s et tourne avec 67 Mio. Le chart étendu rend à l'identique les apps
   existantes.
5. Audit de sécurité du design et du code repris, avec revérification indépendante du constat
   confirmé.

## Ce qu'il faut trancher

- **Constat haut sur `main`** (D13) : `POST /api/deployments` accepte un `github_installation_id`
  fourni par le client. Il n'est pas exploitable tant que CNP-Portal n'est installé que sur
  3-Istor. Je propose de le corriger en tête de ce chantier.
- **Nouvelle app GitHub CNP-Import** (D3) : elle se crée dans les réglages de l'org 3-Istor par
  un owner.
- **Application OAuth sur le GitLab de l'EPITA** : à déclarer par un compte EPITA. Sinon, la
  démonstration part d'un miroir GitHub.
- **Kyverno en mode Deny** sur les namespaces d'apps importées (D16) : les apps existantes ne
  sont pas concernées.
- **Mémoire** : epitweet demande environ 3,4 Gio sur les 4 Gio libres du cluster on-prem.
