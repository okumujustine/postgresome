# Source To Diagnosis Flow

This document defines the long-term product architecture for Postgresome.

The key shift is:

```text
Postgresome is a diagnosis product.
```

## Core model

The durable product model should be:

```text
source
  -> raw evidence
  -> backend normalization
  -> diagnosis engine
  -> findings
  -> linked evidence
  -> recommendations
  -> UI
```

Or more concretely:

```text
Postgres / Supabase / RDS / Cloud SQL
  -> evidence collection
  -> TimescaleDB historical storage
  -> correlation and diagnosis
  -> finding detail
  -> user action
```

## Product belief

Users do not buy Postgresome because it can collect PostgreSQL metrics.

Users buy Postgresome because it can:

- identify what is wrong
- connect signals into one explanation
- reduce time to investigation
- tell them what to do next

That means the backend diagnosis model is the product center.

## Sources

Postgresome should support multiple source types over time.

### 1. Direct database connection

Examples:

- local PostgreSQL
- self-hosted production PostgreSQL
- private network PostgreSQL through a bastion or tunnel

This is the simplest source type conceptually:

```text
Postgresome connects to the database and reads PostgreSQL system views.
```

### 2. Managed provider integration

Examples:

- Supabase
- Neon
- RDS
- Cloud SQL

These should be treated as first-class source integrations, not forced
through a single-connector story.

The collection method may differ by provider:

- direct SQL connection
- provider API + SQL
- OAuth + credentials exchange
- temporary tokens

But once evidence reaches the backend, the diagnosis model should stay the
same.

### 3. Deferred private-network connector

Private-network collection can return later as a dedicated connector when we
need it.

For now, we are simplifying the product around direct connections and managed
integrations so the backend diagnosis model stays clear and the setup story
stays lightweight.

## Current implementation reality

Today the repo is being simplified toward:

```text
source setup
  -> evidence ingestion
  -> TimescaleDB
  -> diagnosis service
  -> findings + evidence
  -> frontend
```

## Backend responsibilities

The backend should own all of the following:

- evidence storage
- baselines and historical comparison
- correlation across signal types
- diagnosis generation
- confidence scoring
- likely-cause ranking
- suggested actions
- fix verification

The backend should answer:

- what changed?
- when did it start?
- what object is affected?
- what evidence supports this?
- what is likely the cause?
- what should the user do next?

## Canonical domain flow

The main domain flow should be:

```text
database source
  -> evidence snapshots
  -> diagnosis run
  -> finding
  -> finding_evidence records
  -> overview / finding detail APIs
```

### Important distinction

Queries, tables, locks, and raw metrics are **not** separate product centers.

They are evidence shapes.

That means:

- a finding may be supported by query evidence
- a finding may be supported by table evidence
- a finding may be supported by maintenance evidence
- a finding may be supported by lock/activity evidence
- one evidence item may contribute to multiple findings later

## Data model direction

The current diagnosis-first model should continue to evolve around these
concepts:

### Source

Represents where evidence came from.

Suggested future fields:

- `source_id`
- `source_kind`
  - `direct`
  - `managed_integration`
- `provider`
  - `postgres`
  - `supabase`
  - `rds`
  - `cloudsql`
  - `neon`
- `connection_scope`
  - project
  - database
  - cluster

### Evidence

Represents raw or derived proof signals.

Examples:

- query latency
- dead tuples
- autovacuum recency
- blocked sessions
- rollback rate

### Finding

Represents the diagnosis.

A finding should answer:

- what issue exists?
- how severe is it?
- what is affected?
- what is the likely impact?
- what should happen next?

### Finding evidence link

Represents the explicit relationship between a diagnosis and its supporting
proof.

This is the core model that keeps the frontend from becoming a dashboard.

## Example: slow query regression

Here is the intended flow for a slow query finding.

### 1. Source reads evidence

A source connection reads:

- `pg_stat_statements`
- optional explain plan
- optional related activity/lock state

### 2. Backend stores raw evidence

The backend stores:

- query snapshot
- execution timing
- block reads
- call count
- explain plan

### 3. Diagnosis engine evaluates

The engine compares:

- current mean execution time
- historical baseline
- current call volume
- disk-read behavior

### 4. Backend creates finding

Example:

```text
Finding:
Slow query regression detected
```

### 5. Backend links evidence

Example evidence items:

- mean execution time rose from 22 ms to 180 ms
- shared block reads increased 240%
- call volume remained stable

### 6. UI shows diagnosis

The UI should render:

```text
Problem
  Slow query regression detected

Evidence
  Mean execution time increased sharply
  Disk reads increased
  This started around 2:10 PM

Action
  Inspect plan and index usage
```

## Example: vacuum lag

The same structure applies to non-query problems.

### Evidence may be:

- dead tuples
- dead-row ratio
- last autovacuum age
- sequential scan increase

This proves the value of the model:

the product is not query-first or table-first.

It is finding-first.

## API direction

The API should gradually become more diagnosis-native.

Preferred top-level surfaces:

- `GET /api/overview`
- `GET /api/findings`
- `GET /api/findings/:id`

Supporting surfaces:

- raw metrics
- query history
- table history
- explain plans

Those supporting surfaces should exist to deepen a finding, not replace it.

## UI direction

The UI should always follow:

```text
finding
  -> explanation
  -> evidence
  -> action
```

Not:

```text
metrics
  -> charts
  -> user figures it out
```

## Implementation ladder

Near-term implementation order should be:

1. Strengthen the findings and evidence model
2. Make findings the center of the UI
3. Add source abstraction in the backend
4. Add provider-specific source implementations
5. Add managed integrations such as Supabase
6. Add deeper correlation and premium diagnosis workflows

## Architectural rule

Before adding any new collector, endpoint, table, or UI area, ask:

```text
Does this improve diagnosis quality, evidence quality, or time to action?
```

If not, it is probably dashboard drift.
