# FinOps de l'IA

> **Statut : v1.0 du 2026-10-25, validée par Brian.** Cette page dit comment la
> consommation de tokens de la landing zone IA entre dans le module FinOps de la CMP. Elle
> s'appuie sur le « Plan FinOps CNP » du 16 octobre 2026 (proposition, pas encore validée ni
> implémentée) et en reprend les choix : format pivot FOCUS, table `cost_line_items` dans le
> PostgreSQL de la CMP, ingesteur `finops-ingester`, onglets du dashboard.

## 1. Deux budgets, deux rôles

| Budget | Unité | Où il s'applique | Effet au dépassement | Phase |
|---|---|---|---|---|
| Budget de tokens | tokens par jour, par agent ou app (`azp`) | agentgateway (D11, D21) | Refus 429 : plafond dur | 1 |
| Budget FinOps | euros par mois, par projet | CMP, module FinOps | Alerte à 70 % et 90 %, comme aujourd'hui | selon le plan FinOps |

Le budget de tokens protège la plateforme et le quota gratuit. Le budget FinOps rend le coût
visible au projet, avec le reste de ses dépenses.

## 2. D'où viennent les chiffres

La gateway compte les tokens de chaque appel et les expose en métriques OpenTelemetry GenAI,
qui arrivent dans VictoriaMetrics par Alloy (D7). C'est la source unique de la consommation
d'IA : elle est plus précise qu'une facture de fournisseur, qui ne connaît ni l'agent ni le
projet.

| Donnée | Source | Dimensions |
|---|---|---|
| Tokens en entrée et en sortie | Métriques GenAI de la gateway | `azp` (agent ou app), `project`, modèle, backend |
| Requêtes, refus, 429 | Métriques de la gateway | idem, plus le motif de refus |
| Prix public par million de tokens | Fichier versionné `finops/ai-prices.yaml` (dans le dépôt du TCO du plan FinOps) | modèle, date d'effet |
| Coût du serveur GPU | TCO on-prem du plan FinOps (énergie et amortissement), quand il existe | — |

## 3. Le coût de l'IA est compté au prix public

**Décision (Brian, 2026-10-25, étendue à Claude le 2026-10-26) : la consommation d'IA est
valorisée au tarif publié de chaque fournisseur, même quand elle est gratuite ou payée par des
crédits.** C'est ce chiffre que montrent les budgets, les dashboards et les exports : il dit ce
que coûte vraiment l'usage. Ce qui est réellement payé (0 € au niveau gratuit de Gemini ;
crédits inclus du Max pour Claude) ne sert qu'à la réconciliation. Le plan FinOps porte
la même règle (D-F8).

| Backend | Coût affiché (`x_cost_eur`) | Base (`x_pricing_basis`) |
|---|---|---|
| Claude | tokens en entrée × prix d'entrée + tokens en sortie × prix de sortie, au tarif publié d'Anthropic pour le modèle (par exemple `claude-haiku-5-5` : 0,10 $ et 0,50 $ par million de tokens pour un prompt de 100 000 tokens au plus), converti en euros au taux BCE du jour | `public_price` |
| Gemini | tokens en entrée × prix d'entrée + tokens en sortie × prix de sortie, au tarif payant du modèle, converti en euros au taux BCE du jour | `public_price` |
| Serveur GPU | part du TCO on-prem du plan FinOps, au prorata des tokens ; 0 tant que le TCO n'existe pas | `tco` |

Les prix vivent dans `finops/ai-prices.yaml`, à côté du TCO : par modèle, prix par million de
tokens en entrée et en sortie, devise, date d'effet et lien vers la page de tarifs du
fournisseur ([Claude](https://platform.claude.com/docs/en/about-claude/pricing),
[Gemini](https://ai.google.dev/gemini-api/docs/pricing)). Un
changement de prix est une PR, avec sa date d'effet.

## 4. Lignes FOCUS

L'ingesteur du plan FinOps reçoit une source de plus, `ai_gateway`, qui écrit chaque jour une
ligne par `azp`, modèle et backend :

| Colonne FOCUS | Valeur |
|---|---|
| `ServiceCategory` | `AI and Machine Learning` |
| `ProviderName`, `ServiceName` | `Anthropic`, `Claude API` ; `Google`, `Gemini API` ; ou `CNP`, `Serveur GPU` |
| `ConsumedQuantity`, `ConsumedUnit` | nombre de tokens, `Tokens` |
| `ListCost` | tokens × tarif payant du jour (`ai-prices.yaml`) |
| `BilledCost` | ce que le fournisseur facture réellement : le prix des tokens pour Claude, 0 au niveau gratuit de Gemini |
| Crédits inclus du Max | une ligne mensuelle `ChargeCategory` = `Credit`, du montant des crédits consommés, pour que la réconciliation retombe sur ce qui a été payé |
| `x_cost_eur`, `x_pricing_basis` | le coût affiché par CNP (§3) et sa base |
| `Tags` | `cnp_project`, `cnp_app` ou `cnp_agent` (nouvelle dimension), `cnp_env` |
| `x_allocation_method` | `direct` |
| `x_source`, `x_is_final` | `ai_gateway` ; définitif à J+1 |

## 5. Écrans

| Où | Quoi | Phase |
|---|---|---|
| Onglet « IA » du projet | Jauge des tokens du jour par agent ou app, face à son budget ; coût IA du mois au prix public | 1 (lu dans VictoriaMetrics) |
| FinOps, vue d'ensemble | Catégorie « IA » dans le filtre et la ventilation par catégorie de service ; KPI « tokens du mois » et « coût IA (prix public) », avec « facturé : 0 € » en petit tant que le niveau gratuit est utilisé | quand `cost_line_items` existe (I2 du plan FinOps) |
| FinOps, explorateur | Regroupement par agent (`cnp_agent`) et par modèle | idem |
| FinOps, budgets | Le budget mensuel du projet compte le coût IA au prix public | idem |
| FinOps, réconciliation | Coût au prix public face à la facture réelle de Google, écart visible | idem |
| Page d'un agent (CMP) | Tokens du jour, budget, coût au prix public, tendance sur 30 jours | 2 |
| Grafana | Tableau « IA » : tokens par projet, agent, modèle ; 429 ; refus ; latence | 1 |

## 6. Avant que le plan FinOps soit implémenté

En phase 1, la CMP lit directement VictoriaMetrics pour l'onglet « IA » du projet
(`GET /api/projects/{projet}/ai/usage`). Les lignes FOCUS n'arrivent qu'avec l'itération I2
du plan FinOps. Le plan FinOps a reçu la source `ai_gateway` et la décision D-F8 le 2026-10-25 ; sa ligne
« fournisseurs LLM » ne couvre plus que les appels qui ne passent pas par la gateway.

## 7. Vérifications

| ID | Phase | Vérification | Attendu |
|---|---|---|---|
| VF1.1 | 1 | 100 appels d'un agent de test, puis l'onglet « IA » du projet | Les tokens affichés égalent ceux des métriques de la gateway, à 1 % près |
| VF1.2 | 1 | Coût au prix public | Tokens en entrée et en sortie × prix de `ai-prices.yaml`, recalculable à la main |
| VF2.1 | I2 du plan FinOps | Lignes `ai_gateway` dans `cost_line_items` | Une ligne par `azp`, modèle, backend et jour ; ingestion idempotente (rejouer J-1 ne duplique rien) |

## 8. Suivre la dépense réelle sur Claude

La dépense réelle se lit dans la Claude Console (usage de l'espace de travail `cnp-demo`, face
à son plafond de 20 $ par mois, D38). La CMP l'affiche sur l'écran Modèles et sur la carte des
backends de l'onglet IA, à partir des métriques de la gateway ; la Console fait foi en cas
d'écart. Les crédits du Max (100 $ par mois pour Max 5x) expirent à chaque cycle sans report :
la dépense du mois ne déborde jamais sur le mois suivant.
