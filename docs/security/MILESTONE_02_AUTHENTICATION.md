# Milestone 02 Authentication Implementation Note

Status: real email OTP flow implemented, 2026-08-19. Local/manual validation is supported; production release remains blocked on the deployment-level emergency recovery control described below.

This document records the Milestone 02 implementation. It does not replace or amend the approved product source-of-truth, architecture, or ADRs. The controlling requirements remain the approved Super Admin authentication behavior and §§13–14 of the [Technical Architecture](../architecture/AI_Sales_Agent_Technical_Architecture_v1.0.md).

## Scope

Milestone 02 implements the single-Super-Admin authentication path:

```text
email + password
-> password verification
-> six-digit email OTP challenge
-> OTP verification
-> new PostgreSQL-backed authenticated session
-> Dashboard
```

It also preserves the Milestone 01 local-only OTP bypass, session lookup, protected Dashboard, and logout. It does not add Organizations, Applications, Packages, Agents, Dashboard business modules, a normal UI recovery bypass, or a general notification-email configuration module.

## Authentication sequence

With `AUTH_OTP_BYPASS=false`, authentication proceeds as follows:

1. `POST /api/v1/auth/login` validates the exact browser origin and a strict email/password request. Redis consumes the password-attempt counters before PostgreSQL account lookup and Argon2id password verification.
2. A valid password does not create a session. The service creates a 256-bit random challenge identifier and an exactly six-digit OTP using `crypto/rand`.
3. PostgreSQL stores an inactive delivery version containing only the challenge-bound HMAC of the OTP. The plaintext OTP exists only in application memory while the email adapter builds and delivers the message.
4. The sender delivers the OTP through the narrow `OTPEmailSender` port. Only after delivery succeeds does PostgreSQL conditionally activate that delivery version.
5. Login returns HTTP `202` with `authentication_state: "OTP_REQUIRED"` and browser-safe challenge metadata. The frontend clears the password, stores only the opaque challenge context in `sessionStorage`, and navigates to `/verify-otp`.
6. The OTP page validates its stored context through the backend status endpoint. It submits only `challenge_id` and `otp` to the verification endpoint.
7. PostgreSQL locks the challenge row and authoritatively checks activation, account state, expiry, attempt count, invalidation, consumption, and the candidate HMAC. On success, one transaction consumes the challenge and inserts a new session.
8. The API places the new independent session secret only in the existing secure HttpOnly cookie. The JSON response contains the safe session view and CSRF token, never the raw session token.

A password match, challenge identifier, frontend route, or successful email send is not authentication. Only the committed OTP-consumption/session-insert transaction establishes an authenticated Super Admin session.

## Challenge data and cryptography

Migration `backend/migrations/00003_super_admin_auth_challenges.sql` adds `super_admin_auth_challenges` without changing historical migrations. Each challenge belongs to exactly one `super_admin_accounts` row through a restrictive foreign key.

The browser-visible challenge ID is 32 bytes from `crypto/rand`, encoded as a canonical 43-character unpadded Base64URL string. It is independent from the account ID, OTP, and eventual session token. The schema prevents sequential or malformed identifiers and includes a partial unique index for one non-terminal challenge per Super Admin. Challenge creation also locks the account row and invalidates any earlier current challenge, so concurrent password successes cannot leave two usable flows.

OTPs are generated uniformly from `000000` through `999999` with `crypto/rand.Int`; leading zeroes remain valid. Plaintext OTPs are never inserted into PostgreSQL, returned by an API, written to browser storage, or intentionally logged.

The stored value is HMAC-SHA-256 over a domain separator, the challenge ID, and the six-digit OTP. The key is the server-only `AUTH_OTP_HMAC_SECRET`, decoded from at least 32 bytes of Base64-encoded key material. Binding the low-entropy code to both an unguessable challenge and a server-held secret prevents a database-only compromise from testing the one-million-code space. Verification rejects malformed digests and uses `hmac.Equal` after exact-length checks.

The challenge table records creation and expiry, failed attempts, resend availability, delivery and active versions, activation, consumption, and invalidation. The OTP hash, account relationship, failure count, and delivery internals are never serialized to the browser.

## Lifecycle and delivery state

PostgreSQL is authoritative for every lifecycle decision. Server timestamps are used; frontend timers are informational.

| State | Meaning and permitted transition |
| --- | --- |
| Delivery pending | The new hash/version is stored with no active version. It cannot verify. A matching successful delivery activation makes it pending for user input; any delivery or activation failure invalidates it fail closed. |
| `PENDING` | The current delivery version is active, unexpired, not consumed or invalidated, and has fewer than five failed attempts. It may be verified or, after the cooldown, resent. |
| `EXPIRED` | `expires_at` is no longer after the server time. It cannot verify or resend; the user must restart login. |
| `ATTEMPTS_EXCEEDED` | Five wrong codes have been committed. The fifth attempt invalidates the challenge, and every later attempt fails even if it supplies the former correct code. |
| `INVALIDATED` | A newer login flow, delivery failure, or other terminal transition made the challenge unusable. |
| `CONSUMED` | Successful verification has already created its one session. The challenge cannot be replayed. |

Each delivered code is valid for 10 minutes. Wrong-code updates run while the row is locked and increment `failed_attempts` in the same transaction. Attempts one through four return the same generic invalid-code result; the fifth returns the attempts-exceeded result and makes the challenge terminal. Correct verification does not increment the counter. The counter is not reset by resend, so resend cannot evade the five-attempt cap.

Resend is available only after 60 seconds according to the persisted server timestamp. Its transaction locks and revalidates the challenge, replaces the HMAC, increments the delivery version, renews the 10-minute validity/cooldown timestamps, and clears activation before committing. This makes the previous OTP unusable before email I/O begins. Successful delivery conditionally activates only the same new version. If sending or activation fails, the pending version is invalidated; the previous OTP is never revived and no session is created.

Verification locks the associated account and then the challenge row with `FOR UPDATE`, checks the active delivery version, performs the comparison, marks the challenge consumed, and inserts the session before one commit. Concurrent verification requests serialize on those locks: at most one can observe an active, unconsumed challenge and commit a session.

## Browser challenge handling

The login page never puts the password in a URL, cookie, `localStorage`, or `sessionStorage`, and clears its password state after the request. The OTP page stores only:

- opaque `challenge_id`;
- `expires_at`;
- `resend_available_at`;
- a masked destination hint such as `a***@example.com`.

The challenge ID is intentionally not placed in the URL. Possession of it grants no account identity or authenticated capability; the valid OTP and all server-side lifecycle checks are still required.

Direct `/verify-otp` access without stored pending context redirects to `/login`. On page load or refresh, the browser asks the backend for current status rather than trusting stored timestamps. Expired, attempts-exceeded, invalidated, and consumed challenges render controlled terminal states with a restart-login action. Backend checks remain authoritative if a caller skips the frontend, modifies storage, copies a network request, or calls the endpoints directly.

## HTTP API

All authentication responses use `Cache-Control: no-store`. Login and OTP endpoints require an exact configured `Origin`; no wildcard CORS policy is introduced. Authentication request bodies require `application/json`, are limited to 8 KiB, reject unknown fields and trailing JSON values, and receive authoritative backend format validation.

| Endpoint | Request | Successful result |
| --- | --- | --- |
| `POST /api/v1/auth/login` | `email`, `password` | HTTP `202` plus `OTP_REQUIRED` challenge metadata when real OTP is enabled; HTTP `200` plus the normal session response only for the allowed local bypass. |
| `POST /api/v1/auth/otp/status` | `challenge_id` | Browser-safe metadata and one of `PENDING`, `EXPIRED`, `ATTEMPTS_EXCEEDED`, `INVALIDATED`, or `CONSUMED`. This POST shape keeps the identifier out of URLs and referrers. |
| `POST /api/v1/auth/otp/verify` | `challenge_id`, `otp` | Consumes the challenge, creates the server session, sets the session cookie, and returns the safe session response. |
| `POST /api/v1/auth/otp/resend` | `challenge_id` | Rotates/delivers the code and returns updated browser-safe challenge metadata. |
| `GET /api/v1/auth/session` | Session cookie | Resolves the authoritative PostgreSQL session for Dashboard refresh and route UX. |
| `POST /api/v1/auth/logout` | Session cookie and `X-CSRF-Token` | Revokes the PostgreSQL session, then clears the cookie. |

The browser never supplies a Super Admin ID or bypass flag. The backend resolves challenge → account exclusively from PostgreSQL. Unknown or modified challenge IDs fail without disclosing account ownership. Safe application error codes distinguish the UX states while SQL, Redis, crypto, and SMTP/provider errors collapse to `AUTHENTICATION_UNAVAILABLE` and are not returned verbatim.

Pre-authentication OTP endpoints cannot require a session CSRF token because no authenticated session exists yet. They require exact Origin validation, strict request formats, opaque challenge state, PostgreSQL lifecycle enforcement, and Redis abuse controls. Logout retains both exact Origin and synchronizer-CSRF validation. Cookie controls remain `HttpOnly`, `SameSite=Strict`, `Path=/`, no `Domain`, and `Secure=true` outside local loopback development.

## Rate limiting and authoritative state

The existing Redis login throttle remains 5 attempts per normalized email and 30 per requesting IP in a 15-minute fixed window. OTP verification and resend use separate hashed Redis keyspaces, each limited to 5 requests per challenge and 30 per IP in 15 minutes. The challenge ID, normalized email, and IP are SHA-256-derived before entering Redis keys; each challenge/IP pair is incremented atomically.

Redis is required layered abuse protection, not the OTP state store. Missing, malformed, or failed Redis operations return a temporary authentication failure and cannot permit password verification, OTP verification, resend, session creation, or a cooldown bypass.

PostgreSQL remains authoritative for:

- the five failed-code attempts;
- 10-minute expiry;
- 60-second resend cooldown;
- active delivery version and old-code invalidation;
- challenge ownership and Super Admin state;
- challenge consumption and replay prevention;
- authenticated sessions and logout revocation.

## Email delivery adapters

Authentication depends only on `OTPEmailSender.SendOTP(context.Context, OTPEmail)`. The message contains the recipient, optional display name, six-digit code, and expiry. It contains no password, session token, CSRF token, challenge/account ID, database value, or provider secret.

Two adapters implement the port:

- `MailpitSender` is restricted to local/test, unauthenticated plaintext SMTP at exactly `127.0.0.1:1025`.
- `SMTPSender` is the provider-ready production boundary. It supports STARTTLS or direct TLS with TLS 1.2 or later. Production configuration requires authenticated SMTP credentials and rejects plaintext transport.

The authentication service does not import a vendor SDK. Transport errors are collapsed at the adapter boundary so SMTP replies and provider topology cannot reach clients. A delivery failure creates no session, returns the controlled unavailable response, and leaves the challenge deterministically inactive/invalidated. There is no silent local bypass or login fallback.

The email is deliberately simple: it identifies Super Admin authentication, includes the six-digit code, states an approximately 10-minute expiry, optionally greets the configured display name, and tells the recipient to ignore an uninitiated login.

## Local Mailpit flow

The development Compose service pins Mailpit and binds both ports to loopback only:

```text
SMTP:  127.0.0.1:1025
Inbox: http://127.0.0.1:8025
```

Mailpit is a local/test capture tool, not a production dependency. `APP_ENV=test` requires the Mailpit mode so automated tests cannot accidentally select a production-capable SMTP adapter. Start and inspect it with:

```bash
make mailpit-up
make mailpit-logs
```

`make infra-up` starts PostgreSQL, Redis, and Mailpit together. With Mailpit running, `make mailpit-test` exercises the real local SMTP adapter and confirms the inbox exposes a message containing a six-digit code.

For a browser test, apply migrations and bootstrap the account, run the API and web application, open `http://127.0.0.1:3001/login`, then read the code at `http://127.0.0.1:8025`. The browser should reach `/verify-otp`, accept the current code once, reach `/dashboard`, survive a Dashboard refresh through the real session API, and lose access after logout.

## Configuration

The real `.env` is intentionally not managed by this milestone. `.env.example` documents these server-side values:

| Variable | Contract |
| --- | --- |
| `APP_ENV` | `local`, `test`, or `production`; controls the bypass, email adapter, HTTPS, and cookie guards. |
| `AUTH_OTP_BYPASS` | `false` for real OTP. `true` is accepted only with `APP_ENV=local`; test and production refuse to start. The browser cannot set it. |
| `AUTH_OTP_HMAC_SECRET` | Required when bypass is false. Base64-encoded key material decoding to at least 32 bytes. Generate locally with `openssl rand -base64 32`; never commit or log it. |
| `AUTH_EMAIL_MODE` | `mailpit` or `smtp`. Test requires `mailpit`; production requires `smtp`. |
| `AUTH_EMAIL_FROM_ADDRESS` | Valid bare sender mailbox without a display name. |
| `AUTH_EMAIL_FROM_NAME` | Required 1–100 character display name without control characters. |
| `AUTH_SMTP_HOST` | Hostname or IP without scheme, port, path, or credentials. Mailpit requires `127.0.0.1`. |
| `AUTH_SMTP_PORT` | TCP port. Mailpit requires `1025`. |
| `AUTH_SMTP_TLS_MODE` | `none`, `starttls`, or `tls`. Mailpit requires `none`; provider SMTP requires `starttls` or `tls`. |
| `AUTH_SMTP_USERNAME` / `AUTH_SMTP_PASSWORD` | Must both be set or both be empty. Production requires both. They must never be logged or browser-visible. |
| `AUTH_SMTP_TIMEOUT` | SMTP operation timeout from 1–30 seconds; defaults to `10s`. |
| `AUTH_SESSION_TTL` | Existing absolute session lifetime; 15 minutes to 24 hours, default `8h`. |
| `APP_ORIGIN` | The one exact browser origin accepted by login, OTP, and authenticated state-changing operations. Production requires HTTPS. |
| `AUTH_TRUSTED_PROXY_CIDRS` | Only controlled reverse-proxy peers allowed to supply the forwarded client-IP chain used for throttling. |
| `GO_API_ORIGIN` | Server-only Next.js rewrite destination; browser requests remain same-origin under `/api/*`. |

When the local bypass is enabled, correct password verification creates the normal PostgreSQL session and the OTP HMAC/email settings are not loaded. Startup refuses `AUTH_OTP_BYPASS=true` for both `APP_ENV=test` and `APP_ENV=production`; no request payload or frontend state can weaken that guard.

For real local OTP, use `APP_ENV=local`, `AUTH_OTP_BYPASS=false`, a newly generated HMAC secret, and the exact loopback Mailpit values in `.env.example`. Production must inject the HMAC and SMTP credentials through deployment secret handling rather than checking them into an environment file.

## API visibility and threat assumptions

The implementation assumes an attacker knows every route, method, request/response shape, frontend branch, and browser-visible identifier and can replay or modify requests with arbitrary clients. Security therefore does not depend on route hiding or frontend order. Every backend operation independently enforces the applicable Origin, request shape, rate limit, challenge format and ownership lookup, lifecycle, attempt/cooldown, and session rules.

The challenge identifier is a locator, not a credential. Changing it does not select an account; without that challenge's current OTP, and without passing every server-side state check, it cannot create a session. Password/account errors remain generic, and unknown accounts execute dummy Argon2id verification to reduce account-enumeration timing differences.

## Session and logout controls

OTP verification never promotes or reuses the challenge as a session secret. It generates independent 32-byte random session and CSRF values. PostgreSQL stores only the SHA-256 session-token hash; the browser receives the raw token only in `sales_agent_session`, an HttpOnly cookie. The response body never contains it.

Dashboard access resolves the cookie against PostgreSQL and rejects expired, revoked, or inactive-account sessions. Logout requires the session, exact Origin, and matching CSRF value; it revokes PostgreSQL state before clearing the cookie. These controls preserve the Milestone 01 session architecture while moving session creation behind successful OTP verification.

## Remaining production work and risks

- **Deployment-level emergency recovery is not implemented.** Architecture §13.9 requires an infrastructure-authorized `platform-admin` recovery command with an explicit reason, one-time short-lived recovery authorization, recovery-material rotation, and an immutable recovery audit/security event. No such command or audit event exists in this milestone. Therefore notification-email/provider failure can still lock out the only Super Admin, and the implementation is not ready to claim full production lockout resilience. There is intentionally no normal UI bypass and the local bypass cannot be enabled in production.
- **SecretReference integration remains deployment work.** `AUTH_OTP_HMAC_SECRET` and SMTP credentials are validated environment inputs today. Production must move them behind the approved SecretReference/managed-secret boundary with least-privilege access, rotation, auditability, and redaction. Rotating the HMAC key invalidates outstanding OTP challenges; operators must plan for affected users to restart login.
- A real production SMTP provider, sending identity/domain policy, delivery monitoring, alerting, and provider-specific operational runbooks still need deployment validation. The generic TLS SMTP adapter is the application boundary; Mailpit must never be used as production evidence or infrastructure.
- Authentication/recovery security-event persistence and operational monitoring are not yet present. Recovery and provider failure evidence must exclude OTPs, passwords, raw session/CSRF tokens, HMAC keys, SMTP credentials, and raw provider errors.
- The existing Milestone 01 session model enforces configured absolute expiry but does not yet enforce an idle-expiry policy. This is separate from OTP correctness and remains production hardening under architecture §13.5.

Until the emergency recovery and managed-secret controls are implemented and exercised, this milestone is suitable for real local OTP/manual browser validation and further deployment integration, but not a final production authentication sign-off.

## Verification commands

Run the normal repository checks from the root:

```bash
./scripts/check-toolchain.sh
make backend-vet
make backend-test
pnpm install --frozen-lockfile
make frontend-typecheck
make frontend-test
make frontend-build
pnpm audit --prod --audit-level=high
git diff --check
```

Real PostgreSQL/Redis integration tests and the migration up/down/up check remain opt-in through their documented test environment variables. The Mailpit and Playwright paths require the local infrastructure and a bootstrapped Super Admin; they must exercise the real Go backend rather than a mocked authentication API.

The complete OTP-mode Playwright suite also ages one real challenge through a
non-HTTP, test-only fixture. Run the API against a dedicated loopback database
whose name ends exactly in `_integration_test`, apply all migrations there, and
bootstrap the E2E account in that same database. `E2E_DATABASE_URL` must be
identical to the API's `DATABASE_URL`; the fixture refuses the normal local
`sales_agent` database and every non-loopback target. With that API, the web
application, Redis, and Mailpit running, execute:

```bash
PLAYWRIGHT_BASE_URL=http://127.0.0.1:3001 \
E2E_AUTH_MODE=otp \
E2E_ADMIN_EMAIL=e2e-admin@example.test \
E2E_ADMIN_PASSWORD='<the disposable E2E account password>' \
E2E_MAILPIT_ORIGIN=http://127.0.0.1:8025 \
E2E_DATABASE_URL='postgres://sales_agent:sales_agent_local@127.0.0.1:5432/sales_agent_m02_integration_test?sslmode=disable' \
pnpm --filter @sales-agent/admin-web test:e2e
```

When credentials select OTP mode, a missing `E2E_DATABASE_URL` fails the
required expiration case instead of reporting a green suite with that case
skipped. Bypass mode is a separate intentional run with
`E2E_AUTH_MODE=bypass`; OTP-only cases are skipped in that mode.
