# Idées retenues, optionnelles, écartées, et revue contradictoire

## La matrice

| Idée | Valeur | Coût | Contre | Verdict |
| --- | --- | --- | --- | --- |
| Full VictoriaMetrics (métriques, logs) | Une famille d'outils, ~1,5 Gio rendus | 2 vagues | Pas de S3 ni de rétention par flux pour les logs | **Retenu** |
| VictoriaTraces | Même famille, léger | faible | Pré-1.0, format non figé | **Retenu**, Tempo en repli |
| SLO générés par app | Cœur de l'offre, répond au premier frein (config) | faible | Faux positifs à faible trafic | **Retenu**, avec seuil de trafic |
| Canary piloté par métriques avec comparaison visuelle | Meilleur moment de démo | faible | Existe déjà en partie | **Retenu** |
| Chaîne d'alerte unique, routage par app | Moins de bruit | moyen | Migration des règles Grafana | **Retenu** |
| Dead man's switch | Seule défense contre une stack muette | nul | Service externe | **Retenu**, fournisseur à trancher |
| Hubble comme carte des services | Déjà là, toutes les apps | faible | L7 seulement avec politiques L7 | **Retenu** |
| Erreurs fréquentes par motif | Très utile au quotidien | faible | | **Retenu** |
| Coût par requête, gaspillage | FinOps concret | faible | | **Retenu** |
| DORA | Lisible par la direction | faible | Peu de déploiements en démo | **Retenu**, 2 métriques d'abord |
| Score de maturité observabilité | Pousse les bonnes pratiques | faible | | **Retenu** |
| Traçabilité des agents IA, enregistrement signé | Recommandé par ANSSI et CNIL, très démontrable | moyen | Dépend de la landing zone IA | **Retenu** (V6) |
| Outils MCP en lecture dans le portail | IA utile sans budget | faible | | **Retenu** |
| OBI (eBPF) | RED sans code pour apps importées | RAM | Noyau 6.8 hors matrice, coexistence Cilium | **Spike** |
| Résumé d'alerte par petit modèle | Effet démo | tokens | Contexte non fiable (logs) | **Option**, en dernier |
| Core Web Vitals | Expérience utilisateur | moyen | Écriture anonyme depuis Internet à sécuriser | **Plus tard** |
| Profiling (Pyroscope) | Flamegraphs impressionnants | 0,5 à 1 Gio | Mode mono-binaire non recommandé en prod | **Option** |
| Hubble UI | Carte en direct | 100 Mio | Doublon partiel avec la vue Topologie | **Option** |
| ntfy (push mobile) | Astreinte légère | nul | Relais ntfy.sh pour iOS | **Option** |
| Contenu des prompts en opt-in | Débogage des agents | faible | RGPD | **Option par projet** |
| VictoriaMetrics en mode cluster | | | Inutile sous 1 M points/s | **Écarté** |
| OpenTelemetry Operator | Injection automatique | | Aucune app Java/Node/.NET, API v1alpha1 | **Écarté** |
| Plusieurs instances VictoriaLogs pour des rétentions différentes | | RAM | Personne n'utilise les classes | **Écarté** |
| Grafana Foundation SDK | | | Pré-1.0 | **Écarté** |
| vmanomaly, Grafana OnCall, Keep, PagerDuty, Opsgenie | | | Enterprise, archivé, lourd, payant, fermé | **Écarté** |
| Session replay, Sentry, Langfuse, Coroot, Pixie, DevLake, Kepler | | Lourd | RGPD, RAM, VM non supportées | **Écarté** |

## Revue contradictoire

Un relecteur indépendant, sans accès au raisonnement du brouillon, a confronté chaque décision au
code (`origin/main` de 7 dépôts), à la documentation des outils et au cluster en lecture seule.
33 constats : 3 bloquants, 18 majeurs, 12 mineurs. Tous les bloquants et majeurs sont traités dans
la version actuelle.

| Constat | Gravité | Traitement |
| --- | --- | --- |
| Isolation vmauth actuelle à durcir | Bloquant | C1, D6, test R1/R2 |
| VMSingle expose écriture et admin aux utilisateurs vmauth | Bloquant | D7 : liste blanche de chemins, clés admin |
| Objets d'observabilité à réserver à la plateforme | Bloquant | C2, D9, sélecteurs de l'opérateur (D1) |
| Pas de NetworkPolicy dans le brouillon | Majeur | D8 |
| Isolation des traces non vérifiée | Majeur | D10 : traces par le portail seulement |
| Usurpation de namespace dans les spans OTLP | Majeur | D10 : association par connexion |
| Beacon Web Vitals = écriture anonyme | Majeur | Reporté, route même-origine |
| Alertes projet en VMRule élargissent les droits | Majeur | `params.extra_filters`, routage par app conservé, règles Grafana perso gardées |
| vmauth point de défaillance unique | Majeur | D7 : plateforme en direct, 2 réplicas |
| La porte du canary a besoin des séries par endpoint | Majeur | D3, D14 |
| Constat Envoy faux dans le brouillon | Majeur | C5 corrigé |
| Volume de logs sous-estimé d'un facteur 50 | Majeur | D5 : dimensionné sur 1 Go/jour |
| Champs de vlagent différents du modèle | Majeur | D4 : schéma commun, test d'usurpation |
| Compteurs de logs perdus avec Alloy | Majeur | Règles `vlogs` en V2 |
| Règles par défaut bruyantes sur k3s | Majeur | D1 : liste blanche |
| Clouds cassés après V1 | Majeur | Anciens noms de Service, relais mis à jour |
| Pas de retour arrière | Majeur | Double écriture, double envoi, bascule dans le CMP |
| Passage de l'archive d'audit sous-spécifié | Majeur | 48 h de double export comparé |
| Réécriture CMP sous-estimée, deux propriétaires des dashboards | Majeur | D17 : le CMP seul propriétaire |
| Budget RAM optimiste, CPU oublié, pic de double run | Majeur | Budget refait, porte RAM avant V4 |
| Quatre chemins de traces, trois cartes | Majeur | Un chemin, une carte |
