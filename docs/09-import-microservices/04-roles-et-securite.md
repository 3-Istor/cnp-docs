# Rôles et sécurité

L'import fait entrer du code étranger dans la plateforme. Chaque identité reçoit donc le minimum
nécessaire : la personne agit avec ses rôles de projet, chaque automate a son identité étroite, et
le code importé ne s'exécute qu'à un seul endroit, sans secret. L'[audit de sécurité](08-audit-securite.md)
a été mené avec la méthode du skill `security-audit` de Cloudflare. Ses exigences sont reprises
ci-dessous.

![Qui peut faire quoi dans l'import](../01-architecture/diagrams/access-matrix-import.light.png#only-light)
![Qui peut faire quoi dans l'import](../01-architecture/diagrams/access-matrix-import.dark.png#only-dark)

## Personnes

| Action | guest | member | admin | owner | infra | security-auditor | finops |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Voir la carte du projet | | ✓ | ✓ | ✓ | ✓ | ✓ | |
| Voir le statut d'un import, ses rapports de scan et son SBOM | | ✓ | ✓ | ✓ | ✓ | ✓ | |
| Installer CNP-Portal sur ses repos et relier son compte GitHub | ✓ | ✓ | ✓ | ✓ | ✓ | | |
| Analyser un repo, éditer le plan | | | ✓ | ✓ | | | |
| Appliquer un import (passkey récente) | | | ✓ | ✓ | | | |
| Relier ou délier deux apps | | | ✓ | ✓ | | | |
| Reconstruire, supprimer une app importée (passkey) | | | ✓ | ✓ | | | |
| Coût par service importé | | ✓ | ✓ | ✓ | ✓ | | ✓ |
| Activer l'import sur la plateforme, gérer le catalogue de bases | | | | | ✓ | | |

Un membre peut lire la carte, mais pas les valeurs de configuration ni les secrets : même règle
que l'onglet Settings d'aujourd'hui. Un agent IA a exactement les droits de la personne pour qui
il agit, avec en plus une confirmation dans le portail pour toute écriture (voir plus bas).

## Identités techniques

| Identité | Ce qu'elle peut faire | Ce qu'elle ne peut pas faire | Où vit son secret |
| --- | --- | --- | --- |
| App GitHub **CNP-Portal**, sur l'org 3-Istor | Ce qu'elle fait aujourd'hui, avec des permissions réduites au besoin réel : créer les repos d'apps, écrire cnp-projects, lancer le workflow de build | Ce que ses permissions réduites n'autorisent plus | Clé privée dans Vault, utilisée par le CMP seul |
| App GitHub **CNP-Portal**, sur un compte utilisateur | Par des jetons restreints : lire le contenu et les métadonnées d'un seul repo choisi ; recevoir les `push` | Par ces jetons : écrire, ouvrir une PR, lire un repo non choisi. L'installation détient plus de droits, jamais utilisés (risque accepté) | Même clé, dans Vault, utilisée par le CMP seul |
| CMP | Émettre pour une installation utilisateur un jeton CNP-Portal d'une heure, `contents:read` et `metadata:read`, limité à un repo, après vérification OIDC du workflow ; écrire cnp-projects | Demander un jeton non restreint pour une installation utilisateur | Vault (rôle Kubernetes `cmp-backend`) |
| Workflow de build central | Job `prepare` : recevoir le jeton du CMP. Job `publish` : pousser `ghcr.io/3-istor/imports/<projet>/<app>` | Aucun secret durable ; aucun `id-token` dans le job qui exécute le code importé | Aucun, identité OIDC GitHub |
| Argo CD | Lire cnp-projects et infra-templates | Lire un repo d'utilisateur | Inchangé |
| Agent IA (MCP) | Lire la carte, analyser, proposer un brouillon | Appliquer, relier ou délier sans confirmation humaine dans le portail | Jeton de l'utilisateur, 5 minutes |

## Décisions de sécurité ajoutées par l'audit

| # | Décision | Constat |
| --- | --- | --- |
| D13 | Le CMP fixe lui-même le propriétaire GitHub et l'installation à la création d'app, au lieu de les lire dans la requête. Correctif sur `main`, avant l'import | Haute, confirmée : `POST /api/deployments` accepte un `github_installation_id` du client. Pas exploitable aujourd'hui, mais il le deviendrait dès qu'une autre installation existerait |
| D14 | La clé de CNP-Portal reste dans le CMP. Le job `prepare` reçoit un jeton d'une heure, en lecture, limité à un repo, en échange de son jeton OIDC | Une clé d'app dans le dépôt de CI pourrait lire les repos de tous les tenants |
| D15 | Jetons utilisateur restreints par repo et par permission : pour une installation utilisateur, le CMP demande toujours `permissions: {contents: read, metadata: read}` et `repositories: [<le repo>]`. Une seule fonction du CMP émet ces jetons, et elle refuse une installation non prouvée | Une app GitHub a un seul jeu de permissions pour toutes ses installations |
| D16 | Valeurs du bloc `workload` générées par le CMP, schéma strict : pas de privileged, de hostPath, de capacités ajoutées, ni de `host` dans les sondes. Kyverno en mode Deny est hors périmètre : chantier séparé | Le chart recopie `securityContext` et les sondes tels quels |
| D17 | Volumes des bases conservés à la suppression du StatefulSet ; le CMP les supprime explicitement après confirmation | Une mauvaise écriture du registre suivie d'un prune Argo détruirait des données |
| D18 | Alias limités aux apps du même projet ; la cible est calculée par le CMP, et les noms réservés (`kubernetes`, `vault`, services plateforme) sont refusés | Un alias libre pourrait relayer vers Vault ou l'API Kubernetes |
| D19 | `import_repo`, `link_apps` et `unlink_apps` au niveau `confirm` de la politique MCP : confirmation dans le portail, pas un booléen fourni par l'agent | Le contenu du repo lu par l'agent peut contenir une injection de prompt |
| D20 | L'analyseur tourne comme un job isolé avec ses limites, pas dans le processus du CMP | Une archive piégée (bombe tar, alias YAML) ferait tomber le portail |
| D21 | Credential de pull GHCR lisible par l'opérateur seulement, jamais monté dans un pod d'app | GitHub ne sait pas limiter un jeton de lecture à un seul paquet |

## Risques acceptés

- CNP-Portal installé chez un utilisateur détient sur ses repos toutes les permissions de l'app,
  que l'utilisateur accorde à l'installation. La plateforme n'en utilise que la lecture, par des
  jetons restreints. Une fuite de la clé privée donnerait ces droits complets : la clé ne sort
  pas de Vault et du CMP. Accepté par Brian le 13/11.
- Le code importé a accès au réseau pendant son build, sur un runner GitHub éphémère. Il ne peut
  nuire qu'à lui-même, borné par les délais.
- Les métriques Hubble d'un projet peuvent montrer le nom d'une charge d'un autre projet qu'il a
  tenté de joindre et qui a été refusée. Divulgation faible.
- Un webhook reçu pendant le sommeil du CMP (01:00 à 07:00) est perdu ; le rattrapage au réveil le
  compense.
