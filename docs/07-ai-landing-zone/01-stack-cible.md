# Stack cible et compromis

Le schéma reprend la vue en couches de la landing zone IA, avec la techno que CNP retient à
chaque étage. Deux couleurs signalent un compromis :

- **orange** : un outil déjà présent est réutilisé pour un usage pour lequel il n'est pas fait ;
- **rouge** : aucun outil, par choix.

Le critère qui tranche est celui de l'architecture cible : environ 4 Go de mémoire pour toute la
zone IA sur le cluster on-prem. Un outil léger ou utile pour apprendre est pris ; un outil qui
consomme beaucoup est écarté quand la plateforme a déjà de quoi couvrir le besoin.

![Stack cible de la landing zone IA](assets/ai-target-stack.png)

Source du schéma : `assets/ai-target-stack.html` (rendu en PNG avec Chromium).

## Deux changements par rapport à l'architecture cible

- **Identité des agents (D3).** Keycloak 26.6 accepte le token du ServiceAccount Kubernetes à la
  place du secret client (*federated client authentication*, passée en « supported »). Le pod
  de l'agent ne porte donc plus aucun secret, sans ajouter de composant.
- **Métriques.** La base de métriques de CNP est VictoriaMetrics. Les vues qui dessinent
  « Prometheus » pour les métriques GenAI doivent lire VictoriaMetrics.

## Les compromis

| # | Étage | Choix | Pourquoi | On gagne | On perd | Pour aller plus loin |
|---|---|---|---|---|---|---|
| 1 | Portail | Agent = chart Helm + `SandboxTemplate`, créé par la CMP. Pas de CRD `Agent`. | Un CRD demande un opérateur à écrire et maintenir ; la CMP et Argo CD savent déjà créer une app. | Aucun code d'opérateur ; même chemin GitOps que les apps. | Pas d'objet unique « Agent » avec un statut : sa config est répartie entre values Helm, client Keycloak et Vault. Lister les agents passe par la CMP. | Un CRD `Agent` léger, réconcilié par la CMP, si le nombre d'agents grandit. |
| 2 | Gateway, autres agents | Pas de proxy A2A ni d'Agent Cards. | Aucun cas d'usage multi-agents. Le support A2A d'Agent Router n'est pas confirmé ; le filtre `a2a` d'Envoy n'apparaît que dans la doc de développement. | Pas de surface d'attaque ni de config pour un besoin absent. | Deux agents ne peuvent pas se parler par un chemin contrôlé ; s'ils le font, c'est hors gateway. | agentgateway (A2A natif), ou le filtre Envoy quand il sera stable. Bon sujet pour s'amuser une fois deux agents en place. |
| 3 | Gateway | Guardrails : Llama Guard servi par Ollama sur le GPU du homelab, appelé en `ext_proc`. | Un classifieur sur CPU dans le cluster mangerait le budget mémoire. | Zéro mémoire sur le cluster ; classification rapide sur GPU. | Dépend du homelab et de WireGuard ; partage le GPU avec l'inférence, donc de la latence. Il faut choisir : bloquer (fail-closed) ou laisser passer si le homelab tombe. | Fail-closed pour les projets sensibles, fail-open avec alerte pour les autres. |
| 4 | Gateway | Masquage PII et secrets : aucun outil. | Jamais conçu ; un service de détection de PII (Python et modèles NLP) est lourd pour le budget. | Rien à faire tourner. | Un secret ou une donnée personnelle dans un prompt peut partir chez Gemini, dont le free tier peut utiliser les données. Llama Guard classe le contenu, il ne masque rien. | Court terme : filtre regex des secrets dans l'`ext_proc` (léger). Plus tard : un service type Presidio sur le GPU. D'ici là, les projets *Local* restent sur Ollama. |
| 5 | Gateway | Audit : access logs de la gateway et décisions OPA → Alloy → Loki, puis l'archive Object Lock existante. | La chaîne d'audit de la plateforme existe déjà (CronJob `audit-archive`, bucket RGW en COMPLIANCE 30 jours). | Aucun composant de plus ; mêmes requêtes LogQL, mêmes alertes, même archive. | Loki est un stockage de logs, pas un outil d'audit : pas de schéma imposé, recherche par agent ou par outil seulement si les champs sont bien posés. 14 jours à chaud par défaut. | Fixer les champs (utilisateur, agent, projet, modèle, outil, tokens, décision), passer `ai-gateway` à 30 jours, ajouter la source à l'export de l'archive. |
| 6 | Modèles | Ollama sur la RTX 5080 du homelab, pas vLLM. | Un seul GPU de 16 Go, un usage de labo. | Un binaire, beaucoup de modèles, changement de modèle à chaud, API compatible OpenAI. | Moins de débit sous charge concurrente que vLLM, moins de réglages de production ; l'Inference Extension ne s'applique pas. | vLLM si plusieurs GPU ou beaucoup d'appels simultanés. |
| 7 | Identité | Pas de SPIRE / SPIFFE. | Avec la federated client authentication, Keycloak donne déjà une identité de workload sans secret. SPIRE ajoute un serveur et un agent par nœud. | Une brique de moins ; le pod n'a plus de secret. | Pas de SVID pour du mTLS entre workloads ; l'identité reste liée à Kubernetes, ce qui complique la fédération avec EKS et GKE. Le support SPIFFE de Keycloak est encore en preview. | SPIRE quand le mTLS entre agents ou la fédération multicloud deviennent utiles. Léger : bon candidat pour s'amuser. |
| 8 | Observabilité | Traces LLM dans Tempo et Grafana, pas Langfuse. | Langfuse auto-hébergé demande au moins ~21,5 Gio de RAM (web, worker, ClickHouse, Postgres, Redis) : hors budget. | Une seule stack d'observabilité, déjà opérée. | Pas d'interface pensée pour les LLM : pas de lecture confortable des prompts et réponses par session, pas d'annotation ni de score, pas de gestion de prompts. | Arize Phoenix : un conteneur, OTel natif. Arize fait tourner ses instances hébergées avec 2 Go de RAM et 1 CPU ; ici il s'appuierait sur un Postgres CloudNativePG. Les évals tournent avec promptfoo dans GitHub Actions, sans rien sur le cluster (MIT ; rachat par OpenAI annoncé en mars 2026, à surveiller). |
| 9 | FinOps | Budgets : quotas de tokens au rate limit de la gateway et OpenCost. | Rien de dédié au FinOps LLM ; la gateway compte déjà les tokens. | Le plafond est appliqué à la gateway, pas seulement constaté. | Le rate limit compte des tokens par fenêtre, pas des euros ; OpenCost chiffre l'infra, pas les appels. La table de prix est à tenir à la main. | Quota par agent en plus du quota par projet ; coût en euros calculé dans Grafana (tokens × prix) ; alerte à 80 %. |
| 10 | FinOps | Plafond d'étapes : limite d'itérations du framework (par exemple `recursion_limit` de LangGraph), posée par le kit d'agent. | La gateway ne voit que des appels, pas les étapes d'un agent. | Zéro composant. | Rien ne l'impose côté plateforme : un agent au code maison peut l'ignorer. | Le quota de tokens par agent sert de plafond dur ; une policy de revue du kit d'agent vérifie la limite. |

## Sources

- [Keycloak 26.6.0 released](https://www.keycloak.org/2026/04/keycloak-2660-released) et [Federated client authentication](https://www.keycloak.org/2026/01/federated-client-authentication)
- [Envoy, filtre a2a](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/a2a_filter) et [FAQ agentgateway](https://agentgateway.dev/docs/faqs)
- [ToolHive, registre personnalisé](https://docs.stacklok.com/toolhive/guides-ui/registry) et [vérification des serveurs MCP](https://stacklok.com/blog/from-unknown-to-verified-solving-the-mcp-server-trust-problem/)
- [Arize Phoenix, dimensionnement en auto-hébergement](https://community.arize.com/x/phoenix-support/5b5uwoj3d0mw/recommended-specs-for-self-hosting-ecs-with-postgr)
- [Frameworks d'évaluation open source en 2026](https://futureagi.com/blog/best-open-source-eval-frameworks-2026)
- Chaîne d'audit existante : `docs/01-architecture/06-audit-logging.md`
