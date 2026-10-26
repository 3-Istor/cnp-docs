# FinOps

> **Statut : plan validé le 2026-10-26, on-prem implémenté, AWS et GCP prêts mais pas
> activés.** Cette section décrit le module FinOps de la CMP tel qu'il est mergé : d'où
> viennent les coûts, comment chaque chiffre est calculé et comment l'exploiter.

Le module FinOps passe d'un moteur simulé à des coûts réels, par projet et par application,
sur on-prem, AWS et GCP. Il y ajoute l'énergie et le carbone, les économies d'offhours-guard
et la réconciliation avec les factures.

## Principe en une phrase

La facture donne le montant, OpenCost donne la répartition, Git donne le prix de l'on-prem,
et chaque chiffre affiché dit d'où il vient.

## Documents

| Page | Contenu |
|---|---|
| [Décisions](01-decisions.md) | D-14 : le plan de données FinOps, avec les huit choix D-F1 à D-F8 |
| [Architecture](02-architecture.md) | Sources, ingestion, stockage, API et interface |
| [Méthodes de calcul](03-methodes.md) | Formule de chaque nombre, valeurs réelles et valeurs moquées |
| [Exploitation](04-exploitation.md) | Réglages, changer un prix, alertes, vérifier l'ingestion, activer un cloud |

## État de la mise en œuvre

| Élément | État |
|---|---|
| OpenCost on-prem, prix issus du TCO versionné | En production |
| Ingestion `opencost`, `ai_gateway`, `saas` dans `cost_line_items` | En production |
| Économies offhours-guard (évité et encaissé) | En production |
| GreenOps on-prem (énergie, carbone RTE et ADEME) | Mergé |
| Ingestion AWS et GCP, partage de facture, réconciliation | Mergé, inactif tant que les clouds ne sont pas activés |
| Source `ai_gateway` | Codée et testée sur fixtures, en attente d'agentgateway |
| Exporter de puissance des hyperviseurs | Reporté, le TCO prend la moyenne mesurée des CSV |
| Clôture mensuelle, exports XLSX et PDF, rôle `cnp-finance`, prévision | À faire |
| Agent de recommandation | Après un mois de données réelles |

## Liens avec le reste de la documentation

- [WS-9, cycle de vie et FinOps](../06-multicloud/04-chantiers-et-sequencement.md) : le chantier qui porte la vue par cloud et la réconciliation.
- [FinOps de l'IA](../07-ai-landing-zone/06-finops-ia.md) : comment les tokens de la gateway deviennent des lignes de coût.
- [Décisions D-01 à D-13](../06-multicloud/02-decisions.md) : D-01 (Git source de vérité), D-10 (tampon des métriques) et D-13 (marquage `cnp.*`) sont des prérequis.
