# Audit du premier import (CNP-96)

Joe a construit un premier parcours d'import entre octobre et le 3 novembre, sur la branche
`import-fichier-joe-raph` de six dépôts. Ce travail a posé les bonnes exigences : commit figé,
scans avant déploiement, droits revérifiés à chaque étape, callbacks de CI authentifiés. Cette page
dit pourquoi on ne le fusionne pas tel quel et ce qu'on en reprend.

| Dépôt | PR | Taille |
| --- | --- | --- |
| CMP | [#172](https://github.com/3-Istor/CMP/pull/172) | 67 fichiers, +8 413 / -425 |
| extern-project-analyzer | [#42](https://github.com/3-Istor/extern-project-analyzer/pull/42) | 52 fichiers, +7 488 / -106 |
| K3s | [#157](https://github.com/3-Istor/K3s/pull/157) | 14 fichiers, +684 |
| app-templates | [#39](https://github.com/3-Istor/app-templates/pull/39) (brouillon) | 6 fichiers, +196 |
| infra-templates | [#20](https://github.com/3-Istor/infra-templates/pull/20) | 3 fichiers, +55 |
| cnp-projects | [#12](https://github.com/3-Istor/cnp-projects/pull/12) | 5 fichiers, +88 |

## Pourquoi on ne fusionne pas

Les branches datent du 2 novembre : 85 commits de retard côté CMP et 8 fichiers en conflit. Elles
ignorent les rôles CNP-228, la refonte de l'interface et les outils MCP. Au-delà du retard, le
modèle lui-même bloque :

| Gravité | Constat | Effet |
| --- | --- | --- |
| Critique | La clé privée de l'app GitHub CNP-Portal est copiée dans un Secret Argo pour chaque repo externe | Qui lit ces Secrets peut émettre des jetons pour toutes les installations de l'app, y compris l'org 3-Istor. Une rotation de la clé casse toutes les apps importées |
| Critique | L'utilisateur doit installer CNP-Portal sur son propre compte | Cette app a presque tous les droits d'une org (administration, workflows, secrets, membres) : elle les obtiendrait sur les repos de l'utilisateur |
| Critique | Un repo donne une seule app, avec une unicité globale sur l'URL et sur le nom d'app | Monorepo impossible ; un même nom d'app interdit dans deux projets |
| Critique | Une app importée est figée : ni release, ni configuration, ni push_files, ni image updater | Aucun chemin pour déployer une version 2 |
| Critique | `deploy/values.yaml` doit être identique octet par octet au contrat généré (256 Mio, une réplique) | Tout réglage bloque l'import ; un service Quarkus à 256 Mio tombe en OOM |
| Majeure | Autorisation propre à l'import, à côté du helper CNP-228 | Ignore les rôles plateforme, l'élévation JIT et le step-up |
| Majeure | La « revue indépendante » de la PR accepte l'approbation de l'importeur lui-même, ou de n'importe qui sur un repo public | La garantie n'existe pas |
| Majeure | Droits d'écriture demandés sur le repo de l'utilisateur, pour ouvrir la PR | Surface inutile |
| Majeure | Template Java incompatible avec Quarkus, conteneurs Python et Java en root | Build ou démarrage en échec |
| Majeure | Une douzaine d'appels distants à chaque étape, expiration à 1 h, échec à chaque push sur la branche | Parcours fragile |
| Majeure | Scanners à base de regex bloquants, Trivy bloquant même sans correctif disponible | La plupart des repos réels échouent |
| Majeure | Smoke test d'un conteneur seul, sans ses dépendances | Impossible pour un service qui attend MongoDB ou Redis |
| Mineure | Rien n'a tourné de bout en bout : ni OAuth dans un navigateur, ni CI jusqu'à Argo | Les tests prouvent l'ordre des transitions, pas le parcours |

## Ce qu'on reprend

| Brique | Où elle va | Ajustement |
| --- | --- | --- |
| Liaison GitHub par OAuth avec PKCE, state à usage unique lié à la session, jetons chiffrés, preuve de l'installation et des droits sur le repo | CMP, liaison des comptes | Clé de chiffrement rotative ; installation de CNP-Import, pas de CNP-Portal. Elle ferme aussi un trou de `main`, qui accepte un `installation_id` sans preuve |
| Vérification OIDC des callbacks de CI, transitions d'état gardées par tentative | CMP, callbacks de build | Épingler `job_workflow_ref` sur une révision taguée, pas sur la tête de `main` |
| Découpage de la CI : préparation sans code utilisateur, build sans secret, publication | Workflow central | Matrice ou reactor unique pour N services |
| Refus des sous-modules, LFS et liens symboliques ; contrôle des chemins | Analyseur | Inchangé |
| Scan du rendu Helm réel | Workflow central | Inchangé |
| Variables Terraform en fichier tfvars JSON plutôt qu'en arguments | Toutes les exécutions Terraform | Corrige aussi `main`, qui passe des secrets en ligne de commande |
| Champs de provenance du registre (SHA source, SHA chart, digest) et attente Argo sur les révisions exactes | cnp-projects, CMP | Inchangé |

Ce qu'on abandonne : la PR obligatoire et la reprise manuelle, l'égalité stricte des values, les
credentials Argo par repo, les introspections de la plateforme à chaque requête et les scanners à
regex comme barrière bloquante, remplacés par gitleaks et `trivy fs`.

Les six PRs seront fermées avec un commentaire qui renvoie vers les nouvelles et dit ce qui a été
repris ([CNP-246](https://3istors.atlassian.net/browse/CNP-246)).
