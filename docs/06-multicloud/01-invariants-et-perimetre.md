# Invariants, périmètre et définition de « terminé »

Posé en amont pour que la revue d'équipe discute des décisions plutôt que des frontières,
et pour que l'implémentation ait une barre d'acceptation non ambiguë.

## Invariants — rien ne peut les violer

| ID | Invariant |
| --- | --- |
| **I-1** | Le trafic applicatif entrant ne transite jamais par on-prem. De Cloudflare Edge vers le connecteur du projet, dans le cluster du projet. |
| **I-2** | Une application déjà déployée continue de servir, d'authentifier et de rafraîchir des jetons avec on-prem injoignable. |
| **I-3** | Aucun service propriétaire d'un cloud sur le chemin applicatif. Compute, L4/L7, stockage bloc, Postgres générique uniquement. |
| **I-4** | Aucune IP publique ni load balancer cloud provisionné pour l'ingress. Connecteurs en sortie seulement. |
| **I-5** | Un projet vit sur exactement un cloud. Prod et Staging ne sont jamais séparés entre deux fournisseurs. |
| **I-6** | Toute ressource créée pour un projet est destructible par l'automatisation qui l'a créée, sans résidu. |

## Non-goals — explicitement hors périmètre

- Monter et exploiter les clusters AWS/GCP, leurs VPC, l'IAM cloud et les tunnels VPN.
  Ils sont consommés comme des cibles déjà connectées.
- Déployer ou provisionner pendant une panne on-prem. Accepté comme indisponible.
- Déplacer un projet existant d'un cloud vers un autre. Hors périmètre v1 : on recrée
  et on migre les données.
- Actif/actif ou bascule automatique entre clouds. Chaque projet vit à un seul endroit.
- La survie des tableaux de bord et du portail pendant une panne on-prem. Accepté comme
  indisponible.
- Remplacer Cloudflare comme edge, ou le modèle de livraison centré sur GitHub.

## Définition de « terminé »

Pas « le code est mergé » : les affirmations suivantes doivent être démontrables sur un
projet pilote vivant.

| # | Affirmation démontrable | Comment on la prouve |
| --- | --- | --- |
| 1 | Un projet créé depuis le portail avec un cloud choisi atterrit intégralement sur ce cloud. | Création via le portail, puis inventaire de chaque objet par fournisseur. |
| 2 | Son application est joignable et ses utilisateurs peuvent se connecter et rafraîchir leurs jetons, VPN on-prem coupé. | Game day GD-1, avec une session navigateur neuve ouverte pendant la coupure. |
| 3 | Le Staging continue de s'éteindre hors heures ouvrées pendant cette coupure. | GD-1, fenêtre planifiée observée on-prem éteint. |
| 4 | La panne d'un cloud est visible depuis les deux autres en quelques minutes. | Maillage de sondes blackbox, vérifié en arrêtant un cluster. |
| 5 | Supprimer le projet ne laisse aucun résidu chez aucun fournisseur. | Teardown, puis le réconciliateur d'orphelins rapporte « propre ». |
| 6 | Perdre le backend d'état AWS ne fait pas perdre le contrôle du parc. | Exercice de restauration depuis la copie on-prem vers un backend jetable. |
| 7 | Les projets on-prem existants ne sont affectés à aucun moment. | Passe de non-régression sur un projet préexistant après chaque phase. |

## Écarts connus entre la note d'architecture v4 et le code

La note v4 décrit plusieurs choses comme déjà vraies que les dépôts ne supportent pas
encore. Ces écarts ne remettent pas la cible en cause, mais ils changent la taille du
travail — et le premier casse la promesse centale de disponibilité telle que le code
est écrit aujourd'hui.

| ID | Écart | Conséquence |
| --- | --- | --- |
| **G-1** | L'authentification des utilisateurs finaux dépend d'on-prem : ce qui existe est un *realm* par projet sur le Keycloak central, et `securitypolicy.yaml` code en dur `issuer: https://auth.3istor.com/realms/<realm>`. | Couper on-prem coupe aujourd'hui la connexion de toutes les applications protégées par SSO, sur tous les clouds. C'est WS-5, et c'est le chemin critique. |
| **G-2** | Prod/Staging n'existe pas comme dimension : aucun champ `environment` nulle part, un seul dépôt, un seul namespace, un seul hostname par application. | Reporté avec la couture posée — voir D-11. |
| **G-3** | Trois régimes de tunnel coexistent, aucun n'est « un par projet ». | Consolidé — voir D-06. |
| **G-4** | Terraform écrit des objets Kubernetes directement, vers un cluster non nommé (`provider "kubernetes" {}`). | Le pivot structurel du chantier — voir D-02. |
| **G-5** | La couche données n'est pas portable : `storageClass: local-path`, `instances: 1`, aucun backup. | Voir D-08. |
| **G-6** | La livraison de secrets est liée à un seul cluster, et l'opérateur réellement utilisé est celui de Ricoberger, pas VSO. | Voir D-07. |
| **G-7** | Il n'existe aucun endroit où stocker le choix du cloud : `ProjectCreate` n'accepte que `project_name`. | Voir D-01. |
| **G-8** | Le verrouillage de l'état Terraform est optionnel et silencieusement ignoré. | Voir D-09. |
| **G-9** | Le chemin d'attachement au Gateway porte un hash généré comme littéral (`...-shared-gateway-ac1e5388...`). | Voir D-05. |
| **G-10** | Un seul jeu de credentials, un seul de chaque chose, dans `core/config.py`. | Voir D-09 et WS-2. |
