# Nettoyage et décommissionnement

Mise en œuvre de [D-13](02-decisions.md#d-13-nettoyage-et-decommissionnement).
I-6 fait du nettoyage une partie de « terminé », pas un travail de suite : toute ressource
créée pour un projet est destructible par l'automatisation qui l'a créée, sans résidu.

## L'outil

`app-templates/scripts/cnp-clean.sh`, successeur scopé de `cleanup-cnp-demo.sh`.

```bash
# Ce qui serait détruit pour un projet — ne détruit rien
./scripts/cnp-clean.sh --project sandbox

# Détruire réellement
./scripts/cnp-clean.sh --project sandbox --yes

# Retirer un cloud entier une fois les tests terminés
./scripts/cnp-clean.sh --cloud aws --yes

# Inventaire seul, tous clouds — sert aussi de rapport d'orphelins
./scripts/cnp-clean.sh --all --inventory-only
```

| Drapeau | Effet |
| --- | --- |
| `--project <nom>` | Portée : un projet et toutes ses applications |
| `--cloud <onprem\|aws\|gcp>` | Portée : tous les projets dont `target_cloud` vaut ce cloud |
| `--all` | Portée : les trois clouds |
| `--yes` | Détruit réellement. Sans lui, l'outil n'affiche que l'inventaire |
| `--inventory-only` | S'arrête après l'inventaire, même avec `--yes` |
| `--keep-registry` | Détruit les ressources mais laisse l'enregistrement du projet dans le registre Git |

Il n'existe **aucune invocation sans portée** : oublier les arguments ne détruit rien.

## Ce qui est détruit, et dans quel ordre

L'ordre importe. Détruire le bootstrap d'un projet avant ses applications laisse des
ressources applicatives sans état pour les retrouver.

1. **Applications du projet** — un `terraform destroy` par état
   `cmp/<cloud>/projects/<projet>/apps/<app>.tfstate` : dépôt GitHub, client OIDC,
   secrets Vault de l'application, tunnel et DNS Cloudflare de l'application, paquets
   GHCR.
2. **Manifestes GitOps du projet** — retrait de l'entrée du registre, ce qui fait
   disparaître les `Application` et l'`AppProject` générés par l'`ApplicationSet`. Argo CD
   les supprime avec `prune`, et les finaliseurs emportent les objets du cluster.
3. **Bootstrap du projet** — `terraform destroy` sur
   `cmp/<cloud>/projects/<projet>/bootstrap.tfstate` : groupes et realm Keycloak, mount et
   politiques Vault, rôles d'authentification Kubernetes du projet, DNS `status-` et
   `offhours-`, tunnel du projet.
4. **Enregistrement CMP** — la ligne `projects` correspondante, pour que le projet
   disparaisse du portail.
5. **Inventaire final** — relecture de chaque fournisseur avec les marqueurs
   `cnp.project` / `cnp.cloud`. L'état de fin attendu est zéro objet retrouvé.

Pour `--cloud`, la boucle applique cette séquence à chaque projet du cloud, puis rapporte
les objets de niveau cluster restants (profil de cluster, secret d'enregistrement Argo CD,
mount d'authentification Vault de ce cluster) — ces derniers sont listés mais **pas**
supprimés automatiquement : ils appartiennent au périmètre de l'équipe cluster.

## Le marquage rend tout cela possible

Un suppresseur qui devine est un suppresseur qui détruit à côté. Chaque objet créé pour un
projet porte, sous la forme native de son fournisseur :

| Fournisseur | Forme du marqueur |
| --- | --- |
| Kubernetes | labels `cnp.3istor.com/project`, `cnp.3istor.com/cloud`, `cnp.3istor.com/env` |
| Cloudflare | le champ `comment` de l'enregistrement DNS et le nom du tunnel |
| Vault | métadonnées du mount |
| Keycloak | convention de nommage `project-<nom>-*` et realm `<nom>` |
| Ressources cloud | tags `cnp.project`, `cnp.cloud`, `cnp.env` |

## Runbook de décommissionnement d'un cloud

À exécuter quand un cloud de test ou de démonstration est retiré pour cesser de payer.

1. Geler les créations vers ce cloud : retirer la valeur de l'énumération `target_cloud`
   acceptée par l'API CMP, pour qu'aucun nouveau projet n'y atterrisse pendant l'opération.
2. `./scripts/cnp-clean.sh --cloud <nom>` — lire l'inventaire, vérifier qu'il correspond
   à ce qu'on croit avoir.
3. `./scripts/cnp-clean.sh --cloud <nom> --yes`.
4. Relancer l'inventaire. Attendu : zéro.
5. Retirer le secret d'enregistrement du cluster dans Argo CD et le mount
   d'authentification Vault de ce cluster.
6. Rendre le cluster à l'équipe cluster, qui démonte le VPC, l'IAM et le VPN.
7. Vérifier la facture du fournisseur au cycle suivant — le seul contrôle qui compte
   vraiment.

## Limite connue

L'outil s'appuie sur les états Terraform et sur le registre pour savoir quoi détruire.
Une ressource créée à la main, hors de ces deux sources, n'est pas trouvée par la phase de
destruction — mais elle l'est par l'inventaire final si elle porte les marqueurs. C'est
la raison d'être du marquage uniforme, et c'est pourquoi l'inventaire se relance **après**
la destruction et pas seulement avant.
