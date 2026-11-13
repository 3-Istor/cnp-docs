# État de l'art (novembre 2026) et besoins des métiers

Synthèse de trois veilles menées le 13/11/2026 : écosystème VictoriaMetrics, observabilité d'une
plateforme interne face à Datadog, traçabilité de l'IA. Les sources sont citées en bas de page.
Les chiffres de RAM des éditeurs sont indicatifs : seule la mesure sur notre cluster compte.

## Ce que veulent les gens, d'après les études

| Étude | Ce qu'elle dit |
| --- | --- |
| Grafana Labs, 4e enquête observabilité (mars 2026, 1 363 répondants) | Premiers freins : complexité 38 %, bruit 34 %, coût 31 %. La fatigue d'alerte est le premier obstacle à une réponse rapide aux incidents (30 %). 77 % jugent l'open source ou les standards ouverts importants. 76 % investissent dans OpenTelemetry. |
| Grafana / ETR, « State of Observability 2026 » (octobre 2026) | 85 % des décideurs acceptent une IA qui enquête en lecture seule, 50 % seulement une IA qui déploie. |
| Enquête Middleware relayée par la CNCF (février 2026) | Le premier problème de mise en place est la configuration des dashboards et des alertes (54 %). Seuls 7,4 % ont une expérience unifiée. |
| DORA 2025 | Cinq métriques : délai de livraison, fréquence de déploiement, temps de rétablissement, taux d'échec des changements, taux de reprise. |

**Ce qu'on en retient pour la CNP** : le gain n'est pas d'ajouter des outils, c'est de **tout
préconfigurer** (dashboards, SLO, alertes) et de **réduire le bruit**. Une plateforme interne est
exactement l'endroit où on peut le faire pour tout le monde d'un coup.

## Besoins par rôle

Les rôles sont ceux de Keycloak, sans nouvelle persona.

| Rôle | Ce qu'il demande | Ce que la CNP doit lui donner |
| --- | --- | --- |
| Membre de projet (dev) | Mon déploiement est-il sain ? Pourquoi c'est lent ou en erreur ? | Signaux dorés par app, marqueurs de déploiement, logs et traces d'une requête, verdict du canary, erreurs les plus fréquentes |
| Owner / admin de projet | SLO, budget d'erreur, qui est alerté, coût, page de statut | Vue SLO, routage Discord par app, statut public, coût par 1 000 requêtes |
| Invité | Est-ce que ça marche ? | Page de statut, vue Overview en lecture |
| infra (admin plateforme) | Santé de la plateforme, voisins bruyants, bruit des alertes | Dashboards plateforme, consommation par projet, qualité des alertes, santé de la stack elle-même |
| security-auditor | Qui a fait quoi, refus réseau, traces des agents IA | Audit k8s, Vault, Keycloak, Hubble, journal des agents, archive Object Lock, lecture seule transverse |
| finops | Coût par projet et par requête, gaspillage | OpenCost, coût par 1 000 requêtes, demandes vs consommation réelle |
| Métier (démo, direction) | La plateforme est-elle fiable et s'améliore-t-elle ? | DORA par projet, atteinte des SLO, score de maturité observabilité |

## Datadog, fonction par fonction

Personne ne branchera Datadog en plus de la CNP. La question est donc : qu'est-ce qu'un client
Datadog utilise vraiment, et qu'est-ce qu'on peut offrir à chaque projet sans qu'il code ?

| Fonction Datadog | Équivalent retenu sur la CNP | Coût | Verdict |
| --- | --- | --- | --- |
| Infrastructure monitoring | vmagent, kube-state-metrics, node-exporter, dashboards | faible | **Socle** |
| Log management | VictoriaLogs, explorateur dans le portail et Grafana | faible | **Socle** |
| APM et traces | SDK OpenTelemetry dans les templates, collecteur, VictoriaTraces | moyen | **Oui**, par app |
| Service map | Hubble (déjà là avec Cilium), arêtes observées dans la vue Topologie | faible | **Oui** |
| SLO | Règles générées par le CMP avec la bibliothèque Go de Sloth | faible | **Oui, cœur de l'offre** |
| Monitors et alertes | vmalert, Alertmanager, Discord par app | faible | **Oui** |
| Synthetics et statut | Gatus par projet (existe), probes générées depuis le catalogue | faible | **Oui** |
| Database monitoring | Métriques CNPG par cluster, panneau Postgres dans l'app | faible | **Oui** |
| Deployment tracking, DORA | Événements Argo CD et Rollouts, calcul dans le CMP | faible | **Oui** |
| Cloud cost | OpenCost, coût par requête, gaspillage | faible | **Oui** |
| LLM observability | Conventions OpenTelemetry GenAI, voir [IA et conformité](07-ia-et-conformite.md) | moyen | **Oui** |
| Bits AI (enquête IA) | Outils MCP en lecture du portail, avec les droits de l'utilisateur | faible | **Oui**, lecture seule |
| RUM | Web Vitals via une route même-origine | moyen | Plus tard |
| Profiler continu | Pyroscope | moyen | Option |
| Watchdog (anomalies) | vmanomaly est Enterprise. Règles z-score simples | faible | Règles simples seulement |
| Session replay | OpenReplay, rrweb | élevé | **Non** (RGPD, RAM) |
| Incident management, on-call | Grafana OnCall OSS archivé, Opsgenie fermé en avril 2027. Discord + ntfy | nul | Minimal |
| Error tracking | Sentry auto-hébergé 8 à 16 Gio | élevé | **Non** |

## L'écosystème VictoriaMetrics en novembre 2026

| Brique | État | Ce que ça change pour nous |
| --- | --- | --- |
| VictoriaMetrics v1.156, opérateur v0.75, chart `victoria-metrics-k8s-stack` 0.98 | Mûr | VMSingle recommandé sous un million de points par seconde : on en fait 10 000. Le mode cluster ne nous apporte rien. |
| vmauth | Mûr, mais trois contournements d'authentification corrigés en v1.151, v1.152 et v1.154 | Version minimale v1.156. Suivre à la lettre la section de la doc sur les arguments de requête imposés. |
| VictoriaLogs v1.54 | Prêt pour la production depuis la v1.0, cadence rapide avec ruptures mineures | **Pas de stockage S3** (PR ouverte, bugs signalés en novembre). **Une seule rétention globale**, pas par flux. Pas d'API Loki en lecture : LogQL à réécrire en LogsQL. |
| vlagent | Collecteur de logs Kubernetes léger (28 Mio dans le benchmark éditeur, Alloy 66 Mio) | Pas de routage de tenant par namespace documenté : un seul tenant, isolation dans vmauth. |
| VictoriaTraces v0.13 | **Pré-1.0**, format de données pas figé | OTLP en entrée, API Jaeger en lecture. Pas de filtre de requête documenté pour isoler un projet. |
| OBI (ex-Beyla, eBPF) | v1.0 le 30/10/2026 | Métriques RED sans code. Noyau 6.8 de nos nœuds absent de la matrice testée. |
| Enterprise seulement | vmanomaly, rétentions multiples, downsampling, vmalert multi-tenant | On s'en passe. |
| Serveurs MCP | `mcp-victoriametrics`, `mcp-victorialogs` | Ils lisent sans filtre de projet : on ne les expose pas tels quels. |

## Ce qui impressionne vraiment en démo, et ce qui est gadget

Fort impact pour un coût faible ou moyen, dans l'ordre :

1. Un déploiement canary qui dérape, le marqueur sur le graphe, l'AnalysisRun qui échoue,
   le rollback automatique, le message Discord avec le lien runbook.
2. Une app créée depuis le catalogue arrive avec son dashboard, son SLO et ses alertes, sans config.
3. D'une ligne de log à la trace de la requête, et retour, en un clic.
4. La carte des services vivante, avec les flux refusés.
5. Le coût par 1 000 requêtes et le gaspillage par app.
6. La trace complète d'un agent IA : demande, appels modèle, appels d'outils, approbation humaine,
   déploiement résultant, coût. Avec l'export de l'enregistrement pour l'audit.
7. Le score de maturité observabilité de chaque app dans le portail.

Gadgets pour notre contexte : énergie et carbone par pod sur des VM (Kepler n'estime pas bien les
VM), session replay, détection d'anomalies par IA, stacks lourdes (Sentry, Langfuse, ClickHouse,
DevLake, Backstage) montées seulement pour la démo.

## Sources principales

- Grafana Labs Observability Survey 2026 : <https://grafana.com/observability-survey/2026/>
- VictoriaMetrics, single vs cluster : <https://docs.victoriametrics.com/victoriametrics/faq/>
- vmauth, fusion des arguments de requête : <https://docs.victoriametrics.com/victoriametrics/vmauth/>
- Fonctions Enterprise : <https://docs.victoriametrics.com/victoriametrics/enterprise/>
- VictoriaLogs, stockage objet : <https://github.com/VictoriaMetrics/VictoriaLogs/issues/48>
- VictoriaLogs, rétention par flux refusée : <https://github.com/VictoriaMetrics/VictoriaLogs/issues/15>
- Benchmark des collecteurs 2026 : <https://victoriametrics.com/blog/log-collectors-benchmark-2026/>
- OBI v1.0 : <https://github.com/open-telemetry/opentelemetry-ebpf-instrumentation/releases/tag/v1.0.0>
- Sloth : <https://github.com/slok/sloth>
- Google SRE Workbook, alerting on SLOs : <https://sre.google/workbook/alerting-on-slos/>

Les rapports complets avec toutes les URL sont dans la PR (annexes de veille).
