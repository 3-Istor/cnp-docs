# Cycle de vie : import d'un repo

Cette page suit un repo importé, de l'analyse à la suppression. Elle décrit ce qui est sur
`main` le 20/11/2026 : CMP `services/imports`, workflow `import.yml` d'extern-project-analyzer,
ApplicationSet `cnp-apps` de K3s et chart infra-templates. Le passage d'un repo importé au
modèle des templates, en service depuis le 20/11, est décrit à la fin, dans
[Après la PR standard](#apres-la-pr-standard).

Les étapes d'analyse et d'application sont détaillées dans [Parcours](03-parcours.md). Le parcours
d'une app créée par template est dans [Cycle de vie : template](10-cycle-template.md).

![Cycle de vie d'un repo importé](../01-architecture/diagrams/lifecycle-import.light.png#only-light)
![Cycle de vie d'un repo importé](../01-architecture/diagrams/lifecycle-import.dark.png#only-dark)

## Qui fait quoi

| Acteur | Rôle |
| --- | --- |
| Admin du projet | Lance l'analyse, relit et applique le plan avec un step-up, supprime l'import |
| Membre du projet | Endort et réveille les apps |
| Module d'import du CMP | Lance le workflow, écrit Vault, le registre et le DNS, suit le déploiement |
| Workflow `import.yml` | Lit le repo à un SHA figé, construit, scanne, signe et pousse les images |
| CNP-Portal | Envoie les webhooks de push du repo importé |
| Argo CD | Rend le chart avec les valeurs de cnp-projects. Ne lit jamais le repo importé |
| offhours-guard | Endort et réveille les apps, vague par vague |
| Admin plateforme | Consulte le contrôle d'orphelins |

## Création

1. L'admin lance l'analyse avec `POST /api/projects/<p>/imports`. Le CMP fige la branche en SHA.
2. Le workflow analyse le repo et renvoie un plan : services, conteneurs d'appui, base Postgres
   managée, recommandations, estimation mémoire et coût.
3. L'admin applique le plan avec un step-up. Il peut exclure ou renommer une app, la rendre
   publique, changer sa mémoire, son port, sa sonde, retirer un lien, activer le pooler ou
   couper la veille.
4. Le workflow construit chaque service et pousse `ghcr.io/3-istor/imports/<p>/<app>` par
   digest, signé avec cosign.
5. Le CMP écrit Vault, puis le registre en une vague par niveau de dépendance : conteneurs
   d'appui d'abord, puis services, puis apps publiques. Chaque vague est un commit.
6. Argo CD synchronise. Le CMP attend Synced, Healthy et l'image exacte. Il teste ensuite
   `GET /` sur les apps publiques.

| Ressource | Nom | Créée par |
| --- | --- | --- |
| Image | `ghcr.io/3-istor/imports/<p>/<app>@sha256:…` | Workflow, job `publish` |
| Entrée du registre | `spec.apps` de `registry/projects/<p>.yaml` : `name`, `type` `service` ou `backing`, `source{repo, path, sha}`, `public`, `hostnames`, `dependsOn` | CMP |
| Fichier de valeurs | `apps/<p>/<app>/values.yaml` dans cnp-projects : image et digest, ports, ressources, env, sondes, `db`, `offhours`, `network.allowFrom` | CMP |
| Secrets Vault | `project-<p>/<app>/imports`. Secret de registre partagé `project-<p>/system/registry` | CMP |
| Politique Vault | `project-<p>-app-<app>-policy` | CMP |
| Rôle Vault | `<p>-<app>-role` | CMP |
| DNS | CNAME `<app>-<p>.3istor.com`, apps publiques seulement | CMP, API Cloudflare |
| Application Argo CD | `<p>-<app>-app` | ApplicationSet `cnp-apps` |
| Namespace | `<p>-<app>`, gardé à la suppression de l'Application | Application `<p>-bootstrap` |
| Workload | Deployment pour un service, StatefulSet et volume pour un conteneur d'appui | Chart |
| Base CNPG | Un `Cluster` par service qui utilisait un Postgres du compose, sauvegarde forcée. `Pooler` si demandé | Chart, bloc `db` |
| Planning de veille | `OffhoursSchedule`, actif par défaut, décalé par vague | Chart, bloc `offhours` |

Le rôle et la politique Vault n'existent que si l'app a un secret, une base, ou si c'est Neo4j.
Un import ne crée ni client Keycloak, ni repo GitHub, ni bucket S3. Aucun Image Updater ne suit
les images importées : le CMP écrit le nouveau digest après chaque build.

## Où vivent les valeurs et les secrets

Le repo importé garde le code. Tout ce qui sert à déployer vit côté plateforme.

| Je veux changer… | Où c'est stocké | Comment le changer aujourd'hui |
| --- | --- | --- |
| Mémoire, port, sonde, exposition publique, pooler, veille | `apps/<p>/<app>/values.yaml` dans cnp-projects | Modifier le plan dans l'onglet Imports, puis appliquer à nouveau |
| Le code, un Dockerfile, le compose | Le repo importé | Push sur la branche importée. Le rebuild suit si `auto_rebuild` est actif |
| Un lien entre deux apps | `dependsOn` du registre | Canevas Topologie, ou `link_apps` en MCP |
| Un secret généré | Vault `project-<p>/<app>/imports` | Écrit par le CMP, pas d'écran dédié |
| Une variable d'environnement libre | `values.yaml` dans cnp-projects | Pas de chemin pris en charge |

Le panneau Configuration du CMP répond 409 sur une app importée, en lecture comme en écriture.
Un commit à la main dans cnp-projects est écrasé au build suivant : l'en-tête du fichier le dit.

## Mise à jour

- **Push :** CNP-Portal envoie le webhook. Le CMP vérifie la signature HMAC et ignore une
  livraison déjà vue. Il ne fait rien si `auto_rebuild` est coupé.
- **Ce qui est reconstruit :** le service dont le dossier ou le Dockerfile a changé. Tous les
  services si le compose ou un manifeste partagé a changé, comme un `pom.xml` parent.
- **Rattrapage :** au démarrage du CMP, puis toutes les 10 minutes. Il compare le dernier SHA
  construit à la tête de la branche.
- **Identité :** le rebuild agit au nom de l'utilisateur qui l'a activé. S'il a perdu ses
  droits, l'import attend qu'un admin confirme avec un step-up.
- **À la main :** bouton « Reconstruire » ou `POST …/rebuild`, avec step-up. Une nouvelle
  analyse se lance comme la première.

## Veille

Chaque app importée reçoit un `OffhoursSchedule`, décalé selon sa vague de déploiement. Les
conteneurs d'appui se réveillent en premier à 07:00 et s'endorment en dernier à 01:30. Chaque
vague suivante se réveille 5 minutes plus tard et s'endort 5 minutes plus tôt. Une app ne se
réveille donc jamais avant ce qu'elle appelle. La veille se coupe app par app dans le plan.

Les boutons Dormir et Réveiller du CMP marchent comme pour une app de template. Un webhook perdu
pendant la nuit du CMP est rattrapé à son réveil.

## Suppression

Une app importée ne se supprime pas seule : le CMP répond 409. On supprime l'import entier
depuis l'onglet Imports, avec un step-up. Un import en cours de build ou de déploiement est
refusé.

1. Un commit retire du registre les entrées de l'import, leurs fichiers de valeurs et les
   `dependsOn` qui les visent. Une app de même nom que l'import n'a pas créée reste en place.
2. Le CMP supprime les CNAME.
3. Il supprime le rôle et la politique Vault.
4. Il attend la disparition de l'Application, puis supprime le namespace avec ses volumes et
   sa base CNPG. Au-delà de 10 minutes, l'import passe en échec.

Restent après la suppression :

- le dossier Vault `project-<p>/<app>/imports` ;
- les images `ghcr.io/3-istor/imports/<p>/<app>` ;
- les sauvegardes CNPG dans Ceph RGW, non vérifié.

Le contrôle d'orphelins voit le dossier Vault restant. Il voit aussi les rôles, politiques,
namespaces et entrées de registre sans app vivante. Il ne voit ni les images GHCR ni les CNAME.
Il est décrit dans [Cycle de vie : template](10-cycle-template.md#controle-dorphelins).

Un projet qui a encore un import ne se supprime pas : il faut d'abord supprimer ses imports.

## Après la PR standard

Depuis le 20/11/2026, une app importée passe au modèle des templates par une PR dans son repo
(CNP-303). EpiTweet a fait le cycle complet le 20/11 : PR 3-Istor/epitweet#1, puis release
`v0.1.0`. Seuls les repos de l'organisation 3-Istor sont éligibles.

![Repo importé après sa PR standard](../01-architecture/diagrams/lifecycle-import-pr.light.png#only-light)
![Repo importé après sa PR standard](../01-architecture/diagrams/lifecycle-import-pr.dark.png#only-dark)

- Après le premier déploiement réussi, le CMP ouvre une PR `cnp/standard-deploy-<sha8>` dans le
  repo importé. Elle ajoute `deploy/values-<app>.yaml` pour chaque app, les Dockerfiles réparés
  par l'analyseur dans `cnp/docker/`, la CI des templates dans `.github/workflows/ci.yml`, plus
  `security.yml`, `dependabot.yml` et `cnp-verify.yml`. Elle ne touche pas au code de l'app.
- Tous ces workflows tournent sur les runners éphémères des imports, jamais sur un runner
  persistant. La CI construit au plus deux images à la fois. Le scan d'image de `security.yml`
  est coupé : l'analyseur scanne déjà les images, puis Trivy dans le cluster.
- Tant que la PR n'est pas fusionnée, rien ne change. Les apps tournent sur la copie de la
  plateforme.
- À la fusion, un seul commit dans cnp-projects pose `repoURL` sur chaque app, supprime son
  `apps/<p>/<app>/values.yaml` et écrit `apps/<p>/<app>/release.yaml` : l'image importée,
  épinglée par digest jusqu'à la première release.
- Argo CD lit alors `deploy/values-<app>.yaml` sur la branche par défaut du repo importé, puis
  `release.yaml` dans cnp-projects. La release ne peut pas vivre dans une autre branche du même
  repo : Argo CD refuse deux révisions d'un même repo dans une application.
- `network.allowFrom` reste calculé depuis le registre et écrit en ligne : le repo ne peut pas
  l'élargir.
- Une release est un tag `vX.Y.Z` du repo importé. La CI publie et signe
  `ghcr.io/3-istor/imports/<p>/<app>/ci:X.Y.Z`. Un Image Updater écrit ce tag dans
  `release.yaml` de cnp-projects. La plateforme n'écrit jamais dans le repo importé.
- Le rebuild de la plateforme s'arrête. `rebuild` et `apply` répondent 409.

Après la fusion, on change une valeur d'une app importée comme pour une app de template : par
un commit dans `deploy/values-<app>.yaml` du repo importé. On la met en production par un tag.

### CNP-304 : schéma fermé et écritures tracées

En service depuis le 20/11/2026 : infra-templates#36, CMP#268,
template-app-webapp-python-fastapi-react#77, template-html-css#23, cnp-docs#57.

- `values.schema.json` du chart devient fermé. Une clé inconnue ou une valeur hors bornes fait
  échouer la CI de l'app et le rendu d'Argo CD.
- Le chart refuse au rendu l'image d'une autre app et une release qui dépasse le quota du
  namespace.
- Le CMP valide le fichier contre ce schéma avant de commit. Ses commits deviennent
  `chore(cmp): update <file> (<user>)`, une ligne par valeur changée, avec un trailer
  `CNP-Changed-By`.
- Les template apps gagnent un workflow `Deploy values` qui vérifie `deploy/`.
- Le CMP#268 ne touche que les apps de template. Les apps importées restent refusées par le
  panneau Configuration. Une fois CNP-303 en place, le même contrôle pourra servir pour elles.
