# Stratégie de Test et Mocking du Backend (FastAPI) 🧪

Ce document détaille la stratégie de test mise en œuvre sur la **Cloud Management Platform (CMP)**. Il présente la structure de la suite de tests (unitaires et d'intégration), la configuration globale de `pytest`, ainsi qu'un guide technique complet expliquant comment simuler (mocker) les dépendances externes critiques (SDKs AWS/OpenStack, API GitHub, et exécutions Terraform).

Pour plus de contexte, vous pouvez consulter :
* Les instructions de déploiement et tests locaux : [08-cmp-local-development.md](file:///home/brian/Documents/aepita/ing2/3-Istor/cnp-docs/05-cmp-backend-api/08-cmp-local-development.md)
* Le fonctionnement du provisionneur : [03-terraform-provisioner.md](file:///home/brian/Documents/aepita/ing2/3-Istor/cnp-docs/04-templates/03-terraform-provisioner.md)
* Le cycle de provisioning applicatif : [01-app-provisioning-flow.md](file:///home/brian/Documents/aepita/ing2/3-Istor/cnp-docs/03-pipelines-and-workflows/01-app-provisioning-flow.md)

---

## 🏛️ Structure des Tests Backend

La suite de tests est localisée dans le dossier `backend/tests/` et utilise le framework `pytest`. Elle est divisée en tests unitaires et tests d'intégration.

```text
backend/tests/
├── conftest.py                 # Fixtures pytest globales (Client API, Base de données de test, Mocks de base)
├── unit/                       # Tests unitaires isolés (sans dépendances réseau ou DB réelle)
│   ├── test_saga_states.py     # États de la machine à état SAGA
│   ├── test_jwt_validation.py  # Validation de signature de clés publiques et claims Keycloak
│   └── test_schemas.py         # Validation de schémas de requêtes et réponses Pydantic
└── integration/                # Tests d'intégration (DB SQLite en mémoire, appels API simulés)
    ├── test_deployments_api.py # Endpoints de cycle de vie des déploiements (/api/deployments)
    ├── test_projects_api.py    # Isolation et droits RBAC des projets
    └── test_catalog_api.py     # Lecture du catalogue et parsing des manifests JSON
```

---

## 🛠️ Focus Technique : Mocking des Dépendances Externes

Pour garantir des tests rapides, déterministes et exécutables hors-ligne, toutes les communications réseau et exécutions de sous-processus système sont interceptées et simulées.

### 1. Mocking des SDKs Cloud (Boto3 & OpenStackSDK)

Lors du provisioning IaaS legacy ou du stockage des fichiers `.tfstate` sur AWS S3, le backend communique avec AWS et OpenStack.

#### A. Simulation de Boto3 (AWS S3)
Nous utilisons le module `unittest.mock` (ou la bibliothèque `moto`) pour surcharger les appels du client S3 afin qu'aucune interaction n'ait lieu avec AWS.

```python
# tests/unit/test_s3_storage.py
import pytest
from unittest.mock import MagicMock, patch
from app.services.s3_service import S3StateService

@pytest.fixture
def mock_s3_client():
    with patch("app.services.s3_service.boto3.client") as mock_client:
        s3_mock = MagicMock()
        mock_client.return_value = s3_mock
        yield s3_mock

def test_s3_upload_state(mock_s3_client):
    service = S3StateService(bucket_name="test-bucket")
    
    # Simulation d'un upload réussi
    mock_s3_client.put_object.return_value = {"ResponseMetadata": {"HTTPStatusCode": 200}}
    
    success = service.upload_state(key="cmp/project-1/app.tfstate", content="{}")
    
    assert success is True
    mock_s3_client.put_object.assert_called_once_with(
        Bucket="test-bucket",
        Key="cmp/project-1/app.tfstate",
        Body="{}"
    )
```

#### B. Simulation de l'OpenStack SDK
Pour le provisionnement des machines virtuelles et bases de données OpenStack, les connexions d'API sont simulées via une surcharge du gestionnaire de connexion OpenStack.

```python
# tests/unit/test_openstack_provisioning.py
import pytest
from unittest.mock import MagicMock, patch
from app.services.openstack_service import OpenStackService

@pytest.fixture
def mock_openstack_conn():
    with patch("app.services.openstack_service.openstack.connect") as mock_connect:
        conn_mock = MagicMock()
        mock_connect.return_value = conn_mock
        yield conn_mock

def test_create_virtual_machine(mock_openstack_conn):
    service = OpenStackService()
    
    # Configurer le comportement attendu du mock OpenStack compute
    mock_server = MagicMock()
    mock_server.id = "os-vm-uuid-1234"
    mock_server.status = "ACTIVE"
    mock_openstack_conn.compute.create_server.return_value = mock_server
    
    vm_id = service.provision_instance(name="test-vm", flavor="m1.medium", image="ubuntu-22.04")
    
    assert vm_id == "os-vm-uuid-1234"
    mock_openstack_conn.compute.create_server.assert_called_once()
```

---

### 2. Mocking de l'API GitHub App

Le backend interagit avec GitHub pour générer des JSON Web Tokens (JWT) signés, récupérer des jetons d'accès d'installation de 60 minutes et créer des dépôts GitOps privés.

Pour mocker ce flux sans appeler l'API REST de GitHub :
* Nous interceptons les requêtes du client HTTP (`httpx.AsyncClient` ou `requests`).
* Nous surchargeons le service de signature de clé privée (`jwt.encode`).

```python
# tests/conftest.py
import pytest
from unittest.mock import AsyncMock, patch

@pytest.fixture
def mock_github_app_service():
    with patch("app.services.github_app.GitHubAppService") as MockService:
        instance = MockService.return_value
        
        # Simule l'obtention d'un jeton temporaire d'installation
        instance.get_installation_token = AsyncMock(return_value="ghs_temp_token_abcdef12345")
        
        # Simule la création réussie d'un dépôt d'application privé
        instance.create_repository = AsyncMock(return_value={
            "id": 87654321,
            "name": "project-alpha-my-app",
            "html_url": "https://github.com/3-Istor/project-alpha-my-app.git"
        })
        
        yield instance
```

---

### 3. Mocking des Exécutions Terraform Locales

Pour les déploiements Kubernetes (PaaS), le backend exécute localement des commandes Terraform (`terraform init`, `terraform apply`, `terraform destroy`) via un sous-processus système (en utilisant `subprocess`).

Pour éviter d'exécuter de vrais binaires Terraform qui nécessiteraient des accès aux clusters K3s ou à AWS, nous simulons la classe d'exécution `TerraformExecutor` ou la fonction système `subprocess.run`.

```python
# tests/unit/test_terraform_executor.py
import pytest
from unittest.mock import MagicMock, patch
from app.services.terraform_executor import TerraformExecutor

@pytest.fixture
def mock_subprocess_run():
    with patch("app.services.terraform_executor.subprocess.run") as mock_run:
        result_mock = MagicMock()
        result_mock.returncode = 0
        result_mock.stdout = "[TF] Apply complete! Resources: 3 added, 0 changed, 0 destroyed."
        result_mock.stderr = ""
        mock_run.return_value = result_mock
        yield mock_run

def test_terraform_apply_execution(mock_subprocess_run):
    executor = TerraformExecutor(
        project_id="alpha",
        deployment_name="frontend",
        variables={
            "vault_token": "hvs.test",
            "github_token": "ghs_test"
        }
    )
    
    # Simulation de la récupération des outputs JSON de Terraform
    with patch.object(executor, "get_outputs") as mock_outputs:
        mock_outputs.return_value = {
            "github_repo_url": {"value": "https://github.com/3-Istor/project-alpha-frontend"},
            "vault_secrets_path": {"value": "kvv2/data/projects/alpha/frontend"},
            "kubernetes_namespace": {"value": "project-alpha"}
        }
        
        success = executor.apply()
        
        assert success is True
        # Vérifie que la commande 'subprocess.run' a bien été appelée avec 'terraform' et 'apply'
        assert mock_subprocess_run.called
        args, kwargs = mock_subprocess_run.call_args
        cmd = args[0]
        assert "terraform" in cmd
        assert "apply" in cmd
```

---

## 🚦 Exécution Locale des Tests

Avant de soumettre du code à la branche principale ou lors des pipelines de CI, les tests doivent être exécutés localement :

```bash
cd backend
# Ajoute le chemin courant au PYTHONPATH pour résoudre les imports d'app
export PYTHONPATH=$PYTHONPATH:.
# Lance la suite complète de tests via pytest
poetry run pytest -v
```

*Note : L'option `-v` active le mode verbeux détaillant chaque scénario validé.*

---
**Next Step**: Continue to [GitHub Repositories Landscape](../04-templates/00-github-repositories-landscape.md) (or return to the [Project Overview](../index.md)).
