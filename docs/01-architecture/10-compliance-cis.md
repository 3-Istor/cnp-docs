# Conformité CIS : preuves hebdomadaires

Chaque semaine, la plateforme on-prem est comparée au CIS Kubernetes Benchmark, au guide
NSA-CISA et à MITRE ATT&CK. Les rapports sont rangés dans l'archive d'audit verrouillée
(voir [Journaux d'audit](06-audit-logging.md#archive-immuable)) : ce sont les preuves
de l'item A1. Le code est dans `K3s/k8s/config/compliance/`.

## Ce qui tourne

| CronJob (`compliance`) | Outil | Ce qui est vérifié | Quand (UTC) |
| --- | --- | --- | --- |
| `kube-bench-server` | kube-bench v0.16.0, profil `k3s-cis-1.9` | serveur `my-k3s-cluster-node` : cibles `master`, `controlplane`, `node`, `policies` | dimanche 03:00 |
| `kube-bench-agent` | kube-bench v0.16.0, profil `k3s-cis-1.9` | agent `k3s-worker-node` : cible `node` | dimanche 03:00 |
| `kubescape` | Kubescape v4.0.20, regolibrary v2.0.39 | objets de l'API : `cis-v1.12.0`, `nsa`, `mitre` | dimanche 03:30 |

Les images sont épinglées par digest. Un job en échec déclenche l'alerte vmalert
`ComplianceScanFailed`.

Un job kube-bench par rôle de nœud, placé par `nodeSelector` (serveur) ou affinité (tout
nœud sans rôle `control-plane`) : le cluster a un serveur et un agent. Avec un deuxième
agent, il faudra une copie de `kube-bench-agent` par nœud, épinglée par
`kubernetes.io/hostname`.

## Où sont les résultats

Bucket RGW `cnp-audit-archive`, préfixe `compliance/` :

```text
compliance/kube-bench/AAAA-MM-JJ/<nœud>.json      rapport complet (--json, sorties des tests incluses)
compliance/kubescape/AAAA-MM-JJ/onprem.json       rapport complet (format json)
compliance/summary/AAAA-MM-JJ/<outil>-<cible>.json  compteurs pour le portail
```

Le résumé donne `pass`, `fail`, `warn` (et `info` pour kube-bench) avec la version de
l'outil et le chemin du rapport. Pour Kubescape, les compteurs portent sur les contrôles
et `warn` compte les contrôles `skipped` (revue manuelle, ou type de ressource non lu),
avec le score de conformité par framework.

Les objets suivent les règles du bucket : Object Lock COMPLIANCE 180 jours, expiration à
190 jours. L'utilisateur RGW `compliance-scans` ne peut qu'écrire sous `compliance/` (bucket
policy posée par le job `audit-archive`) : il ne lit ni ne supprime rien. Pour lire un
rapport, utiliser l'utilisateur `audit-archive` :

```bash
aws --endpoint-url http://192.168.1.110:7480 s3 ls s3://cnp-audit-archive/compliance/summary/ --recursive
```

Point zéro, scan Kubescape manuel du 13/11/2026 avec la même configuration : score
51,5 % (CIS 48,4 %, NSA 61,2 %, MITRE 58,0 %).

## Pourquoi un namespace privilégié

Les contrôles CIS des nœuds lisent les drapeaux des processus k3s, les droits des fichiers
sous `/var/lib/rancher` et `/etc/rancher`, et le journal systemd de l'unité k3s. Le pod
kube-bench a donc `hostPID`, des montages `hostPath` en lecture seule (dont `/` sur `/host`),
tourne en root et garde deux capacités : `SYS_CHROOT` (lancer le `journalctl` du nœud, absent
de l'image) et `DAC_READ_SEARCH`. Aucun niveau Pod Security sous `privileged` ne le permet.

Ce qui l'isole :

- le namespace `compliance` ne contient que ces trois CronJobs, sans Service, entrées
  réseau refusées par NetworkPolicy ;
- le conteneur d'envoi, non root et sans capacité, est le seul à voir le secret RGW ;
- le pod Kubescape respecte le niveau `restricted` (vérifié en dry-run), le label
  `pod-security.kubernetes.io/audit: restricted` signale tout écart ;
- les ClusterRoles sont en lecture seule (`get`, `list`). Celui de Kubescape ne lit pas les
  Secrets ; celui de kube-bench ne lit que pods, service accounts et objets RBAC.

## Écarts acceptés

### Contrôles que K3s traite autrement

D'après les notes de remédiation du profil `k3s-cis-1.9` et la configuration du serveur :

| Contrôles | Pourquoi |
| --- | --- |
| 1.1.1 à 1.1.8 | API server, controller manager, scheduler et etcd tournent dans le processus k3s : pas de manifeste de pod statique. kube-bench les marque `skip`. |
| 1.1.11, 1.1.12, section 2 (etcd) | Le serveur unique stocke son état dans SQLite (kine), pas dans etcd. La cible `etcd` n'est pas lancée. À revoir si le cluster passe en etcd embarqué. |
| 4.1.1, 4.1.2, 4.1.9, 4.1.10 | Le kubelet est embarqué : pas de fichier de service ni de fichier de configuration. |
| 4.2.7 | K3s pose `--hostname-override`, ce que le CIS admet. |
| 4.1.3, 4.1.4, 4.3.1 et Kubescape C-0291 | kube-proxy est désactivé (`--disable-kube-proxy`), Cilium le remplace. |
| 1.1.9, 1.1.10 | Les fichiers CNI sont ceux de Cilium ; un fichier de verrou en 750 est attendu. |

### Choix de la plateforme

| Contrôles | Pourquoi |
| --- | --- |
| 1.2.17, 1.2.18 (`audit-log-maxage`, `audit-log-maxbackup`) | Le fichier d'audit local n'est qu'un tampon (2 jours, 3 fichiers) : Alloy l'envoie dans Loki (30 jours) puis dans l'archive verrouillée (180 jours). |
| 1.2.11 (`AlwaysPullImages`) | Le profil K3s le classe « permissif » : chaque démarrage de pod dépendrait du registre. |
| Version du profil | kube-bench v0.16.0 n'a pas de profil K3s plus récent que `k3s-cis-1.9` (CIS 1.9, Kubernetes 1.27) ; le cluster est en 1.34. Les objets de l'API sont couverts en CIS 1.12 par Kubescape. |
| Kubescape C-0212 (CIS 5.6.4) | Le scanner ne lit pas les Secrets : la partie « Secrets dans `default` » n'est pas évaluée. |
| Contrôles d'hôte Kubescape | `--host-scan=false` : ils déploieraient un capteur privilégié ; kube-bench couvre les nœuds. |

### Écarts réels, à corriger

Les drapeaux du serveur (journal k3s du 12/11/2026) annoncent ces `FAIL`. Ce ne sont pas des
écarts acceptés :

- 1.2.27, 1.2.28 : pas de chiffrement des Secrets au repos (`--secrets-encryption`) ;
- 1.2.9 : pas de plugin d'admission `EventRateLimit` ;
- 1.2.3 : `DenyServiceExternalIPs` n'est pas activé.

## Lancer un passage à la main

```bash
kubectl -n compliance create job kube-bench-server-manual --from=cronjob/kube-bench-server
kubectl -n compliance create job kube-bench-agent-manual --from=cronjob/kube-bench-agent
kubectl -n compliance create job kubescape-manual --from=cronjob/kubescape
kubectl -n compliance logs job/kubescape-manual -c upload
```

Le conteneur `upload` affiche le résumé. Relancer le même jour crée une nouvelle version
des objets ; l'ancienne reste verrouillée.

## Limites

- Kubescape télécharge la regolibrary épinglée depuis GitHub à chaque passage : sans accès
  sortant, le job échoue.
- Les contrôles qui lisent le journal (drapeaux des composants) dépendent de la ligne
  « Running kube-apiserver » écrite au démarrage de k3s. Le journal est persistant sur les
  nœuds, mais une rotation après un long uptime peut faire passer ces contrôles en `WARN`.
- Comme pour l'archive d'audit, un admin Ceph peut effacer le bucket avec `radosgw-admin`.
