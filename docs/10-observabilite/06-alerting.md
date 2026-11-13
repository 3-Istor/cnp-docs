# Alerting : une seule chaîne, moins de bruit

La fatigue d'alerte est le premier frein cité par les équipes en 2026. Aujourd'hui la CNP a quatre
chemins vers Discord avec quatre formats. Cible : **une chaîne pour les règles, un format pour tous
les messages, un routage qui suit l'app**.

![Une alerte de burn rate, de la règle à Discord](../01-architecture/diagrams/observability-alert-flow.light.png#only-light)
![Une alerte de burn rate, de la règle à Discord](../01-architecture/diagrams/observability-alert-flow.dark.png#only-dark)

## La chaîne

1. **vmalert** évalue deux familles de règles : métriques (VMSingle) et logs (VictoriaLogs, type
   `vlogs`). Deux objets VMAlert, un par source.
2. Les règles plateforme sont écrites à la main dans K3s. Les règles des projets (SLO, catalogue
   d'alertes du portail) sont des **VMRule écrites par le CMP** dans `cnp-observability`, chaque
   groupe portant `params.extra_filters` sur les namespaces du projet : même un modèle mal écrit
   ne lit pas les données d'un autre projet.
3. **Alertmanager** route selon les labels `app`, `project` et `severity` :
   app → webhook de l'app, sinon du projet, sinon de la plateforme (le comportement actuel des règles
   Grafana est conservé).
4. **Discord** reçoit un message unique par groupe d'alertes : titre, gravité, app, valeur, lien
   runbook, lien vers l'app dans le portail, lien Grafana, bouton de sourdine (lien portail).

Les événements du CMP (santé, FinOps, sécurité) et de Gatus restent émis par eux, mais adoptent le
même format et les mêmes salons.

## Salons et gravités

| Gravité | Exemples | Salon | Notification |
| --- | --- | --- | --- |
| critical | Budget d'erreur brûlé 14× en 1 h, base indisponible, chaîne d'alerte morte | `#alertes-critiques` du projet ou de la plateforme | `@here`, push ntfy en option |
| warning | Burn rate 6× sur 6 h, disque à 80 %, redémarrages | `#alertes` | Message |
| info | Budget brûlé 1× sur 3 jours, secret de plus de 90 jours | Résumé quotidien | Un message par jour |

## Règles de qualité

- **`runbook_url` et `summary` obligatoires** sur chaque règle, vérifiés en CI avec `vmalert-tool`
  et ses tests unitaires.
- **Inhibitions** : un nœud en panne masque les alertes des pods qu'il portait ; une app endormie par
  offhours-guard ne déclenche ni indisponibilité ni burn rate.
- **Seuil de trafic minimum** sur les alertes de taux, pour ne pas sonner sur 2 requêtes.
- **Règles par défaut du chart en liste blanche** : le mixin Kubernetes complet sonne en permanence
  sur un cluster de 2 nœuds.
- **Qualité mesurée** : un dashboard montre les déclenchements par règle sur 7 jours et le temps de
  résolution. Une règle qui sonne tous les jours sans action est corrigée ou supprimée.
- **Sourdine depuis le portail** : le CMP est le seul client d'Alertmanager (NetworkPolicy). Il
  impose le label `project` dans chaque silence et refuse les regex sur ce label.

## Ce que deviennent les alertes Grafana des projets

Aujourd'hui le CMP crée des règles gérées par Grafana dans l'org du projet, et les éditeurs peuvent
y ajouter les leurs. Migration en trois temps :

1. Le catalogue du portail écrit des VMRule. L'onglet Alertes lit l'état dans vmalert au lieu de
   Grafana.
2. Les règles Grafana du catalogue sont supprimées une fois que leur équivalent a sonné au moins
   une fois dans Discord.
3. Les règles personnalisées des projets restent dans Grafana (inventaire au moment de la
   migration). Le bouton « créer une règle » continue de fonctionner.

## Savoir que la chaîne d'alerte est morte

Une règle `Watchdog` sonne en permanence. Alertmanager l'envoie toutes les 5 minutes à un service
de **dead man's switch** externe, qui prévient Discord quand les pings s'arrêtent. C'est le seul
moyen d'apprendre que vmalert ou Alertmanager sont tombés : en octobre, vmstorage est resté en
lecture seule 5 jours sans que personne le sache.

Choix proposé : healthchecks.io, gratuit jusqu'à 20 contrôles. Alternative auto-hébergée : un
Gatus hors du cluster (sur pae-node-1), plus de travail. À trancher.

## On-call et PagerDuty

Pas d'outil d'astreinte : Grafana OnCall OSS est archivé, Opsgenie ferme en avril 2027, Grafana IRM
gratuit est limité à 3 utilisateurs en cloud. Pour une équipe de 7 sur une plateforme de démo,
**Discord avec `@here` sur les critiques, et ntfy en push mobile en option**, suffisent.
