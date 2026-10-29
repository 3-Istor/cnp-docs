# Index des schémas

Quel ticket a changé quoi, et quel schéma le montre. Tous les schémas sont générés depuis le code avec noodle : sources dans [`architecture/cnp`](https://github.com/3-Istor/cnp-docs/tree/main/architecture/cnp), procédure dans [`architecture/README.md`](https://github.com/3-Istor/cnp-docs/blob/main/architecture/README.md). État au 29/10/2026. EKS et GKE ayant la même plateforme, les schémas les dessinent une seule fois (« cluster cloud public »).

## Par public

| Public | Commencer par |
|---|---|
| Soutenance, décideurs | [Qui utilise quoi](01-system-overview.md#who-uses-what), [Observabilité en un coup d'œil](07-observability.md), [Où tourne quoi](../06-multicloud/06-schemas.md#ou-tourne-quoi) |
| Dev d'une équipe projet | [Un projet et une app sur un cloud public](../06-multicloud/06-schemas.md#un-projet-et-une-app-sur-un-cloud-public), [Promotion et canary](../03-pipelines-and-workflows/06-release-canary.md), [Sauvegarde des bases](08-database-backups.md) |
| Équipe plateforme | [Liens hub / clouds](../06-multicloud/06-schemas.md#liens-hub-clouds), [Qui peut atteindre un projet](../06-multicloud/06-schemas.md#qui-peut-atteindre-un-projet), [Collecte](07-observability.md#collecte), [`cnp.sh up`](../06-multicloud/06-schemas.md#cnpsh-up) |

## Par ticket

| Ticket | Périmètre | Composants | Schémas |
|---|---|---|---|
| CNP-79, 102, 104 | Clusters EKS et GKE enregistrés dans Argo CD | Argo CD, API EKS/GKE, Vault | `multicloud`, `multicloud-links`, `cloud-up` |
| CNP-103, 111 | Tunnels WireGuard, endpoint publié dans un bucket | Pods WireGuard, bucket `cnp-wireguard-endpoints` | `multicloud-links`, `cloud-up` |
| CNP-105, 106 | Socle projet et exposition des apps sur les clouds | `cnp-project-base`, Envoy Gateway, cloudflared, NetworkPolicy | `multicloud`, `multicloud-policies` |
| CNP-107 | Choix AWS / GCP à la création d'un projet | CMP, registre `cnp-projects` | `project-create-cloud` |
| CNP-112 | Synchro cloud qui reprend après un `down` | Applications Argo CD (retry sans fin) | `cloud-up` |
| CNP-128, 160, 162 | État des clusters dans le CMP, « non connecté » | CMP, Application `wireguard-<cloud>`, Secret `cluster-<cloud>` | `cmp-cloud-read`, `cloud-up` |
| CNP-110, 163, 164 | Observabilité des clouds, garde-fous | Alloy, relais 18080/18082/18083, NetworkPolicy `onprem-apps`, ValidatingAdmissionPolicy | `observability`, `observability-read`, `observability-metric-path`, `multicloud-links`, `multicloud-policies` |
| CNP-166 | API EKS/GKE via le proxy SOCKS on-prem | `cnp.sh`, `pae-node-1`, `api_allowed_cidrs` | `multicloud-links`, `cloud-up` |
| CNP-120, 121 | Journaux d'audit, archive immuable | Alloy, Loki, ruler, bucket Object Lock | `observability`, [Journaux d'audit](06-audit-logging.md) |
| CNP-122 | Logs et métriques isolés par projet, alertes | vmauth, tenants Loki, orgs Grafana, Alertmanager | `observability-read`, `observability-simple` |
| CNP-118, 140 | Sauvegarde et restauration des bases | CloudNativePG, plugin barman, Ceph RGW, CMP | `app-backup`, `app-restore` |
| CNP-123, 126 | Staging, promotion, canary | CMP, Argo Rollouts, HTTPRoute, vmselect | `canary-release` |
| — | Vue d'ensemble des acteurs | Dev, admin projet, finance, opérateur | `platform-actors` |
| — | Ce qu'un projet reçoit, et une app vue de l'intérieur | `<p>-system`, namespaces d'apps, AppProjects, realm, mount Vault, org Grafana | `project-anatomy`, `app-runtime` |

Pas encore représenté : OpenCost sur les clouds (CNP-89, en cours).

## Écarts relevés entre la doc et le code

Chaque schéma a une carte *Gaps*. Les principaux :

- La création d'un projet n'est pas bloquée quand le cluster cible est down, seul le déploiement d'une app l'est.
- `vminsert` accepte les écritures des clouds sans authentification ; la Loki gateway n'a pas de NetworkPolicy.
- Sauvegardes CNPG : HTTP clair vers Ceph RGW, identifiants générés par Rook (pas Vault), indisponibles sur EKS/GKE.
