# Postgresome Roadmap

This roadmap reflects Postgresome's diagnosis-first direction.

Postgresome is not trying to become another Grafana, Datadog, or generic PostgreSQL metrics dashboard.

The product we are building is a database doctor.

## Product goal

Reduce time to fix PostgreSQL problems by turning raw signals into:

- diagnosis
- evidence
- impact
- historical context
- suggested action

## Product principles

1. Metrics are evidence, not the product.
2. Historical data exists to explain change over time, not to fill screens with charts.
3. Sources collect signals. The backend interprets them.
4. Every collected metric must justify itself through a diagnosis use case.
5. The default UI experience starts with problems, not dashboards.

## What we are not building first

Do not prioritize:

- configurable dashboard builders
- chart-heavy monitoring surfaces
- infra-monitoring clones
- broad observability features without diagnosis value

## Diagnosis-first architecture

### Source layer

Postgresome should support multiple ingestion paths:

- direct database connection
- managed provider integrations

The current reset deliberately removes the standalone agent runtime so we can
ship a simpler diagnosis platform first.

### Backend responsibilities

The backend should:

- persist raw evidence in TimescaleDB
- compare current metrics with historical baselines
- detect anomalies and degradation
- correlate related events
- produce findings and recommendations
- compute health summaries from diagnoses

### TimescaleDB responsibilities

TimescaleDB is the historical memory layer for diagnosis.

Its job is to preserve enough evidence to answer:

- what changed?
- when did it start?
- is this abnormal for this database?
- what other signals moved at the same time?

## MVP diagnosis categories

The first MVP should focus on these categories:

1. Connection issues
2. Slow queries
3. Table bloat
4. Missing indexes
5. Cache problems
6. Vacuum problems
7. Lock contention

## Phase 1: Source and evidence foundation

Goal:

Establish the minimal signal pipeline needed for diagnosis.

Includes:

- source registration
- metric ingestion
- database activity collection
- table statistics collection
- query statistics collection
- finding ingestion and persistence

Success looks like:

```text
Postgresome can collect evidence continuously without making users think about the collection system.
```

## Phase 2: Source abstraction and backend-owned normalization

Goal:

Make the backend accept evidence from multiple source types without changing
the diagnosis model.

Focus areas:

- formal source model
- shared ingest contracts
- provider-aware connection setup

Success looks like:

```text
Postgresome can support direct and managed PostgreSQL sources without duplicating diagnosis logic.
```

## Phase 3: Diagnosis engine maturity

Goal:

Turn raw evidence into meaningful findings.

Focus areas:

- connection issues
- lock contention
- slow queries
- cache pressure
- vacuum lag
- table bloat
- missing index candidates

Success looks like:

```text
The backend can explain a problem in terms of cause, evidence, and suggested action.
```

## Phase 4: Historical reasoning in TimescaleDB

Goal:

Use historical storage for diagnosis, not just retention.

Add:

- baseline comparisons
- degradation detection
- anomaly detection
- before/after analysis
- correlation across related signals

Success looks like:

```text
Postgresome can say not only what is wrong now, but what changed and when it started getting worse.
```

## Phase 5: Diagnosis-first API surfaces

Goal:

Expose diagnosis-oriented backend responses.

Prioritize endpoints that support:

- database health overview
- findings queue
- issue detail
- evidence retrieval
- table diagnosis context
- query diagnosis context

Raw metric query endpoints remain useful, but as supporting evidence surfaces.

## Phase 6: Diagnosis-first frontend

Goal:

Present the product as a PostgreSQL investigation workflow.

Frontend order of importance:

1. Health overview
2. Current problems
3. Root cause explanation
4. Evidence
5. Suggested fixes
6. Historical charts when they help explain change

Success looks like:

```text
Users land in Postgresome and immediately understand what needs attention and why.
```

## Phase 7: Explainability and confidence

Goal:

Make recommendations more actionable and trustworthy.

Add:

- confidence scoring
- stronger evidence summaries
- clearer impact statements
- better recommendation templates
- links between related findings
- before/after verification of fixes
- source-aware setup experiences for managed providers

## Phase 8: Commercial readiness

Goal:

Turn the diagnosis-first product into a clearly sellable premium offering.

Add:

- explicit buyer positioning
- pricing and packaging
- premium release definition
- diagnosis-first homepage and product messaging
- activation and conversion milestones

Success looks like:

```text
Postgresome has a coherent commercial story that matches the product we are actually building.
```

## Immediate execution filter

Before adding any new collector, endpoint, schema field, or UI module, ask:

```text
What diagnosis does this help us make?
```

If there is no strong answer, it should not be an MVP priority.
