# Plan 003: Add historical reasoning and ship the first premium diagnosis pack

> **Executor instructions**: Follow this plan step by step. Run every
> verification command and confirm the expected result before moving to the
> next step. If anything in the "STOP conditions" section occurs, stop and
> report — do not improvise. When done, update the status row for this plan
> in `docs/plans/README.md` — unless a reviewer dispatched you and told you they
> maintain the index.
>
> **Drift check (run first)**: `git diff --stat bf5cb8f..HEAD -- cloud/internal/storage/repository cloud/internal/analysis/rules cloud/internal/api shared/metrics migrations`
> If any in-scope file changed since this plan was written, compare the
> "Current state" excerpts against the live code before proceeding; on a
> mismatch, treat it as a STOP condition.

## Status

- **Priority**: P1
- **Effort**: L
- **Risk**: HIGH
- **Depends on**: docs/plans/001-backend-diagnosis-execution.md, docs/plans/002-diagnosis-issue-model.md
- **Category**: direction
- **Planned at**: commit `bf5cb8f`, 2026-06-14

## Why this matters

Historical reasoning is what separates a premium diagnosis platform from a
rule-based dashboard with better copy. Premium value comes from saying "this is
abnormal for this database," "it started getting worse here," and "these
signals moved together." This plan adds those reasoning primitives and uses
them to make the first three diagnosis categories feel meaningfully better.

## Current state

- `metric_points`, `table_stats`, and `query_stats` already store evidence.
- `cloud/internal/api/metrics_query.go` exposes time-series retrieval, but only as
  a generic chart-oriented query surface.
- existing findings only record `current_value` and `threshold_value`; there is
  no baseline or change summary yet.
- active detectors exist already for slow queries, high dead rows,
  autovacuum lag, and sequential scan pressure.

Current evidence excerpt:

- `migrations/000003_table_query_stats.sql:1-44`
  ```sql
  CREATE TABLE IF NOT EXISTS table_stats ( ... );
  CREATE TABLE IF NOT EXISTS query_stats ( ... );
  ```

- `cloud/internal/api/metrics_query.go:37-41`
  ```go
  // handleQueryMetrics serves GET /api/metrics/query, returning historical
  // metric evidence aggregated into buckets via TimescaleDB's time_bucket.
  ```

Existing detector coverage:

- `agent/internal/runner.go:102-120`
  ```go
  engine.RegisterTableRules(
      rules.HighDeadRowsRule{...},
      rules.AutovacuumLagRule{...},
      rules.HighSequentialScanRule{...},
  )
  engine.RegisterQueryRules(
      rules.SlowQueryRule{...},
      rules.ExpensiveQueryRule{...},
      rules.DiskHeavyQueryRule{...},
  )
  ```

## Commands you will need

| Purpose | Command | Expected on success |
|---------|---------|---------------------|
| Backend tests | `go test ./...` | exit 0 |
| Frontend build | `cd frontend && npm run build` | exit 0 |
| Frontend lint | `cd frontend && npm run lint` | exit 0 |

## Scope

**In scope**:

- `cloud/internal/storage/repository/*` helpers for historical comparisons
- `cloud/internal/api/*` only where needed to expose historical evidence
- `cloud/internal/analysis/rules/*` for the first premium diagnosis pack
- `shared/metrics/*` only where shared historical abstractions are needed
- new migrations/indexes if required for query performance

**Out of scope**:

- all seven diagnosis categories
- alert delivery
- deep frontend redesign beyond consuming the new historical evidence fields

## Git workflow

- Branch: `advisor/003-historical-reasoning-and-first-pack`
- Commit messages should match the repo's short descriptive style
- Do NOT push or open a PR unless instructed

## Steps

### Step 1: Add reusable historical comparison helpers

Create repository helpers for the backend to ask:

- current value for a resource
- previous period value
- rolling baseline value
- percentage change
- first time abnormality appeared inside a lookback window

Implement these for the evidence tables already present:

- `metric_points`
- `table_stats`
- `query_stats`

Keep the API internal to the backend diagnosis service; do not expose these
raw helpers directly as public endpoints yet.

**Verify**: `go test ./...` → exit 0

### Step 2: Add correlation helpers for the first diagnosis pack

For the first premium pack, add repository-level functions that can correlate:

- dead rows growth + last vacuum/autovacuum gap + sequential scan increase
- query latency growth + call volume + disk read pressure
- cache hit deterioration + disk reads + slow query count

Do not build a generic correlation engine yet. Build narrowly useful helpers
that produce trustworthy evidence for the chosen categories.

**Verify**: `go test ./...` → exit 0

### Step 3: Upgrade three diagnosis categories end to end

Improve these categories first:

1. table bloat
2. vacuum problems
3. slow queries

For each category, update detector output so it populates the richer issue
model with:

- problem summary
- evidence summary
- impact summary
- baseline value
- change summary
- suggested action
- verification hint
- confidence score/label

Example expected outcome for table bloat:

- not just `dead tuples high`
- but `table bloat risk detected`
- with evidence derived from dead row growth, last vacuum gap, and scan trend

**Verify**: `go test ./...` → exit 0

### Step 4: Expose supporting historical evidence for issue detail

Extend the issue detail response from plan 002 with a bounded historical
evidence section, such as:

- `historical_context`
  - `baseline_value`
  - `previous_value`
  - `change_percent`
  - `trend_window`
- `evidence_points`
  - short chart series only for directly relevant metrics

Keep this focused. The endpoint should return only the evidence needed to
support the diagnosis, not every metric series for the resource.

**Verify**: `go test ./...` → exit 0

## Test plan

- Add repository tests for baseline and comparison helpers
- Add rule tests for the three upgraded categories covering:
  - normal behavior does not trigger
  - clear degradation triggers
  - evidence summary fields are populated
  - confidence behaves predictably for weak vs strong evidence
- Add issue detail handler tests ensuring historical evidence is bounded and
  category-relevant
- Verification: `go test ./...` → all pass

## Done criteria

- [ ] Historical comparison helpers exist for current vs previous vs baseline
- [ ] Three categories emit diagnosis-grade evidence and impact fields
- [ ] Issue detail responses include supporting historical context
- [ ] `go test ./...` exits 0
- [ ] `cd frontend && npm run build` exits 0
- [ ] No files outside the in-scope list are modified (`git status`)
- [ ] `docs/plans/README.md` status row updated

## STOP conditions

- Plans 001 or 002 have not landed
- Historical evidence queries require a broader schema redesign than additive
  helper work
- Existing detectors cannot be upgraded without first redefining their input
  data models
- The historical evidence payload starts turning into a generic chart endpoint
  instead of diagnosis support

## Maintenance notes

- Reviewers should scrutinize whether each evidence summary is actually backed
- by the data loaded in this plan, not by hand-wavy prose.
- After these three categories land, use the same pattern for lock contention,
  cache problems, missing indexes, and connection issues.
