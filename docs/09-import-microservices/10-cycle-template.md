# Cycle de vie : app créée par template

Cette page suit une app créée depuis un template, de sa création à sa suppression. Elle décrit
ce qui est sur `main` le 20/11/2026 : CMP, app-templates `templates/k3s-gitops-app`, ApplicationSets
de K3s, chart infra-templates et offhours-guard. Le parcours d'une app importée est dans
[Cycle de vie : import](11-cycle-import.md). Les services de données sont dans
[Providers](12-providers.md).

Le projet existe déjà. Son bootstrap a créé le realm Keycloak `<p>`, le montage Vault
`project-<p>`, le tunnel Cloudflare et l'enregistrement `registry/projects/<p>.yaml` dans
cnp-projects. Le module Terraform `project-bootstrap` ne crée aucun repo GitHub ni AppProject :
Argo CD les tire du registre.

![Cycle de vie d'une app créée par template](../01-architecture/diagrams/lifecycle-template.light.png#only-light)
![Cycle de vie d'une app créée par template](../01-architecture/diagrams/lifecycle-template.dark.png#only-dark)

## Qui fait quoi

| Acteur | Rôle |
| --- | --- |
| Admin du projet | Crée l'app, change sa configuration, coupe une release, la supprime |
| Membre du projet | Endort et réveille l'app |
| CMP backend | Lance Terraform, écrit dans le repo de l'app et dans le registre |
| Argo CD | Lit le registre et le repo, rend le chart, synchronise le cluster |
| CI du repo de l'app | Construit et signe l'image à chaque tag |
| Image Updater | Suit les tags semver et les réécrit dans le repo |
| offhours-guard | Endort les workloads la nuit, les réveille le matin |
| Admin plateforme | Consulte le contrôle d'orphelins |

## Création

L'admin appelle `POST /api/deployments` depuis le portail ou via MCP. Le CMP choisit le type
selon le template : `fullstack` pour `template-app-webapp-python-fastapi-react`, `static` sinon.
Il applique ensuite le module Terraform `k3s-gitops-app`. L'état est rangé sous
`cnp/projects/<p>/<app>`.

| Ressource | Nom | Créée par |
| --- | --- | --- |
| Repo GitHub | `3-Istor/<app>`, privé, généré depuis le template | Terraform |
| Fichiers de valeurs | `deploy/values.yaml` pour static. `deploy/values-frontend.yaml` et `deploy/values-backend.yaml` pour fullstack | Terraform |
| CI de sécurité | `.github/workflows/security.yml`, secret Actions `DISCORD_WEBHOOK_URL` | Terraform |
| Client Keycloak | `cnp-<p>-<app>` dans le realm `<p>` | Terraform |
| Secrets Vault | `project-<p>/<app>` : `username`, `password`, `client-secret`. `project-<p>/<app>/registry` : accès GHCR | Terraform |
| Politique Vault | `project-<p>-app-<app>-policy`, lecture de son seul sous-arbre | Terraform |
| Rôle Vault | `<p>-<app>-role` sur le montage `kubernetes`, ou `kubernetes-<cloud>` hors site | Terraform |
| DNS | CNAME `<app>-<p>.3istor.com` vers le tunnel du projet | Terraform |
| Entrée du registre | `{name, type, repoURL, hostnames}` dans `spec.apps` de `registry/projects/<p>.yaml` | CMP, après Terraform |
| Application Argo CD | `<p>-<app>-app`, ou `<p>-<app>-frontend` et `<p>-<app>-backend` | ApplicationSet `cnp-apps` |
| Namespace | `<p>-<app>`, avec `ResourceQuota` `cnp-project-quota` et `LimitRange` | Application `<p>-bootstrap`, chart cnp-project-base |
| Image Updater | Un par composant, dans `argocd` | ApplicationSet `cnp-image-updaters` |
| Base CNPG | `Cluster` et sauvegarde quotidienne vers Ceph RGW. Active pour le backend fullstack | Chart, bloc `db` |
| Bucket S3 | `ObjectBucketClaim`, désactivé par défaut | Chart, bloc `bucket` |
| Planning de veille | `OffhoursSchedule` `<app>-schedule`, actif par défaut | Chart, bloc `offhours` |
| Secret dans le pod | `VaultSecret` `<app>-secrets`, lu par `envFrom` | Chart |

Argo CD rend le chart infra-templates avec le fichier de valeurs du repo de l'app. Les
paramètres `cnp.project`, `cnp.app`, `cnp.environment`, `cnp.type` et `project_name` gagnent
sur le fichier. Le repo ne peut donc pas changer le chemin Vault, le rôle ni le secret de
registre de l'app. Le chemin Vault et le rôle sont dérivés de ces paramètres. Tout autre chemin
fait échouer le rendu.

## Où vivent les valeurs et les secrets

| Je veux changer… | Où | Comment |
| --- | --- | --- |
| Réplicas, ressources, variables d'environnement, base, veille | `deploy/values*.yaml` du repo `3-Istor/<app>` | Panneau Configuration du CMP, ou commit sur `main` |
| La version déployée | Tag `vX.Y.Z` du repo | « Couper une release » dans le CMP. L'Image Updater écrit le tag dans le fichier |
| Un secret | Vault `project-<p>/<app>` | Saisi dans le CMP. Un agent le demande avec `request_secret` |
| Le nom public | Registre cnp-projects, champ `hostnames` | Fixé à la création, avec le CNAME |

Le panneau Configuration appelle `PATCH /api/deployments/{id}/config`. Il fusionne les clés
autorisées dans le fichier et commit sur `main`. Les clés autorisées sont dans
`app_config_policy.py`. Argo CD synchronise chaque commit sur `main`, avec prune et selfHeal.

## Mise à jour

1. L'admin coupe une release `vX.Y.Z`. Le CMP crée le tag sur le repo.
2. La CI du repo construit l'image `ghcr.io/3-istor/<app>:X.Y.Z` et la signe.
3. L'Image Updater suit les tags semver `X.Y.Z`. Il écrit le nouveau tag dans
   `deploy/values*.yaml` sur `main`.
4. Argo CD synchronise.

Avec un staging, l'updater suit les tags `X.Y.Z-rc.N` et écrit
`deploy/env/staging/<composant>.yaml`. La promotion vers la prod passe par le CMP. Le staging
n'est pas détaillé ici.

## Veille

Chaque app a son `OffhoursSchedule`. Les templates l'activent avec sommeil à 01:00, réveil à
07:00, fuseau Europe/Paris. offhours-guard met à zéro le Rollout, sinon le Deployment, sinon le
StatefulSet. Il garde le nombre de réplicas d'origine pour le réveil. La base CNPG ne dort pas.

Un membre du projet peut endormir ou réveiller l'app tout de suite, depuis le CMP. Le CMP écrit
alors `status.currentState` sur les plannings du namespace prod. Il répond 409 si l'app n'a pas
de planning. La page `offhours-<p>.3istor.com` fait la même chose. Elle exige une connexion au
realm du projet et l'appartenance à un groupe du projet.

## Suppression

L'admin appelle `DELETE /api/deployments/{id}` avec un step-up. Une app protégée contre la
suppression est refusée.

1. Le CMP supprime d'abord le staging s'il existe.
2. `terraform destroy` supprime le repo avec son historique, le client Keycloak, les secrets,
   la politique et le rôle Vault, le CNAME et les images GHCR.
3. Le CMP retire l'app du registre.
4. Argo CD supprime les Applications. cnp-project-base supprime le namespace, la base CNPG et
   ses volumes compris.
5. Le CMP supprime les états Terraform. La ligne de l'app reste en base, au statut `deleted`.

Une suppression interrompue repasse en échec au redémarrage du CMP. La relancer reprend où elle
s'était arrêtée.

### Contrôle d'orphelins

`GET /api/orphans` est réservé aux admins plateforme. Il est en lecture seule. Il compare les
projets et apps vivants du CMP avec :

- les enregistrements et entrées du registre ;
- les namespaces portant le label `cnp.3istor.com/project`, sur chaque cloud ;
- les dossiers KV sous `project-<p>` ;
- les rôles Vault des montages `kubernetes*` ;
- les politiques Vault ;
- les états Terraform sous `cnp/projects/`.

Une source illisible remonte dans `errors`. Le même contrôle existe en script :
`python -m app.scripts.list_orphans`.

Limite connue : les états du bootstrap de projet sont rangés sous `cmp/<cloud>/projects/`. Le
contrôle ne les voit pas. Une image GHCR que GitHub refuse de supprimer reste sans bruit.
