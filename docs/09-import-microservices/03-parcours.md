# Parcours utilisateurs

Chaque parcours a son diagramme de séquence, généré depuis le même modèle que les vues
d'architecture. Les rôles cités sont ceux du portail : *owner*, *admin*, *member*, *guest* du
projet, et les groupes plateforme `infra`, `security-auditor` et `finops`.

## 1. Installer CNP-Portal sur ses repos

*Qui : tout utilisateur connecté. Une seule fois par compte ou par organisation GitHub.*

1. Depuis l'import, s'il n'a pas de compte relié, l'utilisateur clique sur « Connecter GitHub ».
2. GitHub propose d'installer l'app **CNP-Portal** : l'utilisateur choisit les repos qu'il
   expose. Les jetons que le CMP en tire pour ces repos ne donnent que la lecture du contenu
   et des métadonnées, pour un seul repo à la fois.
3. Au retour, le CMP lance un OAuth explicite avec un nouveau state et PKCE. Seul le serveur
   échange le code, et aucun jeton GitHub n'arrive au navigateur.
4. Le CMP vérifie avec le jeton de l'utilisateur que l'installation lui est visible
   (`GET /user/installations`). Il n'enregistre l'installation qu'ensuite. Un `installation_id`
   seul ne donne jamais aucun droit.

![Installer CNP-Portal](../01-architecture/diagrams/import-link-github.light.png#only-light)
![Installer CNP-Portal](../01-architecture/diagrams/import-link-github.dark.png#only-dark)

## 2. Analyser un repo et relire le plan

*Qui : admin ou owner du projet.*

1. Dans le projet, « Importer un repo » : repo GitHub, branche.
2. Le CMP résout la branche en SHA. À partir de là, tout porte sur ce SHA.
3. L'analyseur télécharge l'archive à ce SHA avec un jeton limité à ce seul repo, en lecture.
   Il borne la taille et le nombre de fichiers, et refuse les liens symboliques sortants. Il
   n'exécute rien.
4. Il produit le plan d'import, affiché en brouillon sur la carte du projet :
    - services et bases détectés ;
    - liens proposés ;
    - méthode de build par service ;
    - confiance par élément ;
    - jauge mémoire ;
    - éléments non supportés ;
    - secrets trouvés.
5. L'utilisateur corrige ce qu'il faut : port, sonde, ressources, nom d'app, lien à retirer,
   base à remplacer. Il peut aussi abandonner, et rien n'a alors été créé.

![Analyser un repo](../01-architecture/diagrams/import-analyze.light.png#only-light)
![Analyser un repo](../01-architecture/diagrams/import-analyze.dark.png#only-dark)

## 3. Appliquer l'import

*Qui : admin ou owner, avec un step-up par passkey.*

1. « Appliquer » demande une passkey récente, comme toute action sensible du portail.
2. Le CMP enregistre l'import et son approbateur, puis lance le workflow de build central avec
   des identifiants opaques. Aucun jeton ne passe en entrée.
3. Le job `prepare` présente son jeton OIDC au CMP, qui lui remet un jeton CNP-Portal d'une
   heure, en lecture, limité au repo. Le job télécharge le source, puis lance gitleaks et
   `trivy fs`, qui bloquent seulement sur les secrets et les erreurs critiques.
4. Le job `build` n'a ni secret ni identité. Il construit chaque image, la scanne et produit son
   SBOM.
5. Le job `publish` pousse les images par digest. Il rappelle ensuite le CMP, qui vérifie le
   jeton OIDC de GitHub : repo, workflow à une révision épinglée, run, tentative.
6. Le CMP revérifie les droits de l'utilisateur. Il écrit en un seul commit les valeurs, les
   digests et les liens dans cnp-projects.
7. Argo CD synchronise. Le CMP attend Synced et Healthy sur les révisions exactes, puis lance le
   test fonctionnel.
8. Le suivi montre chaque service. Un échec sur un service n'arrête pas les autres et dit quoi
   faire.

![Appliquer un import](../01-architecture/diagrams/import-apply.light.png#only-light)
![Appliquer un import](../01-architecture/diagrams/import-apply.dark.png#only-dark)

## 4. Reconstruire à chaque push

*Qui : automatique. Bouton « Reconstruire » pour admin et owner.*

- **Push :** CNP-Portal envoie l'événement `push`. Le CMP vérifie la signature HMAC et ignore
  les livraisons déjà vues. Il ne reconstruit que les services dont le dossier a changé.
- **Au réveil du CMP :** il compare le dernier SHA construit à la tête de chaque branche, et
  rattrape ainsi un webhook perdu la nuit.

![Reconstruire](../01-architecture/diagrams/import-rebuild.light.png#only-light)
![Reconstruire](../01-architecture/diagrams/import-rebuild.dark.png#only-dark)

## 5. Relier deux apps

*Qui : admin ou owner. Marche pour les apps importées et pour celles nées d'un template.*

Sur la carte, l'utilisateur tire un trait d'une app vers une autre, ou clique sur « Déclarer ce
lien » sur une arête ambre. Il choisit le port et l'alias, par défaut le nom de la cible. Le CMP
écrit le `dependsOn`. Le chart de l'appelant rend l'alias ExternalName. L'arête passe en trait
plein dès que le trafic est observé.

![Relier deux apps](../01-architecture/diagrams/app-link.light.png#only-light)
![Relier deux apps](../01-architecture/diagrams/app-link.dark.png#only-dark)

## 6. Un agent IA importe un repo

*Qui : l'agent de l'utilisateur, avec ses droits, via MCP.*

`analyze_repo` et `get_topology` sont en lecture. `import_repo` crée un brouillon et répond
`needs_user_confirmation`. L'utilisateur relit le plan dans le portail et l'applique lui-même,
avec sa passkey. Le contenu du repo est une donnée pour l'agent, jamais une instruction : un
README qui « demande » quelque chose n'a aucun effet sur le plan.

![Un agent importe un repo](../01-architecture/diagrams/import-agent.light.png#only-light)
![Un agent importe un repo](../01-architecture/diagrams/import-agent.dark.png#only-dark)
