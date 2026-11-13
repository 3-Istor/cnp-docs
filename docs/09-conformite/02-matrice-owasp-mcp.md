# Matrice OWASP MCP et agentique

Comment le serveur MCP de la CMP (les assistants IA qui agissent sur CNP au nom d'un
utilisateur, voir [Assistants IA (MCP)](../05-cmp-backend-api/11-cmp-mcp-integration.md))
répond aux deux référentiels OWASP du moment, quel contrôle le porte dans le code, et
quel test le prouve. Un contrôle sans test vert n'est pas revendiqué.

Référentiels :

- **OWASP MCP Top 10**, version 2025, statut bêta (phase 3, « Beta Release and Pilot
  Testing »), CC BY-NC-SA 4.0 :
  <https://owasp.org/projects/mcp-top-10>, consulté le 13/11/2026.
- **OWASP Top 10 for Agentic Applications for 2026**, OWASP GenAI Security Project,
  publié le 9 décembre 2025 :
  <https://genai.owasp.org/resource/owasp-top-10-for-agentic-applications-for-2026/>,
  consulté le 13/11/2026.

Code de référence : CMP au commit
[`31b4486`](https://github.com/3-Istor/CMP/tree/31b4486e33d700e507600fb157ccf03d102c46c0)
(branche `test/mcp-owasp-controls`). Les tests sont dans
[`backend/tests/test_mcp_owasp_controls.py`][tfile] sauf mention contraire, et passent
avec `pytest backend/tests/test_mcp_owasp_controls.py`.

## Écarts trouvés et corrigés

L'écriture des tests a révélé quatre écarts réels. Chacun a un correctif minimal dans
la même PR CMP, et le test correspondant échoue sans lui.

| # | Écart | Correctif | Test |
|---|---|---|---|
| 1 | **Un token d'assistant marchait sur toute l'API REST.** Présenté directement à `/api`, il contournait la politique d'enjeu (`policy.yaml`) et le scope lecture seule, que seuls les outils MCP appliquent. Exemple : une promotion sur un projet `critical` sans approbation, un `DELETE` d'app avec un token en lecture seule | Le token n'est accepté que pendant un appel in-process d'un outil MCP ([`keycloak_service.py:694`][kc694], [`internal_api.py:46`][ia46]) | [`test_an_assistant_token_is_refused_by_the_rest_api_outside_a_tool`][t473], [`test_a_read_only_assistant_cannot_write_through_the_rest_api`][t486] |
| 2 | Un appel refusé par la politique (`deny`) ne laissait aucune trace | Événement `agent.policy.denied` au journal ([`mcp_server.py:289`][m289]) | [`test_a_call_the_stake_policy_refuses_is_journaled`][t379] |
| 3 | `list_app_files` renvoyait les chemins du dépôt sans la mention « données, pas instructions », alors qu'un nom de fichier est écrit par n'importe qui ayant poussé dans le dépôt | Ajout de `note: UNTRUSTED` ([`mcp_server.py:706`][m706]) | [`test_content_written_by_others_is_marked_as_untrusted_data`][t630] |
| 4 | Trois tests existants envoyaient un token d'assistant directement à `/api` | Ils passent désormais par le chemin in-process, le seul possible, et vérifient toujours le refus 403 | `test_agents.py` [L500][a500], [L559][a559], [L659][a659] |

**Changement de comportement** (écart 1) : un client qui appellerait `/api` avec un
token `cnp_at_...` reçoit maintenant 401. Aucun client légitime ne le fait : le
frontend utilise les tokens Keycloak et les assistants passent par `/mcp`.

## OWASP MCP Top 10 (2025)

| Risque | Contrôle CNP | Test qui le prouve | Écart restant |
|---|---|---|---|
| **MCP01:2025** Token Mismanagement & Secret Exposure | Tokens opaques, seul le SHA-256 est stocké ([`store.py:64`][s64]) ; token d'accès 15 min ([`config.py:242`][c242]) ; rotation du refresh token ([`store.py:354`][s354]) ; `push_files` refuse un fichier qui contient un secret ([`mcp_server.py:756`][m756], [`repo.py:99`][r99]) ; `request_secret` fait saisir la valeur dans la CMP, jamais dans la conversation ([`mcp_server.py:825`][m825]) | [`test_assistant_tokens_are_stored_only_as_hashes`][t551], [`test_push_files_containing_a_secret_commits_nothing`][t564], [`test_a_committed_token_is_detected`][a761], [`test_secret_files_are_refused`][a750], [`test_refresh_token_works_once`][a326] | La détection de secrets est une liste de formes connues, pas une mesure d'entropie. Un secret qu'une app écrit dans ses logs part tel quel vers le modèle via `get_app_logs` |
| **MCP02:2025** Privilege Escalation via Scope Creep | Scopes `cnp:read` / `cnp:write` ([`store.py:36`][s36], [`mcp_server.py:149`][m149]) ; le grant ne copie ni groupes ni rôles ([`store.py:45`][s45]) ; politique d'enjeu évaluée côté serveur ([`policy.py:108`][p108]) ; token limité au serveur MCP ([`keycloak_service.py:694`][kc694]) | [`test_read_only_assistant_cannot_create_an_app`][a589], [`test_a_read_only_assistant_cannot_write_through_the_rest_api`][t486], [`test_assistant_claims_never_carry_the_users_platform_groups`][t497], [`test_policy_scales_with_the_project_stake`][a702] | Corrigé dans cette PR : contournement par l'API REST (écart 1). En lecture, l'assistant voit tout ce que voit son utilisateur |
| **MCP03:2025** Tool Poisoning | Outils déclarés dans le code, descriptions = docstrings relues en PR ; instructions du serveur constantes ([`mcp_server.py:57`][m57]) ; aucun outil enregistré à l'exécution | [`test_the_tool_set_is_the_reviewed_one`][t585], [`test_tool_descriptions_are_the_docstrings_in_the_code`][t591], [`test_tool_definitions_do_not_change_with_database_content`][t599] | Les guides (`get_guide`) viennent du dépôt `cnp-agent-kit` à l'exécution ([`skills.py:51`][k51]) et ne portent pas la mention non fiable : leur intégrité repose sur la protection de branche de ce dépôt. Le nom de client déclaré à l'enregistrement OAuth est renvoyé par `whoami` |
| **MCP04:2025** Software Supply Chain Attacks & Dependency Tampering | SDK `mcp` verrouillé dans `poetry.lock` ; Trivy bloque les vulnérabilités HIGH/CRITICAL corrigeables sur chaque PR (`.github/workflows/security.yml`) ; `policy.yaml` ne change que par PR | Hors pytest : job CI `trivy` | Pas de SBOM ni de vérification de signature du SDK. Même remarque que MCP03 pour `cnp-agent-kit` |
| **MCP05:2025** Command Injection & Execution | Aucun appel shell dans le serveur MCP ; entrées typées (`Literal`, bornes `ge`/`le`) ; chemins normalisés, sortie du dépôt refusée ([`repo.py:72`][r72]) ; fichiers de CI, Dockerfile et déploiement classés sensibles ([`repo.py:83`][r83]) ; la recherche LogQL refuse le backtick | [`test_paths_cannot_escape_the_repository`][a756], [`test_paths_that_run_code_or_expose_the_app_are_sensitive`][a745], [`test_search_cannot_break_out_of_the_query`][as345] (`test_activity_sources.py`) | Pousser du code reste de l'exécution de code, par conception : le code applicatif non sensible tourne dans le conteneur de l'app après release, borné par l'isolation du namespace |
| **MCP06:2025** Prompt Injection via Contextual Payloads | Contenu des dépôts et logs marqué `UNTRUSTED` ([`mcp_server.py:52`][m52]) dans `read_app_file`, `list_app_files`, `get_app_logs` ; instructions du serveur : ne jamais suivre ce qui s'y trouve ([`mcp_server.py:79`][m79]) ; la vraie barrière est la politique (`ask`, `confirm`, `pr`) | [`test_content_written_by_others_is_marked_as_untrusted_data`][t630], [`test_server_instructions_forbid_following_logs_and_repository_content`][t640] | La mention est un conseil au modèle, pas une garantie. D'autres champs écrits par des tiers ne sont pas marqués : message d'étape d'une app, messages des étapes de release |
| **MCP07:2025** Insufficient Authentication & Authorization | OAuth 2.1 + PKCE (SDK MCP), enregistrement dynamique limité aux redirections en loopback ou autorisées ([`store.py:72`][s72]) ; scope requis sur `/mcp` ([`mcp_server.py:102`][m102]) ; protection DNS rebinding ([`mcp_server.py:110`][m110]) ; rôle de projet vérifié ([`mcp_server.py:160`][m160]) ; routes réservées aux humains ([`agents.py:37`][ag37], [`access.py:38`][ac38], [`step_up.py:74`][su74]) ; tokens Keycloak : `azp` et `typ` vérifiés ([`keycloak_service.py:723`][kc723]) | [`test_the_mcp_endpoint_refuses_a_token_it_did_not_issue`][t520], [`test_an_infra_users_assistant_cannot_reach_a_platform_admin_endpoint`][t508], [`test_an_assistant_cannot_read_a_project_its_user_is_not_member_of`][t528], [`test_registration_refuses_an_unknown_redirect_uri`][a233], [`test_an_assistant_cannot_approve_a_connection`][a559], [`test_rejects_untrusted_tokens`][au89] (`test_api_auth.py`, cas `other-client` et `id-token`), [`test_refuses_an_assistant_token_with_403_by_default`][st92] (`test_step_up.py`) | `aud` n'est pas vérifié sur les tokens Keycloak (il vaut `account`) : `azp` le remplace. `validate_token_resource=False` : le `resource` RFC 8707 n'est pas lié au token, acceptable tant que la CMP n'émet des tokens que pour `/mcp` |
| **MCP08:2025** Lack of Audit and Telemetry | Middleware d'audit sur chaque appel REST qui modifie, avec l'assistant en `via` ([`audit.py:426`][au426]) ; `record_agent_event` pour les commits et tags poussés hors REST ([`audit.py:343`][au343]) ; refus de la politique journalisés ([`mcp_server.py:289`][m289]) | [`test_every_write_tool_call_is_journaled_with_the_assistant_name`][t358] (6 outils), [`test_every_tool_that_can_change_something_is_in_the_journal_test`][t373], [`test_a_call_the_stake_policy_refuses_is_journaled`][t379] | **Les appels d'outils en lecture ne sont pas journalisés** (`whoami`, `read_app_file`, `get_app_logs`...), ni les arrêts `needs_user_confirmation`. À décider : volume du fil d'activité contre traçabilité, éventuellement dans Loki plutôt que dans `audit_events` |
| **MCP09:2025** Shadow MCP Servers | Un seul serveur MCP, servi par la CMP, OAuth émis par la CMP ; ensemble d'outils figé par test | [`test_the_tool_set_is_the_reviewed_one`][t585] | Hors périmètre de la CMP : les serveurs MCP qu'un développeur lance sur son poste. Relève d'une politique d'usage |
| **MCP10:2025** Context Injection & Over-Sharing | Serveur sans état (`stateless_http=True`, [`mcp_server.py:106`][m106]), aucune mémoire de conversation côté CMP ; chaque appel relit l'utilisateur depuis son token et vérifie le rôle sur le projet | [`test_tool_call_acts_as_the_connected_user`][a307], [`test_an_assistant_cannot_read_a_project_its_user_is_not_member_of`][t528] | `get_app_logs` peut envoyer au modèle externe des données d'utilisateurs finaux de l'app : voir le [registre RGPD](01-registre-rgpd.md), traitement T4 |

## OWASP Top 10 for Agentic Applications 2026

| Risque | Contrôle CNP | Test qui le prouve | Écart restant |
|---|---|---|---|
| **ASI01** Agent Goal Hijack | Comme MCP06, plus la politique qui arrête une action risquée même si le modèle a été détourné | [`test_content_written_by_others_is_marked_as_untrusted_data`][t630], [`test_delete_app_waits_for_the_user_and_deletes_nothing`][t425], [`test_critical_project_waits_for_a_human_approval_in_the_cmp`][a622] | Un modèle détourné peut encore faire tout ce qui est `allow`, par exemple pousser du code non sensible sur un projet `standard` |
| **ASI02** Tool Misuse and Exploitation | Politique par action et par niveau d'enjeu ([`policy.yaml`][py]) ; 3 apps max par sandbox ; quotas 120 lectures et 20 écritures par minute et par grant ([`ratelimit.py:9`][rl9]) ; approbation liée au tag et au commit d'une release | [`test_policy_scales_with_the_project_stake`][a702], [`test_no_write_runs_unattended_on_a_critical_project`][t413], [`test_an_assistant_is_slowed_down_past_its_write_budget`][a838], [`test_cut_release_approval_binds_the_tag_and_the_commit`][a812] | Quotas en mémoire, par réplique : valable tant que la CMP tourne en une réplique |
| **ASI03** Identity and Privilege Abuse | L'assistant agit avec l'identité de l'utilisateur, jamais avec ses groupes plateforme ; token limité au serveur MCP ; routes humaines ; grant révoqué si le compte Keycloak est désactivé ou supprimé | [`test_an_assistant_token_is_refused_by_the_rest_api_outside_a_tool`][t473], [`test_assistant_claims_never_carry_the_users_platform_groups`][t497], [`test_an_infra_users_assistant_cannot_reach_a_platform_admin_endpoint`][t508], [`test_assistant_cannot_approve_its_own_sensitive_action`][a659], [`test_refresh_revokes_the_grant_of_a_disabled_account`][a372] | Corrigé dans cette PR (écart 1) |
| **ASI04** Agentic Supply Chain Vulnerabilities | Comme MCP03 et MCP04 | [`test_the_tool_set_is_the_reviewed_one`][t585], [`test_tool_descriptions_are_the_docstrings_in_the_code`][t591] | Guides de `cnp-agent-kit` lus à l'exécution |
| **ASI05** Unexpected Code Execution (RCE) | Pas d'`eval` ni de shell côté serveur ; un push qui touche la CI, un Dockerfile ou le déploiement suit une règle plus stricte (`deny` en sandbox, `ask` en demo, `pr` au-delà) | [`test_paths_that_run_code_or_expose_the_app_are_sensitive`][a745], [`test_policy_scales_with_the_project_stake`][a702] (cas `code.push_sensitive`) | Voir MCP05 |
| **ASI06** Memory & Context Poisoning | Aucune mémoire longue côté CMP : pas de stockage de conversation, définitions d'outils indépendantes de la base | [`test_tool_definitions_do_not_change_with_database_content`][t599] | Les guides sont mis en cache 10 min ; un guide corrompu dans le dépôt resterait servi jusqu'à correction |
| **ASI07** Insecure Inter-Agent Communication | Sans objet : la CMP n'orchestre pas plusieurs agents. Le seul lien est client MCP ↔ CMP, en HTTPS derrière Cloudflare, avec protection DNS rebinding | Aucun, non revendiqué | À réévaluer si la landing zone IA fait dialoguer des agents |
| **ASI08** Cascading Failures | Quotas par grant ; approbations à usage unique, liées aux arguments exacts, expirées après 30 min ([`policy.py:238`][p238]) ; fenêtre autonome plafonnée à 8 h ([`policy.py:26`][p26]) | [`test_a_cmp_approval_does_not_cover_a_call_with_other_arguments`][t437], [`test_critical_project_waits_for_a_human_approval_in_the_cmp`][a622] (rejeu refusé), [`test_an_assistant_is_slowed_down_past_its_write_budget`][a838], [`test_expired_autonomous_window_restores_the_standard_rules`][a722] | Pas de coupe-circuit global sur l'ensemble des assistants |
| **ASI09** Human-Agent Trust Exploitation | Actions jamais autonomes : `project.create` et `app.delete` ne sont `allow` à aucun niveau, même pendant un incident ; résumé de l'action construit par le serveur à partir des arguments, pas par le modèle ; `confirm` passe par une page de la CMP que l'assistant ne peut pas valider | [`test_never_autonomous_actions_need_a_human_at_every_stake_level`][t403], [`test_delete_app_waits_for_the_user_and_deletes_nothing`][t425], [`test_assistant_cannot_approve_its_own_sensitive_action`][a659] | **`ask` est déclaratif** : `user_confirmed=true` est posé par le modèle, la CMP ne peut pas vérifier que l'utilisateur a vraiment dit oui. Seul `confirm` est une vérification humaine hors bande. Le step-up par passkey est désactivé par défaut (`STEP_UP_ENABLED=False`) et les routes de release et de suppression laissent passer les assistants (`agents_allowed=True`) |
| **ASI10** Rogue Agents | Révocation d'un grant ou de tous les grants d'un utilisateur par `infra` ; journal avec `via` ; refus journalisés ; fenêtre autonome qui expire seule | [`test_revoked_assistant_is_refused_on_its_next_call`][a340], [`test_infra_revokes_every_grant_of_a_user`][a533], [`test_a_call_the_stake_policy_refuses_is_journaled`][t379], [`test_expired_autonomous_window_restores_the_standard_rules`][a722] | Pas d'alerte automatique sur une rafale de `agent.policy.denied` |

## Décisions à prendre

- [ ] Journaliser ou non les appels d'outils en lecture (MCP08), et où.
- [ ] Passer `app.delete` et `project.create` de `ask` à `confirm` sur sandbox et demo,
      puisque `ask` repose sur la parole du modèle (ASI09).
- [ ] Marquer comme non fiables les autres champs écrits par des tiers (MCP06).
- [ ] Alerte Loki ou Alertmanager sur `agent.policy.denied` (ASI10).
- [ ] Politique d'usage : assistants autorisés, serveurs MCP locaux (MCP09), envoi de
      logs contenant des données d'utilisateurs finaux (MCP10).

[tfile]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/tests/test_mcp_owasp_controls.py
[t358]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/tests/test_mcp_owasp_controls.py#L358
[t373]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/tests/test_mcp_owasp_controls.py#L373
[t379]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/tests/test_mcp_owasp_controls.py#L379
[t403]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/tests/test_mcp_owasp_controls.py#L403
[t413]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/tests/test_mcp_owasp_controls.py#L413
[t425]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/tests/test_mcp_owasp_controls.py#L425
[t437]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/tests/test_mcp_owasp_controls.py#L437
[t473]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/tests/test_mcp_owasp_controls.py#L473
[t486]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/tests/test_mcp_owasp_controls.py#L486
[t497]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/tests/test_mcp_owasp_controls.py#L497
[t508]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/tests/test_mcp_owasp_controls.py#L508
[t520]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/tests/test_mcp_owasp_controls.py#L520
[t528]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/tests/test_mcp_owasp_controls.py#L528
[t551]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/tests/test_mcp_owasp_controls.py#L551
[t564]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/tests/test_mcp_owasp_controls.py#L564
[t585]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/tests/test_mcp_owasp_controls.py#L585
[t591]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/tests/test_mcp_owasp_controls.py#L591
[t599]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/tests/test_mcp_owasp_controls.py#L599
[t630]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/tests/test_mcp_owasp_controls.py#L630
[t640]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/tests/test_mcp_owasp_controls.py#L640
[a233]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/tests/test_agents.py#L233
[a307]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/tests/test_agents.py#L307
[a326]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/tests/test_agents.py#L326
[a340]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/tests/test_agents.py#L340
[a372]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/tests/test_agents.py#L372
[a500]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/tests/test_agents.py#L500
[a533]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/tests/test_agents.py#L533
[a559]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/tests/test_agents.py#L559
[a589]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/tests/test_agents.py#L589
[a622]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/tests/test_agents.py#L622
[a659]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/tests/test_agents.py#L659
[a702]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/tests/test_agents.py#L702
[a722]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/tests/test_agents.py#L722
[a745]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/tests/test_agents.py#L745
[a750]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/tests/test_agents.py#L750
[a756]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/tests/test_agents.py#L756
[a761]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/tests/test_agents.py#L761
[a812]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/tests/test_agents.py#L812
[a838]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/tests/test_agents.py#L838
[as345]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/tests/test_activity_sources.py#L345
[au89]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/tests/test_api_auth.py#L89
[st92]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/tests/test_step_up.py#L92
[m52]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/app/mcp_server.py#L52
[m57]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/app/mcp_server.py#L57
[m79]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/app/mcp_server.py#L79
[m102]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/app/mcp_server.py#L102
[m106]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/app/mcp_server.py#L106
[m110]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/app/mcp_server.py#L110
[m149]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/app/mcp_server.py#L149
[m160]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/app/mcp_server.py#L160
[m289]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/app/mcp_server.py#L289
[m706]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/app/mcp_server.py#L706
[m756]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/app/mcp_server.py#L756
[m825]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/app/mcp_server.py#L825
[s36]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/app/services/agents/store.py#L36
[s45]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/app/services/agents/store.py#L45
[s64]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/app/services/agents/store.py#L64
[s72]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/app/services/agents/store.py#L72
[s354]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/app/services/agents/store.py#L354
[p26]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/app/services/agents/policy.py#L26
[p108]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/app/services/agents/policy.py#L108
[p238]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/app/services/agents/policy.py#L238
[py]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/app/services/agents/policy.yaml
[r72]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/app/services/agents/repo.py#L72
[r83]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/app/services/agents/repo.py#L83
[r99]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/app/services/agents/repo.py#L99
[rl9]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/app/services/agents/ratelimit.py#L9
[k51]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/app/services/agents/skills.py#L51
[ia46]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/app/services/agents/internal_api.py#L46
[kc694]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/app/services/keycloak_service.py#L694
[kc723]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/app/services/keycloak_service.py#L723
[c242]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/app/core/config.py#L242
[au343]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/app/services/audit.py#L343
[au426]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/app/services/audit.py#L426
[ag37]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/app/routers/agents.py#L37
[ac38]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/app/routers/access.py#L38
[su74]: https://github.com/3-Istor/CMP/blob/31b4486e33d700e507600fb157ccf03d102c46c0/backend/app/services/step_up.py#L74
