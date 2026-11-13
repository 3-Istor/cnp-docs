# Registre des traitements (RGPD art. 30), brouillon

!!! warning "Brouillon tiré du code, pas un registre validé"
    Cette page décrit ce que le code de la CMP fait des données personnelles au
    13/11/2026 : CMP `main` + PR `test/mcp-owasp-controls`, K3s
    [`205f439`](https://github.com/3-Istor/K3s/tree/205f439570e81dfa14d6b1e73dc113553854f2ab).
    Elle ne décide rien. Les finalités, les bases légales, les durées cibles et le
    statut des destinataires sont des **décisions humaines** à prendre par le
    responsable de traitement, avec le DPO s'il y en a un.

Chaque information porte une étiquette :

| Étiquette | Signification |
|---|---|
| **[Bearer]** | Détecté par Bearer CLI 2.1.3 (`--report privacy` / `--report dataflow`) |
| **[Code]** | Lu dans le code ou la configuration, avec le lien vers la ligne |
| **[Décision]** | À décider et à remplir par une personne. Le texte proposé n'est qu'une proposition |

## Méthode

1. Bearer CLI v2.1.3 (Elastic License 2.0, usage interne), binaire officiel
   `bearer_2.1.3_linux_amd64.tar.gz`, SHA-256
   `974d4ad7bb71bc578675684d444e345e8fbad91c85557aac035406ef5cc04c68` vérifié contre le
   `checksums.txt` de la release. Lancé sur `backend/` et `frontend/` avec
   `--skip-path node_modules,tests`.
2. Lecture des modèles SQLAlchemy (`backend/app/models`), des services qui écrivent
   ailleurs que dans Postgres, et de la configuration Loki et de l'archive d'audit dans
   le dépôt K3s.
3. La CI de la CMP relance les deux rapports sur chaque PR et publie le JSON en
   artefact `bearer-privacy` (job `bearer-privacy` de `.github/workflows/security.yml`).
   Le job ne bloque jamais : une nouvelle détection se classe ici, à la main.

### Ce que Bearer a trouvé

| Catégorie Bearer | Backend | Frontend | Lecture humaine |
|---|---|---|---|
| Username | 13 | 3 | Réel : `preferred_username` Keycloak partout |
| Unique Identifier | 9 | 1 | Réel : `sub` Keycloak, identifiants de revue d'accès |
| Fullname / Firstname / Lastname | 4 / 1 / 0 | 1 / 3 / 2 | Réel : claims `name`, `given_name`, `family_name` |
| Email Address | 2 | 3 | Réel : profil, grants d'assistants, synchro Grafana. 1 alerte moyenne côté backend |
| Image | 1 | 2 | Réel : photo de profil |
| Interactions | 25 | 0 | Appels sortants vers des tiers, voir [Destinataires](#destinataires-et-sous-traitants) |
| Passwords | 4 | 0 | Faux positif : noms de réglages (`LOKI_PASSWORD`...), pas des mots de passe d'utilisateurs |
| Geographic | 1 | 0 | Faux positif : régions cloud (`AWS_DEFAULT_REGION`...) |
| Age Range | 1 | 0 | Faux positif : `STEP_UP_MAX_AGE` |
| Browser Fingerprint | 1 | 0 | Faux positif : empreinte d'une vulnérabilité (`security/store.py`) |

Tiers détectés par Bearer : Cloudflare, Discord, GitHub, Google Cloud APIs, et AWS S3
comme magasin de données. Bearer **ne voit pas** les clients MCP (les assistants IA
appellent la CMP, la CMP ne les appelle pas) ni les services internes (Keycloak, Loki,
Grafana) : ils sont ajoutés à la main ci-dessous.

Le rapport dataflow signale aussi des journaux applicatifs qui contiennent des noms
d'utilisateur, par exemple `routers/projects.py:652` (« X a mis Y au rôle Z »). Ces
lignes partent dans Loki avec les logs du pod.

## Responsable et contacts

| Champ | Valeur |
|---|---|
| Responsable de traitement | **[Décision]** |
| Représentant / DPO | **[Décision]** |
| Équipe technique | Équipe plateforme CNP (3-Istor) |

## Personnes concernées

| Catégorie | Qui | Source **[Code]** |
|---|---|---|
| Utilisateurs de la plateforme | Toute personne qui se connecte à la CMP via Keycloak | `get_current_user`, `keycloak_service.py` |
| Membres de projet | Owner, admins, membres, invités d'un projet | `projects.owner_username`, groupes Keycloak `project-*` |
| Identités plateforme | Groupes `infra`, `security-auditor`, `finops` | `services/authz.py` |
| Personnes ayant connecté un assistant IA | Utilisateur qui a approuvé un client MCP | `agent_grants`, `agent_confirmations` |
| Auteurs d'événements d'audit | Toute personne ou compte qui modifie la plateforme, y compris via `kubectl`, Vault ou Keycloak | `audit_events`, `activity_events`, flux Loki d'audit |
| Utilisateurs finaux des apps hébergées | Personnes qui utilisent une app déployée sur CNP | Indirect : leurs données peuvent apparaître dans les logs d'app que la CMP relit (`get_app_logs`) |

## Traitements

Chaque traitement liste ce que le code stocke. Les colonnes **Finalité** et **Base
légale** sont des propositions à valider.

### T1. Comptes et préférences du portail

| | |
|---|---|
| Finalité **[Décision]** | Proposition : authentifier les utilisateurs et personnaliser le portail |
| Base légale **[Décision]** | Proposition : art. 6.1.b (contrat de travail / d'accès) ou 6.1.f (intérêt légitime) |
| Données **[Code]** | `sub`, `preferred_username`, `name`, `given_name`, `family_name`, `email`, `picture`, groupes, langue, favoris, lien vers l'installation GitHub App |
| Stockage **[Code]** | Keycloak (source) ; Postgres `user_preferences` ([`user_preference.py:14`](https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/app/models/user_preference.py#L14)), `user_github_installations` ([`user_github.py:19`](https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/app/models/user_github.py#L19)) ; photo dans le bucket S3 `user-avatars`, servie publiquement depuis `AVATARS_PUBLIC_URL_BASE` ([`config.py:68`](https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/app/core/config.py#L68)) ; cookie de session NextAuth côté navigateur (nom, email, image : `frontend/src/auth.ts`) |
| Durée configurée **[Code]** | Aucune purge dans le code : les lignes vivent tant que le compte existe. Les comptes inactifs 60 j sont signalés puis désactivés 14 j plus tard (`INACTIVE_AFTER_DAYS`, `INACTIVE_GRACE_DAYS`), sans suppression des données |
| Durée cible **[Décision]** | À fixer, avec la suppression au départ d'un collaborateur |

### T2. Projets, membres et revues d'accès

| | |
|---|---|
| Finalité **[Décision]** | Proposition : gérer les droits sur les projets et prouver leur revue périodique |
| Base légale **[Décision]** | Proposition : 6.1.f, ou 6.1.c si la revue d'accès répond à une obligation (NIS2, ISO 27001...) |
| Données **[Code]** | Owner du projet, demandes d'accès just-in-time (demandeur, motif, décideur), éléments de revue (`subject`, `username`, dernier login, relecteurs, décision), dernier login vu |
| Stockage **[Code]** | Postgres `projects` ([`project.py:54`](https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/app/models/project.py#L54)), `access_requests` ([`access.py:16`](https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/app/models/access.py#L16)), `access_review_campaigns`, `access_review_items`, `account_activity` ([`access_review.py:16`](https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/app/models/access_review.py#L16)) ; groupes Keycloak |
| Durée configurée **[Code]** | Aucune purge. Une demande d'accès expire après 120 min (`ACCESS_REQUEST_TTL_MINUTES`) mais la ligne reste |
| Durée cible **[Décision]** | À fixer : une revue d'accès sert de preuve, sa durée suit souvent celle de l'audit |

### T3. Journal d'audit et fil d'activité

| | |
|---|---|
| Finalité **[Décision]** | Proposition : sécurité, traçabilité des changements, enquête après incident |
| Base légale **[Décision]** | Proposition : 6.1.f (sécurité du SI), voire 6.1.c selon les obligations applicables |
| Données **[Code]** | Auteur (`preferred_username` ou `sub`), action, cible, IP source (`X-Forwarded-For` ou IP du pair), détails filtrés (seuls les champs de `_SAFE_FIELDS` sont copiés, les clés qui ressemblent à un secret sont masquées), assistant IA en cause (`via`). Fil d'activité : mêmes champs plus `user_agent`, groupes et URI de la requête pour les événements Kubernetes, IP pour Keycloak |
| Stockage **[Code]** | Postgres `audit_events` et `activity_events` ([`audit.py:16`](https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/app/models/audit.py#L16)) ; Loki : flux `{job="k8s-audit"}`, Vault, Keycloak (`org.keycloak.events`), Hubble, et les logs de tous les pods dont la CMP ; archive horaire des trois flux d'audit dans un bucket Ceph RGW |
| Durée configurée **[Code]** | Postgres : 30 j, `AUDIT_RETENTION_DAYS` ([`config.py:151`](https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/app/core/config.py#L151)), purge toutes les 6 h ([`audit.py:551`](https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/app/services/audit.py#L551)), aucune surcharge dans K3s. Loki : 14 j par défaut ([`values-loki.yaml:80`](https://github.com/3-Istor/K3s/blob/205f439570e81dfa14d6b1e73dc113553854f2ab/k8s/config/observability/values-loki.yaml#L80)), 7 j ou 30 j par label `retention` ; les flux d'audit sont à 30 j (voir [Journaux d'audit](../01-architecture/06-audit-logging.md)). Archive : Object Lock COMPLIANCE 30 j ([`audit-archive.yaml:48`](https://github.com/3-Istor/K3s/blob/205f439570e81dfa14d6b1e73dc113553854f2ab/k8s/config/observability/audit-archive.yaml#L48)) ; **aucune règle d'expiration** : passé 30 j, les objets restent jusqu'à suppression manuelle |
| Durée cible **[Décision]** | À fixer, en particulier pour l'archive : 30 j de verrou n'est pas une durée de conservation |

### T4. Assistants IA connectés par MCP

| | |
|---|---|
| Finalité **[Décision]** | Proposition : permettre à un utilisateur de faire agir un assistant IA sur ses projets, avec ses droits |
| Base légale **[Décision]** | Proposition : 6.1.b ou 6.1.f ; le choix de l'assistant revient à l'utilisateur, à cadrer par une politique d'usage |
| Données **[Code]** | Identité copiée du token Keycloak à l'approbation : `sub`, `preferred_username`, `name`, `given_name`, `family_name`, `email` ; les groupes et rôles ne sont **pas** copiés ([`store.py:45`](https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/app/services/agents/store.py#L45)). Nom du client déclaré par l'assistant, scopes, dernière utilisation, demandes d'approbation (utilisateur, outil, arguments, résumé, décideur). Tokens stockés en SHA-256 seulement |
| Stockage **[Code]** | Postgres `agent_clients`, `agent_auth_requests`, `agent_auth_codes`, `agent_grants`, `agent_access_tokens`, `agent_confirmations`, `project_agent_settings` ([`agent.py`](https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/app/models/agent.py)) ; GitHub : chaque commit poussé par un assistant porte « Requested by `<username>` via `<assistant>` » ([`mcp_server.py:795`](https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/app/mcp_server.py#L795)) |
| Durée configurée **[Code]** | Demande d'autorisation 10 min, code 5 min, token d'accès 15 min (purgés au passage suivant). Grant : 30 j glissants (`AGENT_REFRESH_TTL_DAYS`), révoqué si le compte Keycloak est désactivé. Les grants révoqués et les approbations restent en base sans purge |
| Durée cible **[Décision]** | À fixer pour les grants révoqués et `agent_confirmations` |

### T5. Notifications et alertes

| | |
|---|---|
| Finalité **[Décision]** | Proposition : prévenir les personnes concernées d'un événement qui demande une action |
| Base légale **[Décision]** | Proposition : 6.1.f |
| Données **[Code]** | Destinataire (`user_key`), paramètres du message (souvent un nom d'utilisateur), lien. Discord : demandeur et décideur d'un accès élevé ([`grants.py:396`](https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/app/services/access/grants.py#L396)), alertes d'app, de budget et de sécurité |
| Stockage **[Code]** | Postgres `notifications` ; serveur Discord via webhooks |
| Durée configurée **[Code]** | Postgres : lues 30 j, toutes 90 j (`notifications.purge_expired`). Discord : hors de notre contrôle |
| Durée cible **[Décision]** | Valider ; décider si Discord est acceptable pour des noms de personnes |

### T6. Traçabilité des releases, de la sécurité et du FinOps

| | |
|---|---|
| Finalité **[Décision]** | Proposition : savoir qui a promu, approuvé une exception ou modifié un budget |
| Base légale **[Décision]** | Proposition : 6.1.f |
| Données **[Code]** | Noms d'utilisateur dans `release_events.actors`, `promotions.actor`, `verification_runs.actor`, `promotion_policies.updated_by`, `security_exceptions.approved_by` / `revoked_by`, `alert_settings.updated_by`, `project_budgets.updated_by`, `month_close.closed_by`, `cost_centers.budget_owner` |
| Stockage **[Code]** | Postgres (`models/release.py`, `environment.py`, `security.py`, `alerting.py`, `finops.py`) |
| Durée configurée **[Code]** | Aucune purge |
| Durée cible **[Décision]** | À fixer |

## Destinataires et sous-traitants

Observés dans le code. Le statut juridique (sous-traitant, responsable conjoint,
destinataire) et l'existence d'un contrat ou de clauses types sont **[Décision]**.

| Destinataire | Données personnelles qui y partent **[Code]** | Où | Détecté par | Hors UE ? **[Décision]** |
|---|---|---|---|---|
| Assistants IA choisis par l'utilisateur (Anthropic Claude, OpenAI ChatGPT/Codex, Cursor...) via MCP | Tout ce que renvoient les outils : nom et nom d'utilisateur (`whoami`), projets, contenu des dépôts, **logs d'app pouvant contenir des données d'utilisateurs finaux** (`get_app_logs`) | `backend/app/mcp_server.py` | Lecture | Probablement (US). À cadrer : qui est responsable de ce transfert |
| GitHub (GitHub App de l'organisation) | Nom d'utilisateur dans les messages de commit des assistants, lien utilisateur ↔ installation | `services/github_service.py`, `mcp_server.py` | Bearer + lecture | Oui (US) |
| Discord (webhooks) | Noms d'utilisateur dans les alertes d'accès et de sécurité | `services/discord_service.py`, `services/access/grants.py`, `services/security/alerts.py` | Bearer + lecture | Oui (US) |
| Cloudflare | Trafic de tous les utilisateurs (IP, en-têtes) en frontal du portail et de Keycloak ; l'API DNS est appelée sans donnée personnelle | `services/project_preflight.py`, `keycloak_service.py:81` | Bearer + lecture | Oui (US) |
| AWS | Pas de donnée personnelle observée : facturation, état Terraform. Le client S3 `boto3` sert aussi aux avatars, mais vers `s3.3istor.com` | `services/finops/ingest/aws_focus.py`, `routers/account.py` | Bearer (S3) | Selon région |
| Google Cloud | Pas de donnée personnelle observée : facturation via fédération d'identité | `services/finops/ingest/gcp_focus.py`, `federation.py` | Bearer | Selon région |
| Grafana (interne) | Email et nom des membres, synchronisés dans les orgs Grafana | `services/grafana_service.py` | Bearer (email) | Non, hébergé sur CNP |
| Keycloak, Loki, Vault, Ceph RGW (internes) | Voir T1 et T3 | | Lecture | Non, hébergés sur CNP |

## Mesures de sécurité (art. 32), résumé

Tiré du code, détaillé dans [Ségrégation des accès](../01-architecture/05-access-segregation.md),
[Journaux d'audit](../01-architecture/06-audit-logging.md) et
[Matrice OWASP MCP](02-matrice-owasp-mcp.md).

- Tokens Keycloak vérifiés côté backend (signature, émetteur, `azp`, `typ`), rôles de
  projet relus en direct dans Keycloak.
- Tokens d'assistants opaques, stockés hachés, utilisables seulement par le serveur MCP.
- Journal d'audit de chaque appel qui modifie, champs filtrés et secrets masqués.
- Archive d'audit inaltérable 30 j.

## À décider

- [ ] Responsable de traitement et contact DPO.
- [ ] Finalité et base légale de T1 à T6.
- [ ] Durées cibles, en particulier : T1 au départ d'une personne, T2, T4 (grants
      révoqués, approbations), T6, et l'expiration de l'archive d'audit après le verrou.
- [ ] Statut et contrat de chaque destinataire hors UE, en premier lieu les assistants IA.
- [ ] Politique d'usage des assistants IA : quels assistants sont autorisés, et si les
      logs d'apps manipulant des données d'utilisateurs finaux peuvent leur être envoyés.
- [ ] Photo de profil servie publiquement : acceptable ou non.
- [ ] Noms d'utilisateur dans les logs applicatifs (Loki 14 j) : acceptables ou à retirer.
