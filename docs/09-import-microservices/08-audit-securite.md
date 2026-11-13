# Audit de sécurité : import de dépôts externes dans CNP

Date : 2026-11-13. Méthode : skill `security-audit` de Cloudflare, mode scoped. Les 14 constats existent aussi au format JSON du skill, validés par son script `validate-findings.cjs`.

## 1. Synthèse

Le design v2 corrige le défaut majeur du design de Joe : il crée une GitHub App dédiée `CNP-Import` en lecture seule au lieu d'installer `CNP-Portal` sur les comptes des utilisateurs. Il faut garder cette décision. L'audit du code montre qu'installer CNP-Portal chez les utilisateurs serait exploitable dès aujourd'hui par un admin de n'importe quel projet. `POST /api/deployments` accepte un `github_installation_id` et un `github_owner` fournis par le client, puis Terraform agit sur ce compte avec les droits complets de CNP-Portal.

Le code de Joe à récupérer est de bonne qualité sur les points sensibles. Le flux OAuth (state à usage unique lié à la session, PKCE S256, jetons Fernet), la vérification OIDC du worker et les transitions d'état gardées résistent aux attaques testées. On peut les reprendre en changeant l'App ciblée.

Les risques principaux restants sont dans le design :

1. La clé privée de CNP-Import est prévue dans le dépôt de CI partagé. Elle peut lire les dépôts de tous les tenants.
2. Le jeton GitLab complet de l'utilisateur est transmis à la CI.
3. Les champs `securityContext` et les probes sont recopiés tels quels par le chart, alors que Kyverno est seulement en mode Audit.
4. Le credential de pull GHCR est valable pour toute l'org.
5. La confirmation MCP « ask » est un booléen fourni par l'agent lui-même, alors que l'agent lit le contenu du dépôt importé.

| Gravité | Nombre |
|---|---|
| High (confirmé) | 1 |
| À valider (design non construit ou fait externe) | 9 |
| Rejetés après vérification | 4 |

## 2. Périmètre et méthode

**Périmètre A (design).** La spécification v2 complète (pages Architecture et Parcours) : liaison GitHub et GitLab, analyseur, workflow central, écritures dans le registre, Argo CD, alias ExternalName, Hubble, outils MCP et webhooks.

**Périmètre B (code à récupérer).**

- Dans `CMP@origin/import-fichier-joe-raph` (9dd4e3e) : `github_account.py`, `routers/account.py`, les routes Next `api/auth/github/{link,callback}` et `repository_imports.py` (`verify_token`, `_worker`, `transition`).
- `extern-project-analyzer@d424023` : `import-project.yml`, `import_worker.py`, `import_runtime.py`, `prepare_service.py`, `git_service.py`.
- Dépendances lues pour le contexte : `CMP@origin/main` (a91377a) pour `account.py`, `deployments.py`, `terraform_orchestrator.py` et `mcp_server.py`, `app-templates` (a811797), `infra-templates@feat/cnp-239-workload` (5b007ae) et `K3s` (ead0530) pour Kyverno.

**Méthode.**

- Lecture source seule. Aucune exécution de code cible, aucun appel à GitHub ni aux environnements déployés.
- Classes d'attaque issues de `WEB-PROTOCOL-AND-AUTH`, `CLOUD-AND-DEPLOYMENT`, `SUPPLY-CHAIN-AND-RELEASE`, `DATA-ISOLATION-AND-LIFECYCLE`, `AI-AND-LLM` et `RESOURCE-EXHAUSTION`.
- Seul le constat confirmé a reçu une revérification indépendante par un agent distinct, qui l'a maintenu. Il a ajouté que l'identifiant stocké est réutilisé ensuite (`deployments.py:510`, `security.py:496`, `gitops.py:60`) et qu'il faut que le template 3-Istor soit lisible par le jeton.
- Un composant non construit ne peut pas être « confirmé ». Les menaces de design sont donc classées `needs_validation`, sans gravité, avec le bloqueur exact. Leur priorité apparaît dans la checklist de la section 4.

**Limites.**

- Profil scoped : les problèmes préexistants de la plateforme ne sont retenus que s'ils interagissent avec l'import.
- Aucun registre de couverture antérieur. Ce passage ne prétend pas à l'exhaustivité.
- Chemins des preuves : relatifs au dossier qui contient les dépôts de la CNP.

## 3. Constats par gravité

### HIGH : CMP fabrique un jeton CNP-Portal pour un `github_installation_id` fourni par le client

`cmp/deployments/client-installation-id-confused-deputy`. Confirmé.

**Preuve.**

1. `CMP/backend/app/routers/deployments.py:182` reçoit `app_config` d'un admin de projet. La ligne 267 le stocke tel quel.
2. `terraform_orchestrator.py:84-97` appelle `get_installation_token(int(app_config["github_installation_id"]))`. Le jeton a toutes les permissions de l'App (`github_service.py:88`, sans restriction aux dépôts ni aux permissions utiles). `terraform_executor.py:470` transmet chaque clé en `-var`.
3. `app-templates/.../providers.tf:30` configure le provider GitHub avec `owner = var.github_owner`. `github_repository` et `github_repository_file` sont donc créés dans ce compte.
4. Sur main, `POST /api/account/github-installation` (`account.py:393`) stocke n'importe quel identifiant sans preuve.

**Exploitabilité.** Aujourd'hui, CNP-Portal n'est installée que sur 3-Istor : un identifiant étranger échoue chez GitHub. Avec le design de Joe, `app_info()` vérifie `GITHUB_APP_ID` dans `github_account.py:446`. L'OAuth et l'installation se font donc sur CNP-Portal elle-même. Un admin de projet A qui connaît l'identifiant d'installation d'un utilisateur V (des entiers non secrets) pourrait créer des dépôts et écrire des fichiers chez V. Le jeton porte aussi administration, workflows et secrets en écriture.

**Correctif.**

- Pour `k3s-gitops-app`, imposer côté serveur `github_owner = 3-Istor` et l'installation de l'org.
- Supprimer ou protéger `POST /account/github-installation`.
- Pour l'import, utiliser une fonction distincte `import_installation_token(source_id)`. Elle ne lit l'identifiant que dans une ligne `import_sources` prouvée par le flux de Joe. Elle utilise la clé CNP-Import et demande `repositories=[un dépôt]` avec `contents:read`.

**Avis sur la demande de Joe d'installer CNP-Portal chez les utilisateurs : à refuser.** L'écran de consentement accorderait des droits de classe admin d'org sur les dépôts personnels. La clé CNP-Portal est déjà répartie entre CMP, Terraform, les repo-creds d'Argo et Image Updater. Chacun de ces détenteurs deviendrait un point de compromission de tous les comptes utilisateurs. Le constat ci-dessus montre qu'un chemin d'abus existe déjà dans le code. À vérifier : la liste `GET /app/installations` de CNP-Portal, pour s'assurer qu'aucune installation de test n'existe hors 3-Istor. Il faut aussi passer l'App en « Only on this account ».

### À valider (design) : les priorités

**D1. Clé privée CNP-Import dans l'environnement du dépôt de CI.** `design/build-repo-import-app-private-key`.

- Source : spec, lignes 65 et 110.
- Problème : cette clé peut fabriquer un jeton pour chaque installation utilisateur. Toute personne capable d'exécuter un workflow modifié dans cet environnement peut l'exfiltrer : push sur main, action ou dépendance pip compromise (`import-project.yml:52` installe des paquets épinglés en version mais sans hash, dans un job `id-token: write`). Elle lit alors les dépôts de tous les tenants. L'épinglage de `workflow_sha` côté CMP protège les callbacks, pas les secrets.
- Correctif : garder la clé dans CMP (Vault). Le job prepare présente son jeton OIDC (run_id et attempt_id liés) et reçoit un jeton d'une heure, limité à un dépôt et à `contents:read`. Ainsi le dépôt de CI ne contient plus aucun secret durable.

**D2. Jeton GitLab utilisateur transmis à la CI.** `design/gitlab-user-token-in-ci`, spec lignes 42 et 66.

- Problème : les scopes `read_api` et `read_repository` couvrent tout le compte, pas un seul dépôt.
- Correctif : CMP télécharge lui-même l'archive au SHA fixé et la sert par une URL à usage unique liée au run. Lier aussi l'instance GitLab au `state` OAuth pour éviter la confusion d'IdP. L'URL d'instance vient d'une allowlist serveur, jamais de l'utilisateur.

**D3. Injection de valeurs de workload.** `design/workload-values-injection-audit-only-pss`.

- Preuve : `infra-templates/templates/_helpers.tpl:212,226,249` rend probes, `securityContext` et `podSecurityContext` avec `toYaml`. `K3s/.../pod-security-restricted.yaml:10` est en `validationActions: Audit`.
- Risque : si l'ImportPlan (issu d'un compose non fiable : `privileged`, `cap_add`, `user: root`) ou une carte éditable passe ces champs, un tenant obtient un pod privilégié. Il peut aussi obtenir une probe `httpGet.host` qui fait de la SSRF depuis le réseau du nœud via le kubelet.
- Correctif :
  - CMP n'écrit que des valeurs générées : `runAsNonRoot`, UID numérique, `drop ALL`, pas de `host` dans les probes.
  - Le chart ajoute un `values.schema.json` restrictif.
  - Kyverno passe en `Deny` sur les namespaces projet avant l'activation de l'import. Les apps template rendent un manifeste identique, donc elles sont déjà conformes.

**D4. Credential de pull GHCR à l'échelle de l'org.** `design/registry-pull-credential-org-wide`, spec ligne 86.

- Problème : GHCR n'a pas de jeton limité à un paquet. Si les service accounts des apps peuvent se connecter au rôle Vault, le code d'une app importée peut tirer les images des autres projets, qui contiennent leur code.
- Correctif : lier le rôle Vault au seul SA de l'opérateur de secrets. À terme, utiliser un registre interne avec un compte robot par projet.

**D5. Confirmation MCP fournie par l'agent.** `cmp/mcp/agent-asserted-confirmation-import-link`.

- Preuve : `mcp_server.py:298` laisse passer l'action dès que `user_confirmed=true`. Or `analyze_repo` renvoie au modèle du contenu contrôlé par l'auteur du dépôt : commentaires, README, variables d'environnement.
- Risque : une injection dans le dépôt peut provoquer `link_apps` ou `import_repo`.
- Correctif : niveau `confirm` (approbation dans CMP liée à l'outil, aux arguments complets, au projet et au demandeur) pour `import.apply` et `app.link`. Ne renvoyer à l'agent que des champs structurés, sans texte libre. Masquer les valeurs d'environnement qui ressemblent à des secrets : le compose d'Epitweet contient par exemple `neo4j/password` et `admin:admin`.

**D6. Artefacts et logs publics.** `epa/public-workflow-artifacts-source-leak`.

- Preuve : `import-project.yml:59,118` publie `source.tar` et `image.tar` en artefacts. La sortie de `docker build` va dans les logs.
- Risque : si `3-Istor/extern-project-analyzer` est public, tout utilisateur GitHub lit le code privé importé.
- Vérification : `gh repo view --json visibility`. Le dépôt doit être privé.

**D7. Binaire Trivy non épinglé.** `epa/trivy-setup-unpinned-binary`, `import-project.yml:100-109`.

- Problème : `skip-setup-trivy: false` sans version ni checksum. Un Trivy compromis, comme dans l'incident de 2026-03, s'exécute dans le job qui produit l'artefact publiable. Il peut donc piéger les images de tous les tenants.
- Correctif : installer Trivy par hash, dans un job de scan séparé. Fixer le digest en sortie d'une étape de build exécutée avant le scan. Ajouter une attestation de provenance au job publish et `verifyImages` dans Kyverno.

**D8. Analyseur exécuté dans CMP.** `design/analyzer-archive-yaml-exhaustion`.

- Preuve : `import_worker.py:122` borne seulement le nombre de fichiers (5000) et la taille par fichier (512 Mio), après un clone non borné.
- Risque : bombe gzip, alias YAML, interpolation `${}` récursive, ReDoS sur la config nginx. Dans le pod CMP, cela bloque le portail pour tous les projets.
- Correctif : Job isolé (512 Mio, 30 s CPU), plafonds de taille totale, de nombre d'entrées et de profondeur, `tarfile` avec `filter="data"` (refuse liens, devices et `..`), loader YAML sans alias, pas de résolution réseau des URL trouvées.

**D9. ExternalName et Gateway.** `design/externalname-alias-gateway-deputy`.

- Constat : un ExternalName n'ouvre aucun flux, car Cilium filtre par IP.
- Risque : si le nom d'alias entre en collision avec un Service référencé par une HTTPRoute, ou si la cible est libre, la Gateway du projet relaie le trafic public vers `vault.vault.svc` ou l'API server (classe CVE-2021-25740).
- Correctif : la cible est calculée par CMP (`<app>.<p>-<app>.svc.cluster.local`, même projet seulement). Le nom d'alias ne doit pas être un nom généré par le chart. Interdire les backends ExternalName dans la Gateway.

### Rejetés après vérification (récupérables)

- **CSRF du callback OAuth GitHub.** State haché, lié à l'utilisateur, au `sid` et à un cookie httpOnly, consommé de façon atomique, avec PKCE (`github_account.py:269-288`).
- **Prise d'un import par un autre run.** Il faut l'`attempt_id` et le même workflow épinglé par SHA. Toutes les transitions sont gardées (`repository_imports.py:558,951`). Invariant à conserver : aucun code non fiable dans un job `id-token: write`, car l'OIDC ne distingue pas les jobs d'un run.
- **Réécriture de `metadata.image` par le job build.** CMP compare avec l'image serveur avant le push (`repository_imports.py:1091`). Résidu : `image_scan_passed` et `http_check_passed` sont déclarés par le job lui-même, d'où le correctif de D7.
- **Traversée de chemin via `project_path`.** `prepare_service.py:256,581` : regex, `resolve` puis `relative_to`, et rejet des symlinks, sous-modules et LFS.

## 4. Exigences de sécurité à intégrer au design

**GitHub App CNP-Import et liaison de compte**

- [ ] App distincte : `contents:read` et `metadata:read`, événement `push`. CNP-Portal reste « Only on this account ».
- [ ] Reprendre `github_account.py` en pointant `app_info` et `generate_jwt` vers CNP-Import, avec une clé et un module séparés.
- [ ] Un identifiant d'installation n'est accepté que via la preuve `/user/installations` et l'appel App JWT de Joe. Les lignes héritées indexées par `preferred_username` sont ignorées.
- [ ] Jetons utilisateur : `MultiFernet` avec rotation, clé dans Vault et non dans les variables d'environnement, révocation à la déliaison.
- [ ] Corriger `deployments.py` comme décrit dans le constat HIGH.

**Autorisation dans CMP**

- [ ] Chaque endpoint `/api/import/*` et topologie vérifie le rôle sur le projet de la ressource, pas seulement sur le projet de l'URL : contrôle d'IDOR par `import_source_id` et `build_id`.
- [ ] `dependsOn` et `link_apps` sont limités au même projet, côté serveur.
- [ ] Rebuild (webhook ou polling) : revérifier à chaque fois le rôle CNP et l'accès au dépôt de l'utilisateur qui a lié la source, comme le fait `authorize` de Joe. S'arrêter si l'utilisateur a quitté le projet.

**Workflow de build**

- [ ] Entrées de dispatch : UUID validés et seulement passés via `env:`. Dépôt et SHA sont obtenus de CMP par OIDC, jamais lus dans les inputs.
- [ ] Aucune clé durable dans le dépôt de CI (D1). Protection de branche et revue obligatoire sur `extern-project-analyzer`.
- [ ] `id-token: write` uniquement dans prepare et publish, qui n'exécutent aucun code utilisateur. Tout code utilisateur, y compris le reactor Maven et Railpack, tourne dans `docker build` et jamais sur l'hôte du runner.
- [ ] Pas de cache GHA ni `actions/cache` partagé entre projets. Le scope de cache inclut le projet, ou il n'y a pas de cache.
- [ ] Le frontend Railpack et les lignes `# syntax=` sont épinglés par digest, ou les directives `syntax` non approuvées sont refusées.
- [ ] Actions épinglées par SHA complet. `requirements` avec `--require-hashes`. Trivy et gitleaks installés par hash dans un job séparé (D7).
- [ ] Le callback OIDC vérifie `repository_id`, `repository_owner_id`, `job_workflow_ref` et `job_workflow_sha`, en plus des contrôles de Joe.
- [ ] Image publiée sous `ghcr.io/3-istor/imports/<projet>/<app>@digest`, avec attestation de provenance. Argo ne déploie que par digest.
- [ ] Dépôt de CI privé (D6). gitleaks avec `--redact`. Pas d'artefact SARIF contenant des secrets.

**Analyseur**

- [ ] Job isolé avec limites (D8). Tarball au SHA fixé. Rejet des liens, hardlinks, devices, LFS et sous-modules.
- [ ] Aucune URL extraite (proxy_pass, variables d'environnement) n'est résolue ni appelée. Pas de SSRF : les hôtes GitHub et GitLab viennent d'une allowlist.
- [ ] Valeurs secrètes détectées masquées dans l'ImportPlan, l'UI, les logs et les réponses MCP.

**GitOps, chart et cluster**

- [ ] Schéma de valeurs restrictif, valeurs générées par CMP, Kyverno en `Deny` (D3).
- [ ] Chemins `apps/<p>/<app>/values.yaml` construits à partir de noms validés par regex. Un commit ne touche que le projet autorisé.
- [ ] ExternalName : cible calculée et noms réservés (D9).
- [ ] Credential de pull lisible par l'opérateur seulement (D4).
- [ ] StatefulSet : `persistentVolumeClaimRetentionPolicy` sur `Retain` par défaut, avec suppression explicite après confirmation. Cela évite qu'un prune Argo dû à une mauvaise écriture du registre détruise des données.
- [ ] Capacité : le quota du projet (2 Gi) reste fixé par la plateforme. Un import ne l'augmente jamais. L'estimation de capacité est vérifiée côté serveur avant Apply (Epitweet demande environ 3,4 Gi).

**Webhooks et topologie**

- [ ] HMAC `X-Hub-Signature-256` comparé en temps constant sur le corps brut. Déduplication par `X-GitHub-Delivery`. Correspondance par `installation.id` et `repository.id`, pas par nom. Le webhook ne fait que déclencher une relecture du HEAD par CMP.
- [ ] Requêtes VictoriaMetrics construites côté serveur avec des matchers de namespace issus de l'autorisation. Pas de PromQL utilisateur.
- [ ] Port 9965 de Hubble joignable par vmagent seulement.

**Agents IA**

- [ ] `import.apply`, `app.link` et `app.unlink` au niveau `confirm` (D5). Le contenu du dépôt est étiqueté comme donnée dans les réponses d'outil.

## 5. Risques résiduels acceptés

- **Code utilisateur avec accès réseau dans les runners GitHub hébergés.** Pendant `docker build` et le smoke test, le code peut exfiltrer sa propre source ou abuser du runner. C'est un impact sur soi uniquement, borné par les timeouts.
- **Booléens de scan déclarés par le job build.** Acceptable une fois le digest fixé avant le scan et le scan isolé dans un job séparé.
- **Fuite de topologie inter-projets.** Les métriques Hubble montrent les flux refusés vers un autre projet, et donc le nom de la charge cible. C'est une divulgation faible.
- **Disponibilité limitée par l'offhours guard.** Les webhooks reçus entre 01:00 et 07:00 sont perdus et rattrapés par la réconciliation au réveil.
- **Jetons utilisateur GitHub et GitLab stockés chiffrés dans la base CMP.** La compromission de CMP expose des jetons en lecture seule, ce qui est acceptable avec CNP-Import en lecture seule. Ce ne serait pas acceptable avec CNP-Portal.
- **Couverture partielle.** Pas d'exécution locale. Vérifié depuis : le dépôt de CI `extern-project-analyzer` est privé, et CNP-Portal n'est installé que sur l'org 3-Istor. Le comportement de la Gateway envers un backend ExternalName reste à tester.
