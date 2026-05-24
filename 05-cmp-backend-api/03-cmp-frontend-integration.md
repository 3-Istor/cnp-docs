# Frontend Phase 4 Implementation Guide

**For**: Frontend Developers (Next.js/React)
**Prerequisites**: Phase 3 Backend Complete ✅
**Estimated Duration**: 2-3 days

---

## Overview

Phase 4 adapte le frontend Next.js pour supporter les déploiements multi-providers. Le backend expose maintenant de nouveaux champs dans l'API Deployment qu'il faut afficher et utiliser.

---

## Quick Reference

### Nouveaux Champs API

```typescript
interface Deployment {
  // Champs existants (inchangés)
  id: number;
  name: string;
  template_id: string;
  status: DeploymentStatus;

  // NOUVEAUX champs Phase 3
  provider_type: "legacy_hybrid" | "kubernetes"; // Discriminateur
  project_id?: string; // Nom du projet
  github_repo_url?: string; // URL du repo GitHub
  argocd_app_name?: string; // Nom de l'app ArgoCD
  k8s_namespace?: string; // Namespace Kubernetes
}
```

### Endpoints Inchangés

Tous les endpoints existants fonctionnent sans modification :

- `GET /api/deployments` - Liste des déploiements
- `GET /api/deployments/{id}` - Détails d'un déploiement
- `POST /api/deployments` - Créer un déploiement
- `DELETE /api/deployments/{id}` - Supprimer un déploiement

---

## Tasks Checklist

### 1. Catalog View - Dual Tabs ⭐ Priority 1

**Fichier**: `frontend/app/catalog/page.tsx`

**Objectif**: Séparer les templates IaaS (legacy) et PaaS (Kubernetes)

**Avant**:

```tsx
<CatalogGrid templates={allTemplates} />
```

**Après**:

```tsx
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";

export default function CatalogPage() {
  const [templates, setTemplates] = useState<Template[]>([]);

  const paasTemplates = templates.filter((t) => t.category === "paas");
  const iaasTemplates = templates.filter((t) => t.category === "iaas");

  return (
    <Tabs defaultValue="paas">
      <TabsList>
        <TabsTrigger value="paas">🚀 Platform as a Service (PaaS)</TabsTrigger>
        <TabsTrigger value="iaas">
          🖥️ Infrastructure as a Service (IaaS)
        </TabsTrigger>
      </TabsList>

      <TabsContent value="paas">
        <p className="text-muted-foreground mb-4">
          Deploy containerized applications with GitOps (GitHub + ArgoCD)
        </p>
        <CatalogGrid templates={paasTemplates} />
      </TabsContent>

      <TabsContent value="iaas">
        <p className="text-muted-foreground mb-4">
          Deploy VMs on OpenStack and AWS Auto Scaling Groups
        </p>
        <CatalogGrid templates={iaasTemplates} />
      </TabsContent>
    </Tabs>
  );
}
```

**Test**:

- [ ] Les templates PaaS s'affichent dans l'onglet PaaS
- [ ] Les templates IaaS s'affichent dans l'onglet IaaS
- [ ] Le changement d'onglet fonctionne
- [ ] L'onglet PaaS est sélectionné par défaut

---

### 2. GitHub Account Linking ⭐ Priority 1

**Fichier**: `frontend/app/account/page.tsx`

**Objectif**: Permettre aux utilisateurs de lier leur compte GitHub

**Nouveau Composant**: `components/GitHubLinkButton.tsx`

```tsx
"use client";

import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { CheckCircle2, Github } from "lucide-react";

interface GitHubLinkButtonProps {
  userId: string;
}

export function GitHubLinkButton({ userId }: GitHubLinkButtonProps) {
  const [isLinked, setIsLinked] = useState(false);
  const [installationId, setInstallationId] = useState<string | null>(null);

  useEffect(() => {
    // Fetch user's GitHub installation ID from Keycloak
    fetchGitHubStatus();
  }, [userId]);

  const fetchGitHubStatus = async () => {
    try {
      const response = await fetch("/api/user/github-status");
      const data = await response.json();
      setIsLinked(!!data.github_installation_id);
      setInstallationId(data.github_installation_id);
    } catch (error) {
      console.error("Failed to fetch GitHub status:", error);
    }
  };

  const handleLink = () => {
    // Redirect to GitHub App installation
    window.location.href =
      "https://github.com/apps/cnp-portal/installations/new";
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Github className="h-5 w-5" />
          GitHub Integration
        </CardTitle>
        <CardDescription>
          Link your GitHub account to deploy Kubernetes applications
        </CardDescription>
      </CardHeader>
      <CardContent>
        {isLinked ? (
          <div className="flex items-center gap-2 text-green-600">
            <CheckCircle2 className="h-5 w-5" />
            <span>
              GitHub account linked (Installation ID: {installationId})
            </span>
          </div>
        ) : (
          <div className="space-y-4">
            <p className="text-sm text-muted-foreground">
              To deploy Kubernetes applications, you need to install the CNP
              GitHub App. This allows the platform to create and manage
              repositories on your behalf.
            </p>
            <Button onClick={handleLink} className="w-full">
              <Github className="mr-2 h-4 w-4" />
              Link GitHub Account
            </Button>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
```

**Intégration dans Account Page**:

```tsx
import { GitHubLinkButton } from "@/components/GitHubLinkButton";

export default function AccountPage() {
  const { user } = useAuth();

  return (
    <div className="space-y-6">
      {/* Existing account sections */}

      <GitHubLinkButton userId={user.id} />
    </div>
  );
}
```

**Backend Endpoint Requis**: `GET /api/user/github-status`

```python
# backend/app/routers/user.py
@router.get("/github-status")
async def get_github_status(current_user: User = Depends(get_current_user)):
    """Get user's GitHub installation status from Keycloak."""
    # Fetch from Keycloak user attributes
    github_installation_id = keycloak_service.get_user_attribute(
        current_user.id,
        "github_installation_id"
    )
    return {"github_installation_id": github_installation_id}
```

**Test**:

- [ ] Le bouton "Link GitHub Account" s'affiche si non lié
- [ ] Le clic redirige vers GitHub App installation
- [ ] Le statut "linked" s'affiche après liaison
- [ ] L'installation ID est visible

---

### 3. Dynamic Deployment Cards ⭐ Priority 2

**Fichier**: `frontend/components/DeploymentCard.tsx`

**Objectif**: Afficher des actions différentes selon le `provider_type`

**Refactoring**:

```tsx
import { ExternalLink, Github, Shield, FileCode } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";

interface DeploymentCardProps {
  deployment: Deployment;
}

export function DeploymentCard({ deployment }: DeploymentCardProps) {
  const isKubernetes = deployment.provider_type === "kubernetes";

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle>{deployment.name}</CardTitle>
          <Badge variant={getStatusVariant(deployment.status)}>
            {deployment.status}
          </Badge>
        </div>
        <p className="text-sm text-muted-foreground">
          {deployment.step_message}
        </p>
      </CardHeader>

      <CardContent>
        <div className="space-y-2">
          {isKubernetes ? (
            <KubernetesActions deployment={deployment} />
          ) : (
            <LegacyActions deployment={deployment} />
          )}
        </div>
      </CardContent>
    </Card>
  );
}

function KubernetesActions({ deployment }: { deployment: Deployment }) {
  return (
    <>
      <Button variant="outline" className="w-full justify-start" asChild>
        <a href={deployment.github_repo_url} target="_blank" rel="noopener">
          <Github className="mr-2 h-4 w-4" />
          Open GitHub Repository
          <ExternalLink className="ml-auto h-4 w-4" />
        </a>
      </Button>

      <Button variant="outline" className="w-full justify-start" asChild>
        <a
          href={`https://argocd.3istor.com/applications/${deployment.argocd_app_name}`}
          target="_blank"
          rel="noopener"
        >
          <FileCode className="mr-2 h-4 w-4" />
          View in ArgoCD
          <ExternalLink className="ml-auto h-4 w-4" />
        </a>
      </Button>

      <Button variant="outline" className="w-full justify-start" asChild>
        <a
          href={`https://vault.3istor.com/ui/vault/secrets/kvv2/show/projects/${deployment.project_id}/${deployment.name}`}
          target="_blank"
          rel="noopener"
        >
          <Shield className="mr-2 h-4 w-4" />
          Manage Secrets
          <ExternalLink className="ml-auto h-4 w-4" />
        </a>
      </Button>

      <div className="pt-2 text-xs text-muted-foreground">
        <p>Namespace: {deployment.k8s_namespace}</p>
        <p>Project: {deployment.project_id}</p>
      </div>
    </>
  );
}

function LegacyActions({ deployment }: { deployment: Deployment }) {
  const outputs = JSON.parse(deployment.terraform_outputs || "{}");
  const awsData = outputs.aws || {};

  return (
    <>
      {awsData.alb_dns && (
        <Button variant="outline" className="w-full justify-start" asChild>
          <a href={`https://${awsData.alb_dns}`} target="_blank" rel="noopener">
            <ExternalLink className="mr-2 h-4 w-4" />
            Open Application
          </a>
        </Button>
      )}

      <div className="pt-2 text-xs text-muted-foreground">
        <p>Type: Legacy Hybrid (OpenStack + AWS)</p>
      </div>
    </>
  );
}
```

**Test**:

- [ ] Les déploiements Kubernetes affichent les 3 boutons (GitHub, ArgoCD, Vault)
- [ ] Les déploiements legacy affichent le bouton "Open Application"
- [ ] Les liens externes s'ouvrent dans un nouvel onglet
- [ ] Le namespace et project s'affichent pour Kubernetes

---

### 4. Create Deployment Form ⭐ Priority 2

**Fichier**: `frontend/components/CreateDeploymentModal.tsx`

**Objectif**: Adapter le formulaire selon le template sélectionné

**Modifications**:

```tsx
"use client";

import { useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";

interface CreateDeploymentModalProps {
  template: Template;
  isOpen: boolean;
  onClose: () => void;
}

export function CreateDeploymentModal({
  template,
  isOpen,
  onClose,
}: CreateDeploymentModalProps) {
  const [formData, setFormData] = useState({
    name: "",
    project_name: "",
    replica_count: 2,
    sso_protected: false,
  });

  const isKubernetes = template.category === "paas";

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    const payload = {
      name: formData.name,
      template_id: template.id,
      provider_type: isKubernetes ? "kubernetes" : "legacy_hybrid",
      app_config: isKubernetes
        ? {
            project_name: formData.project_name,
            github_installation_id: user.github_installation_id, // From context
            replica_count: formData.replica_count,
            sso_protected: formData.sso_protected,
          }
        : {
            // Legacy config
          },
    };

    await fetch("/api/deployments", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });

    onClose();
  };

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Deploy {template.name}</DialogTitle>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <Label htmlFor="name">Application Name</Label>
            <Input
              id="name"
              value={formData.name}
              onChange={(e) =>
                setFormData({ ...formData, name: e.target.value })
              }
              placeholder="my-app"
              required
            />
          </div>

          {isKubernetes && (
            <>
              <div>
                <Label htmlFor="project">Project Name</Label>
                <Input
                  id="project"
                  value={formData.project_name}
                  onChange={(e) =>
                    setFormData({ ...formData, project_name: e.target.value })
                  }
                  placeholder="project-alpha"
                  required
                />
                <p className="text-xs text-muted-foreground mt-1">
                  Used for namespace isolation and RBAC
                </p>
              </div>

              <div>
                <Label htmlFor="replicas">Replica Count</Label>
                <Input
                  id="replicas"
                  type="number"
                  min={1}
                  max={10}
                  value={formData.replica_count}
                  onChange={(e) =>
                    setFormData({
                      ...formData,
                      replica_count: parseInt(e.target.value),
                    })
                  }
                />
              </div>

              <div className="flex items-center space-x-2">
                <Checkbox
                  id="sso"
                  checked={formData.sso_protected}
                  onCheckedChange={(checked) =>
                    setFormData({ ...formData, sso_protected: !!checked })
                  }
                />
                <Label htmlFor="sso" className="cursor-pointer">
                  Enable SSO Protection (Keycloak via Envoy Gateway)
                </Label>
              </div>
            </>
          )}

          <div className="flex justify-end gap-2">
            <Button type="button" variant="outline" onClick={onClose}>
              Cancel
            </Button>
            <Button type="submit">Deploy</Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
```

**Test**:

- [ ] Les champs Kubernetes apparaissent pour les templates PaaS
- [ ] Les champs legacy apparaissent pour les templates IaaS
- [ ] La validation fonctionne (champs requis)
- [ ] Le déploiement se crée avec le bon `provider_type`

---

## API Integration Examples

### Fetch Deployments with Filtering

```typescript
// Fetch only Kubernetes deployments
const response = await fetch("/api/deployments?provider_type=kubernetes");
const data = await response.json();

// Fetch deployments for a specific project
const response = await fetch("/api/deployments?project_id=project-alpha");
const data = await response.json();
```

### Create Kubernetes Deployment

```typescript
const deployment = await fetch("/api/deployments", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({
    name: "my-app",
    template_id: "kubernetes-fastapi",
    provider_type: "kubernetes",
    app_config: {
      project_name: "project-alpha",
      github_installation_id: user.github_installation_id,
      replica_count: 3,
      sso_protected: true,
    },
  }),
});
```

### Poll Deployment Status

```typescript
const pollStatus = async (deploymentId: number) => {
  const interval = setInterval(async () => {
    const response = await fetch(`/api/deployments/${deploymentId}`);
    const deployment = await response.json();

    if (deployment.status === "running" || deployment.status === "failed") {
      clearInterval(interval);
    }

    // Update UI
    setDeployment(deployment);
  }, 3000); // Poll every 3 seconds
};
```

---

## Testing Checklist

### Unit Tests

- [ ] `DeploymentCard` renders correctly for Kubernetes
- [ ] `DeploymentCard` renders correctly for Legacy
- [ ] `GitHubLinkButton` shows correct state
- [ ] `CreateDeploymentModal` validates inputs

### Integration Tests

- [ ] Can create Kubernetes deployment
- [ ] Can create Legacy deployment
- [ ] Deployment status updates correctly
- [ ] GitHub linking flow works end-to-end

### E2E Tests

- [ ] Full deployment flow (Catalog → Create → Monitor → Delete)
- [ ] GitHub account linking
- [ ] Switching between IaaS and PaaS tabs

---

## Documentation References

- **API Spec**: `.kiro/steering/docs/05-backend-api/01-deployment-api.md`
- **Changes Summary**: `.kiro/steering/docs/06-phase3-changes.md`
- **Backend Guide**: `backend/PHASE3_COMPLETE.md`

---

## Support

Questions? Check:

1. API documentation (`.kiro/steering/docs/05-backend-api/`)
2. Backend implementation (`backend/PHASE3_*.md`)
3. Architecture docs (`.kiro/steering/docs/`)
