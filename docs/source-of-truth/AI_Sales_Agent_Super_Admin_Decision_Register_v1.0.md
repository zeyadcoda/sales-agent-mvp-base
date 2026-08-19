# AI Sales Agent — Super Admin Decision Register

| Field | Value |
|---|---|
| Version | 1.0 |
| Status | Approved |
| Last Updated | 2026-08-18 |

This register summarizes the approved Super Admin implementation decisions captured during the design sessions. Detailed behavior is in `AI_Sales_Agent_Super_Admin_CP_MVP_Spec_v1.0.md` and the v1.1 project baseline files.

## Decision Groups

- DEC-SA-001–014: Organization logical deletion, Strategy Credits, Organization overrides, Sales AI cost attribution, authentication.
- DEC-SA-015–022: Dashboard and dual Organization entry paths.
- DEC-SA-023–037: Public application simplification and manual Organization creation.
- DEC-SA-038–054: Organization directory, Organization shell, account/lifecycle, deletion behavior.
- DEC-SA-055–069: Packages, Package Versions, numeric limits/overrides, no feature entitlement overrides.
- DEC-SA-070–081: Sales AI cost and consumption.
- DEC-SA-082–094: Organization integrations.
- DEC-SA-095–107: Core platform integrations.
- DEC-SA-108–128: AI Agent management and Agent Version lifecycle.
- DEC-SA-129–148: Simple Agent testing.
- DEC-SA-149–160: System Health & Operations.
- DEC-SA-161–176: Logs & Agent Runs.
- DEC-SA-177–187: Platform Audit.

## Superseded MVP Concepts

- Platform Support/Admin roles: removed; one internal Super Admin actor only.
- Support Access module: removed from current MVP Super Admin CP.
- Organization Suspension state: removed from current MVP lifecycle.
- Request Information registration state/workflow: removed.
- Duplicate domain as blocking registration rule: removed; duplicate domains are allowed.
- Feature entitlement overrides: removed from MVP.
- Complex Agent Scenario Library / formal Release Evaluation / Version Comparison Test Lab: replaced by simple Agent-specific manual Test Agent workflows.
