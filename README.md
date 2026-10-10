# cnp-docs

## Rôle

Documentation centrale de la CNP : architecture, contrats, composants, workflows et cartographie des dépôts. Les README locaux décrivent leur dépôt et renvoient ici pour comprendre l’ensemble.

## Technologies

Markdown, MkDocs Material, Mermaid, YAML et GitHub Actions ; modèles Noodle et exports draw.io/PNG pour les diagrammes d’architecture existants.

## Entrées

| Origine / destinataire | Contenu et transmission |
| --- | --- |
| Les dix autres dépôts locaux | Code, manifests et workflows consultés par les rédacteurs pour vérifier les descriptions et dépendances ; aucune génération automatique des fiches. |
| Contributeurs | Décisions et procédures techniques documentées, modèles et vues d’architecture. |

## Sorties et consommateurs

| Origine / destinataire | Contenu et transmission |
| --- | --- |
| Équipe CNP / lecteurs | Site de documentation construit avec MkDocs et publié par le workflow GitHub Pages. |
| CMP / développeurs et assistants | Sources Markdown référencées comme sous-module dans `.kiro/steering/docs`. |
| Tous les dépôts CNP | Cartographie centrale vers laquelle pointent les README locaux. |

## Documentation CNP

[Cartographie complète et workflows inter-repo](docs/04-templates/00-github-repositories-landscape.md#cnp-docs).

## Lire la documentation

- [Accueil](docs/index.md).
- [Guide développeur](docs/developer-onboarding-guide.md).
- [Architecture et topologies](docs/01-architecture/01-system-overview.md).
- [Composants de plateforme](docs/02-core-components/01-cmp-dashboard.md).
- [Provisioning et workflows](docs/03-pipelines-and-workflows/01-app-provisioning-flow.md).
- [Templates applicatifs](docs/04-templates/01-git-app-templates.md).
- [Contrats API CMP](docs/05-cmp-backend-api/00-global-api-standards.md).
- [Chantier multicloud](docs/06-multicloud/00-index.md).
- [Roadmap](docs/README_ROADMAP.md).

Les URL de plateforme documentées sont `https://cmp.3istor.com`, `https://auth.3istor.com`, `https://vault.3istor.com` et `https://argocd.3istor.com` ; leur disponibilité n’est pas vérifiée par cette documentation.

## Construire et consulter le site

Avec MkDocs Material installé dans un environnement Python :

```bash
mkdocs serve
```

Pour construire le site sans publier :

```bash
mkdocs build
```

La navigation est définie dans [mkdocs.yml](mkdocs.yml). La cartographie des dépôts conserve sa page dans la rubrique « Modèles & Templates ».

## Maintenir la documentation

Lorsqu’un contrat inter-repo change, actualiser le README du producteur, celui du consommateur et [la cartographie centrale](docs/04-templates/00-github-repositories-landscape.md). Vérifier les relations dans le code et les configurations, puis contrôler les liens et le rendu MkDocs.

Les diagrammes d’architecture Noodle existants ont leurs propres [instructions de génération](architecture/README.md). Le diagramme inter-repo Mermaid est maintenu directement dans la page de cartographie.
