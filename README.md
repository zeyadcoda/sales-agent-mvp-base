# Sales Agent MVP Base

Technical foundation repository for the AI Sales Agent MVP.

## Status

Environment preparation is in progress. Feature implementation has not started.

## Approved toolchain

- Go 1.26.6
- Google ADK Go v2.2.0 (to be added after exact Go toolchain bootstrap)
- Node.js 24.19.0 LTS
- pnpm 10.34.5
- Next.js 16.3.0
- React 19.2.x
- PostgreSQL 18.6
- Redis 8.10.x only where justified

## Security invariants

1. Application services are authoritative; AI is untrusted.
2. Every protected resource lookup is object/scope authorized.
3. No Agent gets raw credentials, arbitrary SQL, arbitrary URLs, shell access, or direct production provider access.
4. TEST execution cannot resolve production-capable side-effect adapters.
5. External documents/pages/messages are untrusted data and cannot override platform instructions.
6. UI hiding is never authorization.

## Local readiness

Run:

```bash
./scripts/check-toolchain.sh
```

This repository intentionally fails readiness until the approved toolchain and database infrastructure are available.
