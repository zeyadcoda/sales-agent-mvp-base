# Sales Agent MVP Base

Technical foundation repository for the AI Sales Agent MVP.

## Status

Milestone 02 adds the production-shaped Super Admin flow: password verification,
a six-digit email OTP, PostgreSQL-backed challenge verification, the existing
server-side session, and logout. The local-only bypass remains available.

## Approved toolchain

- Go 1.26.6
- Google ADK Go v2.2.0 (to be added after exact Go toolchain bootstrap)
- Node.js 24.19.0 LTS
- pnpm 10.34.5
- Next.js 16.3.0
- React 19.2.x
- PostgreSQL 18.6
- Redis 8.10.x (required runtime dependency)
- Mailpit 1.30.7 (local/test email capture only)

## Security invariants

1. Application services are authoritative; AI is untrusted.
2. Every protected resource lookup is object/scope authorized.
3. No Agent gets raw credentials, arbitrary SQL, arbitrary URLs, shell access, or direct production provider access.
4. TEST execution cannot resolve production-capable side-effect adapters.
5. External documents/pages/messages are untrusted data and cannot override platform instructions.
6. UI hiding is never authorization.

## Real local OTP flow

Create `.env` from `.env.example`. Replace the OTP HMAC placeholder with a
locally generated secret; do not copy the output into Git, logs, or chat:

```bash
openssl rand -base64 32
```

Use these local email settings:

```dotenv
APP_ENV=local
APP_ORIGIN=http://127.0.0.1:3001
AUTH_OTP_BYPASS=false
AUTH_SESSION_TTL=8h
AUTH_OTP_HMAC_SECRET=<BASE64_OUTPUT>
AUTH_EMAIL_MODE=mailpit
AUTH_EMAIL_FROM_ADDRESS=no-reply@sales-agent.local
AUTH_EMAIL_FROM_NAME="Sales Agent"
AUTH_SMTP_HOST=127.0.0.1
AUTH_SMTP_PORT=1025
AUTH_SMTP_TLS_MODE=none
AUTH_SMTP_USERNAME=
AUTH_SMTP_PASSWORD=
AUTH_SMTP_TIMEOUT=10s
AUTH_TRUSTED_PROXY_CIDRS=127.0.0.1/32,::1/128
GO_API_ORIGIN=http://127.0.0.1:8081
```

Then run:

```bash
make infra-up
make migrate-up
make bootstrap-super-admin EMAIL=admin@example.com NAME="Super Admin"
make api
```

In a second terminal:

```bash
make web
```

Open `http://127.0.0.1:3001/login`. The bootstrap command prompts twice for a
password without echo and never accepts a password argument. After password
verification, retrieve the six-digit code from the Mailpit inbox at
`http://127.0.0.1:8025` and complete verification in the browser.

Mailpit binds SMTP and its web UI to loopback only. It is a local/test capture
tool, not a production dependency. `make infra-up` starts PostgreSQL, Redis, and
Mailpit; `make mailpit-up` starts only Mailpit, and `make mailpit-logs` follows
its container logs. With Mailpit running, `make mailpit-test` verifies SMTP
delivery and inbox content through the real local adapter.

For the local developer bypass, set only `APP_ENV=local` and
`AUTH_OTP_BYPASS=true`. OTP HMAC/email configuration is unused in that mode.
Startup rejects the bypass in `test` and `production` environments.

For the full OTP Playwright acceptance suite, run the API against a dedicated
loopback PostgreSQL database ending in `_integration_test`, and set
`E2E_DATABASE_URL` to that exact same URL. The expiration test deliberately
refuses the normal `sales_agent` development database and remote targets; a
missing fixture URL fails OTP-mode acceptance. The exact environment and
command are documented in
`docs/security/MILESTONE_02_AUTHENTICATION.md#verification-commands`.

## Production authentication email configuration

Production requires `AUTH_EMAIL_MODE=smtp`, `AUTH_SMTP_TLS_MODE=starttls|tls`,
and a non-empty username/password pair. Set the provider hostname, port, From
address/name, and a bounded `AUTH_SMTP_TIMEOUT` from deployment configuration.
Never configure Mailpit in production.

`AUTH_OTP_HMAC_SECRET` is base64-encoded key material of at least 32 random
bytes. SMTP credentials and this HMAC key are deployment-injected environment
secrets for Milestone 02; production hardening must resolve them through the
approved `SecretReference`/managed-secret boundary instead of storing them in
repository or platform records. The required deployment-level emergency
recovery command and immutable recovery audit remain separate production work;
there is no normal UI OTP bypass.

The session cookie is `HttpOnly`, `SameSite=Strict`, host-only, and `Secure` in
non-local environments. `Secure=false` is used only for the loopback HTTP
development origin because browsers cannot send Secure cookies over local HTTP.

## Local readiness

Run:

```bash
./scripts/check-toolchain.sh
```

PostgreSQL and Redis must both be healthy for API readiness. Browser API calls
remain same-origin through the Next.js `/api/*` rewrite.
