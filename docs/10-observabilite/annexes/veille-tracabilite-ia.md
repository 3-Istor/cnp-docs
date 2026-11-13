# AI traceability for CNP: legal map (EU/France) and VictoriaStack implementation

Research date: 2026-11-13. This is a factual map, not legal advice.
Convention: "LAW SAYS" = binding text. "GUIDANCE" = non-binding official text. "BEST PRACTICE" = my synthesis or practitioner advice.
Confidence: [HIGH] primary source read this session. [MED] secondary source or primary source only partly read. [LOW] inferred or unverified.
Stale flag: anything older than 12 months (before 2025-11-13) is marked [STALE - date].

Access caveat: eur-lex.europa.eu pages do not render through my fetcher (JS-only). For AI Act article text I used the European Commission "AI Act Service Desk" article pages, which state they follow the EUR-Lex consolidated version of 2026-07-27 and flag Omnibus changes with NEW/AMENDED markers. The EUR-Lex consolidated text exists at CELEX 02024R1689-20260727 (seen in search results, not rendered). Treat EUR-Lex as the binding source and re-check anything load-bearing there.

---

## PART A. LEGAL

### A1. EU AI Act (Reg. 2024/1689) as amended by the Digital Omnibus on AI

**Omnibus status in Nov 2026: adopted, in force, high-risk postponed.** [HIGH on dates/content, MED on OJ publication day]
- Regulation (EU) 2026/1744, dated 2026-07-08, amends Reg. 2024/1689, 2018/1139 and 2023/1230. ELI: http://data.europa.eu/eli/reg/2026/1744/oj. Search result metadata from eur-lex.europa.eu/eli/reg/2026/1744 confirms title and date. Publication in the OJ on 2026-07-24 and entry into force 2026-07-27 come from law-firm sources [MED].
- Political deal 2026-05-07, Parliament vote 2026-06-16, Council adoption 2026-06-29 (secondary sources) [MED].
- Dates per the Commission's AI Act Service Desk timeline (accessed 2026-11-13) [HIGH]:

| Date | What applies | Changed by Omnibus |
|---|---|---|
| 2025-02-02 | Chapters I-II: definitions, Art. 4 AI literacy, Art. 5 prohibitions | No |
| 2025-08-02 | GPAI provider obligations (Art. 53-55), governance, national authorities and penalties | No |
| 2026-08-02 | Most rules incl. Art. 50 transparency, Commission enforcement powers on GPAI | No |
| 2026-12-02 | New Art. 5 prohibitions (non-consensual intimate deepfakes, CSAM). Transition for Art. 50(2) marking for generative systems already on the market before 2026-08-02 (Art. 111(4)) | Yes |
| 2027-08-02 | Sandboxes; legacy GPAI models placed before 2025-08-02 must comply | Sandbox date yes |
| 2027-12-02 | High-risk Annex III (Art. 6(2)) rules | Yes (was 2026-08-02) |
| 2028-08-02 | High-risk Annex I products (Art. 6(1)) | Yes (was 2027-08-02) |

  Sources: https://ai-act-service-desk.ec.europa.eu/en/ai-act/timeline/timeline-implementation-eu-ai-act and https://artificialintelligenceact.eu/ai-act-explorer/digital-omnibus/ (the latter is an unofficial but text-verified tracker) [HIGH/MED].
- Art. 4 AI literacy was rewritten: providers and deployers "shall take measures to support the development of AI literacy", no specific individual level required (Service Desk page, all 3 paragraphs marked AMENDED). Still in force since 2025-02-02. [HIGH]
- Art. 12, 19, 21, 26, 72(1) text is NOT amended by the Omnibus (no AMENDED markers on the Service Desk pages; Omnibus summary page says Arts. 12/19/26 are untouched). [HIGH]
- Art. 25 was amended (value chain). Art. 27 FRIA may reuse a GDPR DPIA. [HIGH]
- Art. 72: Service Desk page says the Commission template/guidance for the post-market monitoring plan is due by 2027-09-02. Not cross-checked on EUR-Lex. [MED]
- Not found: whether the 2027-12-02 date can slip again. Annex III content itself unchanged on the page I read.

**Article-by-article (LAW SAYS)**

| Article | What it says | Applies to | URL |
|---|---|---|---|
| Art. 12(1)-(2) | High-risk systems must technically allow automatic recording of events (logs) over the system lifetime, at a level enabling traceability for: (a) identifying risk situations or substantial modification, (b) post-market monitoring (Art. 72), (c) monitoring of operation by deployers (Art. 26(5)). Art. 12(3) lists minimum fields only for remote biometric identification: start/end time of each use, reference database, input data that led to a match, identity of persons verifying results. | Provider design duty, high-risk only | https://ai-act-service-desk.ec.europa.eu/en/ai-act/article-12 |
| Art. 19 | Providers keep Art. 12(1) logs under their control for a period appropriate to the intended purpose, at least 6 months, "without prejudice to applicable Union or national law" (data protection law named). | High-risk providers | https://ai-act-service-desk.ec.europa.eu/en/ai-act/article-19 |
| Art. 21(2) | On reasoned request, provider gives the competent authority access to Art. 12(1) logs under its control. Confidentiality per Art. 78. | High-risk providers | https://ai-act-service-desk.ec.europa.eu/en/ai-act/article-21 |
| Art. 26(1)-(2) | Deployer uses system per instructions; assigns human oversight to persons with competence, training, authority, support. | High-risk deployers | https://ai-act-service-desk.ec.europa.eu/en/ai-act/article-26 |
| Art. 26(5) | Deployer monitors operation on basis of instructions for use; if risk, informs provider/distributor and market surveillance authority without undue delay and suspends use; serious incidents reported to provider first, then importer/distributor and authorities. | High-risk deployers | same |
| Art. 26(6) | Deployer keeps automatically generated logs under its control at least 6 months, unless other Union/national law (esp. data protection) provides otherwise. | High-risk deployers | same |
| Art. 26(7) | Employer-deployers inform workers' representatives and affected workers before using a high-risk system at work. | High-risk deployers | same |
| Art. 26(11) | Deployers of Annex III systems that decide or assist decisions on natural persons inform them they are subject to a high-risk system. | High-risk deployers | same |
| Art. 50(1) | Providers design systems that interact directly with natural persons so people are informed they interact with AI, unless obvious to a reasonably well-informed person. | Any AI system provider, in force since 2026-08-02 | https://ai-act-service-desk.ec.europa.eu/en/ai-act/article-50 |
| Art. 50(2) | Providers of systems (including GPAI systems) generating synthetic audio, image, video or text mark outputs in machine-readable, detectable format. Exempt: assistive function for standard editing or no substantial alteration of input. Transition to 2026-12-02 for systems on market before 2026-08-02. | Providers of generative systems | same |
| Art. 50(4) | Deployers disclose deepfakes and AI text published to inform the public on matters of public interest, unless human review and editorial responsibility. | Deployers publishing content | same |
| Art. 53, 55 | GPAI model providers: technical documentation, downstream information (53(1)(b)), copyright policy, training-data summary; systemic-risk models add evaluation, incident reporting to AI Office (55(1)(c)). | GPAI model providers (Anthropic, Google, Meta, Mistral...), not downstream deployers | n/a |
| Art. 72 | Providers of high-risk systems run documented post-market monitoring collecting performance data over lifetime (deployers' data included). | High-risk providers | https://ai-act-service-desk.ec.europa.eu/en/ai-act/article-72 |
| Art. 73 | Providers of high-risk systems report serious incidents to market surveillance authority of the Member State: general case at most 15 days after awareness; widespread infringement or Art. 3(49)(b) at most 2 days; death at most 10 days; incomplete initial report allowed. | High-risk providers | https://ai-act-service-desk.ec.europa.eu/en/ai-act/article-73 |
| Art. 25 | A deployer or third party becomes provider of a high-risk system if it rebrands it, substantially modifies it, or changes the intended purpose (including of a GPAI system) so that it becomes high-risk. 25(4): written agreement between high-risk provider and third-party suppliers of tools/models/components, FOSS tools excluded, GPAI models not covered by that exemption. | Tenant app builders and the platform as supplier | https://ai-act-service-desk.ec.europa.eu/en/ai-act/article-25 |
| Art. 3(3), 3(4), 3(11) | Provider includes whoever develops a system and puts it into service under its own name, "putting into service" includes in-house first use. Deployer uses a system under its authority. | Role qualification | https://ai-act-service-desk.ec.europa.eu/en/ai-act/article-3 |

All rows [HIGH] except where noted, with the caveat that the service desk summaries paraphrase; verbatim text is on EUR-Lex.

**What applies to CNP today (my reading, not legal advice)** [LOW to MED, role qualification is a legal judgment]
- LAW SAYS: the AI Act has no general logging duty for non-high-risk systems. Art. 12/19/26(6)/72/73 are scoped to high-risk systems, which apply from 2027-12-02 at the earliest for Annex III.
- CNP as a platform deploying Claude/Gemini/Ollama models: the model providers carry Art. 53/55. CNP becomes a GPAI model provider only if it substantially modifies a model. The Commission guidelines (2025-07-18, secondary summaries) use a one-third-of-original-training-compute test. Primary text not read in full. [MED]
- CNP-built assistants and hosted agents (portal MCP server, kagent agents): CNP is probably the "provider" of those AI systems, since in-house putting into service counts (Art. 3(3), 3(11)). They are most likely not high-risk. Duties then: Art. 4 (support literacy), Art. 5 (prohibitions), Art. 50(1) disclosure unless obvious, possibly Art. 50(2) marking of generated text if the exemption for assistive editing does not fit. Whether Art. 50(2) bites on internal dev agents is an open reading. [LOW]
- Tenant app in Annex III (4(a) recruiting/CV screening, 4(b) performance monitoring, 3(a)-(d) education admission/grading/proctoring): the tenant is provider or deployer and carries Art. 12/19/26/72/73. CNP is at most a supplier under Art. 25(4) and gets a contractual duty to give information and technical access. Annex III list: https://ai-act-service-desk.ec.europa.eu/en/ai-act/annex-3 [HIGH]. Art. 6(3) derogation not checked.
- Enforcement authority in France: designation still unsettled. A DGE/DGCCRF scheme was published Sept 2025, a Senate vote on the DDADUE bill on 2026-02-18 was reported, final adoption not found. Sources conflict on CNIL's role. [LOW]. https://www.entreprises.gouv.fr/priorites-et-actions/transition-numerique/soutenir-le-developpement-de-lia-au-service-de-0

**Harmonised standards and logging** [MED]
- CEN-CENELEC JTC 21 is working on 10 Commission-requested areas, including record keeping, human oversight, transparency, cybersecurity. Source: https://digital-strategy.ec.europa.eu/en/policies/ai-act-standardisation [HIGH].
- The first one to reach the OJ is EN 18286 (quality management, Art. 17), via Implementing Decision (EU) 2026/2274, per the same Commission page. Not about logging. [HIGH on Commission page; the 2026-10-14 OJ date is from a secondary source]
- prEN ISO/IEC 24970 "AI system logging": ISO/IEC SC 42 DIS 24970:2025 adopted as CEN draft. CEN public enquiry 2025-11-28 to 2026-01-25 (Lithuanian tracker). DIN lists a follow-on FprEN ISO/IEC 24970:2026 under CEN/CLC/JTC 21/WG 3. ISO page shows "Under publication". I could not confirm final publication nor whether it is cited in the OJ as a harmonised standard. [MED]. Sources: https://standards.iteh.ai/catalog/standards/cen/38472bd6-ab5c-4908-ac64-6990f45027e6/pren-iso-iec-24970 , https://www.iso.org/standard/88723.html (403 to my fetcher, status taken from search snippet).
- ISO/IEC 42001:2023 Annex A control A.6.2.8 (AI system recording of event logs): organisation decides in which lifecycle phases logging is enabled, minimum when in use. Control text is paywalled, wording from vendor sites. Claims of mandatory integrity or immutability are vendor additions. [MED/LOW]
- Practical consequence: no harmonised logging standard gives presumption of conformity yet. Anything built now is a reasonable design, not a presumption of conformity.

### A2. GDPR and CNIL

**LAW SAYS (GDPR, Reg. 2016/679; ELI https://eur-lex.europa.eu/eli/reg/2016/679/oj, not re-fetched this session, article numbers from standard knowledge) [MED]**
- Art. 5(1)(c) minimisation, 5(1)(e) storage limitation, 5(1)(f) integrity and confidentiality, 5(2) accountability, Art. 25 by design, Art. 32 security, Art. 30 register, Art. 35 DPIA, Art. 15 access, Art. 22 automated decisions, Art. 33-34 breach notification (72 h to supervisory authority).
- Prompts and agent traces routinely contain personal data (names, emails, code with data, tickets). So prompt/response logs are personal data processing with CNP or the tenant as controller. A controller needs a purpose and a legal basis for the log processing itself.

**CNIL logging recommendation: deliberation no. 2021-122 of 2021-10-14 [HIGH, read in full] [STALE - 2021-10-14, still the reference]**
PDF: https://www.cnil.fr/sites/cnil/files/atoms/files/recommandation_-_journalisation.pdf . Legifrance JORFTEXT000044272396.
- Scope: application logs, not perimeter logs (para. 2). Non-prescriptive (para. 1).
- Para. 6: record creation, consultation, modification, deletion with individually identified author, timestamp, nature of operation, reference of the data concerned. "Avoid duplicating the processed data within the logs."
- Para. 7: store segregated from the main system, write-only for the applications, no overwrite; access rights to logs based on strict necessity.
- Para. 8: retain 6 months to 1 year. Para. 9: purpose is mainly active security (alerts, detection), formalise alert processing. Para. 10: ex post breach investigation.
- Para. 11: reuse for another purpose may be a purpose diversion; use a usage charter or alert on modification access.
- Para. 12: minimise personal data in logs; ensure confidentiality, availability, integrity; recommends to timestamp and sign logs at creation.
- Para. 13: inform authorised users of logging, data and duration. Data subject rights on logs are exercised with the controller of the main processing.
- Paras. 16-17: more than one year can be justified for internal control if risk of misuse is demonstrated (sensitive data at scale, systematic surveillance) and documented investigation procedures exist.
- Para. 18: do not collect sensitive or highly personal data in logs that is not already in the processing.

**CNIL AI recommendations** [MED, STALE for 2024-2025 fiches]
- Fiches IA (2024-04 to 2025-07) cover development of AI systems: legal basis, DPIA, security, rights, annotation, model status. No fiche on logs, none dedicated to deployment. List: https://www.cnil.fr/fr/les-fiches-pratiques-ia . [HIGH] [STALE - 2025-07]
- Rights over prompts: CNIL states GDPR applies to use of models through prompts, and rights requests should say whether they target training data or the model. https://www.cnil.fr/fr/ia-respecter-lexercice-des-droits-des-personnes [MED, STALE]
- EDPB Opinion 28/2024 (2024-12-17): models are not automatically anonymous, legitimate interest needs a three-step test. [MED, STALE - 2024-12-17]
- **CNIL/CIANum note "IA agentique et protection des donnees personnelles", July 2026** [HIGH, read]. https://www.cnil.fr/sites/default/files/2026-07/ia-cianum-cnil.pdf . This is an analysis note, not a formal CNIL deliberation, and says the AI Act (full application 2027) has no specific agent rules. It proposes:
  - Traceability so that for each task the user can identify personal data used, agents involved, third-party services called, exchanges and their chronology. This also builds a chain of responsibility.
  - Data minimisation and purpose limitation per task, user control over accessible data.
  - Art. 22 GDPR remains fully applicable. CJEU SCHUFA ruling: human intervention must be meaningful.
  - Memory partitioned per agent, expiry or size limits, one session per processing.
  - Sandboxing, human validation for critical actions classified by risk (data access, modification, deletion, outbound data), kill switch, filtering of inputs/outputs, evaluation.
  - Notes that joint EDPB/Commission guidelines on AI Act and GDPR interplay are expected by end 2026 (not found published). [LOW]
- **ANSSI "Les essentiels: Securiser un systeme IA agentique" v1.0, October 2026** [HIGH, read]. https://messervices.cyber.gouv.fr/documents-guides/anssi_essentiels_ia_agentique_1.0.pdf . Seven principles. Directly relevant items:
  - Classify agent actions as autonomous, mandatory human approval, or forbidden. Human approval for critical actions, sensitive data access, identity or rights changes, production deployments. Keep the approval mechanism out of the agent's reach. Show the approver the exact action, purpose, expected effects. Execute only the approved action.
  - Deterministic authorisation independent of the agent, default deny.
  - Keep secrets out of model context, memory and logs.
  - Keep agent logs out of reach of agent modification, so the full sequence "request, authorisation, action, result" can be reconstructed after an incident.
  - Monitor in production against the documented mission (UEBA), emergency stop with access revocation, per-tenant and per-environment memory.
  - Cites the AI Act Art. 6 for high-risk systems.
- ANSSI "Recommandations de securite pour un systeme d'IA generative" ANSSI-PA-102, 2024-04-29 [STALE - 2024-04-29]: R9, no automatic execution of critical IS actions. [MED, from search snippet]

**Tension "log everything" vs "minimise" and how it resolves** [BEST PRACTICE]
- The AI Act has no logging duty for CNP today, and for high-risk the 6-month floor in Art. 19/26(6) is explicitly subject to data protection law. So the legal floor for content logging is zero for CNP, and the GDPR sets a ceiling by purpose.
- Pattern: always log metadata and decisions (who, what, when, model, tokens, tool, approval, outcome, hashes), never content by default, content capture opt-in per project with a stated purpose, short retention for content, redaction at the gateway or collector, hash of content so integrity and non-repudiation hold without the text.
- This matches the OTel spec default (content capture off) and CNIL para. 6, 12, 18.

### A3. Other instruments (short)

| Instrument | Status Nov 2026 | Relevance to CNP |
|---|---|---|
| NIS2 (Dir. 2022/2555) France transposition | Not in force. "Projet de loi Resilience" in Assemblee nationale plenary scheduled 2026-11-09 to 2026-11-13 after being withdrawn on 2026-10-06. Outcome not found. EU deadline was 2024-10-17, Commission action before the CJEU reported. ANSSI expects entity compliance around end 2028. Source: https://www.leto.legal/news/nis-2-projet-loi-resilience-assemblee-nationale-9-novembre-2026 (2026-11-05) [MED] | Applies only if CNP or a client company is an essential/important entity. Art. 23(4) timelines: early warning 24 h, notification 72 h, final report 1 month. [MED] |
| Cyber Resilience Act (Reg. 2024/2847) | Art. 14 reporting of actively exploited vulnerabilities and severe incidents applies since 2026-09-11 via ENISA Single Reporting Platform. Remaining obligations from 2027-12-11. Sources: https://digital-strategy.ec.europa.eu/en/policies/cra-reporting plus law-firm notes [MED] | Only for products with digital elements placed on the EU market by a manufacturer. An internal demo platform is likely out of scope. A tenant shipping a product is not. |
| DORA (Reg. 2022/2554) | Applies since 2025-01-17 | Only if a client is a financial entity. ICT logging and incident duties then cascade. Art. 19(2) AI Act and 26(6) let financial institutions keep logs inside their financial-law documentation. |
| Data Act (Reg. 2023/2854) | Applies since 2025-09-12 | Marginal for self-hosted. Relevant for switching and data sharing clauses if CNP is sold as a service. Not researched further. [LOW] |

### A4. Practical table: requirement, record, retention, readers, applicability

Retention figures marked (L) are legal floors or caps, (BP) are best-practice proposals.
Applicability columns: NOW = CNP as it stands. HR = if a tenant app is high-risk (Annex III). N/A = not applicable.

| Requirement | Source | What must be recorded | Retention | Who may read | CNP now | If tenant app high-risk |
|---|---|---|---|---|---|---|
| Automatic event logs of the AI system | AI Act Art. 12, 19, 26(6) | Per use: timestamp start/end, system and model version, input reference, output or decision reference, identity of human reviewer, anomalies. Art. 12(3) gives explicit minimum only for biometric ID. Content only as far as traceability for risk, post-market monitoring and operation monitoring needs. | At least 6 months (L), subject to GDPR. Suggest 12 months metadata (BP) | Provider, deployer, competent authority on reasoned request (Art. 21(2)); internal: security and compliance role only | N/A (not high-risk). Recommended as BP | Yes. Tenant is provider/deployer. CNP offers the capability |
| Human oversight evidence | AI Act Art. 14, 26(2); ANSSI principle 3; CNIL note 2026 | Action proposed (exact), approver identity, approval/rejection, timestamp, channel (CMP), what was executed vs approved | Same as logs (BP) | Security, compliance, auditors | Not legally required. Strong BP and already modelled by needs_user_confirmation / needs_cmp_approval | Yes |
| Deployer monitoring and incident handling | Art. 26(5) | Alerts, risk signals, suspension decisions, notifications to provider and authority | Duration of system plus at least 6 months (BP) | Incident team, DPO, authorities | N/A | Yes |
| Serious incident reporting | Art. 73 | Incident timeline, causal link analysis, affected persons, corrective action. Needs the logs above | Keep report and evidence at least as long as logs (BP) | Provider, authorities | N/A | Yes. Providers 15/10/2 day clocks. Deployer must inform provider first |
| Post-market monitoring | Art. 72 | Performance data collected from deployers over lifetime, evaluation scores, drift | Lifetime of system (BP) | Provider | N/A | Yes |
| Transparency to persons | Art. 50(1), 26(11); GDPR Art. 13-14, 22 | Evidence that disclosure was shown (banner version, timestamp). For Art. 26(11), that the person was informed | Same as logs (BP) | Compliance | Art. 50(1): likely for assistants, unless obvious. Record the disclosure config | Yes |
| Marking of generated content | Art. 50(2), Code of Practice (final 2026-06-10, voluntary, adequate per Commission 2026-07-08 per one secondary source) | Marking method and version applied to outputs | Config history (BP) | Compliance | Open question for internal dev agents [LOW]. Upstream model providers carry marking for their own systems | Case by case |
| AI literacy | Art. 4 (amended) | Training attendance, awareness material, dates | 3 years (BP) | HR, compliance | Yes, as a support measure | Yes |
| Application logs for personal data processing | GDPR Art. 5(2), 32; CNIL 2021-122 | Individually identified author, timestamp, operation type, reference of data; no duplication of data | 6 to 12 months (CNIL para. 8). Over 1 year only with documented justification (para. 16-17) | Segregated, strict-necessity access (para. 7), usage charter (para. 11) | Yes, for any log containing personal data | Yes |
| Integrity of logs | CNIL para. 12; ANSSI principle 7 | Signed or hash-chained records, logs outside agent write reach | Same as the log | n/a | BP, strongly recommended (agents are an adversarial writer) | Yes |
| Content capture (prompts, outputs, tool args and results) | GDPR minimisation, CNIL para. 18, OTel opt-in default | Redacted or hashed content, project-level consent and purpose | Short: 7 to 30 days (BP), not a legal figure | Project owner, DPO on request | Opt-in only | Opt-in, justified by Art. 12(2) purposes and DPIA |
| Access and erasure requests over logs | GDPR Art. 15, 17; CNIL para. 13 | Ability to find records by data subject identifier | Within response delay | DPO | Yes. Hash user IDs consistently, keep an index | Yes |
| Incident and breach notification | GDPR Art. 33; NIS2 Art. 23 if in scope | Timeline, scope, containment | Per policy | Security lead | GDPR yes. NIS2 only if entity in scope and law in force | Same |
| GPAI provider duties | Art. 53, 55 | Documentation, training summary, incident reports to AI Office | n/a | n/a | N/A, CNP is not a model provider. Request provider documentation under Art. 53(1)(b) as deployer BP | N/A |

---

## PART B. TECHNICAL STATE OF THE ART (Nov 2026)

### B1. OpenTelemetry GenAI semantic conventions

**Status: all Development, not Stable.** [HIGH]
- GenAI and MCP conventions moved out of the main repo in semantic-conventions v1.42.0 (2026-06-12) to https://github.com/open-telemetry/semantic-conventions-genai . No tagged release, CHANGELOG has only "Unreleased", README gives no stability level, Schema URL "TODO". Latest commits seen: 2026-11-06 (#692, #513, #144), so it is moving daily.
- Third-party posts claiming "stable" exist (one dated 2026-06-22). They conflict with the registry badges reported in dev.to (2026-07-16) and my reading of the repo. Treat "stable" claims as wrong. [MED]
- Practical rule: pin the instrumentation versions and expect renames. Opt-in env for latest names: OTEL_SEMCONV_STABILITY_OPT_IN=gen_ai_latest_experimental (secondary sources, John Hodge post updated 2026-11-03 and Coralogix doc) [MED].

**Content in the repo docs (docs/gen-ai/*.md), read from raw.githubusercontent.com on 2026-11-13** [HIGH]
- Spans: inference client spans (client-inference.md), embeddings, `invoke_agent` (INTERNAL for in-process agents, CLIENT variant for remote agent services, `invoke_workflow` when the app controls the flow), `create_agent`, `execute_tool` (INTERNAL, name `execute_tool {gen_ai.tool.name}`), retrieval, a2a, realtime.
- Agent identity: gen_ai.agent.name, gen_ai.agent.id, gen_ai.agent.version, gen_ai.conversation.id (only when a real ID exists), gen_ai.main_agent resource entity.
- Tool attributes: gen_ai.tool.name (required), gen_ai.tool.type, gen_ai.tool.call.id, and opt-in sensitive gen_ai.tool.call.arguments and gen_ai.tool.call.result. execute-tool.md defines NO approval, permission or user-confirmation attributes. Agent-as-tool adds gen_ai.transfer.mode and target. Skills add gen_ai.skill.*, process.executable.*, process.exit.code.
- MCP (mcp.md): span name `{mcp.method.name} {target}`, attributes mcp.method.name, mcp.session.id, mcp.protocol.version, jsonrpc.request.id, error.type (`tool_error` when isError), context propagated in `params._meta` as unprefixed traceparent/tracestate/baggage. Metrics mcp.client|server.operation.duration and session.duration. MCP tool-call spans should merge into an existing `execute_tool` span when detectable.
- Metrics (gen-ai-metrics.md): gen_ai.client.operation.duration, gen_ai.server.request.duration, gen_ai.server.time_per_output_token, gen_ai.server.time_to_first_token, gen_ai.invoke_agent.duration, .inference_calls, .tool_calls, gen_ai.execute_tool.duration.
- **Conflict on token metric name.** The metrics page no longer lists gen_ai.client.token.usage. gen-ai-token-metrics.md defines counters gen_ai.client.inference.usage.input_tokens, .output_tokens, .cache_read.input_tokens, .cache_write.input_tokens, .reasoning.output_tokens and histograms gen_ai.client.inference.operation.input_tokens / output_tokens. Yet agentgateway (agentgateway_gen_ai_client_token_usage), the VictoriaMetrics/OpenLIT blog (gen_ai_client_token_usage_sum) and most SDKs still emit the older histogram. Expect both generations to coexist in your data for a while. Counters are the better base for cost sums. [HIGH on spec, HIGH on agentgateway docs]
- Events: gen_ai.evaluation.result (gen_ai.evaluation.name, score.value, score.label, explanation, gen_ai.response.id), parented to the evaluated operation span when possible. Exception events gen_ai.invoke_agent.exception, gen_ai.execute_tool.exception (WARN severity). Inference-operation events are in client-inference.md (not read).
- **Content capture** (gen-ai-spans.md): opt-in. Env var OTEL_INSTRUMENTATION_GENAI_CAPTURE_CONTENT with NO_CONTENT (default), SPAN_ONLY, EVENT_ONLY, SPAN_AND_EVENT. Governs gen_ai.input.messages, gen_ai.output.messages, gen_ai.system_instructions, tool call args and results, retrieval queries, prompt variables, memory records. `_ref` variants (for example gen_ai.input.messages_ref) store a URI to external storage instead of inline content, which suits a "content in the Object Lock bucket, pointer in the trace" design. Docs warn the attributes "likely contain sensitive information including user/PII data".
- Events vs span attributes: since v1.37.0 (2025-08) messages moved from per-message events to span attributes (secondary: John Hodge). The current spec lets you choose where content goes (span, event, both). Evaluation results are events. [MED]
- **Not found in spec:** attributes for human approval, guardrail verdict, prompt-injection detection, user identity on GenAI spans. Use your own namespace (for example `cnp.approval.*`, `cnp.guardrail.*`) plus standard enduser.id and tenant attributes. This is a gap to document, not a standard to cite.

### B2. Who emits what

| Component | What it emits | Content capture | Evidence |
|---|---|---|---|
| agentgateway (Apache-2.0, 6.2k stars, pushed 2026-11-13; latest v1.7.0 2026-10-27) | Traces via OTLP (policy-enabled, off by default on Kubernetes via AgentgatewayPolicy). Default span attrs: HTTP set, gen_ai.operation.name, provider.name, request.model, response.model, usage.input_tokens/output_tokens, agw.ai.original_model, mcp.method.name, mcp.target, mcp.session.id, mcp.resource.*, gen_ai.tool.name, mcp.error.*; conditional jwt.sub, src.identity. Access logs per request with gen_ai.* fields and agw.ai.usage.cost.total, exportable as OTLP LogRecords. Prometheus: agentgateway_gen_ai_client_token_usage (labels token type, operation, system, request/response model, user_id when per-user API key rate limiting), agentgateway_cost_catalog_lookups_total. v1.7.0 adds cache, reasoning tokens, TTFT, session detection (x-claude-code-session-id), CEL MCP guardrails, guard scopes covering tool outputs by default. Guardrail verdicts and audit-mode detections go to metrics and the structured access log (not documented for traces). | Prompt/completion logging not documented on the pages read. Tool calls opt-in via llm.toolCalls in access-log and tracing attributes (adds overhead). Guard audit mode still calls the provider (cost, latency) | https://agentgateway.dev/docs/kubernetes/latest/documentation/llm/observability/ , https://agentgateway.dev/docs/standalone/latest/documentation/observability/traces/attribute-reference/ , https://agentgateway.dev/docs/kubernetes/main/documentation/llm/guardrails/overview/ , release v1.7.0 on GitHub [HIGH] |
| kagent (Apache-2.0, 4.3k stars, v1.0.0 released 2026-11-09, v1.0.1 2026-11-12) | v1.0.0 notes: OTEL via SDK-spec env vars, Weaver registry as telemetry contract, chat and execute_tool spans from runtimes, Python runtimes on upstream GenAI instrumentation, semconv metrics exported, legacy ADK spans dropped, trace context propagated on outbound LLM and MCP calls, harness invocation identity, trusted user as user.id, HITL pause on input-required. PR #1433 (merged 2026-03-10): controller tracing, A2A propagation, invoke_agent span with GenAI attributes. Docs page for 0.10.x showed span `agent_run [agent]` only and says 0.10.x is superseded by 1.0. | v1.0.1 backport "honor runtime content capture settings" implies content capture is runtime-configurable. Exact setting not found | https://api.github.com/repos/kagent-dev/kagent/releases , https://github.com/kagent-dev/kagent/pull/1433 [HIGH on release notes, docs for 1.0 tracing not found] |
| MCP servers (yours) | Only if instrumented. Spec defines mcp.* conventions and `_meta` propagation. Not researched per server SDK. | opt-in args/results | mcp.md [HIGH on spec] |
| Ollama | Not found: no native OTel. PRs open: #14203 (traces), #6537 (OTel metrics), #16998 (Prometheus metrics endpoint). None merged as of 2026-11-13. Third-party opentelemetry-instrumentation-ollama (OpenLLMetry) instruments the Python client only. | n/a | GitHub issues search 2026-11-13 [HIGH]. So measure local-model calls at agentgateway, not at Ollama. GPU metrics need a separate exporter [LOW, not verified] |
| LiteLLM proxy | OTel callback `callbacks: ["otel"]`; default is legacy span format, newer GenAI names behind OTEL_SEMCONV_STABILITY_OPT_IN=gen_ai_latest_experimental; behavior changed in v1.81.0 (litellm_request span collapsed by default). Enterprise-gated: key redaction in logs, audit logs, SSO. Licence boundary for OTel itself not confirmed. | via callback config | docs.litellm.ai, Coralogix, Fiddler pages [MED]. Not needed if agentgateway stays the gateway |
| OpenLIT SDK | Python/TS auto-instrumentation of 50+ libraries incl. MCP, emits GenAI attributes plus own cost extension (gen_ai.usage.cost). Evals scored through OpenLIT platform. Apache-2.0, 3.1k stars, pushed 2026-11-11. | configurable | https://victoriametrics.com/blog/victoriametrics-openlit-agents-observability/ (2026-07-03) [HIGH for the claims in that post] |
| OpenLLMetry / Traceloop SDK | Apache-2.0, 7.5k stars, pushed 2026-11-09. One secondary analysis says it still emits deprecated gen_ai.prompt/gen_ai.completion in places and adds traceloop.* attributes. Ownership claim (ServiceNow acquisition 2026-03) unverified. | configurable | futureagi.com, traceloop docs [LOW/MED] |
| Claude Code and Claude Agent SDK | CLAUDE_CODE_ENABLE_TELEMETRY=1 plus OTEL_METRICS_EXPORTER, OTEL_LOGS_EXPORTER, OTEL_TRACES_EXPORTER (+CLAUDE_CODE_ENHANCED_TELEMETRY_BETA=1 for traces, beta), standard OTEL_EXPORTER_OTLP_*. Metrics: claude_code.token.usage, cost.usage, session.count, code_edit_tool.decision, active_time.total, lines_of_code, commit, pull_request counts. Log events: user_prompt, api_request (model, cost_usd, tokens, request_id), tool_result, tool_decision (decision plus source: config, hook, user_permanent, user_abort), api_error, mcp_server_connection, permission_mode_changed. Spans: claude_code.interaction > llm_request, tool (> tool.blocked_on_user, tool.execution), hook. W3C TRACEPARENT in and out, forwarded to Bash. Attributes: session.id, prompt.id, user.id (random local ID), user.email when authed. Agent SDK doc: inject enduser.id and tenant.id through OTEL_RESOURCE_ATTRIBUTES to get a per-user audit trail. Does NOT use gen_ai.* names. | Off by default. Opt-in: OTEL_LOG_USER_PROMPTS, OTEL_LOG_TOOL_DETAILS, OTEL_LOG_TOOL_CONTENT (needs tracing), OTEL_LOG_ASSISTANT_RESPONSES, OTEL_LOG_RAW_API_BODIES (high sensitivity). Repo settings cannot enable telemetry, only managed settings | https://code.claude.com/docs/en/monitoring-usage , https://code.claude.com/docs/en/agent-sdk/observability [HIGH] |

### B3. Storage and query on the Victoria stack

State of the stack (from existing scratchpad file research-vm-ecosystem.md, same date, plus this session) [HIGH]
- VictoriaTraces v0.13.0 (2026-11-05) is pre-GA, data structure not frozen. OTLP HTTP and gRPC ingest, Jaeger query API for Grafana, per-day partitions, default retention 7d, tenant by AccountID/ProjectID headers with vmauth for authorization. Docs page: https://docs.victoriametrics.com/victoriatraces/
- VictoriaLogs v1.54.0 (2026-11-13) GA since v1.0.0, OTLP logs at /insert/opentelemetry/v1/logs, no S3 storage, no per-stream or per-tenant retention. Use several instances with different retention for audit logs (30d content, 12m metadata, for example). vmalert supports rule type vlogs.
- VictoriaMetrics: OTLP at /opentelemetry. `-opentelemetry.usePrometheusNaming` turns gen_ai.usage.cost into gen_ai_usage_cost_USD_sum (VM/OpenLIT blog).
- Official reference pipeline: https://victoriametrics.com/blog/victoriametrics-openlit-agents-observability/ (2026-07-03) and companion repo https://github.com/VictoriaMetrics/ai-observability (Apache-2.0, 9 commits, docker-compose with VM, VL, VT, Grafana, vmalert, rules/llm.yml, otelcol.yaml, one dashboard). Collector pattern: metrics pipeline strips high-cardinality labels (conversation, response and tool-call IDs, user.id); logs pipeline to VictoriaLogs with optional prompt capture; traces pipeline with tail sampling (keep errors, expensive runs, runs over 10 s, 5% of the rest). The post estimates 80 to 95% trace volume drop. OpenLIT's own destination doc lists the three OTLP endpoints (/opentelemetry, /insert/opentelemetry for logs and traces) and Grafana datasources: Prometheus type, VictoriaLogs plugin, Jaeger type.
- Audit consequence of tail sampling: tail sampling is right for ops traces and wrong for the compliance record. Run two paths. Ops path sampled. Audit path unsampled but metadata-only, as log records.

Mapping of your needs [BEST PRACTICE unless marked]
- GenAI dashboards: tokens, cost, latency, TTFT, errors from agentgateway Prometheus metrics and gen_ai.* metrics in VictoriaMetrics. Datasource prometheus type or victoriametrics-metrics-datasource. Traces via Jaeger datasource on VictoriaTraces, logs via victoriametrics-logs-datasource.
- Cost per project, agent, user: agentgateway gives cost per request in logs and spans (agw.ai.usage.cost.total) and per-user token series when user_id label is on. A per-project label is not a default metric label. Add it as a custom access-log and span attribute through the policy CEL attributes (documented mechanism), then aggregate in LogsQL stats, and keep metrics labels bounded. Per-project label in metrics: not verified. Claude Code cost per user/team: claude_code.cost.usage with OTEL_RESOURCE_ATTRIBUTES team.id.
- Tool-call audit trail: agentgateway MCP spans and access logs (mcp.method.name, gen_ai.tool.name, session id, jwt.sub) for gateway-routed calls. kagent execute_tool spans. Claude Code tool_decision and tool_result events. All carry trace IDs, so one LogsQL query by trace_id joins logs and spans.
- Guardrail and prompt-injection events: agentgateway guardrail detections land in metrics and the access log. Alert with vmalert type vlogs on a verdict field. Exact field name not found, check a live log line.
- Human-approval events: no OTel standard (confirmed absent in execute-tool.md). Emit one structured log record or span event per decision from CMP and the MCP server: `cnp.approval.state` (requested, confirmed, denied, cmp_pending, cmp_approved), `cnp.approval.approver` (Keycloak sub), `cnp.action.hash` (hash of exact action and parameters), `cnp.approval.channel`, trace context of the run. This mirrors the ANSSI principle of showing the exact action and executing only what was approved.
- Evaluation scores: spec event gen_ai.evaluation.result, or metric gen_ai_evaluation_score as in the VM/OpenLIT blog. A nightly eval job writing these into VM is cheaper than a platform.
- Per-run trace: propagate one traceparent from the CMP request, through MCP `params._meta` (spec), agentgateway, kagent (1.0 propagates on outbound LLM and MCP calls), sandbox processes (Claude Code forwards TRACEPARENT to Bash). Whether agentgateway extracts and continues an inbound traceparent: not verified in docs read, test it.

**Tamper-evident audit trail options for a small stack** [BEST PRACTICE, tools unevaluated]
1. S3 Object Lock in compliance mode (you already have it) as the system of record for the per-run "AI Act record", written at run end as a signed JSON object. This is WORM by storage guarantee. Matches CNIL para. 7 (write-only, segregated) and ANSSI principle 7 (outside agent reach).
2. Hash chaining: each run record includes prev_hash and its own SHA-256; write the daily chain head into Object Lock. Cheap, detects deletion and edit. Matches CNIL para. 12 "timestamp and sign at creation".
3. Signing: sign the daily manifest with a key held in Vault transit (you run Vault) or cosign. Key never visible to agents.
4. Transparency log: Sigstore Rekor v2 is tile-based, needs rekor-tiles plus a timestamp authority, no online inclusion proof API, so it is heavy for a small team. Single-binary Merkle logs such as Etchlog (RFC 6962, single-node) exist, maturity unchecked. Source: Sigstore docs and project READMEs via search. [LOW]. I would skip this for the demo.
5. VictoriaLogs and VictoriaTraces are mutable operational stores. Do not present them as the evidence store. Present them as the query layer, with the Object Lock copy as evidence.

### B4. Open-source LLM observability tools vs Victoria stack

Stars and push dates from GitHub API 2026-11-13.

| Tool | Licence | Weight (self-host) | Adds over Victoria stack | Status notes |
|---|---|---|---|---|
| Langfuse | MIT core, ee features with license key (GitHub reports NOASSERTION), 37.4k stars | Web + worker (docs: at least 2 CPU and 4 GB RAM for all app containers) plus Postgres, ClickHouse, Redis/Valkey, S3. DB sizing not documented on pages read | Prompt management, datasets, eval runs, annotation UI, session views. OTLP over HTTP only at /api/public/otel (v3.22.0+), no gRPC; maps gen_ai.* | Acquired by ClickHouse 2026-01-16, ClickHouse states MIT core stays. One site claims AGPL, contradicted by official statements. |
| Phoenix (Arize) | Elastic License 2.0 per secondary sources, GitHub NOASSERTION, 12.3k stars | Not found on pages read | Evals, datasets, experiments, OTel/OpenInference tracing | Arize independent per Hackernoon 2026-08-27, one site claims a Dynatrace deal, unresolved. Source-available, not OSI. [MED] |
| Opik (Comet) | Apache-2.0, 23.4k stars | Not found | Evals, datasets, online scoring, dashboards | Independent. |
| OpenLIT | Apache-2.0, 3.1k stars | SDK plus optional platform; VM official post shows the pipeline without its UI | Auto-instrumentation, cost, evals (server side in platform) | Active, officially paired with VM stack. |
| Helicone | Apache-2.0, 6.3k stars | Not checked | Proxy-based logging, caching | Acquired by Mintlify 2026-03-03, "maintenance mode". Avoid. |
| Lunary | Licence conflicting in sources (Apache-2.0 vs source-available) | Not checked | Prompt management, evals | GitHub API returned 404 for lunary-ai/lunary on 2026-11-13, status unclear. Avoid. |
| OpenLLMetry (instrumentation only) | Apache-2.0, 7.5k stars | Library | Python/TS instrumentation, not a backend | Emits some deprecated attribute names per one secondary source. |

Sources: https://langfuse.com/self-hosting , https://langfuse.com/self-hosting/deployment/infrastructure/containers , https://langfuse.com/integrations/native/opentelemetry , https://clickhouse.com/blog/clickhouse-acquires-langfuse-open-source-llm-observability , https://hackernoon.com/two-acquisitions-deep-two-still-standing-whats-really-happening-to-open-source-llm-observability

Recommendation for a RAM-scarce cluster [BEST PRACTICE]
- Stay on the Victoria stack plus one OTel Collector. It already gives cost, latency, tool-call audit, guardrail alerts and a join through trace IDs, with no extra datastore. Every tool above brings its own database tier, and the heaviest (Langfuse: ClickHouse + Postgres + Redis + S3) duplicates what you already run.
- Add Langfuse only if a concrete need appears: prompt versioning, datasets, human annotation or eval workflows that tenants ask for. Treat it as an optional per-tenant add-on, not a platform dependency. agentgateway has a documented Langfuse integration (blog 2026-02-17), so you can fan out later.
- Do not adopt Helicone or Lunary.
- Evals: start with a scheduled Go or Python job that replays sampled runs and emits gen_ai.evaluation.result events or a gen_ai_evaluation_score metric.

### B5. What "impressive and compliant" looks like for the demo

One run, one trace, one record:
1. Root span at the CMP/MCP entry: Keycloak user (sub, hashed email), project/tenant, assistant client, requested action, correlation id. Disclosure shown to the user (Art. 50(1) evidence).
2. Child spans: invoke_agent (kagent), chat spans per model call through agentgateway (provider, model, tokens, cost, guardrail verdicts), execute_tool spans per MCP call (tool name, argument hash), sandbox exec span (command hash, exit code, egress policy hits), then the deployment span (Argo CD sync, commit SHA, image digest).
3. Approval events in the same trace: needs_user_confirmation asked, answered by whom, needs_cmp_approval pending, approved by whom in CMP, the exact action hash executed. The Kubernetes, Vault, Keycloak and Argo CD audit logs already exist and join on user and time. Link them by commit SHA, deployment name and Keycloak sub.
4. Grafana "run view" dashboard: trace waterfall, cost, approval timeline, guardrail hits, link to the evidence object.
5. "AI Act record" export per run: a signed JSON (and optional PDF) produced by a small Go tool querying VictoriaTraces and VictoriaLogs by trace_id. Fields: run id, time window, user, project, system and model versions (agent version, gateway version, model name and provider, local model digest), disclosure version, list of model calls with token and cost, list of tool calls with argument hashes, approvals with approver identity, guardrail verdicts, resulting artefacts (commit, image digest, deployment), content references (pointer to opt-in content object or "not captured"), prev_hash and hash. Stored in the Object Lock bucket. Same fields map to Art. 12(2) purposes and the CNIL "demande, autorisation, action, resultat" sequence of ANSSI.
6. Privacy posture to show: content capture off by default, per-project opt-in switch, redaction at the gateway, retention dashboard, a DSAR search by hashed user id.
7. Label it honestly: "built to the Art. 12/19/26(6) shape, not a conformity claim, no harmonised standard yet".

### B6. Open questions and conflicts
- OTel token metric naming (older gen_ai.client.token.usage vs new gen_ai.client.inference.usage.*) differs between spec main and emitters. Pin versions.
- Content-capture env var: spec says OTEL_INSTRUMENTATION_GENAI_CAPTURE_CONTENT, a secondary source says Python instrumentation still reads OTEL_INSTRUMENTATION_GENAI_CAPTURE_MESSAGE_CONTENT. Verify per library.
- agentgateway: prompt/completion content logging not documented on pages read, guardrail verdict field names, inbound traceparent continuation, per-project metric labels. Check live.
- kagent 1.0 tracing docs and the exact content-capture switch not found.
- NIS2 France: outcome of the 2026-11-09 to 2026-11-13 Assemblee session unknown.
- Art. 50(2) applicability to internal dev agents and Art. 72 plan deadline (2027-09-02) not cross-checked on EUR-Lex.
- France AI Act authority designation unsettled.
- EN 18286 OJ date (2026-10-14) from a secondary source. Commission page says it was published as Implementing Decision (EU) 2026/2274.
- prEN ISO/IEC 24970 final status unknown.

---

## Sources (accessed 2026-11-13 unless noted)

Legal, primary or official
- https://ai-act-service-desk.ec.europa.eu/en/ai-act/timeline/timeline-implementation-eu-ai-act
- https://ai-act-service-desk.ec.europa.eu/en/ai-act/article-3
- https://ai-act-service-desk.ec.europa.eu/en/ai-act/article-4
- https://ai-act-service-desk.ec.europa.eu/en/ai-act/article-12
- https://ai-act-service-desk.ec.europa.eu/en/ai-act/article-19 (search result summary)
- https://ai-act-service-desk.ec.europa.eu/en/ai-act/article-21
- https://ai-act-service-desk.ec.europa.eu/en/ai-act/article-25
- https://ai-act-service-desk.ec.europa.eu/en/ai-act/article-26
- https://ai-act-service-desk.ec.europa.eu/en/ai-act/article-50
- https://ai-act-service-desk.ec.europa.eu/en/ai-act/article-72
- https://ai-act-service-desk.ec.europa.eu/en/ai-act/article-73
- https://ai-act-service-desk.ec.europa.eu/en/ai-act/annex-3
- https://eur-lex.europa.eu/eli/reg/2026/1744 (search metadata only, page not rendered)
- https://eur-lex.europa.eu/legal-content/EN/TXT/?uri=CELEX%3A02024R1689-20260727 (search metadata only, not rendered)
- https://artificialintelligenceact.eu/ai-act-explorer/digital-omnibus/ (unofficial tracker)
- https://digital-strategy.ec.europa.eu/en/policies/ai-act-standardisation
- https://digital-strategy.ec.europa.eu/en/policies/guidelines-gpai-providers
- https://digital-strategy.ec.europa.eu/en/policies/cra-reporting (search result)
- https://www.entreprises.gouv.fr/priorites-et-actions/transition-numerique/soutenir-le-developpement-de-lia-au-service-de-0 (search result)
- https://www.cnil.fr/sites/cnil/files/atoms/files/recommandation_-_journalisation.pdf
- https://www.cnil.fr/fr/les-fiches-pratiques-ia
- https://www.cnil.fr/fr/ia-respecter-lexercice-des-droits-des-personnes (search result)
- https://www.cnil.fr/sites/default/files/2026-07/ia-cianum-cnil.pdf
- https://messervices.cyber.gouv.fr/documents-guides/anssi_essentiels_ia_agentique_1.0.pdf
- https://messervices.cyber.gouv.fr/documents-guides/Recommandations_de_s%C3%A9curit%C3%A9_pour_un_syst%C3%A8me_d_IA_g%C3%A9n%C3%A9rative.pdf (search snippet only)
- https://standards.iteh.ai/catalog/standards/cen/38472bd6-ab5c-4908-ac64-6990f45027e6/pren-iso-iec-24970 (search result)
- https://www.iso.org/standard/88723.html (403, search snippet only)

Legal, secondary
- https://technology.org/2026/10/13/eu-ai-act-deadlines-after-the-digital-omnibus-what-applies-from-december-2026-to-august-2028/ (403 to fetcher, search snippet only)
- https://www.leto.legal/news/nis-2-projet-loi-resilience-assemblee-nationale-9-novembre-2026
- https://www.jonesday.com/en/insights/2026/06/european-commission-publishes-final-code-of-practice-on-marking-and-labelling-aigenerated-content
- https://www.wilmerhale.com/en/insights/blogs/wilmerhale-privacy-and-cybersecurity-law/20250724-european-commission-issues-guidelines-for-providers-of-general-purpose-ai-models
- https://cms.law/en/int/legal-updates/edpb-opinion-28-2024-key-takeaways-on-processing-personal-data-in-the-context-of-ai-models
- https://www.isms.online/iso-42001/annex-a-controls/a-6-ai-system-life-cycle/a-6-2-8-ai-system-recording-of-event-logs/
- https://www.mccannfitzgerald.com/knowledge/data-privacy-and-cyber-risk/cyber-resilience-act-reporting-obligations-apply-from-11-september-2026

Technical
- https://github.com/open-telemetry/semantic-conventions-genai (and raw docs: docs/gen-ai/gen-ai-metrics.md, gen-ai-token-metrics.md, gen-ai-spans.md, gen-ai-agent-spans.md, gen-ai-events.md, execute-tool.md, mcp.md; commits API)
- https://opentelemetry.io/docs/specs/semconv/gen-ai/ (moved notice)
- https://john-hodge.com/blog/opentelemetry-genai-semantic-conventions/
- https://dev.to/azena-ai/opentelemetrys-genai-semantic-conventions-are-not-stable-yet-heres-what-actually-shipped-in-2026-3mke
- https://agentgateway.dev/docs/kubernetes/latest/documentation/llm/observability/
- https://agentgateway.dev/docs/standalone/latest/documentation/observability/traces/attribute-reference/
- https://agentgateway.dev/docs/standalone/main/documentation/llm/observability/
- https://agentgateway.dev/docs/kubernetes/main/documentation/llm/guardrails/overview/
- https://api.github.com/repos/agentgateway/agentgateway and /releases
- https://api.github.com/repos/kagent-dev/kagent and /releases
- https://github.com/kagent-dev/kagent/pull/1433
- https://kagent.dev/docs/kagent/observability/tracing
- https://code.claude.com/docs/en/monitoring-usage
- https://code.claude.com/docs/en/agent-sdk/observability
- https://api.github.com/search/issues?q=repo:ollama/ollama+opentelemetry
- https://docs.victoriametrics.com/victoriatraces/
- https://victoriametrics.com/blog/victoriametrics-openlit-agents-observability/
- https://docs.openlit.io/latest/sdk/destinations/victoriametrics-stack.md
- https://github.com/VictoriaMetrics/ai-observability
- https://langfuse.com/self-hosting , /self-hosting/deployment/infrastructure/containers , /integrations/native/opentelemetry
- https://clickhouse.com/blog/clickhouse-acquires-langfuse-open-source-llm-observability
- https://hackernoon.com/two-acquisitions-deep-two-still-standing-whats-really-happening-to-open-source-llm-observability
- GitHub API repo metadata for langfuse/langfuse, Arize-ai/phoenix, comet-ml/opik, openlit/openlit, Helicone/helicone, traceloop/openllmetry, lunary-ai/lunary (404)
- https://pistack.xyz/posts/2026-04-25-self-hosted-log-integrity-tamper-evident-audit-logging-guide-2026/ and Sigstore docs (search results only)
- Local file: scratchpad/research-vm-ecosystem.md (prior research, same date)
