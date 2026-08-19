# AI Sales Agent — Super Admin Vibe Coding Plan

| Field | Value |
|---|---|
| Version | 1.0 |
| Status | Proposed Execution Plan |
| Last Updated | 2026-08-18 |
| Scope | Super Admin CP + internal components + backend required to operate and test Agents |

## Objective

Build the Super Admin Control Panel and the internal platform foundations required to operate it. Use the CP to configure and test all MVP Agents. Do not build the company-facing UI in this phase except backend/domain structures that are necessary for Super Admin workflows or Agent testing.

## Phase 0 — Baseline & Repository Guardrails

1. Add the approved Super Admin specification and decision register to the repository.
2. Add a `/docs/source-of-truth/` index pointing to:
   - BRD v1.1
   - Master PRD v1.1
   - Stage 1 v1.1
   - Super Admin CP Spec
   - Super Admin Decision Register
3. Add rule: coding must not silently invent missing business behavior.
4. Add environment separation: TEST vs PRODUCTION.
5. Add database migrations only from approved domain model decisions.
6. Add seed/bootstrap mechanism for the first Super Admin account.

Exit gate:
- docs committed;
- project boots locally;
- database connects;
- first Super Admin can be provisioned safely.

## Phase 1 — Core Backend Foundation

Build:
- application/config foundation;
- database;
- migrations;
- API error envelope;
- correlation IDs;
- idempotency helper;
- optimistic concurrency helper;
- background-job abstraction;
- secret-reference abstraction;
- audit writer;
- operational logger;
- environment marker.

Core entities:
- SuperAdminAccount
- SuperAdminSession
- AuthenticationChallenge
- Organization
- User
- Membership
- RegistrationApplication
- Package
- PackageVersion
- OrganizationPackageSnapshot
- OrganizationLimitOverride
- PlatformIntegration
- OrganizationIntegration
- AgentDefinition
- AgentVersion
- AgentRun
- AgentTestRun
- OperationalLogEvent
- OperationalIssue
- AuditEvent

Exit gate:
- migrations pass;
- CRUD/domain tests pass;
- tenant/platform scope rules tested.

## Phase 2 — Super Admin Authentication

Build:
- login;
- password verification;
- 6-digit email OTP;
- resend cooldown;
- OTP expiry/attempt limits;
- session creation/revocation;
- protected Super Admin route middleware;
- emergency OTP-provider recovery configuration.

Frontend:
- Login
- OTP Verification

Tests:
- invalid credentials;
- expired OTP;
- resend invalidation;
- rate limiting;
- session expiry;
- logout.

Exit gate:
- Super Admin can securely log in and reach an empty Dashboard.

## Phase 3 — Applications & Organizations

Backend:
- public application record/API;
- duplicate CR + jurisdiction validation;
- duplicate-domain informational matching;
- application state transitions;
- approve & activate;
- reject then later approve;
- manual Organization creation;
- Owner invitation;
- Organization activation;
- logical deletion;
- Organization read-only deleted state.

Frontend:
- Applications List
- Application Detail
- Organizations List
- Create Organization
- Organization Overview
- Account

Tests:
- idempotent public approval;
- existing Owner User reuse;
- duplicate CR block;
- duplicate domain allowed;
- deletion stops customer/external operation eligibility;
- Master Lead/Business Pool retention interfaces preserved.

Exit gate:
- both Organization creation paths work end to end.

## Phase 4 — Packages, Credits & Limits

Backend:
- Package/PackageVersion;
- ACTIVE/INACTIVE;
- snapshot assignment;
- Package change with impact preview;
- numeric overrides only;
- effective limit calculation;
- limit usage query contracts.

Frontend:
- Packages List
- Create/Edit Package
- Package Detail
- Organization Package & Limits

Tests:
- Package edits do not change existing snapshots;
- no feature override;
- credit/lead/project override;
- reducing limits below current usage follows locked continuity rules;
- Strategy Credit cannot go below consumed amount;
- Package change clears overrides.

Exit gate:
- Super Admin can configure commercial limits without ambiguity.

## Phase 5 — AI Cost Attribution

Backend:
- AgentRunUsage;
- AgentRunCost;
- pricing basis/version;
- Organization/Project/Lead/Agent attribution;
- retry/child-run cost handling;
- UNPRICED state;
- aggregation APIs.

Frontend:
- Platform AI Usage & Cost
- Organization AI Usage & Cost

Tests:
- retries counted;
- failed billable runs counted;
- historical cost not silently repriced;
- Test environment excluded from customer Sales AI Cost.

Exit gate:
- Super Admin can explain cost by Organization, Project, and Agent.

## Phase 6 — Integration Management

### 6A Organization Integrations
- Meta Ads
- WhatsApp
- Instagram
- Facebook/Messenger

Build:
- entitlement/readiness separation;
- secure credential-reference storage;
- configure/test/disable/enable;
- capability status;
- no plaintext secret redisplay;
- deleted Organization read-only behavior.

### 6B Core Integrations
- AI Provider(s)
- Google/Search
- Notification Email
- optional File/Processing if needed

Build:
- dependency map;
- impact preview on disable;
- test before Connected;
- provider/model resource registry for Agent configuration.

Exit gate:
- integration states and secret handling validated; no real unsafe provider side effect in test environment.

## Phase 7 — Agent Registry & Version Management

Seed fixed ten Agent Definitions.

Build:
- Agent list/detail;
- immutable Agent Definition boundaries;
- Agent Version cloning;
- Draft editing;
- provider/model/credential reference selection;
- prompt editor;
- tool grants from approved set;
- interaction grants from approved graph;
- runtime settings within hard bounds;
- version states;
- activate/rollback;
- Agent ENABLED/DISABLED with impact preview.

Tests:
- active/historical immutability;
- one Active Version;
- in-flight runs remain version-pinned;
- prompt cannot bypass application safety;
- invalid provider/tool/interaction rejected.

Exit gate:
- every Agent can be configured/versioned safely.

## Phase 8 — Simple Agent Test Harness

Create common Test Agent framework:
- TEST environment only;
- no production mutations or provider side effects;
- upstream Test Results selectable/uploadable;
- result storage;
- cost/duration/tool summaries.

Implement tests in dependency order:
1. Company Understanding
2. Project Understanding
3. Strategy
4. Lead Discovery
5. Lead Enrichment
6. Lead Strategist
7. Sales / Conversation + Customer Simulator
8. Campaign
9. Customer Simulator standalone
10. Evaluator

Specific UX:
- Sales Agent opens simulated chat.
- Customer Simulator uses persona profiles.
- Evaluator accepts completed Agent Test Results and returns PASS/REVIEW/FAIL.

Exit gate:
- every Agent has at least one successful test path;
- no test can invoke real external side effects.

## Phase 9 — System Health & Operations

Backend:
- ServiceHealth;
- OperationalIssue;
- background-job visibility;
- retryability classification;
- issue deduplication/recovery;
- UNKNOWN health on stale monitoring.

Frontend:
- System Health
- Service/Issue Detail
- retry-safe job action
- deep links to owning Agent/Integration/Organization

Exit gate:
- failures surface with actionable recovery and no blind external retries.

## Phase 10 — Logs & Agent Runs

Backend:
- structured operational log events;
- Agent Run parent/child relationships;
- tool calls;
- environment separation;
- sensitive-content classification/reveal;
- filtered queries.

Frontend:
- All Logs
- Log Detail
- Agent Runs
- Agent Run Detail
- deliberate sensitive-content reveal

Exit gate:
- one failing Agent/provider workflow can be traced end to end by correlation ID.

## Phase 11 — Platform Audit

Backend:
- immutable AuditEvent;
- old/new values;
- reason/result;
- atomic/reliable write with protected operations;
- sensitive-trace-access audit.

Frontend:
- Platform Audit
- Audit Detail
- Organization-filtered Audit reuse

Exit gate:
- all high-impact Super Admin mutations produce immutable audit evidence.

## Phase 12 — Super Admin Integration & UX Hardening

Validate complete journeys:
1. Login -> OTP -> Dashboard
2. Public Application -> Approve -> Organization Active
3. Manual Organization -> Package -> Activate
4. Organization -> Override Limit
5. Organization -> Configure Meta/WhatsApp/etc.
6. Configure Core AI Provider
7. Agent -> Create Version -> Configure -> Test -> Approve -> Activate
8. Sales Agent -> Customer Simulator conversation
9. Evaluator -> evaluate Test Result
10. Provider failure -> System Health -> Logs -> remediation
11. High-cost Organization -> AI Usage -> Agent Runs
12. Organization Delete -> operations stop -> historical read-only -> Audit

Run:
- dead-end review;
- loop review;
- stale/concurrency tests;
- duplicate submission tests;
- permission/session tests;
- secret-leak tests;
- test-vs-production isolation tests;
- accessibility/responsive smoke tests.

Exit gate:
- Super Admin CP accepted as a complete internal MVP.

## Phase 13 — Freeze Super Admin Baseline

Before company-side UI work:
- update BRD/PRD/Stage1 if implementation exposed approved changes;
- update ERD/API spec;
- update decision register;
- update traceability;
- produce Super Admin validation report;
- lock Super Admin implementation baseline.

Only then begin Company-Facing UX/coding.

## Coding Session Rules

For every coding session:
1. Name exact module and requirement/decision IDs.
2. Do not implement unrelated company-facing screens.
3. Add/modify data model before UI if the UI needs new authoritative data.
4. Backend enforcement is mandatory for protected rules.
5. UI hiding is never authorization.
6. Every mutation defines success, validation, conflict, retry, and audit behavior.
7. Every async operation exposes real state.
8. Every external side effect is adapter-based and idempotent where required.
9. TEST environment must never fall back to production providers.
10. End each session with:
   - tests run;
   - files changed;
   - requirements implemented;
   - open issues;
   - documentation updates.

## Recommended Session Order

1. Repo/domain foundation
2. Super Admin auth
3. Organizations/applications
4. Packages/overrides
5. Integration framework
6. AI provider registry
7. Agent registry/versioning
8. Agent test harness
9. Agent-by-agent tests
10. AI cost attribution
11. System Health/jobs
12. Logs/Agent Runs
13. Audit
14. Dashboard aggregation
15. End-to-end Super Admin hardening
16. Source-of-truth update/freeze
17. Company-facing design/coding resumes
