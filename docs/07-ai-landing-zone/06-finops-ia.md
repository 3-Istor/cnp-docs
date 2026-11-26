# FinOps de l'IA

> **Statut : v1.1 du 2026-11-22** (v1.0 validée par Brian le 2026-10-25 ; révisée pour l'API
> Anthropic, seul fournisseur). Cette page dit comment la
> consommation de tokens de la landing zone IA entre dans le module FinOps de la CMP. Elle
> s'appuie sur le « Plan FinOps CNP » du 16 octobre 2026 (validé le 2026-10-26 et implémenté
> côté CMP, voir [la section FinOps](../08-finops/00-index.md)) et en reprend les choix : format pivot FOCUS, table `cost_line_items` dans le
> PostgreSQL de la CMP, ingesteur `finops-ingester`, onglets du dashboard.

## 1. Deux budgets, deux rôles

| Budget | Unité | Où il s'applique | Effet au dépassement | Phase |
|---|---|---|---|---|
| Budget de tokens | tokens par jour, par agent ou app (`azp`) | agentgateway (D11, D21) | Refus 429 : plafond dur | 1 |
| Budget FinOps | euros par mois, par projet | CMP, module FinOps | Alerte à 70 % et 90 %, comme aujourd'hui | selon le plan FinOps |
| Plafond de dépenses | dollars, pour toute la clé | Console Anthropic | L'API refuse ; la gateway répond « service indisponible » (D13) : fusible dur | 1 |

**Budget total de la démo : 5 $ cumulés**, dont 1,50 $ pour la construction et les tests. Le plafond global est natif à la gateway ; `budget-guard` ajoute le coût des recherches web, le message de refus et l'alerte à 80 % et 100 % (spec D32). Le budget de tokens protège la plateforme et le crédit (environ 10 $) d'un agent qui boucle.
Le plafond de la console est le dernier fusible, au cas où les budgets de la gateway seraient
mal réglés. Le budget FinOps rend le coût visible au projet, avec le reste de ses dépenses.

**Ordre de grandeur.** Avec Haiku 5.5, le modèle par défaut, 10 $ couvrent environ 20 millions de tokens en
sortie ou 100 millions en entrée ; avec Sonnet 5.5, vingt fois moins. La recherche web se
paie en plus, 10 $ pour 1 000 recherches : c'est elle qui domine le coût d'une réponse de Grok.

**Coûts mesurés à l'essai du 2026-11-22** (Haiku 5.5, `effort: low`) : un fact-check avec une recherche web coûte 0,012 $ (19 081 tokens en entrée, 748 en sortie, dont 0,010 $ de recherche) ; un graphe en code mode coûte 0,0008 $. Avec 5 $, environ 400 fact-checks ou plus de 6 000 graphes. Les budgets quotidiens par `azp` se règlent pour que leur somme sur
la durée de la démo reste sous le plafond de la console.

## 2. D'où viennent les chiffres

La gateway compte les tokens de chaque appel et les expose en métriques OpenTelemetry GenAI,
qui arrivent dans VictoriaMetrics par Alloy (D7). C'est la source unique de la consommation
d'IA : elle est plus précise qu'une facture de fournisseur, qui ne connaît ni l'agent ni le
projet.

| Donnée | Source | Dimensions |
|---|---|---|
| Tokens en entrée et en sortie | Métriques GenAI de la gateway | `azp` (agent ou app), `project`, modèle, backend |
| Requêtes, refus, 429 | Métriques de la gateway | idem, plus le motif de refus |
| Prix public par million de tokens | Fichier versionné `finops/ai-prices.yaml` (dans le dépôt du TCO du plan FinOps) | modèle, type de token (entrée, sortie, lecture de cache), date d'effet |
| Dépense réelle | Console Anthropic (usage et coût de la clé) | modèle, jour ; ne connaît ni l'agent ni le projet |

## 3. Le coût de l'IA est compté au prix public

**Décision (Brian, 2026-10-25, révisée le 2026-11-22) : la consommation d'IA est valorisée
au tarif publié par Anthropic**, converti en euros au taux BCE du jour. Le calcul se fait dans
Grafana depuis la table de prix (tableau « IA ») et dans l'ingesteur FinOps. Il n'y a plus de
niveau gratuit : le prix public et la facture réelle doivent coïncider, et la réconciliation
avec la console Anthropic sert à le vérifier. Le plan FinOps porte la même règle (D-F8).

| Modèle | Entrée | Sortie | Lecture de cache | Base (`x_pricing_basis`) |
|---|---|---|---|---|
| `claude-sonnet-5-5` | 2 $ / MTok | 10 $ / MTok | 0,10 $ / MTok | `public_price` |
| `claude-haiku-5-5` | 0,10 $ / MTok | 0,50 $ / MTok | | `public_price` |

Coût affiché (`x_cost_eur`) = tokens en entrée × prix d'entrée + tokens en sortie × prix de
sortie + tokens lus en cache × prix de cache, en dollars, puis en euros. Les tokens de
l'adaptive thinking sont facturés comme des tokens de sortie. Le catalogue de prix de la gateway ne compte que les tokens : la recherche web (0,01 $ par recherche) est ajoutée par `budget-guard` (constaté à l'essai du 2026-11-22).

Les prix vivent dans `finops/ai-prices.yaml`, à côté du TCO : par modèle, prix par million de
tokens par type, devise (USD), date d'effet et lien vers la
[page des tarifs de l'API Anthropic](https://docs.anthropic.com/en/docs/about-claude/pricing).
Un changement de prix est une PR, avec sa date d'effet.

## 4. Lignes FOCUS

L'ingesteur du plan FinOps reçoit une source de plus, `ai_gateway`, qui écrit chaque jour une
ligne par `azp`, modèle et backend :

| Colonne FOCUS | Valeur |
|---|---|
| `ServiceCategory` | `AI and Machine Learning` |
| `ProviderName`, `ServiceName` | `Anthropic`, `Claude API` |
| `ConsumedQuantity`, `ConsumedUnit` | nombre de tokens, `Tokens` |
| `ListCost` | tokens × tarif publié du jour (`ai-prices.yaml`) |
| `BilledCost` | égal à `ListCost` : pas de niveau gratuit ; l'écart avec la console se voit à la réconciliation |
| `x_cost_eur`, `x_pricing_basis` | le coût affiché par CNP (§3) et sa base |
| `Tags` | `cnp_project`, `cnp_app` ou `cnp_agent` (nouvelle dimension), `cnp_env` |
| `x_allocation_method` | `direct` |
| `x_source`, `x_is_final` | `ai_gateway` ; définitif à J+1 |

## 5. Écrans

| Où | Quoi | Phase |
|---|---|---|
| Onglet « IA » du projet | Jauge des tokens du jour par agent ou app, face à son budget ; coût IA du mois en euros, au prix public | 1 (lu dans VictoriaMetrics) |
| FinOps, vue d'ensemble | Catégorie « IA » dans le filtre et la ventilation par catégorie de service ; KPI « tokens du mois » et « coût IA » en euros | quand `cost_line_items` existe (I2 du plan FinOps) |
| FinOps, explorateur | Regroupement par agent (`cnp_agent`) et par modèle | idem |
| FinOps, budgets | Le budget mensuel du projet compte le coût IA au prix public | idem |
| FinOps, réconciliation | Coût calculé face à la dépense de la console Anthropic, écart visible ; reste du crédit face au plafond | idem |
| Page d'un agent (CMP) | Tokens du jour, budget, coût au prix public, tendance sur 30 jours | 2 |
| Grafana | Tableau « IA » : tokens par projet, agent, modèle ; coût en euros depuis la table de prix ; 429 ; refus ; latence | 1 |

## 6. Avant que la gateway soit déployée

Le plan FinOps est validé depuis le 2026-10-26 et la source `ai_gateway` est implémentée
dans la CMP, testée sur fixtures. Elle n'écrit aucune ligne tant qu'agentgateway n'est pas
déployée, parce que les métriques de tokens n'existent pas encore. En phase 1, la CMP lit
directement VictoriaMetrics pour l'onglet « IA » du projet
(`GET /api/projects/{projet}/ai/usage`). La décision D-F8 est consignée dans
[D-14](../08-finops/01-decisions.md), et sa ligne « fournisseurs LLM » ne couvre que les
appels qui ne passent pas par la gateway.

## 7. Vérifications

| ID | Phase | Vérification | Attendu |
|---|---|---|---|
| VF1.1 | 1 | 100 appels d'un agent de test, puis l'onglet « IA » du projet | Les tokens affichés égalent ceux des métriques de la gateway, à 1 % près |
| VF1.2 | 1 | Coût au prix public | Tokens en entrée et en sortie × prix de `ai-prices.yaml`, recalculable à la main |
| VF1.3 | 1 | Plafond de la console atteint | La gateway répond « service indisponible » ; l'alerte part ; aucun agent ne contourne |
| VF2.1 | Gateway déployée, source `ai_gateway` active | Lignes `ai_gateway` dans `cost_line_items` | Une ligne par `azp`, modèle, backend et jour ; ingestion idempotente (rejouer J-1 ne duplique rien) |
