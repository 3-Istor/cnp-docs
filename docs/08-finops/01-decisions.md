# Décisions FinOps

Signée le 2026-10-26. La numérotation continue celle de
[D-01 à D-13](../06-multicloud/02-decisions.md). D-14 regroupe les huit choix D-F1 à D-F8
du plan FinOps : ils forment un seul plan de données, donc une seule fiche.

---

## D-14 — Plan de données FinOps

**Décision.** Huit choix, qui gardent chacun leur identifiant du plan.

| Id | Choix |
|---|---|
| D-F1 | **OpenCost est le moteur d'allocation, pas le produit.** Une instance par cluster ventile le coût par namespace, pod et label, et mesure l'inoccupation. La CMP reste l'interface, les budgets, le RBAC et les exports. |
| D-F2 | **Un seul format pivot : FOCUS.** Toutes les sources sont normalisées en lignes FOCUS 1.2 dans la table `cost_line_items`, avec des colonnes `x_` pour ce que FOCUS ne couvre pas. Les exports sortent de cette table. |
| D-F3 | **La facture fait foi pour le montant, OpenCost pour la répartition.** Un cluster AWS ou GCP coûte ce que dit son export de facturation. OpenCost dit qui l'a consommé. L'écart est affiché, jamais caché. |
| D-F4 | **Le coût on-prem vient d'un TCO versionné dans Git.** Électricité mesurée, amortissement, réseau. Le résultat est généré comme tarif `customPricing` d'OpenCost. |
| D-F5 | **Deux colonnes d'économies : évitée et encaissée.** Endormir des pods libère de la capacité allouée. La facture ne baisse que si un nœud disparaît. Les deux ne sont jamais additionnées. |
| D-F6 | **GreenOps mesuré à la source on-prem, estimé dans le cloud.** Puissance mesurée, intensité carbone RTE, avec la méthode affichée à côté de chaque chiffre. |
| D-F7 | **Les recommandations passent par le pipeline GitOps existant.** Une PR, déployée en canary par Argo Rollouts, avec retour arrière automatique. Un humain approuve. Hors périmètre tant qu'il n'y a pas un mois de données réelles. |
| D-F8 | **L'IA est comptée au prix public, depuis la gateway.** Tokens × tarif publié, même quand l'usage est gratuit ou payé par des crédits. La facture réelle ne sert qu'à la réconciliation. |

**Alternative écartée.** Réécrire l'allocation en PromQL dans la CMP, ou acheter un outil
FinOps (Kubecost, Vantage, CloudZero). Le premier réimplémente l'inoccupation, les volumes
et les pods éphémères, ce qu'OpenCost fait déjà et que la CNCF maintient. Les seconds sont
payants, sortent les données de la plateforme et ne connaissent pas l'on-prem.

**Conséquences.**

- Il y a trois sources de montant, jamais mélangées sans le dire : la facture
  (`invoice`), le prix public (`public_price`, l'IA) et le TCO (`tco`, l'on-prem). Chaque
  ligne porte sa base dans `x_pricing_basis`, affichée au survol.
- Les lignes OpenCost d'un cloud sont au prix catalogue en USD. Elles ne servent qu'à
  répartir la facture et ne sont jamais comptées dans un total.
- Un prix on-prem change par une PR sur `finops/onprem-tco.yaml`, comme `target_cloud`
  change par un commit (D-01). L'historique et la revue sont gratuits.
- Tant qu'une valeur du TCO est moquée, les coûts bâtis dessus sont marqués « estimé ».
- Sans autoscaler de nœuds, l'économie encaissée dans le cloud vaut 0 €. Le tableau de
  bord l'affiche tel quel. C'est une exigence à ajouter au
  [contrat avec l'équipe cluster](../06-multicloud/03-contrat-equipe-cluster.md).
- Le critère d'acceptation de WS-9 est enrichi : réconciliation par fournisseur et non
  alloué sous 5 %.

**Écarts entre le plan et le code mergé.** Trois points du plan du 16 octobre ont été précisés à la mise en œuvre.

| Plan | Code mergé | Raison |
|---|---|---|
| Postgres pour `cost_line_items` | La base CMP est en SQLite en production. Les tables fonctionnent sur les deux moteurs. | Volume faible, quelques centaines de lignes par jour. |
| VictoriaMetrics en rétention d'au moins 90 jours | Rétention de 14 jours | L'historique des coûts vit dans la base CMP. L'ingestion rejoue J-3 et le p95 des recommandations porte sur 14 jours. |
| Ingesteur en CronJob | Boucle asynchrone dans le backend CMP, comme les autres collecteurs | Un composant de moins à déployer. Activée par `FINOPS_INGEST_ENABLED`. |
