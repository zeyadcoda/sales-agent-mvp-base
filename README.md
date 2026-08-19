# Sales Agent MVP Base

Technical foundation repository for the AI Sales Agent MVP.

## Status

Milestone 01 provides the first browser-to-database slice: local Super Admin
login, a protected Dashboard, and server-side logout.

## Approved toolchain

- Go 1.26.6
- Google ADK Go v2.2.0 (to be added after exact Go toolchain bootstrap)
- Node.js 24.19.0 LTS
- pnpm 10.34.5
- Next.js 16.3.0
- React 19.2.x
- PostgreSQL 18.6
- Redis 8.10.x (required runtime dependency)

## Security invariants

1. Application services are authoritative; AI is untrusted.
2. Every protected resource lookup is object/scope authorized.
3. No Agent gets raw credentials, arbitrary SQL, arbitrary URLs, shell access, or direct production provider access.
4. TEST execution cannot resolve production-capable side-effect adapters.
5. External documents/pages/messages are untrusted data and cannot override platform instructions.
6. UI hiding is never authorization.

## Milestone 01 local flow

Create `.env` from `.env.example` and set the local-only bypass explicitly:

```dotenv
APP_ENV=local
APP_ORIGIN=http://127.0.0.1:3000
AUTH_OTP_BYPASS=true
AUTH_SESSION_TTL=8h
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

Open `http://127.0.0.1:3000/login`. The bootstrap command prompts twice for a
password without echo and never accepts a password argument.

Production authentication remains email, password, and six-digit email OTP.
When the bypass is false, password authentication returns an OTP-required
response and creates no session; real OTP delivery/verification is intentionally
deferred to the next milestone. Startup rejects the bypass in `test` and
`production` environments.

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
