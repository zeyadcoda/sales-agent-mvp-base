# AI Sales Agent — Super Admin Control Panel MVP Specification

| Field | Value |
|---|---|
| Version | 1.0 |
| Status | Approved Working Baseline |
| Owner | Project Owner |
| Last Updated | 2026-08-18 |
| Scope | Super Admin CP, internal platform services required to support it, Agent configuration/testing, operations, logs, and audit |

## 1. Purpose

The Super Admin Control Panel is the single internal platform administration experience for the MVP. There is one internal actor only: Super Admin.

The Super Admin can:
- authenticate with password + email OTP;
- manage Organization accounts and public applications;
- create/configure Packages;
- assign Packages and numeric Organization overrides;
- view Sales AI usage/cost;
- configure Organization Meta/WhatsApp/Instagram/Facebook integrations;
- configure Core Platform Integrations;
- configure, test, activate, disable, and roll back AI Agent Versions;
- monitor System Health and background operations;
- inspect Logs and Agent Runs;
- review immutable Platform Audit.

## 2. Navigation

```text
Dashboard
Organizations
Applications
Packages
AI & Usage
Integrations
AI Agents
System Health
Logs
Audit
```

Organization context:
```text
Overview
Account
Package & Limits
AI Usage & Cost
Integrations
Projects & Activity
Audit
```

Agent context:
```text
Overview
Active Configuration
Versions
Operations
Test Agent
```

## 3. Authentication

- Email + password + 6-digit email OTP.
- OTP expires in 10 minutes.
- Max 5 failed attempts.
- Resend cooldown 60 seconds.
- Resend invalidates prior OTP.
- Logout revokes session.
- Emergency OTP delivery recovery required outside normal logged-in configuration.

## 4. Dashboard

Sections:
1. Needs Attention
2. AI Cost & Consumption
3. Organizations
4. System Health
5. Recent Important Activity

Primary action: Create Organization.

## 5. Applications

States:
`SUBMITTED -> UNDER_REVIEW -> APPROVED`
or
`SUBMITTED -> UNDER_REVIEW -> REJECTED -> APPROVED`

- No Request Information workflow.
- Duplicate CR + jurisdiction blocks.
- Duplicate domain allowed.
- Approval creates + activates Organization.
- Rejected application can later be approved.

## 6. Organizations

Lifecycle:
`INACTIVE -> ACTIVE -> DELETED`
and `INACTIVE -> DELETED`.

- DELETED is logical deletion, terminal/read-only in MVP.
- No Suspend.
- No Restore.
- Master Lead/Business Pool reusable identities remain.
- Private Project Lead intelligence remains private.
- Delete requires reason and stops new external/commercial actions.

## 7. Packages & Limits

Package states: ACTIVE / INACTIVE.

Package fields:
- Maximum Projects
- Strategy Credits per Project
- Lead Limit per Project
- Feature entitlements

Numeric limits support NUMERIC / UNLIMITED.

Organization overrides are numeric only:
- Maximum Projects
- Strategy Credits per Project
- Lead Limit per Project

No feature entitlement override.

## 8. Sales AI Cost

Primary internal cost metric includes:
- Sales / Conversation
- Lead Strategist
- Lead Enrichment when part of active lead sales process

Tracked per Organization, Project, Project Lead where applicable, Agent, Version, model/provider, usage and cost.

## 9. Organization Integrations

- Meta Ads
- WhatsApp
- Instagram
- Facebook/Messenger

States:
NOT_CONFIGURED, CONFIGURING, CONNECTED, PARTIAL, ERROR, REAUTH_REQUIRED, DISABLED.

Secrets never redisplayed.

## 10. Core Integrations

- AI Provider(s)
- Google/Search
- Platform Notification Email
- File/Processing only when Super Admin configuration is required

Agent provider/model selections use configured Core Integration resources.

## 11. AI Agent Registry

Fixed MVP Agent set:
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

Version lifecycle:
`DRAFT -> TESTING -> UNDER_REVIEW -> APPROVED -> ACTIVE -> RETIRED`

Agent operational status ENABLED / DISABLED.

## 12. Simple Agent Testing

Each Agent has Test Agent.

Upstream Test Results can be selected/uploaded as downstream Agent inputs.

Sales Agent opens interactive simulated chat using Customer Simulator persona.

Evaluator consumes Agent Test Result and returns PASS / REVIEW / FAIL.

Tests never create production side effects.

## 13. System Health

States:
HEALTHY, DEGRADED, DOWN, UNKNOWN.

Operational Issues:
OPEN / RESOLVED.

Manual retry only for backend-declared retry-safe jobs.

## 14. Logs & Agent Runs

Logs: INFO / WARNING / ERROR.

Agent Run outcomes:
SUCCESS, RETRYABLE_FAILURE, NEEDS_INFORMATION, NEEDS_HUMAN_REVIEW, BLOCKED, FAILED.

Sensitive execution content hidden by default and deliberate reveal is audited.

## 15. Audit

Immutable, application-created, one platform-wide source.
No manual edit/delete.
No secrets.
Historical references remain resolvable.

## 16. Build Boundary for the Next Coding Session

Build only:
- Super Admin frontend
- Super Admin backend/domain services
- authentication
- Organization/application/package services needed by CP
- numeric limit/override service
- AI usage/cost attribution
- Organization/Core integration configuration layer
- Agent Registry/Version management
- simple Agent Test harness
- Agent runtime interfaces required for testing
- System Health/operations
- Logs/Agent Runs
- Audit
- persistence and APIs required for the above

Do not build the company-facing UI in this coding phase except minimal backend/domain foundations required to support Super Admin data and Agent testing.

