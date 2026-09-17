# Contrat d'interface avec l'équipe cluster

Monter et exploiter les clusters AWS/GCP, leurs VPC, l'IAM cloud et les tunnels VPN sont
hors du périmètre CNP. Ces clusters sont consommés comme des cibles déjà connectées.
Ce document existe parce que « déjà connectées » recouvre une série d'hypothèses qui,
laissées implicites, se découvrent en phase 3 au lieu d'être convenues en phase 0.

## Partage des responsabilités

| Objet | Équipe cluster | CNP |
| --- | --- | --- |
| Cluster Kubernetes, nœuds, mises à jour | ✅ | — |
| VPC, sous-réseaux, IAM cloud, quotas | ✅ | — |
| Tunnel VPN vers on-prem | ✅ | — |
| CNI avec support des `NetworkPolicy` | ✅ | — |
| Contrôleur Envoy Gateway et `GatewayClass` | ✅ | — |
| Objet `Gateway` d'un projet, `HTTPRoute`, `ClientTrafficPolicy` | — | ✅ |
| Namespaces de projet, `NetworkPolicy` de projet | — | ✅ |
| Connecteur Cloudflare par projet | — | ✅ |
| Keycloak IAMaaS par projet | — | ✅ |
| Opérateur de secrets, `vmagent`, `blackbox_exporter`, `offhours-guard` | — | ✅ (profil de cluster) |
| Enregistrement du cluster dans Argo CD | — | ✅ |

## Exigences côté cluster

Ce que CNP a besoin de trouver sur un cluster pour pouvoir le consommer.

1. **Sortie internet** depuis les pods, vers Cloudflare (connecteurs de tunnel) et vers
   `ghcr.io` (images). Aucune IP publique entrante, aucun load balancer cloud n'est
   demandé — l'ingress se fait exclusivement par des connexions sortantes (I-4).
2. **`GatewayClass` installée et fonctionnelle**, avec le contrôleur Envoy Gateway. CNP
   crée et détruit les objets `Gateway` au niveau projet ; il ne touche pas à la classe.
3. **CNI supportant les `NetworkPolicy`** en *default-deny*. Cilium on-prem ; l'équivalent
   fonctionnel est requis sur les clusters cloud (D-12).
4. **Classe de stockage bloc** disponible et nommée, communiquée à CNP pour alimenter le
   profil du cluster (D-08). `local-path` n'existe pas ailleurs qu'on-prem.
5. **API server joignable depuis on-prem par le VPN**, pour Argo CD.
6. **Endpoint `TokenReview` joignable depuis Vault par le VPN**, pour l'authentification
   Kubernetes par cluster retenue en D-07.
7. **CA du cluster et JWT de revue** fournis à CNP pour configurer le mount
   d'authentification Vault de ce cluster.

## Flux autorisés sur le VPN

Le VPN ne porte que du plan de contrôle. Tout autre flux qui le traverse est un bug de
conception, pas une commodité.

| Source | Destination | Protocole / port | Raison |
| --- | --- | --- | --- |
| Argo CD (on-prem) | API server du cluster cible | HTTPS 6443 | Synchronisation GitOps (D-03) |
| Opérateur de secrets (cluster cible) | Vault (on-prem) | HTTPS 8200 | Récupération des secrets du projet |
| Vault (on-prem) | API server du cluster cible | HTTPS 6443 | `TokenReview` au login (D-07) |
| `vmagent` (cluster cible) | VictoriaMetrics (on-prem) | HTTPS | `remote_write`, tamponné sur disque (D-10) |

**Explicitement absent de cette liste :** le trafic applicatif entrant. Il va de
Cloudflare Edge au connecteur du projet, dans le cluster du projet, et ne traverse jamais
le VPN (I-1). Si un jour du trafic utilisateur apparaît sur ce lien, c'est une
régression sur l'invariant central du chantier.

## Ce que CNP s'engage à ne pas faire

- Ne pas demander d'IP publique ni de load balancer cloud pour l'ingress.
- Ne pas utiliser de service propriétaire d'un fournisseur sur le chemin applicatif :
  compute, L4/L7, stockage bloc et Postgres générique uniquement (I-3).
- Ne pas exiger de credentials Terraform permanents sur l'API server des clusters
  distants : depuis D-02, seul Argo CD détient des credentials sur les clusters.
- Ne rien laisser derrière lui à la suppression d'un projet : tout objet créé porte
  `cnp.project`, `cnp.cloud` et `cnp.env`, et le job de nettoyage s'appuie dessus (D-13).

## Points à confirmer avec l'équipe cluster avant la phase 3

- [ ] Quel cloud est le pilote, et un cluster est-il réellement consommable aujourd'hui ?
- [ ] Nom de la classe de stockage bloc sur chaque cluster cible.
- [ ] Confirmation du support `NetworkPolicy` en default-deny sur les clusters cloud.
- [ ] Qui, nommément, porte ce contrat côté équipe cluster.
