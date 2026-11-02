# Décision de placement cloud

Prise le 2026-11-02 par Hugo (product owner), ticket CNP-131. La numérotation continue
celle de [D-14](../08-finops/01-decisions.md).

---

## D-15 — Placement en cloud public

**Contexte.** À la création d'un projet, le portail propose trois cartes : On-premise, AWS,
GCP (`CMP/frontend/src/components/projects/CreateProjectModal.tsx`). CNP-83 les remplace par
deux choix. La vision produit garde les deux clouds publics allumés en permanence : la
plateforme doit être une cible interne crédible face au shadow IT, et ses chiffres de coût
servent aux décideurs.

**Décision.**

1. Le développeur choisit **on-prem** ou **cloud public**, jamais AWS ou GCP.
2. En cloud public, la plateforme choisit AWS (EKS) ou GCP (GKE) :
    1. seul un cloud `up` est éligible. L'état vient de `fresh_state`
       (`CMP/backend/app/services/clusters/state.py`) ; `starting`, `down` et `unknown`
       excluent le cloud ;
    2. entre deux clouds `up`, le moins cher en coût total moyen, d'après les données
       OpenCost (CNP-89). Le nombre d'apps déjà déployées sur un cloud ne compte pas ;
    3. tant que ce coût moyen par cloud n'existe pas, les nouveaux projets alternent entre
       AWS et GCP. `GET /api/finops/estimates` ne suffit pas : c'est le coût mensuel estimé
       d'une app par cluster, compute seul.
3. Le choix n'est pas figé. Un projet peut changer de cloud plus tard :
    - CNP-132 : bascule en un clic par le owner ou un admin du projet ;
    - CNP-133 : migration de masse par un admin plateforme ;
    - CNP-187 : bascule automatique vers le cloud le moins cher, sur opt-in.
4. L'écran affiche la [comparaison](#comparaison-affichee) ci-dessous.

**Alternatives écartées.**

- Laisser le développeur choisir AWS ou GCP, comme aujourd'hui. Le critère qui les
  départage est le coût, une donnée que la plateforme a et pas le développeur.
- Répartir selon le nombre d'apps déjà placées sur chaque cloud. Ne dit rien du coût.

**Conséquences.**

- Aucun cloud `up` : aucun cloud éligible.
- La procédure opérateur demande `down` après usage
  ([procédures EKS/GKE](07-procedures-operateur.md#cout)). Un cloud éteint n'est jamais
  choisi : la règle 2 suppose les deux clouds allumés, comme le veut la vision produit.
- D-01 tient : `target_cloud` reçoit le cloud choisi par la plateforme (`aws` ou `gcp`).
- D-15 amende trois textes signés, à partir de la livraison de CNP-132 : « `target_cloud`
  est immuable après création en v1 » ([D-01](02-decisions.md#d-01-ou-vit-le-choix-du-cloud)),
  le non-goal « Déplacer un projet existant d'un cloud vers un autre »
  ([invariants](01-invariants-et-perimetre.md)) et « immuable ensuite »
  ([vocabulaire](00-index.md#vocabulaire)). D'ici là, changer de cloud reste recréer le
  projet, comme le dit le portail (« This cannot be changed later »).
- I-5 tient : une bascule déplace tout le projet, prod et staging ensemble.

---

## Comparaison affichée

Ce que disent les cartes, et d'où vient chaque affirmation. Aucun chiffre : les dépôts n'ont
ni SLA, ni mesure de latence.

| Axe | On-prem | Cloud public | Source |
|---|---|---|---|
| Coût | Moins cher. Prix du TCO versionné dans Git. | Plus cher. Montant de la facture du fournisseur. | Décision Hugo ; D-F3, D-F4 ([D-14](../08-finops/01-decisions.md)) |
| Disponibilité, résilience | Un seul cluster, qui porte aussi le plan de contrôle. Une panne on-prem arrête l'app. | Continue de servir avec on-prem injoignable. Kubernetes managé. Aujourd'hui un nœud dans une zone par cloud. | I-2 ; `Cloud/terraform/public/aws/EKS/main.tf`, `Cloud/terraform/public/gcp/GKE/main.tf` |
| Souveraineté | Données sur nos serveurs. | Données chez AWS ou Google, région Paris (`eu-west-3`, `europe-west9`). | Décision Hugo ; `EKS/versions.tf`, `GKE/main.tf` |
| Élasticité | Capacité limitée à nos serveurs. | Ajouter des nœuds ne demande pas de matériel. Pas d'autoscaler de nœuds : taille fixée à 1. | `EKS/main.tf` (`min_size`, `max_size`), `GKE/main.tf` (`node_count`), D-14 |
| Sauvegardes | Bases CNPG sur Ceph RGW : quotidienne et WAL continu, restauration à un instant donné depuis le CMP. Opt-in. | Pas disponibles sur EKS/GKE. | [Sauvegarde des bases](../01-architecture/08-database-backups.md) |
| Latence | Pas de mesure. Le trafic entre par Cloudflare puis le connecteur du projet, sur les deux options. | Idem. | I-1 |

La latence n'apparaît pas dans les cartes : rien ne la départage dans les dépôts.

### Cartes du portail

Le portail reste épuré (décision Hugo, 2026-11-02) : chaque carte affiche un nom et une ligne,
en anglais. Le comparatif ci-dessus vit ici, pas dans l'interface.

| Carte | Ligne |
|---|---|
| On-prem | Cheapest. Data stays on our servers. |
| Public cloud | Keeps running if on-prem is down. |
