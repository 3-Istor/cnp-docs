# Offre aux projets : un « Datadog » inclus

Ce qu'un projet obtient en déployant sur la CNP, sans compte à créer ni agent à installer. Trois
niveaux, selon ce que l'app fait d'elle-même.

![Catalogue d'observabilité](../01-architecture/diagrams/observability-catalog.light.png#only-light)
![Catalogue d'observabilité](../01-architecture/diagrams/observability-catalog.dark.png#only-dark)

## Niveau 1 : toute app, zéro code

Vient de la plateforme : la gateway Envoy, Kubernetes, Cilium, Argo CD, Gatus, OpenCost.

| Ce que l'app reçoit | D'où ça vient | Où on le voit |
| --- | --- | --- |
| Requêtes, erreurs, latence par route | Métriques Envoy | Overview de l'app, Grafana du projet |
| CPU, mémoire, redémarrages, OOM, saturation | kube-state-metrics, cAdvisor | Overview |
| Logs de tous les conteneurs, 30 jours, recherche plein texte | vlagent, VictoriaLogs | Onglet Logs, Grafana |
| **Erreurs les plus fréquentes**, regroupées par motif | LogsQL `collapse_nums` + `stats by (_msg)` | Overview |
| **SLO de disponibilité et de latence, budget d'erreur** | Règles générées par le CMP | Overview, Settings pour la cible |
| Alertes par défaut vers le Discord de l'app ou du projet | vmalert, Alertmanager | Discord, onglet Alertes |
| Marqueurs de déploiement sur tous les graphes | Argo CD notifications | Graphes, Delivery |
| **Comparaison canary contre stable** pendant un déploiement | Requêtes de la porte du canary | Delivery |
| Rollback automatique si le canary dégrade les erreurs | Argo Rollouts, analyse sur VictoriaMetrics | Delivery, Discord |
| Page de statut publique (si l'owner l'active) | Gatus du projet | `status-<projet>.3istor.com` |
| Carte des flux : qui parle à qui, ce qui est refusé | Hubble | Vue Topologie du projet |
| Coût par jour et **coût par 1 000 requêtes**, gaspillage | OpenCost, Envoy | Costs |
| Fréquence de déploiement, taux d'échec des changements | Argo CD, Rollouts, alertes | Delivery |
| Score de maturité observabilité | Le CMP vérifie ce qui précède | Overview |

## Niveau 2 : app créée depuis un template

Le template FastAPI + React + PostgreSQL arrive instrumenté.

| Ce que l'app reçoit en plus | Comment |
| --- | --- |
| Métriques applicatives et **métriques métier** (`/metrics`, compteur d'exemple documenté) | `prometheus-client`, scrape déclaré par le CMP |
| **Traces distribuées** : front → API → base | SDK OpenTelemetry, collecteur, VictoriaTraces |
| D'un log à sa trace et retour | `trace_id` dans les logs JSON |
| Panneau Postgres : connexions, requêtes lentes, réplication, taille | Métriques CNPG par cluster |
| Probes de santé réellement rendues | Chart générique corrigé |

## Niveau 3 : options à activer

| Option | Pour qui | Coût | État |
| --- | --- | --- | --- |
| Métriques RED sans code (eBPF) pour une app importée | Apps importées sans SDK (Quarkus d'epitweet) | RAM par nœud | Spike (D12) |
| Core Web Vitals (LCP, INP, CLS) par release | Fronts React | Faible | Plus tard |
| Profiling (flamegraphs) | Apps lentes à creuser | 0,5 à 1 Gio | Option, non planifiée |
| Traces des appels LLM d'une app (tokens, coût, latence) | Apps de projets qui appellent un modèle | Faible | Avec la landing zone IA |

## Qui voit quoi

Les droits suivent les rôles Keycloak et les capacités déjà définies dans le portail.

| Rôle | Métriques | Logs | Traces | Alertes | SLO, coût |
| --- | --- | --- | --- | --- | --- |
| Invité du projet | Overview en lecture | non | non | non | Statut |
| Membre | son projet | son projet | son projet (portail) | voir, mettre en sourdine | voir |
| Admin / owner | son projet | son projet | son projet | voir, mettre en sourdine, router | voir, régler la cible |
| infra | tout | tout | tout | tout | tout |
| security-auditor | tout en lecture | audit et plateforme, pas les logs applicatifs | journal des agents IA | lecture | lecture |
| finops | coûts et volumes | non | non | FinOps | tout en lecture |

Le security-auditor ne lit pas les logs applicatifs des projets, comme aujourd'hui dans
`/activity/logs` : ils peuvent contenir des données personnelles et ne servent pas à l'audit.
