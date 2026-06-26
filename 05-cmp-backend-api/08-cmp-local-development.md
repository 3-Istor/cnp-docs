# Local Development Specifications

This document outlines the procedure to execute and run the Cloud Native Platform (CNP) ecosystem locally for development and testing.

---

## 1. Local Network Port Topology

In a local development setup, the frontend and backend communicate directly, bypassing the Envoy Gateway proxy via Vite's internal development proxy configuration.

```text
[ Developer Browser ]
       │
       ├─► (Port 3001) ──► Vite Dev Server (React Frontend)
       │                        │
       │                        ▼ (Proxy rule: /api/*)
       └─► (Port 8000) ──► Uvicorn ASGI Server (FastAPI Backend)
                                │
                                └─► Local SQLite File (app.db) / Postgres (5432)
```

| Component | Service | Port | Host Address |
| :--- | :--- | :--- | :--- |
| **Frontend** | React / Vite | `3001` | `http://localhost:3001` |
| **Backend** | FastAPI / Uvicorn | `8000` | `http://localhost:8000` |
| **Database** | Postgres (Docker Option) | `5432` | `localhost:5432` |

---

## 2. Environment Variables & Configurations

### A. Backend Configuration (`backend/.env`)
Create a `.env` file inside the `backend/` directory.

```bash
# Database Configuration
DB_ENABLED=true
DB_TYPE=sqlite
SQLITE_PATH=./app.db

# GitHub App Integration (CNP-Portal)
GITHUB_APP_PRIVATE_KEY=""
GITHUB_INSTALLATION_ID="98765432"

# Keycloak IdP integration
KEYCLOAK_URL=https://auth.3istor.com
KEYCLOAK_CLIENT_ID=arcl-cmp
KEYCLOAK_CLIENT_SECRET=my-client-secret
KEYCLOAK_ADMIN_USERNAME=admin
KEYCLOAK_ADMIN_PASSWORD=admin-password

# HashiCorp Vault
VAULT_URL=https://vault.3istor.com
VAULT_TOKEN=hvs.xxxxxxxxxxxxxxxxxxxxxxxx

# Cloudflare Routing
CLOUDFLARE_API_TOKEN=cloudflare-dns-token
CLOUDFLARE_ZONE_ID=cloudflare-zone-id
CLOUDFLARE_ACCOUNT_ID=cloudflare-account-id
```

### B. Frontend Proxy Config (`frontend/vite.config.js`)
Vite is preconfigured to proxy `/api` calls directly to the local backend. This removes CORS complexities during development.

```javascript
// vite.config.js extract
export default defineConfig({
  server: {
    port: 3001,
    proxy: {
      '/api': {
        target: process.env.BACKEND_HOST || 'http://localhost:8000',
        changeOrigin: true,
        rewrite: (path) => path
      }
    }
  }
})
```

---

## 3. Step-by-Step Local Execution

### A. Run Backend (FastAPI)
1. **Navigate and install dependencies**:
   ```bash
   cd backend
   poetry install
   ```
2. **Apply migrations**:
   ```bash
   poetry run alembic upgrade head
   ```
3. **Launch the ASGI development server**:
   ```bash
   poetry run uvicorn app.main:app --reload --host 0.0.0.0 --port 8000
   ```
   *Verify API Swagger availability by accessing: `http://localhost:8000/docs`*

### B. Run Frontend (React + Vite)
1. **Navigate and install Node modules**:
   ```bash
   cd frontend
   npm ci
   ```
2. **Launch the development hot-reloading server**:
   ```bash
   npm run dev
   ```
   *Open your browser and navigate to: `http://localhost:3001`*

---

## 🧪 Testing Backend Locally

Before committing changes, execute the test suite to ensure validation and endpoints meet standards:

```bash
cd backend
export PYTHONPATH=$PYTHONPATH:.
poetry run pytest
```

---
**Next Step**: Continue to [Containerization & Docker Compose Specs](09-cmp-containers.md) (or return to the [Project Overview](../README.md)).
