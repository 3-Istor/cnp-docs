# Méthodes de calcul

Cette page dit comment chaque nombre est calculé. Les valeurs viennent de
`finops/onprem-tco.yaml` au 2026-10-26. Une valeur marquée `mock: true` dans ce fichier est
un substitut en attendant une mesure : les coûts qui en dépendent portent `x_is_estimate`
et l'interface les affiche « estimé ».

## 1. Le coût horaire de l'on-prem

```text
coût_horaire = énergie_h + amortissement_h + réseau_h

énergie_h      = watts_mesurés / 1000 × prix_kWh_TTC
amortissement_h = Σ prix_achat / (durée_amortissement_mois × 730)
réseau_h       = abonnement_internet × part_plateforme / 730
```

730 heures est un mois moyen. Les chiffres du TCO actuel :

| Terme | Valeur | Origine |
|---|---|---|
| Puissance mesurée | 141,3 W (35,2 + 42,4 + 63,7) | Mesurée. CSV `powerstat` toutes les 10 min sur les trois nœuds, moyenne d'octobre 2026 |
| Prix de l'électricité | 0,1845 €/kWh TTC | Tarif, daté dans le fichier |
| Amortissement | 0 € | Réel. Les trois NUC7i7BNH sont offerts |
| Abonnement internet | 30 €/mois | **Moqué** |
| Part imputée à la plateforme | 50 % | **Moqué** |

Résultat :

```text
énergie_h = 141,3 / 1000 × 0,1845          = 0,02607 €/h
réseau_h  = 30 × 0,5 / 730                 = 0,02055 €/h
coût_horaire                               = 0,0466 €/h, soit environ 34 €/mois
```

L'électricité pèse 56 % de ce coût. Le reste est la part de la box internet, donc la partie
moquée.

### Du coût horaire au tarif OpenCost

Le cluster K3s offre 8 vCPU, 27,3 Gio de RAM et 366 Gio de stockage utile (capacité brute
Ceph de 1098 Gio en trois réplicas). Le coût horaire se répartit ainsi : 65 % CPU, 30 % RAM,
5 % stockage. Cette clé suit le rapport des prix par défaut d'OpenCost pour un nœud de cette
forme.

| Ressource | Calcul | Tarif injecté |
|---|---|---|
| CPU | 0,0466 × 0,65 / 8 | 0,00378770 €/vCPU·h |
| RAM | 0,0466 × 0,30 / 27,3 | 0,00051228 €/Gio·h |
| Stockage | 0,0466 × 0,05 / 366 | 0,00000637 €/Gio·h |

Le générateur Go écrit ces tarifs dans `k8s/config/opencost/values-pricing.yaml`, fichier
marqué « ne pas éditer ». La CI exécute `go -C finops run ./tco -check` et échoue si ce
fichier n'est plus celui que produit le TCO. OpenCost ne voit donc jamais un tarif qui ne
vient pas de Git.

Valeurs à confirmer : la part de la box, la consommation moyenne du serveur GPU (80 W) et la
capacité Ceph utile (366 Gio, non marquée `mock` mais à vérifier).

## 2. Coût d'un namespace

OpenCost calcule chaque jour, par namespace, le coût de chaque ressource à partir du plus
grand de la requête et de l'usage. L'ingestion écrit une ligne FOCUS par namespace et par
ressource (CPU, RAM, GPU, volumes, réseau, load balancers). Sur l'on-prem, ce montant est
le coût lui-même, car OpenCost le calcule avec le tarif du TCO. L'inoccupation du cluster
devient le pseudo-projet `__idle__`.

## 3. Facture, répartition et réconciliation

Pour AWS et GCP, le montant vient de la facture et la répartition vient d'OpenCost.

1. L'export de facturation est agrégé par jour. Les nœuds du cluster, sans tag projet, sont rangés sous `__cluster__`. Le reste est affecté par tag.
2. Pour chaque jour, la facture des nœuds est répartie sur les namespaces au prorata de leur coût OpenCost du même jour. Le dernier namespace absorbe l'arrondi : la somme des parts égale exactement la facture des nœuds.
3. Les lignes `__cluster__` d'un jour partagé sont ignorées dans les totaux, remplacées par les lignes `<cloud>_split`. Les montants des lignes de facture ne sont jamais modifiés.
4. Sans donnée OpenCost pour un jour, aucun partage n'est écrit et les nœuds restent comptés comme non alloués.

L'identité de la réconciliation, par fournisseur et par mois :

```text
facture = alloué + inoccupation + plateforme + non alloué + écart
```

`/finops/reconciliation` renvoie chaque terme et `unallocated_pct`. L'écart vaut 0 sauf si
des lignes manquent ou sont incohérentes : c'est une ligne visible, pas un ajustement
silencieux. Le non alloué est constitué des lignes sans projet et des nœuds d'un jour non
partagé. La même route liste les ressources sans tag projet les plus coûteuses.

Objectif de gouvernance : non alloué sous 5 % de la facture, par fournisseur. Le code calcule
et expose ce pourcentage. Aucune alerte automatique ne le surveille aujourd'hui.

Choix d'implémentation : l'inoccupation et la plateforme sont montrées à part, pas
redistribuées au prorata. Le plan recommandait le prorata. Ce point reste ouvert.

### Devises

Le montant d'origine est conservé dans `billed_cost` et `billing_currency`. `x_cost_eur`
est converti avec le dernier taux BCE publié à la date de la ligne ou avant. Si aucun taux
n'a jamais été stocké, un taux de repli de 1,10 USD par euro s'applique et la ligne est
marquée estimée. Un échec de lecture du flux BCE garde les derniers taux connus.

## 4. L'IA au prix public

Pour chaque jour, la gateway fournit les tokens par projet, agent (`azp`), modèle et
backend, en entrée et en sortie.

```text
prix_public_usd = (tokens_entrée × prix_entrée + tokens_sortie × prix_sortie) / 1 000 000
x_cost_eur      = prix_public_usd / taux_USD_par_EUR
```

Les prix par million de tokens sont dans `finops/ai-prices.yaml`, avec leur date d'effet. Le
prix retenu est celui en vigueur le jour des tokens. Un changement de tarif est une nouvelle
entrée, l'ancienne reste pour l'historique.

| Champ de la ligne | Valeur | Raison |
|---|---|---|
| `x_cost_eur`, `list_cost` | Prix public | C'est ce que les budgets et les exports montrent |
| `billed_cost`, `effective_cost` | 0 | Niveau gratuit ou crédits prépayés : rien n'est facturé par appel |
| `x_pricing_basis` | `public_price` | Affiché au survol |

Un modèle absent du fichier de prix produit une ligne à 0 marquée estimée, avec un
avertissement dans les logs.

Les modèles servis par le serveur GPU du homelab (`local_backends`, aujourd'hui `ollama`)
n'ont pas de prix public. Leur coût est l'énergie du serveur, répartie sur les tokens servis
ce jour :

```text
coût_jour_gpu = watts_moyens × 24 / 1000 × prix_kWh        (80 W moqués, soit 0,35 €/jour)
part_ligne    = coût_jour_gpu × tokens_ligne / tokens_locaux_du_jour
```

Les appels qui passent par la gateway ne sont pas recomptés dans la ligne SaaS du
fournisseur.

## 5. Les économies d'offhours-guard

Deux colonnes, jamais additionnées (D-F5). Le registre `savings_ledger` reçoit une ligne par
jour, projet et application, dans la catégorie `offhours`.

```text
fraction_endormie = moyenne sur la journée de offhours_target_asleep, par namespace
heures_endormies  = fraction_endormie × 24

évité (€)    = heures_endormies × coût_horaire_des_requêtes_du_namespace
encaissé (€) = énergie_évitée_kWh × prix_kWh
énergie_évitée_kWh = heures_endormies × cœurs_actifs_éveillés × 15 W / 1000
```

- Le coût horaire des requêtes est la somme des requêtes CPU et mémoire du pic de la journée, au tarif du TCO. Des pods endormis ne demandent rien, donc le pic est la taille éveillée.
- Les cœurs actifs éveillés sont le CPU réellement consommé, ramené aux seules heures d'éveil.
- 15 W par cœur actif est la pente mesurée de la puissance contre l'occupation CPU. Il vient de `watts_per_busy_core` dans le TCO.
- Le registre ne couvre que l'on-prem aujourd'hui. Dans le cloud, l'encaissé vaudra 0 tant qu'aucun autoscaler ne retire de nœud, et l'évité n'a pas encore de ligne.
- Le CO2e évité est l'énergie évitée multipliée par l'intensité du jour (section suivante).

Sur l'on-prem, l'économie encaissée est uniquement énergétique, donc faible mais mesurable.

## 6. Énergie et carbone

### Énergie

Sur l'on-prem, la part électrique du coût d'une ligne OpenCost donne son énergie :

```text
énergie_kWh = coût × (énergie_h / coût_horaire) / prix_kWh
```

La part électrique est de 56 % avec les valeurs actuelles. L'énergie est donc une
attribution proportionnelle au coût OpenCost, donc à l'usage de CPU et de RAM. Les lignes
cloud n'ont pas d'énergie : l'estimation par coefficient régional n'est pas branchée.

### Carbone

Deux méthodes sont calculées et affichées côte à côte par `/finops/carbon`.

| Méthode | Facteur | Source | Colonne |
|---|---|---|---|
| Émissions directes | Intensité moyenne du jour, en gCO2e/kWh | RTE éCO2mix, champ `taux_co2`, jeu temps réel puis consolidé quand il existe | `x_co2e_kg`, `x_carbon_method = rte-direct` |
| Analyse de cycle de vie | 51,9 gCO2e/kWh, fixe | ADEME Base Empreinte, élément 43642, mix moyen 2024, France continentale | Calculée à la lecture |

```text
CO2e_direct (kg) = énergie_kWh × intensité_du_jour_g_par_kWh / 1000
CO2e_ACV (kg)    = énergie_kWh × 51,9 / 1000
```

L'intensité du jour est la moyenne des relevés quart-horaires. Les valeurs consolidées de
RTE remplacent les valeurs temps réel dès qu'elles sont publiées. Tant que la mesure de
puissance repose sur la moyenne des CSV et non sur un exporter continu, l'énergie est
« mesurée en moyenne », pas en continu.

## 7. Finalité, clôture et ajustements

Une ligne `x_is_final = false` est révisée à chaque passe. L'interface marque le chiffre
« provisoire ».

| Source | La ligne devient finale |
|---|---|
| OpenCost, IA, SaaS | Quand le jour a au moins 3 jours (`FINOPS_REPLAY_DAYS`), car l'ingestion ne le rejoue plus |
| Facture AWS et GCP | Quand le mois suivant celui qui suit la période a commencé. Les clouds ferment une facture dans les premiers jours du mois suivant, et des crédits ou ajustements peuvent encore arriver pendant ce mois |

La table `month_close` fige un mois : l'ingestion ne réécrit plus ses lignes et le partage de
facture ne le touche plus. La finance clôture un mois avec `POST /api/finops/close/{AAAA-MM}` :
le mois courant est refusé, et un mois qui a encore des lignes de facture non définitives aussi,
sauf `force=true`. Une révision qui arrive après la clôture devient une ligne d'ajustement
(`ChargeCategory = Adjustment`) sur le premier mois ouvert, égale à l'écart ; rejouer la même
révision n'ajoute rien. Une ligne entièrement nouvelle sur un mois clos est ignorée.
