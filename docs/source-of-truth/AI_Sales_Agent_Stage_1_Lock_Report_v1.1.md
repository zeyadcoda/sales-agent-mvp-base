# AI Sales Agent — Stage 1 Product Redesign & Operating Model Lock Report

| Field | Value |
| --- | --- |
| Document | AI Sales Agent — Stage 1 Product Redesign & Operating Model Lock Report |
| Version | 1.1 |
| Status | Under Review — v1.0 + Approved Super Admin Lock Amendment |
| Owner | Zeyad Al Gharabi |
| Approver | Project Owner — explicit approval required |
| Created | 2026-08-16 |
| Last Updated | 2026-08-18 |
| Scope | Product vision, operating model, MVP boundary, AI/human model, acquisition, administration, agents, UX baseline |


> **Authority note:** This document consolidates the approved working decisions from the product-design sessions. It is a Stage 1 lock candidate. It becomes the approved Stage 1 baseline only after explicit Project Owner approval. The accompanying BRD and Master PRD are derived from the same baseline.

## Table of Contents

- [1. Source Inventory and Treatment](#1-source-inventory-and-treatment)
- [2. Product Vision](#2-product-vision)
- [3. Product Principles Preserved](#3-product-principles-preserved)
- [4. Approved MVP Boundary](#4-approved-mvp-boundary)
- [5. Primary End-to-End Workflows](#5-primary-end-to-end-workflows)
- [6. B2B, B2C, and Hybrid Model](#6-b2b-b2c-and-hybrid-model)
- [7. AI and Human Operating Model](#7-ai-and-human-operating-model)
- [8. Control Panels and Administration](#8-control-panels-and-administration)
- [9. Decision Register](#9-decision-register)
- [10. Contradictions Resolved](#10-contradictions-resolved)
- [11. Open Questions](#11-open-questions)
- [12. Deferred Decisions](#12-deferred-decisions)
- [13. Risk Register](#13-risk-register)
- [14. Documents Updated or Required](#14-documents-updated-or-required)
- [15. Traceability](#15-traceability)
- [16. Stage Acceptance Checklist](#16-stage-acceptance-checklist)
- [17. Recommendation](#17-recommendation)

## 1. Source Inventory and Treatment

The uploaded materials were reviewed as references, not automatically accepted requirements.

| Source | Purpose | Treatment | What remains useful | What is superseded or unresolved |
| --- | --- | --- | --- | --- |
| AI_Sales_Agent_Project_Start_Here_v0.1.md | Project initialization baseline | Reference only | Multi-tenant scope; company/project understanding; provider adapters; test-mode principle; Control Panel; security boundaries; Google ADK preference. | The one-week hard delivery assumption is superseded. Fixed role examples are superseded by direct permissions. Detailed product behavior is superseded by approved working decisions. |
| AI Sales Agent PRD (SalesStrategyAgent_Logic).docx | Earlier Strategy Agent concept | Reference only | Document extraction; product understanding; ICP; positioning; sales/marketing strategy; validation before strategy output. | Fixed category mappings, universal field checklist, universal scoring weights, and older onboarding order are not authoritative. |
| AI Sales Agent PRD (SalesStrategyAgent_Logic) 2.docx | Duplicate/variant of earlier Strategy Agent concept | Reference only / duplicate | Same useful extraction and strategy concepts as the other Strategy document. | Treated as duplicate reference; no independent approval. |
| AI Sales Agent PRD (Channels_Logic).docx | Earlier channel behavior concept | Reference only | Research versus messaging channel distinction; email/WhatsApp/Instagram/Facebook channel considerations; product content sharing. | Provider/API behavior was not specified. Current autonomy, integration ownership, and Meta-to-WhatsApp rules supersede it. |
| AI Sales Agent PRD (AIAgents_Logic).docx | Earlier Lead/Sales Agent concept | Reference only | Initial/dynamic qualification; Next Best Action; persistent context; handover brief; channel transitions. | Temporary-lead-only storage, rigid follow-up timings, old qualification labels, and broad handover triggers are superseded. |

## 2. Product Vision

The product is a standalone, multi-tenant **AI Sales and Marketing execution platform** for B2B, B2C, and hybrid Organizations.

It turns approved business and Project understanding into an executable acquisition and sales operation:

```text
Company registration and activation
→ Company Understanding
→ Project and Product Understanding
→ ICP, Sales Strategy, Marketing Strategy, and acquisition direction
→ Meta and/or Web lead acquisition
→ Master Lead / Business Pool intelligence
→ Enrichment and qualification
→ Dynamic lead-specific Strategy
→ Human-like AI sales conversation
→ Human takeover for final MVP action
→ Recorded conversion
→ Analytics and evidence-backed learning
```

The differentiating product value is not a generic chatbot. It is a governed operating system that combines acquisition, lead intelligence, dynamic sales Strategy, multi-channel AI conversation, human control, attribution, and platform-level agent administration.

## 3. Product Principles Preserved

- Standalone multi-tenant AI-powered Sales and Marketing platform.
- B2B, B2C, and hybrid Organizations are supported; each MVP Project has one primary B2B or B2C mode.
- Organization Intelligence, Project Intelligence, and Project Lead Intelligence are separate scopes.
- The application—not the LLM/agent framework—is the source of truth for permissions, limits, protected state, audit, and side effects.
- Human review gates exist for company understanding, project/product understanding, strategy, campaign publication, normal follow-ups, and final MVP conversion actions.
- Provider integrations use controlled adapters and organization/platform ownership rules.
- Project and tenant isolation apply to data, retrieval, memory, prompts, tools, logs, and background jobs.
- Google ADK is the approved working Agent Component, not the whole application.

## 4. Approved MVP Boundary

- Manual company registration review and activation by Super Admin.
- AI-guided Company Understanding with targeted clarification and human approval.
- AI-guided Project/Product Understanding with versioned human-approved baseline.
- Structured Strategy generation, limited per-Project revisions, channel configuration, and final Project onboarding.
- Action-led Organization and Project dashboards.
- Auto-generated Project Product Catalogue with post-onboarding sales-support enrichment.
- Lead acquisition through real Meta Ads, WhatsApp, Instagram, Facebook Messenger, Web/Google discovery, and generic email.
- Master Lead Pool for people and Business Pool for companies, with strict private/reusable data separation.
- Initial scoring, dynamic qualification, persistent lead strategy, autonomous active sales conversation, and supervised follow-up.
- Claim-based human handover for all final MVP commercial actions.
- Campaign and source attribution through human-recorded conversion.
- Super Admin CP, Organization Administration, direct permissions, Plans/limits, Action Center, in-app/email notifications, audit, and agent observability.
- Customer Simulator/Evaluator capability and safe test/sandbox foundations; detailed customer-facing Test Mode remains deferred.

## 5. Primary End-to-End Workflows

### 5.1 Organization activation and onboarding

```text
Landing page
→ Registration application
→ Duplicate/identity validation
→ Super Admin review/contact
→ Organization + subscription activation
→ First login
→ Company Understanding input
→ AI analysis
→ Initial review
→ Targeted clarification
→ Final Business Overview review
→ Approval
→ Organization Dashboard
```

### 5.2 Project onboarding

```text
Create Project
→ Project name + optional description
→ Free-text commercial brief + documents
→ AI Project/Product analysis
→ Targeted clarification
→ Human review and approval
→ User selects Plan the Strategy
→ Structured Strategy generation
→ Limited Strategy revisions
→ Strategy acceptance
→ Channel/Sales AI configuration
→ Final onboarding approval
→ Project LIVE
→ Project Dashboard
```

### 5.3 Meta acquisition and sales

```text
Lead Discovery recommends Meta
→ Campaign Agent prepares setup + lead path + creative guide
→ Human reviews lead capture and settings
→ Human uploads media
→ AI quality review + deterministic validation
→ Authorized user publishes
→ Meta lead enters WhatsApp first
→ Master Lead create/match
→ Enrichment and Lead Strategist
→ Autonomous active Sales AI
→ Human-approved/delegated follow-up
→ Claim-based human handover
→ Human final action
→ Conversion + attribution
```

### 5.4 Web discovery and email sales

```text
Lead Discovery recommends Web
→ User approves Discovery Plan
→ Fresh Search/Google discovery
→ Contactable candidates enter Master Lead DB
→ Enrichment
→ Project-specific score and ranking
→ Allocate up to Project Lead Limit
→ AI sends personalized first email
→ Generic mailbox ingests threaded replies
→ Autonomous active Sales AI
→ Human-approved/delegated follow-up
→ Human handover and conversion
```

## 6. B2B, B2C, and Hybrid Model

- A hybrid Organization may operate both B2B and B2C Projects.
- Each MVP Project has one primary B2B or B2C mode.
- B2B discovery identifies a matching business and practical user/influencer who can advance purchase—not automatically the CEO.
- B2C acquisition and qualification use consumer-relevant context and never require employer, company size, job title, procurement authority, or buying committee.
- The same Master Lead may match unrelated future Projects; Project-specific score, Strategy, conversation, and outcome remain separate.

## 7. AI and Human Operating Model

### AI may act autonomously
- Analyze company, Project, product, market, and lead context.
- Generate approved structured Strategy candidates.
- Plan acquisition sources and web search.
- Send the initial personalized outreach to an approved eligible web lead.
- Respond to Meta inbound leads and conduct active conversations.
- Qualify naturally, adapt objections, and share approved product content.
- Recommend next actions, follow-ups, handovers, and Strategy improvements.
- Execute follow-ups only when explicitly delegated for that Project Lead.

### Human approval/control remains mandatory
- Approve Company Understanding and Project/Product Understanding.
- Approve the Strategy package and Project onboarding.
- Approve Meta lead-capture setup and explicitly publish campaigns.
- Approve normal follow-ups outside active conversation.
- Take over any conversation at any time.
- Complete every final MVP action: booking/confirmation, quote, negotiation, contract, reservation, payment, discount, or final sale.
- Record the final outcome/conversion.

### Deterministic application services remain mandatory
- Authentication, tenancy, permissions, Project access, entitlements, and limits.
- Communication eligibility and suppression.
- State-transition validation, idempotency, auditing, notification/action creation.
- Credential protection, provider calls, campaign spend controls, and webhook verification.

## 8. Control Panels and Administration

### Super Admin Control Panel
- Registration applications, Organizations, subscriptions, Plans, limits, overrides, suspension, closure, and external-action kill switch.
- Organization-specific Meta Ads, WhatsApp, Instagram, and Facebook integration setup.
- Platform-level Google/Search, AI provider credentials, and shared research-provider configuration.
- Per-agent model, credential reference, prompt/version, tools, status, testing, activation, and rollback.
- Agent runs, tool calls, failures, latency, usage, sensitive trace controls, and end-to-end product observability.
- Controlled and audited support access.

### Organization Administration
- Company profile and Company Understanding.
- Team membership, direct permissions, and Project access.
- Projects and Organization email connection.
- Action Center, in-app/email notifications, read-only Plan/usage, settings, and tenant audit.

## 9. Decision Register

The following stable decisions were carried forward from the working sessions.

| Decision ID | Area | Decision | Status |
| --- | --- | --- | --- |
| D-017 | Company Understanding | Company Understanding begins after the approved company administrator's first login. | APPROVED |
| D-018 | Project Understanding | Every new Project performs its own Project Understanding. | APPROVED |
| D-019 | Project Understanding | Project Understanding produces product intelligence, project knowledge, audience, marketing strategy, and sales strategy inputs. | APPROVED |
| D-020 | Project Products | Core Project products cannot be added after the approved Project onboarding baseline. | APPROVED |
| D-021 | Product Catalogue | The Project Product Catalogue is generated from onboarding and can later be enriched with approved sales-support content. | APPROVED |
| D-022 | Knowledge Isolation | Project knowledge must not contaminate another Project or Organization. | APPROVED |
| D-023 | Lead Intelligence | Each Project Lead receives a dynamically updated AI engagement strategy. | APPROVED |
| D-024 | AI/Human Model | AI conducts outreach and sales interaction; a human completes the final MVP commercial action. | APPROVED |
| D-025 | Master Lead Pool | All captured contactable people enter a persistent Master Lead Pool. | APPROVED |
| D-026 | Business Pool | Companies/organizations have a persistent Business Pool profile for permitted future matching. | APPROVED |
| D-027 | Delivery | The full product is developed iteratively; the first implementation target is a UX prototype in approximately three days. | APPROVED |
| D-028 | Testing | Features are tested incrementally during implementation. | APPROVED |
| D-029 | Company Intelligence | Company onboarding focuses on company/business context, not Project product prices; later changes use a controlled versioned update flow. | APPROVED |
| D-030 | Approval | Project products and Product Understanding require human review and approval before Strategy. | APPROVED |
| D-031 | Acquisition | B2B/B2C influences but does not hard-code acquisition channels. | APPROVED |
| D-032 | Handover | AI remains in control until an unsupported/protected action or configured handover condition occurs. | APPROVED |
| D-033 | Learning | MVP learning is structured outcome capture, lead adaptation, analytics, and human-approved strategy improvement—not uncontrolled self-training. | APPROVED |
| D-034 | Registration | Company registration creates an application that requires Platform approval before activation. | APPROVED |
| D-035 | Plans | Plans are configured by Super Admin. | APPROVED |
| D-035A | Strategy Allowance | `ai_credits` means Project Strategy revision allowance before approval, not model/token credit. | APPROVED |
| D-036 | Business Pool Privacy | Only permitted reusable business understanding enters the Business Pool; private tenant information is not shared. | APPROVED |
| D-037 | Digital Presence | Registration requires a website, official social account, or another approved digital business presence. | APPROVED |
| D-038 | Activation | Platform approval activates the Organization and subscription. | APPROVED |
| D-039 | Company Outputs | Company onboarding produces private Organization Intelligence and a separately governed reusable Business Pool profile. | APPROVED |
| D-040 | Company Updates | Material Company Understanding updates create a new version and require controlled approval. | APPROVED |
| D-041 | Multi-Organization User | One user identity may belong to multiple Organizations. | APPROVED |
| D-042 | CR Uniqueness | CR/business registration number plus jurisdiction must be unique for registration. | APPROVED |
| D-043 | Duplicate Company | A company already registered is blocked and directed to sign in or contact support. | APPROVED |
| D-044 | Duplicate Domain | A duplicate company domain/digital identity blocks a new tenant application by default. | APPROVED |
| D-045 | Clarification | Understanding agents ask only targeted questions needed to resolve material gaps. | APPROVED |
| D-046 | Credits | Company Understanding analysis/revision does not consume Strategy Revision Allowance. | APPROVED |
| D-047 | Company Permission | Organization Owner and explicitly permitted users may complete Company Understanding. | APPROVED |
| D-048 | Zero Allowance | At zero Strategy Revision Allowance the user cannot regenerate Strategy, but may approve the latest version. | APPROVED |
| D-049 | Project Onboarding | Project onboarding is AI-led and progressively asks for missing information. | APPROVED |
| D-050 | Understanding Approval | Project/Product Understanding must be approved before Strategy generation. | APPROVED |
| D-051 | Understanding Revisions | Project/Product Understanding revisions do not consume Strategy Revision Allowance. | APPROVED |
| D-052 | Strategy Trigger | Strategy generation is a separate user-initiated stage. | APPROVED |
| D-053 | Strategy Package | Strategy includes ICP/personas, pains/gains, product fit, market, sales, marketing, channel, qualification, and conversion direction. | APPROVED |
| D-054 | Meta Direction | When Meta is appropriate, Strategy includes campaign overview and execution direction. | APPROVED |
| D-055 | Strategy Revision | Each user-requested Strategy revision consumes one per-Project allowance unit. | APPROVED |
| D-056 | Onboarding Completion | Strategy approval alone completes onboarding. | SUPERSEDED by D-061 |
| D-057 | Project Objective | AI infers what the client is selling and seeks to achieve; the user confirms it. | APPROVED |
| D-058 | Strategy Approval | The user approves one complete Strategy package. | APPROVED |
| D-059 | Execution Boundary | Strategy approval does not automatically launch acquisition or spend. | APPROVED |
| D-060 | Channel Settings | Project/channel AI language, tone, style, restrictions, and instructions are configured before final onboarding approval. | APPROVED |
| D-061 | Project Onboarding Sequence | Strategy is accepted, channel settings are completed, then final Project onboarding is approved. | APPROVED |
| D-062 | Project Dashboard | The Project Dashboard combines performance overview and action-required guidance. | APPROVED |
| D-063 | Catalogue Source | Project Catalogue items originate from approved onboarding, not post-onboarding manual product creation. | APPROVED |
| D-064 | Catalogue Isolation | Sales-support catalogue content can help conversations but does not silently alter Strategy. | APPROVED |
| D-065 | Media Sharing | Sales AI may send approved product media/content when relevant. | APPROVED |
| D-066 | Promotions | Controlled coupons/special offers are supported. | DEFERRED/SUPERSEDED by D-271 for MVP |
| D-067 | Product Mutation | MVP does not support normal material mutation of the approved product baseline. | APPROVED |
| D-068 | Campaign Publish | An authorized user explicitly publishes a Meta campaign. | APPROVED |
| D-069 | Campaign Results | Published campaign pages show status, performance, and generated leads. | APPROVED |
| D-070 | Campaign Changes | Every campaign has an end date; authorized users may increase budget, extend, pause, and resume. | APPROVED |
| D-071 | Lead Persistence | Every captured contactable lead enters the Master Lead DB. | APPROVED |
| D-072 | Search Planning | AI generates web discovery criteria, keywords/query families, and search strategy. | APPROVED |
| D-073 | Discovery Approval | Users review and approve the Discovery Plan before external search execution. | APPROVED |
| D-074 | Pool Reuse | Master profiles are dynamically enriched across permitted sources and may support later acquisition matching. | APPROVED |
| D-075 | Project Context | A Master Lead may be used by many Projects; each Project maintains separate sales context. | APPROVED |
| D-078 | Discovery Volume | AI recommends candidate volume; the default web candidate target is Project Lead Limit × configurable multiplier. | APPROVED |
| D-079 | Fresh First | Current acquisition policy begins with fresh discovery before Master Pool supplementation. | APPROVED |
| D-080 | Contactability | A person becomes a captured Lead only when at least one usable permitted contact method exists. | APPROVED |
| D-081 | B2B Persona | B2B discovery targets practical users/influencers who can advance the purchase, not automatically the CEO. | APPROVED |
| D-082 | Discovery Stop | Discovery stops at target, exhaustion, provider/cost limits, excessive duplication, or manual stop. | APPROVED |
| D-083 | Initial Scoring | Initial scoring is Project-specific and generated from the approved Strategy within platform bounds. | APPROVED |
| D-084 | Lead Allocation | MVP fills the full Project Lead Limit with the highest-ranked valid candidates available. | APPROVED |
| D-085 | Dynamic Qualification | Observed interaction evidence progressively supersedes uncertain initial assumptions. | APPROVED |
| D-086 | Source-Aware Entry | Web leads begin as outbound candidates; Meta inbound leads begin with explicit intent. | APPROVED |
| D-087 | Lead Engagement Plan | Every Project Lead receives a dynamic lead-specific engagement plan. | APPROVED |
| D-088 | Qualified Definition | Qualified requires product fit, genuine need/interest, and a realistic path to progress. | APPROVED |
| D-089 | Qualification Outcomes | Qualified, Nurture, Not Qualified, Not Interested, and Suppressed are distinct outcomes with reasons. | APPROVED |
| D-090 | Qualified Behavior | AI continues selling a Qualified lead until conversion preparation, handover, suppression, or another valid outcome. | APPROVED |
| D-091 | Nurture Rule | Immediate pause-on-objection model. | REJECTED/SUPERSEDED |
| D-092 | Not Qualified | Not Qualified stops this Project's sales activity but preserves the Master Lead. | APPROVED |
| D-093 | Not Interested | Not Interested stops this Project; contextual signal may be retained without global generalization. | APPROVED |
| D-094 | Suppression | Suppression cancels applicable outreach and cannot be overridden by AI. | APPROVED |
| D-095 | Handover | Handover stops autonomous messaging and produces a structured human brief. | APPROVED |
| D-096 | No Response | No response is a signal/attention condition, not an automatic final status. | APPROVED as redesigned |
| D-097 | Human-Like Sales | Sales behavior must be contextual, adaptive, non-repetitive, and professionally natural. | APPROVED |
| D-098 | Follow-Up Strategy | AI determines the best follow-up strategy dynamically rather than using one universal timing sequence. | APPROVED |
| D-100 | Follow-Up Approval | Normal follow-ups outside an active conversation require human approval. | APPROVED |
| D-101 | Initial Outbound | AI automatically sends personalized first outreach to approved eligible web leads. | APPROVED |
| D-102 | Active Conversation | AI autonomously handles Meta inbound and normal active sales conversation replies. | APPROVED |
| D-103 | Handover Principle | Human takeover is the exception and occurs only when genuinely required. | APPROVED |
| D-104 | Active Autonomy | AI may communicate autonomously while the conversation is active. | APPROVED |
| D-105 | Context Inactivity | Conversation inactivity is context/channel aware; there is no universal timeout. | APPROVED |
| D-106 | Inbound Override | A new inbound message reopens the conversation and invalidates obsolete pending follow-ups. | APPROVED |
| D-107 | Scheduled Follow-Up | Approved scheduled follow-ups are revalidated immediately before sending. | APPROVED |
| D-108 | Adaptive Sequence | Follow-ups are proposed one at a time based on the latest state. | APPROVED |
| D-109 | Delegated Follow-Up | A user may explicitly delegate follow-up execution for one Project Lead to AI. | APPROVED |
| D-110 | Unresponsive Alert | Material unresponsiveness creates a visible alert and AI recommendation. | APPROVED |
| D-111 | Manual Takeover | An authorized user may take over any conversation at any time. | APPROVED |
| D-112 | Return to AI | A human may return control to AI after full context reassessment. | APPROVED |
| D-113 | Persistent Sales Memory | Complete conversation history and structured lead memory are application-managed and persistent. | APPROVED |
| D-114 | Living Lead Strategy | Each Project Lead has one structured evolving sales strategy. | APPROVED |
| D-115 | Next Best Action | Each active Project Lead has one current Next Best Action or explicit Wait/No Action with reason. | APPROVED |
| D-116 | Material Event | Lead Strategy updates on meaningful new evidence, not every trivial message. | APPROVED |
| D-117 | Strategy/Message Separation | Commercial approach is decided before channel-specific wording is generated. | APPROVED |
| D-118 | Human Instruction | Users inspect and instruct AI through the workspace instead of directly editing protected strategy/state fields. | APPROVED |
| D-119 | Lead Strategist Agent | Dynamic lead strategy is owned by a dedicated configurable Lead Strategist Agent. | APPROVED |
| D-120 | Agent Boundary | Major reasoning uses agents; deterministic services own security, limits, validation, and protected state. | APPROVED |
| D-121 | Agent Set | The approved major agent set includes understanding, strategy, discovery, enrichment, lead strategy, sales, campaign, simulator, and evaluator agents. | APPROVED |
| D-122 | Governed Handoffs | Agents exchange structured governed state/contracts, not free-form authoritative chat. | APPROVED |
| D-123 | Google ADK | Google ADK is the platform Agent Component; the application is built around it. | APPROVED WORKING DECISION |
| D-124 | Lead Strategist Trigger | Lead Strategist runs initially and on material sales events/human instructions. | APPROVED |
| D-125 | Memory Updates | Lead Strategist proposes structured memory updates; application services persist them. | APPROVED |
| D-126 | Lead Strategist Contract | Lead Strategist returns understanding, score/state recommendation, hypothesis, approach, objection, next action, follow-up, handover, memory updates, and user summary. | APPROVED |
| D-127 | Sales Agent Trigger | Sales Agent runs only when inbound communication or a permitted outbound action requires communication. | APPROVED |
| D-128 | Natural Communication | Sales communication is contextual, non-repetitive, adaptive, multilingual, and channel appropriate. | APPROVED |
| D-129 | Grounded Claims | Sales Agent may make only claims supported by approved Project knowledge/offers. | APPROVED |
| D-130 | Natural Qualification | Qualification is collected naturally through conversation rather than rigid interrogation. | APPROVED |
| D-131 | Sales Agent Contract | Sales Agent returns communication/action plus evidence, material-event indication, references, uncertainty, and strategist re-run signal. | APPROVED |
| D-132 | Enrichment Separation | Reusable Master Lead enrichment and Project-specific enrichment are separately governed. | APPROVED |
| D-133 | Identity Merge | AI recommends identity matches; deterministic identity services perform merges. | APPROVED |
| D-134 | Enrichment Evidence | Enriched facts retain provenance, freshness, confidence, and observed/inferred status. | APPROVED |
| D-135 | Purpose-Limited Enrichment | Enrichment is limited to identity, fit, contactability, qualification, and sales needs. | APPROVED |
| D-136 | Re-Enrichment | Lead Strategist may request targeted additional enrichment when needed. | APPROVED |
| D-137 | Candidate Multiplier | The configurable candidate multiplier applies to candidate-based discovery, not automatically to Meta inbound volume. | APPROVED/MODIFIED |
| D-138 | Fresh Discovery | Lead Discovery executes fresh search first under current policy. | APPROVED |
| D-139 | Buyer Influence | B2B search targets practical adoption/purchase influencers. | APPROVED |
| D-140 | Discovery Provenance | Every discovery result retains source and query/run provenance. | APPROVED |
| D-141 | Source Planning | Lead Discovery determines Web, Meta, or Both based on audience/product/market. | APPROVED |
| D-142 | Source-Specific Plans | Lead Acquisition Plan contains separate Web and Meta execution sections. | APPROVED |
| D-143 | Discovery/Campaign Boundary | Lead Discovery owns acquisition-source intelligence; Campaign Agent owns executable Meta campaign preparation. | APPROVED |
| D-144 | Campaign Agent Ownership | Campaign Agent owns intelligent campaign preparation, publication orchestration, monitoring, and recommendations. | APPROVED |
| D-145 | Creative Assets | MVP user supplies final campaign image/video assets. | APPROVED |
| D-146 | Campaign Advice | Campaign Agent may recommend changes but may not autonomously increase spend or extend campaigns. | APPROVED |
| D-147 | Publish Authority | Campaign publishing requires explicit authorized user action. | APPROVED |
| D-148 | Lead Capture Design | Campaign Agent recommends form/path and questions; human approves. | APPROVED/MODIFIED |
| D-149 | Campaign Validation | Campaigns undergo AI commercial review and deterministic technical/provider validation. | APPROVED |
| D-150 | Campaign Learning | Campaign results may recommend Strategy changes but cannot rewrite Strategy automatically. | APPROVED |
| D-151 | Creative Guide | Campaign Agent provides detailed step-by-step media production instructions. | APPROVED |
| D-152 | Creative Review | Campaign Agent reviews uploaded creative against Strategy, product, and brief. | APPROVED |
| D-153 | Strategy Agent Trigger | Strategy runs after approved understanding and explicit user initiation/revision. | APPROVED |
| D-154 | ICP/Persona | B2B ICP organization and buying personas are separate; B2C uses consumer personas. | APPROVED |
| D-155 | Structured Strategy | One user-facing Strategy package contains structured downstream-consumable components. | APPROVED |
| D-156 | Controlled Revision | Strategy revisions apply requested changes and re-evaluate dependent sections. | APPROVED |
| D-157 | Strategy Completeness | Missing critical data blocks Strategy without consuming revision allowance. | APPROVED |
| D-158 | Project Understanding Trigger | Project Understanding runs through analysis, clarification, review, and approval. | APPROVED |
| D-159 | Strategy-Critical Facts | Project Understanding distinguishes strategy-critical baseline from sales-support content. | APPROVED |
| D-160 | Dynamic Fields | Required Project/product information is dynamically determined by context. | APPROVED |
| D-161 | User Corrections | Authorized corrections override AI inference in the approved baseline and remain audited. | APPROVED |
| D-162 | Understanding Version | Approved Project/Product Understanding is versioned and referenced by Strategy. | APPROVED |
| D-163 | Company Agent Trigger | Company Understanding runs at onboarding and controlled update. | APPROVED |
| D-164 | Company Dual Outputs | Company Agent produces separate private and reusable outputs. | APPROVED |
| D-165 | Company Update Signals | Projects may suggest Company Understanding updates but cannot overwrite it. | APPROVED |
| D-166 | Company Version References | Company Understanding is versioned and Projects may reference the source version. | APPROVED |
| D-167 | Simulator/Evaluator | Customer Simulator and Evaluator are separate roles. | APPROVED |
| D-168 | Evaluation Gate | Major agent/prompt versions require evaluation before production activation. | APPROVED |
| D-169 | Hard Failures | Critical policy/business violations fail evaluation regardless of average score. | APPROVED |
| D-170 | Version Comparison | Test Lab compares multiple agent/prompt versions on the same scenarios. | APPROVED |
| D-171 | Regression Library | Corrected failures become versioned regression scenarios. | APPROVED |
| D-172 | Application Orchestration | Meaningful agent results pass through application validation/persistence. | APPROVED |
| D-173 | Structured Handoffs | Major agent handoffs use defined structured contracts. | APPROVED |
| D-174 | Context Isolation | Agent context is explicitly assembled and tenant/project scoped. | APPROVED |
| D-175 | Specialist Delegation | Agents may request approved specialists through governed flows. | APPROVED |
| D-176 | Loop Limits | Multi-agent workflows have deterministic loop/retry/time/cost limits. | APPROVED |
| D-177 | Failure Model | Agents use normalized Success/Retryable/Needs Information/Needs Human/Blocked/Failed outcomes. | APPROVED |
| D-178 | Version-Pinned Runs | Every run records Agent, Prompt, model, and configuration versions. | APPROVED |
| D-179 | Governed Tools | Agent tools are registered permissioned application interfaces. | APPROVED |
| D-180 | Interaction Graph | Agent-to-agent invocation is deny-by-default and constrained by an interaction matrix. | APPROVED |
| D-181 | Layered Configuration | Super Admin configures behavior, but mandatory platform safety is not editable prompt content. | APPROVED |
| D-182 | Agent Version Lifecycle | Agent versions use Draft→Testing→Review→Approved→Active→Retired with rollback. | APPROVED |
| D-183 | Business Pool Object | Business Pool contains companies/organizations. | APPROVED |
| D-184 | Master Lead Object | Master Lead Pool contains people/consumer identities. | APPROVED |
| D-185 | Person/Business Relations | People may have multiple historical/current business relationships. | APPROVED |
| D-186 | Tenant Linkage | A tenant Organization may link to a Business Pool entity without exposing tenant-private data. | APPROVED |
| D-187 | Reusable Profiling | Reusable lead profiling requires evidence, confidence, freshness, and observed/inferred status. | APPROVED |
| D-188 | Contextual Signals | Past interactions remain contextual and are not generalized incorrectly. | APPROVED |
| D-189 | Project-Private Intelligence | Detailed qualification, objections, conversations, and negotiation remain private to the owning tenant/project. | APPROVED |
| D-190 | Lead Matching | Master Lead matching is confidence-based and uses multiple identifiers. | APPROVED |
| D-191 | Merge Authority | Deterministic identity services perform merges. | APPROVED |
| D-192 | Source History | Merges preserve all provenance/history. | APPROVED |
| D-193 | Conflicts | Conflicting profile data is retained/resolved rather than silently overwritten. | APPROVED |
| D-194 | Reversible Lead Merge | Important lead merges are audited and reversible. | APPROVED |
| D-195 | Business Identifier | CR/business registration number plus jurisdiction is the strongest company identifier where available. | APPROVED |
| D-196 | Domain Matching | Domain is a strong but non-absolute Business Pool match. | APPROVED |
| D-197 | Company Hierarchy | Parent/subsidiary/branch/brand relationships are supported. | APPROVED |
| D-198 | Non-Disclosing Link | Tenant linkage never exposes private workspace/activity. | APPROVED |
| D-199 | Business Enrichment | Business Pool enrichment is evidence-backed and freshness-aware. | APPROVED |
| D-200 | Reversible Company Merge | Important Business Pool merges are audited and reversible. | APPROVED |
| D-201 | Brand Distinction | Brands are distinguished from legal/business entities where evidence allows. | APPROVED |
| D-202 | Conversion Goal | Every Project has a confirmed conversion goal; MVP final execution is human. | APPROVED/MODIFIED |
| D-203 | Conversion Timing | MVP conversion is normally confirmed after human takeover. | APPROVED/MODIFIED |
| D-204 | Opportunity | Opportunity is optional and used only when a continuing deal record is needed. | APPROVED |
| D-205 | Opportunity Creation | AI may recommend an Opportunity; application rules create it. | APPROVED |
| D-206 | Conversion Event | Conversion is an auditable event with product/source/evidence/human attribution. | APPROVED |
| D-207 | Human Outcome | Human-controlled sales outcomes must be recorded. | APPROVED |
| D-208 | Attribution | Conversion retains best available source/campaign/discovery lineage. | APPROVED |
| D-209 | Final Action | MVP final commercial/conversion actions require human action. | APPROVED |
| D-210 | Project Funnel | Analytics track acquisition through human-confirmed conversion. | APPROVED |
| D-211 | Source Quality | Sources/campaigns are evaluated by downstream lead quality and conversion. | APPROVED |
| D-212 | AI Effectiveness | AI performance is measured by commercial progression, not message count. | APPROVED |
| D-213 | Handover Analytics | Human response/outcome after handover is measured separately. | APPROVED |
| D-214 | Learning Levels | Learning is separated into Lead, Project, and future governed Platform levels. | APPROVED |
| D-215 | Strategy Recommendations | Project Strategy changes require evidence and human approval. | APPROVED |
| D-216 | Pattern Learning | Project learning requires meaningful patterns, not isolated outcomes. | APPROVED |
| D-217 | Admin Separation | Platform CP and Organization Administration are separate scopes. | APPROVED |
| D-218 | Platform Privileges | Super Admin and Platform Admin/Support privileges are separated. | APPROVED |
| D-219 | Support Access | Private tenant access requires controlled, purpose-limited, audited support access. | APPROVED |
| D-220 | Integration Scope | Integrations are Platform, Platform-Managed Organization, or Organization Self-Managed. | APPROVED |
| D-221 | Per-Agent Model | Super Admin selects provider/model/credential reference and prompt/version per agent. | APPROVED |
| D-222 | Agent Observability | Super Admin CP shows agent status, runs, tools, failures, versions, latency, and usage. | APPROVED |
| D-223 | Sensitive Logs | Operational telemetry is separated from permission-controlled sensitive execution content. | APPROVED |
| D-224 | Registration Management | Super Admin CP manages registration applications. | APPROVED |
| D-225 | Plan Authority | Super Admin controls Plan/subscription assignment. | APPROVED |
| D-226 | Overrides | Organization-specific limit/entitlement overrides are supported. | APPROVED |
| D-227 | Suspension | Organization suspension is non-destructive. | APPROVED |
| D-228 | Kill Switch | Super Admin can disable external side effects per Organization. | APPROVED |
| D-229 | Plan Model | Plan separates commercial information, entitlements, limits, and usage. | APPROVED |
| D-230 | Lead Limit | Project Lead Limit is per Project. | APPROVED |
| D-231 | Revision Limit | Strategy Revision Allowance is per Project. | APPROVED |
| D-232 | Overflow Leads | Overflow leads may enter Master Pool but not Project autonomous workflow. | APPROVED |
| D-233 | Existing Lead Continuity | Reaching lead limit does not stop existing Project Leads. | APPROVED |
| D-234 | Entitlement Readiness | Commercial entitlement and integration readiness are separate. | APPROVED |
| D-235 | Subscription Snapshot | Plan definition changes do not silently rewrite active subscriptions. | APPROVED |
| D-236 | Manual Billing | MVP subscription administration is manual; automated billing is deferred. | APPROVED |
| D-237 | Unlimited Limits | Each Plan limit can independently be numeric or Unlimited. | APPROVED |
| D-238 | Project Access | Organization users receive explicit Project access. | APPROVED |
| D-239 | Sensitive Permissions | High-risk actions require explicit permission. | APPROVED |
| D-240 | Fixed Roles | Fixed functional roles are the authorization model. | SUPERSEDED |
| D-241 | Direct Permissions | Non-owner users receive explicit permissions plus Project access. | APPROVED |
| D-242 | Permission Presets | Presets are optional setup shortcuts, not the authorization model. | APPROVED |
| D-243 | Organization Owner | Owner retains account-level authority and protected controls. | APPROVED |
| D-244 | Action Center | Action Center is a permission-aware unresolved work queue. | APPROVED |
| D-245 | Deep Links | Action items deep-link to the exact resolution workflow. | APPROVED |
| D-246 | Action Authority | Agents recommend; application services create/manage authoritative action items. | APPROVED |
| D-247 | Notification Separation | Notifications alert; Action Center remains the unresolved-work source of truth. | APPROVED |
| D-248 | Notification Routing | Notifications are filtered by permission and Project access. | APPROVED |
| D-249 | Notification Channels | MVP notification channels are in-app and email. | APPROVED/MODIFIED |
| D-250 | Notification Privacy | Email notifications contain minimum necessary sensitive information. | APPROVED |
| D-251 | Usage Visibility | Authorized Organization users can view effective Plan/limits/usage. | APPROVED |
| D-252 | Plan Change Authority | Only Super Admin changes the assigned Plan in MVP. | APPROVED |
| D-253 | Limit Warnings | Organization UI explains approaching/reached limits and behavior. | APPROVED |
| D-254 | Organization Defaults | Organization communication defaults and mandatory restrictions are inherited by Projects. | APPROVED |
| D-255 | Notification Preferences | Normal operational email preferences are configurable; mandatory alerts remain. | APPROVED |
| D-256 | Closure Request | Customer requests Organization closure; Super Admin executes controlled closure. | APPROVED |
| D-257 | Organization Audit | Organization Admin CP exposes tenant-scoped important activity. | APPROVED |
| D-258 | Audit Authority | Application services create authoritative audit events. | APPROVED |
| D-259 | Audit Retention | User removal does not remove historical audited actions. | APPROVED |
| D-260 | MVP Integrations | MVP requires real Meta Ads, WhatsApp, Instagram, Facebook Messenger, Web/Google discovery, and email. | APPROVED |
| D-261 | Generic Email | MVP uses a generic business mailbox with outbound and inbound capability. | APPROVED |
| D-262 | Email AI | Email uses the same autonomous active-conversation and approval rules as other channels. | APPROVED |
| D-263 | Sales Identity | Organization configures the Sales AI sender identity; AI cannot invent it. | APPROVED |
| D-264 | Email Failure | Bounce/failure is logged, alerted, actioned, and prevents blind continued sending. | APPROVED |
| D-265 | Attachments | AI may send approved requested/relevant product attachments/content. | APPROVED |
| D-266 | Meta Destination | Primary paid Meta conversation destination for MVP is WhatsApp. | APPROVED |
| D-267 | Channel Transition | Channel transitions require configured, healthy, eligible, approved destination and context. | APPROVED WORKING RULE |
| D-268 | Missing Integration | Missing integration does not block onboarding; it blocks execution and creates action required. | APPROVED |
| D-269 | Channel Fallback | A healthy approved alternative channel may be used when the primary channel fails. | APPROVED WORKING RULE |
| D-270 | Channel Pause | Affected sales activity pauses when no valid fallback exists. | APPROVED |
| D-271 | Offers/Coupons | Coupons/discount automation is deferred from MVP; such requests trigger human handling. | APPROVED/DEFERRED |
| D-272 | Meta Overflow | Meta may continue receiving leads after limit; overflow enters Master Pool but autonomous Project sales stops at the limit. | APPROVED/MODIFIED |
| D-273 | Project Lifecycle | MVP Project states are ONBOARDING→LIVE→ENDED; ENDED means the customer closed the sales initiative. | APPROVED |
| D-274 | Campaign Lifecycle | Use DRAFT→READY→PUBLISHING→ACTIVE↔PAUSED→ENDED with PUBLISH_FAILED. | REQUIRES CONFIRMATION |
| D-275 | Ended Campaign | Ended campaigns are historical; duplicate to a new Draft; existing lead conversations continue. | APPROVED |
| D-276 | Claim Handover | Eligible users see handover and first successful claimant becomes the human owner. | APPROVED |
| D-277 | End-to-End Admin Observability | Super Admin can observe onboarding, lead generation, campaign configuration, chats, handovers, and conversions subject to privacy controls. | APPROVED |
| D-278 | B2C Independence | B2C qualification must not require employer, role, company size, procurement, or buying committee. | APPROVED |
| D-279 | Organization Dashboard | Organization Dashboard is action-led: create Project, complete product content, then start recommended acquisition. | APPROVED WORKING DECISION |
| D-280 | Unified Sales Workspace | MVP combines lead queue, conversation, AI Sales Insight, approvals, and takeover controls. | APPROVED |
| D-281 | Project Reopen | Reopening an ENDED Project is deferred. | DEFERRED |
| D-282 | MVP Acceptance | MVP proves a clean intuitive full cycle from onboarding to lead acquisition, AI sales, human handover, and recorded conversion. | APPROVED |
| D-283 | Prototype Scope | First UX prototype covers company onboarding through lead conversation. | APPROVED |

## 10. Contradictions Resolved

| Conflict/Tension | Resolution |
| --- | --- |
| One-week full MVP vs product breadth | The one-week hard build assumption is superseded. Delivery is iterative; the first implementation target is an HTML/CSS prototype in approximately three days. |
| Company Understanding vs product knowledge | Company Understanding describes the business and market context. Project/Product Understanding owns detailed products, prices, commercial constraints, and Strategy inputs. |
| Universal B2B/B2C channel mapping | B2B/B2C is an input to Strategy, not a hard-coded channel decision. Lead Discovery recommends Web, Meta, or Both. |
| Temporary Lead Pool vs persistent data asset | All captured contactable people enter the Master Lead Pool. Non-contactable results remain Research Candidates. |
| One universal lead score | Initial scoring is Project-specific; observed conversation evidence dynamically supersedes uncertain pre-contact assumptions. |
| Rigid follow-up schedule | Follow-up strategy is contextual and AI-designed; normal sends require human approval unless explicitly delegated per Project Lead. |
| Fixed organization roles | Fixed functional roles are superseded by direct permissions plus Project access, with Organization Owner as a protected authority. |
| ADK as application architecture | ADK is the Agent Component. Deterministic application services own business state, security, entitlements, suppression, audit, and provider side effects. |
| Opportunity required for every conversion | Opportunity is optional. Conversion is a Project-specific auditable success event; all final MVP actions are human. |
| Meta campaign stops at lead limit | Campaign may keep receiving leads; overflow enters Master Pool but does not enter autonomous Project sales until Super Admin increases the limit. |

## 11. Open Questions

| ID | Area | Unresolved matter |
| --- | --- | --- |
| OQ-001 | Campaign lifecycle D-274 | The recommended lifecycle is defined but still requires explicit confirmation. |
| OQ-002 | Plan prices and numeric allowances | Commercial values are intentionally not yet set. |
| OQ-003 | Generic inbound mailbox implementation | The product requirement is generic inbound/outbound mailbox support; exact IMAP/provider compatibility belongs to integration design. |
| OQ-004 | Provider and legal policy verification | Meta, WhatsApp, Instagram, Facebook, Google/Search, email, prospect data, privacy, and Oman/GCC obligations require current official research and legal review. |
| OQ-005 | Full Arabic/RTL UI baseline | The exact degree of UI localization, dialect behavior, and RTL coverage remains to be locked. |

## 12. Deferred Decisions

| ID | Decision | Reason/Scope | Target |
| --- | --- | --- | --- |
| DD-001 | Detailed Test Mode product scope | Detailed Organization/Project Test Mode UX and exact side-effect simulation behavior. | Stage 3 PRDs and Stage 4 Test Architecture |
| DD-002 | Knowledge precedence/conflict resolution | Exact rule hierarchy for conflicting Organization, Project, product, source, and lead knowledge. | Knowledge PRD and System Design |
| DD-003 | File/document lifecycle | Supported types, replacement, source deletion, versioning, failed extraction, malicious-file UX. | Company/Project Understanding PRDs and System Design |
| DD-004 | Final glossary and formal non-goals polish | A working terminology baseline exists in the BRD/PRD; final glossary review remains. | Stage 1 closeout / Master Blueprint |
| DD-005 | Coupons and automated special offers | Not in MVP; requests for special commercial terms require human takeover. | Post-MVP |
| DD-006 | Reopen ENDED Project | MVP does not reopen ended Projects. | Post-MVP |
| DD-007 | Automated billing | Plan assignment/subscription status are managed manually by Super Admin. | Post-MVP |
| DD-008 | Custom roles | Direct permissions are used; optional presets may be added. | Post-MVP |
| DD-009 | Full Arabic/RTL application scope | English/Arabic Sales AI behavior is required; exact full UI localization scope requires a dedicated specification. | UX/Localization PRD |
| DD-010 | Advanced platform-level learning | Cross-Project learning requires privacy/legal/governance design. | Post-MVP / Legal Review |

## 13. Risk Register

| Risk ID | Risk | Severity | Impact | Current treatment |
| --- | --- | --- | --- | --- |
| R-001 | External provider approval/policy | High | MVP requires six real integrations. Account review, permissions, templates, API availability, or policy changes may block release. | Official provider verification; adapters; sandbox; readiness checks; action/failure UX. |
| R-002 | Master Lead/Business Pool privacy | Critical | Cross-tenant reuse can expose private or impermissibly collected data. | Strict reusable/private separation, source provenance, legal review, scoped suppression, deletion/correction design. |
| R-003 | Email deliverability and consent | High | Generic outbound email can create spam, reputation, and legal issues. | Eligibility service, rate controls, bounce handling, suppression, sender/domain readiness, official legal review. |
| R-004 | AI hallucination/unauthorized terms | Critical | Sales AI could invent product facts, price, discount, or commitment. | Approved knowledge only, tool/service checks, handover, evaluation hard failures, audit. |
| R-005 | Tenant leakage in agent context | Critical | Agent context, retrieval, memory, or logs could mix tenants/projects. | Explicit context assembly, scoped tools, tenant tests, sensitive-log controls, no unrestricted DB access. |
| R-006 | Human approval bottleneck | Medium | Follow-up approvals and final handovers may delay sales. | Action Center, in-app/email alerts, delegation option, handover claim flow, response analytics. |
| R-007 | Campaign spend risk | High | Incorrect campaign settings or duplicate publication can create real spend. | Explicit Publish, deterministic validation, idempotency, spend controls, mandatory end date, audit. |
| R-008 | Lead identity false merges | High | Incorrect merges can corrupt profiles and reuse signals. | Confidence-based matching, deterministic merge rules, provenance, reversible audited merges. |
| R-009 | Agent complexity/cost/latency | High | Many agents can create loops and excessive calls. | Deny-by-default interaction graph, bounded loops, cost/time limits, deterministic services, evaluation. |
| R-010 | Scope expansion | High | MVP could drift into CRM, billing, contracting, or broad automation. | Locked MVP boundary, deferred register, change control, milestone acceptance. |

## 14. Documents Updated or Required

The following artifacts are produced from this baseline:

1. `AI_Sales_Agent_Stage_1_Lock_Report_v1.0.md`
2. `AI_Sales_Agent_MVP_BRD_v1.0.md`
3. `AI_Sales_Agent_MVP_Master_PRD_Blueprint_v1.0.md`

The following are still required before implementation begins:

- Domain-level PRD refinements where the Master PRD delegates detailed provider or file behavior.
- Current official provider-policy and legal/compliance research.
- System context, domain architecture, data model, ERD, API inventory, tool contracts, provider-adapter specifications, security architecture, and deployment architecture.
- Milestone roadmap, test strategy, Definition of Ready, Definition of Done, and implementation prompts.

## 15. Traceability

- Approved product decisions are consolidated into business requirements (`BR-*`) in the BRD.
- Business requirements are translated into prescriptive product requirements (`PR-*`), state models, workflows, screens, agent contracts, and acceptance criteria in the Master PRD.
- The earlier uploaded documents remain reference inputs and are lower in authority than the approved baseline.
- Stage 4 system/API/data specifications must trace back to the BRD and PRD; no protected behavior may be invented during coding.

## 16. Stage Acceptance Checklist

- [x] Product vision and positioning defined
- [x] Organization, Project, Product, Master Lead, Business Pool, Campaign, Conversation, and Conversion boundaries defined
- [x] Company and Project onboarding operating models defined
- [x] B2B/B2C differences defined and B2C path validated
- [x] MVP acquisition/channel boundary defined
- [x] AI autonomy, follow-up approval, takeover, and final-action boundary defined
- [x] Super Admin and Organization Administration concepts defined
- [x] Plans, limits, unlimited configuration, and direct permissions defined
- [x] Agent set, ADK boundary, collaboration, tools, versioning, and evaluation defined
- [x] Master Lead/Business Pool reusable/private-data separation defined
- [x] Dashboard, Action Center, notifications, Sales Workspace, and prototype boundary defined
- [ ] Campaign lifecycle explicitly confirmed
- [ ] Detailed Test Mode scope locked
- [ ] Knowledge conflict precedence locked
- [ ] File/document lifecycle locked
- [ ] Provider/legal policy verification completed

## 17. Recommendation

The product operating model is sufficiently mature to serve as the baseline for formal business and product requirements. The remaining unresolved matters are either explicitly deferred or clearly identified.

**Recommendation:** approve this Stage 1 report as the product baseline, confirm or modify D-274, and use the accompanying BRD and Master PRD as the authoritative product-requirements source for the first MVP. Do not begin production coding until the architecture, API, data, security, provider, and test specifications required by Stage 4 are completed and approved.


---

**Stage approval statement:** `Stage 1 — Product Idea Redesign and Operating Model is locked.`


# Approved Super Admin Control Panel Amendment — 2026-08-18

> **Authority:** This amendment contains explicitly approved Super Admin MVP decisions from the implementation-design sessions after v1.0. Where this amendment conflicts with earlier Super Admin/Admin/Support wording in this document, this amendment supersedes the earlier wording for the MVP.

## A. Platform Administration Actor

- MVP has one internal platform actor only: **Super Admin**.
- Platform Support, Platform Admin/Support, separate Operations roles, and separate Agent Administrator roles are removed from the MVP control-panel model.
- Super Admin is the sole internal actor for the Super Admin Control Panel.

## B. Super Admin Responsibilities

Super Admin may:
- View platform logs and statistics.
- Create, activate, edit, and logically delete Organization accounts.
- Review public Organization applications and approve or reject them.
- Approve a previously rejected application later without requiring a new application.
- Assign Packages to Organizations.
- Create and configure Packages.
- Increase Organization-specific numeric limits through overrides.
- View AI Agent execution cost and consumption per Organization.
- Manage core platform integrations.
- Configure AI Agents and Agent Versions.
- Manage Organization Meta Ads, WhatsApp, Instagram, and Facebook/Messenger integrations.
- View System Health and operational issues.
- View immutable Platform Audit history.

## C. Super Admin Authentication

- Super Admin authentication is **email + password + email OTP**.
- OTP is 6 digits.
- OTP validity: 10 minutes.
- Maximum failed OTP attempts per challenge: 5.
- OTP resend cooldown: 60 seconds.
- Resending invalidates the previous OTP.
- Successful OTP verification creates the authenticated Super Admin session.
- Logout revokes the active session server-side.
- Direct access to OTP verification without a valid password-authentication challenge redirects to login.
- Authentication email delivery must have an emergency/deployment-level recovery mechanism so a provider configuration failure cannot permanently lock out the sole Super Admin.

## D. Super Admin Dashboard

The Super Admin Dashboard is action-led and contains:
1. Needs Attention
2. AI Cost & Consumption
3. Organizations
4. System Health
5. Recent Important Activity

Rules:
- The Dashboard is not a generic KPI dashboard.
- `Create Organization` is the principal creation action.
- Organization mutations occur inside Organization context, not directly from Dashboard cards.
- AI Cost on the Dashboard refers to actual Agent execution cost, not Strategy Credits.
- System Health exposes product-service health rather than low-level infrastructure noise.
- Dashboard sections may fail independently where technically possible.
- Pending public applications may surface in Needs Attention.

## E. Organization Entry Paths

Two Organization entry methods are supported:
1. **Public Application**
2. **Manual Super Admin Creation**

### Public Application
Lifecycle:
`SUBMITTED -> UNDER_REVIEW -> APPROVED`
or
`SUBMITTED -> UNDER_REVIEW -> REJECTED -> APPROVED`

Rules:
- No `INFORMATION_REQUIRED` state in MVP.
- No Request Information workflow.
- Public approval is `Approve & Activate`.
- Rejected applications remain historical and can later be approved.
- Same CR/business registration number + jurisdiction blocks duplicate Organization creation.
- Duplicate domain is allowed.
- Similar company name is informational only.
- Existing Owner email is allowed; existing User identity is reused.
- Approval is idempotent.

### Manual Creation
- Manual creation creates an `INACTIVE` Organization first.
- Activation is a separate explicit Super Admin action.
- Super Admin never creates or sees a customer password.
- New Owner receives a secure invitation to establish credentials.
- Existing User identity may be linked as Owner.
- Only active Packages can be assigned.
- Numeric limit overrides are applied after Organization creation, not in the creation form.
- Activation revalidates critical identity/readiness rules.
- Owner invitation may remain pending while the Organization is ACTIVE; resend invitation is available.

## F. Organization Lifecycle

MVP Organization lifecycle is exactly:
`INACTIVE -> ACTIVE -> DELETED`

Also allowed:
`INACTIVE -> DELETED`

Rules:
- No Suspension state in MVP.
- No Restore transition in MVP.
- `DELETED` is logical deletion and is terminal in MVP.
- Deleted Organization is read-only historical context.
- Deletion requires a reason.
- Deletion blocks customer access and new external/commercial operations.
- Pending outbound actions are cancelled or made obsolete.
- Historical Organization, Project, campaign, AI cost, integration, and audit history remain.
- Master Lead and Business Pool reusable identities remain.
- Organization-private Project Lead intelligence remains private and does not become reusable because of deletion.
- Deletion must block new external side effects before or atomically with the logical deletion operation.
- Partial external shutdown failure creates an actionable operational warning.
- Organization Owner receives a closure notification.
- Normal Organization editing is available only in INACTIVE and ACTIVE states.
- Owner identity/email changes are handled separately from ordinary account editing.

## G. Organization Information Architecture

Organization context inside Super Admin contains:
- Overview
- Account
- Package & Limits
- AI Usage & Cost
- Integrations
- Projects & Activity
- Audit

Organizations List:
- Deleted Organizations are excluded from the normal default list but remain searchable/filterable.
- Mutations such as Package change, limit override, and Delete occur inside Organization context.
- Organization Overview shows only operationally useful summaries: Package, Projects, Lead Usage, Strategy Credits, AI Cost, Integration status.
- Organization Overview shows an actionable Needs Attention section only when intervention is required.

## H. Packages, Strategy Credits, and Numeric Limits

Super Admin UI terminology is **Package**. Existing `Plan` references map to Package.

Package lifecycle:
`INACTIVE <-> ACTIVE`
Packages are not deleted in MVP.

Package includes:
- Maximum Projects
- Strategy Credits per Project
- Lead Limit per Project
- Approved feature entitlements

Each numeric limit independently supports:
- NUMERIC
- UNLIMITED

Rules:
- Package creation starts INACTIVE.
- Meaningful Package edits create immutable Package Versions.
- Existing Organization Package snapshots never silently change when a Package is edited.
- Strategy Credits mean user-requested Strategy regeneration/revision allowance per Project.
- Initial Strategy generation and missing-information/clarification correction do not consume Strategy Credits.
- AI operational cost is separate from Strategy Credits.
- Organization numeric overrides replace the Package value for that Organization; they are not arithmetic additions.
- Numeric overrides support only:
  - Maximum Projects
  - Strategy Credits per Project
  - Lead Limit per Project
- Feature entitlement overrides are not supported in MVP.
- To gain a feature not included in the current Package, Super Admin must change the Package.
- Package change creates a new Organization Package snapshot.
- Package change clears old Organization-specific overrides.
- Package change requires impact preview.
- Existing Projects/Project Leads continue when a reduced limit falls below current usage; new creation/allocation is blocked.
- Strategy Credit limit cannot be reduced below already-consumed revision usage.
- Deleted Organizations show Package and limit history read-only.

## I. Sales AI Cost & Consumption

Primary internal metric: **Sales AI Cost**.

Sales AI Cost includes production Agent execution for:
- Sales / Conversation Agent
- Lead Strategist Agent
- Lead Enrichment Agent when invoked as part of the active Project Lead sales process

The primary Sales AI Cost metric excludes:
- Company Understanding
- Project Understanding
- Strategy generation
- Campaign Agent
- Customer Simulator
- Evaluator

Rules:
- Every costed Agent run is attributable to Organization, Project, Project Lead where applicable, Agent, Agent Version, model, provider, time, usage, and cost.
- Historical cost is stored with its pricing basis and is not silently recalculated when provider pricing changes.
- Failed/retried Agent runs count toward cost when billable resources were consumed.
- Unpriced runs are marked `UNPRICED`, never treated as zero.
- Platform AI Usage provides Organization, Agent, and time-based cost breakdown.
- Organization AI Usage provides Project and Agent breakdown.
- Internal monetary AI cost is Super Admin-only in MVP.
- Automated cost anomaly detection is deferred.
- AI Cost views link to Agent Runs/Logs for diagnosis.

## J. Organization Integrations

Super Admin-managed Organization integrations are exactly:
- Meta Ads
- WhatsApp
- Instagram
- Facebook/Messenger

Rules:
- Each is tracked as a separate logical capability.
- Package entitlement and integration readiness are separate concepts.
- Integration states:
  - NOT_CONFIGURED
  - CONFIGURING
  - CONNECTED
  - PARTIAL
  - ERROR
  - REAUTH_REQUIRED
  - DISABLED
- Credentials stored does not equal CONNECTED; required capability tests must pass.
- Secrets may be entered/replaced by Super Admin but are never redisplayed as plaintext.
- Organization users may see readiness status but never provider secrets.
- Disabling an integration blocks new affected operations while preserving configuration/history.
- Re-enabling requires fresh connection testing.
- A failed integration blocks only dependent capabilities.
- Removing feature entitlement via Package change preserves connection history but blocks use.
- Re-entitling a feature later requires testing before use.
- Provider-specific field contracts are deferred to the provider integration specification.

## K. Core Platform Integrations

Core Platform Integrations are separate from Organization Integrations.

MVP core integrations:
- AI Provider(s)
- Google/Search / Research Provider
- Platform Notification Email
- File/Processing UI only if implementation requires Super Admin-managed external configuration

Rules:
- Common operational states match Organization Integrations.
- AI Provider configuration establishes usable provider/model/credential resources.
- Individual Agent Versions select from approved configured provider/model resources.
- Multiple AI providers are supported by the platform model.
- Fallback provider/model is never arbitrary; it must be explicitly configured for the affected Agent.
- Disabling a Core Integration requires reason and impact preview.
- Super Admin may disable a critical provider when necessary.
- Re-enabling requires fresh testing.
- Core secrets are never redisplayed and are unavailable directly to Agents.
- Core Integration Detail exposes dependent platform functions.
- Provider-specific configuration fields are deferred to provider contracts.

## L. AI Agent Management

MVP Agent Registry is fixed to:
1. Company Understanding
2. Project Understanding
3. Strategy
4. Lead Discovery
5. Lead Enrichment
6. Lead Strategist
7. Sales / Conversation
8. Campaign
9. Customer Simulator
10. Evaluator

Rules:
- Super Admin cannot create/delete Agent types in MVP.
- Agent Definition contains system-controlled purpose, trigger, context scope, allowed architecture, and hard boundaries.
- Super Admin configures behavior through Agent Versions.
- Exactly one Agent Version is ACTIVE per Agent.
- Agent operational status ENABLED/DISABLED is separate from Version lifecycle and calculated health.
- Version lifecycle:
  `DRAFT -> TESTING -> UNDER_REVIEW -> APPROVED -> ACTIVE -> RETIRED`
- Approved, Active, and Retired Versions are immutable.
- New Draft usually clones an existing Version.
- Provider/model/credential options come from Core AI Providers.
- Prompts cannot override protected platform rules.
- Super Admin can view but cannot arbitrarily expand Agent context scope.
- Tools are system-registered; Super Admin may enable/disable only tools permitted for that Agent.
- Agent-to-Agent interactions are deny-by-default and configurable only within system-approved possibilities.
- Meaningful prompt/model/provider/tool/interaction/runtime changes require testing before production activation.
- Activating a Version affects future runs only; in-flight runs remain version-pinned.
- Rollback uses a previously approved Version only after dependency/readiness validation.
- Disabling an Agent stops new executions but preserves Active Version/history.
- Agent disable requires reason and impact preview.
- Draft Versions may be deleted; production/historical Versions remain.

## M. Simple Agent Testing

Every Agent has a **Test Agent** action.

Testing uses one reusable test shell adapted to each Agent's real inputs/outputs.

Tests:
- use real Agent logic
- use TEST environment
- never create production Organizations, Projects, Leads, campaigns, conversations, conversions, or provider side effects
- retain Agent Version, input, output, model, cost, duration, tool activity, and error information

Agent-specific tests:
- Company Understanding: sample company information -> Company Understanding result
- Project Understanding: Company Understanding + Project/product information -> Project/Product Understanding
- Strategy: Company + Project Understanding results -> Strategy
- Lead Discovery: upstream Company/Project/Strategy results -> acquisition/discovery plan
- Lead Enrichment: sample lead + relevant upstream context -> enrichment result
- Lead Strategist: Lead Enrichment + Strategy + lead/persona context -> lead Strategy
- Sales / Conversation: interactive simulated conversation
- Campaign: Project/Strategy/acquisition/product/upstream test inputs -> campaign planning output
- Customer Simulator: selectable/configurable customer persona used primarily in Sales testing
- Evaluator: uploaded/selected Agent test result -> PASS / REVIEW / FAIL

Additional rules:
- Completed Test Results can be selected directly as downstream test inputs; manual upload also supported.
- Test status:
  READY, RUNNING, COMPLETED, FAILED, CANCELLED
- Complex Scenario Library / Release Evaluation / Version Comparison is excluded from MVP.
- A new Agent Version requires at least one successful test before approval.
- Critical safety failure identified by Evaluator blocks production activation until Version changes and is retested.

## N. System Health & Operations

System Health is a product-operations view, not a low-level infrastructure console.

Health states:
- HEALTHY
- DEGRADED
- DOWN
- UNKNOWN

Covers:
- AI Runtime
- Core Providers
- Search/Research
- Meta/platform dependencies
- Notification Email
- Background Jobs
- File Processing where applicable
- Agent Operations

Operational Issues:
- simple `OPEN` / `RESOLVED` records
- same issues feed Dashboard Needs Attention
- resolve automatically only when underlying condition recovers
- Super Admin cannot manually hide an active technical failure
- System Health deep-links to the owning Agent/Integration/Organization/workflow

Retry:
- Super Admin may manually retry only backend-declared retry-safe jobs.
- Uncertain external side effects never receive a generic blind Retry.
- Health data that is stale becomes UNKNOWN, never falsely HEALTHY.
- System Health does not duplicate Agent/Integration kill switches.

## O. Logs & Agent Runs

Logs = operational history.
Audit = Super Admin change history.

Logs:
- levels: INFO, WARNING, ERROR
- shared filtering across Agent, Integration, Job, Messaging, Campaign, Discovery, Notification, System
- Production and Test separated by environment
- Production is default
- immutable from Super Admin UI
- no general raw-log export or true real-time streaming in MVP

Agent Runs:
- dedicated view
- normalized outcomes:
  SUCCESS
  RETRYABLE_FAILURE
  NEEDS_INFORMATION
  NEEDS_HUMAN_REVIEW
  BLOCKED
  FAILED
- exact Agent/Prompt Versions, model/provider, timing, usage, cost, context references, tool calls, and child-Agent calls
- full customer/prompt/response execution content hidden by default
- deliberate reveal creates Audit event
- hidden model chain-of-thought is not required/exposed
- retries and child runs keep individual cost/status and may show workflow total
- Logs do not provide generic business-workflow retries
- exact retention period deferred to data/security specification

## P. Platform Audit

Platform Audit records important Super Admin actions and protected administrative state changes.

Rules:
- one immutable platform-wide Audit dataset
- Organization/resource views reuse the same dataset through filters
- meaningful old/new values or references are stored
- passwords, OTPs, credentials, tokens, and secrets are never stored
- revealing sensitive Agent execution content is audited without copying the sensitive content
- significant failed administrative state-change attempts may be audited
- ordinary field validation errors are not
- Audit Events are created by trusted application services as part of the business operation
- Super Admin cannot edit or individually delete Audit Events
- Deleted Organizations and historical resources retain resolvable Audit references
- exact retention period deferred
- general Audit export excluded from initial MVP

