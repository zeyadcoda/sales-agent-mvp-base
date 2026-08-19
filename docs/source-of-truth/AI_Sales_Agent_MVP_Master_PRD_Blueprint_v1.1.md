# AI Sales Agent — MVP Master Product Requirements & Build Blueprint

| Field | Value |
| --- | --- |
| Version | 1.1 |
| Status | Under Review — v1.0 + Approved Super Admin Amendment |
| Owner | Zeyad Al Gharabi |
| Approver | Project Owner — explicit approval required |
| Created | 2026-08-16 |
| Last Updated | 2026-08-18 |
| Business Baseline | AI_Sales_Agent_MVP_BRD_v1.0.md |
| Decision Baseline | AI_Sales_Agent_Stage_1_Lock_Report_v1.0.md |
| Scope | Detailed product behavior, UX, states, agents, integrations, controls, acceptance, and build governance for first MVP |


> **Authority:** Once approved, this document is the authoritative product-behavior source for the first MVP. Legacy uploaded PRDs are reference material only. This blueprint does not replace the still-required approved data model, API contracts, security architecture, provider-policy specifications, deployment architecture, and milestone roadmap. Where this document marks an item `REQUIRES CONFIRMATION` or `DEFERRED`, coding must not invent the answer.

## Table of Contents

- [1. Source-of-Truth Hierarchy](#1-source-of-truth-hierarchy)
- [2. Product Definition and MVP Promise](#2-product-definition-and-mvp-promise)
- [3. Product Boundaries and Non-Goals](#3-product-boundaries-and-non-goals)
  - [Included](#included)
  - [Not in MVP](#not-in-mvp)
- [4. Actors and Authorization](#4-actors-and-authorization)
  - [Actor model](#actor-model)
  - [Authorization rule](#authorization-rule)
  - [4.1 Permission Catalogue](#41-permission-catalogue)
- [5. Information Architecture and Navigation](#5-information-architecture-and-navigation)
  - [5.1 Screen Inventory](#51-screen-inventory)
- [6. Global Product and UX Rules](#6-global-product-and-ux-rules)
- [7. Business and Product State Models](#7-business-and-product-state-models)
- [8. Business Rule Precedence](#8-business-rule-precedence)
- [9. Product Requirement Index](#9-product-requirement-index)
- [10. Detailed Product Modules](#10-detailed-product-modules)
  - [M01 — Company Registration and Platform Activation](#m01-company-registration-and-platform-activation)
  - [M02 — Company Understanding Onboarding](#m02-company-understanding-onboarding)
  - [M03 — Organization Dashboard and Administration](#m03-organization-dashboard-and-administration)
  - [M04 — Project Creation, Understanding, and Final Onboarding](#m04-project-creation-understanding-and-final-onboarding)
  - [M05 — Project Dashboard and Product Catalogue](#m05-project-dashboard-and-product-catalogue)
  - [M06 — Strategy Planning and Approval](#m06-strategy-planning-and-approval)
  - [M07 — Lead Acquisition Planning](#m07-lead-acquisition-planning)
  - [M08 — Web/Google Lead Discovery](#m08-webgoogle-lead-discovery)
  - [M09 — Meta Campaign Creation, Publication, and Monitoring](#m09-meta-campaign-creation-publication-and-monitoring)
  - [M10 — Channel Integration Readiness and Meta Messaging](#m10-channel-integration-readiness-and-meta-messaging)
  - [M11 — Generic Business Email Channel](#m11-generic-business-email-channel)
  - [M12 — Business Pool, Master Lead Pool, and Project Lead Context](#m12-business-pool-master-lead-pool-and-project-lead-context)
  - [M13 — Lead Enrichment and Identity Resolution](#m13-lead-enrichment-and-identity-resolution)
  - [M14 — Initial Scoring, Dynamic Qualification, and Lead State](#m14-initial-scoring-dynamic-qualification-and-lead-state)
  - [M15 — Dynamic Lead Strategy and Next Best Action](#m15-dynamic-lead-strategy-and-next-best-action)
  - [M16 — Sales Agent and Unified Conversation Workspace](#m16-sales-agent-and-unified-conversation-workspace)
  - [M17 — Follow-Up, Unresponsive Attention, and Conversation Control](#m17-follow-up-unresponsive-attention-and-conversation-control)
  - [M18 — Human Handover, Opportunity, and Conversion](#m18-human-handover-opportunity-and-conversion)
  - [M19 — Project, Campaign, Sales Analytics, and Learning](#m19-project-campaign-sales-analytics-and-learning)
  - [M20 — Plans, Usage, Limits, and Entitlements](#m20-plans-usage-limits-and-entitlements)
  - [M21 — Direct Permissions, Project Access, Action Center, Notifications, and Audit](#m21-direct-permissions-project-access-action-center-notifications-and-audit)
  - [M22 — Super Admin Control Panel and Platform Operations](#m22-super-admin-control-panel-and-platform-operations)
  - [M23 — Google ADK Agent Component, Agent Registry, and Evaluation](#m23-google-adk-agent-component-agent-registry-and-evaluation)
  - [M24 — Security, Privacy, Compliance, and Trust Controls](#m24-security-privacy-compliance-and-trust-controls)
  - [M25 — Global UX, Accessibility, Language, and Prototype](#m25-global-ux-accessibility-language-and-prototype)
- [11. Agent Registry](#11-agent-registry)
- [12. Agent Interaction Matrix](#12-agent-interaction-matrix)
- [13. Agent Run and Contract Standard](#13-agent-run-and-contract-standard)
- [14. Conceptual Data Ownership and Isolation](#14-conceptual-data-ownership-and-isolation)
- [15. Core Entity Dictionary](#15-core-entity-dictionary)
- [16. Integration Registry](#16-integration-registry)
- [17. Integration Failure and Readiness Rules](#17-integration-failure-and-readiness-rules)
- [18. Analytics Event Catalogue](#18-analytics-event-catalogue)
- [19. Non-Functional Requirements](#19-non-functional-requirements)
- [20. Security and Privacy Checklist](#20-security-and-privacy-checklist)
- [21. Test and Evaluation Strategy](#21-test-and-evaluation-strategy)
  - [Test layers](#test-layers)
  - [Critical evaluation hard failures](#critical-evaluation-hard-failures)
  - [21.1 Critical End-to-End Scenarios](#211-critical-end-to-end-scenarios)
- [22. Initial Three-Day UX Prototype Scope](#22-initial-three-day-ux-prototype-scope)
- [23. MVP Acceptance Walkthrough](#23-mvp-acceptance-walkthrough)
- [24. Definition of Done](#24-definition-of-done)
- [25. Requirements Traceability](#25-requirements-traceability)
- [26. Open Questions and Deferred Decisions](#26-open-questions-and-deferred-decisions)
  - [Open](#open)
  - [Deferred](#deferred)
- [27. Required Specifications Before Production Coding](#27-required-specifications-before-production-coding)
- [28. Change-Control Protocol](#28-change-control-protocol)

## 1. Source-of-Truth Hierarchy

1. Approved Master PRD/Build Blueprint and approved change requests.
2. Approved BRD and Stage 1 Decision Baseline.
3. Approved future ADRs, system/data/API/security/integration specifications.
4. Current code/migrations only where consistent with approved specifications.
5. Uploaded initialization and earlier PRDs as reference material.

No editable Agent prompt, UI behavior, or chat history may silently supersede this hierarchy.

## 2. Product Definition and MVP Promise

A customer can move through a complete, traceable sales operation inside the platform:

```text
Apply and be activated
→ Teach the platform about the company
→ Create and approve a Project/product baseline
→ Generate and approve Strategy
→ Configure Sales AI behavior
→ Acquire leads through Meta and/or Web
→ Enrich, score, and dynamically qualify
→ Run natural multi-channel AI sales conversations
→ Approve/delegate follow-ups
→ Claim human handover for final action
→ Record conversion
→ View source-to-conversion analytics
```

The MVP must be clean, intuitive, and usable without requiring the customer to reconstruct the important sales workflow outside the product. All six required channel capabilities are real MVP integrations, while sandbox adapters remain required for development/evaluation safety.

## 3. Product Boundaries and Non-Goals

### Included

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

### Not in MVP

- Full CRM replacement or arbitrary pipeline designer.
- Customer self-service billing/payment gateway.
- Custom role builder; direct permissions are sufficient.
- Voice AI, automated calls, SIP, or outbound calling.
- Autonomous final meeting confirmation, quotation, negotiation, contract, discount, reservation, payment, or sale.
- Invoicing, accounting, contract generation, or complex quotation engine.
- Automated coupons/special-offer execution.
- Reopening ENDED Projects.
- Unrestricted scraping or use of unapproved data sources.
- Unofficial WhatsApp automation.
- Large integration marketplace or generic no-code workflow builder.
- Production-scale microservices/event-streaming complexity before justified.
- Automatic cross-tenant learning from private conversations.

## 4. Actors and Authorization

### Actor model
- Public applicant
- Platform Super Admin
- Platform Admin/Support with explicit restricted grants
- Organization Owner
- Organization user with explicit permissions and Project access
- Lead/prospect
- Google ADK specialist agents
- Deterministic application services
- External provider adapters

### Authorization rule

```text
Authentication
+ Organization membership
+ explicit permission
+ Project access
+ resource scope/ownership
+ Plan entitlement
+ integration readiness
+ current state/action policy
= Allow or Deny
```

The Organization Owner is protected. Optional presets may assign common permissions but are not the authorization model.

### 4.1 Permission Catalogue

| Permission | Purpose |
| --- | --- |
| organization.company_understanding.view/manage/approve | View/manage/approve Company Understanding. |
| organization.team.manage | Invite/remove users within protected Owner rules. |
| organization.permissions.manage | Grant/revoke permitted permissions. |
| organization.project_access.manage | Assign explicit Project access. |
| organization.settings.manage | Edit company settings/defaults. |
| organization.usage.view | View Plan/limits/usage. |
| organization.audit.view | View tenant audit. |
| integration.email.manage | Configure generic mailbox and Sales identity. |
| project.create/view/end | Create, view, or end Project. |
| project.onboarding.manage/approve | Complete and approve Project onboarding. |
| product.content.view/manage | View/manage sales-support content. |
| strategy.view/generate/revise/approve | Run and approve Strategy workflow. |
| acquisition.plan.view/manage/approve | Review/approve acquisition plan. |
| discovery.run/view/stop | Operate Web discovery. |
| campaign.view/prepare/publish | Prepare and explicitly publish campaign. |
| campaign.increase_budget/extend/pause_resume | Perform real campaign operational changes. |
| lead.view | View Project Lead details. |
| conversation.view/send/takeover/return_to_ai | Operate conversation within control rules. |
| followup.approve/delegate | Approve or explicitly delegate AI follow-up. |
| handover.claim/reassign | Claim or administratively reassign handover. |
| opportunity.manage | Manage optional Opportunity. |
| conversion.record/correct | Record or controlled-correct outcome. |
| analytics.view | View permitted Project/campaign/sales analytics. |
| platform.registration.manage | Review/approve/reject registrations. |
| platform.organization.manage | Plan, status, suspension, closure, overrides. |
| platform.integration.manage | Platform and Organization provider configuration. |
| platform.agent.manage | Agent/model/prompt/tool/version administration. |
| platform.evaluation.run/approve | Run evaluations and approve production eligibility. |
| platform.trace.view_sensitive | View restricted execution content with reason/audit. |
| platform.support.access | Open controlled tenant support access. |
| platform.audit.view | View platform audit. |


## 5. Information Architecture and Navigation

```text
Public
├── Landing / Plans / Registration / Application Status / Login
Organization
├── Dashboard
├── Company Understanding
├── Team & Access
├── Projects
├── Email Integration
├── Action Center / Notifications
├── Usage & Plan
├── Settings
└── Audit
Project
├── Dashboard
├── Products
├── Strategy
├── Lead Acquisition
├── Campaigns
├── Leads
├── Sales Inbox / Conversation
├── Handovers
├── Opportunities / Conversions
├── Analytics
└── Settings
Super Admin
├── Dashboard / Registration / Organizations
├── Plans / Overrides
├── Organization and Platform Integrations
├── Agents / Versions / Prompts / Models / Tools
├── Agent Runs / Test Lab / Evaluations
├── Operations / Errors / Health
├── Support Access
└── Audit
```

### 5.1 Screen Inventory

| Screen ID | Screen | Primary actor | Purpose |
| --- | --- | --- | --- |
| PUB-001 | Landing Page | Public visitor | Explain value and direct to registration/login. |
| PUB-002 | Plan Selection | Applicant | Display active Plans and select one by ID. |
| PUB-003 | Company Registration | Applicant | Submit company/admin/digital presence information. |
| PUB-004 | Application Submitted | Applicant | Show pending review and support guidance. |
| PUB-005 | Application Status / Information Required | Applicant | Provide status and requested next step. |
| AUTH-001 | Login | All users | Authenticate and choose Organization when multiple memberships exist. |
| AUTH-002 | Recovery / Reset | User | Recover account through secure flow. |
| ORG-001 | Organization Dashboard | Organization users | Action-led Organization overview and next steps. |
| ORG-002 | Company Understanding Input | Owner/permitted user | Provide free text and company documents. |
| ORG-003 | Company Understanding Analysis | Owner/permitted user | Show progress and source processing. |
| ORG-004 | Initial Understanding Review | Owner/permitted user | Review facts/inferences/conflicts before questions. |
| ORG-005 | Clarification Conversation | Owner/permitted user | Answer targeted Agent questions. |
| ORG-006 | Final Business Overview | Owner/permitted user | Review and approve versioned overview. |
| ORG-007 | Team & Access | Owner/permitted user | Invite users, grant permissions, assign Projects. |
| ORG-008 | Email Integration | Permitted user | Configure/test generic mailbox and Sales identity. |
| ORG-009 | Usage & Plan | Permitted user | Read-only effective entitlements, limits, usage, overrides. |
| ORG-010 | Organization Settings | Permitted user | Company contact/preferences/default restrictions/notifications. |
| ORG-011 | Organization Audit | Permitted user | Review tenant-scoped important activity. |
| ORG-012 | Action Center | Permitted users | Resolve permission-aware operational work. |
| ORG-013 | Notifications | Permitted users | View in-app alerts and link to action. |
| PRJ-001 | Create Project | Permitted user | Create Project shell. |
| PRJ-002 | Project Brief & Sources | Permitted user | Describe Project/products and upload materials. |
| PRJ-003 | Project Understanding Analysis | Permitted user | Show AI processing and missing information. |
| PRJ-004 | Project/Product Understanding Review | Permitted user | Review/correct/approve baseline. |
| PRJ-005 | Strategy Generation | Permitted user | Run Strategy Agent and show progress/errors. |
| PRJ-006 | Strategy Review | Permitted user | Review complete package, revise, approve. |
| PRJ-007 | Channel & Sales AI Settings | Permitted user | Configure language/tone/restrictions/instructions. |
| PRJ-008 | Final Project Review | Permitted user | Approve Project onboarding into LIVE. |
| PRJ-009 | Project Dashboard | Project users | Show status, statistics, and required actions. |
| PRJ-010 | Product Catalogue | Project users | View generated products and content completeness. |
| PRJ-011 | Product Content Detail | Permitted user | Add descriptions/media/docs without editing baseline. |
| ACQ-001 | Lead Acquisition Plan | Permitted user | Review Web/Meta/Both recommendation and approve. |
| WEB-001 | Web Discovery Plan | Permitted user | Review criteria, queries, target, sources, exclusions. |
| WEB-002 | Discovery Run Progress | Permitted user | Monitor fresh discovery and stop/review. |
| WEB-003 | Discovery Results | Permitted user | View candidates, contactable leads, duplicates, allocation. |
| CAM-001 | Campaign Recommendations | Permitted user | Select AI-suggested campaign concept. |
| CAM-002 | Campaign Configuration | Permitted user | Review/edit audience, geography, budget, end date, channels. |
| CAM-003 | Lead Capture Review | Permitted user | Review/approve Meta lead path/form questions. |
| CAM-004 | Creative Production Guide | Permitted user | Follow detailed media creation instructions. |
| CAM-005 | Media Upload & AI Review | Permitted user | Upload final assets and resolve review findings. |
| CAM-006 | Campaign Final Review | Permitted user | Review readiness and explicitly publish. |
| CAM-007 | Published Campaign | Permitted user | Monitor status, metrics, generated leads, recommendations. |
| CAM-008 | Campaign Failure/Correction | Permitted user | Understand provider failure and retry safely. |
| LED-001 | Leads List | Project users | Search/filter Project Leads and statuses. |
| LED-002 | Lead Profile | Project users | View permitted Master Lead, Business relation, enrichment, source. |
| SAL-001 | Unified Sales Workspace | Project sales users | Lead queue + conversation + AI insight + controls. |
| SAL-002 | Follow-Up Approval | Permitted user | Approve/edit/reject/replan/delegate a follow-up. |
| SAL-003 | Handover Queue | Permitted user | View and claim human-ready conversations. |
| SAL-004 | Conversion / Outcome | Permitted user | Record final human outcome with evidence. |
| ANL-001 | Project Analytics | Permitted user | View acquisition-to-conversion funnel. |
| ANL-002 | Acquisition Analytics | Permitted user | Compare Web/Meta/campaign quality. |
| ANL-003 | Sales & Handover Analytics | Permitted user | View AI and human performance. |
| SAD-001 | Super Admin Dashboard | Platform admin | Platform health, registrations, Organizations, agents, integrations. |
| SAD-002 | Registration Applications | Platform admin | Review/approve/reject applicants. |
| SAD-003 | Organizations List | Platform admin | Search/filter all Organizations. |
| SAD-004 | Organization Detail | Platform admin | Manage Plan, usage, users, Projects, status, audit. |
| SAD-005 | Organization Integrations | Super Admin | Configure Meta Ads/WhatsApp/Instagram/Facebook for one Organization. |
| SAD-006 | Plans | Super Admin | Create/version/activate Plans and Unlimited limits. |
| SAD-007 | Platform Integrations | Super Admin | Configure Google/Search, AI credentials, shared providers. |
| SAD-008 | Agent Registry | Super Admin | Manage Agent definitions and status. |
| SAD-009 | Agent Version & Prompt | Super Admin | Configure model/credential/prompt/tools and lifecycle. |
| SAD-010 | Agent Runs & Traces | Super Admin | Inspect status, tools, latency, errors, permitted content. |
| SAD-011 | Test Lab / Evaluation | Super Admin | Run scenarios, compare versions, inspect hard failures. |
| SAD-012 | System Operations | Platform admin | Jobs, errors, provider outages, kill switches. |
| SAD-013 | Support Access | Platform admin | Open/close audited tenant support session. |
| SAD-014 | Platform Audit | Super Admin | Review high-risk platform actions. |


## 6. Global Product and UX Rules

- Every major screen shall have defined loading, empty, error, permission, limit, integration-readiness, and recovery states.
- Users shall always see the current Project/Organization context and, in conversation, the current AI/human control owner.
- Action Center is authoritative for unresolved work; notifications are supplementary.
- AI outputs shown to customers shall summarize conclusions, evidence, uncertainty, and actions without exposing private chain-of-thought.
- Long AI/provider jobs shall expose progress/status and safe retry/cancel behavior.
- Read-only and blocked states shall explain why and what action can resolve them.
- All deep links shall authenticate and re-authorize before rendering protected content.
- Core Sales Workspace and approvals shall be responsive for desktop and common mobile widths.
- English and Arabic Sales AI behavior are required; exact full RTL UI scope remains a dedicated deferred specification.

## 7. Business and Product State Models

| Model | States | Rules/notes |
| --- | --- | --- |
| Registration Application | DRAFT → SUBMITTED → UNDER_REVIEW → INFORMATION_REQUIRED \| APPROVED \| REJECTED | APPROVED creates/activates Organization; repeated approval idempotent. |
| Organization | PENDING/INACTIVE → ACTIVE → SUSPENDED → CLOSED | Suspension is non-destructive; closure controlled by Super Admin. |
| Subscription | PENDING → ACTIVE → SUSPENDED \| EXPIRED \| CANCELLED | MVP administered manually by Super Admin. |
| Company Understanding | NOT_STARTED → INPUT_COLLECTION → ANALYZING → UNDERSTANDING_REVIEW ↔ CLARIFICATION/REFINING → FINAL_REVIEW → APPROVED | ANALYSIS_FAILED is retryable; approved versions immutable. |
| Project | ONBOARDING → LIVE → ENDED | ENDED means customer closed the sales initiative; no reopen in MVP. |
| Project/Product Understanding | COLLECTING → ANALYZING → CLARIFICATION_REQUIRED ↔ REVIEW → APPROVED | Strategy blocked until approved. |
| Strategy | GENERATING → REVIEW → REVISING ↔ REVIEW → APPROVED | NEEDS_INFORMATION and FAILED exceptions; revisions limited. |
| Discovery Run | PLANNED → APPROVED → RUNNING → COMPLETED \| STOPPED \| FAILED \| PARTIAL | Stop reason mandatory. |
| Research Candidate | DISCOVERED → RESEARCHING → CONTACTABLE \| NO_CONTACT \| REJECTED | CONTACTABLE creates/matches Master Lead. |
| Project Lead | SELECTED → CONTACTING → ENGAGED → QUALIFYING → QUALIFIED → HANDOVER_READY → HUMAN_CONTROLLED → CONVERTED | Side outcomes: NURTURE, NOT_QUALIFIED, NOT_INTERESTED, SUPPRESSED; No Response is a flag. |
| Conversation Control | AI_CONTROLLED ↔ FOLLOWUP_APPROVAL_REQUIRED ↔ AI_FOLLOWUP_DELEGATED ↔ HUMAN_CONTROLLED | Human takeover allowed anytime; return-to-AI requires reassessment. |
| Follow-Up Proposal | PROPOSED → APPROVED_NOW \| APPROVED_SCHEDULED \| REJECTED \| REPLAN_REQUESTED → SENT \| CANCELLED \| OBSOLETE \| FAILED | Pre-send revalidation mandatory. |
| Handover | READY → CLAIMED/HUMAN_CONTROLLED → OUTCOME_RECORDED | Claim is atomic; unclaimed items remain Action Items. |
| Opportunity | OPEN → IN_PROGRESS → WON \| LOST \| ON_HOLD | Optional by Project/process. |
| Conversion | RECORDED → CORRECTED (audited) | Historical event; no silent overwrite. |
| Integration | NOT_CONFIGURED → CONFIGURING → CONNECTED \| PARTIAL \| ERROR \| DISABLED \| REAUTH_REQUIRED | Capabilities tracked independently. |
| Campaign | DRAFT → READY → PUBLISHING → ACTIVE ↔ PAUSED → ENDED; PUBLISH_FAILED exception | Recommended D-274; requires explicit confirmation. |
| Agent Version | DRAFT → TESTING → UNDER_REVIEW → APPROVED → ACTIVE → RETIRED | Rollback activates a previously approved version. |
| Action Item | OPEN → IN_PROGRESS → COMPLETED \| DISMISSED \| OBSOLETE | Notification dismissal does not resolve it. |

## 8. Business Rule Precedence

The application shall enforce this high-level order:

```text
Platform safety, law/policy, tenant isolation, suppression, permissions, limits
→ Organization mandatory restrictions
→ Approved Project/Product baseline
→ Approved Strategy and channel settings
→ Approved Product Catalogue sales-support content
→ Human lead-specific instruction (within allowed bounds)
→ Dynamic Lead Strategy
→ Channel-specific message wording
```

The exact knowledge-source conflict algorithm is deferred, but lower layers may never override higher protected rules.

## 9. Product Requirement Index

| PR ID | Module | Requirement/behavior | Type | BRD trace |
| --- | --- | --- | --- | --- |
| PR-01-001 | Company Registration and Platform Activation | Registration creates an application, not an active tenant. | Product Rule | BR-REG-001—012, BR-PLN-001, BR-ADM-019 |
| PR-01-002 | Company Registration and Platform Activation | CR is unique with jurisdiction. | Product Rule | BR-REG-001—012, BR-PLN-001, BR-ADM-019 |
| PR-01-003 | Company Registration and Platform Activation | Plan values are read server-side from Plan ID; browser-submitted price/limits are ignored. | Product Rule | BR-REG-001—012, BR-PLN-001, BR-ADM-019 |
| PR-01-004 | Company Registration and Platform Activation | One User may belong to multiple Organizations. | Product Rule | BR-REG-001—012, BR-PLN-001, BR-ADM-019 |
| PR-01-005 | Company Registration and Platform Activation | Account/Plan activation is manual in MVP. | Product Rule | BR-REG-001—012, BR-PLN-001, BR-ADM-019 |
| PR-01-006 | Company Registration and Platform Activation | Submitting valid data creates one SUBMITTED application and no active Organization. | Acceptance Behavior | BR-REG-001—012, BR-PLN-001, BR-ADM-019 |
| PR-01-007 | Company Registration and Platform Activation | Duplicate CR/jurisdiction is blocked server-side. | Acceptance Behavior | BR-REG-001—012, BR-PLN-001, BR-ADM-019 |
| PR-01-008 | Company Registration and Platform Activation | Approving the application creates one Organization, one Owner membership, and one subscription snapshot. | Acceptance Behavior | BR-REG-001—012, BR-PLN-001, BR-ADM-019 |
| PR-01-009 | Company Registration and Platform Activation | An existing User email can become Owner of another Organization without a duplicate User record. | Acceptance Behavior | BR-REG-001—012, BR-PLN-001, BR-ADM-019 |
| PR-01-010 | Company Registration and Platform Activation | Repeated approval requests are idempotent. | Acceptance Behavior | BR-REG-001—012, BR-PLN-001, BR-ADM-019 |
| PR-02-001 | Company Understanding Onboarding | Company Understanding covers the company/market context, not detailed Project products/prices. | Product Rule | BR-CMP-001—015, BR-POOL-003, BR-SEC-010 |
| PR-02-002 | Company Understanding Onboarding | Understanding revisions do not use Strategy Revision Allowance. | Product Rule | BR-CMP-001—015, BR-POOL-003, BR-SEC-010 |
| PR-02-003 | Company Understanding Onboarding | User-approved corrections outrank AI inference for the approved Organization baseline. | Product Rule | BR-CMP-001—015, BR-POOL-003, BR-SEC-010 |
| PR-02-004 | Company Understanding Onboarding | Private data never enters the Business Pool by default. | Product Rule | BR-CMP-001—015, BR-POOL-003, BR-SEC-010 |
| PR-02-005 | Company Understanding Onboarding | Projects may generate update suggestions but never auto-rewrite Company Understanding. | Product Rule | BR-CMP-001—015, BR-POOL-003, BR-SEC-010 |
| PR-02-006 | Company Understanding Onboarding | User can complete the analysis-review-clarification-final-review flow without repeating registration fields. | Acceptance Behavior | BR-CMP-001—015, BR-POOL-003, BR-SEC-010 |
| PR-02-007 | Company Understanding Onboarding | Approved version contains source-aware facts and business overview. | Acceptance Behavior | BR-CMP-001—015, BR-POOL-003, BR-SEC-010 |
| PR-02-008 | Company Understanding Onboarding | Private uploaded information is not visible in reusable Business Pool output. | Acceptance Behavior | BR-CMP-001—015, BR-POOL-003, BR-SEC-010 |
| PR-02-009 | Company Understanding Onboarding | Strategy Revision Allowance remains unchanged. | Acceptance Behavior | BR-CMP-001—015, BR-POOL-003, BR-SEC-010 |
| PR-02-010 | Company Understanding Onboarding | A later update preserves the previous approved version. | Acceptance Behavior | BR-CMP-001—015, BR-POOL-003, BR-SEC-010 |
| PR-03-001 | Organization Dashboard and Administration | Organization Dashboard is action-led, not a reporting-only page. | Product Rule | BR-ADM-003—007, BR-ADM-016—018, BR-UX-001, BR-PLN-012—013 |
| PR-03-002 | Organization Dashboard and Administration | Plan configuration is read-only on Organization side. | Product Rule | BR-ADM-003—007, BR-ADM-016—018, BR-UX-001, BR-PLN-012—013 |
| PR-03-003 | Organization Dashboard and Administration | Project-level product/Strategy/channel behavior stays inside the Project. | Product Rule | BR-ADM-003—007, BR-ADM-016—018, BR-UX-001, BR-PLN-012—013 |
| PR-03-004 | Organization Dashboard and Administration | Organization-wide hard communication restrictions are inherited by Projects. | Product Rule | BR-ADM-003—007, BR-ADM-016—018, BR-UX-001, BR-PLN-012—013 |
| PR-03-005 | Organization Dashboard and Administration | A new Organization sees Create First Project as primary CTA. | Acceptance Behavior | BR-ADM-003—007, BR-ADM-016—018, BR-UX-001, BR-PLN-012—013 |
| PR-03-006 | Organization Dashboard and Administration | Users see only actions/projects allowed by permissions. | Acceptance Behavior | BR-ADM-003—007, BR-ADM-016—018, BR-UX-001, BR-PLN-012—013 |
| PR-03-007 | Organization Dashboard and Administration | Plan/usage values match effective subscription snapshot and overrides. | Acceptance Behavior | BR-ADM-003—007, BR-ADM-016—018, BR-UX-001, BR-PLN-012—013 |
| PR-03-008 | Organization Dashboard and Administration | Organization settings cannot edit Project Strategy/products. | Acceptance Behavior | BR-ADM-003—007, BR-ADM-016—018, BR-UX-001, BR-PLN-012—013 |
| PR-03-009 | Organization Dashboard and Administration | Permission changes are audited and take effect server-side. | Acceptance Behavior | BR-ADM-003—007, BR-ADM-016—018, BR-UX-001, BR-PLN-012—013 |
| PR-04-001 | Project Creation, Understanding, and Final Onboarding | Each Project has one primary B2B or B2C mode. | Product Rule | BR-PRJ-001—020, BR-STR-001—017, BR-PLN-004—007 |
| PR-04-002 | Project Creation, Understanding, and Final Onboarding | Products cannot be added after final onboarding. | Product Rule | BR-PRJ-001—020, BR-STR-001—017, BR-PLN-004—007 |
| PR-04-003 | Project Creation, Understanding, and Final Onboarding | Core product facts/Project objective/Strategy are versioned. | Product Rule | BR-PRJ-001—020, BR-STR-001—017, BR-PLN-004—007 |
| PR-04-004 | Project Creation, Understanding, and Final Onboarding | Strategy approval alone does not publish campaigns/start discovery. | Product Rule | BR-PRJ-001—020, BR-STR-001—017, BR-PLN-004—007 |
| PR-04-005 | Project Creation, Understanding, and Final Onboarding | Project lifecycle is ONBOARDING/LIVE/ENDED. | Product Rule | BR-PRJ-001—020, BR-STR-001—017, BR-PLN-004—007 |
| PR-04-006 | Project Creation, Understanding, and Final Onboarding | Cannot generate Strategy before approved understanding. | Acceptance Behavior | BR-PRJ-001—020, BR-STR-001—017, BR-PLN-004—007 |
| PR-04-007 | Project Creation, Understanding, and Final Onboarding | Revision counter changes only for valid user-requested revisions. | Acceptance Behavior | BR-PRJ-001—020, BR-STR-001—017, BR-PLN-004—007 |
| PR-04-008 | Project Creation, Understanding, and Final Onboarding | Final approval creates a LIVE Project with one approved Strategy and channel settings. | Acceptance Behavior | BR-PRJ-001—020, BR-STR-001—017, BR-PLN-004—007 |
| PR-04-009 | Project Creation, Understanding, and Final Onboarding | No acquisition side effect occurs from final approval. | Acceptance Behavior | BR-PRJ-001—020, BR-STR-001—017, BR-PLN-004—007 |
| PR-04-010 | Project Creation, Understanding, and Final Onboarding | Product Catalogue matches approved product baseline. | Acceptance Behavior | BR-PRJ-001—020, BR-STR-001—017, BR-PLN-004—007 |
| PR-05-001 | Project Dashboard and Product Catalogue | Post-onboarding content never silently regenerates Strategy. | Product Rule | BR-PRJ-011—014, BR-SAL-007, BR-UX-002 |
| PR-05-002 | Project Dashboard and Product Catalogue | Sales Agent may use only correct Project/product approved content. | Product Rule | BR-PRJ-011—014, BR-SAL-007, BR-UX-002 |
| PR-05-003 | Project Dashboard and Product Catalogue | Material product changes are not normal MVP behavior. | Product Rule | BR-PRJ-011—014, BR-SAL-007, BR-UX-002 |
| PR-05-004 | Project Dashboard and Product Catalogue | Coupons/offers automation is deferred. | Product Rule | BR-PRJ-011—014, BR-SAL-007, BR-UX-002 |
| PR-05-005 | Project Dashboard and Product Catalogue | Core product fields are read-only after onboarding. | Acceptance Behavior | BR-PRJ-011—014, BR-SAL-007, BR-UX-002 |
| PR-05-006 | Project Dashboard and Product Catalogue | Adding an image does not change Strategy version. | Acceptance Behavior | BR-PRJ-011—014, BR-SAL-007, BR-UX-002 |
| PR-05-007 | Project Dashboard and Product Catalogue | Sales AI can send only assets linked to the correct product/Project. | Acceptance Behavior | BR-PRJ-011—014, BR-SAL-007, BR-UX-002 |
| PR-05-008 | Project Dashboard and Product Catalogue | New product creation is blocked. | Acceptance Behavior | BR-PRJ-011—014, BR-SAL-007, BR-UX-002 |
| PR-05-009 | Project Dashboard and Product Catalogue | Dashboard action deep-links to the incomplete product. | Acceptance Behavior | BR-PRJ-011—014, BR-SAL-007, BR-UX-002 |
| PR-06-001 | Strategy Planning and Approval | One coherent approval package; structured internal components. | Product Rule | BR-STR-001—018, BR-AGT-009 |
| PR-06-002 | Strategy Planning and Approval | Project-specific qualification weights must total/validate within platform bounds. | Product Rule | BR-STR-001—018, BR-AGT-009 |
| PR-06-003 | Strategy Planning and Approval | No invented commercial terms. | Product Rule | BR-STR-001—018, BR-AGT-009 |
| PR-06-004 | Strategy Planning and Approval | Strategy changes after Project learning require a separate human-approved version. | Product Rule | BR-STR-001—018, BR-AGT-009 |
| PR-06-005 | Strategy Planning and Approval | Output contains all mandatory sections. | Acceptance Behavior | BR-STR-001—018, BR-AGT-009 |
| PR-06-006 | Strategy Planning and Approval | Missing data does not consume a revision. | Acceptance Behavior | BR-STR-001—018, BR-AGT-009 |
| PR-06-007 | Strategy Planning and Approval | Requested revision creates one new version and consumes one allowance. | Acceptance Behavior | BR-STR-001—018, BR-AGT-009 |
| PR-06-008 | Strategy Planning and Approval | Approved Strategy references exact understanding versions. | Acceptance Behavior | BR-STR-001—018, BR-AGT-009 |
| PR-06-009 | Strategy Planning and Approval | No campaign/discovery is started automatically. | Acceptance Behavior | BR-STR-001—018, BR-AGT-009 |
| PR-07-001 | Lead Acquisition Planning | B2B/B2C influences but does not hard-code source. | Product Rule | BR-ACQ-001—003, BR-ACQ-016, BR-STR-007, BR-AGT-004 |
| PR-07-002 | Lead Acquisition Planning | Meta inbound volume does not use Web candidate multiplier. | Product Rule | BR-ACQ-001—003, BR-ACQ-016, BR-STR-007, BR-AGT-004 |
| PR-07-003 | Lead Acquisition Planning | Fresh Web discovery first under current policy. | Product Rule | BR-ACQ-001—003, BR-ACQ-016, BR-STR-007, BR-AGT-004 |
| PR-07-004 | Lead Acquisition Planning | Plan approval does not itself spend or send. | Product Rule | BR-ACQ-001—003, BR-ACQ-016, BR-STR-007, BR-AGT-004 |
| PR-07-005 | Lead Acquisition Planning | Plan clearly states Web, Meta, or Both and rationale. | Acceptance Behavior | BR-ACQ-001—003, BR-ACQ-016, BR-STR-007, BR-AGT-004 |
| PR-07-006 | Lead Acquisition Planning | User can approve without launching spend. | Acceptance Behavior | BR-ACQ-001—003, BR-ACQ-016, BR-STR-007, BR-AGT-004 |
| PR-07-007 | Lead Acquisition Planning | Missing integration creates an action and blocks only affected path. | Acceptance Behavior | BR-ACQ-001—003, BR-ACQ-016, BR-STR-007, BR-AGT-004 |
| PR-07-008 | Lead Acquisition Planning | Source-specific brief contains all inputs required by downstream Agent. | Acceptance Behavior | BR-ACQ-001—003, BR-ACQ-016, BR-STR-007, BR-AGT-004 |
| PR-07-009 | Lead Acquisition Planning | Fresh-first policy is visible for Web. | Acceptance Behavior | BR-ACQ-001—003, BR-ACQ-016, BR-STR-007, BR-AGT-004 |
| PR-08-001 | Web/Google Lead Discovery | Every captured contactable person enters Master Lead DB. | Product Rule | BR-ACQ-004—012, BR-POOL-001—014, BR-QLF-001—007 |
| PR-08-002 | Web/Google Lead Discovery | Fresh discovery occurs before Master Pool supplementation. | Product Rule | BR-ACQ-004—012, BR-POOL-001—014, BR-QLF-001—007 |
| PR-08-003 | Web/Google Lead Discovery | Search criteria target practical B2B influencers. | Product Rule | BR-ACQ-004—012, BR-POOL-001—014, BR-QLF-001—007 |
| PR-08-004 | Web/Google Lead Discovery | All source/query/run provenance is retained. | Product Rule | BR-ACQ-004—012, BR-POOL-001—014, BR-QLF-001—007 |
| PR-08-005 | Web/Google Lead Discovery | Discovery loops/time/cost are bounded. | Product Rule | BR-ACQ-004—012, BR-POOL-001—014, BR-QLF-001—007 |
| PR-08-006 | Web/Google Lead Discovery | Candidate target follows effective limit and configured multiplier. | Acceptance Behavior | BR-ACQ-004—012, BR-POOL-001—014, BR-QLF-001—007 |
| PR-08-007 | Web/Google Lead Discovery | Non-contactable results do not become active leads. | Acceptance Behavior | BR-ACQ-004—012, BR-POOL-001—014, BR-QLF-001—007 |
| PR-08-008 | Web/Google Lead Discovery | Every allocated lead has provenance and at least one contact method. | Acceptance Behavior | BR-ACQ-004—012, BR-POOL-001—014, BR-QLF-001—007 |
| PR-08-009 | Web/Google Lead Discovery | No Project receives more active Project Leads than allowance. | Acceptance Behavior | BR-ACQ-004—012, BR-POOL-001—014, BR-QLF-001—007 |
| PR-08-010 | Web/Google Lead Discovery | Run records a deterministic stop reason. | Acceptance Behavior | BR-ACQ-004—012, BR-POOL-001—014, BR-QLF-001—007 |
| PR-09-001 | Meta Campaign Creation, Publication, and Monitoring | User publication is mandatory. | Product Rule | BR-CAM-001—018, BR-ACQ-013—020, BR-SEC-022 |
| PR-09-002 | Meta Campaign Creation, Publication, and Monitoring | AI cannot independently increase spend or extend. | Product Rule | BR-CAM-001—018, BR-ACQ-013—020, BR-SEC-022 |
| PR-09-003 | Meta Campaign Creation, Publication, and Monitoring | End date is mandatory. | Product Rule | BR-CAM-001—018, BR-ACQ-013—020, BR-SEC-022 |
| PR-09-004 | Meta Campaign Creation, Publication, and Monitoring | Primary paid Meta conversation destination is WhatsApp in MVP. | Product Rule | BR-CAM-001—018, BR-ACQ-013—020, BR-SEC-022 |
| PR-09-005 | Meta Campaign Creation, Publication, and Monitoring | Ended campaigns do not restart in place. | Product Rule | BR-CAM-001—018, BR-ACQ-013—020, BR-SEC-022 |
| PR-09-006 | Meta Campaign Creation, Publication, and Monitoring | Recommended campaign lifecycle D-274 requires final confirmation. | Product Rule | BR-CAM-001—018, BR-ACQ-013—020, BR-SEC-022 |
| PR-09-007 | Meta Campaign Creation, Publication, and Monitoring | Publish is impossible without explicit authorized click. | Acceptance Behavior | BR-CAM-001—018, BR-ACQ-013—020, BR-SEC-022 |
| PR-09-008 | Meta Campaign Creation, Publication, and Monitoring | Campaign requires an end date and complete approved media. | Acceptance Behavior | BR-CAM-001—018, BR-ACQ-013—020, BR-SEC-022 |
| PR-09-009 | Meta Campaign Creation, Publication, and Monitoring | Provider failure is actionable and retry-safe. | Acceptance Behavior | BR-CAM-001—018, BR-ACQ-013—020, BR-SEC-022 |
| PR-09-010 | Meta Campaign Creation, Publication, and Monitoring | Generated leads retain campaign attribution. | Acceptance Behavior | BR-CAM-001—018, BR-ACQ-013—020, BR-SEC-022 |
| PR-09-011 | Meta Campaign Creation, Publication, and Monitoring | AI cannot execute budget increase/extension. | Acceptance Behavior | BR-CAM-001—018, BR-ACQ-013—020, BR-SEC-022 |
| PR-09-012 | Meta Campaign Creation, Publication, and Monitoring | Overflow lead behavior respects Project Lead Limit. | Acceptance Behavior | BR-CAM-001—018, BR-ACQ-013—020, BR-SEC-022 |
| PR-10-001 | Channel Integration Readiness and Meta Messaging | Super Admin configures Organization Meta/messaging APIs. | Product Rule | BR-ACQ-016—019, BR-EML-009—012, BR-ADM-008—011, BR-SEC-007—009 |
| PR-10-002 | Channel Integration Readiness and Meta Messaging | Organization users do not see secrets. | Product Rule | BR-ACQ-016—019, BR-EML-009—012, BR-ADM-008—011, BR-SEC-007—009 |
| PR-10-003 | Channel Integration Readiness and Meta Messaging | Entitlement and readiness are distinct. | Product Rule | BR-ACQ-016—019, BR-EML-009—012, BR-ADM-008—011, BR-SEC-007—009 |
| PR-10-004 | Channel Integration Readiness and Meta Messaging | Fallback requires approved, configured, healthy, contactable, eligible destination. | Product Rule | BR-ACQ-016—019, BR-EML-009—012, BR-ADM-008—011, BR-SEC-007—009 |
| PR-10-005 | Channel Integration Readiness and Meta Messaging | No blind repeated provider retries. | Product Rule | BR-ACQ-016—019, BR-EML-009—012, BR-ADM-008—011, BR-SEC-007—009 |
| PR-10-006 | Channel Integration Readiness and Meta Messaging | Organization sees capability as included-but-not-configured when appropriate. | Acceptance Behavior | BR-ACQ-016—019, BR-EML-009—012, BR-ADM-008—011, BR-SEC-007—009 |
| PR-10-007 | Channel Integration Readiness and Meta Messaging | Failed connection blocks only affected execution and creates action. | Acceptance Behavior | BR-ACQ-016—019, BR-EML-009—012, BR-ADM-008—011, BR-SEC-007—009 |
| PR-10-008 | Channel Integration Readiness and Meta Messaging | Verified inbound event maps to the correct Project Lead. | Acceptance Behavior | BR-ACQ-016—019, BR-EML-009—012, BR-ADM-008—011, BR-SEC-007—009 |
| PR-10-009 | Channel Integration Readiness and Meta Messaging | Fallback never uses an unapproved/unavailable contact. | Acceptance Behavior | BR-ACQ-016—019, BR-EML-009—012, BR-ADM-008—011, BR-SEC-007—009 |
| PR-10-010 | Channel Integration Readiness and Meta Messaging | Secrets are inaccessible to Organization users and agents. | Acceptance Behavior | BR-ACQ-016—019, BR-EML-009—012, BR-ADM-008—011, BR-SEC-007—009 |
| PR-11-001 | Generic Business Email Channel | SMTP alone is not sufficient for the product requirement; inbound retrieval/push capability is required. | Product Rule | BR-EML-001—012, BR-SAL-008—016, BR-SEC-006—008 |
| PR-11-002 | Generic Business Email Channel | AI identity is configured by the Organization and cannot be fabricated. | Product Rule | BR-EML-001—012, BR-SAL-008—016, BR-SEC-006—008 |
| PR-11-003 | Generic Business Email Channel | First eligible outreach is automatic; normal inactive follow-up is approval-controlled. | Product Rule | BR-EML-001—012, BR-SAL-008—016, BR-SEC-006—008 |
| PR-11-004 | Generic Business Email Channel | Bounces are channel evidence, not automatic lead disqualification. | Product Rule | BR-EML-001—012, BR-SAL-008—016, BR-SEC-006—008 |
| PR-11-005 | Generic Business Email Channel | A compatible mailbox can send and ingest a threaded reply. | Acceptance Behavior | BR-EML-001—012, BR-SAL-008—016, BR-SEC-006—008 |
| PR-11-006 | Generic Business Email Channel | Reply appears in the same Project Lead conversation. | Acceptance Behavior | BR-EML-001—012, BR-SAL-008—016, BR-SEC-006—008 |
| PR-11-007 | Generic Business Email Channel | AI responds automatically while active and waits for approval for normal follow-up. | Acceptance Behavior | BR-EML-001—012, BR-SAL-008—016, BR-SEC-006—008 |
| PR-11-008 | Generic Business Email Channel | Bounce blocks additional automatic sends to that address and creates action. | Acceptance Behavior | BR-EML-001—012, BR-SAL-008—016, BR-SEC-006—008 |
| PR-11-009 | Generic Business Email Channel | Configured sender identity appears consistently. | Acceptance Behavior | BR-EML-001—012, BR-SAL-008—016, BR-SEC-006—008 |
| PR-12-001 | Business Pool, Master Lead Pool, and Project Lead Context | Business Pool contains companies; Master Lead contains people. | Product Rule | BR-POOL-001—015, BR-SEC-016—018 |
| PR-12-002 | Business Pool, Master Lead Pool, and Project Lead Context | A Master Lead may serve multiple unrelated Projects. | Product Rule | BR-POOL-001—015, BR-SEC-016—018 |
| PR-12-003 | Business Pool, Master Lead Pool, and Project Lead Context | No universal score/Strategy/state on Master Lead. | Product Rule | BR-POOL-001—015, BR-SEC-016—018 |
| PR-12-004 | Business Pool, Master Lead Pool, and Project Lead Context | CR+jurisdiction strongest company identifier where available. | Product Rule | BR-POOL-001—015, BR-SEC-016—018 |
| PR-12-005 | Business Pool, Master Lead Pool, and Project Lead Context | Names alone never auto-merge people. | Product Rule | BR-POOL-001—015, BR-SEC-016—018 |
| PR-12-006 | Business Pool, Master Lead Pool, and Project Lead Context | All reusable classifications require evidence and context. | Product Rule | BR-POOL-001—015, BR-SEC-016—018 |
| PR-12-007 | Business Pool, Master Lead Pool, and Project Lead Context | Same person used by two Projects has one Master Lead and two Project Leads. | Acceptance Behavior | BR-POOL-001—015, BR-SEC-016—018 |
| PR-12-008 | Business Pool, Master Lead Pool, and Project Lead Context | Project A private conversation is unavailable to Project B/Organization B. | Acceptance Behavior | BR-POOL-001—015, BR-SEC-016—018 |
| PR-12-009 | Business Pool, Master Lead Pool, and Project Lead Context | Merge preserves both sources and can be reversed. | Acceptance Behavior | BR-POOL-001—015, BR-SEC-016—018 |
| PR-12-010 | Business Pool, Master Lead Pool, and Project Lead Context | Tenant Organization linkage does not reveal private account activity. | Acceptance Behavior | BR-POOL-001—015, BR-SEC-016—018 |
| PR-12-011 | Business Pool, Master Lead Pool, and Project Lead Context | Contextual Not Interested signal is not generalized globally. | Acceptance Behavior | BR-POOL-001—015, BR-SEC-016—018 |
| PR-13-001 | Lead Enrichment and Identity Resolution | Purpose-limited, not exhaustive background research. | Product Rule | BR-QLF-001—005, BR-POOL-007—010, BR-SEC-016 |
| PR-13-002 | Lead Enrichment and Identity Resolution | Observed/inferred distinction is mandatory. | Product Rule | BR-QLF-001—005, BR-POOL-007—010, BR-SEC-016 |
| PR-13-003 | Lead Enrichment and Identity Resolution | Private conversation-derived data is not reusable Master Lead enrichment. | Product Rule | BR-QLF-001—005, BR-POOL-007—010, BR-SEC-016 |
| PR-13-004 | Lead Enrichment and Identity Resolution | AI recommends match; deterministic service merges. | Product Rule | BR-QLF-001—005, BR-POOL-007—010, BR-SEC-016 |
| PR-13-005 | Lead Enrichment and Identity Resolution | Every persisted fact has source/confidence/freshness/type. | Acceptance Behavior | BR-QLF-001—005, BR-POOL-007—010, BR-SEC-016 |
| PR-13-006 | Lead Enrichment and Identity Resolution | Unknown data is not fabricated. | Acceptance Behavior | BR-QLF-001—005, BR-POOL-007—010, BR-SEC-016 |
| PR-13-007 | Lead Enrichment and Identity Resolution | Reusable and Project-specific facts are separated. | Acceptance Behavior | BR-QLF-001—005, BR-POOL-007—010, BR-SEC-016 |
| PR-13-008 | Lead Enrichment and Identity Resolution | Targeted re-enrichment returns to Lead Strategist. | Acceptance Behavior | BR-QLF-001—005, BR-POOL-007—010, BR-SEC-016 |
| PR-13-009 | Lead Enrichment and Identity Resolution | Agent cannot merge identities directly. | Acceptance Behavior | BR-QLF-001—005, BR-POOL-007—010, BR-SEC-016 |
| PR-14-001 | Initial Scoring, Dynamic Qualification, and Lead State | Initial score is Project-specific, not universal. | Product Rule | BR-QLF-006—017, BR-SAL-002—003 |
| PR-14-002 | Initial Scoring, Dynamic Qualification, and Lead State | Qualified = fit + genuine need/interest + realistic path to progress. | Product Rule | BR-QLF-006—017, BR-SAL-002—003 |
| PR-14-003 | Initial Scoring, Dynamic Qualification, and Lead State | Outcomes are Project-specific except suppression scope. | Product Rule | BR-QLF-006—017, BR-SAL-002—003 |
| PR-14-004 | Initial Scoring, Dynamic Qualification, and Lead State | State describes operation; score describes strength. | Product Rule | BR-QLF-006—017, BR-SAL-002—003 |
| PR-14-005 | Initial Scoring, Dynamic Qualification, and Lead State | Every transition has explanation/evidence. | Product Rule | BR-QLF-006—017, BR-SAL-002—003 |
| PR-14-006 | Initial Scoring, Dynamic Qualification, and Lead State | Web candidate gets an initial score before first contact. | Acceptance Behavior | BR-QLF-006—017, BR-SAL-002—003 |
| PR-14-007 | Initial Scoring, Dynamic Qualification, and Lead State | Meta lead begins with intent evidence. | Acceptance Behavior | BR-QLF-006—017, BR-SAL-002—003 |
| PR-14-008 | Initial Scoring, Dynamic Qualification, and Lead State | Dynamic evidence can change score/state both directions. | Acceptance Behavior | BR-QLF-006—017, BR-SAL-002—003 |
| PR-14-009 | Initial Scoring, Dynamic Qualification, and Lead State | Suppressed lead cannot receive outbound communication. | Acceptance Behavior | BR-QLF-006—017, BR-SAL-002—003 |
| PR-14-010 | Initial Scoring, Dynamic Qualification, and Lead State | Every active lead has one current Next Best Action. | Acceptance Behavior | BR-QLF-006—017, BR-SAL-002—003 |
| PR-15-001 | Dynamic Lead Strategy and Next Best Action | One Lead Strategy per Project Lead. | Product Rule | BR-SAL-001—004, BR-SAL-017, BR-AGT-017 |
| PR-15-002 | Dynamic Lead Strategy and Next Best Action | One current Next Best Action. | Product Rule | BR-SAL-001—004, BR-SAL-017, BR-AGT-017 |
| PR-15-003 | Dynamic Lead Strategy and Next Best Action | Strategy is not a giant free-form document; structured living object. | Product Rule | BR-SAL-001—004, BR-SAL-017, BR-AGT-017 |
| PR-15-004 | Dynamic Lead Strategy and Next Best Action | Human gives instructions through UI; does not edit protected fields. | Product Rule | BR-SAL-001—004, BR-SAL-017, BR-AGT-017 |
| PR-15-005 | Dynamic Lead Strategy and Next Best Action | Lead Strategist never sends external messages. | Product Rule | BR-SAL-001—004, BR-SAL-017, BR-AGT-017 |
| PR-15-006 | Dynamic Lead Strategy and Next Best Action | Material event produces a new structured assessment with reason. | Acceptance Behavior | BR-SAL-001—004, BR-SAL-017, BR-AGT-017 |
| PR-15-007 | Dynamic Lead Strategy and Next Best Action | User sees concise sales insight. | Acceptance Behavior | BR-SAL-001—004, BR-SAL-017, BR-AGT-017 |
| PR-15-008 | Dynamic Lead Strategy and Next Best Action | Agent remembers approaches tried and does not repeat failed angle without reason. | Acceptance Behavior | BR-SAL-001—004, BR-SAL-017, BR-AGT-017 |
| PR-15-009 | Dynamic Lead Strategy and Next Best Action | Follow-up recommendation contains timing/channel/objective/angle/approval. | Acceptance Behavior | BR-SAL-001—004, BR-SAL-017, BR-AGT-017 |
| PR-15-010 | Dynamic Lead Strategy and Next Best Action | Unsafe/missing-knowledge case returns human review instead of fabrication. | Acceptance Behavior | BR-SAL-001—004, BR-SAL-017, BR-AGT-017 |
| PR-16-001 | Sales Agent and Unified Conversation Workspace | Natural and adaptive, not templated blast behavior. | Product Rule | BR-SAL-004—025, BR-EML-004—009, BR-UX-012 |
| PR-16-002 | Sales Agent and Unified Conversation Workspace | Active replies autonomous. | Product Rule | BR-SAL-004—025, BR-EML-004—009, BR-UX-012 |
| PR-16-003 | Sales Agent and Unified Conversation Workspace | Strategy and wording responsibilities remain separate. | Product Rule | BR-SAL-004—025, BR-EML-004—009, BR-UX-012 |
| PR-16-004 | Sales Agent and Unified Conversation Workspace | All messages and media are Project-scoped and attributed. | Product Rule | BR-SAL-004—025, BR-EML-004—009, BR-UX-012 |
| PR-16-005 | Sales Agent and Unified Conversation Workspace | Human can take over anytime. | Product Rule | BR-SAL-004—025, BR-EML-004—009, BR-UX-012 |
| PR-16-006 | Sales Agent and Unified Conversation Workspace | Lead queue, conversation, and AI insight appear in one workspace. | Acceptance Behavior | BR-SAL-004—025, BR-EML-004—009, BR-UX-012 |
| PR-16-007 | Sales Agent and Unified Conversation Workspace | AI responds autonomously to active inbound when permitted. | Acceptance Behavior | BR-SAL-004—025, BR-EML-004—009, BR-UX-012 |
| PR-16-008 | Sales Agent and Unified Conversation Workspace | Human takeover immediately prevents AI send. | Acceptance Behavior | BR-SAL-004—025, BR-EML-004—009, BR-UX-012 |
| PR-16-009 | Sales Agent and Unified Conversation Workspace | Return to AI uses full latest context. | Acceptance Behavior | BR-SAL-004—025, BR-EML-004—009, BR-UX-012 |
| PR-16-010 | Sales Agent and Unified Conversation Workspace | Unsupported product question is not fabricated. | Acceptance Behavior | BR-SAL-004—025, BR-EML-004—009, BR-UX-012 |
| PR-17-001 | Follow-Up, Unresponsive Attention, and Conversation Control | Default follow-up requires human approval. | Product Rule | BR-SAL-011—017, BR-UX-003—008 |
| PR-17-002 | Follow-Up, Unresponsive Attention, and Conversation Control | Delegation is explicit, per Project Lead, revocable, and auditable. | Product Rule | BR-SAL-011—017, BR-UX-003—008 |
| PR-17-003 | Follow-Up, Unresponsive Attention, and Conversation Control | One adaptive follow-up at a time. | Product Rule | BR-SAL-011—017, BR-UX-003—008 |
| PR-17-004 | Follow-Up, Unresponsive Attention, and Conversation Control | No universal fixed timing schedule. | Product Rule | BR-SAL-011—017, BR-UX-003—008 |
| PR-17-005 | Follow-Up, Unresponsive Attention, and Conversation Control | No Response is not final status. | Product Rule | BR-SAL-011—017, BR-UX-003—008 |
| PR-17-006 | Follow-Up, Unresponsive Attention, and Conversation Control | No normal inactive follow-up sends without approval or active delegation. | Acceptance Behavior | BR-SAL-011—017, BR-UX-003—008 |
| PR-17-007 | Follow-Up, Unresponsive Attention, and Conversation Control | Lead reply cancels obsolete pending send. | Acceptance Behavior | BR-SAL-011—017, BR-UX-003—008 |
| PR-17-008 | Follow-Up, Unresponsive Attention, and Conversation Control | Delegation can be revoked and cannot bypass suppression/final action boundary. | Acceptance Behavior | BR-SAL-011—017, BR-UX-003—008 |
| PR-17-009 | Follow-Up, Unresponsive Attention, and Conversation Control | Unresponsive flag appears as action, not forced status. | Acceptance Behavior | BR-SAL-011—017, BR-UX-003—008 |
| PR-17-010 | Follow-Up, Unresponsive Attention, and Conversation Control | Scheduled send rechecks all mandatory conditions. | Acceptance Behavior | BR-SAL-011—017, BR-UX-003—008 |
| PR-18-001 | Human Handover, Opportunity, and Conversion | All final MVP actions are human. | Product Rule | BR-CON-001—012, BR-ANL-005, BR-UX-006 |
| PR-18-002 | Human Handover, Opportunity, and Conversion | High interest alone does not require handover until final/protected boundary. | Product Rule | BR-CON-001—012, BR-ANL-005, BR-UX-006 |
| PR-18-003 | Human Handover, Opportunity, and Conversion | Opportunity is optional. | Product Rule | BR-CON-001—012, BR-ANL-005, BR-UX-006 |
| PR-18-004 | Human Handover, Opportunity, and Conversion | Conversion is an event, not a boolean. | Product Rule | BR-CON-001—012, BR-ANL-005, BR-UX-006 |
| PR-18-005 | Human Handover, Opportunity, and Conversion | Master Lead is not globally Won/Lost. | Product Rule | BR-CON-001—012, BR-ANL-005, BR-UX-006 |
| PR-18-006 | Human Handover, Opportunity, and Conversion | All eligible users see handover; exactly one claim succeeds. | Acceptance Behavior | BR-CON-001—012, BR-ANL-005, BR-UX-006 |
| PR-18-007 | Human Handover, Opportunity, and Conversion | AI sends no sales messages after claim. | Acceptance Behavior | BR-CON-001—012, BR-ANL-005, BR-UX-006 |
| PR-18-008 | Human Handover, Opportunity, and Conversion | Human can record final outcome without large CRM form. | Acceptance Behavior | BR-CON-001—012, BR-ANL-005, BR-UX-006 |
| PR-18-009 | Human Handover, Opportunity, and Conversion | Conversion includes source/campaign and human evidence. | Acceptance Behavior | BR-CON-001—012, BR-ANL-005, BR-UX-006 |
| PR-18-010 | Human Handover, Opportunity, and Conversion | Analytics distinguishes AI preparation from human result. | Acceptance Behavior | BR-CON-001—012, BR-ANL-005, BR-UX-006 |
| PR-19-001 | Project, Campaign, Sales Analytics, and Learning | Message volume is not a success metric by itself. | Product Rule | BR-ANL-001—014, BR-CON-011—012 |
| PR-19-002 | Project, Campaign, Sales Analytics, and Learning | Lead-level adaptation is automatic; Project Strategy changes are not. | Product Rule | BR-ANL-001—014, BR-CON-011—012 |
| PR-19-003 | Project, Campaign, Sales Analytics, and Learning | One outcome never rewrites Strategy. | Product Rule | BR-ANL-001—014, BR-CON-011—012 |
| PR-19-004 | Project, Campaign, Sales Analytics, and Learning | Master Lead is not the primary reporting dimension. | Product Rule | BR-ANL-001—014, BR-CON-011—012 |
| PR-19-005 | Project, Campaign, Sales Analytics, and Learning | Dashboard traces conversion to source and Project Lead. | Acceptance Behavior | BR-ANL-001—014, BR-CON-011—012 |
| PR-19-006 | Project, Campaign, Sales Analytics, and Learning | Cheap but low-quality leads are distinguishable from high-quality sources. | Acceptance Behavior | BR-ANL-001—014, BR-CON-011—012 |
| PR-19-007 | Project, Campaign, Sales Analytics, and Learning | Human delay is distinguishable from AI performance. | Acceptance Behavior | BR-ANL-001—014, BR-CON-011—012 |
| PR-19-008 | Project, Campaign, Sales Analytics, and Learning | Recommendation states evidence and confidence. | Acceptance Behavior | BR-ANL-001—014, BR-CON-011—012 |
| PR-19-009 | Project, Campaign, Sales Analytics, and Learning | Rejecting a recommendation leaves current Strategy unchanged. | Acceptance Behavior | BR-ANL-001—014, BR-CON-011—012 |
| PR-20-001 | Plans, Usage, Limits, and Entitlements | Plan, entitlement, limit, and usage are separate concepts. | Product Rule | BR-PLN-001—014, BR-ADM-012, BR-SEC-004 |
| PR-20-002 | Plans, Usage, Limits, and Entitlements | Readiness is separate from entitlement. | Product Rule | BR-PLN-001—014, BR-ADM-012, BR-SEC-004 |
| PR-20-003 | Plans, Usage, Limits, and Entitlements | All enforcement is server-side. | Product Rule | BR-PLN-001—014, BR-ADM-012, BR-SEC-004 |
| PR-20-004 | Plans, Usage, Limits, and Entitlements | Manual billing/subscription status in MVP. | Product Rule | BR-PLN-001—014, BR-ADM-012, BR-SEC-004 |
| PR-20-005 | Plans, Usage, Limits, and Entitlements | Numeric and Unlimited limits enforce correctly. | Acceptance Behavior | BR-PLN-001—014, BR-ADM-012, BR-SEC-004 |
| PR-20-006 | Plans, Usage, Limits, and Entitlements | Organization cannot alter Plan through browser request. | Acceptance Behavior | BR-PLN-001—014, BR-ADM-012, BR-SEC-004 |
| PR-20-007 | Plans, Usage, Limits, and Entitlements | Existing subscription does not silently change after Plan edit. | Acceptance Behavior | BR-PLN-001—014, BR-ADM-012, BR-SEC-004 |
| PR-20-008 | Plans, Usage, Limits, and Entitlements | Overflow lead behavior matches D-272. | Acceptance Behavior | BR-PLN-001—014, BR-ADM-012, BR-SEC-004 |
| PR-20-009 | Plans, Usage, Limits, and Entitlements | Warnings explain exact blocked/new versus continuing behavior. | Acceptance Behavior | BR-PLN-001—014, BR-ADM-012, BR-SEC-004 |
| PR-21-001 | Direct Permissions, Project Access, Action Center, Notifications, and Audit | Direct permissions + Project access are the authorization model. | Product Rule | BR-ADM-003—007, BR-UX-003—010, BR-REG-009 |
| PR-21-002 | Direct Permissions, Project Access, Action Center, Notifications, and Audit | Organization Owner is protected. | Product Rule | BR-ADM-003—007, BR-UX-003—010, BR-REG-009 |
| PR-21-003 | Direct Permissions, Project Access, Action Center, Notifications, and Audit | Action Center, not notification, is unresolved-work source of truth. | Product Rule | BR-ADM-003—007, BR-UX-003—010, BR-REG-009 |
| PR-21-004 | Direct Permissions, Project Access, Action Center, Notifications, and Audit | Agents do not create authoritative audit records. | Product Rule | BR-ADM-003—007, BR-UX-003—010, BR-REG-009 |
| PR-21-005 | Direct Permissions, Project Access, Action Center, Notifications, and Audit | User sees only authorized Projects and actions. | Acceptance Behavior | BR-ADM-003—007, BR-UX-003—010, BR-REG-009 |
| PR-21-006 | Direct Permissions, Project Access, Action Center, Notifications, and Audit | Direct unauthorized API attempt is denied even if UI control hidden. | Acceptance Behavior | BR-ADM-003—007, BR-UX-003—010, BR-REG-009 |
| PR-21-007 | Direct Permissions, Project Access, Action Center, Notifications, and Audit | Action completion and notification dismissal behave independently. | Acceptance Behavior | BR-ADM-003—007, BR-UX-003—010, BR-REG-009 |
| PR-21-008 | Direct Permissions, Project Access, Action Center, Notifications, and Audit | Removed user's historical audit remains. | Acceptance Behavior | BR-ADM-003—007, BR-UX-003—010, BR-REG-009 |
| PR-21-009 | Direct Permissions, Project Access, Action Center, Notifications, and Audit | Email notification does not expose full conversation. | Acceptance Behavior | BR-ADM-003—007, BR-UX-003—010, BR-REG-009 |
| PR-22-001 | Super Admin Control Panel and Platform Operations | Super Admin CP is distinct from Organization Admin. | Product Rule | BR-ADM-001—020, BR-AGT-009—015, BR-SEC-021, BR-ANL-013 |
| PR-22-002 | Super Admin Control Panel and Platform Operations | Platform Admin/Support does not automatically receive Super Admin powers. | Product Rule | BR-ADM-001—020, BR-AGT-009—015, BR-SEC-021, BR-ANL-013 |
| PR-22-003 | Super Admin Control Panel and Platform Operations | Platform-managed Organization integration resides inside Organization detail. | Product Rule | BR-ADM-001—020, BR-AGT-009—015, BR-SEC-021, BR-ANL-013 |
| PR-22-004 | Super Admin Control Panel and Platform Operations | Platform-wide APIs reside under Platform Integrations. | Product Rule | BR-ADM-001—020, BR-AGT-009—015, BR-SEC-021, BR-ANL-013 |
| PR-22-005 | Super Admin Control Panel and Platform Operations | End-to-end observability is subject to tenant/privacy controls. | Product Rule | BR-ADM-001—020, BR-AGT-009—015, BR-SEC-021, BR-ANL-013 |
| PR-22-006 | Super Admin Control Panel and Platform Operations | Super Admin can configure Organization X Meta capabilities without exposing secrets to Organization X. | Acceptance Behavior | BR-ADM-001—020, BR-AGT-009—015, BR-SEC-021, BR-ANL-013 |
| PR-22-007 | Super Admin Control Panel and Platform Operations | Super Admin can choose model/credential/prompt version per agent. | Acceptance Behavior | BR-ADM-001—020, BR-AGT-009—015, BR-SEC-021, BR-ANL-013 |
| PR-22-008 | Super Admin Control Panel and Platform Operations | Agent run view shows exact version/tools/status and hides sensitive content by default. | Acceptance Behavior | BR-ADM-001—020, BR-AGT-009—015, BR-SEC-021, BR-ANL-013 |
| PR-22-009 | Super Admin Control Panel and Platform Operations | External-action kill switch blocks provider sends. | Acceptance Behavior | BR-ADM-001—020, BR-AGT-009—015, BR-SEC-021, BR-ANL-013 |
| PR-22-010 | Super Admin Control Panel and Platform Operations | Support access is recorded and expires/ends. | Acceptance Behavior | BR-ADM-001—020, BR-AGT-009—015, BR-SEC-021, BR-ANL-013 |
| PR-23-001 | Google ADK Agent Component, Agent Registry, and Evaluation | ADK is Agent Component, not source of business truth. | Product Rule | BR-AGT-001—020, BR-TST-001—008 |
| PR-23-002 | Google ADK Agent Component, Agent Registry, and Evaluation | Agent interaction graph and tools are deny-by-default. | Product Rule | BR-AGT-001—020, BR-TST-001—008 |
| PR-23-003 | Google ADK Agent Component, Agent Registry, and Evaluation | Editable prompt cannot disable platform safety. | Product Rule | BR-AGT-001—020, BR-TST-001—008 |
| PR-23-004 | Google ADK Agent Component, Agent Registry, and Evaluation | History is version-pinned. | Product Rule | BR-AGT-001—020, BR-TST-001—008 |
| PR-23-005 | Google ADK Agent Component, Agent Registry, and Evaluation | Customer Simulator never causes live side effects. | Product Rule | BR-AGT-001—020, BR-TST-001—008 |
| PR-23-006 | Google ADK Agent Component, Agent Registry, and Evaluation | Agent cannot retrieve unrelated tenant context. | Acceptance Behavior | BR-AGT-001—020, BR-TST-001—008 |
| PR-23-007 | Google ADK Agent Component, Agent Registry, and Evaluation | Invalid structured output is rejected/retried/blocked, not persisted. | Acceptance Behavior | BR-AGT-001—020, BR-TST-001—008 |
| PR-23-008 | Google ADK Agent Component, Agent Registry, and Evaluation | Changing active version affects future runs only. | Acceptance Behavior | BR-AGT-001—020, BR-TST-001—008 |
| PR-23-009 | Google ADK Agent Component, Agent Registry, and Evaluation | Simulator sends no real provider action. | Acceptance Behavior | BR-AGT-001—020, BR-TST-001—008 |
| PR-23-010 | Google ADK Agent Component, Agent Registry, and Evaluation | Critical violation prevents activation. | Acceptance Behavior | BR-AGT-001—020, BR-TST-001—008 |
| PR-24-001 | Security, Privacy, Compliance, and Trust Controls | AI is untrusted. | Product Rule | BR-SEC-001—025, BR-REG-009, BR-EML-012 |
| PR-24-002 | Security, Privacy, Compliance, and Trust Controls | UI hiding is never authorization. | Product Rule | BR-SEC-001—025, BR-REG-009, BR-EML-012 |
| PR-24-003 | Security, Privacy, Compliance, and Trust Controls | Suppression and protected states are deterministic. | Product Rule | BR-SEC-001—025, BR-REG-009, BR-EML-012 |
| PR-24-004 | Security, Privacy, Compliance, and Trust Controls | Private conversation is not reusable Master Lead data. | Product Rule | BR-SEC-001—025, BR-REG-009, BR-EML-012 |
| PR-24-005 | Security, Privacy, Compliance, and Trust Controls | Production data is not casually used in test/local environments. | Product Rule | BR-SEC-001—025, BR-REG-009, BR-EML-012 |
| PR-24-006 | Security, Privacy, Compliance, and Trust Controls | Cross-tenant resource request is denied server-side. | Acceptance Behavior | BR-SEC-001—025, BR-REG-009, BR-EML-012 |
| PR-24-007 | Security, Privacy, Compliance, and Trust Controls | Suppressed lead cannot receive a message through any Agent path. | Acceptance Behavior | BR-SEC-001—025, BR-REG-009, BR-EML-012 |
| PR-24-008 | Security, Privacy, Compliance, and Trust Controls | Agent cannot invoke unregistered tool or bypass provider service. | Acceptance Behavior | BR-SEC-001—025, BR-REG-009, BR-EML-012 |
| PR-24-009 | Security, Privacy, Compliance, and Trust Controls | Invalid webhook is not processed. | Acceptance Behavior | BR-SEC-001—025, BR-REG-009, BR-EML-012 |
| PR-24-010 | Security, Privacy, Compliance, and Trust Controls | Unverified legal/provider requirement remains visibly blocked. | Acceptance Behavior | BR-SEC-001—025, BR-REG-009, BR-EML-012 |
| PR-25-001 | Global UX, Accessibility, Language, and Prototype | Product feels like an intelligent sales workspace, not an Agent debugger. | Product Rule | BR-UX-001—012, BR-REG-003, BR-SAL-025 |
| PR-25-002 | Global UX, Accessibility, Language, and Prototype | User always knows what happens next and who controls the conversation. | Product Rule | BR-UX-001—012, BR-REG-003, BR-SAL-025 |
| PR-25-003 | Global UX, Accessibility, Language, and Prototype | Technical traces stay in Super Admin operations, not normal customer UI. | Product Rule | BR-UX-001—012, BR-REG-003, BR-SAL-025 |
| PR-25-004 | Global UX, Accessibility, Language, and Prototype | First prototype scope ends at lead conversation but includes navigation to key acquisition paths. | Product Rule | BR-UX-001—012, BR-REG-003, BR-SAL-025 |
| PR-25-005 | Global UX, Accessibility, Language, and Prototype | Prototype demonstrates registration, onboarding, Project, Strategy, product content, acquisition, leads, and conversation. | Acceptance Behavior | BR-UX-001—012, BR-REG-003, BR-SAL-025 |
| PR-25-006 | Global UX, Accessibility, Language, and Prototype | Every major screen has loading/empty/error/permission/limit state. | Acceptance Behavior | BR-UX-001—012, BR-REG-003, BR-SAL-025 |
| PR-25-007 | Global UX, Accessibility, Language, and Prototype | Control owner is visible in conversation. | Acceptance Behavior | BR-UX-001—012, BR-REG-003, BR-SAL-025 |
| PR-25-008 | Global UX, Accessibility, Language, and Prototype | Action Item opens exact work location. | Acceptance Behavior | BR-UX-001—012, BR-REG-003, BR-SAL-025 |
| PR-25-009 | Global UX, Accessibility, Language, and Prototype | Core flows are usable on common desktop and mobile widths. | Acceptance Behavior | BR-UX-001—012, BR-REG-003, BR-SAL-025 |

## 10. Detailed Product Modules

### M01 — Company Registration and Platform Activation

**Purpose:** Capture a legitimate company application, prevent duplicate tenant creation, and activate an Organization and subscription only after Platform review.

**Mapped BRD requirements:** BR-REG-001—012, BR-PLN-001, BR-ADM-019

#### Actors

- Public applicant
- Platform Super Admin
- Platform Admin/Support with permission
- System

#### Representative User Stories


- **US-M01-001:** As a Public applicant, I want this capability to capture a legitimate company application, prevent duplicate tenant creation, and activate an Organization and subscription only after Platform review.

- **US-M01-002:** As the platform, I need all protected actions in this module to be permissioned, auditable, tenant-scoped, and state-valid.

- **US-M01-003:** As an authorized reviewer, I want clear loading, empty, failure, and recovery behavior so that work never depends on hidden system state.

#### Preconditions

- At least one active Plan is available for registration.
- Registration endpoint is available.
- Applicant is not already authenticated as the same existing company.

#### Inputs

- Company name
- CR/business registration number
- Country/jurisdiction
- Company contact email
- Company phone
- Website/social/other approved digital presence
- First admin full name/email/phone/password
- Selected Plan ID

#### Outputs

- Registration Application
- Applicant acknowledgement
- Super Admin Action Item
- On approval: Organization, Owner membership, subscription snapshot, effective entitlements/limits, onboarding status

#### Main Flow

1. Applicant opens registration and selects a Plan.
2. Client and server validate required fields and normalize email, phone, CR, jurisdiction, and domain/digital identity.
3. System checks CR/jurisdiction, company identity, and domain/digital identity for existing or pending records.
4. System creates a SUBMITTED Registration Application; it does not activate the Organization.
5. Super Admin reviews the application, contacts the applicant outside the platform if needed, and records notes.
6. Super Admin may request information, reject, or approve.
7. Approval creates/activates the Organization, first Owner membership, subscription snapshot, effective limits, and Company Understanding NOT_STARTED.
8. Applicant receives in-app/email account activation information and can log in.

#### Alternative, Exception, and Failure Flows

- Duplicate CR/company: block submission and direct applicant to sign in or support.
- Duplicate domain/digital identity: block by default; support resolves legitimate group/subsidiary exceptions.
- Existing user email: reuse the User identity and create a new membership after approval.
- Information required: application becomes INFORMATION_REQUIRED and preserves history.
- Rejected: no Organization access; applicant-facing and internal reasons are stored separately.
- Approval failure: transaction rolls back or resumes idempotently without duplicate Organization/subscription creation.

#### Product Rules

- Registration creates an application, not an active tenant.
- CR is unique with jurisdiction.
- Plan values are read server-side from Plan ID; browser-submitted price/limits are ignored.
- One User may belong to multiple Organizations.
- Account/Plan activation is manual in MVP.

#### Responsibility Split


**AI responsibilities**

- No AI reasoning is required in the registration approval decision.
- Company Understanding Agent does not run until activation and authorized login.


**Deterministic application responsibilities**

- Validate and normalize data.
- Perform duplicate checks.
- Create application states and audit.
- Perform transactional activation.
- Enforce password/security/rate controls.
- Create Action Items and notifications.


**Human responsibilities**

- Applicant provides truthful data.
- Super Admin reviews and decides.
- Organization Owner later completes Company Understanding.

#### Permissions

- Public application submission; registration review/approve/reject permissions; Organization activation permission.

#### Data Requirements

- RegistrationApplication
- User
- Organization
- Membership
- SubscriptionSnapshot
- PlanReference
- AuditEvent
- ActionItem
- Notification

#### UX States

- Loading Plan options
- Validation errors
- Duplicate blocked state
- Submitted/pending page
- Information required
- Rejected
- Approved/activated
- Server failure/retry

#### Notifications and Action Center

- In-app/email to Platform reviewers for new application.
- Applicant email for submitted, information required, approved, or rejected.
- No private internal review notes in applicant email.

#### Audit Requirements

- Submission
- Duplicate detection
- Review state change
- Information requested
- Approval/rejection
- Organization/subscription creation
- Actor and timestamp

#### Analytics Events and Metrics

- Applications by state
- Approval time
- Duplicate rate
- Activation failure rate

#### Security and Privacy

- Rate limiting and bot controls
- Secure password storage
- No privileged client fields
- PII-minimized logs
- IDOR protection for application status

#### B2B / B2C Behavior

- Same flow for B2B and B2C Organizations; Organization mode is determined later during Company Understanding.

#### Formal Product Requirements

| PR ID | Requirement/behavior | Type | BRD trace |
| --- | --- | --- | --- |
| PR-01-001 | Registration creates an application, not an active tenant. | Product Rule | BR-REG-001—012, BR-PLN-001, BR-ADM-019 |
| PR-01-002 | CR is unique with jurisdiction. | Product Rule | BR-REG-001—012, BR-PLN-001, BR-ADM-019 |
| PR-01-003 | Plan values are read server-side from Plan ID; browser-submitted price/limits are ignored. | Product Rule | BR-REG-001—012, BR-PLN-001, BR-ADM-019 |
| PR-01-004 | One User may belong to multiple Organizations. | Product Rule | BR-REG-001—012, BR-PLN-001, BR-ADM-019 |
| PR-01-005 | Account/Plan activation is manual in MVP. | Product Rule | BR-REG-001—012, BR-PLN-001, BR-ADM-019 |
| PR-01-006 | Submitting valid data creates one SUBMITTED application and no active Organization. | Acceptance Behavior | BR-REG-001—012, BR-PLN-001, BR-ADM-019 |
| PR-01-007 | Duplicate CR/jurisdiction is blocked server-side. | Acceptance Behavior | BR-REG-001—012, BR-PLN-001, BR-ADM-019 |
| PR-01-008 | Approving the application creates one Organization, one Owner membership, and one subscription snapshot. | Acceptance Behavior | BR-REG-001—012, BR-PLN-001, BR-ADM-019 |
| PR-01-009 | An existing User email can become Owner of another Organization without a duplicate User record. | Acceptance Behavior | BR-REG-001—012, BR-PLN-001, BR-ADM-019 |
| PR-01-010 | Repeated approval requests are idempotent. | Acceptance Behavior | BR-REG-001—012, BR-PLN-001, BR-ADM-019 |

#### Acceptance Criteria

- [ ] **AC-M01-001:** Submitting valid data creates one SUBMITTED application and no active Organization.
- [ ] **AC-M01-002:** Duplicate CR/jurisdiction is blocked server-side.
- [ ] **AC-M01-003:** Approving the application creates one Organization, one Owner membership, and one subscription snapshot.
- [ ] **AC-M01-004:** An existing User email can become Owner of another Organization without a duplicate User record.
- [ ] **AC-M01-005:** Repeated approval requests are idempotent.

### M02 — Company Understanding Onboarding

**Purpose:** Create an approved, versioned understanding of the registered business while separating private Organization Intelligence from reusable Business Pool data.

**Mapped BRD requirements:** BR-CMP-001—015, BR-POOL-003, BR-SEC-010

#### Actors

- Organization Owner
- Organization user with Company Understanding permission
- Company Understanding Agent
- System
- Platform reviewer for later material updates

#### Representative User Stories


- **US-M02-001:** As a Organization Owner, I want this capability to create an approved, versioned understanding of the registered business while separating private Organization Intelligence from reusable Business Pool data.

- **US-M02-002:** As the platform, I need all protected actions in this module to be permissioned, auditable, tenant-scoped, and state-valid.

- **US-M02-003:** As an authorized reviewer, I want clear loading, empty, failure, and recovery behavior so that work never depends on hidden system state.

#### Preconditions

- Organization and subscription are active.
- User has Company Understanding permission.
- At least one digital presence exists from registration.

#### Inputs

- Registration data
- Free-text business explanation
- Company-profile documents
- Approved digital presence
- User corrections/answers
- For updates: current approved version and new evidence

#### Outputs

- Initial Understanding
- Clarification questions
- Final Business Overview
- Approved Company Understanding version
- Private Organization Intelligence
- Sanitized Business Pool proposal

#### Main Flow

1. User lands in Company Understanding after first approved login.
2. User provides free text and/or uploads company profiles.
3. Agent analyzes registration context, digital presence, and provided sources.
4. System presents Initial Understanding with Confirmed/Inferred/Missing/Conflicting labels and evidence.
5. User reviews before clarification.
6. Agent asks only material targeted questions; user answers.
7. Agent refines until critical gaps are resolved or explicitly acknowledged.
8. System presents a clean Final Business Overview and evidence summary.
9. Authorized user approves the version.
10. System persists private Organization Intelligence and separately validates reusable Business Pool attributes.
11. User is routed to the Organization Dashboard.

#### Alternative, Exception, and Failure Flows

- No document/free text: prompt the user to provide at least one substantive source.
- Source conflict: show conflict and require correction/confirmation where material.
- Analysis failure: allow retry and preserve submitted sources.
- Non-critical unknown: permit approval if it does not prevent reliable business classification.
- Later material update: create a proposed new version; do not overwrite the current version until approved.

#### Product Rules

- Company Understanding covers the company/market context, not detailed Project products/prices.
- Understanding revisions do not use Strategy Revision Allowance.
- User-approved corrections outrank AI inference for the approved Organization baseline.
- Private data never enters the Business Pool by default.
- Projects may generate update suggestions but never auto-rewrite Company Understanding.

#### Responsibility Split


**AI responsibilities**

- Extract and classify business facts.
- Identify missing/conflicting information.
- Generate targeted questions.
- Produce Business Overview and reusable/private classification proposal.
- Never self-approve.


**Deterministic application responsibilities**

- Assemble authorized sources.
- Scan/store files through governed ingestion.
- Persist version/evidence/corrections.
- Apply reusable-data policy.
- Authorize and audit approval.


**Human responsibilities**

- Provide context and documents.
- Review/correct/answer.
- Approve final overview.
- Request later material update.

#### Permissions

- company_understanding.view/manage/approve; business_pool.review where applicable; platform update-approval permission for later material changes.

#### Data Requirements

- CompanyUnderstandingVersion
- CompanyFact
- SourceReference
- ClarificationQuestion
- UserCorrection
- OrganizationKnowledgeItem
- BusinessPoolProposal
- Approval
- AuditEvent

#### UX States

- Empty input
- Uploading/processing
- Analyzing
- Initial review
- Clarification required
- Refining
- Final review
- Approved
- Analysis failed

#### Notifications and Action Center

- Action Item when clarification or final approval is needed.
- In-app/email for prolonged failed analysis or later update approval.

#### Audit Requirements

- Source added/removed reference
- Analysis run/version
- Question/answer
- Correction
- Approval
- Business Pool inclusion/exclusion

#### Analytics Events and Metrics

- Onboarding completion rate/time
- Questions asked
- Corrections rate
- Analysis failure
- Business Pool reusable attributes produced

#### Security and Privacy

- Tenant-scoped retrieval
- Untrusted document/web content
- Sensitive source separation
- No cross-tenant context
- Restricted trace access

#### B2B / B2C Behavior

- B2B/B2C/both is an Organization-level classification only; detailed Project mode is chosen per Project.

#### Formal Product Requirements

| PR ID | Requirement/behavior | Type | BRD trace |
| --- | --- | --- | --- |
| PR-02-001 | Company Understanding covers the company/market context, not detailed Project products/prices. | Product Rule | BR-CMP-001—015, BR-POOL-003, BR-SEC-010 |
| PR-02-002 | Understanding revisions do not use Strategy Revision Allowance. | Product Rule | BR-CMP-001—015, BR-POOL-003, BR-SEC-010 |
| PR-02-003 | User-approved corrections outrank AI inference for the approved Organization baseline. | Product Rule | BR-CMP-001—015, BR-POOL-003, BR-SEC-010 |
| PR-02-004 | Private data never enters the Business Pool by default. | Product Rule | BR-CMP-001—015, BR-POOL-003, BR-SEC-010 |
| PR-02-005 | Projects may generate update suggestions but never auto-rewrite Company Understanding. | Product Rule | BR-CMP-001—015, BR-POOL-003, BR-SEC-010 |
| PR-02-006 | User can complete the analysis-review-clarification-final-review flow without repeating registration fields. | Acceptance Behavior | BR-CMP-001—015, BR-POOL-003, BR-SEC-010 |
| PR-02-007 | Approved version contains source-aware facts and business overview. | Acceptance Behavior | BR-CMP-001—015, BR-POOL-003, BR-SEC-010 |
| PR-02-008 | Private uploaded information is not visible in reusable Business Pool output. | Acceptance Behavior | BR-CMP-001—015, BR-POOL-003, BR-SEC-010 |
| PR-02-009 | Strategy Revision Allowance remains unchanged. | Acceptance Behavior | BR-CMP-001—015, BR-POOL-003, BR-SEC-010 |
| PR-02-010 | A later update preserves the previous approved version. | Acceptance Behavior | BR-CMP-001—015, BR-POOL-003, BR-SEC-010 |

#### Acceptance Criteria

- [ ] **AC-M02-001:** User can complete the analysis-review-clarification-final-review flow without repeating registration fields.
- [ ] **AC-M02-002:** Approved version contains source-aware facts and business overview.
- [ ] **AC-M02-003:** Private uploaded information is not visible in reusable Business Pool output.
- [ ] **AC-M02-004:** Strategy Revision Allowance remains unchanged.
- [ ] **AC-M02-005:** A later update preserves the previous approved version.

### M03 — Organization Dashboard and Administration

**Purpose:** Provide an action-led Organization home and customer-side administration without duplicating Project configuration.

**Mapped BRD requirements:** BR-ADM-003—007, BR-ADM-016—018, BR-UX-001, BR-PLN-012—013

#### Actors

- Organization Owner
- Authorized Organization users
- System

#### Representative User Stories


- **US-M03-001:** As a Organization Owner, I want this capability to provide an action-led Organization home and customer-side administration without duplicating Project configuration.

- **US-M03-002:** As the platform, I need all protected actions in this module to be permissioned, auditable, tenant-scoped, and state-valid.

- **US-M03-003:** As an authorized reviewer, I want clear loading, empty, failure, and recovery behavior so that work never depends on hidden system state.

#### Preconditions

- Organization is active.
- Company Understanding is approved for normal workflow.

#### Inputs

- Organization status
- Projects
- Plan/usage
- Action Items
- Notifications
- Integrations
- Permission scope

#### Outputs

- Organization Dashboard
- Next-action guidance
- Team/permission administration
- Email integration access
- Usage/Plan view
- Organization settings
- Audit view

#### Main Flow

1. After Company Understanding, route user to Organization Dashboard.
2. If no Project exists, primary CTA is Create First Project.
3. If a Project completed onboarding but product sales-support content is incomplete, show Complete Product Content.
4. When products are ready, show Start Recommended Lead Acquisition.
5. Show high-level organization/project statistics and open actions filtered by permission/Project access.
6. Provide navigation to Company, Team & Access, Projects, Email, Action Center, Usage & Plan, Settings, and Audit.

#### Alternative, Exception, and Failure Flows

- No permission for an action: do not show executable control; show read-only status where appropriate.
- Plan limit reached: explain block and direct user to contact platform team.
- Integration missing: show entitled but not configured and create action.
- No open actions: show healthy empty state and next strategic recommendation.

#### Product Rules

- Organization Dashboard is action-led, not a reporting-only page.
- Plan configuration is read-only on Organization side.
- Project-level product/Strategy/channel behavior stays inside the Project.
- Organization-wide hard communication restrictions are inherited by Projects.

#### Responsibility Split


**AI responsibilities**

- May summarize next recommended organizational action but does not directly modify business state.


**Deterministic application responsibilities**

- Filter data by membership/permissions/Project access.
- Compute effective Plan/usage.
- Create/deduplicate actions.
- Render tenant-scoped audit and notification preferences.


**Human responsibilities**

- Manage users/permissions/Project access where authorized.
- Connect email.
- Review usage/settings.
- Follow next actions.

#### Permissions

- organization.view/settings.manage/team.manage/permissions.manage/project.create/integration.email.manage/usage.view/audit.view and Project-specific permissions.

#### Data Requirements

- OrganizationSettings
- Membership
- PermissionGrant
- ProjectAccess
- UsageRecord
- ActionItem
- NotificationPreference
- OrganizationAuditEvent

#### UX States

- New organization
- No Projects
- Actions open
- Healthy/no actions
- Limit warning
- Integration warning
- Permission denied
- Load/error

#### Notifications and Action Center

- In-app/email according to event routing and preferences.
- Dashboard badge counts for open actions/unread notifications.

#### Audit Requirements

- Permission changes
- Project access changes
- Email connection changes
- Settings changes
- Closure request

#### Analytics Events and Metrics

- Dashboard task completion
- Time to first Project
- Product-content completion
- Time to first acquisition

#### Security and Privacy

- No cross-Organization data
- No secret display
- Owner protection
- Audit of permission escalation

#### B2B / B2C Behavior

- Same administration model; Projects determine B2B/B2C behavior.

#### Formal Product Requirements

| PR ID | Requirement/behavior | Type | BRD trace |
| --- | --- | --- | --- |
| PR-03-001 | Organization Dashboard is action-led, not a reporting-only page. | Product Rule | BR-ADM-003—007, BR-ADM-016—018, BR-UX-001, BR-PLN-012—013 |
| PR-03-002 | Plan configuration is read-only on Organization side. | Product Rule | BR-ADM-003—007, BR-ADM-016—018, BR-UX-001, BR-PLN-012—013 |
| PR-03-003 | Project-level product/Strategy/channel behavior stays inside the Project. | Product Rule | BR-ADM-003—007, BR-ADM-016—018, BR-UX-001, BR-PLN-012—013 |
| PR-03-004 | Organization-wide hard communication restrictions are inherited by Projects. | Product Rule | BR-ADM-003—007, BR-ADM-016—018, BR-UX-001, BR-PLN-012—013 |
| PR-03-005 | A new Organization sees Create First Project as primary CTA. | Acceptance Behavior | BR-ADM-003—007, BR-ADM-016—018, BR-UX-001, BR-PLN-012—013 |
| PR-03-006 | Users see only actions/projects allowed by permissions. | Acceptance Behavior | BR-ADM-003—007, BR-ADM-016—018, BR-UX-001, BR-PLN-012—013 |
| PR-03-007 | Plan/usage values match effective subscription snapshot and overrides. | Acceptance Behavior | BR-ADM-003—007, BR-ADM-016—018, BR-UX-001, BR-PLN-012—013 |
| PR-03-008 | Organization settings cannot edit Project Strategy/products. | Acceptance Behavior | BR-ADM-003—007, BR-ADM-016—018, BR-UX-001, BR-PLN-012—013 |
| PR-03-009 | Permission changes are audited and take effect server-side. | Acceptance Behavior | BR-ADM-003—007, BR-ADM-016—018, BR-UX-001, BR-PLN-012—013 |

#### Acceptance Criteria

- [ ] **AC-M03-001:** A new Organization sees Create First Project as primary CTA.
- [ ] **AC-M03-002:** Users see only actions/projects allowed by permissions.
- [ ] **AC-M03-003:** Plan/usage values match effective subscription snapshot and overrides.
- [ ] **AC-M03-004:** Organization settings cannot edit Project Strategy/products.
- [ ] **AC-M03-005:** Permission changes are audited and take effect server-side.

### M04 — Project Creation, Understanding, and Final Onboarding

**Purpose:** Create a versioned, approved Project/Product baseline and channel configuration that downstream agents can safely execute.

**Mapped BRD requirements:** BR-PRJ-001—020, BR-STR-001—017, BR-PLN-004—007

#### Actors

- Authorized Organization user
- Project Understanding Agent
- Strategy Agent
- System

#### Representative User Stories


- **US-M04-001:** As a Authorized Organization user, I want this capability to create a versioned, approved Project/Product baseline and channel configuration that downstream agents can safely execute.

- **US-M04-002:** As the platform, I need all protected actions in this module to be permissioned, auditable, tenant-scoped, and state-valid.

- **US-M04-003:** As an authorized reviewer, I want clear loading, empty, failure, and recovery behavior so that work never depends on hidden system state.

#### Preconditions

- Organization active; Company Understanding approved; Project capacity available; user has project.create.

#### Inputs

- Project name
- Optional description
- Free-text Project brief
- Products/services to sell
- Goal/context
- Brochures/product lists/prices/specifications/previous marketing and sales material
- Clarification answers

#### Outputs

- Project shell
- Project/Product Understanding version
- Confirmed objective
- Product Catalogue proposal
- Approved Strategy version
- Channel/Sales AI settings
- LIVE Project

#### Main Flow

1. Create Project shell in ONBOARDING.
2. Collect Project brief and sources.
3. Project Understanding Agent analyzes and asks targeted questions.
4. Present Project Scope, objective, products, pricing/constraints, assumptions, and missing/conflicts.
5. User comments/corrects through the AI interaction and approves Project/Product Understanding.
6. System creates the Product Catalogue baseline.
7. User explicitly chooses Plan the Strategy.
8. Strategy Agent generates the structured package or returns Needs Information without consuming allowance.
9. User approves or requests a limited revision.
10. After Strategy acceptance, user configures Project and channel communication settings.
11. Present final Project onboarding summary.
12. Authorized user approves; Project becomes LIVE and routes to Project Dashboard.

#### Alternative, Exception, and Failure Flows

- Project limit reached: block creation; existing Projects unaffected.
- Critical product data missing: clarification loop; Strategy blocked.
- Zero Strategy revisions: latest Strategy can be approved but not regenerated.
- Missing channel integration: allow final onboarding, mark Action Required, block affected execution.
- User identifies error before final approval: return to the relevant review stage.
- Material product error after LIVE: no normal edit; requires future exceptional recovery/change process.

#### Product Rules

- Each Project has one primary B2B or B2C mode.
- Products cannot be added after final onboarding.
- Core product facts/Project objective/Strategy are versioned.
- Strategy approval alone does not publish campaigns/start discovery.
- Project lifecycle is ONBOARDING/LIVE/ENDED.

#### Responsibility Split


**AI responsibilities**

- Project Agent extracts/asks/proposes.
- Strategy Agent produces package/revisions.
- Neither self-approves or launches execution.


**Deterministic application responsibilities**

- Enforce Plan/permissions.
- Isolate sources/context.
- Persist versions and approvals.
- Count revisions.
- Create Product Catalogue.
- Manage Project state and actions.


**Human responsibilities**

- Provide sources/answers.
- Review/correct/approve understanding.
- Review/revise/approve Strategy.
- Configure channels.
- Final approve.

#### Permissions

- project.create/view; project.onboarding.manage/approve; strategy.generate/revise/approve; channel.settings.manage.

#### Data Requirements

- Project
- ProjectUnderstandingVersion
- ProductUnderstandingVersion
- ProjectObjective
- ProjectSource
- ProductCatalogueItem
- StrategyVersion
- ChannelSettings
- Approval
- UsageRecord

#### UX States

- Draft inputs
- Uploading/analyzing
- Clarification
- Understanding review
- Strategy generating
- Strategy review
- No revisions
- Channel settings
- Final review
- LIVE
- Failure

#### Notifications and Action Center

- Action Items for missing data, reviews, Strategy approval, channel readiness, final approval.
- Email for long-running/failed approval-required items.

#### Audit Requirements

- Project creation
- Source upload
- Agent runs
- Corrections
- Understanding approval
- Strategy revisions/approval
- Settings
- LIVE transition

#### Analytics Events and Metrics

- Time to LIVE
- Questions/revisions
- Strategy approval rate
- Missing integration count
- Project abandonment

#### Security and Privacy

- Project isolation
- Untrusted inputs
- No privileged client state
- Version/reference integrity
- Permission/entitlement checks

#### B2B / B2C Behavior

- B2B and B2C use the same onboarding shell but different Agent-generated information requirements, ICP/personas, scoring, and acquisition recommendations.

#### Formal Product Requirements

| PR ID | Requirement/behavior | Type | BRD trace |
| --- | --- | --- | --- |
| PR-04-001 | Each Project has one primary B2B or B2C mode. | Product Rule | BR-PRJ-001—020, BR-STR-001—017, BR-PLN-004—007 |
| PR-04-002 | Products cannot be added after final onboarding. | Product Rule | BR-PRJ-001—020, BR-STR-001—017, BR-PLN-004—007 |
| PR-04-003 | Core product facts/Project objective/Strategy are versioned. | Product Rule | BR-PRJ-001—020, BR-STR-001—017, BR-PLN-004—007 |
| PR-04-004 | Strategy approval alone does not publish campaigns/start discovery. | Product Rule | BR-PRJ-001—020, BR-STR-001—017, BR-PLN-004—007 |
| PR-04-005 | Project lifecycle is ONBOARDING/LIVE/ENDED. | Product Rule | BR-PRJ-001—020, BR-STR-001—017, BR-PLN-004—007 |
| PR-04-006 | Cannot generate Strategy before approved understanding. | Acceptance Behavior | BR-PRJ-001—020, BR-STR-001—017, BR-PLN-004—007 |
| PR-04-007 | Revision counter changes only for valid user-requested revisions. | Acceptance Behavior | BR-PRJ-001—020, BR-STR-001—017, BR-PLN-004—007 |
| PR-04-008 | Final approval creates a LIVE Project with one approved Strategy and channel settings. | Acceptance Behavior | BR-PRJ-001—020, BR-STR-001—017, BR-PLN-004—007 |
| PR-04-009 | No acquisition side effect occurs from final approval. | Acceptance Behavior | BR-PRJ-001—020, BR-STR-001—017, BR-PLN-004—007 |
| PR-04-010 | Product Catalogue matches approved product baseline. | Acceptance Behavior | BR-PRJ-001—020, BR-STR-001—017, BR-PLN-004—007 |

#### Acceptance Criteria

- [ ] **AC-M04-001:** Cannot generate Strategy before approved understanding.
- [ ] **AC-M04-002:** Revision counter changes only for valid user-requested revisions.
- [ ] **AC-M04-003:** Final approval creates a LIVE Project with one approved Strategy and channel settings.
- [ ] **AC-M04-004:** No acquisition side effect occurs from final approval.
- [ ] **AC-M04-005:** Product Catalogue matches approved product baseline.

### M05 — Project Dashboard and Product Catalogue

**Purpose:** Give users an operational Project home and enrich approved products with safe sales-support content.

**Mapped BRD requirements:** BR-PRJ-011—014, BR-SAL-007, BR-UX-002

#### Actors

- Authorized Project users
- System
- Sales Agent

#### Representative User Stories


- **US-M05-001:** As a Authorized Project users, I want this capability to give users an operational Project home and enrich approved products with safe sales-support content.

- **US-M05-002:** As the platform, I need all protected actions in this module to be permissioned, auditable, tenant-scoped, and state-valid.

- **US-M05-003:** As an authorized reviewer, I want clear loading, empty, failure, and recovery behavior so that work never depends on hidden system state.

#### Preconditions

- Project exists; Product Catalogue baseline generated.

#### Inputs

- Project state
- Strategy
- Products
- Campaigns
- Lead/Conversation statistics
- Open actions
- Descriptions/media/documents

#### Outputs

- Project Dashboard
- Product Catalogue pages
- Approved media/content available to Sales Agent

#### Main Flow

1. Display Project overview, status, Strategy version, products, selected channels, and core statistics.
2. Display action-required items such as incomplete product content, campaign ready/publish, handover, follow-up approval, integration failure.
3. Product list is read from approved onboarding baseline.
4. Authorized user opens a product and adds/edits sales-support description, images, video, brochures, or documents.
5. System validates files/content and records approval/source metadata.
6. Sales Agent retrieves the smallest relevant approved content set during conversations.

#### Alternative, Exception, and Failure Flows

- No sales-support content: show completion action but allow Project LIVE.
- Attempt to add new product: block and explain baseline rule.
- Attempt to edit strategy-critical field: read-only; direct user to controlled future change process.
- Unsupported/malicious file: reject/quarantine with actionable message.
- Deleted media referenced by a future message: prevent send and request replacement.

#### Product Rules

- Post-onboarding content never silently regenerates Strategy.
- Sales Agent may use only correct Project/product approved content.
- Material product changes are not normal MVP behavior.
- Coupons/offers automation is deferred.

#### Responsibility Split


**AI responsibilities**

- May suggest missing content and select relevant approved media for a lead.
- Cannot alter strategy-critical facts.


**Deterministic application responsibilities**

- Separate baseline fields from sales-support fields.
- Authorize upload/read.
- Validate/scan/store files.
- Version content and preserve references.
- Emit audit/analytics.


**Human responsibilities**

- Add and maintain sales-support content.
- Review actions.
- Do not add core products.

#### Permissions

- product.view; product.content.manage; product.content.approve if separated.

#### Data Requirements

- ProductCatalogueItem
- ProductBaselineSnapshot
- SalesContentAsset
- AssetVersion
- AssetApproval
- ProductUsageReference

#### UX States

- Empty content
- Upload progress
- Processing
- Ready
- Rejected/failed
- Read-only baseline
- Permission denied

#### Notifications and Action Center

- Action Item for incomplete recommended content or rejected asset.
- No email for routine successful uploads.

#### Audit Requirements

- Asset add/update/remove
- Approval
- AI use/send reference
- Blocked product-add attempt

#### Analytics Events and Metrics

- Product completeness
- Asset usage in conversations
- Missing content requests
- Failed assets

#### Security and Privacy

- Tenant/Project-scoped signed access
- File validation/scanning
- No cross-product retrieval
- Sensitive document controls

#### B2B / B2C Behavior

- Same content model. B2C may emphasize visual media; B2B may emphasize brochures/specifications, but no fixed assumption.

#### Formal Product Requirements

| PR ID | Requirement/behavior | Type | BRD trace |
| --- | --- | --- | --- |
| PR-05-001 | Post-onboarding content never silently regenerates Strategy. | Product Rule | BR-PRJ-011—014, BR-SAL-007, BR-UX-002 |
| PR-05-002 | Sales Agent may use only correct Project/product approved content. | Product Rule | BR-PRJ-011—014, BR-SAL-007, BR-UX-002 |
| PR-05-003 | Material product changes are not normal MVP behavior. | Product Rule | BR-PRJ-011—014, BR-SAL-007, BR-UX-002 |
| PR-05-004 | Coupons/offers automation is deferred. | Product Rule | BR-PRJ-011—014, BR-SAL-007, BR-UX-002 |
| PR-05-005 | Core product fields are read-only after onboarding. | Acceptance Behavior | BR-PRJ-011—014, BR-SAL-007, BR-UX-002 |
| PR-05-006 | Adding an image does not change Strategy version. | Acceptance Behavior | BR-PRJ-011—014, BR-SAL-007, BR-UX-002 |
| PR-05-007 | Sales AI can send only assets linked to the correct product/Project. | Acceptance Behavior | BR-PRJ-011—014, BR-SAL-007, BR-UX-002 |
| PR-05-008 | New product creation is blocked. | Acceptance Behavior | BR-PRJ-011—014, BR-SAL-007, BR-UX-002 |
| PR-05-009 | Dashboard action deep-links to the incomplete product. | Acceptance Behavior | BR-PRJ-011—014, BR-SAL-007, BR-UX-002 |

#### Acceptance Criteria

- [ ] **AC-M05-001:** Core product fields are read-only after onboarding.
- [ ] **AC-M05-002:** Adding an image does not change Strategy version.
- [ ] **AC-M05-003:** Sales AI can send only assets linked to the correct product/Project.
- [ ] **AC-M05-004:** New product creation is blocked.
- [ ] **AC-M05-005:** Dashboard action deep-links to the incomplete product.

### M06 — Strategy Planning and Approval

**Purpose:** Produce a structured commercial Strategy that downstream acquisition and sales agents can execute consistently.

**Mapped BRD requirements:** BR-STR-001—018, BR-AGT-009

#### Actors

- Authorized Project user
- Strategy Agent
- System

#### Representative User Stories


- **US-M06-001:** As a Authorized Project user, I want this capability to produce a structured commercial Strategy that downstream acquisition and sales agents can execute consistently.

- **US-M06-002:** As the platform, I need all protected actions in this module to be permissioned, auditable, tenant-scoped, and state-valid.

- **US-M06-003:** As an authorized reviewer, I want clear loading, empty, failure, and recovery behavior so that work never depends on hidden system state.

#### Preconditions

- Approved Project/Product Understanding and confirmed objective.
- Project has remaining revision allowance for any requested revision.

#### Inputs

- Approved understanding versions
- Organization relevant context
- Project mode/geography/objective
- Products
- Supported channels/capabilities
- User revision instruction

#### Outputs

- Strategy Version
- ICP/personas
- Pain/gain and product fit
- Sales/marketing/acquisition direction
- Qualification framework
- Meta concepts
- Conversion/handover model
- Risks/assumptions

#### Main Flow

1. User clicks Plan the Strategy.
2. System validates complete approved inputs.
3. Strategy Agent produces one structured package.
4. UI renders an executive summary plus detailed sections.
5. User approves or enters a revision instruction.
6. System validates remaining allowance; Agent applies requested change and dependent updates.
7. System stores a new version and change summary.
8. After acceptance, route to channel/Sales AI configuration.

#### Alternative, Exception, and Failure Flows

- Missing critical data: return Needs Information; no allowance consumed.
- Agent failure: retry within policy and preserve inputs.
- Revision exceeds allowance: block regeneration, retain latest Strategy.
- User change would require new product: reject and direct to Project/Product baseline rule.
- Unsupported acquisition recommendation: show unavailable capability/action required.

#### Product Rules

- One coherent approval package; structured internal components.
- Project-specific qualification weights must total/validate within platform bounds.
- No invented commercial terms.
- Strategy changes after Project learning require a separate human-approved version.

#### Responsibility Split


**AI responsibilities**

- Research/reason about market and ICP.
- Generate Strategy/qualification/acquisition recommendations.
- Explain risks and assumptions.


**Deterministic application responsibilities**

- Validate inputs/allowance/schema.
- Persist versions and traceability.
- Expose downstream components by contract.
- Audit approval.


**Human responsibilities**

- Review, instruct revision, approve.
- Configure communication settings after acceptance.

#### Permissions

- strategy.view/generate/revise/approve; project access.

#### Data Requirements

- StrategyVersion
- ICP
- Persona
- PainGainMap
- ProductAudienceFit
- QualificationFramework
- AcquisitionDirection
- CampaignConcept
- StrategyApproval

#### UX States

- Generating
- Needs information
- Review
- Revision limit warning
- Approved
- Agent failure

#### Notifications and Action Center

- Action Item for Strategy review/approval or missing information.
- Email based on preference/urgency.

#### Audit Requirements

- Generation/version
- User instruction
- Allowance consumption
- Approval
- Downstream use

#### Analytics Events and Metrics

- Time/generation
- Revision count
- Approval rate
- Later recommendation acceptance
- Downstream conversion by Strategy version

#### Security and Privacy

- Approved context only
- No provider side effects
- No cross-tenant market/private inputs
- Prompt/version trace

#### B2B / B2C Behavior

- B2B: separate company ICP and buying personas. B2C: consumer personas and consumer qualification.

#### Formal Product Requirements

| PR ID | Requirement/behavior | Type | BRD trace |
| --- | --- | --- | --- |
| PR-06-001 | One coherent approval package; structured internal components. | Product Rule | BR-STR-001—018, BR-AGT-009 |
| PR-06-002 | Project-specific qualification weights must total/validate within platform bounds. | Product Rule | BR-STR-001—018, BR-AGT-009 |
| PR-06-003 | No invented commercial terms. | Product Rule | BR-STR-001—018, BR-AGT-009 |
| PR-06-004 | Strategy changes after Project learning require a separate human-approved version. | Product Rule | BR-STR-001—018, BR-AGT-009 |
| PR-06-005 | Output contains all mandatory sections. | Acceptance Behavior | BR-STR-001—018, BR-AGT-009 |
| PR-06-006 | Missing data does not consume a revision. | Acceptance Behavior | BR-STR-001—018, BR-AGT-009 |
| PR-06-007 | Requested revision creates one new version and consumes one allowance. | Acceptance Behavior | BR-STR-001—018, BR-AGT-009 |
| PR-06-008 | Approved Strategy references exact understanding versions. | Acceptance Behavior | BR-STR-001—018, BR-AGT-009 |
| PR-06-009 | No campaign/discovery is started automatically. | Acceptance Behavior | BR-STR-001—018, BR-AGT-009 |

#### Acceptance Criteria

- [ ] **AC-M06-001:** Output contains all mandatory sections.
- [ ] **AC-M06-002:** Missing data does not consume a revision.
- [ ] **AC-M06-003:** Requested revision creates one new version and consumes one allowance.
- [ ] **AC-M06-004:** Approved Strategy references exact understanding versions.
- [ ] **AC-M06-005:** No campaign/discovery is started automatically.

### M07 — Lead Acquisition Planning

**Purpose:** Select and approve the best available source mix—Web, Meta, or Both—and produce source-specific execution briefs.

**Mapped BRD requirements:** BR-ACQ-001—003, BR-ACQ-016, BR-STR-007, BR-AGT-004

#### Actors

- Lead Discovery Agent
- Authorized Project user
- System

#### Representative User Stories


- **US-M07-001:** As a Lead Discovery Agent, I want this capability to select and approve the best available source mix—Web, Meta, or Both—and produce source-specific execution briefs.

- **US-M07-002:** As the platform, I need all protected actions in this module to be permissioned, auditable, tenant-scoped, and state-valid.

- **US-M07-003:** As an authorized reviewer, I want clear loading, empty, failure, and recovery behavior so that work never depends on hidden system state.

#### Preconditions

- Project is LIVE; Strategy approved; user has acquisition permission; relevant feature entitled.

#### Inputs

- Strategy acquisition direction
- ICP/personas
- Products
- Geography
- Project Lead Limit
- Integration readiness
- Existing Project leads/runs
- Platform policies

#### Outputs

- Lead Acquisition Plan
- Web Discovery Plan
- Meta Acquisition Brief
- Expected source role/volume
- Exclusions
- Readiness actions

#### Main Flow

1. User starts Lead Acquisition.
2. Lead Discovery Agent evaluates audience availability and supported sources.
3. Agent recommends Web, Meta, or Both and explains why.
4. Agent produces source-specific plans.
5. User reviews and instructs adjustments if needed.
6. User approves the plan.
7. System starts Web Discovery and/or exposes Meta campaign recommendations for Campaign Agent.

#### Alternative, Exception, and Failure Flows

- Recommended source not entitled: show Plan limitation and prevent execution.
- Entitled but not configured: allow plan approval; create Action Required; block affected execution.
- User removes one source: Agent updates assumptions/expected contribution.
- No viable source: Needs Human Review with reasons rather than fabricate plan.

#### Product Rules

- B2B/B2C influences but does not hard-code source.
- Meta inbound volume does not use Web candidate multiplier.
- Fresh Web discovery first under current policy.
- Plan approval does not itself spend or send.

#### Responsibility Split


**AI responsibilities**

- Analyze source suitability and produce plans.
- Explain evidence/assumptions.
- Do not call provider publish/send directly.


**Deterministic application responsibilities**

- Resolve effective entitlements/readiness.
- Persist plan/version.
- Authorize execution.
- Create actions.


**Human responsibilities**

- Review, adjust, approve, initiate source workflow.

#### Permissions

- acquisition.plan.view/manage/approve; campaign.prepare; discovery.run.

#### Data Requirements

- AcquisitionPlan
- SourcePlan
- SourceReadiness
- AcquisitionApproval
- ProjectLeadLimitSnapshot

#### UX States

- Analyzing
- Recommendation
- Review
- Missing integration
- No viable source
- Approved
- Execution started

#### Notifications and Action Center

- Action Item for missing integration or approval.
- In-app/email for blocked acquisition.

#### Audit Requirements

- Plan generation/approval
- Source selected
- Readiness block
- Execution handoff

#### Analytics Events and Metrics

- Source recommendation distribution
- Plan-to-execution time
- Source contribution
- Plan accuracy vs downstream outcomes

#### Security and Privacy

- No unapproved provider use
- Context isolation
- No private Master Pool details in user-facing plan

#### B2B / B2C Behavior

- B2B may favor Web; B2C may favor Meta; both remain evidence-based recommendations.

#### Formal Product Requirements

| PR ID | Requirement/behavior | Type | BRD trace |
| --- | --- | --- | --- |
| PR-07-001 | B2B/B2C influences but does not hard-code source. | Product Rule | BR-ACQ-001—003, BR-ACQ-016, BR-STR-007, BR-AGT-004 |
| PR-07-002 | Meta inbound volume does not use Web candidate multiplier. | Product Rule | BR-ACQ-001—003, BR-ACQ-016, BR-STR-007, BR-AGT-004 |
| PR-07-003 | Fresh Web discovery first under current policy. | Product Rule | BR-ACQ-001—003, BR-ACQ-016, BR-STR-007, BR-AGT-004 |
| PR-07-004 | Plan approval does not itself spend or send. | Product Rule | BR-ACQ-001—003, BR-ACQ-016, BR-STR-007, BR-AGT-004 |
| PR-07-005 | Plan clearly states Web, Meta, or Both and rationale. | Acceptance Behavior | BR-ACQ-001—003, BR-ACQ-016, BR-STR-007, BR-AGT-004 |
| PR-07-006 | User can approve without launching spend. | Acceptance Behavior | BR-ACQ-001—003, BR-ACQ-016, BR-STR-007, BR-AGT-004 |
| PR-07-007 | Missing integration creates an action and blocks only affected path. | Acceptance Behavior | BR-ACQ-001—003, BR-ACQ-016, BR-STR-007, BR-AGT-004 |
| PR-07-008 | Source-specific brief contains all inputs required by downstream Agent. | Acceptance Behavior | BR-ACQ-001—003, BR-ACQ-016, BR-STR-007, BR-AGT-004 |
| PR-07-009 | Fresh-first policy is visible for Web. | Acceptance Behavior | BR-ACQ-001—003, BR-ACQ-016, BR-STR-007, BR-AGT-004 |

#### Acceptance Criteria

- [ ] **AC-M07-001:** Plan clearly states Web, Meta, or Both and rationale.
- [ ] **AC-M07-002:** User can approve without launching spend.
- [ ] **AC-M07-003:** Missing integration creates an action and blocks only affected path.
- [ ] **AC-M07-004:** Source-specific brief contains all inputs required by downstream Agent.
- [ ] **AC-M07-005:** Fresh-first policy is visible for Web.

### M08 — Web/Google Lead Discovery

**Purpose:** Execute an approved fresh-discovery plan, create contactable Master Leads with provenance, and allocate the best Project Leads within limits.

**Mapped BRD requirements:** BR-ACQ-004—012, BR-POOL-001—014, BR-QLF-001—007

#### Actors

- Lead Discovery Agent
- Lead Enrichment Agent
- System
- Authorized Project user

#### Representative User Stories


- **US-M08-001:** As a Lead Discovery Agent, I want this capability to execute an approved fresh-discovery plan, create contactable Master Leads with provenance, and allocate the best Project Leads within limits.

- **US-M08-002:** As the platform, I need all protected actions in this module to be permissioned, auditable, tenant-scoped, and state-valid.

- **US-M08-003:** As an authorized reviewer, I want clear loading, empty, failure, and recovery behavior so that work never depends on hidden system state.

#### Preconditions

- Approved Web Discovery Plan
- Web/Search entitlement and healthy platform integration
- Project LIVE

#### Inputs

- Target business/person criteria
- Query families
- Geography
- Exclusions
- Required contact methods
- Candidate target
- Allowed providers/tools

#### Outputs

- Research Candidates
- Captured Master Leads
- Discovery Run result
- Duplicates/matches
- Project Lead candidates
- Stop reason

#### Main Flow

1. System calculates candidate target from effective Project Lead Limit × configurable multiplier.
2. Lead Discovery Agent executes approved query/source sequence through governed tools.
3. Raw results become Research Candidates with provenance.
4. System/Enrichment attempts identity/contact resolution.
5. Contactable people create/match Master Leads; companies create/match Business Pool entities.
6. Enrichment gathers purpose-limited fit evidence.
7. Initial scoring ranks valid candidates.
8. System allocates highest-ranked candidates up to Project Lead Limit.
9. If insufficient, query eligible Master Pool matches and refresh as needed.
10. Persist Discovery Run summary and hand allocated leads to Lead Strategist.

#### Alternative, Exception, and Failure Flows

- Provider limit/failure: stop/pause with reason and action.
- Excessive duplicates/irrelevance: adapt approved query strategy within run bounds; stop if exhausted.
- No contact method: keep Research Candidate; do not create active Lead.
- Identity match medium confidence: keep separate/pending review; no unsafe merge.
- Fewer valid candidates than limit: allocate fewer; never fabricate contact data.
- Project limit reached before completion: store captured Master Leads but do not allocate beyond limit.

#### Product Rules

- Every captured contactable person enters Master Lead DB.
- Fresh discovery occurs before Master Pool supplementation.
- Search criteria target practical B2B influencers.
- All source/query/run provenance is retained.
- Discovery loops/time/cost are bounded.

#### Responsibility Split


**AI responsibilities**

- Plan/execute approved search reasoning.
- Adapt query families within boundaries.
- Return structured candidates/provenance/stop reason.


**Deterministic application responsibilities**

- Provider tools, rate/cost limits, URL/source policy, identity resolution, persistence, scoring, allocation, audit.


**Human responsibilities**

- Approve plan/start/stop run; review results and issues.

#### Permissions

- discovery.run/view/stop; Master Pool details limited by Project use and permissions.

#### Data Requirements

- DiscoveryRun
- SearchQueryFamily
- ResearchCandidate
- SourceEvidence
- BusinessPoolEntity
- MasterLead
- ProjectLeadAllocation
- StopReason

#### UX States

- Ready
- Running/progress
- Partial results
- Stopped
- Failed
- Completed
- Insufficient valid leads

#### Notifications and Action Center

- Action Item/email for failed/stalled run or result requiring review.
- Completion in-app notification.

#### Audit Requirements

- Plan/run/version
- Queries/providers
- Result/source
- Identity match
- Allocation
- Stop/failure

#### Analytics Events and Metrics

- Candidates per query
- Contactable rate
- Duplicate rate
- Allocation rate
- Provider cost/latency
- Qualified/conversion downstream

#### Security and Privacy

- Approved source/tool only
- URL/SSRF controls
- Source licensing flags
- No private cross-tenant data
- Rate limits

#### B2B / B2C Behavior

- B2B company-first/person-influence search. B2C Web discovery is allowed only when Strategy and policy deem it appropriate.

#### Formal Product Requirements

| PR ID | Requirement/behavior | Type | BRD trace |
| --- | --- | --- | --- |
| PR-08-001 | Every captured contactable person enters Master Lead DB. | Product Rule | BR-ACQ-004—012, BR-POOL-001—014, BR-QLF-001—007 |
| PR-08-002 | Fresh discovery occurs before Master Pool supplementation. | Product Rule | BR-ACQ-004—012, BR-POOL-001—014, BR-QLF-001—007 |
| PR-08-003 | Search criteria target practical B2B influencers. | Product Rule | BR-ACQ-004—012, BR-POOL-001—014, BR-QLF-001—007 |
| PR-08-004 | All source/query/run provenance is retained. | Product Rule | BR-ACQ-004—012, BR-POOL-001—014, BR-QLF-001—007 |
| PR-08-005 | Discovery loops/time/cost are bounded. | Product Rule | BR-ACQ-004—012, BR-POOL-001—014, BR-QLF-001—007 |
| PR-08-006 | Candidate target follows effective limit and configured multiplier. | Acceptance Behavior | BR-ACQ-004—012, BR-POOL-001—014, BR-QLF-001—007 |
| PR-08-007 | Non-contactable results do not become active leads. | Acceptance Behavior | BR-ACQ-004—012, BR-POOL-001—014, BR-QLF-001—007 |
| PR-08-008 | Every allocated lead has provenance and at least one contact method. | Acceptance Behavior | BR-ACQ-004—012, BR-POOL-001—014, BR-QLF-001—007 |
| PR-08-009 | No Project receives more active Project Leads than allowance. | Acceptance Behavior | BR-ACQ-004—012, BR-POOL-001—014, BR-QLF-001—007 |
| PR-08-010 | Run records a deterministic stop reason. | Acceptance Behavior | BR-ACQ-004—012, BR-POOL-001—014, BR-QLF-001—007 |

#### Acceptance Criteria

- [ ] **AC-M08-001:** Candidate target follows effective limit and configured multiplier.
- [ ] **AC-M08-002:** Non-contactable results do not become active leads.
- [ ] **AC-M08-003:** Every allocated lead has provenance and at least one contact method.
- [ ] **AC-M08-004:** No Project receives more active Project Leads than allowance.
- [ ] **AC-M08-005:** Run records a deterministic stop reason.

### M09 — Meta Campaign Creation, Publication, and Monitoring

**Purpose:** Turn the approved Meta brief into a complete campaign, require human publication, and connect campaign performance to lead quality and conversion.

**Mapped BRD requirements:** BR-CAM-001—018, BR-ACQ-013—020, BR-SEC-022

#### Actors

- Campaign Agent
- Authorized campaign user
- Campaign Service
- Meta Adapter
- System

#### Representative User Stories


- **US-M09-001:** As a Campaign Agent, I want this capability to turn the approved Meta brief into a complete campaign, require human publication, and connect campaign performance to lead quality and conversion.

- **US-M09-002:** As the platform, I need all protected actions in this module to be permissioned, auditable, tenant-scoped, and state-valid.

- **US-M09-003:** As an authorized reviewer, I want clear loading, empty, failure, and recovery behavior so that work never depends on hidden system state.

#### Preconditions

- Project LIVE
- Approved Meta Acquisition Brief
- Meta feature entitled
- Organization Meta capabilities configured for publish
- User has campaign permissions

#### Inputs

- Product/Strategy/ICP/persona
- Audience/geography
- Budget
- End date
- Lead path/form
- Ad copy
- Creative brief
- Uploaded media
- Organization Meta account configuration

#### Outputs

- Campaign proposal
- Approved lead capture
- Creative guide/media review
- Provider campaign IDs/status
- Performance metrics
- Generated leads
- AI recommendations

#### Main Flow

1. User selects a suggested campaign.
2. Campaign Agent generates configuration, audience, lead path/form questions, copy, and detailed creative production guide.
3. User reviews/edits settings and explicitly approves lead capture.
4. User creates/uploads media.
5. Campaign Agent reviews media and commercial coherence.
6. System validates permissions, entitlement, spend, mandatory end date, required fields/media, integration readiness, and provider mapping.
7. User explicitly clicks Publish.
8. Campaign Service executes idempotent Meta Adapter operation and stores provider IDs/status.
9. Published Campaign page displays performance and generated leads.
10. Campaign Agent analyzes performance and downstream lead quality and recommends actions.
11. Authorized user may increase budget, extend end date, pause, or resume.

#### Alternative, Exception, and Failure Flows

- Missing integration during Project onboarding: Project may be LIVE but publish is blocked with Action Required.
- Provider rejects publish: PUBLISH_FAILED; explain, preserve draft, allow correction/retry.
- Duplicate publish request: idempotent return of existing operation/result.
- Budget/end date edit violates policy/limit: block with explanation.
- Media inconsistency: warn/recommend; human decides unless deterministic/provider validation blocks.
- Project lead limit reached: campaign may keep receiving leads; overflow enters Master Pool but no autonomous Project sales.
- Campaign end: retain analytics; late existing lead conversations continue; duplicate to new Draft for reuse.

#### Product Rules

- User publication is mandatory.
- AI cannot independently increase spend or extend.
- End date is mandatory.
- Primary paid Meta conversation destination is WhatsApp in MVP.
- Ended campaigns do not restart in place.
- Recommended campaign lifecycle D-274 requires final confirmation.

#### Responsibility Split


**AI responsibilities**

- Generate/assess campaign intelligence and creative instructions.
- Review uploaded media.
- Interpret performance/failures.
- Recommend, never authorize spend.


**Deterministic application responsibilities**

- Validate permissions/entitlements/spend/dates/media/provider.
- Call adapter.
- Persist state/provider IDs/webhooks.
- Idempotency/audit.
- Create leads/actions.


**Human responsibilities**

- Approve form/settings/media.
- Explicitly publish.
- Approve operational changes.
- Respond to recommendations.

#### Permissions

- campaign.view/prepare/publish/increase_budget/extend/pause_resume; Project access.

#### Data Requirements

- Campaign
- CampaignVersion
- AudienceConfig
- LeadCaptureConfig
- CreativeBrief
- CreativeAsset
- PublishOperation
- ProviderReference
- CampaignMetric
- CampaignRecommendation

#### UX States

- Draft
- Ready
- Publishing
- Active
- Paused
- Ended
- Publish failed
- Integration unavailable
- Limit warning

#### Notifications and Action Center

- Ready/publish failure/end date/limit/recommendation notifications based on permissions.
- Email for urgent failure/expiry.

#### Audit Requirements

- Configuration changes
- Lead form approval
- Media review
- Publish
- Budget/end-date/pause/resume
- Provider events
- Agent recommendations

#### Analytics Events and Metrics

- Spend/reach/impressions/clicks/leads/CPL where available
- Qualified handovers/conversions
- Lead quality by campaign
- Failure rate

#### Security and Privacy

- Explicit spend action
- Credential protection
- Webhook verification/replay
- Idempotency
- Policy validation
- No AI provider credentials

#### B2B / B2C Behavior

- Audience/creative varies by Project mode. B2C must not inherit B2B role targeting.

#### Formal Product Requirements

| PR ID | Requirement/behavior | Type | BRD trace |
| --- | --- | --- | --- |
| PR-09-001 | User publication is mandatory. | Product Rule | BR-CAM-001—018, BR-ACQ-013—020, BR-SEC-022 |
| PR-09-002 | AI cannot independently increase spend or extend. | Product Rule | BR-CAM-001—018, BR-ACQ-013—020, BR-SEC-022 |
| PR-09-003 | End date is mandatory. | Product Rule | BR-CAM-001—018, BR-ACQ-013—020, BR-SEC-022 |
| PR-09-004 | Primary paid Meta conversation destination is WhatsApp in MVP. | Product Rule | BR-CAM-001—018, BR-ACQ-013—020, BR-SEC-022 |
| PR-09-005 | Ended campaigns do not restart in place. | Product Rule | BR-CAM-001—018, BR-ACQ-013—020, BR-SEC-022 |
| PR-09-006 | Recommended campaign lifecycle D-274 requires final confirmation. | Product Rule | BR-CAM-001—018, BR-ACQ-013—020, BR-SEC-022 |
| PR-09-007 | Publish is impossible without explicit authorized click. | Acceptance Behavior | BR-CAM-001—018, BR-ACQ-013—020, BR-SEC-022 |
| PR-09-008 | Campaign requires an end date and complete approved media. | Acceptance Behavior | BR-CAM-001—018, BR-ACQ-013—020, BR-SEC-022 |
| PR-09-009 | Provider failure is actionable and retry-safe. | Acceptance Behavior | BR-CAM-001—018, BR-ACQ-013—020, BR-SEC-022 |
| PR-09-010 | Generated leads retain campaign attribution. | Acceptance Behavior | BR-CAM-001—018, BR-ACQ-013—020, BR-SEC-022 |
| PR-09-011 | AI cannot execute budget increase/extension. | Acceptance Behavior | BR-CAM-001—018, BR-ACQ-013—020, BR-SEC-022 |
| PR-09-012 | Overflow lead behavior respects Project Lead Limit. | Acceptance Behavior | BR-CAM-001—018, BR-ACQ-013—020, BR-SEC-022 |

#### Acceptance Criteria

- [ ] **AC-M09-001:** Publish is impossible without explicit authorized click.
- [ ] **AC-M09-002:** Campaign requires an end date and complete approved media.
- [ ] **AC-M09-003:** Provider failure is actionable and retry-safe.
- [ ] **AC-M09-004:** Generated leads retain campaign attribution.
- [ ] **AC-M09-005:** AI cannot execute budget increase/extension.
- [ ] **AC-M09-006:** Overflow lead behavior respects Project Lead Limit.

### M10 — Channel Integration Readiness and Meta Messaging

**Purpose:** Provide reliable Organization-scoped WhatsApp, Instagram, and Facebook messaging capabilities with readiness, fallback, and failure controls.

**Mapped BRD requirements:** BR-ACQ-016—019, BR-EML-009—012, BR-ADM-008—011, BR-SEC-007—009

#### Actors

- Platform Super Admin
- Organization user
- System
- Sales Agent
- Lead Strategist
- Provider adapters

#### Representative User Stories


- **US-M10-001:** As a Platform Super Admin, I want this capability to provide reliable Organization-scoped WhatsApp, Instagram, and Facebook messaging capabilities with readiness, fallback, and failure controls.

- **US-M10-002:** As the platform, I need all protected actions in this module to be permissioned, auditable, tenant-scoped, and state-valid.

- **US-M10-003:** As an authorized reviewer, I want clear loading, empty, failure, and recovery behavior so that work never depends on hidden system state.

#### Preconditions

- Platform integration definition exists.
- Super Admin has Organization-specific provider account information.
- Project has selected/approved channels.

#### Inputs

- Organization provider credentials/account IDs
- Capability permissions
- Project channel settings
- Lead contact identifiers
- Provider health/status

#### Outputs

- Capability status per Organization
- Channel readiness
- Provider message events
- Fallback/pause actions
- Action Items

#### Main Flow

1. Super Admin opens Organization → Integrations and configures Meta Ads, WhatsApp, Instagram, and Facebook capabilities independently.
2. System stores secret references, validates account/permission information, and tests connection.
3. Organization/Project UI displays entitlement and readiness without exposing secrets.
4. Before each action, system checks capability status and communication eligibility.
5. Incoming messages/webhooks are verified and resolved to Organization, Project, Master Lead, and Project Lead.
6. Sales AI handles eligible active conversations.
7. If a channel fails, Lead Strategist/system evaluates approved alternatives.
8. If valid alternative exists, preserve context and transition appropriately; otherwise pause affected activity and alert.

#### Alternative, Exception, and Failure Flows

- Token expired/reauthorization required: set state, block sends, create Action Item.
- Capability partially available: enable only validated functions.
- Unknown incoming identity: create/match Master Lead using governed identity flow.
- Duplicate/replayed webhook: ignore idempotently and audit as needed.
- Destination contact unavailable: do not transition; continue/hold current channel.
- Project ENDED or Organization external-action disabled: accept/log inbound where policy allows but do not send outbound.

#### Product Rules

- Super Admin configures Organization Meta/messaging APIs.
- Organization users do not see secrets.
- Entitlement and readiness are distinct.
- Fallback requires approved, configured, healthy, contactable, eligible destination.
- No blind repeated provider retries.

#### Responsibility Split


**AI responsibilities**

- Sales Agent handles content.
- Lead Strategist recommends transition/next action.
- Agents do not access raw secrets.


**Deterministic application responsibilities**

- Credential vault/reference, connection tests, capability state, provider adapter, webhook verification, idempotency, contact/eligibility, failure actions.


**Human responsibilities**

- Super Admin configures/repairs connections.
- Organization users see status/action; permitted users manage Project channel behavior.

#### Permissions

- platform integration.manage; organization integration.configure by Super Admin; channel settings/manage; conversation permissions.

#### Data Requirements

- ProviderConnection
- OrganizationCapability
- CredentialReference
- WebhookEvent
- ChannelIdentity
- ConnectionTest
- IntegrationError

#### UX States

- Not configured
- Configuring
- Connected
- Partial
- Error
- Disabled
- Reauthorization required
- Provider outage

#### Notifications and Action Center

- In-app/email to relevant users for failure, reauth, pause, or restored status.
- Super Admin operations alert for broad provider issue.

#### Audit Requirements

- Credential config/rotation
- Connection tests
- Capability enable/disable
- Provider event verification
- Fallback/pause

#### Analytics Events and Metrics

- Connection success
- Send/receive failure
- Webhook duplicates
- Fallback use
- Downtime

#### Security and Privacy

- Secrets never logged/displayed
- Webhook signatures/replay
- Tenant resolution before processing
- No cross-org provider IDs
- Kill switch

#### B2B / B2C Behavior

- Same channel infrastructure; Strategy and qualification determine B2B/B2C messaging behavior.

#### Formal Product Requirements

| PR ID | Requirement/behavior | Type | BRD trace |
| --- | --- | --- | --- |
| PR-10-001 | Super Admin configures Organization Meta/messaging APIs. | Product Rule | BR-ACQ-016—019, BR-EML-009—012, BR-ADM-008—011, BR-SEC-007—009 |
| PR-10-002 | Organization users do not see secrets. | Product Rule | BR-ACQ-016—019, BR-EML-009—012, BR-ADM-008—011, BR-SEC-007—009 |
| PR-10-003 | Entitlement and readiness are distinct. | Product Rule | BR-ACQ-016—019, BR-EML-009—012, BR-ADM-008—011, BR-SEC-007—009 |
| PR-10-004 | Fallback requires approved, configured, healthy, contactable, eligible destination. | Product Rule | BR-ACQ-016—019, BR-EML-009—012, BR-ADM-008—011, BR-SEC-007—009 |
| PR-10-005 | No blind repeated provider retries. | Product Rule | BR-ACQ-016—019, BR-EML-009—012, BR-ADM-008—011, BR-SEC-007—009 |
| PR-10-006 | Organization sees capability as included-but-not-configured when appropriate. | Acceptance Behavior | BR-ACQ-016—019, BR-EML-009—012, BR-ADM-008—011, BR-SEC-007—009 |
| PR-10-007 | Failed connection blocks only affected execution and creates action. | Acceptance Behavior | BR-ACQ-016—019, BR-EML-009—012, BR-ADM-008—011, BR-SEC-007—009 |
| PR-10-008 | Verified inbound event maps to the correct Project Lead. | Acceptance Behavior | BR-ACQ-016—019, BR-EML-009—012, BR-ADM-008—011, BR-SEC-007—009 |
| PR-10-009 | Fallback never uses an unapproved/unavailable contact. | Acceptance Behavior | BR-ACQ-016—019, BR-EML-009—012, BR-ADM-008—011, BR-SEC-007—009 |
| PR-10-010 | Secrets are inaccessible to Organization users and agents. | Acceptance Behavior | BR-ACQ-016—019, BR-EML-009—012, BR-ADM-008—011, BR-SEC-007—009 |

#### Acceptance Criteria

- [ ] **AC-M10-001:** Organization sees capability as included-but-not-configured when appropriate.
- [ ] **AC-M10-002:** Failed connection blocks only affected execution and creates action.
- [ ] **AC-M10-003:** Verified inbound event maps to the correct Project Lead.
- [ ] **AC-M10-004:** Fallback never uses an unapproved/unavailable contact.
- [ ] **AC-M10-005:** Secrets are inaccessible to Organization users and agents.

### M11 — Generic Business Email Channel

**Purpose:** Support any compatible business mailbox for first outreach, threaded inbound replies, active AI sales, attachments, follow-up control, and delivery failure handling.

**Mapped BRD requirements:** BR-EML-001—012, BR-SAL-008—016, BR-SEC-006—008

#### Actors

- Authorized Organization user
- Sales Agent
- Lead Strategist
- Email Service/Adapter
- System

#### Representative User Stories


- **US-M11-001:** As a Authorized Organization user, I want this capability to support any compatible business mailbox for first outreach, threaded inbound replies, active AI sales, attachments, follow-up control, and delivery failure handling.

- **US-M11-002:** As the platform, I need all protected actions in this module to be permissioned, auditable, tenant-scoped, and state-valid.

- **US-M11-003:** As an authorized reviewer, I want clear loading, empty, failure, and recovery behavior so that work never depends on hidden system state.

#### Preconditions

- Email feature entitled.
- User has email integration permission.
- A compatible outbound and inbound mailbox connection can be established.
- Project LIVE and email approved.

#### Inputs

- Outbound server/provider configuration
- Inbound mailbox retrieval configuration
- Mailbox address
- Sales AI display name/signature
- Lead email
- Thread/message data
- Approved product attachments

#### Outputs

- Connection status
- Outbound messages
- Inbound thread events
- Delivery status/bounces
- Conversation timeline
- Action Items

#### Main Flow

1. Organization user configures mailbox and approved Sales AI identity.
2. System tests outbound and inbound capabilities without exposing credentials.
3. Selected Web Project Lead receives personalized first email automatically after all checks.
4. System records Message-ID/threading/provider references.
5. Inbound reply is retrieved/pushed, deduplicated, mapped to the conversation, and recorded.
6. Sales Agent/Lead Strategist handles active email conversation autonomously.
7. Sales AI may attach approved relevant product material.
8. When inactive, Lead Strategist proposes follow-up; user approves unless delegated.
9. Delivery bounce/failure updates channel quality, blocks blind retries, and creates an action/notification.

#### Alternative, Exception, and Failure Flows

- Outbound works but inbound does not: connection is partial/not ready for autonomous threaded sales; action required.
- Wrong credentials/SSL/config: show test details safely; do not store plaintext secrets in logs.
- Unknown reply/thread: attempt safe correlation; otherwise hold for review.
- Auto-reply/out-of-office: classify as non-material unless it contains actionable timing/contact information.
- Attachment too large/unsupported: send alternative approved content/link or request human review.
- Mailbox disconnected during active conversation: pause/fallback under channel rules.

#### Product Rules

- SMTP alone is not sufficient for the product requirement; inbound retrieval/push capability is required.
- AI identity is configured by the Organization and cannot be fabricated.
- First eligible outreach is automatic; normal inactive follow-up is approval-controlled.
- Bounces are channel evidence, not automatic lead disqualification.

#### Responsibility Split


**AI responsibilities**

- Generate natural email content using current Lead Strategy.
- Classify reply evidence.
- Select approved attachments.
- Never invent sender identity or product facts.


**Deterministic application responsibilities**

- Mailbox connection/test, secure credentials, threading, send/retrieve, delivery events, idempotency, eligibility/suppression, audit.


**Human responsibilities**

- Configure identity/mailbox.
- Monitor conversations.
- Approve/delegate follow-ups.
- Take over when needed.

#### Permissions

- integration.email.manage; lead/contact eligibility; conversation.view/takeover; followup.approve/delegate.

#### Data Requirements

- EmailConnection
- SalesIdentity
- EmailMessage
- EmailThread
- DeliveryEvent
- AttachmentReference
- ConversationChannelState

#### UX States

- Not configured
- Testing
- Connected
- Partial
- Error
- Sending
- Delivered
- Bounced
- Reply received
- Reauth required

#### Notifications and Action Center

- Connection failure
- Bounce
- Follow-up approval
- Handover
- Urgent thread failure; in-app/email without private content.

#### Audit Requirements

- Connection/identity changes
- Send/receive
- Delivery status
- Thread match
- Attachment use
- Follow-up approval/control

#### Analytics Events and Metrics

- Send/delivery/bounce/reply rates
- Time to response
- Qualification/conversion by sender/domain/campaign where appropriate

#### Security and Privacy

- Credential encryption
- Sender spoofing prevention
- Suppression before every send
- Inbound content untrusted
- Attachment safety
- No sensitive log payload

#### B2B / B2C Behavior

- B2B Web acquisition is a primary use. B2C email remains available when Strategy and eligibility support it.

#### Formal Product Requirements

| PR ID | Requirement/behavior | Type | BRD trace |
| --- | --- | --- | --- |
| PR-11-001 | SMTP alone is not sufficient for the product requirement; inbound retrieval/push capability is required. | Product Rule | BR-EML-001—012, BR-SAL-008—016, BR-SEC-006—008 |
| PR-11-002 | AI identity is configured by the Organization and cannot be fabricated. | Product Rule | BR-EML-001—012, BR-SAL-008—016, BR-SEC-006—008 |
| PR-11-003 | First eligible outreach is automatic; normal inactive follow-up is approval-controlled. | Product Rule | BR-EML-001—012, BR-SAL-008—016, BR-SEC-006—008 |
| PR-11-004 | Bounces are channel evidence, not automatic lead disqualification. | Product Rule | BR-EML-001—012, BR-SAL-008—016, BR-SEC-006—008 |
| PR-11-005 | A compatible mailbox can send and ingest a threaded reply. | Acceptance Behavior | BR-EML-001—012, BR-SAL-008—016, BR-SEC-006—008 |
| PR-11-006 | Reply appears in the same Project Lead conversation. | Acceptance Behavior | BR-EML-001—012, BR-SAL-008—016, BR-SEC-006—008 |
| PR-11-007 | AI responds automatically while active and waits for approval for normal follow-up. | Acceptance Behavior | BR-EML-001—012, BR-SAL-008—016, BR-SEC-006—008 |
| PR-11-008 | Bounce blocks additional automatic sends to that address and creates action. | Acceptance Behavior | BR-EML-001—012, BR-SAL-008—016, BR-SEC-006—008 |
| PR-11-009 | Configured sender identity appears consistently. | Acceptance Behavior | BR-EML-001—012, BR-SAL-008—016, BR-SEC-006—008 |

#### Acceptance Criteria

- [ ] **AC-M11-001:** A compatible mailbox can send and ingest a threaded reply.
- [ ] **AC-M11-002:** Reply appears in the same Project Lead conversation.
- [ ] **AC-M11-003:** AI responds automatically while active and waits for approval for normal follow-up.
- [ ] **AC-M11-004:** Bounce blocks additional automatic sends to that address and creates action.
- [ ] **AC-M11-005:** Configured sender identity appears consistently.

### M12 — Business Pool, Master Lead Pool, and Project Lead Context

**Purpose:** Maintain governed reusable identities for companies and people while isolating every Organization's Project-specific sales intelligence.

**Mapped BRD requirements:** BR-POOL-001—015, BR-SEC-016—018

#### Actors

- System
- Lead Discovery Agent
- Lead Enrichment Agent
- Platform data steward/Super Admin
- Authorized Project user

#### Representative User Stories


- **US-M12-001:** As a System, I want this capability to maintain governed reusable identities for companies and people while isolating every Organization's Project-specific sales intelligence.

- **US-M12-002:** As the platform, I need all protected actions in this module to be permissioned, auditable, tenant-scoped, and state-valid.

- **US-M12-003:** As an authorized reviewer, I want clear loading, empty, failure, and recovery behavior so that work never depends on hidden system state.

#### Preconditions

- A company/person is discovered, registered, imported in future, or interacts through a channel.

#### Inputs

- Identity/contact evidence
- Business registration/domain/name evidence
- Source restrictions
- Interest/professional signals
- Project association

#### Outputs

- Business Pool entity
- Master Lead
- Person-business relationships
- Project Lead context
- Merge recommendation/history
- Reuse/privacy flags

#### Main Flow

1. Normalize incoming company/person identifiers.
2. Search Business/Master Pools for potential matches.
3. Compute deterministic and AI-supported match confidence.
4. High-confidence valid match updates existing identity and provenance; medium-confidence remains pending/separate; low-confidence creates new identity.
5. Create/update person-business relationships with role, dates, source, and confidence.
6. Create a separate Project Lead context when an Organization Project uses a Master Lead.
7. Persist reusable evidence separately from private Project-specific sales data.
8. Authorized Platform user may review/reverse important merges.

#### Alternative, Exception, and Failure Flows

- Conflicting employers/locations/contact methods: retain all evidence/currentness; do not blindly overwrite.
- Same person with different email/phone: merge only when evidence threshold met.
- Company domain shared by group: preserve distinct entities and relationships.
- Registered tenant also appears as prospect: link real-world Business Pool identity without exposing tenant account data.
- Deletion/correction request: mark and process under future detailed retention policy; do not silently discard audit/provenance.

#### Product Rules

- Business Pool contains companies; Master Lead contains people.
- A Master Lead may serve multiple unrelated Projects.
- No universal score/Strategy/state on Master Lead.
- CR+jurisdiction strongest company identifier where available.
- Names alone never auto-merge people.
- All reusable classifications require evidence and context.

#### Responsibility Split


**AI responsibilities**

- Enrichment/Discovery may propose attributes and matches.
- Agents never directly merge or expose private tenant information.


**Deterministic application responsibilities**

- Normalize/match/merge/reverse, enforce reuse policy, maintain provenance/conflicts, create Project Lead, tenant isolation.


**Human responsibilities**

- Authorized Platform user resolves ambiguous/high-risk merges; Project users see only permitted Project-relevant profile.

#### Permissions

- Platform identity.merge/reverse/review; Project lead.view; no cross-tenant raw pool browsing unless explicitly permitted by product flow.

#### Data Requirements

- BusinessPoolEntity
- MasterLead
- PersonBusinessRelationship
- ReusableSignal
- SourceEvidence
- ProjectLead
- MergeDecision
- CorrectionRequest
- SuppressionRecord

#### UX States

- New
- Possible match
- Merged
- Conflict
- Correction pending
- Suppressed/restricted
- Stale data warning

#### Notifications and Action Center

- Platform action for ambiguous merge/correction; Project action only when contact/identity blocks work.

#### Audit Requirements

- Create/update/match/merge/reverse
- Reusable/private classification
- Relationship change
- Suppression/correction

#### Analytics Events and Metrics

- Duplicate/merge rate
- False-merge corrections
- Profile freshness
- Reusable match contribution
- Source quality

#### Security and Privacy

- Cross-tenant privacy boundary
- Source licensing/reuse flags
- Sensitive attribute minimization
- Reversible merge
- Suppression scope

#### B2B / B2C Behavior

- B2B uses Business Pool + Master Lead relationship; B2C may use Master Lead without company.

#### Formal Product Requirements

| PR ID | Requirement/behavior | Type | BRD trace |
| --- | --- | --- | --- |
| PR-12-001 | Business Pool contains companies; Master Lead contains people. | Product Rule | BR-POOL-001—015, BR-SEC-016—018 |
| PR-12-002 | A Master Lead may serve multiple unrelated Projects. | Product Rule | BR-POOL-001—015, BR-SEC-016—018 |
| PR-12-003 | No universal score/Strategy/state on Master Lead. | Product Rule | BR-POOL-001—015, BR-SEC-016—018 |
| PR-12-004 | CR+jurisdiction strongest company identifier where available. | Product Rule | BR-POOL-001—015, BR-SEC-016—018 |
| PR-12-005 | Names alone never auto-merge people. | Product Rule | BR-POOL-001—015, BR-SEC-016—018 |
| PR-12-006 | All reusable classifications require evidence and context. | Product Rule | BR-POOL-001—015, BR-SEC-016—018 |
| PR-12-007 | Same person used by two Projects has one Master Lead and two Project Leads. | Acceptance Behavior | BR-POOL-001—015, BR-SEC-016—018 |
| PR-12-008 | Project A private conversation is unavailable to Project B/Organization B. | Acceptance Behavior | BR-POOL-001—015, BR-SEC-016—018 |
| PR-12-009 | Merge preserves both sources and can be reversed. | Acceptance Behavior | BR-POOL-001—015, BR-SEC-016—018 |
| PR-12-010 | Tenant Organization linkage does not reveal private account activity. | Acceptance Behavior | BR-POOL-001—015, BR-SEC-016—018 |
| PR-12-011 | Contextual Not Interested signal is not generalized globally. | Acceptance Behavior | BR-POOL-001—015, BR-SEC-016—018 |

#### Acceptance Criteria

- [ ] **AC-M12-001:** Same person used by two Projects has one Master Lead and two Project Leads.
- [ ] **AC-M12-002:** Project A private conversation is unavailable to Project B/Organization B.
- [ ] **AC-M12-003:** Merge preserves both sources and can be reversed.
- [ ] **AC-M12-004:** Tenant Organization linkage does not reveal private account activity.
- [ ] **AC-M12-005:** Contextual Not Interested signal is not generalized globally.

### M13 — Lead Enrichment and Identity Resolution

**Purpose:** Gather enough reliable, permitted evidence to support contactability, scoring, and dynamic sales Strategy without unlimited profiling.

**Mapped BRD requirements:** BR-QLF-001—005, BR-POOL-007—010, BR-SEC-016

#### Actors

- Lead Enrichment Agent
- Lead Strategist Agent
- System
- Authorized Platform reviewer

#### Representative User Stories


- **US-M13-001:** As a Lead Enrichment Agent, I want this capability to gather enough reliable, permitted evidence to support contactability, scoring, and dynamic sales Strategy without unlimited profiling.

- **US-M13-002:** As the platform, I need all protected actions in this module to be permissioned, auditable, tenant-scoped, and state-valid.

- **US-M13-003:** As an authorized reviewer, I want clear loading, empty, failure, and recovery behavior so that work never depends on hidden system state.

#### Preconditions

- Master Lead exists or a contactable candidate is being created.
- Approved sources/tools are available.

#### Inputs

- Lead identity/provenance
- Existing reusable profile
- Project ICP/product/scoring needs
- Targeted research request

#### Outputs

- Reusable Master Lead facts
- Project-specific fit evidence
- Contact verification
- Business/person relationships
- Missing/conflicting data
- Enough-for-scoring indicator

#### Main Flow

1. System defines the enrichment objective from Project Strategy and current gaps.
2. Agent reviews existing evidence and identifies only necessary research.
3. Agent calls approved tools/sources.
4. Agent returns observed/inferred facts with evidence/confidence/freshness and reuse classification.
5. System validates and persists reusable and Project-specific results separately.
6. Identity service evaluates merge recommendations.
7. System signals Lead Strategist when enough evidence exists or important gaps remain.
8. Lead Strategist may request later targeted re-enrichment.

#### Alternative, Exception, and Failure Flows

- No reliable source: keep Unknown and allow conversational qualification if appropriate.
- Conflicting source: preserve conflict and currentness evidence.
- No valid contact after candidate research: remain Research Candidate.
- Source restricted from retention/reuse: store only permitted references/derived Project evidence.
- Loop/cost limit reached: return partial result + Needs Review/Information.

#### Product Rules

- Purpose-limited, not exhaustive background research.
- Observed/inferred distinction is mandatory.
- Private conversation-derived data is not reusable Master Lead enrichment.
- AI recommends match; deterministic service merges.

#### Responsibility Split


**AI responsibilities**

- Research/interpret evidence and confidence.
- Return structured results and missing information.


**Deterministic application responsibilities**

- Source policy, tool authorization, persistence, match/merge, reuse/private routing, loop/cost limits.


**Human responsibilities**

- Review only exceptions/ambiguity as permitted.

#### Permissions

- enrichment.run/view; platform match.review; Project context access.

#### Data Requirements

- EnrichmentRun
- EnrichmentFact
- SourceEvidence
- ContactMethod
- ProjectFitEvidence
- IdentityMatchRecommendation

#### UX States

- Queued
- Running
- Partial
- Complete
- Needs information
- Needs human review
- Failed
- Source blocked

#### Notifications and Action Center

- Action for unresolved identity/contact or failed enrichment when it blocks acquisition.
- No routine email.

#### Audit Requirements

- Run/tool/source
- Facts persisted/rejected
- Match recommendation
- Cost/stop reason

#### Analytics Events and Metrics

- Enrichment success
- Contact verification
- Evidence freshness
- Cost/latency
- Downstream scoring impact

#### Security and Privacy

- Approved sources
- No unrestricted person profiling
- No private cross-tenant context
- Sensitive content controls

#### B2B / B2C Behavior

- B2B enrichment may research company/role/trigger; B2C uses lead-provided/campaign/permitted public data and stricter minimization.

#### Formal Product Requirements

| PR ID | Requirement/behavior | Type | BRD trace |
| --- | --- | --- | --- |
| PR-13-001 | Purpose-limited, not exhaustive background research. | Product Rule | BR-QLF-001—005, BR-POOL-007—010, BR-SEC-016 |
| PR-13-002 | Observed/inferred distinction is mandatory. | Product Rule | BR-QLF-001—005, BR-POOL-007—010, BR-SEC-016 |
| PR-13-003 | Private conversation-derived data is not reusable Master Lead enrichment. | Product Rule | BR-QLF-001—005, BR-POOL-007—010, BR-SEC-016 |
| PR-13-004 | AI recommends match; deterministic service merges. | Product Rule | BR-QLF-001—005, BR-POOL-007—010, BR-SEC-016 |
| PR-13-005 | Every persisted fact has source/confidence/freshness/type. | Acceptance Behavior | BR-QLF-001—005, BR-POOL-007—010, BR-SEC-016 |
| PR-13-006 | Unknown data is not fabricated. | Acceptance Behavior | BR-QLF-001—005, BR-POOL-007—010, BR-SEC-016 |
| PR-13-007 | Reusable and Project-specific facts are separated. | Acceptance Behavior | BR-QLF-001—005, BR-POOL-007—010, BR-SEC-016 |
| PR-13-008 | Targeted re-enrichment returns to Lead Strategist. | Acceptance Behavior | BR-QLF-001—005, BR-POOL-007—010, BR-SEC-016 |
| PR-13-009 | Agent cannot merge identities directly. | Acceptance Behavior | BR-QLF-001—005, BR-POOL-007—010, BR-SEC-016 |

#### Acceptance Criteria

- [ ] **AC-M13-001:** Every persisted fact has source/confidence/freshness/type.
- [ ] **AC-M13-002:** Unknown data is not fabricated.
- [ ] **AC-M13-003:** Reusable and Project-specific facts are separated.
- [ ] **AC-M13-004:** Targeted re-enrichment returns to Lead Strategist.
- [ ] **AC-M13-005:** Agent cannot merge identities directly.

### M14 — Initial Scoring, Dynamic Qualification, and Lead State

**Purpose:** Rank pre-contact candidates and continuously reassess lead fit/readiness from real evidence.

**Mapped BRD requirements:** BR-QLF-006—017, BR-SAL-002—003

#### Actors

- Strategy Agent
- Lead Strategist Agent
- System
- Authorized users

#### Representative User Stories


- **US-M14-001:** As a Strategy Agent, I want this capability to rank pre-contact candidates and continuously reassess lead fit/readiness from real evidence.

- **US-M14-002:** As the platform, I need all protected actions in this module to be permissioned, auditable, tenant-scoped, and state-valid.

- **US-M14-003:** As an authorized reviewer, I want clear loading, empty, failure, and recovery behavior so that work never depends on hidden system state.

#### Preconditions

- Approved Strategy qualification framework.
- Project Lead exists.
- For Web: enrichment sufficient for initial scoring; for Meta: source-intent signal exists.

#### Inputs

- Project-specific scoring dimensions/weights
- Lead evidence
- Source intent
- Conversation events
- Human instruction

#### Outputs

- Initial score
- Dynamic score/history
- Qualification outcome/state
- Reason codes
- Confidence/explanation
- Next Best Action trigger

#### Main Flow

1. Strategy Agent proposes qualification dimensions/weights; system validates platform bounds.
2. For Web, system/Lead Strategist calculates initial score from research evidence.
3. System ranks candidates and allocates up to limit.
4. For Meta, create initial assessment with explicit inbound intent.
5. On each material event, Lead Strategist reassesses fit, need/interest, ability to progress, score, state, reasons, and confidence.
6. System validates protected transitions and stores history.
7. Update Project analytics and the Lead Strategy/Next Best Action.

#### Alternative, Exception, and Failure Flows

- Insufficient evidence: maintain uncertainty and choose a natural information-gathering action.
- Strong initial score but explicit rejection: dynamic score/state changes sharply; initial remains historical.
- Low initial score but strong inbound buying evidence: dynamic score may rise sharply.
- No response: flag attention; do not automatically set Not Interested.
- Opt-out: deterministic suppression overrides score/qualification.

#### Product Rules

- Initial score is Project-specific, not universal.
- Qualified = fit + genuine need/interest + realistic path to progress.
- Outcomes are Project-specific except suppression scope.
- State describes operation; score describes strength.
- Every transition has explanation/evidence.

#### Responsibility Split


**AI responsibilities**

- Interpret evidence and recommend score/state/reason/confidence.
- Never directly change suppression.


**Deterministic application responsibilities**

- Validate formula/weights/transitions.
- Persist score/state history.
- Apply suppression/eligibility.
- Allocate limit.


**Human responsibilities**

- Review insight and instruct/take over; no direct protected field editing.

#### Permissions

- lead.view; strategy insight; state override only through defined human actions/outcomes.

#### Data Requirements

- QualificationFramework
- ScoreSnapshot
- QualificationEvidence
- LeadStateHistory
- ReasonCode
- SuppressionRecord
- NextBestAction

#### UX States

- Awaiting score
- Scored
- Qualifying
- Qualified
- Nurture
- Not Qualified
- Not Interested
- Suppressed
- Handover Ready

#### Notifications and Action Center

- Unresponsive/attention
- Qualified/handover
- Suppression/blocked communication
- Limit overflow where relevant

#### Audit Requirements

- Framework/version
- Score/state changes
- Evidence
- Manual outcome
- Suppression

#### Analytics Events and Metrics

- Score distribution
- State conversion
- False positives/negatives from outcomes
- Time to qualification

#### Security and Privacy

- No sensitive inference without policy
- Suppression deterministic
- No cross-Project score reuse
- Explainability

#### B2B / B2C Behavior

- B2B dimensions may include company/person influence. B2C dimensions cannot require B2B data.

#### Formal Product Requirements

| PR ID | Requirement/behavior | Type | BRD trace |
| --- | --- | --- | --- |
| PR-14-001 | Initial score is Project-specific, not universal. | Product Rule | BR-QLF-006—017, BR-SAL-002—003 |
| PR-14-002 | Qualified = fit + genuine need/interest + realistic path to progress. | Product Rule | BR-QLF-006—017, BR-SAL-002—003 |
| PR-14-003 | Outcomes are Project-specific except suppression scope. | Product Rule | BR-QLF-006—017, BR-SAL-002—003 |
| PR-14-004 | State describes operation; score describes strength. | Product Rule | BR-QLF-006—017, BR-SAL-002—003 |
| PR-14-005 | Every transition has explanation/evidence. | Product Rule | BR-QLF-006—017, BR-SAL-002—003 |
| PR-14-006 | Web candidate gets an initial score before first contact. | Acceptance Behavior | BR-QLF-006—017, BR-SAL-002—003 |
| PR-14-007 | Meta lead begins with intent evidence. | Acceptance Behavior | BR-QLF-006—017, BR-SAL-002—003 |
| PR-14-008 | Dynamic evidence can change score/state both directions. | Acceptance Behavior | BR-QLF-006—017, BR-SAL-002—003 |
| PR-14-009 | Suppressed lead cannot receive outbound communication. | Acceptance Behavior | BR-QLF-006—017, BR-SAL-002—003 |
| PR-14-010 | Every active lead has one current Next Best Action. | Acceptance Behavior | BR-QLF-006—017, BR-SAL-002—003 |

#### Acceptance Criteria

- [ ] **AC-M14-001:** Web candidate gets an initial score before first contact.
- [ ] **AC-M14-002:** Meta lead begins with intent evidence.
- [ ] **AC-M14-003:** Dynamic evidence can change score/state both directions.
- [ ] **AC-M14-004:** Suppressed lead cannot receive outbound communication.
- [ ] **AC-M14-005:** Every active lead has one current Next Best Action.

### M15 — Dynamic Lead Strategy and Next Best Action

**Purpose:** Maintain the live sales brain for each Project Lead and provide one explainable commercial next step.

**Mapped BRD requirements:** BR-SAL-001—004, BR-SAL-017, BR-AGT-017

#### Actors

- Lead Strategist Agent
- Sales Agent
- Authorized Organization user
- System

#### Representative User Stories


- **US-M15-001:** As a Lead Strategist Agent, I want this capability to maintain the live sales brain for each Project Lead and provide one explainable commercial next step.

- **US-M15-002:** As the platform, I need all protected actions in this module to be permissioned, auditable, tenant-scoped, and state-valid.

- **US-M15-003:** As an authorized reviewer, I want clear loading, empty, failure, and recovery behavior so that work never depends on hidden system state.

#### Preconditions

- Project Lead exists; approved Project Strategy/product context available; authorized evidence/memory assembled.

#### Inputs

- Project Strategy
- Product Knowledge
- Reusable permitted Master Lead profile
- Project Lead state/score
- Persistent memory
- Latest material event
- Human instruction
- Channel/control context

#### Outputs

- Updated lead understanding
- Score/state recommendation
- Sales hypothesis
- Current objective
- Current approach
- Objection analysis
- Tried approaches
- Missing information
- Next Best Action
- Follow-up/handover recommendation
- Memory updates
- User summary

#### Main Flow

1. Material event triggers Lead Strategist.
2. Agent identifies exactly what changed.
3. Agent updates needs, pains, gains, product interest, timing, influence/readiness, objections, and commitments.
4. Agent reassesses qualification and current sales hypothesis.
5. Agent chooses one current objective and approach.
6. Agent records successful/failed approaches to prevent repetition.
7. Agent selects one Next Best Action or Wait/No Action and determines execution/approval/handover requirement.
8. Agent proposes structured memory changes.
9. Application validates and persists current Lead Strategy/memory/state.
10. Sales Agent receives the authoritative current Strategy if communication is required.

#### Alternative, Exception, and Failure Flows

- Missing external evidence: request Lead Enrichment within loop/cost bounds.
- Conflicting evidence: lower confidence and ask/verify naturally.
- Human instruction conflicts with Project restriction/approved commercial terms: block instruction and explain/takeover option.
- No safe action: Needs Human Review.
- Trivial message: no full re-strategy unless material evidence threshold is met.

#### Product Rules

- One Lead Strategy per Project Lead.
- One current Next Best Action.
- Strategy is not a giant free-form document; structured living object.
- Human gives instructions through UI; does not edit protected fields.
- Lead Strategist never sends external messages.

#### Responsibility Split


**AI responsibilities**

- All strategic responsibilities described above.
- Request approved Enrichment specialist.
- Return structured result.


**Deterministic application responsibilities**

- Assemble context, enforce hierarchy/limits, persist result, validate state/action, trigger Sales Agent/Action Item.


**Human responsibilities**

- Inspect summary, Ask AI, give lead-specific instruction, approve follow-up, take over.

#### Permissions

- conversation.view; lead.strategy.view/instruct; followup approval; takeover.

#### Data Requirements

- LeadStrategyVersion/Snapshot
- SalesHypothesis
- Objection
- ApproachAttempt
- NextBestAction
- MemoryUpdate
- FollowUpProposal
- HandoverRecommendation

#### UX States

- Updating
- Current
- Low confidence
- Needs enrichment
- Needs human review
- Wait
- Handover recommended

#### Notifications and Action Center

- Action/notification for follow-up approval, unresponsive attention, handover, missing critical knowledge.

#### Audit Requirements

- Trigger/event
- Input versions
- Strategy changes
- Next action
- Human instruction
- Specialist calls

#### Analytics Events and Metrics

- Strategy change frequency
- Next-action outcomes
- Repeated approach rate
- Human override rate
- Handover quality

#### Security and Privacy

- Authorized context only
- No raw cross-tenant pool
- No direct send/state override
- Prompt/tool injection controls

#### B2B / B2C Behavior

- Project-specific logic generated by Strategy; no universal B2B assumptions.

#### Formal Product Requirements

| PR ID | Requirement/behavior | Type | BRD trace |
| --- | --- | --- | --- |
| PR-15-001 | One Lead Strategy per Project Lead. | Product Rule | BR-SAL-001—004, BR-SAL-017, BR-AGT-017 |
| PR-15-002 | One current Next Best Action. | Product Rule | BR-SAL-001—004, BR-SAL-017, BR-AGT-017 |
| PR-15-003 | Strategy is not a giant free-form document; structured living object. | Product Rule | BR-SAL-001—004, BR-SAL-017, BR-AGT-017 |
| PR-15-004 | Human gives instructions through UI; does not edit protected fields. | Product Rule | BR-SAL-001—004, BR-SAL-017, BR-AGT-017 |
| PR-15-005 | Lead Strategist never sends external messages. | Product Rule | BR-SAL-001—004, BR-SAL-017, BR-AGT-017 |
| PR-15-006 | Material event produces a new structured assessment with reason. | Acceptance Behavior | BR-SAL-001—004, BR-SAL-017, BR-AGT-017 |
| PR-15-007 | User sees concise sales insight. | Acceptance Behavior | BR-SAL-001—004, BR-SAL-017, BR-AGT-017 |
| PR-15-008 | Agent remembers approaches tried and does not repeat failed angle without reason. | Acceptance Behavior | BR-SAL-001—004, BR-SAL-017, BR-AGT-017 |
| PR-15-009 | Follow-up recommendation contains timing/channel/objective/angle/approval. | Acceptance Behavior | BR-SAL-001—004, BR-SAL-017, BR-AGT-017 |
| PR-15-010 | Unsafe/missing-knowledge case returns human review instead of fabrication. | Acceptance Behavior | BR-SAL-001—004, BR-SAL-017, BR-AGT-017 |

#### Acceptance Criteria

- [ ] **AC-M15-001:** Material event produces a new structured assessment with reason.
- [ ] **AC-M15-002:** User sees concise sales insight.
- [ ] **AC-M15-003:** Agent remembers approaches tried and does not repeat failed angle without reason.
- [ ] **AC-M15-004:** Follow-up recommendation contains timing/channel/objective/angle/approval.
- [ ] **AC-M15-005:** Unsafe/missing-knowledge case returns human review instead of fabrication.

### M16 — Sales Agent and Unified Conversation Workspace

**Purpose:** Conduct natural multi-channel sales communication while exposing clear AI insight and human control.

**Mapped BRD requirements:** BR-SAL-004—025, BR-EML-004—009, BR-UX-012

#### Actors

- Sales Agent
- Lead Strategist
- Authorized sales user
- System
- Lead

#### Representative User Stories


- **US-M16-001:** As a Sales Agent, I want this capability to conduct natural multi-channel sales communication while exposing clear AI insight and human control.

- **US-M16-002:** As the platform, I need all protected actions in this module to be permissioned, auditable, tenant-scoped, and state-valid.

- **US-M16-003:** As an authorized reviewer, I want clear loading, empty, failure, and recovery behavior so that work never depends on hidden system state.

#### Preconditions

- Project LIVE
- Project Lead active within limit
- Approved product/Strategy/channel settings
- Eligible healthy channel
- AI control permitted

#### Inputs

- Inbound message or outbound action
- Current Lead Strategy
- Project Strategy
- Product Knowledge/media
- Persistent memory
- Language/tone/restrictions
- Control state

#### Outputs

- Message/action
- Attachments
- Detected evidence
- Material-event flag
- Strategist re-run request
- Conversation updates
- Safety/handover flags

#### Main Flow

1. System records inbound event/message and resolves Project Lead.
2. Sales Agent interprets intent and identifies material evidence.
3. If material, Lead Strategist reassesses before response where needed.
4. Sales Agent creates a natural response to the current objective/approach.
5. System validates grounded references, restrictions, control mode, permission, eligibility, suppression, and channel readiness.
6. Provider service sends or blocks the message.
7. Conversation timeline updates with actor/channel/version.
8. User can monitor, Ask AI, Instruct AI, or Take Over.
9. Sales Agent may retrieve/send relevant approved media/documents.

#### Alternative, Exception, and Failure Flows

- Knowledge missing: acknowledge uncertainty, ask clarification or request human review; never invent.
- Restricted word/claim: block or rewrite safely; if unresolved, Needs Review.
- Human Controlled: AI may monitor/summarize but cannot send.
- Project ENDED/Organization disabled: stop new outbound.
- Lead message outside business scope: politely redirect.
- Language unsupported by Project settings: follow defined fallback or human review.

#### Product Rules

- Natural and adaptive, not templated blast behavior.
- Active replies autonomous.
- Strategy and wording responsibilities remain separate.
- All messages and media are Project-scoped and attributed.
- Human can take over anytime.

#### Responsibility Split


**AI responsibilities**

- Generate communication, classify evidence, select approved content, flag material change/handover.
- No protected state change or provider credential access.


**Deterministic application responsibilities**

- Conversation persistence, context assembly, validation, send, audit, control state, UI updates.


**Human responsibilities**

- Monitor, instruct, take over, send manually in Human Controlled, return to AI.

#### Permissions

- conversation.view/send/takeover/return_to_ai; lead/project access; media access.

#### Data Requirements

- Conversation
- Message
- Attachment
- ConversationSummary
- SalesMemory
- ControlState
- ChannelThreadReference
- AgentRunReference

#### UX States

- Lead queue loading/empty
- Conversation loading
- AI controlled
- Human controlled
- Follow-up approval
- Handover ready
- Sending/failed
- Integration unavailable

#### Notifications and Action Center

- Handover/follow-up/unresponsive/failure notifications; no notification for routine active replies.

#### Audit Requirements

- Every message/send/block
- Media retrieval/send
- Control changes
- AI/human instruction
- Material evidence

#### Analytics Events and Metrics

- Response time
- Engagement
- Qualification progress
- Grounding failures
- Human takeover/return
- Channel transition

#### Security and Privacy

- Suppression before send
- Tenant/project context
- Untrusted inbound content
- No secrets
- Sensitive trace controls
- Manual-send authorization

#### B2B / B2C Behavior

- Language/tone/persona differ by Project. English and Arabic conversation support are baseline; exact full UI localization deferred.

#### Formal Product Requirements

| PR ID | Requirement/behavior | Type | BRD trace |
| --- | --- | --- | --- |
| PR-16-001 | Natural and adaptive, not templated blast behavior. | Product Rule | BR-SAL-004—025, BR-EML-004—009, BR-UX-012 |
| PR-16-002 | Active replies autonomous. | Product Rule | BR-SAL-004—025, BR-EML-004—009, BR-UX-012 |
| PR-16-003 | Strategy and wording responsibilities remain separate. | Product Rule | BR-SAL-004—025, BR-EML-004—009, BR-UX-012 |
| PR-16-004 | All messages and media are Project-scoped and attributed. | Product Rule | BR-SAL-004—025, BR-EML-004—009, BR-UX-012 |
| PR-16-005 | Human can take over anytime. | Product Rule | BR-SAL-004—025, BR-EML-004—009, BR-UX-012 |
| PR-16-006 | Lead queue, conversation, and AI insight appear in one workspace. | Acceptance Behavior | BR-SAL-004—025, BR-EML-004—009, BR-UX-012 |
| PR-16-007 | AI responds autonomously to active inbound when permitted. | Acceptance Behavior | BR-SAL-004—025, BR-EML-004—009, BR-UX-012 |
| PR-16-008 | Human takeover immediately prevents AI send. | Acceptance Behavior | BR-SAL-004—025, BR-EML-004—009, BR-UX-012 |
| PR-16-009 | Return to AI uses full latest context. | Acceptance Behavior | BR-SAL-004—025, BR-EML-004—009, BR-UX-012 |
| PR-16-010 | Unsupported product question is not fabricated. | Acceptance Behavior | BR-SAL-004—025, BR-EML-004—009, BR-UX-012 |

#### Acceptance Criteria

- [ ] **AC-M16-001:** Lead queue, conversation, and AI insight appear in one workspace.
- [ ] **AC-M16-002:** AI responds autonomously to active inbound when permitted.
- [ ] **AC-M16-003:** Human takeover immediately prevents AI send.
- [ ] **AC-M16-004:** Return to AI uses full latest context.
- [ ] **AC-M16-005:** Unsupported product question is not fabricated.

### M17 — Follow-Up, Unresponsive Attention, and Conversation Control

**Purpose:** Enable strategic re-engagement while preserving default human approval and optional explicit AI delegation.

**Mapped BRD requirements:** BR-SAL-011—017, BR-UX-003—008

#### Actors

- Lead Strategist
- Sales Agent
- Authorized user
- System

#### Representative User Stories


- **US-M17-001:** As a Lead Strategist, I want this capability to enable strategic re-engagement while preserving default human approval and optional explicit AI delegation.

- **US-M17-002:** As the platform, I need all protected actions in this module to be permissioned, auditable, tenant-scoped, and state-valid.

- **US-M17-003:** As an authorized reviewer, I want clear loading, empty, failure, and recovery behavior so that work never depends on hidden system state.

#### Preconditions

- Conversation is inactive or material no-response condition exists.
- Project/lead/channel remain eligible.

#### Inputs

- Conversation history
- Reason/objection
- Lead state
- Recommended timing/channel/objective
- Control/delegation state

#### Outputs

- Follow-Up Proposal
- Approval decision
- Scheduled send
- Delegation state
- Unresponsive Action Item
- Cancelled/obsolete proposal

#### Main Flow

1. Lead Strategist determines that a follow-up is appropriate and why.
2. Sales Agent drafts a natural channel-specific message.
3. System creates a Follow-Up Proposal and Action Item.
4. Authorized user approves now/scheduled, edits, rejects, asks AI to replan, or delegates follow-up for this Project Lead.
5. If approved, system schedules and revalidates immediately before send.
6. If delegated, AI may plan/execute future follow-ups within boundaries until revoked/outcome/suppression/takeover.
7. If lead replies before send, system reopens active conversation and cancels/obsoletes the proposal.
8. Meaningful unresponsiveness creates an alert and displays AI assessment/alternative approach.

#### Alternative, Exception, and Failure Flows

- Channel becomes unavailable: fallback only if approved; otherwise pause and action.
- Lead opted out: cancel all proposals and block.
- Project ENDED: cancel future sends.
- Proposal becomes stale due to product/Strategy/state change: require replan/approval.
- User edits message but violates restriction: block and explain.

#### Product Rules

- Default follow-up requires human approval.
- Delegation is explicit, per Project Lead, revocable, and auditable.
- One adaptive follow-up at a time.
- No universal fixed timing schedule.
- No Response is not final status.

#### Responsibility Split


**AI responsibilities**

- Recommend timing/strategy/message; execute only under delegated mode.
- Reassess after result.


**Deterministic application responsibilities**

- Create/authorize/schedule/revalidate/cancel proposals.
- Manage delegation and Action Items.
- Check suppression/state/readiness.


**Human responsibilities**

- Approve/edit/reject/replan/delegate/revoke/take over.

#### Permissions

- followup.approve/delegate; conversation.takeover; Project access.

#### Data Requirements

- FollowUpProposal
- FollowUpDecision
- ScheduledAction
- DelegationGrant
- UnresponsiveFlag
- ActionItem

#### UX States

- Awaiting approval
- Approved scheduled
- Sending
- Sent
- Rejected
- Cancelled
- Obsolete
- Delegated
- Paused

#### Notifications and Action Center

- In-app/email for approval and material unresponsiveness; no repeated spam notifications without state change.

#### Audit Requirements

- Proposal/reason
- Approval/edit/reject
- Delegation/revoke
- Pre-send validation
- Cancel/obsolete
- Send result

#### Analytics Events and Metrics

- Approval delay
- Approval rate
- Delegated follow-up outcomes
- Unresponsive recovery
- Cancelled stale sends

#### Security and Privacy

- Suppression/eligibility immediately before send
- Permission/Project access
- No unauthorized background send
- Audit

#### B2B / B2C Behavior

- Same control model; timing/wording/strategy adapt to channel and Project.

#### Formal Product Requirements

| PR ID | Requirement/behavior | Type | BRD trace |
| --- | --- | --- | --- |
| PR-17-001 | Default follow-up requires human approval. | Product Rule | BR-SAL-011—017, BR-UX-003—008 |
| PR-17-002 | Delegation is explicit, per Project Lead, revocable, and auditable. | Product Rule | BR-SAL-011—017, BR-UX-003—008 |
| PR-17-003 | One adaptive follow-up at a time. | Product Rule | BR-SAL-011—017, BR-UX-003—008 |
| PR-17-004 | No universal fixed timing schedule. | Product Rule | BR-SAL-011—017, BR-UX-003—008 |
| PR-17-005 | No Response is not final status. | Product Rule | BR-SAL-011—017, BR-UX-003—008 |
| PR-17-006 | No normal inactive follow-up sends without approval or active delegation. | Acceptance Behavior | BR-SAL-011—017, BR-UX-003—008 |
| PR-17-007 | Lead reply cancels obsolete pending send. | Acceptance Behavior | BR-SAL-011—017, BR-UX-003—008 |
| PR-17-008 | Delegation can be revoked and cannot bypass suppression/final action boundary. | Acceptance Behavior | BR-SAL-011—017, BR-UX-003—008 |
| PR-17-009 | Unresponsive flag appears as action, not forced status. | Acceptance Behavior | BR-SAL-011—017, BR-UX-003—008 |
| PR-17-010 | Scheduled send rechecks all mandatory conditions. | Acceptance Behavior | BR-SAL-011—017, BR-UX-003—008 |

#### Acceptance Criteria

- [ ] **AC-M17-001:** No normal inactive follow-up sends without approval or active delegation.
- [ ] **AC-M17-002:** Lead reply cancels obsolete pending send.
- [ ] **AC-M17-003:** Delegation can be revoked and cannot bypass suppression/final action boundary.
- [ ] **AC-M17-004:** Unresponsive flag appears as action, not forced status.
- [ ] **AC-M17-005:** Scheduled send rechecks all mandatory conditions.

### M18 — Human Handover, Opportunity, and Conversion

**Purpose:** Transfer a prepared lead to an authorized human for every final MVP action and preserve outcome/attribution.

**Mapped BRD requirements:** BR-CON-001—012, BR-ANL-005, BR-UX-006

#### Actors

- Lead Strategist
- Sales Agent
- Authorized Project user
- System

#### Representative User Stories


- **US-M18-001:** As a Lead Strategist, I want this capability to transfer a prepared lead to an authorized human for every final MVP action and preserve outcome/attribution.

- **US-M18-002:** As the platform, I need all protected actions in this module to be permissioned, auditable, tenant-scoped, and state-valid.

- **US-M18-003:** As an authorized reviewer, I want clear loading, empty, failure, and recovery behavior so that work never depends on hidden system state.

#### Preconditions

- Project Lead active; final/protected action or manual takeover condition exists.

#### Inputs

- Handover reason
- Lead state/score
- Current request
- Conversation summary
- Products/objections/commitments
- Source/campaign
- Recommended human action

#### Outputs

- Handover Action Item
- Assigned human owner
- Human Controlled conversation
- Optional Opportunity
- Human-recorded outcome
- Conversion event

#### Main Flow

1. Lead Strategist or user triggers Handover Ready.
2. System stops autonomous sales sending and creates a structured handover brief.
3. All authorized users with Project access see the claimable item.
4. First valid Take Over claim atomically assigns the human owner.
5. Human continues conversation and performs the final action externally/through future supported workflow.
6. Human records Meeting/Appointment/Quote/Opportunity/Sale/Nurture/Lost/Still Working or configured outcome.
7. If continuing deal tracking is needed, application creates optional Opportunity per rules.
8. On success, create Conversion event with source/Project/product/evidence/human attribution.
9. Update analytics and learning evidence.

#### Alternative, Exception, and Failure Flows

- No authorized user claims: repeat in-app/email alert/escalation according to notification policy.
- Two users claim simultaneously: one succeeds, one sees current owner.
- Human releases/reassignment: future detailed behavior may be simple Owner/Admin reassignment in MVP.
- Human does not record outcome: Action Item remains open and reminders apply.
- Lead asks AI after human takeover: AI remains silent unless human returns control.

#### Product Rules

- All final MVP actions are human.
- High interest alone does not require handover until final/protected boundary.
- Opportunity is optional.
- Conversion is an event, not a boolean.
- Master Lead is not globally Won/Lost.

#### Responsibility Split


**AI responsibilities**

- Prepare recommendation/brief and monitor if permitted.
- No final action.


**Deterministic application responsibilities**

- Atomic claim, control state, assignment, reminders, Opportunity/Conversion persistence, attribution, audit.


**Human responsibilities**

- Claim, communicate, complete final action, record outcome, optionally return to AI before final outcome if appropriate.

#### Permissions

- conversation.takeover; handover.claim/reassign; conversion.record; opportunity.manage.

#### Data Requirements

- Handover
- HandoverClaim
- HumanAssignment
- Opportunity
- Conversion
- OutcomeEvidence
- AttributionReference

#### UX States

- Handover ready
- Claimed/Human controlled
- Waiting human
- Still working
- Converted
- Nurture
- Lost
- Outcome overdue

#### Notifications and Action Center

- Immediate in-app/email handover alert; reminders if unclaimed/outcome missing; conversion confirmation.

#### Audit Requirements

- Trigger/brief
- Claim conflict
- Human messages
- Outcome
- Opportunity/Conversion
- Attribution

#### Analytics Events and Metrics

- Handover volume
- Claim time
- Human response time
- Handover-to-conversion
- Unclaimed/overdue outcomes

#### Security and Privacy

- Permission/Project access
- Atomic claim
- No AI send during Human Controlled
- Sensitive email minimization

#### B2B / B2C Behavior

- Final human action applies to both B2B and B2C; exact outcome type is Project-specific.

#### Formal Product Requirements

| PR ID | Requirement/behavior | Type | BRD trace |
| --- | --- | --- | --- |
| PR-18-001 | All final MVP actions are human. | Product Rule | BR-CON-001—012, BR-ANL-005, BR-UX-006 |
| PR-18-002 | High interest alone does not require handover until final/protected boundary. | Product Rule | BR-CON-001—012, BR-ANL-005, BR-UX-006 |
| PR-18-003 | Opportunity is optional. | Product Rule | BR-CON-001—012, BR-ANL-005, BR-UX-006 |
| PR-18-004 | Conversion is an event, not a boolean. | Product Rule | BR-CON-001—012, BR-ANL-005, BR-UX-006 |
| PR-18-005 | Master Lead is not globally Won/Lost. | Product Rule | BR-CON-001—012, BR-ANL-005, BR-UX-006 |
| PR-18-006 | All eligible users see handover; exactly one claim succeeds. | Acceptance Behavior | BR-CON-001—012, BR-ANL-005, BR-UX-006 |
| PR-18-007 | AI sends no sales messages after claim. | Acceptance Behavior | BR-CON-001—012, BR-ANL-005, BR-UX-006 |
| PR-18-008 | Human can record final outcome without large CRM form. | Acceptance Behavior | BR-CON-001—012, BR-ANL-005, BR-UX-006 |
| PR-18-009 | Conversion includes source/campaign and human evidence. | Acceptance Behavior | BR-CON-001—012, BR-ANL-005, BR-UX-006 |
| PR-18-010 | Analytics distinguishes AI preparation from human result. | Acceptance Behavior | BR-CON-001—012, BR-ANL-005, BR-UX-006 |

#### Acceptance Criteria

- [ ] **AC-M18-001:** All eligible users see handover; exactly one claim succeeds.
- [ ] **AC-M18-002:** AI sends no sales messages after claim.
- [ ] **AC-M18-003:** Human can record final outcome without large CRM form.
- [ ] **AC-M18-004:** Conversion includes source/campaign and human evidence.
- [ ] **AC-M18-005:** Analytics distinguishes AI preparation from human result.

### M19 — Project, Campaign, Sales Analytics, and Learning

**Purpose:** Measure the complete commercial funnel and turn repeated evidence into human-approved Project improvements.

**Mapped BRD requirements:** BR-ANL-001—014, BR-CON-011—012

#### Actors

- Authorized Organization users
- Super Admin
- Campaign Agent
- Lead Strategist
- System

#### Representative User Stories


- **US-M19-001:** As a Authorized Organization users, I want this capability to measure the complete commercial funnel and turn repeated evidence into human-approved Project improvements.

- **US-M19-002:** As the platform, I need all protected actions in this module to be permissioned, auditable, tenant-scoped, and state-valid.

- **US-M19-003:** As an authorized reviewer, I want clear loading, empty, failure, and recovery behavior so that work never depends on hidden system state.

#### Preconditions

- Events and outcomes are recorded with Project/source/version references.

#### Inputs

- Acquisition/discovery events
- Campaign metrics
- Project Lead state history
- Messages/handovers
- Conversion/outcomes
- Strategy/agent versions

#### Outputs

- Project funnel dashboard
- Source/campaign quality
- AI/human performance
- Objection/pattern insights
- Strategy improvement recommendations

#### Main Flow

1. Aggregate Project funnel from acquired through converted.
2. Calculate Web and Meta source performance with downstream qualification/handover/conversion.
3. Separate AI sales effectiveness from human handover response/outcome.
4. Identify repeated persona/product/objection/channel patterns with confidence/sample-size context.
5. Generate evidence-backed Strategy improvement recommendation.
6. Authorized user reviews/accepts/rejects/refines recommendation.
7. Accepted recommendation enters controlled Strategy version workflow.
8. Expose Super Admin aggregate/authorized operational results.

#### Alternative, Exception, and Failure Flows

- Insufficient sample: show observation with low confidence, not recommendation.
- Missing human outcome: flag incomplete attribution/Action Item.
- Provider metric delayed: mark freshness/partial data.
- Attribution ambiguous: retain best available lineage and uncertainty.
- Project ENDED: analytics remain available and continue receiving late outcome corrections.

#### Product Rules

- Message volume is not a success metric by itself.
- Lead-level adaptation is automatic; Project Strategy changes are not.
- One outcome never rewrites Strategy.
- Master Lead is not the primary reporting dimension.

#### Responsibility Split


**AI responsibilities**

- Campaign Agent interprets campaign quality.
- Lead Strategist adapts individuals.
- Analytics/Strategy Agent may generate recommendation.


**Deterministic application responsibilities**

- Event collection, attribution, aggregation, freshness, permissions, version references, recommendation workflow.


**Human responsibilities**

- Review performance and recommendations; record missing outcomes; approve new Strategy version.

#### Permissions

- analytics.view; strategy.recommendation.review; Super Admin aggregate/authorized sensitive access.

#### Data Requirements

- AnalyticsEvent
- FunnelSnapshot
- Attribution
- MetricSnapshot
- PatternEvidence
- StrategyRecommendation
- RecommendationDecision

#### UX States

- No data
- Partial/stale
- Healthy
- Limit/quality warning
- Recommendation available
- Outcome missing

#### Notifications and Action Center

- In-app/email for material recommendation, severe campaign issue, missing handover outcome, or limit warning.

#### Audit Requirements

- Metric calculation versions
- Recommendation/evidence
- Decision
- Outcome correction

#### Analytics Events and Metrics

- Acquired/contacted/engaged/qualified/handover/conversion
- Qualified Handover Rate
- Handover-to-Conversion
- Source quality
- AI/human timing

#### Security and Privacy

- Tenant/project filtering
- Sensitive conversation not exposed in aggregate without permission
- No automatic cross-tenant learning

#### B2B / B2C Behavior

- B2B/B2C funnels share stages but use Project-specific qualification evidence and conversion types.

#### Formal Product Requirements

| PR ID | Requirement/behavior | Type | BRD trace |
| --- | --- | --- | --- |
| PR-19-001 | Message volume is not a success metric by itself. | Product Rule | BR-ANL-001—014, BR-CON-011—012 |
| PR-19-002 | Lead-level adaptation is automatic; Project Strategy changes are not. | Product Rule | BR-ANL-001—014, BR-CON-011—012 |
| PR-19-003 | One outcome never rewrites Strategy. | Product Rule | BR-ANL-001—014, BR-CON-011—012 |
| PR-19-004 | Master Lead is not the primary reporting dimension. | Product Rule | BR-ANL-001—014, BR-CON-011—012 |
| PR-19-005 | Dashboard traces conversion to source and Project Lead. | Acceptance Behavior | BR-ANL-001—014, BR-CON-011—012 |
| PR-19-006 | Cheap but low-quality leads are distinguishable from high-quality sources. | Acceptance Behavior | BR-ANL-001—014, BR-CON-011—012 |
| PR-19-007 | Human delay is distinguishable from AI performance. | Acceptance Behavior | BR-ANL-001—014, BR-CON-011—012 |
| PR-19-008 | Recommendation states evidence and confidence. | Acceptance Behavior | BR-ANL-001—014, BR-CON-011—012 |
| PR-19-009 | Rejecting a recommendation leaves current Strategy unchanged. | Acceptance Behavior | BR-ANL-001—014, BR-CON-011—012 |

#### Acceptance Criteria

- [ ] **AC-M19-001:** Dashboard traces conversion to source and Project Lead.
- [ ] **AC-M19-002:** Cheap but low-quality leads are distinguishable from high-quality sources.
- [ ] **AC-M19-003:** Human delay is distinguishable from AI performance.
- [ ] **AC-M19-004:** Recommendation states evidence and confidence.
- [ ] **AC-M19-005:** Rejecting a recommendation leaves current Strategy unchanged.

### M20 — Plans, Usage, Limits, and Entitlements

**Purpose:** Enforce commercial packaging and usage consistently without allowing UI, agents, or Organizations to bypass controls.

**Mapped BRD requirements:** BR-PLN-001—014, BR-ADM-012, BR-SEC-004

#### Actors

- Super Admin
- Organization users
- System

#### Representative User Stories


- **US-M20-001:** As a Super Admin, I want this capability to enforce commercial packaging and usage consistently without allowing UI, agents, or Organizations to bypass controls.

- **US-M20-002:** As the platform, I need all protected actions in this module to be permissioned, auditable, tenant-scoped, and state-valid.

- **US-M20-003:** As an authorized reviewer, I want clear loading, empty, failure, and recovery behavior so that work never depends on hidden system state.

#### Preconditions

- At least one Plan definition exists.

#### Inputs

- Plan commercial fields
- Feature entitlements
- Numeric/Unlimited limits
- Organization subscription snapshot
- Overrides
- Usage events

#### Outputs

- Effective entitlements/limits
- Usage counters
- Limit warnings/blocks
- Read-only Organization Plan view

#### Main Flow

1. Super Admin creates/updates Plan definitions.
2. Super Admin assigns a Plan during registration approval or later change.
3. System stores subscription snapshot and applies explicit Organization overrides.
4. At each protected action, system calculates effective entitlement/readiness/limit.
5. Usage is recorded transactionally/idempotently.
6. Organization UI shows effective values and approaching/reached warnings.
7. Only Super Admin changes Plan or overrides.
8. Existing records continue according to each limit's defined behavior.

#### Alternative, Exception, and Failure Flows

- Unlimited: bypass commercial numeric block only; continue provider/cost/safety checks.
- Plan definition edited: existing subscription remains unchanged until explicit migration.
- Override expires: future allowance uses new effective value; existing records are not deleted.
- Lead limit reached: existing leads continue; overflow Master Leads are not autonomously worked.
- Project limit reached: block new Project only.
- Strategy revision limit reached: block new revision only.

#### Product Rules

- Plan, entitlement, limit, and usage are separate concepts.
- Readiness is separate from entitlement.
- All enforcement is server-side.
- Manual billing/subscription status in MVP.

#### Responsibility Split


**AI responsibilities**

- No agent may decide or override entitlement/usage.


**Deterministic application responsibilities**

- Effective-limit service, counters, transactional checks, snapshots, overrides, warnings/actions, audit.


**Human responsibilities**

- Super Admin assigns/changes; Organization views; user contacts platform team for change.

#### Permissions

- plan.manage/assign; override.manage; usage.view; no Organization Plan edit.

#### Data Requirements

- Plan
- PlanVersion
- SubscriptionSnapshot
- Entitlement
- LimitDefinition
- OrganizationOverride
- UsageRecord
- LimitEvent

#### UX States

- Plan list/edit
- Assigned Plan
- Usage normal/near/reached
- Unlimited
- Override active/expired
- Permission denied

#### Notifications and Action Center

- In-app/email near/reached thresholds; admin alert for anomalous usage.

#### Audit Requirements

- Plan/version changes
- Assignment/migration
- Override
- Usage block
- Unlimited config

#### Analytics Events and Metrics

- Plan adoption
- Usage utilization
- Limit blocks
- Override frequency
- Estimated provider/AI cost

#### Security and Privacy

- No client-trusted counters
- Idempotent usage
- Permission protection
- Unlimited does not bypass abuse/safety

#### B2B / B2C Behavior

- Same commercial model; Projects may use different acquisition features based on entitlement/readiness.

#### Formal Product Requirements

| PR ID | Requirement/behavior | Type | BRD trace |
| --- | --- | --- | --- |
| PR-20-001 | Plan, entitlement, limit, and usage are separate concepts. | Product Rule | BR-PLN-001—014, BR-ADM-012, BR-SEC-004 |
| PR-20-002 | Readiness is separate from entitlement. | Product Rule | BR-PLN-001—014, BR-ADM-012, BR-SEC-004 |
| PR-20-003 | All enforcement is server-side. | Product Rule | BR-PLN-001—014, BR-ADM-012, BR-SEC-004 |
| PR-20-004 | Manual billing/subscription status in MVP. | Product Rule | BR-PLN-001—014, BR-ADM-012, BR-SEC-004 |
| PR-20-005 | Numeric and Unlimited limits enforce correctly. | Acceptance Behavior | BR-PLN-001—014, BR-ADM-012, BR-SEC-004 |
| PR-20-006 | Organization cannot alter Plan through browser request. | Acceptance Behavior | BR-PLN-001—014, BR-ADM-012, BR-SEC-004 |
| PR-20-007 | Existing subscription does not silently change after Plan edit. | Acceptance Behavior | BR-PLN-001—014, BR-ADM-012, BR-SEC-004 |
| PR-20-008 | Overflow lead behavior matches D-272. | Acceptance Behavior | BR-PLN-001—014, BR-ADM-012, BR-SEC-004 |
| PR-20-009 | Warnings explain exact blocked/new versus continuing behavior. | Acceptance Behavior | BR-PLN-001—014, BR-ADM-012, BR-SEC-004 |

#### Acceptance Criteria

- [ ] **AC-M20-001:** Numeric and Unlimited limits enforce correctly.
- [ ] **AC-M20-002:** Organization cannot alter Plan through browser request.
- [ ] **AC-M20-003:** Existing subscription does not silently change after Plan edit.
- [ ] **AC-M20-004:** Overflow lead behavior matches D-272.
- [ ] **AC-M20-005:** Warnings explain exact blocked/new versus continuing behavior.

### M21 — Direct Permissions, Project Access, Action Center, Notifications, and Audit

**Purpose:** Make work and authority explicit for each Organization user and preserve authoritative records of actions.

**Mapped BRD requirements:** BR-ADM-003—007, BR-UX-003—010, BR-REG-009

#### Actors

- Organization Owner
- Authorized Organization users
- System
- Agents

#### Representative User Stories


- **US-M21-001:** As a Organization Owner, I want this capability to make work and authority explicit for each Organization user and preserve authoritative records of actions.

- **US-M21-002:** As the platform, I need all protected actions in this module to be permissioned, auditable, tenant-scoped, and state-valid.

- **US-M21-003:** As an authorized reviewer, I want clear loading, empty, failure, and recovery behavior so that work never depends on hidden system state.

#### Preconditions

- Organization active; Owner membership exists.

#### Inputs

- Membership
- Permission grants
- Project access
- Agent recommendations
- System events
- Notification preferences

#### Outputs

- Authorization decisions
- Action Items
- In-app/email notifications
- Tenant audit history

#### Main Flow

1. Owner invites or links a user to the Organization.
2. Owner grants explicit permissions and Project access; optional presets only preselect grants.
3. Every protected request evaluates permission and Project access.
4. Agents/system recommend work; application creates/deduplicates Action Items.
5. Action Center filters items by current user authority and deep-links to resolution.
6. Notification router sends in-app and configured email alerts.
7. Completing underlying work resolves the Action Item; dismissing notification does not.
8. Application creates immutable audit events for important actions.
9. User removal revokes access but preserves actor history.

#### Alternative, Exception, and Failure Flows

- Owner cannot be removed/demoted by a user without protected authority.
- Permission revoked while page open: next server action denied.
- User belongs to multiple Organizations: active Organization context is explicit.
- Action no longer relevant: application marks completed/obsolete.
- Email delivery failure: in-app remains authoritative; email failure logged.
- Sensitive event: email contains minimal details and authenticated link.

#### Product Rules

- Direct permissions + Project access are the authorization model.
- Organization Owner is protected.
- Action Center, not notification, is unresolved-work source of truth.
- Agents do not create authoritative audit records.

#### Responsibility Split


**AI responsibilities**

- Recommend actions and provide summaries; no permission or audit authority.


**Deterministic application responsibilities**

- Membership/permission/access enforcement, action lifecycle, routing, audit, preference/mandatory notification rules.


**Human responsibilities**

- Manage permissions; resolve actions; configure preferences; inspect audit.

#### Permissions

- team.manage; permission.manage; project_access.manage; per-action explicit permissions.

#### Data Requirements

- Membership
- PermissionGrant
- ProjectAccessGrant
- ActionItem
- Notification
- NotificationPreference
- AuditEvent

#### UX States

- Invitation pending
- Active/suspended/removed
- No actions
- Open action
- Notification unread/read
- Audit filters
- Permission denied

#### Notifications and Action Center

- In-app and email only for MVP; permission-aware and minimal-content.

#### Audit Requirements

- Invites/membership
- Permission/access changes
- Every sensitive business action
- Notification delivery
- Action resolution

#### Analytics Events and Metrics

- Open action age
- Approval latency
- Notification delivery/read
- Permission denial
- Audit coverage

#### Security and Privacy

- No privilege escalation
- Tenant-scoped audit
- Email privacy
- Owner protection
- Server-side checks

#### B2B / B2C Behavior

- No difference in authorization model; Project mode changes available work, not permission mechanics.

#### Formal Product Requirements

| PR ID | Requirement/behavior | Type | BRD trace |
| --- | --- | --- | --- |
| PR-21-001 | Direct permissions + Project access are the authorization model. | Product Rule | BR-ADM-003—007, BR-UX-003—010, BR-REG-009 |
| PR-21-002 | Organization Owner is protected. | Product Rule | BR-ADM-003—007, BR-UX-003—010, BR-REG-009 |
| PR-21-003 | Action Center, not notification, is unresolved-work source of truth. | Product Rule | BR-ADM-003—007, BR-UX-003—010, BR-REG-009 |
| PR-21-004 | Agents do not create authoritative audit records. | Product Rule | BR-ADM-003—007, BR-UX-003—010, BR-REG-009 |
| PR-21-005 | User sees only authorized Projects and actions. | Acceptance Behavior | BR-ADM-003—007, BR-UX-003—010, BR-REG-009 |
| PR-21-006 | Direct unauthorized API attempt is denied even if UI control hidden. | Acceptance Behavior | BR-ADM-003—007, BR-UX-003—010, BR-REG-009 |
| PR-21-007 | Action completion and notification dismissal behave independently. | Acceptance Behavior | BR-ADM-003—007, BR-UX-003—010, BR-REG-009 |
| PR-21-008 | Removed user's historical audit remains. | Acceptance Behavior | BR-ADM-003—007, BR-UX-003—010, BR-REG-009 |
| PR-21-009 | Email notification does not expose full conversation. | Acceptance Behavior | BR-ADM-003—007, BR-UX-003—010, BR-REG-009 |

#### Acceptance Criteria

- [ ] **AC-M21-001:** User sees only authorized Projects and actions.
- [ ] **AC-M21-002:** Direct unauthorized API attempt is denied even if UI control hidden.
- [ ] **AC-M21-003:** Action completion and notification dismissal behave independently.
- [ ] **AC-M21-004:** Removed user's historical audit remains.
- [ ] **AC-M21-005:** Email notification does not expose full conversation.

### M22 — Super Admin Control Panel and Platform Operations

**Purpose:** Operate the SaaS, customer lifecycle, integrations, agent runtime, observability, safety, and support from a separate protected control plane.

**Mapped BRD requirements:** BR-ADM-001—020, BR-AGT-009—015, BR-SEC-021, BR-ANL-013

#### Actors

- Platform Super Admin
- Platform Admin/Support with restricted grants
- System

#### Representative User Stories


- **US-M22-001:** As a Platform Super Admin, I want this capability to operate the SaaS, customer lifecycle, integrations, agent runtime, observability, safety, and support from a separate protected control plane.

- **US-M22-002:** As the platform, I need all protected actions in this module to be permissioned, auditable, tenant-scoped, and state-valid.

- **US-M22-003:** As an authorized reviewer, I want clear loading, empty, failure, and recovery behavior so that work never depends on hidden system state.

#### Preconditions

- Platform administrator authenticated with explicit platform permissions.

#### Inputs

- Registration/Organization/Plan/usage data
- Provider/credential references
- Agent/model/prompt configurations
- Agent telemetry
- Support requests
- System health

#### Outputs

- Managed Organizations/subscriptions
- Configured integrations
- Agent versions/status
- Operational dashboards
- Support sessions
- Audit/safety actions

#### Main Flow

1. Super Admin dashboard shows registrations, Organization status, usage/limits, integration errors, agent health, failed jobs, and safety alerts.
2. Registration Applications module reviews/activates/rejects applicants.
3. Organization Detail provides Overview, Plan/Subscription, Limits/Usage, Integrations, Users, Projects, Company Understanding, Support Access, Audit, and Administrative Actions.
4. Platform Integrations manages Google/Search, AI credentials, and shared providers.
5. Organization Integrations manages Meta Ads/WhatsApp/Instagram/Facebook credentials and capabilities for the selected Organization.
6. Agents module manages Agent Registry, models, credential references, prompt/agent versions, tools, allowed interactions, evaluation, activation, and rollback.
7. Agent Operations visualizes runs, tool calls, errors, latency, usage, status, and permitted traces.
8. Super Admin may suspend/reactivate Organization, change Plan, apply override, disable external actions, and process closure.
9. Support uses controlled audited access.

#### Alternative, Exception, and Failure Flows

- Sensitive trace access: require elevated permission/reason and audit.
- Provider secret update: select/store credential reference; never redisplay plaintext after save.
- Organization suspension: external actions stop; data retained.
- Global/provider incident: kill switch/feature disable prevents affected side effects.
- Agent version failure: rollback to previous approved version.
- Support session expiry: access ends automatically.

#### Product Rules

- Super Admin CP is distinct from Organization Admin.
- Platform Admin/Support does not automatically receive Super Admin powers.
- Platform-managed Organization integration resides inside Organization detail.
- Platform-wide APIs reside under Platform Integrations.
- End-to-end observability is subject to tenant/privacy controls.

#### Responsibility Split


**AI responsibilities**

- No special AI autonomy; agents provide telemetry/results, not platform authorization.


**Deterministic application responsibilities**

- Platform RBAC, credential vault, integration status, observability ingestion, support access, kill switches, audit.


**Human responsibilities**

- Configure/approve/operate within platform grants.

#### Permissions

- Platform permissions for registration, Organizations, Plans, providers, agents, traces, support, suspension, audit.

#### Data Requirements

- PlatformUser
- PlatformPermission
- OrganizationAdminView
- CredentialReference
- ProviderConfig
- AgentDefinition
- AgentVersion
- AgentRun
- SupportSession
- SystemAlert

#### UX States

- Healthy dashboard
- Pending registrations
- Integration error
- Agent degraded
- Sensitive access prompt
- Suspended org
- System failure

#### Notifications and Action Center

- Platform in-app/email for critical provider/agent/security/system events.

#### Audit Requirements

- Every configuration, activation, rollback, support access, trace access, suspension, override, credential rotation.

#### Analytics Events and Metrics

- Org activation/time
- Integration health
- Agent success/latency/cost
- Provider failures
- Support usage
- Safety blocks

#### Security and Privacy

- Strong platform auth
- Least privilege
- Secrets protection
- Sensitive trace controls
- Audit immutability
- No tenant context leakage

#### B2B / B2C Behavior

- Platform administration is mode-independent.

#### Formal Product Requirements

| PR ID | Requirement/behavior | Type | BRD trace |
| --- | --- | --- | --- |
| PR-22-001 | Super Admin CP is distinct from Organization Admin. | Product Rule | BR-ADM-001—020, BR-AGT-009—015, BR-SEC-021, BR-ANL-013 |
| PR-22-002 | Platform Admin/Support does not automatically receive Super Admin powers. | Product Rule | BR-ADM-001—020, BR-AGT-009—015, BR-SEC-021, BR-ANL-013 |
| PR-22-003 | Platform-managed Organization integration resides inside Organization detail. | Product Rule | BR-ADM-001—020, BR-AGT-009—015, BR-SEC-021, BR-ANL-013 |
| PR-22-004 | Platform-wide APIs reside under Platform Integrations. | Product Rule | BR-ADM-001—020, BR-AGT-009—015, BR-SEC-021, BR-ANL-013 |
| PR-22-005 | End-to-end observability is subject to tenant/privacy controls. | Product Rule | BR-ADM-001—020, BR-AGT-009—015, BR-SEC-021, BR-ANL-013 |
| PR-22-006 | Super Admin can configure Organization X Meta capabilities without exposing secrets to Organization X. | Acceptance Behavior | BR-ADM-001—020, BR-AGT-009—015, BR-SEC-021, BR-ANL-013 |
| PR-22-007 | Super Admin can choose model/credential/prompt version per agent. | Acceptance Behavior | BR-ADM-001—020, BR-AGT-009—015, BR-SEC-021, BR-ANL-013 |
| PR-22-008 | Agent run view shows exact version/tools/status and hides sensitive content by default. | Acceptance Behavior | BR-ADM-001—020, BR-AGT-009—015, BR-SEC-021, BR-ANL-013 |
| PR-22-009 | External-action kill switch blocks provider sends. | Acceptance Behavior | BR-ADM-001—020, BR-AGT-009—015, BR-SEC-021, BR-ANL-013 |
| PR-22-010 | Support access is recorded and expires/ends. | Acceptance Behavior | BR-ADM-001—020, BR-AGT-009—015, BR-SEC-021, BR-ANL-013 |

#### Acceptance Criteria

- [ ] **AC-M22-001:** Super Admin can configure Organization X Meta capabilities without exposing secrets to Organization X.
- [ ] **AC-M22-002:** Super Admin can choose model/credential/prompt version per agent.
- [ ] **AC-M22-003:** Agent run view shows exact version/tools/status and hides sensitive content by default.
- [ ] **AC-M22-004:** External-action kill switch blocks provider sends.
- [ ] **AC-M22-005:** Support access is recorded and expires/ends.

### M23 — Google ADK Agent Component, Agent Registry, and Evaluation

**Purpose:** Use Google ADK for configurable specialist agents while keeping product state and protected behavior in deterministic application services.

**Mapped BRD requirements:** BR-AGT-001—020, BR-TST-001—008

#### Actors

- Platform Super Admin
- Agent runtime/orchestrator
- Application services
- Customer Simulator
- Evaluator

#### Representative User Stories


- **US-M23-001:** As a Platform Super Admin, I want this capability to use Google ADK for configurable specialist agents while keeping product state and protected behavior in deterministic application services.

- **US-M23-002:** As the platform, I need all protected actions in this module to be permissioned, auditable, tenant-scoped, and state-valid.

- **US-M23-003:** As an authorized reviewer, I want clear loading, empty, failure, and recovery behavior so that work never depends on hidden system state.

#### Preconditions

- Approved Agent definitions and versions exist.
- Application tool contracts and context assembly exist.

#### Inputs

- Application event/user action
- Authorized structured context
- Agent/prompt/model version
- Allowed tools/specialists
- Execution limits

#### Outputs

- Structured Agent Result
- Agent Run/trace metadata
- Validated state/event
- Evaluation result
- Activation/rollback decision

#### Main Flow

1. Application event selects the approved active Agent version.
2. Context assembler provides only authorized tenant/Project/lead data.
3. ADK runs the specialist Agent with allowed tools/interactions and bounded limits.
4. Agent returns structured result and normalized outcome.
5. Application validates schema, permission, business rules, and state transition.
6. Application persists authoritative result and triggers next event/Agent if allowed.
7. Every run records exact versions/tools/timing/status.
8. Draft Agent versions run in Test Lab against versioned simulator scenarios.
9. Evaluator scores quality and hard failures; Super Admin approves/activates or rejects/rolls back.

#### Alternative, Exception, and Failure Flows

- Retryable model failure: application retries within configured policy.
- Needs Information: create relevant action/question; no fabricated result.
- Needs Human Review/Blocked: stop workflow and alert.
- Repeated specialist loop: terminate at deterministic limit.
- Critical evaluation hard failure: production activation blocked.
- Model/provider unavailable: use only approved fallback if configured; otherwise fail safely.

#### Product Rules

- ADK is Agent Component, not source of business truth.
- Agent interaction graph and tools are deny-by-default.
- Editable prompt cannot disable platform safety.
- History is version-pinned.
- Customer Simulator never causes live side effects.

#### Responsibility Split


**AI responsibilities**

- Execute specialist reasoning per Agent contracts.
- No direct protected state or raw provider credential access.


**Deterministic application responsibilities**

- Orchestrate, assemble context, authorize tools, enforce limits, validate/persist, audit, activate/rollback.


**Human responsibilities**

- Super Admin configures/tests/approves versions and inspects operations.

#### Permissions

- agent.manage; version.test/approve/activate/rollback; trace.view_sensitive; tool.manage; evaluation.run.

#### Data Requirements

- AgentDefinition
- AgentVersion
- PromptVersion
- ModelConfiguration
- ToolDefinition
- InteractionRule
- AgentRun
- ToolCall
- EvaluationScenario
- EvaluationResult

#### UX States

- Draft
- Testing
- Under review
- Approved
- Active
- Retired
- Run success/retry/blocked/fail
- Evaluation pass/fail

#### Notifications and Action Center

- Platform alerts for elevated failure/latency/cost, hard evaluation fail, or fallback use.

#### Audit Requirements

- Version/config changes
- Run/context refs
- Tools/specialists
- Validation decision
- Evaluation/activation/rollback

#### Analytics Events and Metrics

- Run count/success/latency/cost
- Tool failure
- Hard failure
- Version comparison
- Regression pass

#### Security and Privacy

- Context isolation
- Tool authorization
- Prompt injection controls
- Secrets outside Agent
- Sensitive logs
- Loop/cost limits

#### B2B / B2C Behavior

- Agents consume Project-specific B2B/B2C contexts; no hard-coded universal targeting/qualification in the framework.

#### Formal Product Requirements

| PR ID | Requirement/behavior | Type | BRD trace |
| --- | --- | --- | --- |
| PR-23-001 | ADK is Agent Component, not source of business truth. | Product Rule | BR-AGT-001—020, BR-TST-001—008 |
| PR-23-002 | Agent interaction graph and tools are deny-by-default. | Product Rule | BR-AGT-001—020, BR-TST-001—008 |
| PR-23-003 | Editable prompt cannot disable platform safety. | Product Rule | BR-AGT-001—020, BR-TST-001—008 |
| PR-23-004 | History is version-pinned. | Product Rule | BR-AGT-001—020, BR-TST-001—008 |
| PR-23-005 | Customer Simulator never causes live side effects. | Product Rule | BR-AGT-001—020, BR-TST-001—008 |
| PR-23-006 | Agent cannot retrieve unrelated tenant context. | Acceptance Behavior | BR-AGT-001—020, BR-TST-001—008 |
| PR-23-007 | Invalid structured output is rejected/retried/blocked, not persisted. | Acceptance Behavior | BR-AGT-001—020, BR-TST-001—008 |
| PR-23-008 | Changing active version affects future runs only. | Acceptance Behavior | BR-AGT-001—020, BR-TST-001—008 |
| PR-23-009 | Simulator sends no real provider action. | Acceptance Behavior | BR-AGT-001—020, BR-TST-001—008 |
| PR-23-010 | Critical violation prevents activation. | Acceptance Behavior | BR-AGT-001—020, BR-TST-001—008 |

#### Acceptance Criteria

- [ ] **AC-M23-001:** Agent cannot retrieve unrelated tenant context.
- [ ] **AC-M23-002:** Invalid structured output is rejected/retried/blocked, not persisted.
- [ ] **AC-M23-003:** Changing active version affects future runs only.
- [ ] **AC-M23-004:** Simulator sends no real provider action.
- [ ] **AC-M23-005:** Critical violation prevents activation.

### M24 — Security, Privacy, Compliance, and Trust Controls

**Purpose:** Make tenant safety, communication permission, provider integrity, and AI trust boundaries mandatory product behavior.

**Mapped BRD requirements:** BR-SEC-001—025, BR-REG-009, BR-EML-012

#### Actors

- All actors
- System
- Security/Privacy/Legal reviewers

#### Representative User Stories


- **US-M24-001:** As a All actors, I want this capability to make tenant safety, communication permission, provider integrity, and AI trust boundaries mandatory product behavior.

- **US-M24-002:** As the platform, I need all protected actions in this module to be permissioned, auditable, tenant-scoped, and state-valid.

- **US-M24-003:** As an authorized reviewer, I want clear loading, empty, failure, and recovery behavior so that work never depends on hidden system state.

#### Preconditions

- Security controls are designed before protected feature implementation.

#### Inputs

- Identity/session
- Resource/tenant IDs
- Permissions
- Suppression/eligibility
- Provider events
- Uploaded/external content
- Agent/tool requests
- Logs

#### Outputs

- Allow/deny decision
- Blocked action reason
- Security/audit event
- Correction/suppression state
- Provider/legal readiness status

#### Main Flow

1. Authenticate every protected request.
2. Resolve Organization/Project/resource ownership server-side.
3. Evaluate permission, entitlement, readiness, state, suppression, and action policy.
4. Validate inputs and reject privileged/mass-assigned fields.
5. Authorize Agent tools with scoped context.
6. Verify provider webhooks/mail events and replay/idempotency controls.
7. Treat documents/pages/messages as untrusted and isolate instructions from policy.
8. Protect/minimize secrets and sensitive logs.
9. Create audit/security actions for blocked/high-risk events.
10. Verify current official provider/legal requirements before live production release.

#### Alternative, Exception, and Failure Flows

- Policy not verified: mark integration/release blocked; do not assume compliance.
- Prompt injection tries to call tool/change instructions: deny and record.
- Cross-tenant ID access: deny without revealing existence.
- Sensitive trace requested: require elevated permission/reason/audit.
- Opt-out/suppression: cancel pending sends and block future applicable sends.
- Provider signature/replay failure: reject event.

#### Product Rules

- AI is untrusted.
- UI hiding is never authorization.
- Suppression and protected states are deterministic.
- Private conversation is not reusable Master Lead data.
- Production data is not casually used in test/local environments.

#### Responsibility Split


**AI responsibilities**

- May classify content/risk or recommend review, but cannot override controls.


**Deterministic application responsibilities**

- All listed controls, audits, data boundaries, provider verification state, incident/kill switch.


**Human responsibilities**

- Use permitted actions; reviewers approve policy/legal readiness.

#### Permissions

- Security/platform permissions; Organization permissions; support access controls.

#### Data Requirements

- SecurityEvent
- SuppressionRecord
- Consent/EligibilityRecord
- PolicyVerification
- CredentialReference
- WebhookEvent
- AuditEvent
- DataCorrectionRequest

#### UX States

- Access denied
- Suppressed
- Policy blocked
- Provider unverified
- Security review
- Incident/kill switch

#### Notifications and Action Center

- Security/platform email/in-app for high-risk event; tenant notifications only when appropriate/minimal.

#### Audit Requirements

- All allow/deny high-risk decisions
- Support/trace access
- Suppression
- Credential/provider/security configuration

#### Analytics Events and Metrics

- Blocked attacks
- Cross-tenant tests
- Suppression enforcement
- Webhook invalid/replay
- Sensitive trace access

#### Security and Privacy

- Core module itself.

#### B2B / B2C Behavior

- B2C profiling/outreach requires especially strict minimization; B2B does not remove privacy/communication obligations.

#### Formal Product Requirements

| PR ID | Requirement/behavior | Type | BRD trace |
| --- | --- | --- | --- |
| PR-24-001 | AI is untrusted. | Product Rule | BR-SEC-001—025, BR-REG-009, BR-EML-012 |
| PR-24-002 | UI hiding is never authorization. | Product Rule | BR-SEC-001—025, BR-REG-009, BR-EML-012 |
| PR-24-003 | Suppression and protected states are deterministic. | Product Rule | BR-SEC-001—025, BR-REG-009, BR-EML-012 |
| PR-24-004 | Private conversation is not reusable Master Lead data. | Product Rule | BR-SEC-001—025, BR-REG-009, BR-EML-012 |
| PR-24-005 | Production data is not casually used in test/local environments. | Product Rule | BR-SEC-001—025, BR-REG-009, BR-EML-012 |
| PR-24-006 | Cross-tenant resource request is denied server-side. | Acceptance Behavior | BR-SEC-001—025, BR-REG-009, BR-EML-012 |
| PR-24-007 | Suppressed lead cannot receive a message through any Agent path. | Acceptance Behavior | BR-SEC-001—025, BR-REG-009, BR-EML-012 |
| PR-24-008 | Agent cannot invoke unregistered tool or bypass provider service. | Acceptance Behavior | BR-SEC-001—025, BR-REG-009, BR-EML-012 |
| PR-24-009 | Invalid webhook is not processed. | Acceptance Behavior | BR-SEC-001—025, BR-REG-009, BR-EML-012 |
| PR-24-010 | Unverified legal/provider requirement remains visibly blocked. | Acceptance Behavior | BR-SEC-001—025, BR-REG-009, BR-EML-012 |

#### Acceptance Criteria

- [ ] **AC-M24-001:** Cross-tenant resource request is denied server-side.
- [ ] **AC-M24-002:** Suppressed lead cannot receive a message through any Agent path.
- [ ] **AC-M24-003:** Agent cannot invoke unregistered tool or bypass provider service.
- [ ] **AC-M24-004:** Invalid webhook is not processed.
- [ ] **AC-M24-005:** Unverified legal/provider requirement remains visibly blocked.

### M25 — Global UX, Accessibility, Language, and Prototype

**Purpose:** Deliver a clean, intuitive intelligent sales workspace with complete state visibility and an implementation-first prototype path.

**Mapped BRD requirements:** BR-UX-001—012, BR-REG-003, BR-SAL-025

#### Actors

- All product users
- Product/UX team
- System

#### Representative User Stories


- **US-M25-001:** As a All product users, I want this capability to deliver a clean, intuitive intelligent sales workspace with complete state visibility and an implementation-first prototype path.

- **US-M25-002:** As the platform, I need all protected actions in this module to be permissioned, auditable, tenant-scoped, and state-valid.

- **US-M25-003:** As an authorized reviewer, I want clear loading, empty, failure, and recovery behavior so that work never depends on hidden system state.

#### Preconditions

- Approved product model exists.

#### Inputs

- Actor/permission/context
- Data/state
- Actions/errors
- Language settings

#### Outputs

- Responsive accessible screens
- Consistent navigation
- Loading/empty/error/permission/limit states
- Initial prototype

#### Main Flow

1. Use separate Public, Organization, Project, and Super Admin information architectures.
2. Present action-first dashboards and deep-linked work queues.
3. Use one Unified Sales Workspace for lead queue, conversation, AI insight, and controls.
4. Show current AI/human control state and pending approval clearly.
5. Show live/test status and integration readiness prominently where applicable.
6. Provide English/Arabic conversation behavior; design all UI components so RTL/localization can be completed.
7. Prototype the path from registration through lead conversation before deep visual polish.

#### Alternative, Exception, and Failure Flows

- No data: provide a meaningful CTA, not blank screen.
- Permission denied: explain access without exposing hidden resource details.
- Limit reached: explain what is blocked and what continues.
- Integration missing/failure: show action and fallback/pause behavior.
- Long AI/provider operation: show progress, safe cancel/return, and final result/action.
- Mobile: prioritize Action Center, lead list, conversation, takeover, and approvals.

#### Product Rules

- Product feels like an intelligent sales workspace, not an Agent debugger.
- User always knows what happens next and who controls the conversation.
- Technical traces stay in Super Admin operations, not normal customer UI.
- First prototype scope ends at lead conversation but includes navigation to key acquisition paths.

#### Responsibility Split


**AI responsibilities**

- Provide concise explanations/recommendations; no raw hidden chain-of-thought.


**Deterministic application responsibilities**

- Consistent design system/state components, permission-aware navigation, accessibility hooks, localization-ready layout.


**Human responsibilities**

- Complete tasks with minimal configuration and clear review/approval controls.

#### Permissions

- UI reflects backend authorization; no UI-only security.

#### Data Requirements

- ScreenState
- NavigationItem
- UserPreference
- LocalizationResource
- AccessibilityMetadata

#### UX States

- Loading
- Empty
- Error
- Permission
- Limit
- Integration
- AI working
- Approval
- Success
- Live/Ended

#### Notifications and Action Center

- Consistent in-app/email alerts; no excessive routine alerts.

#### Audit Requirements

- Major UX action audit through underlying business event.

#### Analytics Events and Metrics

- Task completion
- Drop-off
- Error recovery
- Time to first Project/lead/conversation
- Accessibility defects

#### Security and Privacy

- Do not expose sensitive traces/IDs
- Authenticated deep links
- Safe errors

#### B2B / B2C Behavior

- Sales Agent supports English/Arabic. Full RTL UI is a planned specification; prototype should avoid layouts that block it.

#### Formal Product Requirements

| PR ID | Requirement/behavior | Type | BRD trace |
| --- | --- | --- | --- |
| PR-25-001 | Product feels like an intelligent sales workspace, not an Agent debugger. | Product Rule | BR-UX-001—012, BR-REG-003, BR-SAL-025 |
| PR-25-002 | User always knows what happens next and who controls the conversation. | Product Rule | BR-UX-001—012, BR-REG-003, BR-SAL-025 |
| PR-25-003 | Technical traces stay in Super Admin operations, not normal customer UI. | Product Rule | BR-UX-001—012, BR-REG-003, BR-SAL-025 |
| PR-25-004 | First prototype scope ends at lead conversation but includes navigation to key acquisition paths. | Product Rule | BR-UX-001—012, BR-REG-003, BR-SAL-025 |
| PR-25-005 | Prototype demonstrates registration, onboarding, Project, Strategy, product content, acquisition, leads, and conversation. | Acceptance Behavior | BR-UX-001—012, BR-REG-003, BR-SAL-025 |
| PR-25-006 | Every major screen has loading/empty/error/permission/limit state. | Acceptance Behavior | BR-UX-001—012, BR-REG-003, BR-SAL-025 |
| PR-25-007 | Control owner is visible in conversation. | Acceptance Behavior | BR-UX-001—012, BR-REG-003, BR-SAL-025 |
| PR-25-008 | Action Item opens exact work location. | Acceptance Behavior | BR-UX-001—012, BR-REG-003, BR-SAL-025 |
| PR-25-009 | Core flows are usable on common desktop and mobile widths. | Acceptance Behavior | BR-UX-001—012, BR-REG-003, BR-SAL-025 |

#### Acceptance Criteria

- [ ] **AC-M25-001:** Prototype demonstrates registration, onboarding, Project, Strategy, product content, acquisition, leads, and conversation.
- [ ] **AC-M25-002:** Every major screen has loading/empty/error/permission/limit state.
- [ ] **AC-M25-003:** Control owner is visible in conversation.
- [ ] **AC-M25-004:** Action Item opens exact work location.
- [ ] **AC-M25-005:** Core flows are usable on common desktop and mobile widths.

## 11. Agent Registry

| Agent ID | Agent | Purpose | Inputs | Outputs | Hard boundary |
| --- | --- | --- | --- | --- | --- |
| A-001 | Company Understanding Agent | Understand/classify business; ask questions; create Business Overview | Registration data, sources, digital presence | Company understanding, questions, reusable/private proposal | No Project Strategy, no self-approval |
| A-002 | Project Understanding Agent | Understand Project/products and readiness | Approved company context, Project brief/files | Project/Product Understanding, catalogue proposal, missing data | No Strategy approval, no product invention |
| A-003 | Strategy Agent | Generate complete Project Strategy and qualification framework | Approved understanding/objective/capabilities | Structured Strategy, ICP/personas, acquisition direction | No lead finding, campaign publication, messages |
| A-004 | Lead Discovery Agent | Choose Web/Meta/Both and execute approved Web discovery | Strategy, ICP, geography, lead limit, providers | Acquisition plan, candidates, provenance, stop reason | No final qualification or outreach |
| A-005 | Lead Enrichment Agent | Gather purpose-limited evidence | Lead identity, provenance, Project research objective | Reusable/project facts, match recommendation, missing data | No outreach or direct merge |
| A-006 | Lead Strategist Agent | Maintain dynamic Project Lead sales Strategy | Strategy, product, lead evidence, memory, latest event | Score/state recommendation, hypothesis, objective, approach, NBA, follow-up/handover | No message send or protected state |
| A-007 | Sales / Conversation Agent | Communicate naturally on approved channels | Lead Strategy, product knowledge, memory, channel settings | Message/action, evidence, attachments, material-event signal | No independent Strategy rewrite or invented facts |
| A-008 | Campaign Agent | Prepare/monitor Meta campaigns and creative | Meta brief, Strategy, product, budget/end-date constraints | Configuration, form/copy/creative guide, review, recommendations | No publish/spend action without human/service |
| A-009 | Customer Simulator Agent | Act as realistic prospect in sandbox | Scenario/persona/hidden state | Simulated messages/behavior | No production data/side effects |
| A-010 | Evaluator Agent | Evaluate versions and behavior independently | Transcript, traces, expected behavior, versions | Metrics, hard failures, pass/fail/review | No production activation or prompt changes |

## 12. Agent Interaction Matrix

| Caller/Agent | May request/consume | Governed handoff rule |
| --- | --- | --- |
| A-001 Company Understanding | None normally | Application persists approved Company Understanding before downstream use. |
| A-002 Project Understanding | A-001 context (read only) | Consumes approved Company Understanding; returns result to application. |
| A-003 Strategy | A-002 output | Runs only on approved understanding; returns structured Strategy to application. |
| A-004 Lead Discovery | A-003 Strategy; A-008 handoff for Meta | May request A-005 enrichment for candidates through application. |
| A-005 Lead Enrichment | A-004 or A-006 request | Returns validated evidence to application; A-006 then re-runs. |
| A-006 Lead Strategist | A-005 evidence; A-007 material event | May request A-005; sends authoritative approach to A-007 only after persistence. |
| A-007 Sales | A-006 current Lead Strategy | May flag material evidence to trigger A-006; cannot invoke arbitrary agents. |
| A-008 Campaign | A-004 Meta brief; A-003 Strategy | Provider publication occurs through Campaign Service only. |
| A-009 Simulator | A-006/A-007 in sandbox | Only evaluation environment. |
| A-010 Evaluator | A-009 scenario + A-006/A-007 traces | Produces evaluation result for Super Admin; no product state mutation. |

## 13. Agent Run and Contract Standard

Every Agent definition/version shall specify:

- Agent ID, purpose, owner, trigger, active status, version lifecycle.
- Authorized context scopes and context assembler.
- Structured input and output schemas.
- Allowed tools and allowed specialist Agents.
- Forbidden actions and protected boundaries.
- Prompt version, model/provider, credential reference, runtime settings.
- Retry, timeout, loop, cost, and fallback behavior.
- Test-mode behavior and evaluation suite.
- Audit/trace requirements and sensitive-content classification.
- Acceptance criteria and hard-failure conditions.

Every run shall return one normalized outcome: `SUCCESS`, `RETRYABLE_FAILURE`, `NEEDS_INFORMATION`, `NEEDS_HUMAN_REVIEW`, `BLOCKED`, or `FAILED`.

## 14. Conceptual Data Ownership and Isolation

| Scope | Authoritative data |
| --- | --- |
| Platform scope | Plan definitions, platform users/permissions, provider definitions, agent registry/versions, global feature flags, system audit/operations. |
| Organization private scope | Subscription snapshot, memberships/permissions, Company Understanding, private Organization knowledge, Organization integrations, settings/audit. |
| Project private scope | Project/Product Understanding, Product Catalogue, Strategy, campaigns, acquisition plans, Project Leads, qualification, Lead Strategy, conversations, handovers, opportunities, conversions, analytics. |
| Reusable Business Pool | Permitted company identities/classifications/evidence and company hierarchy. |
| Reusable Master Lead Pool | Permitted person identity/contact/professional/interest/context signals and person-business relationships. |
| Sensitive execution scope | Agent prompts/responses, retrieved context, tool traces, full conversations; restricted and retention-controlled. |

## 15. Core Entity Dictionary

| Entity | Minimum product-level fields/relationships |
| --- | --- |
| Organization | id, identity, status, Company Understanding version, subscription snapshot, settings, external-action state |
| Membership | user, Organization, Owner flag, status, explicit permissions |
| Project | id, Organization, name, mode, lifecycle, objective, understanding/Strategy versions, effective limits |
| Product Catalogue Item | Project, baseline snapshot, sales-support content/assets, approval/source/version |
| Strategy Version | Project, source understanding versions, structured components, revision instruction, status, approver |
| Business Pool Entity | identity, CR/jurisdiction, domains, classification, hierarchy, evidence, reuse/correction flags |
| Master Lead | person identity/contact, reusable signals, person-business relationships, evidence, suppression/correction flags |
| Project Lead | Project + Master Lead, source, scores/states, Lead Strategy, conversation, outcome |
| Campaign | Project, Strategy/version, provider refs, audience, budget, end date, state, metrics |
| Conversation | Project Lead, channel threads, control state, messages, summary, memory, actor/version refs |
| Lead Strategy | Project Lead, current objective/hypothesis/approach, objections, attempts, next action, follow-up/handover |
| Follow-Up Proposal | Project Lead, channel, timing, objective, message, reason, approval/delegation, state |
| Handover | Project Lead, brief, state, claimant/owner, timestamps, outcome requirement |
| Conversion | Project Lead, type, evidence, human actor, source/campaign lineage, optional value, time |
| Agent Version | Agent, prompt/model/credential/tool config, lifecycle, evaluation eligibility |
| Agent Run | versions, context refs, tool calls, outcome, latency, cost/usage, sensitive-content classification |
| Action Item | scope, type, affected resource, reason, priority, assignee eligibility, state, deep link |
| Audit Event | actor, scope, action, resource, old/new where appropriate, time, correlation ID |

## 16. Integration Registry

| Integration ID | Integration | Scope | Configuration location | Business use | Mandatory controls |
| --- | --- | --- | --- | --- | --- |
| INT-META-ADS | Meta Ads | Platform-Managed Organization | Super Admin → Organization X → Integrations | Campaign create/publish/update/status/metrics/leads | Explicit publish; budget/end-date changes user-authorized; provider policy verification required |
| INT-WA | WhatsApp Messaging | Platform-Managed Organization | Super Admin → Organization X → Integrations | Inbound/outbound messages; primary paid Meta destination | Eligibility/suppression; webhook verification; policy/template requirements pending official verification |
| INT-IG | Instagram Messaging | Platform-Managed Organization | Super Admin → Organization X → Integrations | Direct inbound/outbound messaging where supported | Capability/permission dependent; channel transitions controlled |
| INT-FB | Facebook Messenger | Platform-Managed Organization | Super Admin → Organization X → Integrations | Direct inbound/outbound messaging where supported | Capability/permission dependent; channel transitions controlled |
| INT-GSEARCH | Google/Search Provider | Platform | Super Admin → Platform Integrations | Web/Maps/search-based discovery and evidence | Source/API terms and storage restrictions require verification |
| INT-EMAIL | Generic Business Mailbox | Organization Self-Managed | Organization Admin → Email Integration | Outbound SMTP-equivalent + inbound IMAP/provider-equivalent threaded replies | Sender identity, deliverability, anti-spam, suppression, compatible provider validation |
| INT-AI | AI Model Provider | Platform/per Agent | Super Admin → Platform Integrations + Agent config | LLM/model execution | Credential refs, model per Agent, fallbacks only if approved |
| INT-FILE | File Storage/Processing | Platform | Platform configuration | Company/Project sources, product assets, attachments | Detailed file lifecycle deferred; scanning and signed access mandatory |
| INT-NOTIFY | Email Notification Provider | Platform | Platform configuration | Operational staff/customer email alerts | Minimum content; Action Center remains authoritative |

## 17. Integration Failure and Readiness Rules

- An entitled but unconfigured integration is visible as unavailable and creates an Action Item; it does not block Project onboarding.
- No Agent may treat a missing/failed provider as permission to use an unapproved source or channel.
- Approved channel fallback is allowed only when the destination is healthy, permitted by Strategy, contactable, eligible, and relevant.
- If no fallback exists, only the affected activity pauses; the user receives in-app/email alert.
- Provider errors are normalized, retain provider references, and support safe bounded retry where permitted.
- Inbound provider events are verified, deduplicated, replay-protected, and mapped to the correct tenant/Project/lead.
- Real MVP integration acceptance is separate from sandbox/mock adapter acceptance.

## 18. Analytics Event Catalogue

| Event ID | Event | Minimum properties |
| --- | --- | --- |
| EVT-REG-SUBMITTED | Registration submitted | application_id, plan_id, country |
| EVT-ORG-ACTIVATED | Organization activated | organization_id, subscription_snapshot_id |
| EVT-COMPANY-UNDERSTANDING-APPROVED | Company Understanding approved | organization_id, version_id, approver |
| EVT-PROJECT-CREATED | Project created | project_id, mode |
| EVT-PROJECT-LIVE | Project onboarding approved | project_id, strategy_version, channels |
| EVT-STRATEGY-GENERATED | Strategy generated | project_id, version, revision_consumed |
| EVT-ACQ-PLAN-APPROVED | Acquisition Plan approved | project_id, sources |
| EVT-DISCOVERY-COMPLETED | Web discovery completed | run_id, target, candidates, contactable, allocated, stop_reason |
| EVT-CAMPAIGN-PUBLISHED | Campaign published | campaign_id, provider_id, budget, end_date |
| EVT-LEAD-CAPTURED | Master Lead captured | master_lead_id, source, project_id? |
| EVT-PROJECT-LEAD-ACTIVATED | Project Lead allocated | project_lead_id, initial_score, source |
| EVT-FIRST-OUTREACH | First outreach sent | project_lead_id, channel, agent_version |
| EVT-LEAD-REPLIED | Lead replied | project_lead_id, channel, material_event |
| EVT-QUALIFICATION-CHANGED | Score/state changed | project_lead_id, old/new, evidence |
| EVT-FOLLOWUP-PROPOSED | Follow-up proposed | project_lead_id, timing, approval_mode |
| EVT-FOLLOWUP-DECIDED | Follow-up approved/edited/rejected/delegated | actor, decision |
| EVT-HANDOVER-READY | Handover created | project_lead_id, reason |
| EVT-HANDOVER-CLAIMED | Handover claimed | project_lead_id, human_user_id |
| EVT-CONVERSION-RECORDED | Conversion recorded | project_lead_id, type, source, campaign, human |
| EVT-PROJECT-ENDED | Project ended | project_id, actor, reason |
| EVT-AGENT-RUN | Agent run completed | agent/version/prompt/model/status/latency/tools |
| EVT-SECURITY-BLOCK | Protected action blocked | actor/context/policy/reason |

## 19. Non-Functional Requirements

| NFR ID | Area | Requirement/target | Status |
| --- | --- | --- | --- |
| NFR-SEC-001 | Tenant isolation | 100% of protected requests, jobs, agent contexts, files, and provider events are tenant/project scoped and server-authorized. | Mandatory baseline |
| NFR-SEC-002 | Authorization | No protected state change relies on UI hiding or Agent self-authorization. | Mandatory baseline |
| NFR-SEC-003 | Secrets | Provider/model/email credentials are encrypted/protected and represented by secret references in UI/logs. | Mandatory baseline |
| NFR-REL-001 | Idempotency | Campaign publish, spend updates, message sends where retryable, webhook processing, handover claims, and conversion recording have explicit idempotency/replay behavior. | Mandatory baseline |
| NFR-OBS-001 | Observability | Every Agent run and external side effect has correlation IDs, actor/context, version, status, timestamps, and error information. | Mandatory baseline |
| NFR-PRV-001 | Sensitive logs | Operational telemetry is available by default; full prompt/response/conversation content requires elevated audited access. | Mandatory baseline |
| NFR-UX-001 | Responsive UX | Core dashboards, Action Center, Lead list, Sales Workspace, approvals, and takeover are usable on desktop and common mobile widths. | MVP requirement |
| NFR-UX-002 | Accessibility | Keyboard navigation, semantic labels, focus, contrast, error association, and accessible status communication are required; formal conformance target requires Stage 4/UX validation. | MVP target |
| NFR-L10N-001 | Language | Sales AI supports English and Arabic behavior from MVP baseline; full UI RTL/localization scope remains deferred. | MVP + deferred detail |
| NFR-PERF-001 | Core UI response | Initial target: non-AI/non-provider pages provide interactive feedback within 3 seconds at the 95th percentile under MVP load. | Proposed; technical validation required |
| NFR-PERF-002 | Long operations | AI/discovery/campaign/provider operations display immediate progress/status and do not leave the user without feedback. | MVP requirement |
| NFR-REL-002 | Provider outage | Provider failures are isolated, retried only within policy, and produce fallback/pause/action behavior. | MVP requirement |
| NFR-SCL-001 | Commercial scale | Numeric and Unlimited Plans must not remove technical rate/cost/abuse controls. | Mandatory baseline |
| NFR-MNT-001 | Maintainability | One authoritative implementation of each business rule; Agent prompts cannot duplicate protected application logic. | Mandatory baseline |
| NFR-PORT-001 | Provider portability | All external providers use adapters and normalized application errors/status. | MVP architecture requirement |
| NFR-TEST-001 | Testability | All agents have structured contracts and scenario evaluation; all external side effects have sandbox/mock capability even though real integrations are required for MVP completion. | MVP requirement |
| NFR-DR-001 | Backup/restore | Backup, restore, retention, and deletion targets require Stage 4 validation before production release. | Open technical specification |

## 20. Security and Privacy Checklist

- [ ] Every protected route/service/job/tool applies tenant and Project scope.
- [ ] Authentication, permission, Project access, entitlement, readiness, state, and ownership are checked server-side.
- [ ] IDOR and mass assignment tests exist for every core resource.
- [ ] Suppression/eligibility is checked immediately before every outbound send.
- [ ] Agent context excludes unrelated Organizations/Projects/leads.
- [ ] Agents cannot call unregistered tools or access raw credentials.
- [ ] Provider webhooks/mail events are verified, replay-protected, and idempotent.
- [ ] Uploaded/external/inbound content is treated as untrusted prompt-injection input.
- [ ] Sensitive traces require elevated audited access.
- [ ] Support access is purpose-limited and audited.
- [ ] Master Lead/Business Pool reuse is source-aware and excludes tenant-private content.
- [ ] Production data is not used casually in local/test environments.
- [ ] Campaign spend and final commercial actions require explicit authorized human action.

## 21. Test and Evaluation Strategy

### Test layers
- Unit tests for deterministic business rules, state transitions, limits, eligibility, suppression, identity matching, and permission evaluation.
- Integration tests for database/services, email threading, provider adapters, webhooks, background jobs, files, and Agent tool contracts.
- End-to-end tests for the complete customer journeys.
- Security tests for tenant leakage, IDOR, mass assignment, prompt injection, tool abuse, replay, secret exposure, and support access.
- Agent evaluations using versioned Customer Simulator scenarios and Evaluator hard-failure rules.
- Regression scenarios for every corrected production Agent failure.

### Critical evaluation hard failures
- Invented price, feature, availability, guarantee, discount, or commitment.
- Suppression/opt-out violation.
- Cross-tenant/private information exposure.
- Message sent in Human Controlled mode.
- Unapproved follow-up or campaign publication/spend.
- Final commercial action performed autonomously.
- Unauthorized tool/provider access.

### 21.1 Critical End-to-End Scenarios

| Scenario ID | Scenario | Required journey |
| --- | --- | --- |
| E2E-001 | B2B Web→Email | Register/activate company; Company Understanding; B2B Project; Strategy recommends Web; approve discovery; find company + practical influencer + email; score/rank; first outreach; threaded reply; dynamic qualification; human handover; record conversion. |
| E2E-002 | B2C Meta→WhatsApp | B2C Project; Campaign Agent creates plan/creative guide; user publishes; consumer initiates WhatsApp; dynamic qualification/objection; product media; human handover; order outcome. |
| E2E-003 | Follow-up approval | Conversation becomes inactive; AI proposes contextual follow-up; user edits/approves scheduled send; lead replies before time; scheduled send cancels and active AI resumes. |
| E2E-004 | Delegated follow-up | User delegates one lead; AI strategically follows up; user revokes; no further automatic follow-up. |
| E2E-005 | Human takeover/return | Human takes over mid-conversation; AI stops; human answers; return to AI triggers full reassessment and correct continuation. |
| E2E-006 | Suppression | Lead opts out; all pending sends cancel; all Agent/provider paths block future applicable outreach. |
| E2E-007 | Meta overflow | Project reaches lead limit; additional Meta lead enters Master Lead Pool with attribution; no Project autonomous sales; user alerted; Super Admin increases limit; lead can then be activated under controlled allocation. |
| E2E-008 | Integration outage | Email/WhatsApp fails; approved healthy fallback used when eligible; otherwise affected activity pauses and action/notification created. |
| E2E-009 | Agent version evaluation | Draft Sales Agent/Lead Strategist versions run against simulator scenarios; hard failure blocks activation; corrected scenario becomes regression; Super Admin activates passing version. |
| E2E-010 | Tenant isolation | Organization A user/agent/tool attempts Organization B Project/lead/file/trace access and is denied without data disclosure. |


## 22. Initial Three-Day UX Prototype Scope

The prototype is a visual/interaction prototype, not production functionality. It shall demonstrate:

1. Landing, Plans, Company Registration, Pending Approval.
2. Company Understanding input, analysis, clarification, final approval.
3. Organization Dashboard and Create First Project.
4. Project brief/documents, Project/Product Understanding review.
5. Strategy review/revision and channel/Sales AI settings.
6. Project Dashboard and generated Product Catalogue.
7. Lead Acquisition Plan.
8. Meta Campaign configuration/creative guide/media upload.
9. Web Discovery Plan/results.
10. Leads list and Unified Sales Workspace.
11. AI Sales Insight, autonomous conversation, follow-up approval, unresponsive alert, and Take Over control.

Prototype navigation and components must be compatible with the later permission, state, mobile, and Arabic/RTL requirements.

## 23. MVP Acceptance Walkthrough

1. Applicant registers company and selects Plan.
2. Super Admin reviews and activates Organization and subscription.
3. Owner completes and approves Company Understanding.
4. User creates Project and supplies Project/product information and documents.
5. AI asks targeted questions; user approves Project/Product Understanding and objective.
6. User generates, optionally revises within allowance, and approves one Strategy package.
7. User configures Project/channel Sales AI behavior and approves Project onboarding.
8. Project becomes LIVE and user completes Product Catalogue content.
9. Lead Discovery recommends Web, Meta, or Both.
10. For Meta: user prepares media from AI guide, approves lead capture, explicitly publishes, and sees campaign/lead results.
11. For Web: user approves search plan, discovery creates contactable Master Leads, and highest-ranked leads are allocated.
12. Lead Enrichment and Lead Strategist create the initial lead context/Strategy.
13. Sales AI automatically starts eligible outreach or responds to inbound WhatsApp/email/social message.
14. AI conducts natural active conversation, adapts Strategy, and shares approved product content.
15. Normal inactive follow-up is approved or explicitly delegated.
16. Authorized human monitors and can Take Over at any time.
17. Final MVP action produces Handover Ready; an authorized user claims it.
18. Human completes and records the final outcome/conversion.
19. Project/campaign/source/AI/human analytics reflect the journey.
20. Super Admin can observe permitted end-to-end results and exact Agent/provider versions.

## 24. Definition of Done

- [ ] Approved BRD/PRD requirement and acceptance criteria implemented.
- [ ] Server-side authentication, tenant scope, explicit permission, Project access, entitlement, and resource/state checks.
- [ ] IDOR and mass-assignment review completed.
- [ ] AI tool authorization and context isolation verified.
- [ ] Suppression and communication eligibility checked for every outbound path.
- [ ] External side effects are explicit, idempotent, auditable, and test/sandbox safe.
- [ ] Loading, empty, error, permission, limit, integration, and recovery states implemented.
- [ ] Audit events and analytics events emitted.
- [ ] Unit, integration, end-to-end, security, and agent-evaluation tests pass.
- [ ] No unresolved critical/high security issue.
- [ ] Documentation, decision register, and traceability updated.
- [ ] Local build/deployment and the relevant end-to-end walkthrough verified.

## 25. Requirements Traceability

- Every `PR-*` item maps to one or more `BR-*` requirements in the Product Requirement Index and module sections.
- Every screen maps to at least one module and one permission/state model.
- Every Agent result maps to a structured contract and deterministic application validation.
- Every external side effect maps to an authorized application service and provider adapter.
- Every high-risk action maps to audit, notification/action behavior, and security tests.
- Every implementation milestone must cite the exact BR/PR/decision IDs it implements.

## 26. Open Questions and Deferred Decisions

### Open

| ID | Area | Unresolved matter |
| --- | --- | --- |
| OQ-001 | Campaign lifecycle D-274 | The recommended lifecycle is defined but still requires explicit confirmation. |
| OQ-002 | Plan prices and numeric allowances | Commercial values are intentionally not yet set. |
| OQ-003 | Generic inbound mailbox implementation | The product requirement is generic inbound/outbound mailbox support; exact IMAP/provider compatibility belongs to integration design. |
| OQ-004 | Provider and legal policy verification | Meta, WhatsApp, Instagram, Facebook, Google/Search, email, prospect data, privacy, and Oman/GCC obligations require current official research and legal review. |
| OQ-005 | Full Arabic/RTL UI baseline | The exact degree of UI localization, dialect behavior, and RTL coverage remains to be locked. |

### Deferred

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

## 27. Required Specifications Before Production Coding

- Approved campaign state model confirmation (D-274).
- System context and domain architecture.
- Conceptual/logical data model, ERD, entity dictionary, retention/deletion, history/versioning.
- Complete API inventory and contracts for client, Control Panel, internal services, Agent tools, jobs, webhooks, and provider adapters.
- Google ADK implementation topology, session/context assembly, tool schemas, callback/observability, fallback, and cost design.
- Security threat model, authorization model, secrets, webhook/replay, file/URL processing, prompt-injection controls, and test plan.
- Current official Meta/WhatsApp/Instagram/Facebook/Google/email policy and legal/privacy review.
- Deployment, local environment, background jobs, storage, observability, backup/restore, and rollback.
- Milestone roadmap, Definition of Ready, release/test strategy, and coding-session prompts.

## 28. Change-Control Protocol

1. A proposed change identifies affected decision, BR, PR, state, screen, Agent, integration, data, security, tests, and milestones.
2. The Project Owner approves, rejects, modifies, or defers it.
3. Approved changes receive a stable Change Request/Decision ID.
4. The BRD/PRD and traceability are updated before implementation.
5. If architecture changes, an ADR is required.
6. Code never becomes the new requirement merely because it was implemented.


---

**Approval statement:** `AI_Sales_Agent_MVP_Master_PRD_Blueprint_v1.0 is approved as the authoritative MVP product-requirements baseline.`


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

