# Containerization & Docker Compose Specifications

For consistent, production-like local development and deployment testing, CNP leverages containerized execution topologies through Docker and Docker Compose.

---

## 1. Local Containerized Network Topology

When executing under Docker Compose, the services run within an isolated virtual network, mapping specific external ports to the host:

```text
[ Developer Host ]
       │
       ├─► (Port 3001) ──► [ react-frontend Container ] (Port 3001)
       │                        │
       │                        ▼ (Internal network: http://backend:8000)
       ├─► (Port 8081) ──► [ fastapi-backend Container ] (Port 8000)
       │                        │
       │                        ▼ (Internal network: postgres-db:5432)
       └─► (Port 5432) ──► [ postgres-db Container ] (Port 5432)
```

---

## 2. Dockerfile Specifications

### A. Backend (`backend/Dockerfile`)
The backend uses a single-stage, lightweight Debian-based python image optimized for faster image sizes and low runtime overhead.

```dockerfile
FROM python:3.11-slim

WORKDIR /app

# Install system dependencies (including git for template repo sync)
RUN apt-get update && apt-get install -y \
    git \
    && rm -rf /var/lib/apt/lists/*

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

EXPOSE 8000

CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8000"]
```

### B. Frontend (`frontend/Dockerfile`)
The frontend implements a **multi-stage build** to optimize dependency footprints and execution speed.

```dockerfile
# Stage 1: Build dependencies
FROM node:20-slim AS build
WORKDIR /app
COPY package.json .
RUN npm install
COPY . .

# Stage 2: Runtime image
FROM node:20-slim
WORKDIR /app
COPY --from=build /app .
EXPOSE 3001
CMD ["npm", "run", "dev"]
```

---

## 3. Docker Compose Orchestration

The `docker-compose.yml` integrates the frontend, backend, and an optional PostgreSQL database using Docker profiles.

```yaml
services:
  backend:
    build: ./backend
    image: ghcr.io/3-istor/template-app-webapp-python-fastapi-react/backend:latest
    container_name: fastapi-backend
    ports:
      - "8081:8000"
    volumes:
      - ./backend:/app
      - ~/.kube/config:/root/.kube/config:ro   # ☸️ Crucial: Mounts K8s auth for Terraform inside container
    env_file:
      - path: .env
        required: false
    environment:
      PYTHONUNBUFFERED: 1

  frontend:
    build: ./frontend
    image: ghcr.io/3-istor/template-app-webapp-python-fastapi-react/frontend:latest
    container_name: react-frontend
    ports:
      - "3001:3001"
    volumes:
      - ./frontend:/app
      - /app/node_modules
    depends_on:
      - backend

  db:
    image: postgres:16-alpine
    container_name: postgres-db
    profiles:
      - postgres                               # Only boots up when explicitly requested
    ports:
      - "5432:5432"
    environment:
      POSTGRES_DB: ${DB_NAME:-app_db}
      POSTGRES_USER: ${DB_USER:-postgres}
      POSTGRES_PASSWORD: ${DB_PASSWORD:-postgres}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U \${DB_USER:-postgres} -d \${DB_NAME:-app_db}"]
      interval: 5s
      timeout: 5s
      retries: 5

volumes:
  postgres_data:
```

---

## 4. Runbooks: Orchestrating Containers

### Start standard stack (SQLite mode - Fast)
This launches the frontend and backend instantly, persisting data to the local SQLite `app.db` file.
```bash
docker compose up --build
```

### Start full-resilience stack (PostgreSQL mode)
This activates the `postgres` profile, bootstrapping a dedicated database container and verifying its health status before making connections.
```bash
docker compose --profile postgres up --build
```

### Stop and clean state
Wipes container instances and internal network configurations.
```bash
docker compose down -v
```

---
**Next Step**: Continue to [Onboarding Runbook & Guide](10-cmp-onboarding-runbook.md) (or return to the [Project Overview](../README.md)).
