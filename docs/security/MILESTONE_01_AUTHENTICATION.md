# Milestone 01 Authentication Implementation Note

Status: implemented local end-to-end slice, 2026-08-19

This is an implementation and verification note. It does not replace or amend the approved source-of-truth, architecture, or ADRs.

## Scope and traceability

Milestone 01 implements Super Admin password authentication, a task-authorized local-only OTP bypass, PostgreSQL-backed sessions, protected Dashboard identity, and logout. It intentionally does not implement email OTP delivery or verification, emergency recovery, or later Dashboard modules.

The controlling product behavior remains email, password, six-digit email OTP, then an authenticated session, as defined by:

- [Source-of-truth precedence](../source-of-truth/README.md)
- [Super Admin CP MVP Specification §3](../source-of-truth/AI_Sales_Agent_Super_Admin_CP_MVP_Spec_v1.0.md#3-authentication)
- The approved Super Admin Authentication amendment (§C) in the Master PRD, BRD, and Stage 1 baseline

The implementation follows `ARCH-PR-002` (application services are authoritative), `ARCH-PR-004` (deny by default), `ARCH-PR-005` (server-side security), and the authentication/browser/configuration controls in §§11–14, 31.3, and 37 of the [Technical Architecture](../architecture/AI_Sales_Agent_Technical_Architecture_v1.0.md). It also follows [ADR-0001](../adr/ADR-0001-go-modular-monolith.md) for the Go modular-monolith/PostgreSQL boundary and [ADR-0002](../adr/ADR-0002-required-redis-runtime.md) for required, non-authoritative, fail-closed Redis use.

## Implemented security boundaries

| Boundary | Implemented control |
| --- | --- |
| Password | Argon2id via `golang.org/x/crypto/argon2`; PHC encoding with version and parameters; 19 MiB memory, 2 iterations, parallelism 1; random 16-byte salt; bounded parser and constant-time hash comparison. Unknown accounts perform dummy-hash verification. |
| Login input | Explicit email/password DTO, 8 KiB body limit, JSON content-type requirement, unknown-field rejection, one-object enforcement, bounded email/password validation, and parameterized PostgreSQL queries. Unknown email, wrong password, and inactive account use the same credential error. |
| Login throttling | Redis fixed-window counters: 5 attempts per normalized email and 30 per requesting IP per 15 minutes. Email/IP identifiers are SHA-256-derived before entering Redis keys. One Lua script updates both counters and TTLs atomically. Missing or failed Redis returns authentication unavailable; it never permits login. |
| Proxy identity | `RemoteAddr` is authoritative unless the immediate peer is inside `AUTH_TRUSTED_PROXY_CIDRS`. A trusted `X-Forwarded-For` chain is evaluated right-to-left; malformed chains fall back to the known peer. Only actual controlled proxy networks may be configured. |
| Session | A new independent 32-byte `crypto/rand` secret is issued after authentication. The raw token appears only in the cookie; PostgreSQL stores its SHA-256 hash. Lookup checks expiry, revocation, and active Super Admin state. PostgreSQL is authoritative; Redis stores no session identity or authorization state. |
| Cookie | `sales_agent_session`; host-only with no `Domain`; `HttpOnly`; `SameSite=Strict`; `Path=/`; absolute expiry. `Secure=true` outside local development. `Secure=false` is limited to local loopback HTTP because browsers cannot send Secure cookies over that development origin. |
| CSRF | Login and logout require exact configured `Origin`. Logout additionally requires the independent `X-CSRF-Token` returned by the same-origin session API and held only in frontend runtime memory. Logout revokes PostgreSQL state before clearing the cookie. |
| API responses | Auth responses use `Cache-Control: no-store`, safe structured errors and correlation IDs. Password hashes, raw session tokens, database/Redis errors, and privileged fields are not returned. No wildcard CORS headers are added. |
| Browser state | Browser code uses cookies with `credentials: same-origin`; it stores no auth material in `localStorage` or `sessionStorage`. `/dashboard` resolves the real session API and redirects unauthenticated users for UX; the Go API remains authoritative. |

The schema is added by `backend/migrations/00002_super_admin_auth.sql`. `super_admin_accounts.email` is normalized and unique, password hashes are never browser-controlled, and `super_admin_sessions.token_hash` is constrained to a 32-byte digest. The first account is created only through `cmd/bootstrap-super-admin`, which requires an interactive terminal, reads and confirms the password without echo, enforces 12–128 characters, and never accepts a password flag.

## Local-only OTP bypass

`AUTH_OTP_BYPASS` defaults to `false`. The startup configuration rejects `true` unless `APP_ENV=local`; both `APP_ENV=test` and `APP_ENV=production` fail startup. The frontend cannot enable or request the bypass.

When the bypass is enabled locally, a correct password creates the normal PostgreSQL-backed session. When it is disabled, a correct password returns HTTP `428` with code `OTP_REQUIRED` and creates no session. The `local_development` response field drives only the visible local-environment badge; it grants no authentication capability.

This exception is limited to the milestone’s local browser journey and does not change production authentication behavior.

## Same-origin request path

Browser code calls `/api/v1/auth/*` on the Next.js origin. `apps/admin-web/next.config.ts` rewrites `/api/:path*` to the server-only `GO_API_ORIGIN`; the browser never calls the Go origin directly and no broad CORS policy is introduced.

For local development:

```text
Browser: http://127.0.0.1:3001
Next.js rewrite: /api/*
Go API: http://127.0.0.1:8081
```

`APP_ORIGIN` must exactly match the browser origin. `AUTH_TRUSTED_PROXY_CIDRS` must contain only the Next.js/reverse-proxy networks that are actually controlled; the local example trusts loopback only.

## Verification

Run the normal quality gates from the repository root:

```bash
./scripts/check-toolchain.sh
make backend-vet
make backend-test
pnpm install --frozen-lockfile
make frontend-typecheck
make frontend-test
make frontend-build
git diff --check
```

Database and Redis integration tests are opt-in safeguards against accidental external connections:

```bash
cd backend
TEST_DATABASE_URL="postgres://..." TEST_REDIS_URL="redis://..." go test ./...
APP_ENV=test TEST_MIGRATION_DATABASE_URL="postgres://user:password@127.0.0.1:5432/sales_agent_migration_test?sslmode=disable" go test ./internal/database -run TestMigrationsUpAndDown
```

The migration rollback test refuses to run unless `APP_ENV=test`, PostgreSQL is
on a loopback host, and the dedicated database name ends exactly with
`_migration_test`. The test destroys all schema objects managed by Goose in
that database.

For the real browser workflow, provision a local Super Admin, start PostgreSQL/Redis, the API, and the web app, then provide `E2E_ADMIN_EMAIL` and `E2E_ADMIN_PASSWORD` to:

```bash
pnpm --filter @sales-agent/admin-web test:e2e
```

The browser suite covers logged-out Dashboard rejection, generic wrong-password failure, login, refresh persistence, logout, and rejection after logout.

## Deferred items and residual risks

- Real OTP challenge generation, protected storage, email delivery, verification, resend/attempt controls, and deployment-level emergency recovery are the next authentication milestone.
- Sessions currently enforce absolute expiry only. `last_seen_at` is updated, but an idle-expiry policy is not yet enforced; architecture §13.5 requires both capabilities before production hardening is complete.
- Trusted-proxy parsing is implemented and tested, but production safety still depends on configuring only the actual proxy networks and confirming the deployment platform’s forwarding behavior. An empty list safely ignores forwarded headers; an incorrect broad list weakens client-IP throttling.
- The current response-header baseline includes anti-framing, no-sniff, and restrictive referrer controls. Production CSP, HSTS, and Permissions Policy remain deployment/frontend hardening work.
- Authentication-specific security-event monitoring and recovery audit evidence are not yet implemented. Secrets, passwords, CSRF values, and raw session tokens must remain excluded when observability is added.
- The auth repository uses explicit parameterized pgx queries. sqlc generation from the ADR-0001 database-access baseline has not yet been introduced.
