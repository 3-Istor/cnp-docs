# Architecture cible

Tout ce qui est décrit ici est **proposé**, rien n'est déployé. Les décisions D1 à D21 ont été
écrites, puis confrontées au code par une revue contradictoire indépendante (33 constats, dont
3 bloquants). La version ci-dessous intègre ses corrections ; les points encore ouverts sont dans
[Décisions de Brian](00-index.md#decisions-de-brian-1411).

![Architecture cible de l'observabilité](../01-architecture/diagrams/observability-target.light.png#only-light)
![Architecture cible de l'observabilité](../01-architecture/diagrams/observability-target.dark.png#only-dark)

## En une phrase

Trois bases VictoriaMetrics (métriques, logs, traces), deux collecteurs légers, une seule chaîne
d'alerte vers Discord, Grafana pour les dashboards, et le portail qui montre à chaque projet ses
signaux sans qu'il code. Le tout pour moins de RAM qu'aujourd'hui.

## Avant, après

| Fonction | Aujourd'hui | Cible |
| --- | --- | --- |
| Collecte des métriques | Alloy (DaemonSet, 1,8 Gio) | **vmagent** (1 Deployment) |
| Collecte des logs | Alloy | **Fluent Bit** (DaemonSet léger) |
| Métriques | VictoriaMetrics cluster (4 composants) | **VMSingle** |
| Logs | Loki + gateway nginx, S3 RGW | **VictoriaLogs** single, volume Ceph |
| Traces | rien | **Collecteur OpenTelemetry → VictoriaTraces** |
| Alertes | vmalert, ruler Loki, règles Grafana, CMP, Gatus | **vmalert (métriques + logs) → Alertmanager → Discord** ; le CMP et Gatus gardent leurs événements, au même format |
| Isolation par projet | vmauth (métriques), tenants Loki (logs) | **vmauth pour les trois signaux**, réservé aux lectures des projets |
| Déploiement | charts Helm séparés | **opérateur VictoriaMetrics**, objets VMSingle, VMAgent, VMAlert, VMUser… |
| Dashboards | Grafana, IDs gnet | Grafana, JSON générés et versionnés, un dossier par rôle |

## Décisions

### Socle

**D1 · Opérateur VictoriaMetrics, verrouillé.** Chart de l'opérateur seul et nos propres objets
(VM v1.156). `victoria-metrics-k8s-stack` va chercher ses règles par défaut sur GitHub au moment du
déploiement, hors de git : les règles retenues sont copiées dans le dépôt.
Notre Grafana reste, scrapes etcd, scheduler et controller-manager
désactivés (k3s). Le convertisseur des CRDs Prometheus est limité aux namespaces plateforme.
`selectAllByDefault: false` et sélecteurs de namespace explicites : l'opérateur ne lit les objets
que dans les namespaces plateforme et dans `cnp-observability`, un namespace dont **le CMP est le
seul écrivain**. Règles par défaut en liste blanche (node-exporter, kube-state-metrics, santé VM,
Alertmanager) : le reste du mixin Kubernetes sonne en permanence sur 2 nœuds. Argo CD en
`ServerSideApply` (CRDs trop grandes pour l'annotation client).
*Pourquoi l'opérateur* : les scrapes, règles et utilisateurs deviennent des objets Kubernetes que le CMP
écrit, versionnés et visibles, au lieu d'un Secret patché.

**D2 · VMSingle remplace le mode cluster.** Rétention 30 jours (14 aujourd'hui) **après** le régime
de D3, volume 40 Gio, limite 1,5 Gio (vmstorage utilise déjà 997 Mio), limites de requête
`-search.maxUniqueTimeseries`, `-search.maxMemoryPerQuery`, `-search.maxQueryDuration`.
Sauvegarde quotidienne `vmbackup` vers le RGW Ceph. Alerte sur l'espace disque libre.
**Les anciens noms de Service sont gardés** (`…-vmselect`, `…-vminsert`, `…-vmauth`) et pointent
sur la nouvelle stack avec réécriture des chemins : aucun projet n'a à rejouer son Terraform,
les clouds et Argo Rollouts continuent de marcher, et le retour arrière est un changement de
sélecteur.

**D3 · Régime de cardinalité.** Suppression des buckets d'histogrammes apiserver et kubelet, du
conteneur shutdown-manager d'Envoy et du double scrape de shared-gateway. On garde
`envoy_cluster_endpoint_rq_*`, indispensable au canary. Limites `sample_limit` par scrape et
`-remoteWrite.maxDailySeries`. Cible : moins de 150 000 séries, mesurées avec l'explorateur de
cardinalité avant et après.

**D4 · Collecte.** vmagent (un Deployment, `honorLabels` désactivé, le vrai namespace posé par la
découverte : un label `namespace` exposé par un pod devient `exported_namespace`) pour les
métriques. Fluent Bit (DaemonSet) pour les logs des pods et le journal d'audit k3s.
*Pourquoi Fluent Bit et pas vlagent* (constaté au développement, 14/11) : vlagent mélange le JSON
de l'app et les métadonnées Kubernetes dans le même enregistrement, sans préfixe, donc une app peut
forger `kubernetes.pod_namespace` ou le label de projet. Un filtre Lua de Fluent Bit reconstruit
chaque enregistrement à partir des seules métadonnées et range le contenu de l'app sous `log` et
`json.*`. vmauth filtre ensuite sur le champ `namespace`, jamais sur `project`. **Un seul schéma de champs** pour tous les clusters : `namespace`, `project`, `cluster`, `app`.
Sur les clouds, vmagent garde un tampon disque (`-remoteWrite.tmpDataPath`) quand le tunnel
WireGuard tombe : c'est le besoin FinOps de CNP-110. Alloy disparaît partout.

**D5 · Logs.** VictoriaLogs single, rétention globale 30 jours, plafond disque
(`-retention.maxDiskSpaceUsageBytes`). Dimensionné pour environ 1 Go par jour (98 % de logs
plateforme), avec un régime : exclusion des lignes debug et des namespaces bavards au collecteur.
Les classes 7, 14 et 30 jours disparaissent : seuls vault et keycloak les utilisaient, en 30 jours.
Pas de S3 tant que VictoriaLogs ne le supporte pas en stable. Les journaux d'audit gardent leur
archive Object Lock de 180 jours ; l'export horaire passe de `logcli` à LogsQL après 48 h de
double export comparé ligne à ligne.

### Isolation et sécurité

**D6 · Un tenant de logs, isolation par vmauth.** Même modèle que les métriques : chaque projet a
un utilisateur vmauth qui force le filtre sur ses namespaces. **Tous les arguments de filtre sont
imposés par vmauth**, comme le demande sa documentation, et aucun ne vient du client. Un test de CI
vérifie qu'un projet ne lit rien d'un autre.

**D7 · vmauth pour les projets seulement.** Deux réplicas et un PDB. Chaque utilisateur n'a accès
qu'aux chemins de **lecture** (`url_map` avec liste blanche) : VMSingle, VictoriaLogs et
VictoriaTraces servent aussi l'écriture et l'administration sur le même port. Clés d'auth sur les
endpoints admin. Les consommateurs plateforme (vmalert, Argo Rollouts, OpenCost, CMP, org 1 de
Grafana) lisent les bases directement, protégés par NetworkPolicy : une panne de vmauth ne doit
pas faire avorter tous les canaries en cours.

**D8 · NetworkPolicy par brique, dans chaque vague.** Les tenants ont un egress ouvert par
conception : chaque nouvelle base est joignable par défaut. Écrivains : vmagent, Fluent Bit, vmalert,
relais WireGuard. Lecteurs : vmauth, vmalert, Grafana, CMP, OpenCost, Rollouts. Alertmanager :
vmalert et le CMP seulement (C3).

**D9 · Les tenants ne peuvent pas créer d'objets d'observabilité.** Les groupes
`operator.victoriametrics.com`, `monitoring.coreos.com` et `opentelemetry.io` sont réservés à la
plateforme (C2). Les scrapes et règles par app sont écrits par le CMP dans
`cnp-observability`, jamais par les charts des apps. Aucun impact sur le travail quotidien :
aucun projet n'en utilise aujourd'hui.

### Clouds publics

![Observabilité des clouds publics](../01-architecture/diagrams/observability-multicloud.light.png#only-light)
![Observabilité des clouds publics](../01-architecture/diagrams/observability-multicloud.dark.png#only-dark)

**D22 · Une porte, une identité par cloud.** EKS et GKE reçoivent les mêmes collecteurs que
l'on-prem (vmagent, Fluent Bit, un collecteur OpenTelemetry en mode agent), installés par Argo CD
depuis le dépôt Cloud. Tout ce qu'un cloud envoie ou lit passe par **vmauth**, derrière le relais
18080 (écritures) ou 18083 (lectures), avec **un utilisateur par cloud** dont le mot de passe vient
de Vault. vmauth ajoute lui-même `cluster=aws|gcp` : `extra_label` sur les métriques,
`VL-Extra-Fields` sur les logs, un en-tête que la passerelle OpenTelemetry transforme en attribut
sur les spans. Un cloud ne peut donc pas écrire au nom d'un autre cluster, et les lectures
d'Argo Rollouts et d'OpenCost restent bornées à leur cloud. Aujourd'hui, les relais vont droit sur
vminsert et vmselect sans authentification : le pare-feu de namespaces du cloud est la seule
barrière.
*Pourquoi un collecteur de traces sur le cloud* : la passerelle de l'on-prem associe un span à
son pod par l'adresse source, que le relais masque. Le collecteur du cloud fait cette association
sur place, avec l'API du cloud.
*Clouds éteints* : vmagent et Fluent Bit gardent une file sur disque (2 Gio) pendant une coupure du
tunnel et la vident au retour ; `cnp.sh down` attend que les files soient vides. Deux alertes :
`CloudTelemetryMissing` (Argo CD dit le cloud allumé, aucun échantillon depuis 10 minutes) et
`CloudQueueBacklog` (file au-delà de 500 Mo). Les alertes de nœud, aujourd'hui limitées à
`cluster="onprem"`, s'appliquent aux clouds en `warning`. Alertmanager groupe aussi par `cluster`.
*Relais* : 18080 et 18083 vont vers vmauth, 18082 continue d'alimenter Loki jusqu'à la fin de V2
puis disparaît. Le relais 18081 (on-prem vers OpenCost du cloud) ne change pas.

### Traces et instrumentation

**D10 · Un seul chemin de traces.** SDK OpenTelemetry dans le template FastAPI (et OpenTelemetry
natif de Quarkus pour les apps importées) → **un collecteur OpenTelemetry** → VictoriaTraces.
Le collecteur associe chaque span au pod **par la connexion** et supprime tout attribut `k8s.*`,
`project` ou `service.namespace` envoyé par le client : un tenant ne peut pas se faire passer pour
un autre. Il retire aussi les query strings et les en-têtes sensibles. Repli si VictoriaTraces
(pré-1.0) casse : Tempo monolithique derrière le même collecteur.
**Lecture des traces (Q9, tranché le 14/11)** : **un tenant VictoriaTraces par projet**. Le
collecteur range chaque span dans le tenant de son projet (`AccountID` = identifiant numérique du
projet) grâce à une table de routage que le CMP écrit, comme il écrit déjà la config de vmauth. Les
orgs Grafana des projets ont une datasource Jaeger qui passe par vmauth, et vmauth **impose**
l'en-tête `AccountID` de l'utilisateur : le client ne peut pas le choisir. `/select/tenant_ids`,
`/delete`, `/internal` et LogsQL sont fermés à ces utilisateurs.
*Pourquoi pas un filtre comme pour les métriques* : les API Jaeger et Tempo de VictoriaTraces
ignorent `extra_filters` (issue VictoriaTraces #178, correctif #184 non fusionné). Un filtre posé par
vmauth n'isolerait rien, même pas la recherche par identifiant de trace.
*Coût* : un pipeline et un exporteur par projet dans le collecteur, et un redémarrage progressif
à la création d'un projet, sans coupure : deux réplicas, `maxUnavailable: 0`, le nouveau pod est
prêt avant que l'ancien s'arrête, l'arrêt vide les files en mémoire, et les SDK réessaient au
pire quelques secondes. Le projet créé n'a pas encore de trace à perdre. Le routage par ressource
est en alpha. Recette : la même trace
écrite dans deux tenants, la trace d'un autre projet demandée par son identifiant, un en-tête
`AccountID` forgé : tout doit revenir vide.

**D11 · Templates instrumentés.** FastAPI : `/metrics`, traces OpenTelemetry, logs JSON avec
`trace_id`. Chart générique : rendu des probes et des `podAnnotations` (C7). Postgres : le CMP
déclare le scrape du cluster CNPG de chaque app. Ce qu'une app gagne est listé dans
[Offre aux projets](04-offre-projets.md).

**D12 · OBI (eBPF) en spike limité dans le temps.** Métriques RED sans code pour les apps
importées. Porte de sortie : coexistence avec Cilium et noyau 6.8 vérifiés, RAM mesurée, sinon
abandon. Découverte limitée aux namespaces de projets, jamais vault, keycloak, `*-system` ni le CMP.
Pas d'OpenTelemetry Operator : aucune app Java, Node ou .NET dans le catalogue, et son API est
encore en v1alpha1.

**D13 · Carte des services = Hubble.** Métriques Hubble par workload (déjà prévues par le chantier
import, CNP-242) et Hubble UI en option. Une seule carte, alimentée par le réseau, valable pour
toutes les apps.

### Produit

**D14 · Le SLO comme produit.** Chaque app de prod avec une route HTTP reçoit un SLO par défaut :
disponibilité 99,5 % et latence p95 < 500 ms, sur 7 jours pour les apps de démo, 28 jours sinon.
Généré par le CMP avec la bibliothèque Go de Sloth en VMRule (enregistrement + alertes
multi-fenêtres 14,4 / 6 / 1 du SRE Workbook), avec un seuil de trafic minimum et une prise en compte
d'offhours-guard (une app endormie ne brûle pas son budget). La porte du canary reste une requête
distincte, dérivée du même modèle : un SLI par route ne sait pas séparer canary et stable.
L'owner règle la cible dans Settings.

**D15 · Une seule chaîne d'alerte.** Voir [Alerting](06-alerting.md).

**D16 · Le portail d'abord, Grafana pour creuser.** Voir [Dashboards et vues](05-dashboards.md).

**D17 · Dashboards versionnés.** JSON générés et commités dans K3s, un dossier par rôle dans
l'org 1. Dans les orgs projet, **un seul propriétaire par objet** : Terraform crée l'org, le CMP
crée datasources et dashboards. Pas de SDK Grafana Foundation (pré-1.0).

**D18 · Marqueurs de déploiement et DORA.** Argo CD notifications pose une annotation Grafana et
incrémente un compteur par app. Le CMP calcule d'abord la fréquence de déploiement et le taux
d'échec des changements, le reste ensuite.

**D19 · FinOps.** Coût par 1 000 requêtes (OpenCost / requêtes Envoy) et gaspillage par app
(requests contre p95 réel), en reprenant `recommendations_real.py`.

### IA

**D20 · Observabilité de la landing zone IA et traçabilité.** Remplace Tempo, Alloy et Loki dans
la spec de la landing zone (ses décisions D7, D14, D24, D30). Voir [IA et conformité](07-ia-et-conformite.md).

**D21 · IA bon marché, en lecture seule.** Le serveur MCP du CMP gagne `get_app_metrics`,
`query_app_logs`, `list_app_traces`, `get_alerts`, `get_slo`, avec les droits de l'utilisateur. On
n'expose pas `mcp-victoriametrics` qui lirait sans filtre de projet. Pas de résumé d'alerte par un
modèle (décision Q7 du 14/11) : le message Discord porte déjà le contexte préparé (dernier
déploiement, redémarrages, erreurs fréquentes) sans dépenser de tokens.

## Budget mémoire et CPU

Estimations à confirmer par la mesure. Requests, en Mio.

| Brique | Aujourd'hui (utilisé) | Cible (request) |
| --- | --- | --- |
| Alloy | 1 826 | 0 |
| Loki + gateway | 916 | 0 |
| VictoriaMetrics cluster | 1 096 | 0 |
| VMSingle | | 1 024 |
| vmagent | | 256 |
| Fluent Bit ×2 | | 96 |
| VictoriaLogs | | 512 |
| vmalert ×2, Alertmanager, vmauth ×2, opérateur | 44 | 300 |
| kube-state-metrics, node-exporter, Grafana | 309 | 400 |
| **Socle** | **≈ 3 300 utilisés** | **≈ 2 600 demandés** |
| Collecteur OpenTelemetry ×2 + VictoriaTraces | | 512 |
| Hubble relay, métriques en plus | | 150 |
| **Avec les nouveautés** | | **≈ 3 250 demandés** |

- **Pic de double fonctionnement** : pendant 14 jours en vague 1 et 2, ancienne et nouvelle stack
  coexistent, soit environ +1,5 à 2 Gio. On ne lance la vague 2 qu'après avoir retiré Alloy des
  métriques en vague 1.
- **CPU** : le nœud principal est à 82 %. Requests et limites CPU sur chaque brique, `GOMAXPROCS`
  sur les binaires VictoriaMetrics.
- **La migration ne règle pas les OOM de pae-node-3** : la RAM d'une VM de 20 Gio déjà touchée ne
  redescend pas côté hyperviseur.
- La vague 4 (traces, OBI) ne démarre qu'avec les gains des vagues 1 et 2 mesurés.
