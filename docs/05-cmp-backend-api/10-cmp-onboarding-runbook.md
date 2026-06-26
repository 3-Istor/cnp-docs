# Onboarding Runbook: Developer Setup Guide

This runbook is a step-by-step checklist designed to get a new developer or platform administrator from a clean local machine to their first successful deployment on the Cloud Native Platform (CNP).

---

## 📋 Prerequisites & Local Requirements

Ensure your machine has the following tools installed:
* [ ] **Python**: version `3.11` exactly.
* [ ] **Node.js**: version `20` or higher.
* [ ] **Docker**: version `20.10+` with Docker Compose V2.
* [ ] **Kubectl**: configured to access your development/lab K3s cluster.

---

## 🚀 Setup Checklist

### Step 1: Clone the Codebase
Clone the control plane repository and navigate to its root folder.
```bash
git clone https://github.com/3-Istor/arcl-cmp.git
cd arcl-cmp
```

### Step 2: Initialize Environment Variables
Generate the local `.env` configuration file from the provided template.
```bash
# Copy template to active env
cp .env.example .env
```

Edit the generated `.env` file to configure your local credentials:
```bash
# backend/.env extract
DB_ENABLED=true
DB_TYPE=sqlite
SQLITE_PATH=./app.db

# Vault & Keycloak development tokens
VAULT_TOKEN="hvs.xxxxxxxxxxxxxxxxx"
KEYCLOAK_ADMIN_PASSWORD="admin-password"
```

### Step 3: Run Database Migrations
Initialize the local SQLite database schema.
```bash
cd backend
python3.11 -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt

# Run migrations (or initialize db schema)
export PYTHONPATH=$PYTHONPATH:.
pytest
```

### Step 4: Boot up the Local Stack
Use Docker Compose to run the entire architecture locally in hot-reloaded development mode.
```bash
# Start backend, frontend, and database
docker compose up --build -d
```

---

## 🔍 Post-Setup Verification

To confirm your local environment is correctly configured, run the following verification steps:

### Test 1: Verify API Gateway is up
Query the backend root endpoint:
```bash
curl -X GET "http://localhost:8081/"
```
**Expected Response (`200 OK`)**:
```json
{
  "message": "Bienvenue sur l'API Cloud Native Template",
  "version": "1.0.0",
  "endpoints": {
    "hello": "/api/hello",
    "database": "/api/db/status",
    "health": "/health"
  }
}
```

### Test 2: Verify Database Connectivity
Check if the local SQLite or PostgreSQL database is connected and unblocked:
```bash
curl -X GET "http://localhost:8081/api/db/status"
```
**Expected Response (`200 OK`)**:
```json
{
  "enabled": true,
  "type": "sqlite",
  "connected": true
}
```

---

## 🔑 Default Test Accounts & Seed Credentials

If your dev stack is configured against the shared staging cluster, use these credentials to log in:

* **Keycloak Admin Panel**: `https://admin-auth.3istor.com`
  * **Username**: `admin`
  * **Password**: *Refer to Vault path `kvv2/keycloak/admin`*
* **Test Client Profile**:
  * **Username**: `client.test@epita.fr`
  * **Password**: *Generates dynamically. Check output of `terraform output client_test_password`*

---

## 🚨 Troubleshooting Common Setup Obstacles

### 1. "Database is locked" (SQLite)
* **Symptom**: FastAPI requests hang indefinitely.
* **Reason**: Parallel write processes (e.g. background Terraform runs) are colliding on the single SQLite file.
* **Fix**: Ensure WAL (Write-Ahead Logging) is enabled in `app/core/database.py`. If locked, run:
  ```bash
  sqlite3 backend/app.db "PRAGMA journal_mode=WAL;"
  ```

### 2. "Unauthorized" K8s API
* **Symptom**: Terraform execution fails with `Error: Unauthorized` during day-0 provisioning.
* **Reason**: The Docker container running the backend has no valid `~/.kube/config` mapped or lacks active cluster permissions.
* **Fix**: Ensure your local kubeconfig is correctly mapped in the `volumes` section of your `docker-compose.yml`:
  ```yaml
  volumes:
    - ~/.kube/config:/root/.kube/config:ro
  ```

---
**Next Step**: Complete the setup and start deploying assets from your CMP Dashboard.
