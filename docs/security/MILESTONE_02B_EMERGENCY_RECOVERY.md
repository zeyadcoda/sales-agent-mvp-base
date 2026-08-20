# Milestone 02B Emergency Super Admin Recovery

Status: implementation and validation note, 2026-08-20

This document records the Milestone 02B recovery design and operator runbook. It does not replace or amend the approved product source-of-truth, architecture, or ADRs. The controlling authentication behavior remains email, password, six-digit email OTP, and then a PostgreSQL-backed Super Admin session. The architectural decision is recorded in [ADR-0003](../adr/ADR-0003-super-admin-emergency-recovery.md).

## Scope

Milestone 02B adds only the deployment-level mechanism needed to recover the sole Super Admin when the configured notification-email provider is unavailable.

It adds:

- a deployment-only `platform-admin auth-recovery` CLI;
- a PostgreSQL-authoritative, 10-minute, one-time recovery authorization;
- recovery consumption after successful password verification;
- normal PostgreSQL-backed session creation without email OTP for that one login;
- a minimum reusable, append-only Platform Audit foundation;
- safe status, revocation, and audit inspection commands;
- database lifecycle/concurrency, provider-outage, rate-limit, Audit, and CLI tests.

It does not add a recovery page, HTTP endpoint, request field, browser identifier, user-entered recovery code, reusable token, permanent OTP bypass, special session, Platform Support role, or Agent capability. It does not add the Audit UI or any other Super Admin business module.

**The Super Admin frontend is unchanged.** The existing Login, OTP, session cookie, Dashboard protection, and logout behavior remain the browser experience.

## Threat being solved

The MVP has one Super Admin and normally depends on email OTP delivery. A configured provider outage, rejected credentials at send time, or local Mailpit outage can prevent OTP delivery even when the Super Admin knows the correct password. Without a deployment-level alternative, that dependency could permanently lock out the only internal operator.

Recovery is intentionally narrower than password reset. It does not help someone who has forgotten the password, and it does not authenticate a caller merely because email is down. It lets an already-authorized deployment operator grant one short window in which the next successful password login for one exact Super Admin may skip OTP once.

## Deployment authority boundary

The recovery CLI is authorized by existing deployment access, not by a browser role. The operator must already be able to run the deployed command with server-side configuration and database connectivity.

Production must restrict that capability through deployment/IAM controls. Recommended controls include:

- named operator identities with strong authentication and MFA;
- least-privilege, just-in-time access to the application runtime and database network;
- protected bastion, deployment job, or equivalent controlled execution path;
- independent infrastructure logging of who invoked the command;
- incident/change-ticket linkage in the operator label or reason where appropriate;
- no database URL, password, token, or other credential in command arguments or shell history.

The operator label and reason are written to Platform Audit. They must be meaningful, but they must not contain passwords, OTPs, session tokens, HMAC keys, SMTP credentials, database credentials, API keys, or other secrets.

Agents, prompts, frontend code, Organization users, and the authenticated Super Admin browser cannot invoke the CLI. No application HTTP permission represents deployment authority.

## Recovery lifecycle

The server derives lifecycle state from PostgreSQL timestamps and terminal fields. Recovery persistence materializes `clock_timestamp()` after acquiring the account lock, so a request that waited for another account operation cannot reuse a transaction-start timestamp. Application, CLI input, operator, and browser clocks are not authoritative.

| State | Meaning |
| --- | --- |
| `ACTIVE` | The authorization belongs to the exact Super Admin, `expires_at` is after PostgreSQL's current time, and `consumed_at`, `revoked_at`, and `expired_at` are all unset. It may be consumed once after a correct password. |
| `EXPIRED` | `expired_at` is set or `expires_at` is no longer after PostgreSQL's current time. It is permanently unusable and cannot be renewed. |
| `CONSUMED` | A password login used the authorization and created its one normal session. It cannot be replayed. |
| `REVOKED` | A deployment operator revoked the unused authorization. It cannot be consumed or reactivated. |

The validity is 10 minutes. There is no duration flag, extension command, renewal, or reusable recovery material. A new authorization is a separate reasoned and audited operation.

At most one unresolved authorization can exist for a Super Admin. Before a replacement is created, an elapsed unresolved row is terminalized by setting `expired_at`; the database's partial unique index then prevents concurrent authorization commands from leaving two unresolved or usable grants. An expired record is never selected as active.

## PostgreSQL model

Goose migration `00004_super_admin_emergency_recovery.sql` creates `super_admin_recovery_authorizations`, which stores only server-side recovery state:

- opaque event/authorization ID;
- exact `super_admin_id` with a restrictive account relationship;
- `created_at` and `expires_at`;
- nullable, mutually exclusive `consumed_at`, `revoked_at`, and `expired_at` terminal timestamps;
- mandatory reason;
- mandatory deployment/operator identifier;
- correlation ID.

It stores no plaintext password, OTP, recovery code, token, session token, CSRF token, HMAC key, provider credential, database credential, or other secret. The authorization ID is not sent to the browser and is not an authentication credential.

Constraints enforce valid timestamp ordering, a maximum lifetime of 10 minutes, and no more than one terminal state. A partial unique index permits at most one unresolved row per Super Admin, while active lookups additionally require `expires_at` to be after the PostgreSQL operation time. Every consuming SQL mutation rechecks that the row is unconsumed, unrevoked, unexpired, and owned by the locked account.

## Authorization creation

The `authorize` application operation performs the following sequence:

1. Validate and normalize the email, reason, and operator label.
2. Display the exact target, reason, operator, 10-minute validity, and one-time password-login impact.
3. Require interactive confirmation; `y` or `yes` confirms case-insensitively, while empty input, EOF, or every other response declines the operation. There is no `--yes` override.
4. Begin a PostgreSQL transaction and lock the exact Super Admin account first.
5. Reject an unknown or ineligible account safely.
6. Resolve any existing authorization state and enforce at most one active grant.
7. Materialize `clock_timestamp()` after the account lock and insert the new authorization with `expires_at` exactly 10 minutes later.
8. Append `SUPER_ADMIN_RECOVERY_AUTHORIZED` to Platform Audit with the same reason, operator, target reference, result, and correlation ID.
9. Commit the authorization and Audit Event together.

The CLI prints a safe status and expiry. It prints no code, token, password, session value, credential, or secret.

## Password login and consumption

Recovery is integrated into the existing `POST /api/v1/auth/login` use case; no recovery API is added.

The effective sequence when `AUTH_OTP_BYPASS=false` is:

```text
exact Origin and request validation
-> Redis email/IP password-login throttle
-> account lookup or dummy Argon2id verification
-> account-state and password verification
-> attempt one atomic PostgreSQL recovery consumption
   -> if consumed: normal session + Audit Event -> existing session cookie -> Dashboard
   -> if not consumed: normal OTP challenge -> email -> OTP verification -> normal session
```

Unknown email, wrong password, and ineligible account behavior remains the existing generic credential failure. Recovery state is not consumed, changed, or disclosed. A wrong password does not become a recovery operation and does not cause an authorization-specific browser response.

After a correct password and an active-row read hint, the application generates the same independent random raw session and CSRF material used by normal authentication. The raw session token is hashed before persistence. The recovery transaction then:

1. locks the Super Admin account first;
2. locks/selects the one eligible authorization for that account;
3. materializes `clock_timestamp()` after the account lock and rechecks expiry, consumption, and revocation;
4. marks the authorization consumed;
5. invalidates older pending OTP challenges for the same account;
6. inserts the normal session state containing only the SHA-256 session-token hash;
7. appends `SUPER_ADMIN_RECOVERY_CONSUMED` with safe references, original reason/operator attribution, result, and the login correlation ID;
8. commits consumption, OTP invalidation, session insertion, and Audit Event atomically.

Only after commit does the existing handler set the existing HttpOnly session cookie. The raw session token is never stored in PostgreSQL, Audit, logs, or a JSON response. Recovery does not reuse an OTP challenge or authorization identifier as the session secret.

The recovery-created session is not special. It has the same expiry, lookup, CSRF, cookie, revocation, Dashboard, and logout behavior as a session created after OTP verification.

## Concurrency behavior

All operations that lock both account and authorization rows use the same account-first order.

If two correct-password requests race for one authorization, PostgreSQL serializes them. Only the first transaction that still sees an active row can mark it consumed, insert the recovery-created session, and append the consumption Audit Event. The other request cannot reuse the row and continues through normal OTP behavior or fails safely.

Consequently:

- one authorization creates at most one recovery-authenticated session;
- consumption and session creation cannot commit separately;
- a failed/rolled-back session insertion leaves the authorization unconsumed;
- a failed/rolled-back Audit insertion leaves the authorization and session uncommitted;
- a browser disconnect after commit does not make the authorization reusable.

## Email-provider outage behavior

A valid recovery authorization is created entirely through PostgreSQL and does not require Mailpit or SMTP.

After correct password verification, a successfully consumed authorization bypasses OTP challenge generation and does not call the email sender. A runtime provider outage therefore cannot prevent that one login.

Without a valid authorization, the existing OTP path remains mandatory. If email delivery fails, the challenge remains unusable according to Milestone 02 fail-closed behavior, no session is created, and the caller is not authenticated.

After consumption, the next login immediately returns to normal OTP behavior. An email outage by itself never authenticates anyone.

## Rate limiting and dependency failure

Recovery does not bypass the existing Redis password-login throttle. The throttle runs before account lookup, password verification, or recovery consumption and remains 5 attempts per normalized email and 30 per requesting IP in a 15-minute fixed window.

Redis remains a required, non-authoritative dependency. A missing, malformed, failed, or timed-out Redis operation returns authentication unavailable before recovery state is consumed or a session is created. The authorization remains in PostgreSQL and may be used later if it is still unexpired.

The deployment CLI uses PostgreSQL for durable recovery and Audit state; it does not move authorization state into Redis or create a permissive in-memory fallback. Deployment/IAM access, confirmation, exact-account locking, the single-active-grant constraint, and immutable Audit are the CLI abuse controls. Browser password attempts remain Redis-throttled regardless of the grant.

## Relationship to the local OTP bypass

`AUTH_OTP_BYPASS` and emergency recovery are independent mechanisms.

| Mechanism | Purpose | Environment | Persistence | Audit/reason | Use |
| --- | --- | --- | --- | --- | --- |
| `AUTH_OTP_BYPASS` | Local developer convenience | Only `APP_ENV=local` | Configuration only | No emergency authorization | Every correct local password login while intentionally enabled |
| Emergency recovery | Provider-outage security operation | May be used in production | One PostgreSQL authorization | Mandatory operator/reason and Audit | One correct password login within 10 minutes |

Startup continues to reject `AUTH_OTP_BYPASS=true` in test and production. When the local bypass is intentionally enabled, it follows its existing branch and does not inspect or consume an emergency authorization. Recovery behavior is exercised with `AUTH_OTP_BYPASS=false`.

## Platform Audit foundation

Migration `00004_super_admin_emergency_recovery.sql` also creates the one platform-wide append-only `platform_audit_events` source. Recovery appends to that reusable table; no recovery-specific audit silo is created.

The reusable record supports:

- opaque event ID;
- actor type and actor identifier/label;
- server timestamp;
- action;
- resource type and safe reference;
- nullable Organization reference;
- safe old/new structured values or references where useful;
- mandatory recovery reason;
- result;
- correlation ID.

Recovery actions include:

- `SUPER_ADMIN_RECOVERY_AUTHORIZED`;
- `SUPER_ADMIN_RECOVERY_CONSUMED`;
- `SUPER_ADMIN_RECOVERY_REVOKED` when an active authorization is explicitly revoked;
- `SUPER_ADMIN_RECOVERY_AUTHORIZATION_FAILED` for a protected authorization attempt rejected because the target is ineligible or an active grant already exists;
- `SUPER_ADMIN_RECOVERY_REVOCATION_FAILED` for a protected revocation attempt rejected because the target is ineligible or no active grant exists.

Those protected failed deployment operations append a sanitized `FAILURE` result in the same transaction that resolves the failure. CLI parsing errors, missing flags, and declined confirmation never open the deployment services and do not pretend a protected database mutation occurred. Wrong browser passwords retain normal authentication failure behavior and do not disclose or consume recovery state.

Audit records are created only by trusted application services and CLI operations. Application behavior provides no update/delete path, frontend write API, Agent tool, or individual deletion operation. Database triggers reject every `UPDATE`, `DELETE`, or `TRUNCATE` against `platform_audit_events`. Recovery state changes and their successful Audit Events commit in the same database transaction.

Audit never contains passwords, OTPs, raw session/CSRF tokens, recovery codes or tokens, HMAC keys, SMTP/API/database credentials, secret values, or raw provider errors. Reasons and safe structured values are bounded and treated as untrusted display text.

## CLI runbook

The examples below are run from the repository root. The local examples load the existing `.env` into a subshell without modifying it, then run the command from the Go module in `backend`. When the subshell exits, its exported values disappear. The CLI itself requires `DATABASE_URL`; do not pass a password, database URL, or other credential as a flag.

### Authorize one recovery login

```bash
(
  set -a
  . ./.env
  set +a
  cd backend
  go run ./cmd/platform-admin auth-recovery authorize \
    --email admin@example.com \
    --reason "notification provider outage" \
    --operator "local-developer"
)
```

The command shows this impact summary before opening deployment services:

```text
Emergency Super Admin recovery authorization
Target Super Admin: admin@example.com
Reason: notification provider outage
Deployment operator: local-developer
Recovery validity: 10 minutes
Behavior: next successful password login bypasses email OTP once
No recovery code, token, or browser secret will be created.
Confirm? [y/N]
```

Type `y` or `yes` only after checking the target, reason, and operator. The command creates no authorization on the default response, and it has no non-interactive confirmation flag. On success it prints only the target, RFC 3339 expiry, and Audit correlation ID; it never prints the internal authorization ID or authentication material.

### Check status

```bash
(
  set -a
  . ./.env
  set +a
  cd backend
  go run ./cmd/platform-admin auth-recovery status \
    --email admin@example.com
)
```

Status reports safe metadata such as `ACTIVE`, `EXPIRED`, `CONSUMED`, `REVOKED`, or no authorization, plus timestamps where applicable. It prints no authentication material.

### Revoke an unused authorization

```bash
(
  set -a
  . ./.env
  set +a
  cd backend
  go run ./cmd/platform-admin auth-recovery revoke \
    --email admin@example.com \
    --reason "provider restored before recovery was used" \
    --operator "local-developer"
)
```

Revocation shows the exact target and impact and requires interactive `y` or `yes` confirmation. It atomically marks the active authorization revoked and appends `SUPER_ADMIN_RECOVERY_REVOKED`. It does not revoke an already consumed authorization or make an expired authorization usable.

### Inspect recovery Audit Events safely

```bash
(
  set -a
  . ./.env
  set +a
  cd backend
  go run ./cmd/platform-admin auth-recovery audit \
    --email admin@example.com
)
```

This is the preferred local verification command for developers who do not know SQL. It lists at most the 50 newest matching events and displays timestamp, action, result, actor type/identifier, reason, and correlation ID without displaying internal structured values, credentials, OTPs, raw session values, or other secrets.

In production, use the deployment platform's approved secret injection and command execution mechanism instead of sourcing a local `.env` file.

## Beginner-friendly Mailpit outage test

Prerequisites:

- the latest migrations are applied;
- `admin@example.com` is an active local Super Admin and its password is known;
- PostgreSQL and Redis are running;
- the web application is running at `http://127.0.0.1:3001` (run `make web` in its own terminal if needed);
- no previous testing has already exhausted the Redis login limit for the email/IP. If it has, wait for the 15-minute window to expire rather than disabling the limiter.

For a fresh local checkout, `make infra-up` starts PostgreSQL, Redis, and Mailpit, and `make migrate-up` applies the migrations. If the account does not exist yet, `make bootstrap-super-admin EMAIL=admin@example.com` prompts for its password without modifying `.env` or accepting the password as an argument.

Perform these steps exactly:

1. Run the API with `AUTH_OTP_BYPASS=false` without editing `.env`. Stop any existing API process, then run from the repository root:

   ```bash
   (
     set -a
     . ./.env
     set +a
     export AUTH_OTP_BYPASS=false
     cd backend
     go run ./cmd/api
   )
   ```

2. In another terminal, stop Mailpit to simulate a notification-provider outage:

   ```bash
   docker compose -f infra/compose/docker-compose.yml stop mailpit
   ```

3. Open `http://127.0.0.1:3001/login` and attempt a normal login for `admin@example.com` with the correct password. The email send must fail safely. The browser must not reach Dashboard and no authenticated session may be created.

4. In a terminal, authorize one recovery login with the mandatory reason and operator label:

   ```bash
   (
     set -a
     . ./.env
     set +a
     cd backend
     go run ./cmd/platform-admin auth-recovery authorize \
       --email admin@example.com \
       --reason "manual test: Mailpit outage" \
       --operator "local-developer"
   )
   ```

   Review the impact summary and type `y`. No recovery code or secret is printed.

5. Attempt Login again with the correct password before the 10-minute expiry. The browser must reach Dashboard directly without showing the email OTP page. This is a normal session using the existing HttpOnly cookie.

6. Logout through the existing Dashboard/logout control. The server-side session must be revoked normally.

7. While Mailpit remains stopped, attempt Login again with the same correct password. The consumed authorization must not work again. Email delivery must fail safely, and the browser must not authenticate.

8. Restart Mailpit:

   ```bash
   docker compose -f infra/compose/docker-compose.yml up -d mailpit
   ```

9. Login again with the correct password. The browser must follow the normal password to OTP flow. Open `http://127.0.0.1:8025`, read the latest six-digit OTP, verify it, and confirm Dashboard loads. Logout when finished.

After the exercise, inspect the recovery events without SQL:

```bash
(
  set -a
  . ./.env
  set +a
  cd backend
  go run ./cmd/platform-admin auth-recovery audit \
    --email admin@example.com
)
```

The output must contain the authorization and one consumption event, preserve `manual test: Mailpit outage` and `local-developer`, and contain no password, OTP, session token, recovery secret, or SMTP credential. `status --email admin@example.com` must report the authorization as consumed rather than active.

## Browser and API visibility

Attackers are assumed to know every browser API URL and request shape. Recovery adds no route for them to discover or call.

- The Login request remains only email and password.
- The browser sends no recovery flag, authorization ID, operator, reason, or code.
- The Login page, OTP page, Dashboard, and browser storage contain no recovery identifier.
- DevTools shows only the ordinary Login response: either OTP is required, authentication is unavailable/fails, or the existing normal session response succeeds.
- There is no frontend or Agent client for the recovery tables or CLI.

The recovery decision is an internal branch after correct password verification. Knowing that recovery exists does not provide the deployment authority, password, active PostgreSQL state, or ability to bypass Redis throttling.

## Security verification

The implementation and its tests must explicitly prove:

- no recovery HTTP endpoint, UI, browser field, route, or identifier exists;
- no password, OTP, raw session/CSRF value, recovery material, or provider/database credential is logged or audited;
- wrong/unknown credentials neither consume nor disclose recovery;
- account state and existing password/login throttling still apply;
- Redis failure prevents recovery authentication and leaves the authorization unconsumed;
- authorization is tied to exactly one Super Admin and expires after 10 minutes;
- consumed, expired, and revoked authorizations cannot be reused;
- two concurrent correct logins create at most one recovery-authenticated session;
- consumption, normal hashed-session insertion, and Audit append are atomic;
- valid recovery never calls the unavailable email sender;
- no authorization plus email failure never creates a session;
- local OTP bypass remains rejected outside local and does not consume recovery;
- Agents and frontend code have no recovery or generic Audit-write capability;
- authorization, consumption, and revocation preserve safe reason/operator attribution.

## Verification commands

Run formatting over the changed Go files, then run the normal repository gates from the repository root. (`gofmt -w .` does not recursively format a Go module, so pass the changed `.go` files to `gofmt -w` explicitly.) The final delivery report records the exact commands and results.

```bash
./scripts/check-toolchain.sh
make backend-vet
make backend-test
make frontend-typecheck
make frontend-test
make frontend-build
git diff --check
```

The focused unit/package run is:

```bash
(
  cd backend
  go test ./cmd/platform-admin ./internal/platform/audit ./internal/platform/auth ./internal/httpapi ./internal/requestmeta -count=1
)
```

This includes the CLI confirmation/secret-output tests, safe Audit validation, wrong-password behavior, one-time and provider-outage recovery, race-loser behavior, Redis fail-closed behavior, normal hashed-session creation, local-bypass separation, and the explicit no-recovery-route case in `TestAuthRouteErrorsRetainJSONNoStoreAndAllowSemantics`.

PostgreSQL integration uses a migrated, disposable loopback database through `TEST_DATABASE_URL`. The following exact recovery tests cover authorization/expiry/replacement/consumption/revocation and Audit, failed-target Audit, rollback atomicity, database time after lock waiting, inactive-account rechecks, and simultaneous consumption:

- `TestPostgresRecoveryAuthorizationLifecycleAndAudit`
- `TestPostgresRecoveryUnknownTargetFailureIsAudited`
- `TestPostgresRecoveryConsumptionRollsBackWhenSessionInsertFails`
- `TestPostgresRecoveryUsesDatabaseTimeAfterWaitingForAccountLock`
- `TestPostgresRecoveryCannotConsumeAfterAccountDeactivation`
- `TestPostgresRecoveryConcurrentConsumptionCreatesOneSessionAndAudit`

Run them, together with the Platform Audit round trip, from the repository root:

```bash
(
  cd backend
  TEST_DATABASE_URL='<dedicated migrated loopback PostgreSQL URL>' \
    go test ./internal/platform/auth ./internal/platform/audit \
      -run '^(TestPostgresRecovery|TestPostgresAuditAppendAndListRoundTrip)' \
      -count=1 -v
)
```

The migration test performs the Goose chain's up/down/up validation and deliberately destroys the schema in its target. It refuses to run unless `APP_ENV=test`, the host is loopback, and the database name ends exactly in `_migration_test`:

```bash
(
  cd backend
  APP_ENV=test \
  TEST_MIGRATION_DATABASE_URL='<dedicated loopback PostgreSQL URL ending in _migration_test>' \
    go test ./internal/database -run '^TestMigrationsUpAndDown$' -count=1 -v
)
```

Run the real Redis counter integration separately; service tests still prove that a Redis error prevents recovery inspection or consumption:

```bash
(
  cd backend
  TEST_REDIS_URL='redis://127.0.0.1:6379/0' \
    go test ./internal/cache \
      -run '^(TestRedisPing|TestIncrementLoginAttemptCountersWithRedis)$' \
      -count=1 -v
)
```

The frontend is unchanged. With the API running against the same dedicated migrated E2E database, `AUTH_OTP_BYPASS=false`, Redis and Mailpit available, and the web application at `http://127.0.0.1:3001`, run the existing real-backend OTP regression suite exactly as documented by Milestone 02:

```bash
PLAYWRIGHT_BASE_URL=http://127.0.0.1:3001 \
E2E_AUTH_MODE=otp \
E2E_ADMIN_EMAIL=e2e-admin@example.test \
E2E_ADMIN_PASSWORD='<the disposable E2E account password>' \
E2E_MAILPIT_ORIGIN=http://127.0.0.1:8025 \
E2E_DATABASE_URL='<the same dedicated loopback *_integration_test PostgreSQL URL used by the API>' \
pnpm --filter @sales-agent/admin-web test:e2e
```

That suite is the regression proof for Login, OTP, session refresh, Dashboard protection, expiry, and logout. The manual Mailpit exercise above is the recovery-specific browser validation.

## Remaining production assumptions and risks

- The CLI cannot recover a forgotten password and deliberately still requires the correct password.
- PostgreSQL, Redis, the API, and deployment access must be available. Recovery is not a database, cache, network, or whole-deployment disaster-recovery system.
- The API still requires structurally valid startup configuration. Recovery handles a configured provider that fails at send time; it cannot help if invalid/missing mandatory configuration prevents the API from starting.
- Production access to `platform-admin`, server-side configuration, and database connectivity is a high-risk deployment capability. Final IAM, MFA, bastion/job, approval, and infrastructure-log controls are deployment responsibilities.
- A compromised database owner or deployment administrator is outside the protection provided by application-level append-only behavior. Least-privilege database roles, infrastructure audit, backup protection, and operational monitoring remain necessary.
- `AUTH_OTP_HMAC_SECRET`, SMTP credentials, and other production secrets still need the approved managed `SecretReference`/vault integration, least-privilege access, rotation, and redaction controls described in Milestone 02.
- Production SMTP identity, deliverability, provider monitoring, and incident response still require deployment validation.
- Session idle-expiry policy remains separate Milestone 01/02 production hardening. Recovery does not alter the existing absolute session lifetime.
- Platform Audit retention, archival, backup/restore, and database-owner tamper protections require later production policy and infrastructure work. No Audit UI or export exists in this milestone.
- Recovery reasons are audited operational data. Operators must avoid placing secrets or unnecessary personal data in them.

Subject to successful automated and manual validation, this milestone closes the notification-provider lockout gap without changing the approved browser authentication experience or expanding product scope.
