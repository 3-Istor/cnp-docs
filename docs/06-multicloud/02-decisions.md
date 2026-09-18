# Décisions D-01 à D-13

Signées le 2026-09-15. Chaque décision porte un identifiant stable : les chantiers
(WS-x) nomment les décisions qu'ils supposent, donc annuler une décision ne force à
replanifier que les chantiers qui la citent.

!!! note "Deux décisions renversent la recommandation initiale"
    **D-07** retient l'authentification Kubernetes plutôt que JWT à clés épinglées, et
    **D-13** demande un job de suppression réel plutôt qu'un simple rapporteur
    d'orphelins. Les conséquences de ces deux choix sont détaillées dans leurs fiches.

---

## D-01 — Où vit le choix du cloud

**Décision.** Le registre Git est la source de vérité du placement. Chaque projet
possède un enregistrement `cnp-projects/registry/projects/<nom>.yaml` portant
`target_cloud`, et toute modification passe par un commit — ce qui donne gratuitement
l'historique, la revue et le rollback.

La base CMP conserve une **copie** du champ (colonne `target_cloud` sur la table
`projects`), pour que le portail puisse lister et afficher les projets sans lire Git à
chaque requête. En cas de divergence entre les deux, le fichier Git fait foi.

**Alternative écartée.** Faire de la base CMP la source de vérité et projeter vers Git.
Techniquement équivalent, mais cela place l'autorité dans le composant qui disparaît
pendant une panne on-prem, alors que Git reste lisible.

**Conséquences.**

- `target_cloud` est immuable après création en v1. Le changer, c'est recréer le projet :
  l'API rejette explicitement la modification plutôt que d'orphéliner des ressources.
- Base et Git peuvent diverger. Le réconciliateur de WS-9 vérifie les deux sens.
- La création de projet devient un aller-retour par commit. C'est assumé : c'est
  exactement le comportement déjà en place aujourd'hui pour le manifeste projet.

---

## D-02 — Comment l'état Kubernetes atteint un cluster distant

**Décision.** Terraform cesse complètement d'écrire des objets Kubernetes. Le provider
`kubernetes` est retiré du module applicatif ; chaque objet qu'il créait devient soit un
secret Vault synchronisé, soit un manifeste rendu dans Git.
**Argo CD devient le seul détenteur de credentials sur les clusters distants.**

**Alternative écartée.** Donner au runner Terraform on-prem des credentials permanents à
privilèges élevés sur les trois API servers. Surface de privilège permanente, dépendance
dure au VPN sur le chemin de provisioning, et un apply partiel laisse des objets sur un
cluster distant sans réconciliateur pour converger.

**Conséquences.**

- Le provisioning devient *eventually consistent* : l'API ne peut plus répondre
  « namespace créé » à l'instant où l'apply rend la main. Le statut de déploiement du
  portail se dérive désormais de l'état de synchronisation et de santé Argo CD.
- Gain de second ordre : cinq des six objets déplacés étaient créés à chaque déploiement
  applicatif. Les passer par un générateur supprime toute une classe d'échecs partiels,
  et le contournement `time_sleep` disparaît.

**Ce que Terraform garde.** Cloudflare, GitHub, Vault, les API des fournisseurs cloud,
la configuration d'identité. Tout ce qu'il joint par internet ou par une API unique.

---

## D-03 — Mécanique multi-cluster d'Argo CD

**Décision.**

- Chaque cible est enregistrée comme secret de cluster nommé d'après le cloud
  (`onprem`, `aws`, `gcp`), étiqueté `cnp.3istor.com/cloud`.
- Les Applications adressent `destination.name`, jamais une URL de serveur : l'URL change
  quand un cluster est reconstruit, le nom non.
- La génération passe dans un `ApplicationSet` avec une matrice « générateur de fichiers
  Git sur `registry/projects/*.yaml` » × « liste d'applications du projet ». Ajouter un
  projet est un commit ; ajouter un cloud est un secret de cluster.
- Un `AppProject` par projet, avec les destinations restreintes au cluster et aux
  namespaces de ce projet — une Application mal générée ne peut pas atterrir sur le
  mauvais cloud.

**Conséquences.** Argo CD a besoin de joindre chaque API server distant par le VPN.
C'est l'unique dépendance dure du plan de contrôle, et elle est cohérente avec le
contrat : VPN coupé signifie plus de déploiements, pas plus de service. Argo CD devient
aussi un vrai point de défaillance unique pour la livraison — d'où la sauvegarde
explicite de ses secrets de cluster aux côtés de l'état Terraform (D-09).

---

## D-04 — Keycloak par projet : construction, amorçage, alimentation

**Décision.** Une instance Keycloak par projet, dans `<projet>-system`, sur le cluster du
projet, livrée par `cnp-project-base`.

- **Issuer** — `auth-<projet>.3istor.com`, routé par le tunnel et le Gateway du projet.
  La chaîne d'issuer doit être identique que le jeton soit validé par le navigateur ou
  par Envoy. Le piège classique est qu'Envoy résolve une URL interne puis rejette la
  revendication `iss` : on épingle le hostname public comme issuer partout, et on donne à
  Envoy un chemin de résolution local pour ce même nom.
- **Amorçage du realm** — `keycloak-config-cli` en Job de sync-wave, lisant les
  définitions de realm depuis Git, plutôt que le provider Terraform Keycloak. Le provider
  exigerait qu'on-prem joigne le Keycloak de chaque projet à chaque apply, ce qui
  réintroduirait exactement la dépendance qu'on supprime.
- **Secrets de client** — Terraform génère le secret OIDC de chaque application et
  l'écrit dans Vault ; l'opérateur de secrets le synchronise dans le namespace ;
  `keycloak-config-cli` le consomme pour créer le client avec ce secret exact ; la
  `SecurityPolicy` d'Envoy lit le même Secret Kubernetes. Une valeur, un écrivain,
  pas de problème d'amorçage circulaire.

**Maîtrise du coût.** Un pod Keycloak par projet est la frontière d'isolation qui vaut la
peine d'être payée ; une base Postgres par projet, non. Un cluster CNPG mutualisé par
cluster cible, une base par projet à l'intérieur, et Keycloak réglé sur un petit tas
mémoire. À cent projets, cela fait cent petits pods et un cluster de base de données par
cloud, pas deux cents pods.

**Tranché (2026-09-18).** Négligeable — aucun projet existant n'a de vrais utilisateurs
finaux dans son realm actuel. WS-5 se limite donc aux nouveaux projets ; pas de double
exécution ni de migration de realm à prévoir.

---

## D-05 — Un Gateway par projet, avec un Service stable devant

**Décision déjà actée en v4.** Un Gateway par projet ; l'équipe cluster possède la
`GatewayClass` et le contrôleur. Ce qui reste est la mécanique qui le fait fonctionner.

- Créer un `Service` ClusterIP ordinaire dans `<projet>-system`, sélectionnant les pods
  Envoy par `gateway.envoyproxy.io/owning-gateway-name` et `owning-gateway-namespace`.
  La configuration du tunnel vise alors un nom qu'on maîtrise, et non un hash généré par
  Envoy Gateway (`envoy-gateway-infra-shared-gateway-ac1e5388`, qui différera sur un
  cluster neuf et sur chaque Gateway par projet).
- Remplacer le sélecteur global `prod-gateway-access: "true"` par un label par projet,
  pour qu'une route du projet A ne puisse pas s'attacher au Gateway du projet B.
- `httproute.yaml` cesse de coder `parentRefs` en dur et les prend depuis les values,
  avec par défaut le Gateway partagé actuel, pour que les applications existantes ne
  bougent pas.
- Reporter la `ClientTrafficPolicy` (un saut de confiance, `X-Forwarded-For`) sur chaque
  Gateway par projet : elle n'est attachée qu'au Gateway partagé aujourd'hui, et la
  perdre casse silencieusement l'IP client partout.

---

## D-06 — Un tunnel par projet, configuré depuis Git

**Décision.** Un tunnel Cloudflare et un connecteur par projet, dans `<projet>-system`,
portant tous les hostnames du projet — prod, staging, `auth-`, `status-`, `offhours-`.
La configuration d'ingress passe de `config_src = "cloudflare"` à un fichier local livré
par Argo CD. Terraform continue de créer le tunnel, le credential et les enregistrements
DNS ; la table de routage devient de l'état Git.

**Pourquoi la config locale.** Avec la config distante, ajouter une route à un projet en
production exige l'API Cloudflare, donc le runner on-prem. Avec la config locale, les
changements de route passent par Argo CD comme le reste, et le connecteur les recharge
lui-même. Cela permet aussi de supprimer le contournement `time_sleep` au destroy.

**Limites de compte Cloudflare.** Non bloquant : aucun problème de plafond constaté
jusqu'ici avec un nombre déjà conséquent de projets hébergés, donc le design par projet
est retenu pour le POC. À rechiffrer avant de viser l'échelle « milliers de projets »
évoquée dans la note ; le repli documenté est un tunnel partagé par cluster avec des
routes par projet — isolation moindre, même chemin de trafic.

---

## D-07 — Authentification Vault entre clusters

**Décision — retenue contre la recommandation initiale.** Un *mount* d'authentification
Kubernetes par cluster : `kubernetes-onprem`, `kubernetes-aws`, `kubernetes-gcp`, chacun
avec le CA de son cluster et un JWT de revue. Modèle familier, aligné sur ce qui tourne
déjà, et qui ne demande pas de procédure de rotation de clés nouvelle.

**Alternative écartée.** Authentification JWT avec les clés publiques de chaque cluster
épinglées statiquement, où Vault valide la signature hors ligne sans jamais joindre les
clusters. Elle supprimait une dépendance réseau au lieu d'en ajouter une, mais au prix
d'une procédure de rotation à écrire et à tenir.

**Conséquence à assumer.** Vault doit joindre l'endpoint `TokenReview` de chaque cluster
au moment du login. C'est une seconde dépendance du plan de contrôle, dans le sens
inverse de celle d'Argo CD, et elle doit être inscrite dans la liste des flux autorisés
sur le VPN (voir le contrat d'interface).

**Ce que cela ne casse pas.** Ce login n'a lieu qu'au moment où l'opérateur de secrets
synchronise. Un Secret Kubernetes déjà synchronisé n'en dépend pas : I-2 reste tenable.

**Question de contrat restée ouverte, à tester et non à supposer.** Que fait l'opérateur
Ricoberger d'un Secret déjà synchronisé quand Vault est injoignable ? I-2 exige qu'il le
laisse tranquille. C'est un point de premier rang du game day GD-1. Si la réponse est
mauvaise, le choix de l'opérateur se rouvre — c'est le seul scénario où migrer vers VSO
de HashiCorp devient nécessaire plutôt qu'optionnel.

---

## D-08 — Les bases de données restent auto-hébergées et portables

**Décision.** CNPG partout, aucune fonctionnalité spécifique à un fournisseur. Cela
satisfait I-3 directement, garde un seul modèle opérationnel sur trois fournisseurs, et
évite la tarification des bases managées par cloud.

- La classe de stockage devient une valeur de profil par cluster.
- Le nombre d'instances devient une valeur par environnement (prod ≥ 2, staging 1).
- Le backup vers un stockage objet compatible S3 est activé. `K3s/k8s/app/garage-s3.yaml`
  fournit déjà une cible on-prem opérationnelle.

**Point de vigilance.** Interaction avec `offhours-guard` : éteindre une base staging qui
a un PVC et un planning de backup actif demande un ordonnancement explicite, sinon on
récolte des backups en échec et des alertes bruyantes toutes les nuits. À spécifier une
fois dans le chart plutôt qu'à découvrir projet par projet.

---

## D-09 — Disposition de l'état, verrouillage et seconde copie

**Décision.**

- La disposition des clés devient `cmp/<cloud>/projects/<projet>/bootstrap.tfstate` et
  `cmp/<cloud>/projects/<projet>/apps/<app>.tfstate`. Cela préserve l'isolation actuelle
  tout en faisant d'un inventaire par cloud un simple listage de préfixe.
- Le verrouillage devient non optionnel dans le code : si le backend est activé et
  qu'aucun verrouillage n'est configuré, le runner refuse de démarrer au lieu d'avertir.
- Versioning du bucket vérifié actif, et réplication planifiée à sens unique de tout le
  préfixe vers Garage on-prem. Copie en lecture seule, pas un backend de bascule.
- Les secrets de cluster Argo CD et le matériel de descellement Vault sont sauvegardés à
  la même cadence et au même endroit : l'état seul ne permet pas de reconstruire le
  contrôle.

**Note de migration.** Changer la disposition des clés implique de déplacer les objets
d'état existants. C'est une opération `terraform state` scriptée par projet avec une
passe de vérification, pas un simple renommage : mode simulation, un projet à la fois,
vérification entre chaque.

---

## D-10 — Observabilité et chemin d'alerte qui survit à on-prem

**Décision.**

- Un `vmagent` par cluster, en `remote_write` vers la VictoriaMetrics on-prem par le VPN,
  avec le tampon disque dimensionné explicitement pour la plus longue panne dont on
  accepte de perdre les données. Le chiffre se choisit, il ne s'hérite pas du défaut.
- Un `blackbox_exporter` dans les trois clusters, chacun sondant le hostname public de
  **tous** les projets — pas seulement les siens — par le vrai chemin internet, à travers
  Cloudflare. Chacun est scrapé par le `vmagent` local, avec le même tampon.
  Résultat : si un cloud tombe, les deux autres continuent de rapporter son
  indisponibilité en quasi temps réel.
- La liste des cibles de sonde est générée depuis le même registre que le reste.
- **Le chemin d'alerte pendant une panne on-prem est le Gatus par projet**, qui poste
  déjà directement vers Discord et tourne déjà sur le cluster du projet. C'est la seule
  route de notification qui ne traverse pas un plan de contrôle mort : elle doit être
  explicitée et testée.

Les tableaux de bord restent hébergés on-prem et peuvent être indisponibles pendant la
panne. Accepté : seule la collecte doit continuer.

**Volontairement reporté.** Exposer un endpoint d'ingestion authentifié à travers
Cloudflare pour que les métriques ne dépendent plus du tout du VPN. Vraie amélioration,
non nécessaire au contrat, et elle ajoute un chemin d'écriture exposé sur internet.
À revoir après le pilote.

---

## D-11 — Prod/Staging est-il dans le périmètre de ce chantier

**Décision — reporté, avec la couture posée.** Le multi-environnement n'est pas le
périmètre de ce chantier, mais il ne doit pas être rendu plus coûteux par lui.

Concrètement : `environment` est porté dès maintenant dans le schéma du registre, dans
le nommage des namespaces et dans les values du chart, avec `prod` par défaut pour tout
l'existant, et il sera « allumé » en travail de suite. Le multicloud se prouve contre un
modèle mono-environnement connu et sain.

**Conséquence documentaire.** La note d'architecture v4 décrit Prod/Staging comme
l'anatomie *actuelle* d'un projet. C'est la cible, pas l'état du code (écart G-2). La
note doit être amendée avant d'être diffusée, sinon la revue d'équipe calibrera tout le
reste de travers.

---

## D-12 — Segmentation réseau

**Décision.** **Un VPN par cloud**, qui discute directement avec l'API du cluster ; le
système CNP se branche par-dessus. Pas de tunnel par projet : le trafic applicatif ne
passe de toute façon plus par ce VPN (il passe par le tunnel Cloudflare de chaque
projet), donc multiplier les tunnels ne ferait que multiplier la gestion de clés et de
certificats à l'échelle de milliers de projets.

Le vrai risque à traiter est le mouvement latéral à l'intérieur d'un cloud si un projet
est compromis. L'isolation est donc poussée là où elle est peu chère et expressive,
au niveau du projet :

- `NetworkPolicy` en *default-deny* dans chaque namespace de projet, avec une liste
  d'autorisations explicite : DNS, les namespaces du projet lui-même, la sortie du
  connecteur de tunnel vers Cloudflare, le chemin de l'opérateur de secrets vers Vault,
  le Keycloak du projet, le scraping de métriques, et GHCR.
- Des groupes de sécurité cloud limités aux ressources propres du projet.

Cilium est déjà le CNI on-prem, donc le langage de politique est disponible ; les
clusters cloud ont besoin de l'équivalent — c'est une exigence à transmettre à l'équipe
cluster.

**Frontière à transmettre.** Le VPN ne doit porter que des flux de plan de contrôle :
Argo CD vers l'API server, l'opérateur de secrets et Vault (dans les deux sens, cf.
D-07), et le `remote_write` de `vmagent`. Tout le reste qui le traverse est un bug de
conception. C'est écrit dans le contrat d'interface plutôt que laissé en hypothèse.

---

## D-13 — Nettoyage et décommissionnement

**Décision — renversée vers un suppresseur réel.** Ce qui est demandé est un job de
nettoyage qu'on peut lancer à la demande et qui **supprime réellement** les ressources,
pour repartir d'un état neuf — le comportement du script de clean actuel, mais scopé
proprement.

Trois portées, de la plus fine à la plus large :

| Portée | Effet |
| --- | --- |
| `--project <nom>` | Détruit toutes les ressources du projet **et de ses applications** sur son cloud, puis retire le projet de la plateforme (registre, enregistrement CMP, groupes et realm Keycloak, mount Vault, DNS et tunnels Cloudflare). |
| `--cloud <onprem\|aws\|gcp>` | Applique la portée projet à **tous** les projets dont `target_cloud` vaut ce cloud, dans un ordre de destruction déterminé. Sert à retirer un cloud une fois les tests terminés et à ne plus payer. |
| `--all` | Les trois clouds. Remise à zéro complète, réservée aux environnements de démonstration. |

**Garde-fous, parce que c'est un suppresseur.**

- **Simulation par défaut.** Sans `--yes`, l'outil n'affiche que l'inventaire de ce qu'il
  détruirait. Détruire exige un drapeau explicite.
- **Un inventaire avant d'agir.** L'outil liste ce qu'il a trouvé par fournisseur avant
  de toucher à quoi que ce soit, et ce même inventaire sert de rapport d'orphelins après
  coup (l'état de fin attendu est « zéro orphelin »).
- **Marquage uniforme obligatoire.** Chaque objet créé pour un projet porte `cnp.project`,
  `cnp.cloud` et `cnp.env` — en tags cloud, labels Kubernetes, commentaires Cloudflare et
  métadonnées Vault. Sans marqueur uniforme, « trouver tout ce qui appartient à X » est
  de la devinette, et un suppresseur qui devine est un suppresseur qui détruit à côté.
- **Ordre de destruction documenté** par projet, et runbook de décommissionnement d'un
  cloud qui fait passer chaque projet de ce cloud par cet ordre.

**Ce qui reste vrai de la prudence initiale.** Le mode simulation est le mode par
défaut, et la production reste protégée par le fait que la portée soit toujours
explicite : il n'existe pas d'invocation sans portée.
