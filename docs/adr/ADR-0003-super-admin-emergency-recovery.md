# ADR-0003 — Super Admin Emergency Recovery

Status: APPROVED WORKING BASELINE
Date: 2026-08-20

## Context

The MVP has one internal platform actor: Super Admin. Its approved authentication flow is email, password, a six-digit email OTP, and then a PostgreSQL-backed session. If the notification-email provider is unavailable, the sole Super Admin could otherwise be locked out indefinitely.

A normal browser recovery flow, reusable recovery code, permanent OTP-disable setting, or direct session-creation command would create a second authentication mechanism with a broader attack surface. The deployment operator already has a separate infrastructure authority boundary through controlled access to the deployed application, server-side configuration, and PostgreSQL.

The local `AUTH_OTP_BYPASS` setting does not solve this production problem. It is a developer convenience that is accepted only when `APP_ENV=local` and remains separate from emergency recovery.

## Decision

Emergency recovery uses a deployment-authorized, one-time, server-side authorization.

- An infrastructure-authorized operator runs `platform-admin auth-recovery authorize` outside HTTP and identifies exactly one Super Admin by normalized email.
- The command requires a non-empty reason and deployment/operator label, shows the target and impact, and requires an explicit interactive `y` or `yes` confirmation. The default is no, and there is no non-interactive confirmation flag.
- PostgreSQL stores the authorization. It contains no password, OTP, recovery code, token, session secret, API key, or provider credential.
- An authorization is valid for 10 minutes from a PostgreSQL `clock_timestamp()` materialized after the account lock, is tied to one Super Admin, does not renew, and becomes unusable when expired, consumed, or revoked.
- At most one unresolved authorization may exist for a Super Admin; elapsed rows are terminalized with `expired_at` before replacement. Creation, revocation, and consumption use account-first locking and database constraints so concurrent operations serialize consistently.
- Existing password-login throttling remains first in the login path. Redis remains required and fails closed.
- The login service does not inspect or mutate recovery state until the account is eligible and the password has been verified successfully. Unknown account, wrong password, and inactive account behavior remains safe and does not disclose an authorization.
- After a correct password, the application generates independent random session/CSRF material immediately before invoking the consuming operation. One PostgreSQL transaction may then atomically consume an eligible authorization, invalidate older pending OTP challenges, insert the normal session state containing only the session-token hash, and append the consumption Audit Event.
- The raw session token is delivered only through the existing HttpOnly session cookie. A recovery session has the same lifetime, revocation, CSRF, and lookup rules as an OTP-created session.
- If no authorization is consumed, login follows the existing OTP creation and delivery path. Email failure never authenticates a caller.
- Concurrent correct-password requests may produce at most one recovery-authenticated session. A request that loses the recovery-consumption race follows the normal OTP path or fails safely.
- Authorization, consumption, and revocation are recorded in the one platform-wide append-only Audit source. Database triggers reject `UPDATE`, `DELETE`, and `TRUNCATE`; Audit stores safe references and operational attribution, never credentials or raw authentication material.
- Deployment-only `status`, `revoke`, and `audit` subcommands provide safe operational inspection. Revocation requires a reason, operator label, impact summary, and interactive confirmation.

There is no recovery HTTP endpoint, browser route, UI control, browser-visible recovery identifier, or user-entered recovery code. The existing Login page and its request shape remain unchanged.

## Security implications

- Deployment access is the authority boundary. Production access to the command, application runtime, database network, and server-side configuration must be restricted through deployment/IAM controls, preferably with strong authentication, least privilege, just-in-time access, and operator-level traceability.
- The command does not accept a password or database credential as an argument and prints no secret. Production credentials must be injected by the deployment platform or managed secret system rather than copied into shell history.
- PostgreSQL is authoritative for authorization identity, operation time, expiry, consumption, revocation, session creation, and Audit Events. Recovery operations use `clock_timestamp()` after account locking rather than application, operator, or browser time. Redis is not a recovery-state store.
- Redis unavailability still blocks password login before recovery can be consumed. Recovery does not weaken account, password, IP, email, or dependency-failure controls.
- Account-first locking gives authorization, revocation, and login consumption a consistent lock order. The consuming update rechecks account ownership, expiry, and terminal fields in SQL.
- Audit is created by trusted application services/CLI operations. Browsers and Agents cannot create, edit, delete, or invoke recovery events.
- Application logs and Audit Events exclude passwords, OTPs, raw session/CSRF tokens, recovery secrets, HMAC keys, SMTP credentials, database credentials, and raw provider errors.

## Alternatives rejected

- A permanent or environment-controlled production OTP bypass: too easy to leave enabled and not one-time.
- A recovery code or token printed by the CLI and entered in the browser: exposes a credential to terminal history, copy/paste paths, browser state, and network requests.
- A hidden recovery HTTP endpoint: attackers are assumed to know all routes and request shapes.
- Direct database insertion of a session: bypasses password verification, login throttling, normal session creation, and reliable Audit behavior.
- A Super Admin or Platform Support recovery UI: cannot solve provider lockout safely and would reintroduce a removed platform role/capability.
- Redis-only or in-memory authorization: not durable or authoritative and could be lost or recreated unsafely.

## Consequences

- A configured but unavailable notification-email provider no longer permanently locks out the sole Super Admin, provided the API, PostgreSQL, Redis, correct password, and deployment authority remain available.
- Recovery cannot solve a forgotten password, unavailable PostgreSQL/Redis/API, or a deployment that cannot start because mandatory configuration is structurally invalid.
- Operators need a protected production runbook and IAM policy for the CLI. Possession of deployment/database authority remains highly privileged and must be monitored outside the application as well as within Platform Audit.
- The implementation introduces a reusable platform-wide Audit foundation but no Audit UI, export, generic audit-write API, or unrelated administration framework.
- Audit retention policy and production managed-secret integration remain separate hardening work.
- The Super Admin frontend remains unchanged.

## What remains prohibited

- A recovery endpoint, recovery page, hidden browser route, recovery request field, or browser-visible recovery identifier.
- A reusable recovery token/code, password bypass, or login without the correct password.
- A special, longer-lived, or differently protected recovery session.
- Bypassing Redis login throttling or permitting authentication when Redis fails.
- Reusing, renewing, or consuming an expired, consumed, or revoked authorization.
- Allowing Agents, prompts, frontend code, or Organization users to authorize, inspect, revoke, or consume recovery.
- Logging or auditing authentication secrets, credentials, or raw provider errors.
- Treating the local `AUTH_OTP_BYPASS` setting as production recovery.
