# Observabilité : full VictoriaMetrics et offre aux projets

**Statut : proposition à valider, rien n'est déployé.** Ticket [CNP-185](https://3istors.atlassian.net/browse/CNP-185)
(epic de Hugo), dont ce dossier élargit le périmètre.

## En bref

Aujourd'hui l'observabilité de la CNP est un mélange : Alloy fait tout et pèse 1,8 Gio, les logs
sont dans Loki mais leurs alertes dans VictoriaMetrics, quatre chemins mènent à Discord, la stack
ne se surveille pas, et les apps n'exposent ni métriques, ni traces, ni logs structurés.

La cible :

- **Trois bases VictoriaMetrics** (métriques, logs, traces), **deux collecteurs légers**, **une
  chaîne d'alerte** routée par app vers Discord, Grafana pour les dashboards. Moins de RAM
  qu'aujourd'hui, traces comprises.
- **Un « Datadog » inclus pour chaque projet**, sans code : signaux dorés, logs, erreurs
  fréquentes, SLO avec budget d'erreur, canary piloté par les métriques, marqueurs de déploiement,
  page de statut, carte des flux, coût par requête, DORA. Les apps créées depuis un template
  ajoutent traces, métriques métier et Postgres.
- **La traçabilité des agents IA** au niveau recommandé par l'ANSSI et la CNIL en 2026 : une trace
  par tâche, approbations comprises, et un enregistrement d'audit signé en archive Object Lock
  180 jours.
- **Trois durcissements de sécurité d'abord**, trouvés pendant l'inventaire, sur l'isolation entre
  projets (détail hors de ce dépôt public tant qu'ils ne sont pas corrigés).

## Les pages

| Page | Contenu |
| --- | --- |
| [Existant](01-existant.md) | Inventaire mesuré, le « mix », 8 constats à corriger tout de suite |
| [État de l'art](02-etat-de-l-art.md) | Études 2026, besoins par rôle, Datadog fonction par fonction, écosystème VictoriaMetrics |
| [Architecture cible](03-architecture.md) | Décisions D1 à D21, avant et après, budget RAM et CPU |
| [Offre aux projets](04-offre-projets.md) | Ce qu'une app obtient à chaque niveau, qui voit quoi |
| [Dashboards et vues](05-dashboards.md) | Portail, Grafana plateforme par rôle, orgs projet |
| [Alerting](06-alerting.md) | Chaîne unique, salons, qualité, dead man's switch |
| [IA et conformité](07-ia-et-conformite.md) | Droit en vigueur, ce qu'on trace par tâche d'agent, enregistrement signé |
| [Plan et recette](08-plan-et-recette.md) | 6 vagues avec retour arrière, recette, répartition |
| [Idées et revue](09-idees-et-revue.md) | Retenu, optionnel, écarté ; constats de la revue contradictoire |

## Comment ce dossier a été construit

1. Inventaire du code (`origin/main` de 8 dépôts) et mesures en lecture seule sur le cluster.
2. Trois veilles en parallèle : écosystème VictoriaMetrics, observabilité d'une plateforme et
   parité Datadog, traçabilité de l'IA (droit européen et français, conventions OpenTelemetry).
3. Brouillon en 20 décisions, puis **revue contradictoire** par un relecteur indépendant qui a
   confronté chaque décision au code et à la doc des outils : 33 constats, dont 3 bloquants,
   tous intégrés.
4. Schémas noodle de la cible, du flux d'alerte, de la trace d'un agent et du catalogue.

## Ce que Brian doit trancher

| # | Question | Proposition |
| --- | --- | --- |
| Q1 | Valider le périmètre élargi de CNP-185 et le découpage avec Hugo (V2 et V6 pour Hugo, V0, V1, V3 pour Brian) | Oui, à voir avec Hugo |
| Q2 | Constats C1 à C3 : les vérifier et ouvrir les correctifs tout de suite, avant le reste | Oui |
| Q3 | Rétention des logs | 30 jours pour tous, audit 180 jours en archive |
| Q4 | Traces : VictoriaTraces (pré-1.0, tout VictoriaMetrics) ou Tempo (stable) | VictoriaTraces, Tempo en repli |
| Q5 | Dead man's switch : healthchecks.io (gratuit, externe) ou Gatus sur pae-node-1 | healthchecks.io |
| Q6 | Contenu des prompts des agents : jamais, ou opt-in par projet à 7 jours | Opt-in |
| Q7 | Résumé d'alerte par IA : modèle local (RTX 5080), Gemini gratuit, ou rien | Modèle local, en dernier |
| Q8 | Options : profiling Pyroscope, Hubble UI, ntfy | Hubble UI oui, les deux autres plus tard |
| Q9 | Traces des projets visibles dans le portail seulement, pas dans leur Grafana, tant que VictoriaTraces n'a pas de filtre imposable | Oui |
