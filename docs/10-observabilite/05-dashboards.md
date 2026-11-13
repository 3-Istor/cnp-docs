# Dashboards et vues, par rôle

Règle : **le portail montre ce dont on a besoin pour décider, Grafana sert à creuser.** Chaque vue
du portail a un lien « Ouvrir dans Grafana » avec les mêmes filtres et la même période.

## Dans le portail (CMP)

| Vue | Contenu | Rôles |
| --- | --- | --- |
| **App · Overview** | Requêtes, erreurs, p95 ; budget d'erreur et burn rate ; marqueurs de déploiement ; 5 erreurs les plus fréquentes ; dernières traces lentes ; alertes en cours avec bouton Sourdine ; score de maturité | membre et plus, invité en lecture partielle |
| **App · Delivery** | Canary contre stable en direct (erreurs, latence par version), verdict de l'analyse, fréquence de déploiement, taux d'échec | membre et plus |
| **App · Logs** | Explorateur LogsQL avec niveaux, recherche, contexte autour d'une ligne, lien vers la trace | membre et plus |
| **App · Traces** | Liste des traces (lentes, en erreur), cascade d'une trace, lien vers les logs | membre et plus |
| **App · Costs** | Coût par jour, coût par 1 000 requêtes, gaspillage requests contre usage | membre et plus, finops |
| **Projet · Topologie** | Apps, liens déclarés et observés, flux refusés | membre et plus |
| **Projet · SLO** | Tous les SLO du projet, budget restant, tendance | membre et plus |
| **Plateforme · Santé** | Santé de chaque brique, stack d'observabilité comprise, qualité des alertes | infra |
| **Activité** (existe) | Audit, Vault, Keycloak, refus réseau, actions des agents IA | infra, auditor, membres pour leur projet |

## Dans Grafana, org 1 (plateforme)

Un dossier par rôle, des dashboards générés et versionnés dans K3s.

| Dossier | Dashboards | Pour |
| --- | --- | --- |
| **Plateforme** | Vue d'ensemble du cluster (USE nœuds, pods, PVC), Ceph, Argo CD, cert-manager, CNPG, Envoy Gateway, Keycloak, Vault | infra |
| **Observabilité** | Santé VMSingle, VictoriaLogs, VictoriaTraces, vmagent, vlagent, vmalert, Alertmanager ; cardinalité ; volume de logs par namespace | infra |
| **Alertes** | Qualité des alertes : déclenchements par règle sur 7 jours, temps de résolution, alertes limitées, règles jamais déclenchées | infra |
| **Projets** | Voisins bruyants : CPU, mémoire, logs, séries par projet ; SLO de tous les projets | infra |
| **Sécurité** | Refus réseau Hubble, échecs de connexion Keycloak, refus Vault, exec dans les pods, élévations JIT, CVE Trivy, décisions des garde-fous IA | security-auditor, infra |
| **FinOps** | Coût et énergie (existe), coût par requête, gaspillage, tokens IA par projet | finops, infra |
| **IA** | Appels de modèles, tokens, coût, latence, refus, appels d'outils, approbations | infra, auditor |
| **Direction** | DORA par projet, atteinte des SLO, disponibilité de la plateforme, adoption | tout le monde en lecture |

## Dans Grafana, orgs projet

Créés par le CMP à la création du projet, avec les datasources du projet (vmauth) :

- **App · signaux dorés** : un dashboard par app, variables app et version.
- **Projet · vue d'ensemble** : toutes les apps, SLO, coût.
- **Logs** : explorateur VictoriaLogs du projet.
- **Postgres** : un panneau par cluster CNPG.

Les membres du projet sont Editor de leur org : ils peuvent ajouter leurs propres dashboards
métier, qui restent dans leur org.
