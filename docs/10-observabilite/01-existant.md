# Existant : ce qui tourne le 13/11/2026

Mesures prises en lecture seule sur le cluster de production et sur `origin/main` des dépôts K3s,
CMP, app-templates, infra-templates, cnp-project-base et Cloud.

## La stack

| Brique | Version | Rôle | RAM utilisée | Limite |
| --- | --- | --- | --- | --- |
| Grafana Alloy (DaemonSet ×2) | chart 1.11.1 | Logs, audit k3s, tous les scrapes, tenants Loki, compteurs de logs | 877 + 949 Mi | 1000 Mi ×2 |
| Loki SingleBinary + gateway nginx | chart 7.3.0 | Logs multi-tenant (un tenant par projet), ruler sécurité | 904 + 12 Mi | 2 Gi |
| VictoriaMetrics cluster | chart 0.48.0 | vminsert, vmselect, vmstorage, vmauth | 55 + 164 + 871 + 6 Mi | 256 / 256 / 1024 / 128 Mi |
| vmalert | chart 0.46.0 | 29 règles métriques | 20 Mi | 128 Mi |
| Alertmanager | chart 1.42.0 | Deux receivers Discord sur le même webhook | 24 Mi | 128 Mi |
| Grafana | chart 10.5.15 (12.3) | Org 1 + une org par projet | 90 Mi | 256 Mi |
| kube-state-metrics, node-exporter | 2.18 | Métriques Kubernetes et nœuds | 134 + 85 Mi | |
| Gatus | 5.37 | Statut plateforme + un Gatus par projet | | |
| OpenCost | | Coût par namespace | 82 Mi | |

**Total observabilité : environ 3,3 Gio utilisés** sur un cluster de 2 nœuds (20 + 8 Gio, 2 × 4 vCPU).
Le nœud principal est à 61 % de mémoire et **82 % de CPU**.

## Les volumes

| Mesure | Valeur | Commentaire |
| --- | --- | --- |
| Séries actives | environ 307 000 | `sum(scrape_samples_scraped)` |
| dont kubelet + apiserver | 175 000 (57 %) | Surtout des histogrammes que personne ne lit |
| dont Envoy (8 pods) | 43 700 | Scrapés deux fois pour shared-gateway, et aussi sur le conteneur shutdown-manager |
| Disque vmstorage | 12,5 Gio pour 14 jours | Le volume de 10 Gio a déjà été plein : VictoriaMetrics en lecture seule 5 jours en octobre |
| Logs envoyés par Alloy | environ 930 Mo/jour (compressé) | Dont 20 Mo/jour seulement pour les apps des projets : 98 % sont des logs plateforme |
| Namespaces avec une rétention de logs explicite | 2 (vault, keycloak, 30 j) | Aucune app n'utilise le label `cnp.3istor.com/log-retention` |

## Le « mix bizarre », point par point

1. **Deux moteurs, deux chemins de logs croisés.** Les logs sont dans Loki, mais les alertes sur les
   logs tournent dans VictoriaMetrics à partir de compteurs fabriqués par Alloy
   (`loki_process_custom_*`). Et une règle Loki surveille VictoriaMetrics.
2. **Alloy fait tout** : logs, audit, tous les scrapes, lecture des PodMonitor, routage des tenants.
   C'est le plus gros consommateur de RAM de la stack.
3. **Quatre canaux d'alerte vers Discord** : vmalert et le ruler Loki via Alertmanager, les règles
   Grafana créées par le CMP pour les projets, les messages directs du CMP (santé, FinOps, sécurité),
   et Gatus. Formats, seuils et silences différents.
4. **Nommage Prometheus partout** alors qu'il n'y a pas de Prometheus : datasource « Prometheus »,
   CRDs prometheus-operator installées seulement pour ServiceMonitor, label `release: kube-prometheus-stack`.
5. **La stack ne se surveille pas** : vmstorage, vminsert, vmselect, vmalert et Loki ne sont pas scrapés.
6. **Les apps n'exposent rien.** Le template FastAPI n'a ni `/metrics`, ni logs JSON, ni OpenTelemetry.
   Le chart générique ne rend ni `podAnnotations` ni les probes déclarées dans ses values.
7. **Aucune trace distribuée** ni carte des services, en dehors des refus réseau de Hubble.

## Constats de l'inventaire à corriger tout de suite (vague 0)

| Id | Constat | Gravité |
| --- | --- | --- |
| C1 | Isolation des métriques entre projets à durcir dans la configuration vmauth générée par le CMP. Détail dans un canal privé tant que ce n'est pas corrigé. | Haute |
| C2 | Les objets d'observabilité (CRDs de l'opérateur, ServiceMonitor, PodMonitor) doivent être réservés à la plateforme avant d'installer l'opérateur. Détail dans un canal privé. | Haute |
| C3 | Alertmanager doit être réservé à vmalert et au CMP par NetworkPolicy. | Moyenne |
| C4 | vmselect OOM-killé le 13/11 par une seule requête large (limite 256 Mi), vmstorage à 997 Mi sur 1 Gi. Aucune limite de requête (`-search.max*`). | Moyenne |
| C5 | Envoy scrapé deux fois pour shared-gateway : `sum()` double les requêtes de ces routes dans l'onglet Overview. | Moyenne |
| C6 | Pods de test Helm Kyverno en Error depuis 208 jours ; sidecar Gatus en CrashLoopBackOff dans `iot-system` et `self-runner-system` (2 725 redémarrages). | Basse |
| C7 | Chart générique : probes et `podAnnotations` jamais rendues. | Basse |
| C8 | Modèle de la doc en retard : OpenCost « prévu », règle Grafana auto-assign, vues observabilité absentes de l'index. | Basse |
