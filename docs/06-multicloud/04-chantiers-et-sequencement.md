# Chantiers et séquencement

Une branche et une revue par chantier. Chaque chantier nomme les décisions qu'il suppose,
les fichiers qu'il touche, et la barre qu'il doit franchir.

## Principe d'ordonnancement

Prouver l'abstraction sur le cluster qu'on a déjà, avant d'introduire un cluster qu'on
n'a pas. Ajouter un cloud est la partie peu chère ; c'est la dernière, pas la première.

Une porte n'est pas une réunion de statut : c'est une chose précise qui doit être
démontrable avant que la phase suivante commence.

## Les chantiers

| ID | Chantier | Suppose | Barre d'acceptation |
| --- | --- | --- | --- |
| **WS-0** | Garde-fous avant tout mouvement | — | Un changement de chart qui modifie le rendu d'une application existante fait échouer la CI. |
| **WS-1** | Registre de projets et choix du cloud | D-01, D-11 | Un projet créé via le portail produit un enregistrement dont `target_cloud` correspond au choix ; tout projet préexistant se relit en `onprem` sans changement de comportement. |
| **WS-2** | Terraform : contrat agnostique, implémentation par fournisseur | D-02, D-09, D-10 | Déployer une application on-prem par le module refondu produit un ensemble d'objets vivants identique à l'ancien chemin, et le module planifie proprement sans provider Kubernetes dans son fichier de verrouillage. |
| **WS-3** | GitOps multi-cluster | D-01, D-02, D-03 | Déplacer le `target_cloud` d'un projet de test dans le registre déplace ses charges sur l'autre cluster sans exécution Terraform, et une Application pointée manuellement vers le mauvais cluster est rejetée par son `AppProject`. |
| **WS-4** | L'unité d'ingress par projet | D-05, D-06 | Un projet pilote sert par son propre Gateway et son propre connecteur ; supprimer le projet retire les deux, plus ses enregistrements DNS et son tunnel, sans rien laisser chez Cloudflare. |
| **WS-5** | L'identité comme service local *(chemin critique)* | D-04 | VPN on-prem coupé, un utilisateur qui ne s'est jamais authentifié réalise une connexion complète sur une application pilote, et une session existante rafraîchit son jeton. |
| **WS-6** | Secrets entre clusters | D-07 | Un pod qui redémarre sur un cluster distant pendant une panne Vault simulée monte quand même ses secrets, et une lecture de chemin inter-projets est refusée par la politique. |
| **WS-7** | Couche de données portable | D-08 | Une base pilote est restaurée depuis un backup dans un namespace neuf, et le même chart rend correctement sur les trois profils de classe de stockage. |
| **WS-8** | Observabilité et signal de disponibilité | D-10 | Arrêter un cluster produit une alerte issue des deux autres dans la fenêtre convenue, et les métriques du cluster arrêté se rattrapent à la reconnexion sans trou. |
| **WS-9** | Cycle de vie, FinOps et décommissionnement | D-13 | Supprimer un projet pilote laisse le réconciliateur à zéro orphelin chez tous les fournisseurs, et la vue FinOps ventile le coût par cloud. |

## Phases et portes

### Phase 0 — Décisions et sondes

Signature des treize décisions. Trois sondes en parallèle, toutes peu coûteuses :
le comportement de l'opérateur de secrets quand Vault est injoignable (D-07), le Service
stable devant un Gateway par projet (D-05), et les plafonds Cloudflare (D-06). Les
garde-fous WS-0 atterrissent ici.

> **Porte** — chaque décision est signée ou renversée, et aucun chantier ne dépend d'une
> question sans réponse.

### Phase 1 — On-prem devient une cible comme les autres

WS-1, WS-2, WS-3, WS-4 — toute la refonte, avec exactement un cluster en jeu. Rien du
comportement de la plateforme ne doit changer ; tout de sa structure change. C'est la
phase la plus grosse et la plus risquée, et c'est précisément pour cela qu'elle tourne
sans nouveau fournisseur pour brouiller l'attribution des pannes.

> **Porte** — le projet témoin de non-régression est identique dans le cluster, et un
> nouveau projet se provisionne de bout en bout par le nouveau chemin.

### Phase 2 — L'identité devient locale

WS-5 et WS-6, toujours on-prem. Le Keycloak par projet tourne sur le cluster on-prem et
est traité comme s'il était distant : Envoy valide contre l'issuer du projet, les secrets
passent par Vault, et le premier game day de coupure VPN se joue ici. Prouver l'autonomie
sans second cloud est possible, et beaucoup moins cher à déboguer.

> **Porte** — GD-1 passe on-prem : connexion et rafraîchissement fonctionnent avec le
> chemin de plan de contrôle sectionné.

### Phase 3 — Le second cluster

WS-7 et WS-8 atterrissent au moment où le profil de cluster est instancié pour le cloud
pilote. Un projet de test y est créé via le portail, passé dans toute la matrice
d'acceptation, puis GD-1 à nouveau, pour de vrai. Ce n'est qu'ici que le code rencontre
un fournisseur qu'il n'a jamais vu.

> **Porte** — toute la matrice de définition de « terminé » passe sur le cloud pilote.

### Phase 4 — Durcissement et transfert

WS-9, le troisième cloud, la migration des projets existants selon la réponse sur les
realms, les game days restants, `cnp-docs` à jour, runbooks, transfert.

> **Porte** — l'équipe sait créer, exploiter et décommissionner un projet sur n'importe
> quel cloud sans vous dans la pièce.

## Sur les estimations

Délibérément aucune ici. Les deux variables dominantes — le périmètre Prod/Staging
(tranché : reporté, D-11) et la migration des realms existants (toujours sans réponse) —
conditionnent le dimensionnement de WS-5, qui est le chemin critique. Un chiffre produit
avant cette réponse serait faux d'un facteur, pas d'un pourcentage.

## Registre des risques

| # | Risque | Gravité | Atténuation |
| --- | --- | --- | --- |
| R-1 | Le chart partagé est rendu par chaque application vivante ; un changement négligent les casse toutes en même temps. | Critique | Règle de compatibilité WS-0 plus un job de diff de manifeste rendu en CI. Toute nouvelle value vaut par défaut le rendu actuel. |
| R-2 | La reconstruction de l'identité (WS-5) est sous-dimensionnée : le flux d'authentification Terraform existant est substantiel et son portage en config de realm est facile à rater subtilement. | Critique | Porter le flux d'abord, et diffuser l'export de realm contre l'actuel avant toute autre chose dans WS-5. Un écart de comportement est bloquant. |
| R-3 | Désaccord d'issuer OIDC entre la vue du navigateur et celle d'Envoy — l'échec classique du Keycloak par projet, qui se manifeste à la validation du jeton et non au déploiement. | Critique | Épingler le hostname public comme issuer partout ; test explicite dans la matrice d'acceptation du pilote, pas un simple test de fumée. |
| R-4 | La migration d'état vers la nouvelle disposition de clés corrompt ou perd l'état d'un projet existant. | Critique | Versioning vérifié d'abord, copie on-prem en place avant de commencer, mode simulation, un projet à la fois avec vérification entre chaque. |
| R-5 | L'opérateur de secrets réconcilie un Secret synchronisé quand Vault est injoignable, cassant I-2 exactement dans le scénario pour lequel le design existe. | Élevée | Testé explicitement dans GD-1. En cas d'échec, la migration d'opérateur devient obligatoire et passe sur le chemin critique. |
| R-6 | Argo CD devient un point de défaillance unique pour toute la livraison sur trois clouds. | Élevée | Accepté par le contrat — la livraison peut s'arrêter. Adossé à la sauvegarde des secrets de cluster et de la configuration, avec une reconstruction documentée. |
| R-7 | Le coût du Keycloak par projet croît linéairement et ne se remarque qu'une fois cher. | Élevée | Base mutualisée dès le premier jour, requests de ressources explicites, et alerte budgétaire par cloud dans WS-9 avant que le second cloud reçoive des projets. |
| R-8 | Le provisioning devient *eventually consistent* et le portail continue d'annoncer un succès pour des choses qui échoueront ensuite à se synchroniser. | Moyenne | Statut dérivé de la santé Argo CD dans WS-3, avec un état « soumis » distinct de « en fonctionnement ». |
| R-9 | Les dépendances à l'équipe cloud — `GatewayClass`, sortie réseau, support `NetworkPolicy` — arrivent tard ou différemment de l'hypothèse. | Moyenne | Contrat d'interface écrit, convenu avant la phase 3 plutôt que découvert pendant. |
| R-10 | Les plafonds d'objets Cloudflare sont atteints à l'échelle, après que le design par projet a été engagé. | Moyenne | Non bloquant au stade POC (D-06). Repli : un tunnel partagé par cluster avec des routes par projet. |

## Game days

Un contrat de disponibilité qui n'a jamais été exercé est une hypothèse.

### GD-1 — la coupure VPN

Joué deux fois : une fois en phase 2 on-prem, une fois en phase 3 sur le cloud pilote pour
de vrai. On sectionne le chemin de plan de contrôle, puis on déroule le tableau. Ce qui
n'est pas vert est un constat, pas une nuance.

| Vérification | Attendu | Rattaché à |
| --- | --- | --- |
| Une session existante charge l'application. | Fonctionne | I-1 |
| Un utilisateur qui ne s'est jamais connecté réalise une connexion complète. | Fonctionne | I-2, WS-5 |
| Un jeton d'accès expire et se rafraîchit. | Fonctionne | I-2, WS-5 |
| Un pod est supprimé et replanifié ; il monte ses secrets. | Fonctionne | R-5, WS-6 |
| La base continue de servir ; une fenêtre de backup passe. | Sert ; le backup peut échouer bruyamment | WS-7 |
| Le Staging s'éteint à l'heure planifiée. | Fonctionne | WS-9 |
| Gatus détecte une panne applicative induite et poste sur Discord. | Fonctionne | D-10 |
| `vmagent` tamponne et rattrape à la reconnexion. | Aucun trou après reprise | WS-8 |
| Un déploiement est tenté. | Échoue avec une erreur claire et honnête | D-02, R-8 |
| Le portail est ouvert. | Indisponible — attendu | non-goal |

### GD-2 — perte d'un cluster

Arrêter entièrement le cluster pilote. Les deux autres doivent le signaler dans la fenêtre
convenue, et rien sur eux ne doit se dégrader. Prouve que le maillage de sondes est
réellement indépendant.

### GD-3 — décommissionnement

Faire passer un projet jetable du cloud pilote par tout le runbook de décommissionnement.
Le succès, c'est le réconciliateur d'orphelins à zéro chez tous les fournisseurs ensuite.

### GD-4 — restauration d'état

Reconstruire un backend Terraform fonctionnel depuis la copie on-prem vers un bucket
jetable et planifier contre lui sans diff. Prouve que la seconde copie est une copie et
pas un réconfort.
