# Plan de migration et recette

Six vagues. Chacune a un retour arrière et une porte de sortie mesurable. Ordre choisi pour livrer
tôt l'histoire de démo (SLO, canary, Discord) et ne démarrer les nouveautés coûteuses qu'avec les
gains de RAM mesurés.

| Vague | Contenu | Dépôts | Retour arrière | Porte de sortie |
| --- | --- | --- | --- | --- |
| **V0 · Sécurité et hygiène** | C1 durcissement vmauth + test d'isolation ; C2 CRDs d'observabilité réservées à la plateforme ; C3 NetworkPolicy Alertmanager ; C4 limites de requête et vmstorage à 1,5 Gio ; C5 un seul scrape Envoy ; C6 pods Kyverno et sidecar Gatus ; C7 probes du chart | CMP, cnp-project-base, K3s, infra-templates | Revert des PRs | Le test d'isolation renvoie 0 série ; plus de pods en erreur ; séries Envoy divisées par 2 |
| **V1 · Métriques** | Opérateur verrouillé (D1), VMSingle avec anciens noms de Service (D2), vmagent (D4), régime (D3), auto-surveillance, NetworkPolicies, Alloy ne fait plus que les logs | K3s | **Double écriture** : vmagent écrit dans les deux stacks pendant 14 jours ; les consommateurs passent un par un ; retour = sélecteur des Services | < 150 k séries ; tous les dashboards et règles identiques côte à côte ; Alloy < 300 Mio |
| **V2 · Logs** | VictoriaLogs + vlagent (D5, D6), schéma de champs commun, règles `vlogs` qui remplacent les compteurs Alloy, port des 7 règles de sécurité, export d'audit LogsQL, CMP derrière `LOGS_BACKEND=loki\|vl` | K3s, CMP, app-templates | **Double envoi** 14 jours ; retour = variable du CMP | Test d'usurpation (un pod qui écrit `namespace: victim` dans son JSON n'apparaît pas chez la victime) ; 48 h d'export d'audit identiques ligne à ligne ; Loki et Alloy supprimés |
| **V3 · Alerting et SLO** | Chaîne unique (D15), routage par app, format Discord, dead man's switch, SLO générés (D14), sourdine depuis le portail, qualité des alertes | K3s, CMP | Les règles Grafana restent jusqu'à ce que leur équivalent ait sonné | Démo : canary dégradé → rollback → message Discord avec runbook ; un SLO par app de prod |
| **V4 · Traces et templates** | Collecteur OpenTelemetry + VictoriaTraces (D10), template FastAPI instrumenté (D11), CNPG par app, Hubble (D13), spike OBI (D12) | K3s, app-templates, infra-templates | Désactivation par app ; Tempo en repli | Une requête suivie du front à la base ; RAM totale sous le budget |
| **V5 · Produit** | Overview, Delivery, Traces, Costs dans le portail (D16), dashboards par rôle (D17), DORA (D18), coût par requête (D19), score de maturité, outils MCP en lecture (D21) | CMP, K3s | Fonctions derrière des feature flags | Parcours de démo complet, revue a11y |
| **V6 · IA et clouds** | Observabilité et enregistrement d'audit des agents (D20), quand la landing zone arrive ; vmagent et vlagent sur EKS et GKE **quand Brian rallume les clouds** | K3s, Cloud, CMP | Par composant | Une tâche d'agent tracée et exportée ; les trois clusters dans Grafana avec le filtre `cluster` |

## Ce qui doit fonctionner à chaque instant

Ces consommateurs lisent la stack actuelle. Chaque vague garde leur chemin intact ou le bascule
explicitement :

- **Argo Rollouts** (`cnp-error-rate`) : anciens noms de Service, `envoy_cluster_endpoint_rq_*` conservé.
- **OpenCost** et **FinOps** du CMP (vmselect, `extra_filters[]` sur `cluster`).
- **Datasources des orgs projet** créées par Terraform (`…-vmauth:8427`) : rien à rejouer.
- **Relais WireGuard** des clouds (:18080, :18082, :18083) : cibles mises à jour en V1 et V2,
  même si les clouds sont éteints.
- **CMP** : `activity_sources.py`, `activity_collector.py`, `app_overview.py`, `alerts_catalog.py`,
  `alerts_grafana.py`, MCP `get_app_logs`.
- **Archive d'audit** horaire et alerte `AuditArchiveFailed`.

## Recette de bout en bout

| # | Vérification | Attendu |
| --- | --- | --- |
| R1 | Un membre du projet A tente de lire ou d'écrire des métriques, logs et traces du projet B par tous les chemins connus | Rien |
| R2 | Un job planifié rejoue R1 chaque nuit avec un compte jetable | Alerte si une donnée revient |
| R3 | Créer une app depuis le catalogue | Dashboard, SLO, alertes et page de statut présents en moins de 5 minutes |
| R4 | Déployer une version qui renvoie 10 % de 500 | Canary avorté, rollback, message Discord avec runbook et lien |
| R5 | Cliquer sur une ligne de log en erreur | La trace de la requête s'ouvre |
| R6 | Arrêter vmalert | Le dead man's switch prévient Discord en moins de 15 minutes |
| R7 | Mettre une alerte en sourdine depuis le portail | Silence limité au projet, visible dans l'onglet Alertes |
| R8 | Mesurer la RAM de `observability` | Inférieure à aujourd'hui (3,3 Gio) avec traces et Hubble compris |
| R9 | Une tâche d'agent IA | Trace complète, coût, approbation, enregistrement signé vérifiable |

## Qui fait quoi

CNP-185 est l'epic de Hugo, décidé avec Brian le 31/10. Ce dossier en élargit le périmètre. Proposition
de découpage à valider avec lui : Hugo garde V2 (logs) et V6 (clouds, qu'il connaît par CNP-110),
Brian prend V0, V1 et V3. V4 et V5 se répartissent après V3. CNP-191 (GitOps base et overlays)
touche les mêmes fichiers : V1 se fait dans la nouvelle structure si CNP-191 est prêt, sinon
CNP-191 reprend les fichiers après V1.
