# Runbook · Budget d'erreur qui brûle (SLO)

Votre app consomme son budget d'erreur plus vite que prévu. Le message Discord dit à quel rythme
et combien de temps il reste avant d'épuiser le budget de la période.

## Ce que l'alerte veut dire

Chaque app de prod a deux objectifs par défaut : 99,5 % de requêtes sans erreur 5xx et 95 % des
requêtes servies en moins de 500 ms, sur 7 jours (projets de démo) ou 28 jours. Le CMP génère
trois alertes par objectif :

| Alerte | Rythme | Fenêtres | Sévérité | Budget épuisé en |
| --- | --- | --- | --- | --- |
| Brûlure rapide | 14,4 fois | 1 h et 5 min | critical | environ 2 jours sur 28 |
| Brûlure moyenne | 6 fois | 6 h et 30 min | critical | environ 5 jours sur 28 |
| Brûlure lente | 1 fois | 3 jours et 6 h | warning | la fin de la période |

L'alerte ne sonne pas sous 60 requêtes par heure ni pendant que l'app dort (offhours-guard).

## Vérifier

1. Ouvrez le lien **Portail** du message : l'onglet Overview montre le taux d'erreur, la latence
   et le budget restant, avec les déploiements en marqueurs.
2. Un déploiement juste avant le début ? Comparez canary et stable dans l'onglet Releases.
3. Onglet Logs : les erreurs les plus fréquentes de la dernière heure, groupées par motif. Cliquez
   une ligne pour ouvrir sa trace.
4. Une dépendance lente (base, autre app) ? L'onglet Traces montre où part le temps.

## Agir

- **Version fautive** : revenez à la version précédente depuis Releases. Si le canary tournait,
  il a peut-être déjà avorté de lui-même.
- **Charge inhabituelle** : vérifiez les limites mémoire et CPU dans Settings, et les redémarrages
  dans Overview.
- **Dépendance** : prévenez son équipe ; l'alerte reste ouverte tant que le taux ne redescend pas.
- **Objectif trop strict** pour cette app : l'owner le règle dans Settings, onglet SLO.

## Mettre en sourdine

Le lien **Sourdine** du message ouvre le portail. La sourdine ne vaut que pour votre projet et
dure au plus 7 jours. Elle n'arrête pas la brûlure du budget : elle évite seulement de répéter le
message pendant que vous corrigez.
