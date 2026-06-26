# Variables d'environnement de la Plateforme (CNP) ⚡

Ce document recense l'intégralité des variables d'environnement configurées et utilisées au sein de l'écosystème Cloud Native Platform (CNP). Ces variables sont réparties entre le Backend (FastAPI), le Frontend (Next.js/Vite), la Base de données, les runners Terraform de bootstrapping, et les pipelines d'intégration continue (CI/CD).

Pour des détails sur la configuration locale ou l'exécution de ces composants, consultez :
* Le guide de développement local : [08-cmp-local-development.md](file:///home/brian/Documents/aepita/ing2/3-Istor/cnp-docs/05-cmp-backend-api/08-cmp-local-development.md)
* Le guide des conteneurs : [09-cmp-containers.md](file:///home/brian/Documents/aepita/ing2/3-Istor/cnp-docs/05-cmp-backend-api/09-cmp-containers.md)
* Le fonctionnement du provisionneur : [03-terraform-provisioner.md](file:///home/brian/Documents/aepita/ing2/3-Istor/cnp-docs/04-templates/03-terraform-provisioner.md)

---

## 📋 Table de Référence des Variables d'Environnement

Le tableau ci-dessous regroupe toutes les variables d'environnement utilisées dans le projet :

| Nom | Composant Cible | Type | Statut | Description |
| :--- | :--- | :--- | :--- | :--- |
| **`DB_ENABLED`** | Backend (FastAPI) | `Boolean` | Optionnelle | Active ou désactive la connexion à la base de données. Si désactivée, les appels d'API dépendant de la DB échoueront. (Défaut : `true`). |
| **`DB_TYPE`** | Backend (FastAPI) | `String` | Optionnelle | Spécifie le type de base de données pour le stockage CMP. Supporte principalement `sqlite` et `postgres`. (Défaut : `sqlite`). |
| **`SQLITE_PATH`** | Backend (FastAPI) | `String` | Optionnelle | Chemin vers le fichier de base de données SQLite local. Requis uniquement si `DB_TYPE` est configuré sur `sqlite`. (Défaut : `./app.db`). |
| **`DB_NAME`** | Backend / Docker Compose | `String` | Optionnelle | Nom de la base de données PostgreSQL externe ou de conteneur. (Défaut : `app_db`). |
| **`DB_USER`** | Backend / Docker Compose | `String` | Optionnelle | Nom d'utilisateur de connexion à la base de données PostgreSQL. (Défaut : `postgres`). |
| **`DB_PASSWORD`** | Backend / Docker Compose | `String` | Optionnelle | Mot de passe associé à l'utilisateur de base de données PostgreSQL. (Défaut : `postgres`). |
| **`GITHUB_APP_PRIVATE_KEY`** | Backend / Terraform | `String` | **Requise** | Clé privée RSA (format PEM) de l'application GitHub (`CNP-Portal`) pour l'authentification et l'échange de jetons d'accès temporaires (60 min). |
| **`GITHUB_INSTALLATION_ID`** | Backend (FastAPI) | `String` / `Integer` | **Requise** | Identifiant de l'installation de l'application GitHub sur le compte ou l'organisation ciblée pour l'automatisation. |
| **`KEYCLOAK_URL`** | Backend (FastAPI) | `URL` | **Requise** | URL racine du serveur d'identité Keycloak (IdP) pour la gestion du SSO et des protocoles OIDC (ex: `https://auth.3istor.com`). |
| **`KEYCLOAK_CLIENT_ID`** | Backend (FastAPI) | `String` | **Requise** | Identifiant client OIDC défini dans Keycloak pour l'application CMP (ex: `arcl-cmp`). |
| **`KEYCLOAK_CLIENT_SECRET`** | Backend (FastAPI) | `String` | **Requise** | Secret client OIDC associé au client ID Keycloak confidentiel pour l'authentification backend. |
| **`KEYCLOAK_ADMIN_USERNAME`** | Backend (FastAPI) | `String` | Optionnelle | Nom d'utilisateur de l'administrateur Keycloak pour la gestion des API d'administration et la synchronisation des locataires. (Défaut : `admin`). |
| **`KEYCLOAK_ADMIN_PASSWORD`** | Backend / Terraform | `String` | **Requise** | Mot de passe associé à l'administrateur Keycloak pour les appels d'API admin (passé également en variable de runner Terraform). |
| **`VAULT_URL`** | Backend (FastAPI) | `URL` | **Requise** | URL de base de l'instance HashiCorp Vault (ex: `https://vault.3istor.com`). |
| **`VAULT_TOKEN`** | Backend / Terraform | `String` | **Requise** | Jeton d'accès Vault (`hvs.xxx`) permettant au backend et à Terraform de créer des chemins de secrets et de les y injecter. |
| **`CLOUDFLARE_API_TOKEN`** | Backend (FastAPI) | `String` | **Requise** | Jeton d'accès API Cloudflare pour configurer dynamiquement les DNS et tunnels lors du déploiement. |
| **`CLOUDFLARE_ZONE_ID`** | Backend (FastAPI) | `String` | **Requise** | Identifiant de la zone DNS Cloudflare associée au domaine géré (ex: `3istor.com`). |
| **`CLOUDFLARE_ACCOUNT_ID`** | Backend (FastAPI) | `String` | **Requise** | Identifiant de compte Cloudflare hébergeant les tunnels de routage `cloudflared`. |
| **`TF_BACKEND_S3_ENABLED`** | Backend (FastAPI) | `Boolean` | Optionnelle | Indique si le stockage distant de l'état Terraform (Micro-State) doit s'effectuer à distance sur AWS S3. (Défaut : `false`). |
| **`TF_BACKEND_S3_BUCKET`** | Backend (FastAPI) | `String` | **Requise** (si S3 actif) | Nom du bucket AWS S3 hébergeant les fichiers d'état `.tfstate` du CMP. |
| **`TF_BACKEND_S3_DYNAMODB_TABLE`** | Backend (FastAPI) | `String` | **Requise** (si S3 actif) | Nom de la table DynamoDB de verrouillage d'état concurrent pour Terraform. |
| **`TF_BACKEND_AWS_REGION`** | Backend (FastAPI) | `String` | **Requise** (si S3 actif) | Région AWS hébergeant le bucket S3 et la table DynamoDB de verrouillage d'état. |
| **`PYTHONUNBUFFERED`** | Backend (Docker) | `Boolean` (`0`/`1`) | Optionnelle | Désactive la mise en mémoire tampon de la sortie standard de Python, pour un affichage direct des logs de conteneur. (Défaut : `1`). |
| **`PYTHONPATH`** | Backend (Local) | `String` | Optionnelle | Configure le chemin de recherche des modules Python dans l'environnement local (ex: `$PYTHONPATH:.`). |
| **`BACKEND_HOST`** | Frontend (Vite) | `URL` | Optionnelle | URL cible du backend FastAPI, utilisée par le proxy de développement Vite pour rediriger les requêtes vers `/api`. (Défaut : `http://localhost:8000`). |
| **`POSTGRES_DB`** | Base de données (DB) | `String` | **Requise** (interne) | Configure le nom de la base de données PostgreSQL générée au premier démarrage du conteneur (dérivé de `DB_NAME`). |
| **`POSTGRES_USER`** | Base de données (DB) | `String` | **Requise** (interne) | Configure le nom d'utilisateur administrateur PostgreSQL généré au premier démarrage (dérivé de `DB_USER`). |
| **`POSTGRES_PASSWORD`** | Base de données (DB) | `String` | **Requise** (interne) | Configure le mot de passe administrateur PostgreSQL généré au premier démarrage (dérivé de `DB_PASSWORD`). |
| **`TF_VAR_vault_token`** | Terraform Runner | `String` | **Requise** (injectée) | Jeton Vault injecté automatiquement par le backend dans l'environnement du processus Terraform. |
| **`TF_VAR_keycloak_admin_password`** | Terraform Runner | `String` | **Requise** (injectée) | Mot de passe de l'administrateur Keycloak injecté pour mapper les identités OIDC des tenants. |
| **`TF_VAR_github_app_private_key`** | Terraform Runner | `String` | **Requise** (injectée) | Clé privée de l'app GitHub injectée pour la configuration initiale du dépôt Git d'application. |
| **`TF_VAR_github_token`** | Terraform Runner | `String` | **Requise** (injectée) | Jeton temporaire d'installation (valide 60 min) généré dynamiquement et injecté sous forme de variable Terraform. |
| **`REGISTRY`** | CI/CD (GitHub Actions) | `String` / `URL` | Optionnelle | Adresse du registre de conteneurs cible pour les builds d'images (Défaut : `ghcr.io`). |
| **`IMAGE_NAME`** | CI/CD (GitHub Actions) | `String` | Optionnelle | Nom de l'image de conteneur générée par GitHub Actions, basée sur `${{ github.repository }}`). |
| **`GITHUB_TOKEN`** | CI/CD (GitHub Actions) | `String` | **Requise** (fournie) | Secret fourni dynamiquement par le runner GitHub Actions pour s'authentifier auprès de GHCR. |

---

## 🔒 Sécurité et Injection Dynamique (`TF_VAR_*`)

Pour empêcher toute fuite d'informations sensibles (tokens Vault, secrets Keycloak, clés d'API) dans l'historique d'exécution système (visible via les commandes `ps aux` ou les logs de processus), le backend du CMP n'utilise jamais de drapeaux de ligne de commande `-var`.

À la place, il utilise le préfixe `TF_VAR_` reconnu nativement par Terraform. Les variables correspondantes sont injectées directement dans l'environnement système du sous-processus d'exécution :

```python
# app/services/terraform_executor.py
env = os.environ.copy()
env["TF_VAR_vault_token"] = settings.VAULT_TOKEN
env["TF_VAR_keycloak_admin_password"] = settings.KEYCLOAK_ADMIN_PASSWORD
env["TF_VAR_github_app_private_key"] = settings.GITHUB_APP_PRIVATE_KEY
```

---

## 🐋 Orchestration de Base de Données (Docker Compose)

Lorsqu'on utilise la base de données PostgreSQL intégrée via le profil docker compose `postgres`, le conteneur PostgreSQL (`db`) consomme les variables d'environnement système pour instancier la base et configurer les privilèges de connexion :

```yaml
# Extrait du docker-compose.yml
db:
  image: postgres:16-alpine
  container_name: postgres-db
  profiles:
    - postgres
  ports:
    - "5432:5432"
  environment:
    POSTGRES_DB: ${DB_NAME:-app_db}
    POSTGRES_USER: ${DB_USER:-postgres}
    POSTGRES_PASSWORD: ${DB_PASSWORD:-postgres}
```

*Note : Les variables `${DB_NAME}`, `${DB_USER}`, et `${DB_PASSWORD}` proviennent du fichier local `.env` configuré à la racine du projet et sont transmises par substitution.*
