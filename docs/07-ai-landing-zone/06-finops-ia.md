# FinOps de l'IA

> **Statut : brouillon v0.1 du 2026-10-25, à valider.** Cette page dit comment la
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

## 3. Lignes FOCUS

L'ingesteur du plan FinOps reçoit une source de plus, `ai_gateway`, qui écrit chaque jour une
ligne par `azp`, modèle et backend :

| Colonne FOCUS | Valeur |
|---|---|
| `ServiceCategory` | `AI and Machine Learning` |
| `ProviderName`, `ServiceName` | `Google`, `Gemini API` ; ou `CNP`, `Serveur GPU` |
| `ConsumedQuantity`, `ConsumedUnit` | nombre de tokens, `Tokens` |
| `ListCost` | tokens × prix public du jour (`ai-prices.yaml`) : ce que l'usage coûterait au niveau payant |
| `BilledCost`, `EffectiveCost` | 0 pour le niveau gratuit de Gemini ; part du TCO pour le serveur GPU, au prorata des tokens, quand le TCO existe |
| `Tags` | `cnp_project`, `cnp_app` ou `cnp_agent` (nouvelle dimension), `cnp_env` |
| `x_allocation_method` | `direct` |
| `x_source`, `x_is_final` | `ai_gateway` ; définitif à J+1 |

Afficher `ListCost` à côté d'un `BilledCost` nul répond à la question « combien coûterait
cette démo si on passait au payant ». Le dashboard le présente comme une **valeur au prix
public**, jamais comme une dépense.

## 4. Écrans

| Où | Quoi | Phase |
|---|---|---|
| Onglet « IA » du projet | Jauge des tokens du jour par agent ou app, face à son budget ; valeur au prix public du mois | 1 (lu dans VictoriaMetrics) |
| FinOps, vue d'ensemble | Catégorie « IA » dans le filtre et la ventilation par catégorie de service ; KPI « tokens du mois » et « valeur au prix public » | quand `cost_line_items` existe (I2 du plan FinOps) |
| FinOps, explorateur | Regroupement par agent (`cnp_agent`) et par modèle | idem |
| FinOps, budgets | Le budget mensuel du projet inclut la part d'IA facturée (aujourd'hui 0 pour Gemini gratuit) | idem |
| Page d'un agent (CMP) | Tokens du jour, budget, coût et valeur au prix public, tendance sur 30 jours | 2 |
| Grafana | Tableau « IA » : tokens par projet, agent, modèle ; 429 ; refus ; latence | 1 |

## 5. Avant que le plan FinOps soit implémenté

En phase 1, la CMP lit directement VictoriaMetrics pour l'onglet « IA » du projet
(`GET /api/projects/{projet}/ai/usage`). Les lignes FOCUS n'arrivent qu'avec l'itération I2
du plan FinOps. Le format ci-dessus est à ajouter à ce plan comme source `ai_gateway`, et la
ligne « fournisseurs LLM » de ses sources SaaS devient inutile pour les appels qui passent par
la gateway.

## 6. Vérifications

| ID | Phase | Vérification | Attendu |
|---|---|---|---|
| VF1.1 | 1 | 100 appels d'un agent de test, puis l'onglet « IA » du projet | Les tokens affichés égalent ceux des métriques de la gateway, à 1 % près |
| VF1.2 | 1 | Valeur au prix public | Tokens × prix de `ai-prices.yaml`, recalculable à la main |
| VF2.1 | I2 du plan FinOps | Lignes `ai_gateway` dans `cost_line_items` | Une ligne par `azp`, modèle, backend et jour ; ingestion idempotente (rejouer J-1 ne duplique rien) |
