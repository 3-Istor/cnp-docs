# Phase 3 Changes Summary

**Date**: 2026-05-24
**Status**: ✅ Complete and Verified

## Overview

Phase 3 adds Kubernetes provider support to the CMP backend, enabling modern GitOps-based deployments alongside the existing legacy hybrid (OpenStack + AWS) infrastructure.

---

## What Changed

### 1. Multi-Provider Architecture

Le backend supporte maintenant deux types de déploiements :

- **`legacy_hybrid`** : OpenStack VMs + AWS ASG (existant, inchangé)
- **`kubernetes`** : GitHub + Terraform + ArgoCD (nouveau)

### 2. Nouveaux Champs dans l'API

Tous les endpoints `/api/deployments` retournent maintenant ces champs supplémentaires :

```typescript
interface Deployment {
  // Existant
  id: number;
  name: string;
  template_id: string;
  status: DeploymentStatus;

  // NOUVEAU - Phase 3
  provider_type: "legacy_hybrid" | "kubernetes";
  project_id?: string;
  github_repo_url?: string;
  argocd_app_name?: string;
  k8s_namespace?: string;
}
```

### 3. GitHub App Integration

Pour les déploiements Kubernetes, l'utilisateur doit lier son compte GitHub :

1. **Frontend** : Bouton "Link GitHub Account" dans la page Account
2. **OAuth Flow** : Redirection vers GitHub App installation
3. **Callback** : Stockage du `installation_id` dans Keycloak
4. **Utilisation** : Passé dans `app_config` lors de la création

---

## Impact Frontend

### Changements Requis (Phase 4)

#### 1. Catalog View - Séparation IaaS / PaaS

```typescript
// Avant
<CatalogGrid templates={allTemplates} />

// Après
<Tabs>
  <Tab label="Platform as a Service (PaaS)">
    <CatalogGrid
      templates={templates.filter(t => t.category === 'paas')}
    />
  </Tab>
  <Tab label="Infrastructure as a Service (IaaS)">
    <CatalogGrid
      templates={templates.filter(t => t.category === 'iaas')}
    />
  </Tab>
</Tabs>
```

#### 2. Account Page - GitHub Linking

```typescript
// Nouveau composant
<GitHubLinkButton
  isLinked={user.github_installation_id !== null}
  onLink={() => {
    window.location.href =
      'https://github.com/apps/cnp-portal/installations/new';
  }}
/>
```

#### 3. Deployment Cards - Conditional Rendering

```typescript
function DeploymentCard({ deployment }: Props) {
  if (deployment.provider_type === 'kubernetes') {
    return (
      <Card>
        <Title>{deployment.name}</Title>
        <Status>{deployment.status}</Status>

        {/* Kubernetes-specific actions */}
        <Actions>
          <Button href={deployment.github_repo_url}>
            📦 GitHub Repo
          </Button>
          <Button href={`https://argocd.3istor.com/applications/${deployment.argocd_app_name}`}>
            🐙 ArgoCD
          </Button>
          <Button href={`https://vault.3istor.com/ui/vault/secrets/kvv2/show/projects/${deployment.project_id}/${deployment.name}`}>
            🔒 Secrets
          </Button>
        </Actions>
      </Card>
    );
  }

  // Legacy hybrid rendering
  return <LegacyDeploymentCard deployment={deployment} />;
}
```

#### 4. Create Deployment Form

```typescript
function CreateDeploymentForm() {
  const [providerType, setProviderType] = useState<'kubernetes' | 'legacy_hybrid'>('kubernetes');

  return (
    <Form>
      <Select
        label="Provider Type"
        value={providerType}
        onChange={setProviderType}
      >
        <option value="kubernetes">Kubernetes (GitOps)</option>
        <option value="legacy_hybrid">Legacy Hybrid</option>
      </Select>

      {providerType === 'kubernetes' && (
        <>
          <Input
            label="Project Name"
            name="project_name"
            required
          />
          <Input
            label="Replica Count"
            name="replica_count"
            type="number"
            defaultValue={2}
          />
          <Checkbox
            label="Enable SSO Protection"
            name="sso_protected"
          />
        </>
      )}

      {/* Legacy fields */}
    </Form>
  );
}
```

---

## Changements Backend

### Database Schema

Nouvelle migration : `c4d8f2a91b3e_add_kubernetes_provider_support`

```sql
ALTER TABLE deployments ADD COLUMN provider_type VARCHAR(13) DEFAULT 'legacy_hybrid';
ALTER TABLE deployments ADD COLUMN project_id VARCHAR(100);
ALTER TABLE deployments ADD COLUMN github_repo_url VARCHAR(255);
ALTER TABLE deployments ADD COLUMN argocd_app_name VARCHAR(100);
ALTER TABLE deployments ADD COLUMN k8s_namespace VARCHAR(100);
```

### Nouveaux Services

1. **`app/services/github_service.py`**
   - `generate_jwt()` : Génère un JWT signé avec la clé privée GitHub App
   - `get_installation_token()` : Échange le JWT contre un token d'installation
   - `create_repository()` : Crée un repo privé via l'API GitHub

2. **`app/terraform/github_bootstrap/`**
   - Module Terraform pour le Day-0 provisioning
   - Crée : GitHub repo, K8s namespace, Vault secrets, ArgoCD Application

3. **Saga Orchestrator Refactoring**
   - Route les déploiements selon `provider_type`
   - Nouveau flow Kubernetes : GitHub → Terraform → ArgoCD
   - Flow legacy préservé : OpenStack → AWS

---

## Compatibilité

### ✅ Backward Compatible

- Tous les déploiements existants continuent de fonctionner
- `provider_type` par défaut : `legacy_hybrid`
- Aucun changement breaking dans l'API
- Migration de base de données additive uniquement

### Nouveaux Champs Optionnels

Les champs Kubernetes sont `null` pour les déploiements legacy :

```json
{
  "id": 1,
  "provider_type": "legacy_hybrid",
  "project_id": null,
  "github_repo_url": null,
  "argocd_app_name": null,
  "k8s_namespace": null
}
```

---

## Configuration Requise

### Backend (.env)

```bash
# GitHub App Integration (requis pour Kubernetes)
GITHUB_APP_PRIVATE_KEY=""

# S3 Backend (requis pour Terraform state)
TF_BACKEND_S3_ENABLED=true
TF_BACKEND_S3_BUCKET=3-istor-tf-infra-aws
TF_BACKEND_S3_DYNAMODB_TABLE=terraform-state-lock
```

### Frontend

Aucune configuration supplémentaire requise. Les changements sont dans le code uniquement.

---

## Testing

### Backend

```bash
cd backend

# Vérifier que tout fonctionne
poetry run python verify_phase3.py

# Tester l'intégration GitHub
poetry run python test_github_service.py
```

### Frontend (à venir Phase 4)

```bash
cd frontend

# Tests unitaires pour les nouveaux composants
npm test -- DeploymentCard
npm test -- GitHubLinkButton

# Tests d'intégration
npm run test:e2e
```

---

## Documentation

### Backend

- `backend/PHASE3_COMPLETE.md` : Guide complet
- `backend/QUICKSTART_PHASE3.md` : Setup rapide
- `backend/PHASE3_IMPLEMENTATION.md` : Détails techniques
- `backend/PHASE3_ARCHITECTURE.md` : Diagrammes

### API

- `.kiro/steering/docs/05-backend-api/01-deployment-api.md` : Spécification API complète

### Architecture

- `.kiro/steering/docs/01-architecture/01-system-overview.md` : Vue d'ensemble
- `.kiro/steering/docs/03-pipelines-and-workflows/01-app-provisioning-flow.md` : Flow de provisioning

---

## Prochaines Étapes (Phase 4)

### Frontend Refactoring

1. **Dual Catalog View** : Séparer les templates IaaS et PaaS
2. **GitHub Account Linking** : OAuth flow dans la page Account
3. **Dynamic Deployment Cards** : UI conditionnelle selon `provider_type`
4. **ArgoCD Health Integration** : Statut de sync en temps réel
5. **Day-2 Operations UI** : Scaling, toggle SSO, rollback

### Estimation

- **Durée** : 2-3 jours
- **Complexité** : Moyenne
- **Dépendances** : Phase 3 (✅ Complete)

---

## Support

### Questions Backend

Voir la documentation dans `backend/PHASE3_*.md`

### Questions Frontend

Voir `.kiro/steering/docs/05-backend-api/01-deployment-api.md` pour l'intégration API

### Questions Architecture

Voir `.kiro/steering/docs/` pour la documentation complète du système
