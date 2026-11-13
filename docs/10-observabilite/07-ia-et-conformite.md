# IA : observabilité des agents et traçabilité

Ce que la loi demande, ce que les autorités françaises recommandent, et ce que la CNP met en place
pour la landing zone IA (gateway, agents hébergés, outils MCP, serveur MCP du portail). **Ceci est
une cartographie factuelle, pas un avis juridique.** Les articles cités sont à revérifier sur
EUR-Lex avant toute affirmation de conformité.

![Trace d'une tâche d'agent, de la demande au déploiement](../01-architecture/diagrams/observability-ai-run.light.png#only-light)
![Trace d'une tâche d'agent, de la demande au déploiement](../01-architecture/diagrams/observability-ai-run.dark.png#only-dark)

## Ce que dit le droit en novembre 2026

| Texte | Ce qu'il impose | S'applique à la CNP ? |
| --- | --- | --- |
| AI Act, art. 4 | Favoriser la culture IA des personnes qui utilisent les systèmes (allégé par l'Omnibus) | Oui, déjà |
| AI Act, art. 50(1) | Dire à une personne qu'elle parle à une IA | Oui, pour les assistants et agents de la CNP (depuis le 02/08/2026) |
| AI Act, art. 12 | Journalisation automatique des événements | Systèmes **à haut risque** seulement |
| AI Act, art. 19 et 26(6) | Garder ces journaux **au moins 6 mois**, sous réserve du droit des données personnelles | Haut risque seulement |
| AI Act, art. 26 | Surveillance humaine, suivi du fonctionnement par le déployeur | Haut risque seulement |
| AI Act, art. 73 | Signaler un incident grave en 2, 10 ou 15 jours | Haut risque seulement |
| Digital Omnibus, Règl. (UE) 2026/1744 | Reporte le haut risque de l'annexe III au **02/12/2027** | Calendrier |
| RGPD + recommandation CNIL journalisation (2021) | Journaux qui identifient l'auteur, l'horodatage, l'opération ; séparés, en écriture seule ; 6 à 12 mois ; minimiser les données personnelles | **Oui**, pour tous nos journaux |
| CNIL / CIANum, IA agentique (juillet 2026) | Traçabilité par tâche, validation humaine des actions critiques, mémoire cloisonnée, bac à sable, arrêt d'urgence | Recommandation |
| ANSSI, sécuriser un système d'IA agentique v1.0 (octobre 2026) | Approbation hors de portée de l'agent ; journaux hors de portée de l'agent, qui permettent de reconstruire demande, autorisation, action, résultat | Recommandation |
| NIS2 (France) | Loi de transposition en discussion, non en vigueur au 13/11 | Non |

**Lecture pour la CNP :**

- La CNP **déploie** des modèles tiers (Claude, Gemini, un modèle local) et **fournit** ses propres
  agents et assistants, pour un usage interne d'outillage. Ce n'est pas du haut risque : l'AI Act ne
  lui impose pas de journalisation détaillée aujourd'hui.
- Un **projet tenant** qui ferait du recrutement ou de la notation d'étudiants (annexe III) serait
  lui-même soumis aux art. 12, 19 et 26. La CNP doit alors **lui donner les moyens** de s'y
  conformer : c'est un argument d'offre fort.
- Les recommandations ANSSI et CNIL de 2026 décrivent presque mot pour mot ce qu'il faut tracer pour
  un agent qui agit. On les suit : c'est l'état de l'art et c'est démontrable.
- La tension « tout journaliser » contre « minimiser » se règle ainsi : **métadonnées et décisions
  toujours, contenu des prompts jamais par défaut**, en opt-in par projet, expurgé et à courte durée.

## Ce qu'on enregistre pour chaque tâche d'agent

Une tâche d'agent = **une trace** OpenTelemetry, et **un enregistrement d'audit** signé.

| Étape | Span | Champs clés |
| --- | --- | --- |
| Demande dans le portail ou via MCP | `cmp.request` | utilisateur (sub Keycloak), projet, agent, client MCP (`azp`) |
| Exécution de l'agent | `invoke_agent` | agent, version de ses consignes, modèle par défaut |
| Appel au modèle | `chat` (gateway) | modèle, fournisseur, tokens entrée et sortie, coût, latence, décision de budget, verdict des garde-fous |
| Appel d'outil | `execute_tool` / `mcp.method.name` | outil, serveur MCP, **hash** des arguments, décision de politique |
| Demande d'approbation | `cnp.approval` | action demandée, hash de l'action, approbateur, décision, délai |
| Exécution en bac à sable | `cnp.sandbox.exec` | image, durée, code de sortie, ressources |
| Effet sur la plateforme | `cnp.deploy` | app, version, commit, release, lien Argo CD |

Les conventions GenAI d'OpenTelemetry sont encore au statut *Development* : les noms changent
(`gen_ai.client.token.usage` vient d'être remplacé dans la spec). On fige les noms dans une couche
de traduction du collecteur, et les champs sans convention (approbation, garde-fous, identité) sont
dans un espace `cnp.*`.

## Où ça va

| Donnée | Base | Rétention | Lecteurs |
| --- | --- | --- | --- |
| Traces d'agents (opérationnel) | VictoriaTraces | 7 jours | membre du projet de l'agent (portail), infra |
| Tokens, coût, latence, refus par projet, agent, modèle | VMSingle | 30 jours, historique FinOps dans Postgres | finops, membres, infra |
| Journaux d'accès de la gateway, décisions de politique | VictoriaLogs | 30 jours | infra, security-auditor |
| **Enregistrement d'audit par tâche** (JSON signé) | **Archive Object Lock** (RGW, mode COMPLIANCE) | **180 jours** | security-auditor, infra |
| Contenu des prompts et réponses | VictoriaLogs, flux séparé | 7 jours, opt-in projet, expurgé | membres du projet qui l'ont activé |

**Chemin d'audit séparé et non échantillonné.** Les traces opérationnelles peuvent être
échantillonnées pour économiser de la place ; l'enregistrement d'audit, lui, est complet. Un petit
job Go lit chaque trace terminée, en tire l'enregistrement (métadonnées seulement), le **signe avec
Vault transit**, chaîne son hash au précédent, et l'écrit dans le bucket Object Lock. La tête de
chaîne du jour est archivée elle aussi : une suppression ou une modification se voit.
Les agents n'ont aucun droit sur ce bucket ni sur Vault transit : leurs journaux sont hors de leur
portée, comme le demande l'ANSSI.

180 jours couvrent le plancher de 6 mois de l'AI Act pour le jour où un tenant serait à haut risque,
et restent dans la fourchette de la CNIL.

## Ce que ça donne en démo

1. Un utilisateur demande à l'assistant du portail de corriger une app en erreur.
2. Dans le portail, onglet Traces de l'agent : la cascade complète, demande, raisonnement en appels
   de modèle, lecture des logs via l'outil MCP, proposition de correctif, **approbation de
   l'utilisateur**, release, rollout.
3. Coût de la tâche en euros et en tokens, comparé au budget de l'agent.
4. Bouton **« Exporter l'enregistrement d'audit »** : le JSON signé, sa vérification de signature et
   sa position dans la chaîne.
5. Côté security-auditor : la liste de toutes les actions d'agents de la semaine, refus compris.

Libellé honnête : « enregistrement au format des art. 12, 19 et 26(6), sans prétention de
conformité ».

## Changements dans la spec de la landing zone IA

| Décision de la spec | Aujourd'hui | Devient |
| --- | --- | --- |
| D7 | OpenTelemetry → Alloy → Tempo | OpenTelemetry → collecteur → VictoriaTraces |
| D14 | Logs d'accès → Alloy → Loki, archive 30 jours | Logs d'accès → vlagent → VictoriaLogs, enregistrement signé en archive 180 jours |
| D24 | Traces 7 jours, logs 30 jours, prompts jamais | Inchangé, plus l'opt-in contenu à 7 jours |
| D30 | Tempo monolithique | Supprimé, Tempo reste le repli |
| Hors périmètre « conformité AI Act » | | Entre dans le périmètre, au niveau décrit ici |

## Outils écartés

| Outil | Pourquoi pas |
| --- | --- |
| Langfuse | Postgres, ClickHouse, Redis et S3 : trop lourd. Option par tenant si un projet demande gestion de prompts et jeux de données |
| Phoenix | Licence Elastic 2.0 |
| Helicone | En maintenance depuis son rachat |
| OpenLIT | Bon candidat (Apache-2.0, intégration documentée avec VictoriaMetrics), mais agentgateway et kagent émettent déjà les spans : pas besoin d'un SDK de plus |
| Rekor / Sigstore pour le journal | Lourd pour notre échelle ; Vault transit + Object Lock suffisent |
