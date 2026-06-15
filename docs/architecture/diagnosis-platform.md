# Postgresome Diagnosis Platform Direction

This document is the working product and architecture brief for Postgresome.

## One-line definition

Postgresome is a PostgreSQL diagnosis platform that analyzes database behavior, explains problems, shows evidence, and recommends what to fix.

## Why this direction matters

There are already many tools that can show PostgreSQL charts.

That is not the hard part for users.

The hard part is:

- knowing which signals matter
- knowing whether a change is abnormal
- connecting multiple signals into one explanation
- deciding what to fix first

Postgresome should solve that problem.

## The mental model

Users should experience Postgresome like a database doctor:

1. Observe symptoms
2. Compare against history
3. Identify likely cause
4. Show evidence
5. Estimate impact
6. Recommend next action

## Diagnosis template

Each important finding should trend toward this structure:

```text
Problem:
What issue is happening?

Evidence:
What data proves it?
What changed?
What related signals support the explanation?

Impact:
Why should the user care?

Suggested action:
What should the user do next?
```

## Architecture split

### Source setup layer

The connection layer should stay simple.

Its responsibilities:

- connect to PostgreSQL-compatible sources
- collect raw evidence
- normalize it into ingestible payloads
- send the data to the backend

Its non-responsibilities:

- making product-level diagnoses
- computing user-facing recommendations
- carrying heavy correlation logic

### Backend

The backend is the diagnosis engine.

Its responsibilities:

- store raw evidence in TimescaleDB
- retain enough context to reason historically
- detect issues from single signals and correlated signals
- compare current values to historical baselines
- generate findings and recommendations
- expose diagnosis-first API responses

### TimescaleDB

TimescaleDB is the historical memory layer.

Use it to answer:

- what changed recently?
- when did a degradation start?
- is this behavior unusual for this database?
- which other signals moved at the same time?

## What the frontend should optimize for

The frontend should feel like an investigation tool, not a wall of telemetry.

Default information hierarchy:

1. Overall database health
2. Current problems
3. Root cause explanation
4. Evidence
5. Suggested fixes
6. Supporting history and charts

Historical visualization is useful only when it strengthens diagnosis.

## MVP categories

The first categories that matter most:

1. Connection issues
2. Slow queries
3. Table bloat
4. Missing indexes
5. Cache problems
6. Vacuum problems
7. Lock contention

## Collection filter

Every metric should justify itself through diagnosis value.

Ask:

```text
What diagnosis can this metric help detect, explain, or prove?
```

Good reasons to collect a metric:

- it helps detect a problem
- it helps establish historical baseline
- it helps correlate or confirm a cause
- it strengthens evidence for a recommendation

Bad reasons to collect a metric:

- it might look nice on a chart
- it is commonly present in monitoring tools
- it fills space in a dashboard

## Design implications

UI should avoid:

- chart-first landing pages
- random KPIs without explanation
- broad “monitoring console” patterns
- generic infrastructure-dashboard hierarchy

UI should emphasize:

- diagnosis queue
- issue detail
- problem/evidence/fix structure
- historical context only when it helps explain change

## Notes for implementation

- Route and type names in the current codebase may still use `dashboard` language. That is legacy naming, not product intent.
- We can preserve stable internal APIs while changing the product behavior toward diagnosis-first outputs.
- Schema evolution should favor storing reusable evidence rather than product-specific widget data.
