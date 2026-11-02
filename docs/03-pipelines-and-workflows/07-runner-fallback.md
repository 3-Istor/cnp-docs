# Runner de secours (quota Actions)

L'org est en plan GitHub Free : 2 000 min/mois de runners hébergés pour les dépôts **privés** (un dépôt public n'est jamais bloqué). Les workflows restent sur `ubuntu-latest`. Le runner d'org `epimac-github-runner` (labels `self-hosted`, `Linux`, `X64`, groupe Default) ne sert que quand ce quota est épuisé.

## Reconnaître le blocage

- Run en échec immédiat, jobs non démarrés, annotation : `The job was not started because recent account payments have failed or your spending limit needs to be increased`.
- Confirmer dans Settings › Billing de l'org (owner) : minutes Actions consommées.
- Un job `Queued` sans fin est un autre problème : aucun runner en ligne ne porte les labels demandés.

## Ce qui bascule

Fichiers d'une app générée par `app-templates/templates/k3s-gitops-app` :

| Fichier | Origine | Jobs |
|---|---|---|
| `.github/workflows/ci.yml` | copie du template à la création | fastapi-react : `backend-test`, `migration-lint`, `frontend-build-check`, `docker-build-and-push`, `dependabot-merge` ; html-css : `static-check`, `docker-build` (+ `dependabot-merge`) |
| `.github/workflows/cnp-verify.yml` | copie du template | `verify`, lancé par le CMP avant une promotion |
| `.github/workflows/security.yml` | écrit par Terraform (`github_repository_file.security_workflow`, source `k3s-gitops-app/security.yml`) | `gitleaks`, `trivy-fs`, `trivy-config`, `discover-images`, `trivy-image`, `report` |

Le `security.yml` des templates est remplacé dans l'app générée. Modifier un template ne touche que les apps créées ensuite.

## Basculer

`runs-on` est écrit en dur : il faut éditer chaque dépôt privé bloqué.

```bash
gh repo clone 3-Istor/<app> && cd <app>
sed -i 's/runs-on: ubuntu-latest/runs-on: [self-hosted, Linux, X64]/' .github/workflows/*.yml
git commit -am "ci: run on the self-hosted runner while the Actions quota is exhausted"
git push   # puis relancer les runs bloqués : gh run rerun <id>
```

!!! warning "Terraform réécrit `security.yml`"
    Un nouvel `apply` de l'app remet `ubuntu-latest` dans `security.yml` (contenu géré par `github_repository_file`).

Un seul runner : les jobs (et les matrices) passent un par un.

## Revenir

Quand le quota revient (nouveau cycle de facturation ou plafond relevé) : `git revert <sha du commit de bascule>` et push.

## Sans éditer chaque dépôt (non implémenté)

Il faudrait, dans chaque job des trois fichiers (templates et `k3s-gitops-app/security.yml`) :

```yaml
runs-on: ${{ fromJSON(vars.CNP_RUNS_ON || '"ubuntu-latest"') }}
```

Bascule : `gh variable set CNP_RUNS_ON -R 3-Istor/<app> --body '["self-hosted","Linux","X64"]'` ; retour : `gh variable delete CNP_RUNS_ON -R 3-Istor/<app>`. En plan Free, une variable d'org n'est pas lue par les dépôts privés : la variable est par dépôt. Les apps déjà créées devraient quand même être éditées une fois.

## Sécurité

- Le runner exécute le code de toutes les apps générées : chaque push, PR (Dependabot compris) et Dockerfile.
- Runner unique pour l'org et persistant (non `--ephemeral`) : un job peut laisser fichiers ou processus au job suivant, d'un autre projet.
- L'utilisateur `github-runner` est dans le groupe `docker`, donc root sur la VM. Cette VM (projet OpenStack EpiMac) héberge aussi le runner GitLab et un miroir Docker Hub.
- Les secrets du job (`GITHUB_TOKEN` avec `packages: write`, `DISCORD_WEBHOOK_URL`) sont lisibles depuis l'hôte.
- Donc : bascule temporaire, dépôts privés bloqués seulement, retour dès que le quota revient.

## À vérifier sur la VM

`Cloud/terraform/openstack/projects/epimac/runner.tf` installe Docker, git, curl, unzip et jq, pas `yq` (utilisé par `security.yml`) ni explicitement `docker-buildx-plugin` (utilisé par `docker/build-push-action`). Ce fichier enregistre le runner sur `TheGostsniperfr/infra-dockair-sandbox`, pas sur l'org : l'enregistrement d'org n'est pas dans l'IaC.
