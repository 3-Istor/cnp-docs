# Plan et recette

Le travail est découpé en dix sous-tâches de
[CNP-158](https://3istors.atlassian.net/browse/CNP-158). L'ordre fait tourner epitweet de bout en
bout le plus tôt possible, puis industrialise l'import.

| Étape | Contenu | Tickets | Jalon |
| --- | --- | --- | --- |
| 1. Socle | Chart `workload` et StatefulSet (fait en local), presets des bases, types `service` et `backing` dans le schéma, ApplicationSet, image updater qui ignore ces types, credential de registre par projet | CNP-239 | Une app `backing` MongoDB tourne dans un projet de test |
| 2. Liens | `dependsOn`, alias ExternalName, tests réseau | CNP-240 | Un pod joint `mongodb:27017` par l'alias |
| 3. Epitweet à la main | Valeurs écrites à la main dans cnp-projects, images construites par une exécution du workflow | CNP-238, CNP-239 | **Epitweet tourne de bout en bout** |
| 4. Build | Workflow central (Dockerfile du repo, Dockerfile réparé, Railpack), callback OIDC, commit atomique du registre | CNP-238 | Un build relancé redéploie un seul service |
| 5. Analyse | Analyseur Python et plan d'import | CNP-237 | Le plan d'epitweet liste 7 apps, 4 bases et les liens |
| 6. Parcours | Liaison GitHub et GitLab, tables d'import, brouillon sur la carte, Apply avec step-up, suivi | CNP-241 | Epitweet importé depuis le portail |
| 7. Topologie | Métriques Hubble dans VictoriaMetrics, API, onglet avec React Flow | CNP-242 | Les arêtes apparaissent quand on clique dans epitweet |
| 8. Rebuild et MCP | Webhook, polling, rattrapage ; outils MCP | CNP-243, CNP-244 | Un push redéploie le service touché |
| 9. Recette et doc | Corpus, epitweet, fermeture des PRs de Joe | CNP-245, CNP-246 | Tout ci-dessous est vert |

## Recette

Tout se passe sur le vrai cluster, dans des projets de test supprimés à la fin.

**Epitweet**

- [ ] Le plan liste 7 apps, 4 bases et les liens du front vers les 6 services, et signale api-gateway comme non supporté.
- [ ] Créer un compte, poster, suivre, liker et chercher fonctionnent depuis l'URL publique.
- [ ] Les arêtes observées apparaissent sur la carte en moins de 2 minutes.
- [ ] Les mots de passe du repo ne sont pas utilisés : les secrets viennent de Vault.
- [ ] La suppression du projet ne laisse ni namespace, ni volume, ni image.

**Corpus de repos publics** (au moins 9 sur 10 déployés et joignables)

| Type | Exemple visé |
| --- | --- |
| Site statique | HTML et CSS seuls |
| SPA React ou Vite | front sans Dockerfile, construit par Railpack |
| FastAPI | API Python avec Dockerfile |
| Django | avec Postgres, remplacé par CNPG |
| Express | Node sans Dockerfile |
| Go | binaire HTTP |
| Spring Boot | Maven, avec Dockerfile multi-étapes |
| Quarkus | Dockerfile qui attend un artefact compilé |
| Monorepo avec docker-compose | plusieurs services et une base |
| Repo GitLab | importé par OAuth GitLab |

**Sécurité**

- [ ] Un member ne peut ni analyser ni importer : le bouton est absent, et l'API répond 403.
- [ ] Un callback de build rejoué, ou venant d'une autre tentative, est refusé.
- [ ] Un webhook à signature invalide est refusé et journalisé.
- [ ] Un alias vers un service hors du projet est refusé à l'écriture.
- [ ] Une archive piégée (lien symbolique sortant, fichier énorme, chemin `..`) est refusée sans lecture hors du dossier.
- [ ] Les constats critiques et majeurs de l'audit de sécurité sont fermés ou acceptés par écrit.
