# ADR-0001 — Go Modular Monolith Technical Baseline

Status: APPROVED WORKING BASELINE
Date: 2026-08-18

## Decision

Use a domain-oriented modular monolith with:

- Go backend and worker processes.
- Google ADK Go v2 for the Agent component only.
- Next.js + React + strict TypeScript for Super Admin UI.
- PostgreSQL as authoritative data store.
- pgx + sqlc for database access after dependency bootstrap.
- REST/OpenAPI and SSE where streaming is required.
- Separate PRODUCTION and TEST execution planes/adapters.
- Application-owned authorization, state, side effects, logs, audit, and provider/model registry.

## Security invariants

- AI is untrusted and never authorizes itself.
- No Agent receives raw credentials or generic database/provider access.
- Every protected object lookup is scope-authorized server-side.
- TEST never resolves live side-effect adapters.
- Sensitive external input is treated as untrusted data, not instructions.
