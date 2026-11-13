# Import de repos, microservices et topologie

**Statut : validé le 13/11/2026, implémentation en cours.** Ticket parent
[CNP-158](https://3istors.atlassian.net/browse/CNP-158), sous-tâches CNP-237 à CNP-246.

Une équipe arrive avec son code existant : un repo GitHub, parfois un monorepo de
microservices avec son docker-compose et ses bases. La CNP le lit à un commit figé et propose un
plan d'import sur la carte du projet. Après validation, elle construit chaque service, lance les
bases, relie les services entre eux sans toucher à leur code et montre en continu qui parle à qui.
Le repo de l'utilisateur reste la source de vérité. La carte montre aussi les apps créées par
template ; les relier viendra en v2, le schéma du registre ne l'accepte pas encore.

La démonstration est epitweet : 6 services Quarkus, un front React, MongoDB, Redis, Neo4j et
Elasticsearch, importés tels quels.

## Où on en est (13/11/2026 au soir)

| Brique | Livré dans | État |
| --- | --- | --- |
| Correctif du constat haut, jetons restreints | CMP#212 | En prod |
| Types `service` et `backing`, liens, chart `workload` | infra-templates#24 #27, cnp-project-base#33, K3s#181, cnp-projects#17 | En prod |
| Analyseur v2 et workflow central | extern-project-analyzer#43 #44 | Mergé |
| Backend d'import, liaison GitHub OAuth | CMP#217 | En prod (v0.3.105), import désactivé |
| Onglet Topologie, assistant d'import | CMP#220 | En prod (v0.3.105) |
| Métriques de flux Hubble | K3s#187 | En prod |
| Configuration de l'import, clé de chiffrement | K3s#189 | Appliqué |
| Runners des repos privés, réglages de CNP-Portal, premier import d'epitweet | | À faire |

## Les pages

| Page | Ce qu'on y trouve |
| --- | --- |
| [Valeur pour les utilisateurs](01-valeur.md) | Études 2025-2026, personas, fonctions retenues en v1 et après |
| [Architecture](02-architecture.md) | Ce qui est ajouté composant par composant, sources de la topologie, décisions D1 à D12 |
| [Parcours](03-parcours.md) | Les 6 parcours avec leurs diagrammes de séquence |
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

## Décisions de Brian du 13/11

- **Une seule app GitHub : CNP-Portal.** Pas d'app CNP-Import. Ses permissions sont réduites à
  ce que la plateforme utilise vraiment. L'utilisateur installe CNP-Portal sur son compte ou son
  organisation, en choisissant ses repos. Pour une installation utilisateur, le CMP ne demande
  que des jetons `contents:read` et `metadata:read`, limités à un repo. Le risque résiduel est
  accepté par écrit (D14, D15).
- **Constat haut sur `main`** (D13) : corrigé en tête du chantier. Il devient exploitable dès
  que CNP-Portal est installé hors de 3-Istor.
- **Pas de GitLab dans ce chantier.** Epitweet est poussé sur GitHub. Un import depuis GitLab
  reste une extension possible plus tard.
- **Kyverno en mode Deny : chantier séparé.** Ce chantier garde le schéma strict des valeurs
  générées par le CMP (D16).
- **Mémoire** : environ 4 Gio pour epitweet, accepté.
- Le reste du dossier est approuvé tel quel.
