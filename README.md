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

### Agent

The agent is a lightweight PostgreSQL signal collector. It gathers raw evidence such as:

- `pg_stat_database`
- `pg_stat_activity`
- `pg_stat_user_tables`
- `pg_stat_statements`
- lock information
- vacuum history and lag signals
- index usage
- query performance

The agent should stay lightweight.

It should not be responsible for diagnosis logic.

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

- `agent_id`
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

The repository already contains the beginnings of this architecture:

- collectors for database, activity, table, and query signals
- a backend analysis engine that turns evidence into findings
- TimescaleDB-backed metric storage
- a frontend that is being repositioned around health, issues, queries, and tables

The next steps should continue pushing the codebase away from dashboard thinking and toward diagnosis-first product behavior.
