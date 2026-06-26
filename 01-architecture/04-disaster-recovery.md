# Plan de Reprise d'Activité (PRA) et Restauration du Control Plane 🌋

Ce document définit les procédures de secours, les commandes d'administration système et les scénarios de restauration en cas de panne majeure affectant le Control Plane de la **Cloud Management Platform (CNP)**.

Pour la description générale des composants et des topologies associés, consultez :
* L'architecture globale : [01-system-overview.md](file:///home/brian/Documents/aepita/ing2/3-Istor/cnp-docs/01-architecture/01-system-overview.md)
* La configuration ArgoCD : [04-gitops-argocd.md](file:///home/brian/Documents/aepita/ing2/3-Istor/cnp-docs/02-core-components/04-gitops-argocd.md)
* Le stockage d'état Terraform : [03-terraform-provisioner.md](file:///home/brian/Documents/aepita/ing2/3-Istor/cnp-docs/04-templates/03-terraform-provisioner.md)

---

## 🏛️ Scénario 1 : Perte de la base de données du CMP (SQLite/PostgreSQL)

Dans ce scénario, la base de données contenant les informations de déploiements, de projets, de liaison de comptes GitHub et de statuts est corrompue ou détruite.

### A. Restauration depuis les sauvegardes régulières

#### 1. Cas SQLite (Développement/Staging local)
Si le backend tourne en mode SQLite (`DB_TYPE=sqlite`), la base de données est contenue dans le fichier défini par `SQLITE_PATH` (ex: `app.db`).

* **Sauvegarde manuelle à chaud** :
```bash
# Sauvegarde sécurisée de la base SQLite en activant le mode de sauvegarde en ligne
sqlite3 backend/app.db ".backup 'backend/app.db.bak'"
```

* **Procédure de restauration** :
```bash
# Arrêter le backend FastAPI pour libérer les verrous
docker compose stop backend

# Remplacer le fichier corrompu par la sauvegarde
mv backend/app.db.bak backend/app.db

# Relancer le backend
docker compose start backend
```

#### 2. Cas PostgreSQL (Production)
Si la production utilise un conteneur ou une instance PostgreSQL.

* **Sauvegarde automatique (cron quotidien)** :
```bash
docker exec -t postgres-db pg_dump -U postgres -d app_db -F c -b -v -f /var/lib/postgresql/data/db_backup.dump
```

* **Restauration de la base de données** :
```bash
# 1. Arrêter le backend FastAPI pour couper les connexions actives
docker compose stop backend

# 2. Recréer une base de données propre
docker exec -it postgres-db psql -U postgres -c "DROP DATABASE IF EXISTS app_db;"
docker exec -it postgres-db psql -U postgres -c "CREATE DATABASE app_db;"

# 3. Restaurer le dump PostgreSQL
docker exec -it postgres-db pg_restore -U postgres -d app_db -v /var/lib/postgresql/data/db_backup.dump

# 4. Redémarrer le service backend
docker compose start backend
```

---

### B. Reconstruction depuis l'État Réel (S3 & Git - Zero DB Backup Fallback)

Si aucune sauvegarde de la base de données n'est disponible, l'architecture CNP permet de reconstruire l'intégralité de la base de données à partir de l'état réel des infrastructures (**Single Source of Truth**), contenu dans les dépôts Git privés et les fichiers d'état Terraform stockés sur AWS S3.

#### Script de réconciliation d'urgence
Un script Python d'administration (`backend/app/scripts/reconstruct_db.py`) est fourni pour scanner le bucket S3 contenant les fichiers `.tfstate` et réinjecter les déploiements dans la DB :

```python
# backend/app/scripts/reconstruct_db.py
import json
import boto3
from app.core.database import SessionLocal
from app.models.deployment import Deployment

s3 = boto3.client("s3")
db = SessionLocal()

def reconstruct():
    # Scanner les préfixes cmp/projects/ pour lister les états
    bucket = "3-istor-tf-infra-aws"
    response = s3.list_objects_v2(Bucket=bucket, Prefix="cmp/projects/")
    
    for obj in response.get("Contents", []):
        key = obj["Key"]
        if not key.endswith(".tfstate"):
            continue
            
        # Exemple de chemin : cmp/projects/{project_id}/{deployment_name}.tfstate
        parts = key.split("/")
        project_id = parts[2]
        deployment_name = parts[3].replace(".tfstate", "")
        
        # Récupérer l'état Terraform depuis S3
        state_obj = s3.get_object(Bucket=bucket, Key=key)
        state_data = json.loads(state_obj["Body"].read().decode("utf-8"))
        
        # Extraire les outputs du state
        outputs = state_data.get("outputs", {})
        repo_url = outputs.get("github_repo_url", {}).get("value")
        ns = outputs.get("kubernetes_namespace", {}).get("value")
        
        # Enregistrer à nouveau le déploiement dans la DB
        existing = db.query(Deployment).filter_by(name=deployment_name, project_id=project_id).first()
        if not existing:
            new_dep = Deployment(
                name=deployment_name,
                project_id=project_id,
                github_repo_url=repo_url,
                k8s_namespace=ns,
                provider_type="kubernetes",
                status="RUNNING" # L'infrastructure existe déjà
            )
            db.add(new_dep)
            db.commit()
            print(f"Réconcilié : {project_id}/{deployment_name}")

if __name__ == "__main__":
    reconstruct()
```

---

## 📡 Scénario 2 : Désynchronisation totale d'ArgoCD avec les dépôts Git

Ce scénario survient lorsque ArgoCD perd l'accès aux dépôts Git (identifiants expirés, révocation de l'application GitHub, corruption de cache d'index) entraînant l'état `Unknown` ou `OutOfSync` de l'intégralité des applications déployées.

### Étape 1 : Diagnostic de la connectivité Git
Vérifier l'état de connexion des dépôts dans ArgoCD :
```bash
# Connexion au serveur ArgoCD en ligne de commande
argocd login argocd.3istor.com --username admin --password "$ARGOCD_ADMIN_PASSWORD" --insecure

# Lister l'état de synchronisation des applications
argocd app list | grep -E "OutOfSync|Unknown"
```

### Étape 2 : Recréation des secrets d'accès GitHub App
Si les jetons d'accès ou secrets GitHub App ont été révoqués, ArgoCD ne peut plus interroger les dépôts privés. Appliquez à nouveau le secret contenant la clé privée de l'application GitHub mise à jour :

```bash
# Générer le manifeste du secret de credentials Git
cat <<EOF > argocd-git-creds.yaml
apiVersion: v1
kind: Secret
metadata:
  name: github-app-creds
  namespace: argocd
  labels:
    argocd.argoproj.io/secret-type: repo-creds
stringData:
  type: git
  url: https://github.com/3-Istor
  githubAppID: "3836905"
  githubAppInstallationID: "98765432"
  githubAppPrivateKey: |
$(cat path/to/new_private_key.pem | sed 's/^/    /')
EOF

# Appliquer le secret sur le cluster Kubernetes
kubectl apply -f argocd-git-creds.yaml
```

### Étape 3 : Nettoyage et rafraîchissement forcé du cache d'ArgoCD
Pour forcer ArgoCD à reconstruire son index local et à vider son cache de validation d'API :

```bash
# Forcer la réconciliation à froid d'une application
argocd app get <app-name> --hard-refresh

# Lancer une synchronisation forcée en élaguant les ressources orphelines
argocd app sync <app-name> --force --prune
```

---

## 🛠️ Gestion des États Terraform Orphelins (Orphaned States)

Lors d'un plantage du backend FastAPI ou d'une annulation forcée du SAGA Orchestrator au milieu d'un déploiement, l'état Terraform peut être bloqué par DynamoDB ou laissé à l'abandon sur S3 sans enregistrement dans la DB locale.

### 1. Libérer un verrou d'état concurrent (DynamoDB Lock Release)
Si le déploiement échoue avec le message d'erreur `Error: Error acquiring the state lock`, cela signifie qu'un verrou d'écriture persiste dans DynamoDB.

1. **Identifier le Lock ID** : Le message d'erreur de Terraform affiche le `ID` du verrou (ex: `c87b-1a2b-3c4d-5e6f`).
2. **Forcer le déverrouillage** :
```bash
# Naviguer dans le dossier temporaire du déploiement
cd /tmp/cmp-deployments/project-alpha/app-frontend/

# Exécuter la commande force-unlock en passant l'ID du verrou
terraform force-unlock c87b-1a2b-3c4d-5e6f
```

### 2. Détecter et supprimer les états orphelins sur S3
Pour lister les fichiers d'état présents sur AWS S3 qui ne possèdent pas de correspondance dans la base de données locale :

```bash
# Lister tous les états enregistrés dans le bucket S3
aws s3 ls s3://3-istor-tf-infra-aws/cmp/projects/ --recursive
```

Si un état doit être nettoyé manuellement (par exemple, après un échec de la phase de destruction) :
```bash
# Télécharger l'état localement pour validation
aws s3 cp s3://3-istor-tf-infra-aws/cmp/projects/alpha/frontend.tfstate ./temp.tfstate

# Supprimer l'état sur S3 pour libérer l'espace
aws s3 rm s3://3-istor-tf-infra-aws/cmp/projects/alpha/frontend.tfstate
```

---
**Next Step**: Continue to [CMP Core Dashboard & API](../02-core-components/01-cmp-dashboard.md) (or return to the [Project Overview](../README.md)).
