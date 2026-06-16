# Postgresome

Postgresome is a PostgreSQL diagnosis platform.

The product goal is not to expose every database metric or become a generic monitoring dashboard. The goal is to reduce time to fix by answering:

- What is wrong?
- Why is it happening?
- What evidence proves it?
- What changed over time?
- What should I fix next?

## Product philosophy

PostgreSQL exposes hundreds of useful signals, but users should not need to translate all of them by hand.

Postgresome treats raw metrics as evidence, not as the product itself.

Instead of showing only:

```text
Dead tuples: 2.5M
```

Postgresome should diagnose:

```text
Problem:
Table bloat risk detected

Evidence:
- users table dead tuples increased from 50k to 2.5M
- last vacuum happened 5 days ago
- sequential scans increased 40%

Impact:
Queries touching this table may become slower.

Suggested action:
Review autovacuum configuration or schedule maintenance.
```

## Architecture direction

### Source connectors

Postgresome should be source-first.

The near-term product should work with PostgreSQL evidence coming from:

- direct database connections
- managed provider integrations such as Supabase, Neon, or RDS

All evidence should feed the same backend diagnosis model.

We are intentionally removing the standalone agent runtime for now so the
product can focus on a simpler connection and diagnosis flow.

### API / backend

The backend is the diagnosis engine. It is responsible for:

- storing raw time-series evidence in TimescaleDB
- comparing current behavior with historical behavior
- identifying degradation and anomalies
- correlating related signals
- producing findings, health summaries, and recommendations

### TimescaleDB

TimescaleDB is not just chart storage.

It is Postgresome's historical memory layer. It should support:

- understanding normal database behavior
- comparing current vs historical patterns
- detecting anomalies
- identifying degradation over time
- correlating related events across signals

The storage model should support:

- `source_id`
- `source_kind`
- `database_id`
- `metric name`
- dimensions such as table, query, user, lock target, or database
- timestamp
- value

## Frontend priority

The frontend should optimize for diagnosis flow:

1. Database health overview
2. Current problems
3. Root cause explanation
4. Evidence
5. Suggested fixes
6. Historical charts only when needed

Avoid building:

- huge configurable dashboards first
- random charts everywhere
- infrastructure-monitoring clones

Build toward:

```text
Your database has a problem.
Here is why.
Here is proof.
Here is what changed.
Here is how to fix it.
```

## MVP diagnosis categories

The first MVP should focus on:

1. Connection issues
2. Slow queries
3. Table bloat
4. Missing indexes
5. Cache problems
6. Vacuum problems
7. Lock contention

Every collected metric should answer one question:

```text
What diagnosis can this help us make?
```

If a metric cannot help detect, explain, or prove a problem, we should avoid collecting it for now.

## Current repository direction

The repository is being simplified around two main product surfaces:

- a backend app that stores evidence and runs diagnosis
- a diagnosis-first frontend

The next steps should continue pushing the codebase away from dashboard thinking and toward diagnosis-first product behavior.

For the target long-term flow, see:

- `docs/architecture/diagnosis-platform.md`
- `docs/architecture/source-to-diagnosis-flow.md`

## Repository layout

- `backend/` — Go backend app, module, and migration entrypoints
- `backend/internal/` — backend-owned API, diagnosis, analysis, storage, and supporting packages
- `backend/internal/metrics/` — backend-owned evidence and metric types
- `backend/migrations/` — schema changes for the Postgresome storage database
- `frontend/` — diagnosis UI
- `docker/` — remaining local infrastructure assets
- `docs/` — architecture docs, roadmap, implementation plans, and mockups
