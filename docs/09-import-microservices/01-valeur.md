# Valeur pour les utilisateurs

Les templates ne couvrent que les nouvelles apps. Or la plupart des équipes arrivent avec un
existant : un monorepo, un docker-compose, des bases, des services qui s'appellent entre eux.
L'import fait de la CNP une plateforme qu'on adopte avec son code actuel. La topologie la rend
lisible, et sûre à faire évoluer.

## Ce que disent les études de 2025-2026

- **L'adoption est le premier frein**, pas la technique. 45,3 % des équipes plateforme peinent à
  faire adopter leur plateforme, et 43,9 % citent la complexité de l'existant
  ([State of Platform Engineering vol. 4](https://platformengineering.org/blog/the-biggest-challenges-platform-engineering-teams-are-facing-in-2026), 518 répondants).
- **L'existant ne rentre pas dans les golden paths.** Dans un cas rapporté par la CNCF, un golden
  path a atteint 85 % d'adoption, puis 60 % du temps de l'équipe plateforme est parti dans les
  exceptions ([CNCF, 01/09/2026](https://www.cncf.io/blog/2026/09/01/platform-engineering-maturity-from-toolchain-to-self-service/)).
- **Les développeurs attendent.** 78 % patientent un jour ou plus pour une aide DevOps, et 3 %
  seulement font entièrement confiance aux métadonnées de leur portail
  ([Port, 2025](https://www.port.io/state-of-internal-developer-portals), étude d'un éditeur).
- **L'IA ne vaut que gouvernée.** La confiance dans l'IA monte à 94 % sous une gouvernance formelle,
  contre 51 % sans
  ([Perforce, 2026](https://www.perforce.com/press-releases/state-of-platform-engineering-2026)).
  DORA 2025 décrit l'IA comme un amplificateur de la qualité de la plateforme
  ([Google Cloud](https://cloud.google.com/blog/products/ai-machine-learning/announcing-the-2025-dora-report)).
- **L'import d'un docker-compose avec ses bases reste un créneau ouvert.** Railway détecte les
  monorepos JavaScript ; on n'a trouvé ni chez Render ni chez Northflank d'import d'un compose
  avec ses bases. La demande se déduit des freins ci-dessus, aucune étude ne la mesure.

## Pour qui, et pour quoi faire

| Persona | Importer l'existant | Comprendre le système | Changer sans casser |
| --- | --- | --- | --- |
| Développeur | Monter son repo sans le réécrire, voir ce que l'import décide avant qu'il s'applique | Savoir qui appelle son service et ce dont il a besoin | Reconstruire à chaque push, revenir en arrière |
| Tech lead | Choisir l'ordre de migration des services | Vérifier que l'architecture déclarée est celle qui tourne | Voir l'impact d'un lien avant de l'ajouter |
| Équipe plateforme | Ne plus traiter les imports à la main | Garder un catalogue juste sans effort | Durcir le réseau sans casser les équipes |
| RSSI | Savoir ce qui est importé, avec quels secrets et quelle exposition | Voir qui parle à quoi, et ce qui est refusé | Une preuve par changement : plan, approbateur, SBOM, scans |
| FinOps | Coût rattaché au projet et à chaque service dès l'import | Coût par service | Estimation avant d'appliquer |
| Auditeur | Journal de qui a importé et approuvé quoi | Inventaire avec propriétaires | Chaque changement relié à une approbation |
| Agent IA | Analyser et proposer un import avec les droits de l'utilisateur | Lire un graphe du projet exploitable par machine | Proposer, puis attendre la confirmation humaine |

## Les fonctions retenues

La v1 vise ce qui répond aux freins mesurés et se montre en démonstration. Le reste est noté pour
la suite, avec son effort.

| Fonction | Persona | Ce qu'elle change | Version |
| --- | --- | --- | --- |
| Plan d'import relisible sur la carte du projet : services, bases, liens, méthode de build, confiance par élément, rien ne s'applique avant validation | Développeur, RSSI | L'existant entre sans réécriture, et l'utilisateur voit chaque décision | v1 |
| Jauge de capacité dans le plan (RAM demandée contre RAM libre du projet) | Développeur, FinOps | Pas de déploiement qui échoue faute de place | v1 |
| Panneau « non supporté » : composant sans source, port hôte ignoré, route de santé absente | Développeur | Confiance : la plateforme dit ce qu'elle ne sait pas faire | v1 |
| Secrets en dur détectés et remplacés par des secrets Vault générés | RSSI | Aucun secret du repo en production | v1 |
| Liens déclarés et trafic observé sur la même carte, bouton « Déclarer ce lien » | Tech lead, RSSI | Le catalogue reste juste parce qu'il se compare au réel | v1 |
| Relier deux apps en un clic, y compris des apps nées d'un template | Développeur | Les noms d'origine (`user-service:8081`) marchent sans toucher au code | v1 |
| Dossier de preuve par import : plan, approbateur, SHA, digests, SBOM, résultats de scan | Auditeur, RSSI | La conformité sans travail en plus ; utile pour le Cyber Resilience Act | v1 |
| Test fonctionnel après déploiement, affiché dans le suivi | Développeur | « Ça tourne » prouvé, pas supposé | v1 |
| Agents IA : analyser, lire la carte, proposer un import, toujours confirmé par l'humain | Agent, développeur | IA utile et gouvernée | v1 |
| « Explique mon architecture » et « que casse la panne de X » depuis le graphe, par MCP | Nouveau venu, manager | Prise en main plus rapide | v1 (lecture) |
| Isolation par liens avec prévisualisation de ce qui serait bloqué | RSSI | Moindre privilège réseau sans casse | v2 |
| Coût par service et par lien sur la carte | FinOps | Arbitrages visibles | v2 |
| Rapport de migration : score de préparation par service, ordre conseillé | Tech lead, manager | Démarrer la migration d'un portefeuille | v2 |
| Sortie sans enfermement : export des manifests et de la config Argo | Manager | Réponse directe à la peur du lock-in | v2 |
| Environnement de prévisualisation par PR, bases comprises | Développeur | Revue sur un environnement réel | v3 |
