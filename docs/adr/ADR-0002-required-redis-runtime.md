# ADR-0002 — Required Redis Runtime

Status: APPROVED WORKING BASELINE
Date: 2026-08-19

## Context

The initial architecture treated Redis as optional. The implemented MVP foundation now uses Redis as a required application-runtime dependency and includes it in API readiness.

This decision supersedes only the earlier “Redis not required initially” position. It does not change approved product behavior or the [ADR-0001](ADR-0001-go-modular-monolith.md) decision that PostgreSQL and deterministic application services remain authoritative.

## Decision

- Redis is required for the API runtime in local, CI, staging, and production environments.
- PostgreSQL remains the authoritative source of durable business and security state.
- Redis is limited to short-lived, reconstructible concerns such as rate limiting, coordination, justified caching, and other explicitly designed ephemeral functions.
- `GET /health/ready` returns ready only when bounded PostgreSQL and Redis checks both succeed.
- `GET /health/live` remains independent of PostgreSQL and Redis.
- Redis configuration is loaded only from validated server-side configuration and fails closed when missing or invalid.

## Security implications

- Redis failure must never trigger an unsafe or permissive fallback.
- Redis failure must never bypass authentication, authorization, tenant isolation, rate limits, environment isolation, or side-effect safeguards.
- Redis loss or eviction must never cause loss of authoritative business records; Redis-held data must be safe to discard and reconstruct.
- Redis failure must never produce a false `READY` or `HEALTHY` signal.
- Redis credentials, connection URLs, and raw dependency errors must not be exposed to browsers, Agents, logs, or health responses.

## Consequences

- Every API environment must provision, secure, monitor, and test Redis alongside PostgreSQL.
- PostgreSQL or Redis unavailability makes the API not ready while process liveness remains available for orchestration and diagnosis.
- Redis-backed functions must define bounded timeouts, failure behavior, and reconstruction semantics.
- Dependency tests must cover healthy, failed, timed-out, and missing PostgreSQL and Redis states.
- Redis adds runtime infrastructure and operational cost without changing the durable data model.

## What remains prohibited

- Using Redis as an authoritative or sole durable datastore.
- Storing the sole copy of Organization, Package, Agent, Agent Version, Audit, authentication identity, authorization grant, outbox, or other durable business records in Redis.
- Allowing Redis to replace PostgreSQL transactions, constraints, tenant-scoped repositories, immutable audit, or application-owned security decisions.
- Silent fallback to process-local memory, an implicit Redis endpoint, permissive behavior, or a production-capable adapter.
- Giving browsers, Agents, prompts, or generic application packages unrestricted Redis access.
