# Réglages d'une app : CMP et YAML du dépôt (CNP-304)

Un développeur qui gère son dépôt n'a pas besoin du CMP. Le CMP sert aux débutants, le YAML aux habitués du code, et les deux écrivent **le même fichier** : `deploy/values.yaml` (app statique) ou `deploy/values-frontend.yaml` et `deploy/values-backend.yaml` (app fullstack), plus `deploy/env/<env>/<composant>.yaml` quand l'app a un staging. Argo CD rend ce fichier avec le chart `cnp-generic-app` d'infra-templates.

Ticket : [CNP-304](https://3istors.atlassian.net/browse/CNP-304).

## Principe

- **Un seul fichier, deux éditeurs.** Le CMP lit le fichier du dépôt et y écrit par un commit sur `main`. Il n'a pas de copie à côté : un réglage qui vit dans le YAML n'est stocké nulle part ailleurs.
- **Mêmes règles des deux côtés.** `values.schema.json` du chart est fermé. La CI du dépôt, le CMP avant de commiter et Argo CD au rendu appliquent le même schéma, lu sur `main` d'infra-templates.
- **Traçable.** Chaque modification faite dans le CMP est un commit dans le dépôt de l'app, qui nomme le fichier, l'auteur et chaque valeur modifiée.

## Inventaire

Légende de la colonne YAML : **oui** la clé existe dans le fichier de l'app et le chart la lit ; **ajouté** la clé est ajoutée ou bornée par CNP-304.

### Réglages éditables dans le CMP et dans le YAML

| Réglage | Clé YAML | Où dans le CMP | YAML |
| --- | --- | --- | --- |
| Réplicas | `replicaCount` (0 à 10) | Configuration | oui, borné |
| Ressources | `resources.requests/limits` (`cpu`, `memory`) | Configuration, FinOps (recommandation appliquée) | oui, formats bornés |
| Autoscaling | `autoscaling.*` (max 10) | Configuration | oui, borné |
| Version de l'image | `image.tag`, `image.digest` | Releases, Configuration | oui |
| Variables d'environnement | `env[]` (`name`, `value` en chaîne, `valueFrom`) | Configuration | oui, borné |
| ConfigMap | `configMap.*` | Configuration | oui |
| Service | `service.*` (`ClusterIP` seul) | Configuration | oui, borné |
| Sondes | `livenessProbe`, `readinessProbe`, `startupProbe` | Configuration | oui |
| Supervision | `monitoring.*` | Configuration | oui |
| Métriques et traces | `metrics.*`, `tracing.*` | Configuration | oui, ajouté côté CMP |
| Rétention des logs | `logRetention` (`7d`, `14d`, `30d`) | Configuration | oui, borné |
| Annotations et labels de pod | `podAnnotations`, `podLabels` | Configuration | oui |
| Exposition, SSO, groupes | `ingress.enabled`, `ingress.sso_protected`, `ingress.allowedGroups` | Sécurité (préréglages), Configuration | oui |
| Base de données | `db.enabled`, `db.name`, `db.storage`, `db.metrics` | Configuration | oui, borné |
| Pooler PgBouncer | `db.pooler.*` | Configuration | oui, ajouté côté CMP |
| Sauvegardes de la base | `db.backup.enabled`, `keepAfterDelete`, `retentionPolicy` | Sécurité (sauvegardes) | oui, borné |
| Restauration | `db.restore.*` | Sécurité (restauration) | oui, écrit par le CMP |
| Bucket S3 | `bucket.enabled` | Configuration | oui, ajouté côté CMP |
| Mise en veille (prod) | `offhours.*` (cron à 5 champs) | Configuration | oui, borné |
| Mise en veille (staging) | `offhours.*` de `deploy/env/staging/<composant>.yaml` | Environnements | oui |
| Stratégie de release, canary | `release.*` | Release | oui, borné |
| Migrations | `migrations.enabled`, `command` (Release), `timeout`, `resources` (Configuration) | Release, Configuration | oui, borné |
| Appelants autorisés | `network.allowFrom[]` | Configuration | oui, ajouté côté CMP |
| Secrets attendus | `secrets.keys[]` (noms seulement) | Configuration, panneau Secrets | **ajouté** |
| Blocage de la CI sécurité | `deploy/security.yaml` | Sécurité (réglages) | oui |

`release.*`, `migrations.enabled/command`, `db.backup.*` et `db.restore.*` ne passent plus par le formulaire générique du CMP : leur écran dédié vérifie ce qu'une fusion de YAML ne voit pas (canary en cours, règles du chart pour la release, bucket de sauvegarde d'un cloud public, source d'une restauration). Le formulaire générique renvoie vers cet écran. Le YAML reste modifiable directement, sous le schéma.

### Réglages du YAML sans écran dans le CMP

`ingress.hostname`, `ingress.publicPaths`, `workload.*` (apps importées), `image.repository`, `secrets.vaultPath`. Ils ne sont pas éditables dans le CMP : le dépôt est leur seul éditeur, sous le schéma et les contrôles de rendu.

### Réglages qui restent dans le CMP seulement, volontairement

| Réglage | Où il vit | Raison |
| --- | --- | --- |
| Politique de promotion (suites, soak, approbation, promotion auto) | Base du CMP | C'est un garde-fou de gouvernance : ceux qui poussent dans le dépôt ne doivent pas pouvoir désactiver l'approbation qui les contrôle. Le droit d'écrire dans le dépôt n'est pas le rôle admin du projet. |
| Activer ou retirer le staging | Base du CMP, registre cnp-projects, Terraform | Provisionne un client OIDC, un realm, un chemin Vault et un namespace, sous un plafond par projet (`maxApps`). Les fichiers `deploy/env/*` qui en résultent sont, eux, dans le dépôt. |
| Liens entre apps (`dependsOn`) | Registre cnp-projects | Un lien crée un objet dans le namespace de l'appelant et ouvre le réseau de la cible : le dépôt de l'appelant ne peut pas accorder ce que la cible n'a pas accepté. Côté cible, `network.allowFrom` est dans le YAML. |
| Budget du projet | Base du CMP | Réglage financier du projet, réservé au propriétaire. |
| Règles d'alerte, objectifs SLO | Base du CMP, Grafana | Réconciliées avec Grafana par le CMP. Candidat à un passage dans le YAML (`alerting`, `slo`) dans un ticket suivant. |
| Protection contre la suppression, accès client | Registre cnp-projects, Keycloak | Réglages de projet, pas d'app, réservés aux admins humains. |
| Membres | Keycloak | Identités, pas de la configuration. |
| Exceptions de sécurité | Base du CMP, rendues dans `deploy/security-exceptions.yaml` | Chaque exception porte un demandeur et une piste d'audit ; le fichier du dépôt est généré depuis la base. |
| Webhook d'alertes sécurité | Vault | C'est un secret. |
| Veille, réveil, pause, redémarrage | Statut de l'OffhoursSchedule | Actions ponctuelles, pas des réglages. |
| Sauvegarde à la demande | Ressource CNPG `Backup` | Action ponctuelle. |
| Niveau d'autonomie de l'assistant | Base du CMP | Gouvernance de l'assistant. |
| Icône, favoris, langue, notifications | Base du CMP, S3 | Préférences d'interface. |
| Auto-rebuild d'un import | Base du CMP | Propre aux imports (CNP-303). |

## Secrets : les noms dans le YAML, les valeurs dans Vault

```yaml
secrets:
  enabled: true
  keys: [STRIPE_API_KEY, SMTP_PASSWORD]
```

`secrets.keys` déclare les noms attendus, jamais les valeurs : le schéma refuse un élément qui n'est pas un nom de variable (`API_KEY=sk_live...` échoue). Le panneau Secrets du CMP compare ces noms au chemin Vault de l'app et affiche les manquants avec un lien vers Vault. Une valeur de secret ne va jamais dans `env[].value`.

## Ce qui est refusé avant le déploiement

| Contrôle | Où | Effet |
| --- | --- | --- |
| Clé inconnue, valeur hors schéma | `values.schema.json` | Échec de la CI du dépôt, refus du CMP (422) avant le commit, échec du rendu Argo CD (l'ancienne version reste en place). |
| Image d'une autre app | Rendu du chart (`cnp-generic-app.checkImageOwner`) | Une image `ghcr.io/3-istor/...` doit être `ghcr.io/3-istor/<app>[/...]` ou `ghcr.io/3-istor/imports/<projet>/...`. Même règle que la politique Kyverno `restrict-tenant-images`, qui reste le contrôle à l'admission. |
| Quota du namespace dépassé | Rendu du chart (`templates/quota-report.yaml`) et `ci/check-app-values.sh` | Voir ci-dessous. |

### Le quota : pourquoi au rendu

Chaque namespace d'app a le même ResourceQuota (cnp-project-base, `templates/namespace.yaml`) : `requests.cpu: 1`, `requests.memory: 2Gi`, `pods: 10`.

- Un maximum dans le schéma ne suffit pas : le quota porte sur une somme (réplicas fois requêtes, plus la base, le pooler, le Job de migration, le surplus d'un rolling update ou d'un canary), qu'un schéma JSON ne calcule pas.
- Un contrôle par l'API du CMP ne couvrirait que les modifications faites dans le CMP, pas un push dans le dépôt.
- Le rendu du chart voit tout le fichier et tourne à la fois dans la CI du dépôt et dans Argo CD. Le chart calcule le pic de la release et échoue au-delà du quota. Le quota y est recopié en constantes (`cnp-generic-app.quota`), pas en valeurs, pour qu'une app ne puisse pas relever son propre plafond. Si le quota change dans cnp-project-base, ces constantes changent dans le même lot.
- Une app fullstack a deux releases dans le même namespace : aucun rendu seul ne voit la somme. `ci/check-app-values.sh` rend les deux composants et additionne leurs pics.

### La CI du dépôt

Les templates d'app ont un workflow `Deploy values` (`.github/workflows/deploy-values.yml`), lancé sur les PR et les push qui touchent `deploy/`. Il récupère infra-templates sur `main` et lance `ci/check-app-values.sh`, qui rend chaque composant et chaque environnement comme l'ApplicationSet `cnp-apps`. Les contrôles vivent dans infra-templates : le dépôt de l'app suit le chart avec lequel Argo CD déploie.

## Traçabilité des écritures du CMP

```text
chore(cmp): update deploy/values.yaml (alice)

- replicaCount
- resources.requests.cpu

CNP-Changed-By: alice
```

Aucun commit quand rien ne change. Les écrans dédiés (release, sauvegardes, exposition, staging) gardent leurs propres messages de commit, toujours dans le dépôt de l'app.

## Deux éditeurs : les règles

- Le CMP écrit dans le dépôt, jamais à côté. Les réglages de la liste « CMP seulement » n'ont pas de clé YAML, donc pas de double source.
- Le CMP envoie le SHA du fichier lu : si le dépôt a bougé entre la lecture et l'écriture, il refuse (409) et relit.
- Le CMP refuse ce que le schéma refuserait, et ne change que les valeurs de sa liste ; le reste du fichier (commentaires compris) est conservé.

## Limites connues

- Les dépôts déjà générés gardent la CI de leur création : ils n'ont pas le workflow `Deploy values` tant qu'on ne le leur ajoute pas. Le schéma et les contrôles de rendu s'appliquent quand même à leur synchronisation Argo CD.
- infra-templates est privé : le workflow lit le chart avec le secret d'organisation `CNP_INFRA_TEMPLATES_TOKEN` (lecture seule du contenu d'infra-templates).
- Une modification du CMP est un commit sur `main` : la CI tourne en même temps que la synchronisation. Le CMP valide le schéma avant de commiter ; le quota et le propriétaire de l'image sont contrôlés par le rendu Argo CD, qui refuse sans rien déployer.
