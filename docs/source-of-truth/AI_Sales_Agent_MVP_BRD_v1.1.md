# AI Sales Agent — MVP Business Requirements Document

| Field | Value |
| --- | --- |
| Version | 1.1 |
| Status | Under Review — v1.0 + Approved Super Admin Amendment |
| Owner | Zeyad Al Gharabi |
| Approver | Project Owner — explicit approval required |
| Created | 2026-08-16 |
| Last Updated | 2026-08-18 |
| Source Documents | Stage 1 Lock Report; PROJECT_START_HERE v0.1; earlier Strategy/Channel/Agent PRDs as reference only |
| Related Decisions | D-017 through D-283 as consolidated in the Stage 1 Lock Report |
| Scope | Business requirements for the first usable multi-tenant AI Sales and Marketing MVP |


> **Source-of-truth statement:** Once approved, this BRD is authoritative for MVP business scope, outcomes, actors, business rules, and acceptance. The Master PRD is authoritative for detailed product behavior. Exact system architecture, API contracts, data schemas, provider mappings, and deployment decisions require later approved Stage 4 specifications.

## Table of Contents

- [1. Executive Summary](#1-executive-summary)
- [2. Business Problem](#2-business-problem)
- [3. Business Objectives](#3-business-objectives)
- [4. Business Value](#4-business-value)
- [5. Customer and Market Model](#5-customer-and-market-model)
- [6. Stakeholders](#6-stakeholders)
- [7. Product Scope](#7-product-scope)
- [8. Operating Model](#8-operating-model)
- [9. Core Business Objects](#9-core-business-objects)
- [10. Business Requirements Catalogue](#10-business-requirements-catalogue)
- [11. Business Rule Catalogue](#11-business-rule-catalogue)
- [12. Roles, Responsibility, and Authorization](#12-roles-responsibility-and-authorization)
- [13. AI and Human Responsibility Matrix](#13-ai-and-human-responsibility-matrix)
- [14. Plans, Limits, and Entitlements](#14-plans-limits-and-entitlements)
- [15. Control Panel Business Requirements](#15-control-panel-business-requirements)
- [16. Analytics and Success Measures](#16-analytics-and-success-measures)
- [17. Security and Privacy Expectations](#17-security-and-privacy-expectations)
- [18. Compliance and External Policy Dependencies](#18-compliance-and-external-policy-dependencies)
- [19. Assumptions and Dependencies](#19-assumptions-and-dependencies)
- [20. Risks](#20-risks)
- [21. Deferred and Open Items](#21-deferred-and-open-items)
- [22. MVP Business Acceptance Criteria](#22-mvp-business-acceptance-criteria)
- [23. Traceability Matrix](#23-traceability-matrix)
- [24. Change Control and Approval](#24-change-control-and-approval)

## 1. Executive Summary

The AI Sales Agent platform enables a business to register, be approved, teach the platform about the company, create a commercial Project, define products and objectives, generate an approved sales/marketing Strategy, acquire leads through Meta or Web/Google research, enrich and qualify those leads, conduct human-like multi-channel AI sales conversations, hand over final actions to authorized humans, record conversions, and learn from outcomes.

The product is a multi-tenant SaaS for B2B and B2C commercial operations. It is not a generic chatbot or full CRM. Its central business value is the combination of governed acquisition, persistent lead/business intelligence, dynamic lead-specific Strategy, autonomous active sales conversation, human control, attribution, and platform-level agent administration.

## 2. Business Problem

Businesses often operate sales and marketing as disconnected activities. Company/product knowledge is incomplete, targeting depends on individual expertise, lead acquisition does not reliably connect to sales follow-through, sales representatives repeat manual research and follow-up, and campaign reporting stops at lead volume instead of commercial outcome.

The platform addresses these problems by creating one controlled workflow from approved business understanding to human-confirmed conversion, while preserving customer ownership, permissions, safety, provider compliance, and auditability.

## 3. Business Objectives

| ID | Objective |
| --- | --- |
| BO-001 | Reduce the time and expertise required for a business to move from company information to an executable sales and marketing operation. |
| BO-002 | Acquire relevant B2B and B2C leads through supported paid and research-driven channels. |
| BO-003 | Provide dynamic, professional AI sales conversations that improve qualification and prepare leads for human final action. |
| BO-004 | Preserve human authority over spend, follow-up approval by default, final commercial action, and protected commitments. |
| BO-005 | Create a reusable, governed Business Pool and Master Lead Pool without exposing tenant-private information. |
| BO-006 | Provide traceability from acquisition source through conversation, handover, and human-confirmed conversion. |
| BO-007 | Enable the platform operator to manage Organizations, Plans, agents, models, credentials, integrations, usage, safety, and support. |
| BO-008 | Support B2B and B2C Projects without forcing one model's qualification or identity assumptions onto the other. |
| BO-009 | Make agent behavior testable, versioned, observable, reversible, and governed through Google ADK plus application controls. |
| BO-010 | Deliver an intuitive MVP that users can operate without reconstructing critical sales work outside the platform. |

## 4. Business Value

- Reduce onboarding and Strategy-planning effort through AI-guided understanding.
- Turn approved Project knowledge into repeatable acquisition and sales behavior.
- Create more relevant lead pools through adaptive Web/Meta source selection.
- Preserve and improve permitted platform intelligence instead of re-researching every identity from zero.
- Give businesses an AI salesperson that adapts by lead while keeping final commercial authority human.
- Make campaign and sales performance measurable from source through human conversion.
- Give the platform operator complete governance over customers, agents, models, integrations, limits, and safety.

## 5. Customer and Market Model

### B2B Organizations
May acquire businesses and practical buyer influencers through Web/Google discovery, Meta, or both. Project qualification may consider company fit, operational need, user/influencer role, product fit, intent, and ability to advance the decision.

### B2C Organizations
Primarily acquire consumers through Meta and supported messaging, while remaining compatible with other permitted sources. Qualification uses consumer-relevant fit, need, intent, eligibility, geography, readiness, and contactability.

### Hybrid Organizations
May operate both B2B and B2C Projects. Each MVP Project has one primary mode to preserve clear identity, targeting, scoring, Strategy, and reporting behavior.

## 6. Stakeholders

| Stakeholder | Responsibility |
| --- | --- |
| Project Owner | Approves product scope, commercial model, and baselines. |
| Platform Super Admin | Operates platform, Plans, Organizations, integrations, agents, models, credentials, observability, and safety controls. |
| Platform Admin/Support | Handles approved registration/support activities within restricted permissions. |
| Organization Owner | Owns the customer account and grants user permissions/Project access. |
| Authorized Organization User | Creates/approves Projects, campaigns, follow-ups, conversations, and outcomes according to explicit permissions. |
| Sales Operator | Monitors AI sales, approves follow-ups, takes over, and completes final actions where permitted. |
| Marketing Operator | Prepares, reviews, publishes, and monitors campaigns where permitted. |
| Lead/Prospect | Receives or initiates permitted communications and interacts with Sales AI/human users. |
| Security/Privacy Reviewer | Validates tenant isolation, data reuse, suppression, logging, retention, and provider risks. |
| Legal/Policy Reviewer | Reviews outreach, profiling, data-source licensing, provider terms, and jurisdiction obligations. |

## 7. Product Scope

### Included in MVP

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

### Explicit Non-Goals

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

## 8. Operating Model

```text
Platform
└── Organization
    ├── Company Understanding
    ├── Users / permissions / Project access
    ├── Subscription / Plan / integrations
    └── Projects
        ├── Project/Product Understanding
        ├── Product Catalogue
        ├── Strategy
        ├── Acquisition / Campaigns
        ├── Project Leads
        ├── Conversations / Lead Strategies
        ├── Human handovers
        └── Conversions / analytics
```

Reusable intelligence is separated from tenant-private sales data:

```text
Business Pool (companies)        Master Lead Pool (people)
            \                    /
             \ person-business  /
              relationships
                    |
              Project Lead
      (private Project-specific sales context)
```

## 9. Core Business Objects

| Object | Business meaning |
| --- | --- |
| Platform | The multi-tenant SaaS and its operator-controlled services. |
| Organization | A customer tenant and security boundary. |
| User | A platform identity that may belong to multiple Organizations. |
| Membership | A User's relationship to an Organization, including Owner status and permissions. |
| Permission | Explicit authorization to perform a sensitive or scoped action. |
| Project | The primary commercial execution unit; one primary B2B or B2C mode; lifecycle ONBOARDING/LIVE/ENDED. |
| Company Understanding | Approved versioned intelligence describing the registered business and market context. |
| Project Understanding | Approved versioned Project scope and objective. |
| Product Understanding | Approved strategy-critical facts for products/services included in the Project. |
| Product Catalogue Item | An approved Project product plus editable sales-support media/content. |
| Strategy | One approved package containing structured market, ICP/persona, sales, marketing, acquisition, qualification, and conversion direction. |
| Business Pool Entity | A reusable permitted company/organization identity. |
| Master Lead | A reusable permitted person/consumer identity. |
| Project Lead | A Project-specific relationship containing fit, score, Strategy, conversation, state, and outcome. |
| Research Candidate | A potentially relevant discovery result that is not yet contactable. |
| Campaign | A Project-scoped Meta acquisition activity with user-controlled publication/spend. |
| Conversation | A Project Lead's channel-specific or unified sales interaction history. |
| Lead Strategy | The evolving Project-specific sales approach maintained by Lead Strategist. |
| Follow-up Proposal | An AI-recommended future contact requiring approval unless explicitly delegated. |
| Handover | A claimable transition from AI to authorized human control. |
| Opportunity | Optional continuing deal record after qualification/handover. |
| Conversion | An auditable human-confirmed Project success event. |
| Plan | Commercial package containing entitlements and numeric/unlimited limits. |
| Integration | A platform or Organization-scoped provider connection. |
| Agent | A versioned Google ADK-based reasoning component. |
| Agent Run | A version-pinned, auditable execution of an agent. |
| Action Item | Authoritative unresolved work record. |
| Notification | In-app/email alert that directs an authorized user to an Action Item or event. |
| Audit Event | Application-created immutable record of an important action. |

## 10. Business Requirements Catalogue


### 10.1 Registration, Identity, and Tenancy


| Requirement ID | Business requirement | Primary actor/scope | Priority | Delivery | Objective | Decision trace |
| --- | --- | --- | --- | --- | --- | --- |
| BR-REG-001 | The platform shall provide a public company-registration application containing company name, CR/business registration number, jurisdiction/country, company contact email, company phone, approved digital presence, first administrator name/email/phone/password, and selected Plan. | Applicant | Must | MVP | BO-001 | D-034,D-037 |
| BR-REG-002 | Registration shall create a pending application and shall not automatically activate an Organization or subscription. | System | Must | MVP | BO-004 | D-034,D-038 |
| BR-REG-003 | Super Admin shall review, request information, approve, or reject registration applications before activation. | Super Admin | Must | MVP | BO-007 | D-224 |
| BR-REG-004 | The same user email may be a member of multiple Organizations. | System | Must | MVP | BO-001 | D-041 |
| BR-REG-005 | CR/business registration number shall be unique within its normalized jurisdiction; duplicate registration shall be blocked. | System | Must | MVP | BO-005 | D-042 |
| BR-REG-006 | A detected existing company shall be directed to sign in or contact support instead of creating a second tenant. | System | Must | MVP | BO-005 | D-043 |
| BR-REG-007 | A duplicate company domain/digital identity shall block registration by default and require support resolution. | System | Must | MVP | BO-005 | D-044 |
| BR-REG-008 | Activation shall create or activate the Organization, the first Owner membership, the subscription snapshot, effective entitlements/limits, and Company Understanding status. | System/Super Admin | Must | MVP | BO-007 | D-038 |
| BR-REG-009 | Every Organization-owned record shall be tenant-scoped and server-side authorized. | System | Must | MVP | BO-005 | D-120,D-174 |
| BR-REG-010 | Organization suspension shall stop applicable access/external actions without deleting historical data. | Super Admin/System | Must | MVP | BO-007 | D-227 |
| BR-REG-011 | Super Admin shall be able to disable all external side effects for one Organization while preserving safe internal analysis/read access. | Super Admin | Must | MVP | BO-004 | D-228 |
| BR-REG-012 | Customer-requested Organization closure shall require a controlled Super Admin process. | Organization Owner/Super Admin | Must | MVP | BO-007 | D-256 |



### 10.2 Plans, Entitlements, and Usage


| Requirement ID | Business requirement | Primary actor/scope | Priority | Delivery | Objective | Decision trace |
| --- | --- | --- | --- | --- | --- | --- |
| BR-PLN-001 | Super Admin shall create, activate, deactivate, and assign Plans. | Super Admin | Must | MVP | BO-007 | D-035,D-225 |
| BR-PLN-002 | A Plan shall separately define commercial information, feature entitlements, limits, and usage semantics. | Super Admin/System | Must | MVP | BO-007 | D-229 |
| BR-PLN-003 | Each Plan limit shall support Numeric or Unlimited mode independently. | Super Admin/System | Must | MVP | BO-007 | D-237 |
| BR-PLN-004 | The Project Lead Limit shall apply per Project. | System | Must | MVP | BO-002 | D-230 |
| BR-PLN-005 | The Strategy Revision Allowance shall apply per Project and shall count only user-requested revisions after initial Strategy output. | System | Must | MVP | BO-001 | D-035A,D-231 |
| BR-PLN-006 | Understanding clarification, missing-information correction, and initial Strategy generation shall not consume Strategy Revision Allowance. | System | Must | MVP | BO-001 | D-046,D-051,D-157 |
| BR-PLN-007 | At zero remaining Strategy revisions, the latest Strategy shall remain approvable but shall not be regenerated. | System/User | Must | MVP | BO-004 | D-048 |
| BR-PLN-008 | Existing Project Leads shall continue operating when a Project reaches its Lead Limit. | System | Must | MVP | BO-003 | D-233 |
| BR-PLN-009 | Overflow Meta leads shall enter/match the Master Lead Pool but shall not enter autonomous Project sales beyond the effective limit. | System | Must | MVP | BO-005 | D-232,D-272 |
| BR-PLN-010 | Super Admin may apply Organization-specific feature/limit overrides without editing the underlying Plan. | Super Admin | Must | MVP | BO-007 | D-226 |
| BR-PLN-011 | Editing a Plan shall not silently alter existing Organization subscription snapshots. | System/Super Admin | Must | MVP | BO-004 | D-235 |
| BR-PLN-012 | Organization users shall be able to view, but not change, their assigned Plan, effective limits, entitlements, usage, and overrides. | Organization User | Must | MVP | BO-010 | D-251,D-252 |
| BR-PLN-013 | The platform shall warn authorized users when usage approaches or reaches a limit and explain the resulting behavior. | System | Must | MVP | BO-010 | D-253 |
| BR-PLN-014 | MVP subscription activation, renewal/status, and Plan changes shall be administered manually by Super Admin. | Super Admin | Must | MVP | BO-007 | D-236 |



### 10.3 Company Understanding


| Requirement ID | Business requirement | Primary actor/scope | Priority | Delivery | Objective | Decision trace |
| --- | --- | --- | --- | --- | --- | --- |
| BR-CMP-001 | After Organization activation, an authorized user shall complete Company Understanding before normal Project creation/use. | Organization Owner/Permitted User | Must | MVP | BO-001 | D-017,D-047 |
| BR-CMP-002 | Company Understanding inputs shall include existing registration data plus free text and/or company-profile documents supplied by the user. | Organization User | Must | MVP | BO-001 | D-029,D-045 |
| BR-CMP-003 | The Company Understanding Agent shall analyze the approved digital presence and provided sources and classify material information as Confirmed, Inferred, Missing/Unknown, or Conflicting. | Company Understanding Agent | Must | MVP | BO-001 | D-163 |
| BR-CMP-004 | The user shall review the initial understanding before the Agent asks targeted clarification questions. | Organization User/Agent | Must | MVP | BO-010 | D-045 |
| BR-CMP-005 | The Agent shall not ask for information already available in registration data or supplied sources unless a conflict requires clarification. | Agent | Must | MVP | BO-010 | D-045 |
| BR-CMP-006 | The user shall review and approve a final Business Overview before Company Understanding becomes authoritative. | Organization User | Must | MVP | BO-004 | D-039 |
| BR-CMP-007 | Company Understanding shall cover business identity, industry/categories, business model, geographies, broad capabilities/services, market background, and customer/business types. | Agent/System | Must | MVP | BO-001 | D-029 |
| BR-CMP-008 | Detailed Project products, prices, offers, audiences, and Project Strategy shall not be owned by Company Understanding. | System/Agent | Must | MVP | BO-001 | D-029 |
| BR-CMP-009 | Approved Company Understanding shall be versioned and auditable. | System | Must | MVP | BO-006 | D-040,D-166 |
| BR-CMP-010 | The platform shall separately create private Organization Intelligence and a sanitized reusable Business Pool proposal. | System/Agent | Must | MVP | BO-005 | D-036,D-039,D-164 |
| BR-CMP-011 | Private documents, internal business Strategy, Project data, and customer data shall not become reusable Business Pool data by default. | System | Must | MVP | BO-005 | D-036 |
| BR-CMP-012 | Project activity may generate a suggested Company Understanding update but shall not overwrite the approved profile automatically. | System/Agent | Must | MVP | BO-004 | D-165 |
| BR-CMP-013 | Material Company Understanding updates shall create a new version through authorized review/approval. | Organization User/Platform Admin | Must | MVP | BO-006 | D-040 |
| BR-CMP-014 | Company Understanding failures or missing information shall produce actionable retry/clarification states rather than fabricated output. | Agent/System | Must | MVP | BO-010 | D-177 |
| BR-CMP-015 | Completion shall route the user to the Organization Dashboard and the next action to create the first Project. | System | Must | MVP | BO-010 | D-279 |



### 10.4 Projects and Product Understanding


| Requirement ID | Business requirement | Primary actor/scope | Priority | Delivery | Objective | Decision trace |
| --- | --- | --- | --- | --- | --- | --- |
| BR-PRJ-001 | A Project shall be the primary commercial execution unit and shall have one primary B2B or B2C mode in MVP. | Organization User | Must | MVP | BO-008 | D-031 |
| BR-PRJ-002 | Project creation shall require a name and allow an optional short description. | Organization User | Must | MVP | BO-001 | D-049 |
| BR-PRJ-003 | The user shall provide a free-text commercial brief and may upload product lists, brochures, prices, specifications, prior sales/marketing material, and other relevant sources. | Organization User | Must | MVP | BO-001 | D-049 |
| BR-PRJ-004 | The Project Understanding Agent shall use approved Company Understanding only as relevant context and shall not access unrelated Project knowledge. | Agent/System | Must | MVP | BO-005 | D-022,D-158 |
| BR-PRJ-005 | The Agent shall dynamically identify missing material information and ask targeted questions rather than require one universal product form. | Agent | Must | MVP | BO-010 | D-160 |
| BR-PRJ-006 | The Agent shall produce Project Scope Understanding, inferred Project objective, Product Understanding per product/service, source evidence, conflicts, assumptions, and readiness. | Agent | Must | MVP | BO-001 | D-158 |
| BR-PRJ-007 | The user shall confirm what is being sold and the commercial outcome the Project seeks. | Organization User | Must | MVP | BO-004 | D-057 |
| BR-PRJ-008 | The user shall review, correct, and approve Project/Product Understanding before Strategy generation. | Organization User | Must | MVP | BO-004 | D-030,D-050 |
| BR-PRJ-009 | Authorized user corrections shall override AI inference in the approved baseline and preserve the original value in audit/history. | System/User | Must | MVP | BO-006 | D-161 |
| BR-PRJ-010 | The approved baseline shall distinguish strategy-critical facts from editable sales-support content. | Agent/System | Must | MVP | BO-001 | D-159 |
| BR-PRJ-011 | The Product Catalogue shall be generated from the approved Product Understanding. | System | Must | MVP | BO-001 | D-021,D-063 |
| BR-PRJ-012 | Users shall not add new core Project products after final onboarding approval. | System/User | Must | MVP | BO-004 | D-020 |
| BR-PRJ-013 | Users may add/edit approved descriptions, images, videos, brochures, and sales-support documents without changing Strategy. | Organization User | Must | MVP | BO-003 | D-064 |
| BR-PRJ-014 | The Sales AI may retrieve only content associated with the correct Project and product. | System/Agent | Must | MVP | BO-005 | D-065 |
| BR-PRJ-015 | Approved Project/Product Understanding shall be versioned and referenced by the Strategy version generated from it. | System | Must | MVP | BO-006 | D-162 |
| BR-PRJ-016 | Project lifecycle shall be ONBOARDING, LIVE, and ENDED. | System/User | Must | MVP | BO-010 | D-273 |
| BR-PRJ-017 | LIVE shall mean onboarding is complete and acquisition/sales operation is permitted subject to integrations, permissions, and limits. | System | Must | MVP | BO-002 | D-273 |
| BR-PRJ-018 | ENDED shall mean the customer intentionally closed the sales initiative; new acquisition/outreach shall stop while history and analytics remain. | Organization User/System | Must | MVP | BO-006 | D-273 |
| BR-PRJ-019 | Reopening an ENDED Project shall not be supported in MVP. | System | Must | MVP | BO-004 | D-281 |
| BR-PRJ-020 | Project knowledge and memory shall remain isolated from other Projects even within the same Organization unless an approved source is explicitly reused. | System | Must | MVP | BO-005 | D-022 |



### 10.5 Strategy


| Requirement ID | Business requirement | Primary actor/scope | Priority | Delivery | Objective | Decision trace |
| --- | --- | --- | --- | --- | --- | --- |
| BR-STR-001 | Strategy generation shall require approved Project/Product Understanding and explicit user initiation. | User/Strategy Agent | Must | MVP | BO-001 | D-052,D-153 |
| BR-STR-002 | If critical information is missing, Strategy generation shall be blocked and missing information requested without consuming revision allowance. | Agent/System | Must | MVP | BO-004 | D-157 |
| BR-STR-003 | The Strategy Agent shall produce one structured package containing executive summary, objective, market overview, positioning, ICPs, personas, pains/gains, product-to-audience fit, value propositions, sales Strategy, marketing Strategy, acquisition direction, qualification framework, conversion/handover model, campaign concepts, assumptions, risks, and learning hypotheses. | Strategy Agent | Must | MVP | BO-001 | D-053,D-155 |
| BR-STR-004 | For B2B, the Strategy shall separately model ideal customer organization and relevant user/influencer/buyer personas. | Strategy Agent | Must | MVP | BO-008 | D-154 |
| BR-STR-005 | For B2C, the Strategy shall use consumer-relevant personas and shall not require B2B attributes. | Strategy Agent | Must | MVP | BO-008 | D-278 |
| BR-STR-006 | The Strategy shall map product capabilities/benefits to specific persona pains, gains, decision criteria, and value propositions. | Strategy Agent | Must | MVP | BO-003 | D-053 |
| BR-STR-007 | The Strategy shall recommend acquisition direction—Web, Meta, or Both—without hard-coding by B2B/B2C classification. | Strategy Agent | Must | MVP | BO-002 | D-031 |
| BR-STR-008 | The Strategy shall define Project-specific initial qualification dimensions and proposed weights within platform bounds. | Strategy Agent/System | Must | MVP | BO-003 | D-083 |
| BR-STR-009 | The Strategy shall define the Project conversion goal and the MVP human handover boundary. | Strategy Agent/User | Must | MVP | BO-004 | D-202,D-209 |
| BR-STR-010 | The Strategy shall not invent prices, discounts, guarantees, capabilities, or contractual terms. | Strategy Agent | Must | MVP | BO-004 | D-129,D-271 |
| BR-STR-011 | The user shall review one coherent Strategy package and may approve or request revision through instructions. | Organization User | Must | MVP | BO-010 | D-058 |
| BR-STR-012 | Each requested revision shall create a new version and consume one allowance unit. | System/Agent | Must | MVP | BO-006 | D-055 |
| BR-STR-013 | A revision shall apply requested changes, re-evaluate dependent sections, and preserve unaffected content where still valid. | Strategy Agent | Must | MVP | BO-010 | D-156 |
| BR-STR-014 | The system shall retain previous Strategy versions and the revision instruction that produced each version. | System | Must | MVP | BO-006 | D-156 |
| BR-STR-015 | After Strategy acceptance, the user shall configure Project and channel communication settings before final onboarding approval. | Organization User | Must | MVP | BO-003 | D-060,D-061 |
| BR-STR-016 | Organization-wide mandatory communication restrictions shall be inherited and shall not be overridden by Project settings. | System | Must | MVP | BO-004 | D-254 |
| BR-STR-017 | Strategy approval shall not launch acquisition, publish campaigns, or create spend automatically. | System/User | Must | MVP | BO-004 | D-059,D-147 |
| BR-STR-018 | Project-level Strategy improvement recommendations shall require evidence, user review, and a new approved version. | Agent/User/System | Must | MVP | BO-009 | D-215 |



### 10.6 Lead Acquisition and Discovery


| Requirement ID | Business requirement | Primary actor/scope | Priority | Delivery | Objective | Decision trace |
| --- | --- | --- | --- | --- | --- | --- |
| BR-ACQ-001 | The MVP shall support real Meta Ads, WhatsApp, Instagram messaging, Facebook Messenger, Web/Google discovery, and generic email outreach. | Platform/Organization | Must | MVP | BO-002 | D-260 |
| BR-ACQ-002 | The Lead Discovery Agent shall evaluate audience availability, ICP, persona, product, geography, objective, and supported capabilities to recommend Web, Meta, or Both. | Lead Discovery Agent | Must | MVP | BO-002 | D-141 |
| BR-ACQ-003 | The user shall review the Lead Acquisition Plan before external Web discovery or campaign execution. | Organization User | Must | MVP | BO-004 | D-073 |
| BR-ACQ-004 | Web Discovery plans shall contain target-company/person criteria, practical buyer-influence roles, geography, exclusions, contact requirements, source sequence, and query families. | Lead Discovery Agent | Must | MVP | BO-002 | D-072,D-139 |
| BR-ACQ-005 | The default Web candidate target shall be effective Project Lead Limit multiplied by a configurable platform multiplier, initially 10. | System/Lead Discovery Agent | Must | MVP | BO-002 | D-078,D-137 |
| BR-ACQ-006 | The system shall execute fresh Web discovery first and supplement with eligible Master Lead Pool matches only when needed under current policy. | System/Lead Discovery Agent | Must | MVP | BO-005 | D-079,D-138 |
| BR-ACQ-007 | A Research Candidate without usable contact information shall not become a Master Lead until a contactable identity is found. | System | Must | MVP | BO-005 | D-080 |
| BR-ACQ-008 | Every captured contactable person from every approved source shall create or match a Master Lead record. | System | Must | MVP | BO-005 | D-071 |
| BR-ACQ-009 | Every captured lead shall retain acquisition source, provider reference/URL, query/discovery run, timestamp, confidence, and contact provenance. | System | Must | MVP | BO-006 | D-140 |
| BR-ACQ-010 | B2B Web discovery shall identify matching businesses and practical users/influencers able to advance purchase rather than default to executive titles. | Lead Discovery Agent | Must | MVP | BO-008 | D-081 |
| BR-ACQ-011 | Discovery shall stop at target, search exhaustion, provider limit, cost/usage limit, excessive duplication/irrelevance, failure, or manual stop and shall record the stop reason. | System/Agent | Must | MVP | BO-007 | D-082 |
| BR-ACQ-012 | A Master Lead Pool match shall never expose another tenant's private conversation or Project sales intelligence. | System | Must | MVP | BO-005 | D-036,D-189 |
| BR-ACQ-013 | Meta inbound leads shall be treated as having an explicit source-intent signal. | System/Lead Strategist | Must | MVP | BO-003 | D-086 |
| BR-ACQ-014 | The primary paid Meta conversation destination in MVP shall be WhatsApp. | Campaign Agent/System | Must | MVP | BO-002 | D-266 |
| BR-ACQ-015 | Direct Instagram/Facebook/WhatsApp inbound messages shall be handled through the same Project Lead and Sales AI controls where eligible. | System/Sales Agent | Must | MVP | BO-003 | D-260,D-262 |
| BR-ACQ-016 | A recommended channel that is not configured shall not block Project onboarding but shall block execution and create an Action Item. | System | Must | MVP | BO-010 | D-268 |
| BR-ACQ-017 | If a live channel fails, the system may transition to another approved, configured, healthy, contactable, and eligible channel when strategically appropriate. | System/Lead Strategist | Must | MVP | BO-003 | D-267,D-269 |
| BR-ACQ-018 | If no valid channel fallback exists, affected activity shall pause and authorized users shall be alerted. | System | Must | MVP | BO-004 | D-270 |
| BR-ACQ-019 | Channel transitions shall preserve conversation context and shall never invent contact data. | System/Agent | Must | MVP | BO-006 | D-267 |
| BR-ACQ-020 | Acquisition results beyond the active Project Lead allowance shall remain available in Master Pool/source analytics but shall not receive autonomous Project sales. | System | Must | MVP | BO-005 | D-272 |



### 10.7 Meta Campaigns


| Requirement ID | Business requirement | Primary actor/scope | Priority | Delivery | Objective | Decision trace |
| --- | --- | --- | --- | --- | --- | --- |
| BR-CAM-001 | The Campaign Agent shall convert an approved Meta acquisition brief into a structured campaign proposal. | Campaign Agent | Must | MVP | BO-002 | D-144 |
| BR-CAM-002 | The proposal shall include campaign identity/objective, product, target audience, geography, channels/placements, budget, mandatory end date, lead path, copy direction, creative brief, and risks. | Campaign Agent | Must | MVP | BO-002 | D-070,D-144 |
| BR-CAM-003 | The Campaign Agent shall recommend the lead form/path and qualification questions; an authorized human shall approve them. | Campaign Agent/User | Must | MVP | BO-004 | D-148 |
| BR-CAM-004 | The Campaign Agent shall provide detailed step-by-step image/video production instructions appropriate to the campaign and selected placement. | Campaign Agent | Must | MVP | BO-010 | D-151 |
| BR-CAM-005 | The user shall create/upload final media assets for MVP. | Organization User | Must | MVP | BO-010 | D-145 |
| BR-CAM-006 | The Campaign Agent shall review uploaded media against Product Knowledge, Strategy, approved claims, and creative brief and shall flag issues. | Campaign Agent | Must | MVP | BO-004 | D-152 |
| BR-CAM-007 | The user shall be able to review and edit campaign settings and copy within permitted boundaries before publication. | Organization User | Must | MVP | BO-010 | D-144 |
| BR-CAM-008 | Campaign publication shall require an explicit action by an authorized user. | Organization User | Must | MVP | BO-004 | D-068,D-147 |
| BR-CAM-009 | Before publication, the platform shall perform AI commercial-quality review and deterministic permission, entitlement, spend, date, media, integration, and provider-readiness validation. | System/Agent | Must | MVP | BO-004 | D-149 |
| BR-CAM-010 | The Campaign Service, not the Agent, shall call the Meta provider adapter/API. | System | Must | MVP | BO-004 | D-120,D-144 |
| BR-CAM-011 | Campaign publication and updates shall be idempotent and auditable. | System | Must | MVP | BO-006 | D-147 |
| BR-CAM-012 | Authorized users shall be able to increase budget, extend end date, pause, and resume an active campaign. | Organization User | Must | MVP | BO-004 | D-070 |
| BR-CAM-013 | The Campaign Agent may recommend budget, extension, pause, or replacement actions but shall not execute them independently. | Campaign Agent | Must | MVP | BO-004 | D-146 |
| BR-CAM-014 | Published campaign pages shall show configuration, status, performance, generated leads, downstream qualification, handovers, and conversions as available. | System | Must | MVP | BO-006 | D-069,D-211 |
| BR-CAM-015 | Campaign provider failures shall create an actionable error, retain provider details, and allow correction/retry without duplicate publication. | System/Agent | Must | MVP | BO-010 | D-149 |
| BR-CAM-016 | Ended campaigns shall remain historical, shall not restart in place, and may be duplicated into a new Draft. | System/User | Must | MVP | BO-006 | D-275 |
| BR-CAM-017 | Existing lead conversations shall continue after the source campaign ends. | System | Must | MVP | BO-003 | D-275 |
| BR-CAM-018 | The exact campaign state machine shall be confirmed before implementation; the current recommended model is DRAFT/READY/PUBLISHING/ACTIVE/PAUSED/ENDED with PUBLISH_FAILED. | Project Owner | Must | MVP | BO-010 | D-274 |



### 10.8 Email and Messaging Channels


| Requirement ID | Business requirement | Primary actor/scope | Priority | Delivery | Objective | Decision trace |
| --- | --- | --- | --- | --- | --- | --- |
| BR-EML-001 | Organization users with permission shall connect a generic business mailbox supporting outbound sending and inbound reply retrieval. | Organization User/System | Must | MVP | BO-002 | D-261 |
| BR-EML-002 | The Organization shall configure an approved Sales AI display identity, mailbox, and signature; AI shall not invent identity. | Organization User | Must | MVP | BO-004 | D-263 |
| BR-EML-003 | The Sales AI shall automatically generate and send a personalized first email to approved eligible web leads. | Sales Agent/System | Must | MVP | BO-003 | D-101,D-262 |
| BR-EML-004 | Inbound email replies shall be ingested into the correct Project Lead thread and handled as active conversation context. | System/Sales Agent | Must | MVP | BO-003 | D-262 |
| BR-EML-005 | Normal active email replies shall be autonomous within approved Project/channel rules. | Sales Agent | Must | MVP | BO-003 | D-102,D-262 |
| BR-EML-006 | Email follow-ups outside active conversation shall require human approval unless follow-up control was explicitly delegated for that Project Lead. | System/User | Must | MVP | BO-004 | D-100,D-109 |
| BR-EML-007 | Email bounces/rejections shall be recorded, update channel quality, create an Action Item/notification, and prevent blind repeated sending. | System | Must | MVP | BO-004 | D-264 |
| BR-EML-008 | Sales AI may send approved product attachments/content when requested or strategically relevant. | Sales Agent | Must | MVP | BO-003 | D-265 |
| BR-EML-009 | The same active/inactive conversation, follow-up approval, takeover, return-to-AI, suppression, and audit rules shall apply across email, WhatsApp, Instagram, and Facebook messaging. | System/Agent | Must | MVP | BO-004 | D-102,D-104 |
| BR-EML-010 | Provider connection status shall expose Not Configured, Configuring, Connected, Error, Disabled, and Reauthorization Required or equivalent states. | System | Must | MVP | BO-010 | D-220 |
| BR-EML-011 | Platform-managed Organization Meta/WhatsApp/Instagram/Facebook secrets shall not be exposed to Organization users. | System/Super Admin | Must | MVP | BO-007 | D-220 |
| BR-EML-012 | All outbound communications shall be checked for permission, Project access, entitlement, integration readiness, eligibility, suppression, control mode, and policy before provider execution. | System | Must | MVP | BO-004 | D-120,D-179 |



### 10.9 Business Pool and Master Lead Pool


| Requirement ID | Business requirement | Primary actor/scope | Priority | Delivery | Objective | Decision trace |
| --- | --- | --- | --- | --- | --- | --- |
| BR-POOL-001 | The Business Pool shall store companies/organizations; the Master Lead Pool shall store people/consumer identities. | System | Must | MVP | BO-005 | D-183,D-184 |
| BR-POOL-002 | A Master Lead may link to multiple current/historical Business Pool entities with role, relationship type, dates, source, and confidence. | System | Must | MVP | BO-005 | D-185 |
| BR-POOL-003 | A registered tenant Organization may link to a Business Pool entity without exposing tenant-private data or activity. | System | Must | MVP | BO-005 | D-186,D-198 |
| BR-POOL-004 | Master Lead reusable attributes shall retain source, confidence, freshness, observed/inferred status, and reuse eligibility. | System/Agent | Must | MVP | BO-005 | D-187 |
| BR-POOL-005 | Historical product/category interest or rejection signals shall retain context and shall not be generalized beyond evidence. | System/Agent | Must | MVP | BO-005 | D-188 |
| BR-POOL-006 | Project-specific qualification, detailed objections, conversation, budget, negotiation, Lead Strategy, and outcome shall remain tenant/project private. | System | Must | MVP | BO-005 | D-189 |
| BR-POOL-007 | Master Lead matching shall use normalized multiple identifiers and confidence levels rather than name alone. | System/Agent | Must | MVP | BO-005 | D-190 |
| BR-POOL-008 | AI may recommend identity matches; deterministic identity services shall merge records. | System | Must | MVP | BO-005 | D-191 |
| BR-POOL-009 | Lead and company merges shall preserve all source history and conflicting values. | System | Must | MVP | BO-006 | D-192,D-193 |
| BR-POOL-010 | Authorized Platform Administration shall be able to reverse audited incorrect merges. | Super Admin | Must | MVP | BO-007 | D-194,D-200 |
| BR-POOL-011 | CR plus jurisdiction shall be the strongest company identity where available. | System | Must | MVP | BO-005 | D-195 |
| BR-POOL-012 | Company domain shall be a strong but non-absolute match and shall not collapse legitimate subsidiaries/brands/branches. | System | Must | MVP | BO-005 | D-196 |
| BR-POOL-013 | The Business Pool shall support parent, subsidiary, branch, brand, franchise, and related-entity relationships. | System | Must | MVP | BO-005 | D-197,D-201 |
| BR-POOL-014 | Reusable Business Pool enrichment shall retain evidence, confidence, freshness, and observed/inferred status. | System/Agent | Must | MVP | BO-005 | D-199 |
| BR-POOL-015 | Correction, suppression, reuse eligibility, and future deletion/retention controls shall be represented for reusable profiles. | System | Must | MVP | BO-005 | D-187 |



### 10.10 Enrichment and Qualification


| Requirement ID | Business requirement | Primary actor/scope | Priority | Delivery | Objective | Decision trace |
| --- | --- | --- | --- | --- | --- | --- |
| BR-QLF-001 | The Lead Enrichment Agent shall separately produce reusable Master Lead enrichment and Project-specific enrichment. | Lead Enrichment Agent | Must | MVP | BO-005 | D-132 |
| BR-QLF-002 | Enrichment shall be purpose-limited to identity, contactability, ICP/product fit, qualification, and sales Strategy. | Lead Enrichment Agent | Must | MVP | BO-004 | D-135 |
| BR-QLF-003 | Every material enrichment fact shall retain provenance, freshness, confidence, and observed/inferred status. | System/Agent | Must | MVP | BO-006 | D-134 |
| BR-QLF-004 | Missing information shall remain Unknown rather than be fabricated. | Agent | Must | MVP | BO-004 | D-129,D-134 |
| BR-QLF-005 | Lead Strategist may request targeted re-enrichment when missing external evidence materially affects sales Strategy. | Lead Strategist/Enrichment Agent | Must | MVP | BO-003 | D-136 |
| BR-QLF-006 | The platform shall calculate/store an initial Project-specific score before Web outreach using Strategy-defined dimensions validated within platform bounds. | System/Lead Strategist | Must | MVP | BO-003 | D-083 |
| BR-QLF-007 | The platform shall rank contactable Web candidates and allocate the highest-ranked candidates up to the Project Lead Limit. | System | Must | MVP | BO-002 | D-084 |
| BR-QLF-008 | Meta inbound leads shall start with an explicit intent signal and shall enter active qualification/conversation rather than cold outreach. | System/Lead Strategist | Must | MVP | BO-003 | D-086 |
| BR-QLF-009 | Dynamic qualification shall use observed replies, questions, objections, product preference, timing, intent, and progress evidence and shall progressively supersede uncertain initial assumptions. | Lead Strategist/System | Must | MVP | BO-003 | D-085 |
| BR-QLF-010 | A Qualified lead shall have evidence of product fit, genuine need/interest, and a realistic path to progress. | Lead Strategist/System | Must | MVP | BO-003 | D-088 |
| BR-QLF-011 | Qualification outcomes shall distinguish Qualified, Nurture, Not Qualified, Not Interested, and Suppressed and record structured reasons. | System/Lead Strategist | Must | MVP | BO-006 | D-089 |
| BR-QLF-012 | Not Qualified and Not Interested shall apply only to the Project Lead and shall not invalidate the Master Lead globally. | System | Must | MVP | BO-005 | D-092,D-093 |
| BR-QLF-013 | Suppression shall be a protected state that blocks applicable outbound communication regardless of Agent recommendation. | System | Must | MVP | BO-004 | D-094 |
| BR-QLF-014 | No Response shall be treated as evidence/attention condition and shall not automatically mark a lead Not Interested or final. | System/Lead Strategist | Must | MVP | BO-003 | D-096 |
| BR-QLF-015 | Every active Project Lead shall have exactly one current Next Best Action or explicit Wait/No Action with reason. | Lead Strategist/System | Must | MVP | BO-003 | D-115 |
| BR-QLF-016 | Project-specific scoring and qualification shall never require B2B attributes for a B2C Project. | System/Agent | Must | MVP | BO-008 | D-278 |
| BR-QLF-017 | Initial and dynamic score/state history shall be retained for audit and learning. | System | Must | MVP | BO-006 | D-085 |



### 10.11 Sales Conversation and Control


| Requirement ID | Business requirement | Primary actor/scope | Priority | Delivery | Objective | Decision trace |
| --- | --- | --- | --- | --- | --- | --- |
| BR-SAL-001 | Each Project Lead shall have a structured evolving Lead Strategy maintained by a dedicated Lead Strategist Agent. | Lead Strategist | Must | MVP | BO-003 | D-114,D-119 |
| BR-SAL-002 | Lead Strategy shall contain updated understanding, qualification/score recommendation, hypothesis, objective, approach, objections, approaches tried, missing information, Next Best Action, follow-up/handover recommendation, and user summary. | Lead Strategist | Must | MVP | BO-003 | D-126 |
| BR-SAL-003 | Lead Strategy shall update only on material new evidence or authorized human instruction. | Lead Strategist/System | Must | MVP | BO-003 | D-116,D-124 |
| BR-SAL-004 | The Sales Agent shall execute the current commercial approach through natural channel-appropriate communication and shall not independently rewrite Strategy. | Sales Agent | Must | MVP | BO-003 | D-117 |
| BR-SAL-005 | Sales AI communication shall be contextual, professional, adaptive, non-repetitive, multilingual within configured languages, and grounded in conversation memory. | Sales Agent | Must | MVP | BO-003 | D-097,D-128 |
| BR-SAL-006 | Qualification questions shall be integrated naturally into conversation rather than delivered as a rigid interrogation. | Sales Agent | Must | MVP | BO-003 | D-130 |
| BR-SAL-007 | Sales AI shall only make product/commercial claims supported by approved Project knowledge. | Sales Agent/System | Must | MVP | BO-004 | D-129 |
| BR-SAL-008 | The complete conversation history, messages, attachments, extracted sales memory, summary, commitments, and current Strategy shall be application-persisted. | System | Must | MVP | BO-006 | D-113 |
| BR-SAL-009 | Initial outreach to an approved eligible Web lead shall be generated and sent automatically within Project/channel rules. | Sales Agent/System | Must | MVP | BO-003 | D-101 |
| BR-SAL-010 | Meta inbound and normal active-conversation replies shall be autonomous within approved boundaries. | Sales Agent/System | Must | MVP | BO-003 | D-102,D-104 |
| BR-SAL-011 | The active/inactive determination shall use channel context, conversation context, explicit lead statements, and configured rules rather than one universal timeout. | System/Lead Strategist | Must | MVP | BO-003 | D-105 |
| BR-SAL-012 | Normal follow-ups outside an active conversation shall require human approval. | System/User | Must | MVP | BO-004 | D-100 |
| BR-SAL-013 | A user may explicitly delegate follow-up execution for one Project Lead to AI and may revoke it. | User/System | Must | MVP | BO-004 | D-109 |
| BR-SAL-014 | Follow-ups shall be adaptive, proposed one at a time, and include timing, channel, objective, message, reason, and approval state. | Lead Strategist/Sales Agent | Must | MVP | BO-003 | D-098,D-108 |
| BR-SAL-015 | An approved scheduled follow-up shall be revalidated for reply, suppression, status, Project state, channel availability, and relevance before send. | System | Must | MVP | BO-004 | D-107 |
| BR-SAL-016 | Any new inbound lead message shall reopen active conversation and cancel/obsolete conflicting unsent follow-ups. | System | Must | MVP | BO-003 | D-106 |
| BR-SAL-017 | Material unresponsiveness shall create a visible alert with AI assessment and recommended approach. | System/Lead Strategist | Must | MVP | BO-010 | D-110 |
| BR-SAL-018 | Authorized users shall be able to monitor any accessible conversation without interrupting AI. | Organization User | Must | MVP | BO-010 | D-111 |
| BR-SAL-019 | An authorized user may take over an AI-controlled conversation at any time; AI sending shall stop immediately. | Organization User/System | Must | MVP | BO-004 | D-111 |
| BR-SAL-020 | An authorized user may return conversation control to AI only after AI reassesses current history, human messages, state, commitments, and Strategy. | Organization User/System/Agent | Must | MVP | BO-004 | D-112 |
| BR-SAL-021 | The Sales Workspace shall distinguish lead messages, AI messages, human messages, internal AI instructions, system events, and control-state changes. | System/UI | Must | MVP | BO-006 | D-280 |
| BR-SAL-022 | Users shall be able to Ask AI, Instruct AI, Take Over, Return to AI, and approve/edit/reject follow-up proposals according to permission/control mode. | Organization User/UI | Must | MVP | BO-010 | D-118,D-280 |
| BR-SAL-023 | Organization/Project mandatory restrictions, suppression, permissions, and approved knowledge shall override lead-specific human instructions. | System | Must | MVP | BO-004 | D-118,D-181 |
| BR-SAL-024 | Active conversation and follow-up events shall be audited and attributed to AI/human actor and version. | System | Must | MVP | BO-006 | D-178,D-258 |
| BR-SAL-025 | The Unified Sales Workspace shall provide a lead queue, conversation timeline, AI Sales Insight, control actions, and lead detail context. | UI/System | Must | MVP | BO-010 | D-280 |



### 10.12 Handover, Opportunity, and Conversion


| Requirement ID | Business requirement | Primary actor/scope | Priority | Delivery | Objective | Decision trace |
| --- | --- | --- | --- | --- | --- | --- |
| BR-CON-001 | All final MVP commercial actions shall require human takeover. | System/Lead Strategist/User | Must | MVP | BO-004 | D-209 |
| BR-CON-002 | Handover shall be triggered by unsupported/protected final action, explicit human request, insufficient reliable knowledge, policy requirement, or authorized manual takeover. | System/Agent/User | Must | MVP | BO-004 | D-032,D-103 |
| BR-CON-003 | A handover shall contain lead, channel, state/score, reason, current request, last message, products, objections, commitments, and recommended human action. | System/Lead Strategist | Must | MVP | BO-010 | D-095 |
| BR-CON-004 | Handover-ready items shall be visible to all authorized users with Project access; the first valid claimant becomes assigned human owner. | System/User | Must | MVP | BO-004 | D-276 |
| BR-CON-005 | The claim operation shall prevent simultaneous conflicting ownership. | System | Must | MVP | BO-004 | D-276 |
| BR-CON-006 | Opportunity creation shall be optional and used only when a continuing deal record is needed. | System/User | Must | MVP | BO-010 | D-204 |
| BR-CON-007 | Lead Strategist may recommend an Opportunity; deterministic Project rules/application services shall create it. | System/Agent | Must | MVP | BO-004 | D-205 |
| BR-CON-008 | The human owner shall record the final outcome through a concise workflow. | Organization User | Must | MVP | BO-006 | D-207 |
| BR-CON-009 | Conversion shall be an auditable event containing type, time, Project Lead, product, source/campaign/discovery lineage, evidence, human actor, optional value, and outcome. | System/User | Must | MVP | BO-006 | D-206,D-208 |
| BR-CON-010 | The Master Lead shall not be globally classified by a Project conversion/loss outcome. | System | Must | MVP | BO-005 | D-188,D-189 |
| BR-CON-011 | The platform shall preserve AI contribution, handover time, human response time, and final human outcome. | System | Must | MVP | BO-006 | D-213 |
| BR-CON-012 | Conversion records shall update Project/campaign/source analytics and learning evidence. | System | Must | MVP | BO-006 | D-210,D-215 |



### 10.13 Analytics and Learning


| Requirement ID | Business requirement | Primary actor/scope | Priority | Delivery | Objective | Decision trace |
| --- | --- | --- | --- | --- | --- | --- |
| BR-ANL-001 | Project analytics shall track leads acquired, selected, contacted, engaged, qualifying, qualified, handed over, and human-converted. | System | Must | MVP | BO-006 | D-210 |
| BR-ANL-002 | Campaign and source analytics shall connect volume/cost to downstream qualification, handover, and conversion. | System | Must | MVP | BO-006 | D-211 |
| BR-ANL-003 | Web discovery analytics shall include candidate target, candidates found, contactable leads, selected leads, duplicates, and stop reason. | System | Must | MVP | BO-006 | D-140 |
| BR-ANL-004 | AI performance shall be measured through progression, response, qualification, objection handling, grounding, follow-up, handover quality, and conversion preparation—not message count alone. | System | Must | MVP | BO-009 | D-212 |
| BR-ANL-005 | Human handover acceptance, response time, final action, and outcome shall be measured separately. | System | Must | MVP | BO-006 | D-213 |
| BR-ANL-006 | Lead-level Strategy adaptation shall occur automatically on material evidence. | Lead Strategist | Must | MVP | BO-003 | D-214 |
| BR-ANL-007 | Project-level learning shall produce evidence-backed recommendations rather than automatic Strategy changes. | System/Agent | Must | MVP | BO-009 | D-214,D-215 |
| BR-ANL-008 | Project-level recommendations shall consider pattern consistency, sample size, confidence, persona, product, source, and attribution. | System/Agent | Must | MVP | BO-009 | D-216 |
| BR-ANL-009 | Accepted recommendations shall create a new Strategy version through controlled approval. | System/User | Must | MVP | BO-009 | D-215 |
| BR-ANL-010 | Future platform-level cross-Project learning shall remain disabled/deferred until governance and legal/privacy rules are approved. | System | Must | MVP | BO-005 | D-214 |
| BR-ANL-011 | Organization and Project dashboards shall present concise operational statistics and actions requiring attention. | UI/System | Must | MVP | BO-010 | D-062,D-279 |
| BR-ANL-012 | Project Dashboard actions shall deep-link to Product Catalogue, campaign setup/publication, leads, conversations, handovers, and other relevant workflows. | UI/System | Must | MVP | BO-010 | D-062,D-245 |
| BR-ANL-013 | Super Admin shall be able to observe aggregated and permitted end-to-end onboarding, acquisition, agent, conversation, handover, and conversion results. | Super Admin | Must | MVP | BO-007 | D-277 |
| BR-ANL-014 | Analytics shall preserve Project/source context and shall not use Master Lead as the only reporting dimension. | System | Must | MVP | BO-006 | D-188 |



### 10.14 Administration and Permissions


| Requirement ID | Business requirement | Primary actor/scope | Priority | Delivery | Objective | Decision trace |
| --- | --- | --- | --- | --- | --- | --- |
| BR-ADM-001 | The platform shall provide separate Super Admin CP and Organization Administration scopes. | System | Must | MVP | BO-007 | D-217 |
| BR-ADM-002 | Platform Super Admin and Platform Admin/Support shall have separate configurable privilege sets. | System | Must | MVP | BO-007 | D-218 |
| BR-ADM-003 | Each Organization shall have an Owner with protected account-level authority. | System | Must | MVP | BO-004 | D-243 |
| BR-ADM-004 | Other Organization users shall receive explicit permissions and explicit Project access; fixed functional roles shall not be the authorization source. | Organization Owner/System | Must | MVP | BO-004 | D-241 |
| BR-ADM-005 | Permission presets may exist only as setup shortcuts. | Organization Owner/UI | Should | MVP | BO-010 | D-242 |
| BR-ADM-006 | Authorization shall evaluate authentication, Organization membership, explicit permission, Project access, entitlement/readiness, resource ownership, action policy, and state. | System | Must | MVP | BO-004 | D-238,D-239 |
| BR-ADM-007 | Organization Owner shall grant/revoke permitted user permissions and Project access; users shall never grant themselves access. | Organization Owner/System | Must | MVP | BO-004 | D-243 |
| BR-ADM-008 | Platform-managed Organization integrations shall be configured under Super Admin CP → Organization → Integrations. | Super Admin | Must | MVP | BO-007 | D-220 |
| BR-ADM-009 | Super Admin shall configure Meta Ads, WhatsApp, Instagram, and Facebook capabilities separately per Organization. | Super Admin | Must | MVP | BO-007 | D-220 |
| BR-ADM-010 | Platform-wide Google/Search and AI/provider credentials shall be managed as platform integrations. | Super Admin | Must | MVP | BO-007 | D-220 |
| BR-ADM-011 | Email shall be Organization self-managed through Organization Administration. | Organization User | Must | MVP | BO-002 | D-220,D-261 |
| BR-ADM-012 | Super Admin shall configure per-agent model/provider, credential reference, prompt/version, runtime settings, tools, status, and evaluation state. | Super Admin | Must | MVP | BO-009 | D-221 |
| BR-ADM-013 | Super Admin CP shall show agent runs, tool calls, specialist invocations, status, errors, latency, usage, and versions. | Super Admin | Must | MVP | BO-009 | D-222 |
| BR-ADM-014 | Access to full prompt/response/customer content in agent logs shall be restricted and audited separately from operational telemetry. | System/Super Admin | Must | MVP | BO-005 | D-223 |
| BR-ADM-015 | Support access to a tenant shall be purpose-limited, time-limited where practical, permission-controlled, and audited. | Platform Support/System | Must | MVP | BO-005 | D-219 |
| BR-ADM-016 | Organization Administration shall provide Company, Team/Access, Projects, Email integration, Action Center, Usage/Plan, Settings, and Audit. | Organization User | Must | MVP | BO-010 | D-217 |
| BR-ADM-017 | Organization communication defaults and mandatory restrictions shall be configurable and inherited by Projects. | Organization Owner/Permitted User | Must | MVP | BO-004 | D-254 |
| BR-ADM-018 | Organization Admin CP shall expose tenant-scoped important actions while preserving historical actor information after user removal. | System | Must | MVP | BO-006 | D-257,D-259 |
| BR-ADM-019 | Super Admin shall manage Organization registration, Plan, usage/overrides, integrations, suspension, closure, support, and audit from one Organization detail view. | Super Admin | Must | MVP | BO-007 | D-224-D-228 |
| BR-ADM-020 | Organization users shall not access platform-global agents, provider secrets, other Organizations, or platform configuration. | System | Must | MVP | BO-005 | D-217 |



### 10.15 AI Agents and Governance


| Requirement ID | Business requirement | Primary actor/scope | Priority | Delivery | Objective | Decision trace |
| --- | --- | --- | --- | --- | --- | --- |
| BR-AGT-001 | Google ADK shall be the approved working Agent Component for major reasoning/orchestration capabilities. | Platform | Must | MVP | BO-009 | D-123 |
| BR-AGT-002 | The application shall remain authoritative for tenancy, permissions, limits, suppression, protected state, audit, and provider side effects. | System | Must | MVP | BO-004 | D-120,D-172 |
| BR-AGT-003 | The MVP agent registry shall include Company Understanding, Project Understanding, Strategy, Lead Discovery, Lead Enrichment, Lead Strategist, Sales/Conversation, Campaign, Customer Simulator, and Evaluator. | Super Admin/System | Must | MVP | BO-009 | D-121 |
| BR-AGT-004 | Agents shall exchange defined structured contracts and shall not use free-form inter-agent conversation as authoritative state. | System/Agents | Must | MVP | BO-009 | D-122,D-173 |
| BR-AGT-005 | Every agent context shall be explicitly assembled and tenant/Project/lead scoped. | System | Must | MVP | BO-005 | D-174 |
| BR-AGT-006 | Agent-to-agent invocation and tool access shall be deny-by-default and controlled by an interaction/tool registry. | System/Super Admin | Must | MVP | BO-004 | D-179,D-180 |
| BR-AGT-007 | Multi-agent workflows shall enforce retry, loop, timeout, and cost/usage limits. | System | Must | MVP | BO-007 | D-176 |
| BR-AGT-008 | Agents shall return normalized Success, Retryable Failure, Needs Information, Needs Human Review, Blocked, or Failed outcomes. | Agents/System | Must | MVP | BO-009 | D-177 |
| BR-AGT-009 | Every agent run shall record exact agent version, prompt version, model/provider configuration, timestamps, context references, tools, and result status. | System | Must | MVP | BO-006 | D-178 |
| BR-AGT-010 | Agent configuration shall separate editable behavior from non-editable platform safety and authorization controls. | System/Super Admin | Must | MVP | BO-004 | D-181 |
| BR-AGT-011 | Agent versions shall move through controlled Draft, Testing, Under Review, Approved, Active, and Retired states with rollback. | Super Admin/System | Must | MVP | BO-009 | D-182 |
| BR-AGT-012 | Customer Simulator and Evaluator shall be separate evaluation roles using sandboxed data/tools. | System | Must | MVP | BO-009 | D-167 |
| BR-AGT-013 | Major agent/prompt/model changes shall require evaluation before production activation. | Super Admin/System | Must | MVP | BO-009 | D-168 |
| BR-AGT-014 | Evaluation shall hard-fail critical violations such as invented price, suppression breach, tenant leak, unauthorized spend/discount, or messaging after takeover. | Evaluator/System | Must | MVP | BO-004 | D-169 |
| BR-AGT-015 | Test Lab shall support version comparison and versioned regression scenarios. | Super Admin | Must | MVP | BO-009 | D-170,D-171 |
| BR-AGT-016 | Agents shall not directly access provider credentials or protected database interfaces outside governed tools/services. | System/Agents | Must | MVP | BO-004 | D-179 |
| BR-AGT-017 | Lead Strategist shall not send messages; Sales Agent shall not autonomously rewrite Lead Strategy. | Agents | Must | MVP | BO-009 | D-117,D-119 |
| BR-AGT-018 | Campaign Agent shall not publish or modify spend without an explicit authorized application action. | Campaign Agent/System | Must | MVP | BO-004 | D-147 |
| BR-AGT-019 | Agent-generated recommendations that affect product state shall pass deterministic validation before persistence/execution. | System | Must | MVP | BO-004 | D-172 |
| BR-AGT-020 | Historical agent runs shall remain linked to the versions used and shall not be rewritten by later configuration changes. | System | Must | MVP | BO-006 | D-178 |



### 10.16 Action Center, Notifications, Audit, and UX


| Requirement ID | Business requirement | Primary actor/scope | Priority | Delivery | Objective | Decision trace |
| --- | --- | --- | --- | --- | --- | --- |
| BR-UX-001 | The Organization Dashboard shall be action-led and guide a new customer from first Project creation to Product Catalogue completion and lead acquisition. | UI/System | Must | MVP | BO-010 | D-279 |
| BR-UX-002 | The Project Dashboard shall show Project status, products, Strategy, acquisition, lead/conversation/conversion statistics, and action-required items. | UI/System | Must | MVP | BO-010 | D-062 |
| BR-UX-003 | The Action Center shall be the authoritative permission-aware queue of unresolved operational work. | System/UI | Must | MVP | BO-010 | D-244 |
| BR-UX-004 | Action Items shall deep-link to the exact work screen and support Open, In Progress, Completed, and Dismissed or equivalent statuses. | UI/System | Must | MVP | BO-010 | D-245 |
| BR-UX-005 | Agents may recommend actions; application services shall create, deduplicate, assign, complete, and audit Action Items. | System | Must | MVP | BO-006 | D-246 |
| BR-UX-006 | MVP notifications shall be in-app and email and shall be routed by permission and Project access. | System | Must | MVP | BO-010 | D-248,D-249 |
| BR-UX-007 | Dismissing a notification shall not resolve the underlying Action Item. | System/UI | Must | MVP | BO-010 | D-247 |
| BR-UX-008 | Email notifications shall contain minimum necessary information and direct users to authenticated application content. | System | Must | MVP | BO-005 | D-250 |
| BR-UX-009 | Users may configure non-critical operational email preferences; mandatory account/security notifications remain enabled. | Organization User/System | Must | MVP | BO-010 | D-255 |
| BR-UX-010 | The application shall create audit events for important approvals, state changes, publication, spend changes, follow-up delegation, takeover, permissions, integrations, conversion, support access, and agent configuration. | System | Must | MVP | BO-006 | D-257,D-258 |
| BR-UX-011 | The first prototype shall cover company registration/onboarding through Project onboarding, acquisition setup, leads, and Sales Workspace. | Product Team | Must | Prototype | BO-010 | D-283 |
| BR-UX-012 | Important screens shall provide clear loading, empty, error, permission, integration-readiness, and limit-reached states. | UI/System | Must | MVP | BO-010 | D-282 |



### 10.17 Security, Privacy, Compliance, and Trust


| Requirement ID | Business requirement | Primary actor/scope | Priority | Delivery | Objective | Decision trace |
| --- | --- | --- | --- | --- | --- | --- |
| BR-SEC-001 | Tenant isolation shall be enforced server-side for all protected records, retrieval, agent context, tools, background jobs, files, logs, and provider events. | System | Must | MVP | BO-005 | Security baseline |
| BR-SEC-002 | Every protected request shall verify authentication, Organization membership, permission, Project access, resource ownership/scope, entitlement, and state. | System | Must | MVP | BO-004 | Security baseline |
| BR-SEC-003 | IDOR and mass-assignment protections shall be explicitly tested for all tenant/project resources. | System | Must | MVP | BO-005 | Security baseline |
| BR-SEC-004 | Privileged values supplied by browsers or agents shall not be trusted without server-side derivation/validation. | System | Must | MVP | BO-004 | Security baseline |
| BR-SEC-005 | AI shall be treated as an untrusted actor and shall not change its own permissions or bypass application controls. | System | Must | MVP | BO-004 | Security baseline |
| BR-SEC-006 | Every outbound communication shall enforce suppression/opt-out and channel eligibility immediately before send. | System | Must | MVP | BO-004 | Security baseline |
| BR-SEC-007 | Provider tokens and API credentials shall be encrypted/protected and exposed to agents only through controlled services. | System | Must | MVP | BO-005 | Security baseline |
| BR-SEC-008 | Incoming provider webhooks/mail events shall be verified, replay-protected, idempotent, and tenant/project resolved. | System | Must | MVP | BO-004 | Security baseline |
| BR-SEC-009 | Critical publication, spend update, message send, handover claim, and conversion operations shall be idempotent where applicable. | System | Must | MVP | BO-004 | Security baseline |
| BR-SEC-010 | Uploaded documents, external pages, inbound messages, and retrieved content shall be treated as untrusted and potential prompt-injection sources. | System/Agents | Must | MVP | BO-005 | Security baseline |
| BR-SEC-011 | Agents shall not follow untrusted-content instructions that conflict with platform policy, Project rules, or tool authorization. | System/Agents | Must | MVP | BO-004 | Security baseline |
| BR-SEC-012 | Sensitive information shall be redacted/minimized in normal logs and email notifications. | System | Must | MVP | BO-005 | Security baseline |
| BR-SEC-013 | Full prompt/response/conversation traces shall require elevated controlled audited access. | System | Must | MVP | BO-005 | Security baseline |
| BR-SEC-014 | Important data corrections, consent changes, suppression changes, merges, and support access shall be auditable. | System | Must | MVP | BO-006 | Security baseline |
| BR-SEC-015 | Production tenant data shall not be casually copied into local development or evaluation environments. | System/Operations | Must | MVP | BO-005 | Security baseline |
| BR-SEC-016 | Master Lead/Business Pool reuse shall be limited to permitted reusable attributes and shall preserve source/licensing restrictions. | System | Must | MVP | BO-005 | Security baseline |
| BR-SEC-017 | Detailed conversation, budget, objections, negotiation, and tenant-supplied private data shall not be reused cross-tenant. | System | Must | MVP | BO-005 | Security baseline |
| BR-SEC-018 | The platform shall support correction, suppression, and future deletion/retention controls for person/company profiles. | System | Must | MVP | BO-005 | Security baseline |
| BR-SEC-019 | The platform shall not assume every discovered email/phone is legally or contractually eligible for outreach. | System | Must | MVP | BO-004 | Security baseline |
| BR-SEC-020 | Provider policy, consent, outreach, profiling, scraping/search, retention, and Oman/GCC/international privacy requirements shall be verified against current official sources before production release. | Legal/Platform | Must | MVP | BO-004 | Security baseline |
| BR-SEC-021 | The platform shall implement per-Organization external-action kill switch and platform-level provider/agent kill switches. | Super Admin/System | Must | MVP | BO-007 | Security baseline |
| BR-SEC-022 | Campaign spend and budget changes shall require explicit authorized action and audit. | System/User | Must | MVP | BO-004 | Security baseline |
| BR-SEC-023 | The MVP shall not allow AI to create or negotiate discounts, contracts, reservations, payment, or final commitments. | System/Agent | Must | MVP | BO-004 | Security baseline |
| BR-SEC-024 | Support access shall be purpose-limited and auditable. | System/Support | Must | MVP | BO-005 | Security baseline |
| BR-SEC-025 | Security, privacy, and provider-policy failures shall create actionable blocked states rather than silent fallback to unsafe behavior. | System | Must | MVP | BO-010 | Security baseline |



### 10.18 Testing and Evaluation


| Requirement ID | Business requirement | Primary actor/scope | Priority | Delivery | Objective | Decision trace |
| --- | --- | --- | --- | --- | --- | --- |
| BR-TST-001 | The platform shall provide a sandboxed Customer Simulator and Evaluator using real agent logic and blocked external side effects. | System/Super Admin | Must | MVP | BO-009 | D-167 |
| BR-TST-002 | Evaluation scenarios shall include interested, price-sensitive, difficult, wrong-fit, competitor, ready-to-buy, handover, opt-out, hallucination, memory, language, and human-takeover cases. | Super Admin/System | Must | MVP | BO-009 | D-171 |
| BR-TST-003 | Evaluation results shall include agent/prompt/model versions, transcript, tool traces, quality metrics, hard failures, warnings, and Pass/Fail/Review result. | Evaluator/System | Must | MVP | BO-009 | D-169 |
| BR-TST-004 | Super Admin shall compare agent/prompt versions on identical scenarios before activation. | Super Admin | Must | MVP | BO-009 | D-170 |
| BR-TST-005 | Critical failures shall block production eligibility until resolved and re-evaluated. | System/Super Admin | Must | MVP | BO-004 | D-169 |
| BR-TST-006 | Important production defects shall become regression scenarios. | Super Admin/System | Must | MVP | BO-009 | D-171 |
| BR-TST-007 | Detailed Organization/Project Test Mode UX is deferred but the architecture shall preserve sandbox adapters and side-effect blocking. | System | Must | MVP foundation | BO-009 | DD-001 |
| BR-TST-008 | All MVP requirements shall have traceable acceptance tests before implementation completion. | Product/QA | Must | MVP | BO-010 | Specification-driven delivery |



## 11. Business Rule Catalogue

| Rule ID | Rule |
| --- | --- |
| BRULE-001 | A registration application does not create an active tenant. |
| BRULE-002 | CR/business registration number uniqueness is evaluated with jurisdiction. |
| BRULE-003 | One User may hold memberships in multiple Organizations. |
| BRULE-004 | Only Super Admin may assign/change the Organization Plan in MVP. |
| BRULE-005 | A Plan limit is either Numeric or Unlimited; Unlimited does not bypass provider, cost, abuse, or safety controls. |
| BRULE-006 | Strategy Revision Allowance is per Project and is consumed only by user-requested revisions after the first output. |
| BRULE-007 | Company/Project clarification and missing-data correction never consume Strategy Revision Allowance. |
| BRULE-008 | Company Understanding does not own Project products, prices, audience, or Strategy. |
| BRULE-009 | Project/Product Understanding must be approved before Strategy. |
| BRULE-010 | Strategy must be accepted, channel settings completed, and final Project onboarding approved before Project becomes LIVE. |
| BRULE-011 | Core Project products cannot be added after onboarding approval. |
| BRULE-012 | Post-onboarding sales-support content cannot silently alter approved Strategy. |
| BRULE-013 | A Project is either B2B or B2C in MVP; an Organization may operate both types. |
| BRULE-014 | A lead must have at least one usable permitted contact method before becoming a captured Master Lead. |
| BRULE-015 | All captured contactable people enter the Master Lead Pool regardless of source. |
| BRULE-016 | Business Pool stores companies; Master Lead Pool stores people. |
| BRULE-017 | A Master Lead has no universal Project score, state, Strategy, or outcome. |
| BRULE-018 | Project-specific sales intelligence remains private to the owning Organization/Project. |
| BRULE-019 | AI may recommend identity matches; deterministic services merge records. |
| BRULE-020 | Merges preserve provenance/conflicts and are reversible by authorized Platform Administration. |
| BRULE-021 | Fresh discovery is attempted before Master Pool supplementation under current policy. |
| BRULE-022 | Web candidate target defaults to Project Lead Limit × configurable multiplier; current default is 10. |
| BRULE-023 | Project Lead allocation fills the limit with highest-ranked valid candidates available. |
| BRULE-024 | Meta inbound leads carry an explicit source-intent signal. |
| BRULE-025 | Meta paid acquisition uses WhatsApp as the primary MVP conversation destination. |
| BRULE-026 | Initial score is Project-specific; dynamic observed evidence can move the score/state materially in either direction. |
| BRULE-027 | Qualified requires fit, need/interest, and realistic ability to progress. |
| BRULE-028 | Not Qualified/Not Interested are Project-specific; Suppressed is an overriding communication restriction. |
| BRULE-029 | No Response is an attention signal, not automatic Not Interested. |
| BRULE-030 | Every active Project Lead has one Next Best Action or Wait/No Action with reason. |
| BRULE-031 | Lead Strategist decides the approach; Sales Agent decides how to communicate it. |
| BRULE-032 | AI may automatically send initial eligible outreach and active-conversation replies. |
| BRULE-033 | Normal follow-ups require human approval unless explicitly delegated per Project Lead. |
| BRULE-034 | Any inbound reply reopens active conversation and invalidates obsolete unsent follow-ups. |
| BRULE-035 | Human may take over at any time; AI sending stops immediately. |
| BRULE-036 | Human may return control to AI only after current-context reassessment. |
| BRULE-037 | Persistent memory is application-managed; an ADK session is not the sole source of truth. |
| BRULE-038 | Sales AI may use only approved Project knowledge and approved product content. |
| BRULE-039 | MVP AI may not finalize meetings, quotes, contracts, discounts, reservations, payments, or sales. |
| BRULE-040 | All final MVP commercial actions are human and must be recorded. |
| BRULE-041 | Campaign publication always requires explicit authorized user action. |
| BRULE-042 | Campaign Agent advises; Campaign Service and provider adapter execute. |
| BRULE-043 | Every campaign requires an end date. |
| BRULE-044 | Authorized users may increase budget, extend, pause, or resume; AI may only recommend. |
| BRULE-045 | Meta overflow leads remain in Master Pool but autonomous Project sales stops at the effective limit. |
| BRULE-046 | Missing channel configuration does not block onboarding but blocks affected execution. |
| BRULE-047 | A failed channel may fall back only to another configured, healthy, eligible, approved channel. |
| BRULE-048 | Organization Owner is protected; other users receive explicit permissions and Project access. |
| BRULE-049 | Agents cannot directly change permissions, suppression, entitlements, spend, or protected state. |
| BRULE-050 | All agent tools are governed application interfaces. |
| BRULE-051 | All meaningful agent outputs are validated/persisted before downstream use. |
| BRULE-052 | Agent runs are version-pinned and auditable. |
| BRULE-053 | Critical evaluation failures block agent activation. |
| BRULE-054 | Action Center is authoritative for unresolved work; notifications are alerts only. |
| BRULE-055 | Email notifications minimize sensitive content. |
| BRULE-056 | Project state is ONBOARDING, LIVE, or ENDED; ENDED retains history and stops new commercial execution. |
| BRULE-057 | Opportunity is optional; Conversion is an auditable event. |
| BRULE-058 | Learning may adapt a Lead automatically but may only recommend Project Strategy changes. |
| BRULE-059 | Provider/legal requirements must be verified from current official sources before production release. |
| BRULE-060 | No critical business rule may exist only in an editable prompt or UI. |

## 12. Roles, Responsibility, and Authorization

### Platform authority
- **Platform Super Admin:** Plans, Organizations, subscriptions, integration credentials/configuration, agents/models/prompts, evaluation, observability, safety controls, suspension, closure, and support authorization.
- **Platform Admin/Support:** limited registration/support/operations permissions; no automatic access to high-risk agent/provider/security configuration.

### Organization authority
- **Organization Owner:** protected account authority; manages user permissions and Project access.
- **Organization Users:** receive explicit permissions and Project access. Functional labels/presets may simplify assignment but are not the authorization source.

### Authorization formula

```text
Authenticated identity
+ active Organization membership
+ explicit permission
+ Project access
+ resource ownership/scope
+ Plan entitlement
+ integration readiness
+ action/state policy
= Allow or Deny
```

## 13. AI and Human Responsibility Matrix

| Capability | AI responsibility | Human responsibility | Application-service responsibility |
| --- | --- | --- | --- |
| Company Understanding | Analyze, classify, ask targeted questions, propose overview | Review/correct/approve | Scope context, persist versions, authorize, audit |
| Project/Product Understanding | Extract, ask questions, propose approved baseline | Review/correct/approve | File/source access, version, isolate, audit |
| Strategy | Generate structured package and controlled revisions | Approve/revise; configure channels | Validate allowance, versions, readiness |
| Lead acquisition planning | Recommend Web/Meta/Both and produce source plans | Approve plan | Entitlement, provider readiness, limits, execution jobs |
| Campaign | Prepare settings/copy/form/creative guide; monitor/recommend | Approve form/media/settings; explicitly publish/change spend | Permission, spend, dates, provider API, idempotency, audit |
| Lead enrichment | Research permitted evidence | Review exceptions if needed | Source permissions, identity merge, storage |
| Lead Strategy | Assess, score, hypothesize, decide Next Best Action | Inspect/instruct/take over | Validate/persist state and memory |
| Active sales conversation | Communicate autonomously within rules | Monitor; intervene anytime | Eligibility, suppression, channel send, audit |
| Follow-up | Recommend timing/angle/message; execute if delegated | Approve/edit/reject or delegate | Schedule/revalidate/send/cancel |
| Final action | Prepare handover brief | Claim handover; complete and record final action | Claim locking, outcome persistence, attribution |
| Learning | Adapt individual lead; recommend Project improvements | Approve new Project Strategy version | Aggregate evidence, preserve versioning |

## 14. Plans, Limits, and Entitlements

Plans may include any combination of enabled features and Numeric/Unlimited limits. Initial required fields are Plan name, description, price/reference, status, maximum Projects, Strategy Revision Allowance per Project, Project Lead Limit per Project, and channel/feature entitlements.

Commercial Unlimited does not disable provider quotas, cost safeguards, abuse controls, security, or policy enforcement. Organization-specific overrides may change the effective allowance. The Organization sees its effective configuration but only Super Admin changes it.

## 15. Control Panel Business Requirements

### Super Admin CP
Required modules: Dashboard, Registration Applications, Organizations, Organization Detail, Plans, Overrides, Platform Integrations, Organization Integrations, Users, Agents, Prompt/Agent Versions, Model/Credential Assignment, Agent Operations, Test Lab/Evaluations, Usage/Cost, Jobs/Errors, Audit, Support Access, and System Settings.

### Organization Administration
Required modules: Organization Overview, Company Understanding, Team & Access, Projects, Email Integration, Action Center, Usage & Plan, Organization Settings, Notification Preferences, and Audit/Activity.

### Project Workspace
Required modules: Project Dashboard, Products, Strategy, Campaigns, Lead Acquisition, Leads, Sales Inbox/Conversation, Handovers, Opportunities/Conversions, Analytics, and Project Settings.

## 16. Analytics and Success Measures

The MVP shall measure:
- Acquisition volume and contactability.
- Initial fit and dynamic qualification.
- Engagement and qualification rates.
- Qualified Handover Rate.
- Human handover acceptance/response time.
- Handover-to-Conversion Rate.
- Campaign spend, cost per lead where available, and downstream lead quality.
- Web discovery candidate/contactable/selected/converted funnel.
- AI grounding, objection handling, follow-up, memory, and handover quality.
- Human final-action outcomes.
- Strategy recommendation evidence and approval.

## 17. Security and Privacy Expectations

Security and privacy are business requirements, not later implementation details. The platform must protect tenant/project isolation, communication eligibility, suppression, credentials, webhook integrity, audit, data minimization, sensitive logging, support access, and prompt/tool boundaries.

The Master Lead and Business Pools create material privacy and licensing risk. No cross-tenant reuse is allowed without permitted, evidence-backed, source-aware reusable classification. Provider terms and jurisdiction requirements must be researched from current official sources before production release. No legal certainty is claimed by this BRD.

## 18. Compliance and External Policy Dependencies

- Meta Marketing API, advertising terms, account permissions, lead forms, Click-to-WhatsApp, and campaign operations.
- WhatsApp Business Platform, opt-in/template/conversation requirements, prohibited categories, and enforcement.
- Instagram and Facebook Messenger API permissions and messaging restrictions.
- Google Search/Maps/custom search and other research-provider storage/licensing restrictions.
- Generic email sender authentication, inbound mailbox access, anti-spam, unsubscribe, suppression, reputation, and jurisdiction obligations.
- Oman/GCC and intended international privacy, profiling, marketing, retention, deletion, correction, and cross-border transfer requirements.
- Prospect-data and website-source terms, robots/access limitations, and data licensing.

## 19. Assumptions and Dependencies

### Assumptions

- Super Admin can obtain and configure required Organization-specific Meta/WhatsApp/Instagram/Facebook credentials and permissions.
- Organizations can supply a legitimate business mailbox and approved Sales AI identity.
- Current provider APIs permit the intended use after required reviews/permissions; this remains unverified until official research.
- Organization customers accept human approval for normal follow-ups and all final MVP actions.
- The platform may retain permitted reusable lead/business data subject to later legal/privacy validation.
- The first MVP will be locally deployable/testable before production hardening.

### Dependencies

- Approved provider accounts, permissions, credentials, and Organization-specific connection data.
- Google ADK suitability validation for the required agent topology and observability.
- A secure application backend, tenant-aware database, job system, storage, provider adapters, and audit capability.
- Approved Project/Product knowledge sufficient to ground Sales AI.
- Human users available to approve follow-ups and claim final-action handovers.
- Current policy/legal research before live outreach and cross-tenant reusable profiling.

## 20. Risks

| Risk ID | Risk | Severity | Impact | Treatment |
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

## 21. Deferred and Open Items

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

### Open

| ID | Area | Unresolved matter |
| --- | --- | --- |
| OQ-001 | Campaign lifecycle D-274 | The recommended lifecycle is defined but still requires explicit confirmation. |
| OQ-002 | Plan prices and numeric allowances | Commercial values are intentionally not yet set. |
| OQ-003 | Generic inbound mailbox implementation | The product requirement is generic inbound/outbound mailbox support; exact IMAP/provider compatibility belongs to integration design. |
| OQ-004 | Provider and legal policy verification | Meta, WhatsApp, Instagram, Facebook, Google/Search, email, prospect data, privacy, and Oman/GCC obligations require current official research and legal review. |
| OQ-005 | Full Arabic/RTL UI baseline | The exact degree of UI localization, dialect behavior, and RTL coverage remains to be locked. |

## 22. MVP Business Acceptance Criteria

| Acceptance ID | Criterion |
| --- | --- |
| BAC-001 | A valid applicant can submit a company application; no Organization access is granted until Super Admin approval. |
| BAC-002 | Duplicate CR/company/domain cases are blocked with actionable login/support guidance. |
| BAC-003 | Super Admin can activate an Organization with a selected Plan and effective limits. |
| BAC-004 | An authorized user can complete Company Understanding through free text/documents, AI review, targeted questions, final approval, and versioned output. |
| BAC-005 | Private Organization Intelligence and reusable Business Pool data remain visibly/securably separate. |
| BAC-006 | A user can create a Project, provide product/material inputs, answer dynamic questions, approve Project/Product Understanding, and confirm objective. |
| BAC-007 | The system generates a structured Strategy package; revision usage is correctly counted and enforced. |
| BAC-008 | The user configures channel/Sales AI behavior and completes Project onboarding into LIVE. |
| BAC-009 | The Product Catalogue is auto-populated, does not allow new core products, and accepts sales-support media/content. |
| BAC-010 | The Organization and Project dashboards identify the correct next action and show relevant usage/statistics. |
| BAC-011 | The Lead Discovery Agent can recommend Web, Meta, or Both and produce a human-reviewable plan. |
| BAC-012 | Web discovery can create contactable Master Leads with provenance, rank them, and allocate Project Leads up to the limit. |
| BAC-013 | A Meta campaign can be configured, creatively guided, reviewed, explicitly published, and monitored from the platform. |
| BAC-014 | Meta-generated leads enter the Master Lead Pool and start WhatsApp qualification/conversation. |
| BAC-015 | A generic business mailbox can send first outreach, ingest threaded replies, and support autonomous active email conversation. |
| BAC-016 | Bounce and channel failures create logged, visible, actionable states and do not cause unsafe blind retries. |
| BAC-017 | Master Lead and Business Pool matching preserves provenance and does not leak tenant-private information. |
| BAC-018 | Initial score/ranking and dynamic qualification operate differently for Web and Meta sources while using one Project Lead model. |
| BAC-019 | Sales AI communicates naturally, remembers context, does not repeat answered questions, and does not invent product/commercial facts. |
| BAC-020 | Normal active conversation is autonomous; normal follow-up requires approval unless delegated. |
| BAC-021 | A user can approve/edit/reject a follow-up, delegate follow-up, monitor a conversation, take over, and return to AI. |
| BAC-022 | Unresponsive leads are flagged with strategy/recommendation without automatic Not Interested classification. |
| BAC-023 | Suppression prevents all applicable sends even if an agent recommends contact. |
| BAC-024 | Final commercial action triggers or requires human control; an authorized user can claim handover. |
| BAC-025 | The human can record a conversion/outcome and source/campaign attribution remains intact. |
| BAC-026 | Campaign/source/AI/human analytics reflect the end-to-end funnel. |
| BAC-027 | Super Admin can manage Organizations, Plans, overrides, integrations, agents, models, credentials, versions, logs, and safety controls. |
| BAC-028 | Organization Owner can manage users through direct permissions and Project access. |
| BAC-029 | Action Center and in-app/email notifications reach only authorized relevant users. |
| BAC-030 | Every high-risk action creates a tenant-scoped audit event. |
| BAC-031 | Agent versions can be evaluated, compared, activated, and rolled back without rewriting historical runs. |
| BAC-032 | The same Project Lead/conversation cannot be accessed from another Organization or unrelated Project. |
| BAC-033 | Project Lead overflow does not stop existing conversations and does not activate new autonomous sales beyond the allowance. |
| BAC-034 | Ending a Project stops new acquisition/outreach while preserving historical data/analytics. |
| BAC-035 | The first prototype demonstrates company onboarding through lead conversation in a clean, intuitive navigation flow. |

## 23. Traceability Matrix

| Business objective | Primary requirement domains |
| --- | --- |
| BO-001 | Registration; Company Understanding; Projects/Product Understanding; Strategy |
| BO-002 | Lead Acquisition; Meta Campaigns; Email and Messaging |
| BO-003 | Enrichment/Qualification; Lead Strategy; Sales Conversation |
| BO-004 | Plans/limits; AI/Human controls; Security; Handover; Campaign publish |
| BO-005 | Tenancy; Pools; identity matching; privacy; agent context isolation |
| BO-006 | Audit; attribution; conversion; analytics; versioning |
| BO-007 | Super Admin CP; Plans; integrations; agent operations; usage |
| BO-008 | Project modes; ICP/personas; B2B/B2C qualification |
| BO-009 | ADK agents; evaluation; versioning; observability; learning |
| BO-010 | Dashboards; Action Center; Sales Workspace; errors; acceptance walkthrough |

## 24. Change Control and Approval

- Any change to approved business scope, autonomy, spend, data reuse, permissions, Project lifecycle, final-action boundary, or provider behavior requires a new decision/change request.
- The BRD must not be silently edited to match code.
- The Master PRD and later system/API/data specifications must trace to this BRD.
- Production coding may begin only after the required system, data, API, security, integration, and test specifications are approved.


---

**Approval statement:** `AI_Sales_Agent_MVP_BRD_v1.0 is approved as the MVP business baseline.`


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

