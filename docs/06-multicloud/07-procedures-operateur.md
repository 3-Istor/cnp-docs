# Procédures opérateur EKS / GKE

L'opérateur lance les racines Terraform `EKS/` et `GKE/` (dépôt Cloud) avec `cnp.sh`. Les boutons up/down du CMP ne sont pas construits (CNP-109). Vue d'ensemble : [schémas](06-schemas.md).

## Prérequis

- `terraform`, `kubectl`, `vault` ; GKE : `gcloud` + `gke-gcloud-auth-plugin`.
- `export AWS_PROFILE=cnp` (EKS et GKE : GKE écrit l'endpoint dans S3).
- GKE : `gcloud auth application-default login`.
- Vault, pour `plan`, `up` **et** `down` :

```bash
export VAULT_ADDR=https://vault.3istor.com   # écraser tout VAULT_ADDR du shell qui pointe ailleurs
vault login
```

## Mise en place (une fois)

1. Clé WireGuard : `vault kv put kvv2/cnp-vpn/<cloud> privatekey=...`
2. Mount d'auth `kubernetes-<cloud>` (K3s/terraform) : `terraform apply -var-file=secrets_variables.tfvars -target='vault_auth_backend.kubernetes_cloud' -target='vault_kubernetes_auth_backend_role.vault_secrets_operator_role_cloud'` (4 ressources ; `prevent_destroy`).
3. `terraform.tfvars` dans `EKS/` ou `GKE/` : `api_allowed_cidrs` (IP de sortie du K3s + IP de l'opérateur), `vpn_source_cidrs`, `admin_principal_arn` (EKS).

## Commandes

Depuis `EKS/` ou `GKE/` :

| Commande | Effet |
|---|---|
| `./cnp.sh init` | `terraform init` (backend S3 / GCS) |
| `./cnp.sh plan` | plan, sans modification |
| `./cnp.sh up` | apply, kubeconfig, attend le nœud Ready |
| `./cnp.sh down` | refuse s'il reste un LoadBalancer, puis destroy |
| `./cnp.sh kubeconfig` | régénère le kubeconfig |
| `./cnp.sh endpoint` | affiche `<IP du nœud>:31820` et la clé publique WireGuard (diagnostic ; la publication dans S3 est faite par `up`) |

## Automatique après `up`

Rien à synchroniser à la main :

1. Terraform pose le jeton `argocd-manager` dans Vault (`kvv2/argocd/clusters/<cloud>`), la config `auth/kubernetes-<cloud>/config`, la clé WireGuard (ns `cnp-vpn`) et l'objet S3 d'endpoint.
2. Le VSO du hub met à jour le Secret de cluster : moins d'une minute.
3. Argo synchronise `platform-<cloud>` et `wireguard-<cloud>`, puis bootstraps et apps du cloud.
4. Le pod WireGuard on-prem relit l'endpoint : au plus ~60 s après 180 s sans handshake.

## Vérifications

```bash
kubectl get nodes
kubectl -n argocd get app platform-<cloud> wireguard-<cloud>   # Synced / Healthy (contexte K3s)
vault read -field=kubernetes_host auth/kubernetes-<cloud>/config   # IP du cluster neuf
kubectl get svc -A | grep LoadBalancer # doit être vide
```

## `down`

- Survit : entrée de cluster dans Vault, mounts et rôles CMP, config d'auth (réécrite au prochain `up`), `kvv2/cnp-vpn/*`, buckets d'état et d'endpoint, registre, Applications Argo (état Unknown).
- Perdu : cluster, disque du nœud (bases `local-path`), kubeconfig SSM, objet d'endpoint.

## Problèmes connus

| Ticket | Symptôme | Contournement |
|---|---|---|
| CNP-112 | Après un `down` long, l'opération de sync Argo retente l'ancienne adresse | Sur K3s : `kubectl -n argocd annotate app <app> argocd.argoproj.io/refresh=hard --overwrite` ; si l'opération reste bloquée, `kubectl -n argocd delete app <app>` : l'ApplicationSet la recrée (vécu le 18/10 sur `wireguard-gcp`) |
| CNP-114 | nginx du frontend résout le backend au démarrage et plante tant que le Service n'existe pas | Se rétablit seul dès que le backend existe (redémarrages du pod frontend) |
| CNP-115 | Apps fullstack : le VaultSecret `app-registry` est rendu par frontend et backend, l'un reste OutOfSync | Cosmétique, sans effet sur le service |

## Coût

~0,29 $/h avec les deux clusters up. Toujours `down` après usage.
