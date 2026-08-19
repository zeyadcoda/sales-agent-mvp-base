# Environment Preparation Runbook

Status: IN PROGRESS
Date: 2026-08-18

## Required toolchain

- Go 1.26.6
- Node.js 24.19.0 LTS
- pnpm 10.34.5
- Docker Engine + Docker Compose v2
- PostgreSQL 18.6 local container
- Redis 8.10.0 local container (only for features that require it)

## Ordered preparation

1. Validate toolchain with `./scripts/check-toolchain.sh`.
2. Start local infrastructure with `make infra-up`.
3. Confirm PostgreSQL health.
4. Install backend dependencies and add pgx/sqlc/migrations.
5. Add Google ADK Go v2.2.0 only after Go 1.26.6 is active.
6. Install frontend dependencies with `pnpm install --frozen-lockfile` after the first lockfile is generated.
7. Create local `.env` from `.env.example`; never add live production provider secrets.
8. Run migrations from an empty database.
9. Run backend tests.
10. Run frontend typecheck/build/tests.
11. Start API and frontend.
12. Verify TEST cannot resolve production-capable adapters.
13. Provision the first local Super Admin only after auth foundation is implemented.

## Current sandbox status

Available:
- Debian 13
- Git
- Go 1.23.2 (temporary runner only)
- Node 22.16.0 (temporary runner only)

Unavailable / blocking readiness:
- Approved Go 1.26.6 toolchain
- Node 24.19.0
- pnpm
- Docker/Compose
- PostgreSQL

The backend bootstrap intentionally has no external dependencies yet, so its environment-safety and health tests can run with `GOTOOLCHAIN=local`. This is not permission to begin feature coding on the older Go version.

## Exit gate

Environment preparation is complete only when:

- exact toolchain check passes;
- local infrastructure is healthy;
- backend connects to PostgreSQL;
- migrations run from empty database;
- backend tests pass;
- frontend installs/builds/tests;
- API liveness and readiness are healthy;
- frontend can call backend locally;
- no secrets are committed;
- TEST-side-effect isolation test passes;
- first Super Admin bootstrap mechanism is ready at the authentication milestone.
