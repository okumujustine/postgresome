# Plan 001: Move diagnosis execution from the agent into the backend

> **Executor instructions**: Follow this plan step by step. Run every
> verification command and confirm the expected result before moving to the
> next step. If anything in the "STOP conditions" section occurs, stop and
> report — do not improvise. When done, update the status row for this plan
> in `docs/plans/README.md` — unless a reviewer dispatched you and told you they
> maintain the index.
>
> **Drift check (run first)**: `git diff --stat bf5cb8f..HEAD -- agent/internal/runner.go agent/internal/client/client.go cloud/internal/api/handlers.go cloud/internal/api/server.go cloud/internal/analysis/engine.go cloud/internal/storage/repository agent/internal/collector migrations`
> If any in-scope file changed since this plan was written, compare the
> "Current state" excerpts against the live code before proceeding; on a
> mismatch, treat it as a STOP condition.

## Status

- **Priority**: P1
- **Effort**: L
- **Risk**: HIGH
- **Depends on**: none
- **Category**: direction
- **Planned at**: commit `bf5cb8f`, 2026-06-14

## Why this matters

The repo already collects strong PostgreSQL evidence, but the agent still runs
the diagnosis engine locally and sends finished findings to the API. That makes
historical comparison, cross-signal correlation, and premium diagnosis logic
much harder because the backend never sees the full raw context. Moving
diagnosis execution into the backend is the architectural pivot that every
later premium feature depends on.

## Current state

- `agent/internal/runner.go` — the agent currently owns rule execution and
  sends completed findings to the API.
- `agent/internal/client/client.go` — still exposes `SendFindings(...)` to
  `POST /api/findings/ingest`.
- `cloud/internal/api/handlers.go` — the API only persists findings; it does not
  generate them.
- `cloud/internal/analysis/engine.go` — rule engine exists already and can be reused
  in the backend.
- `migrations/000003_table_query_stats.sql` — the backend stores table and
  query evidence, but no pg_stat_activity snapshot or EXPLAIN evidence.

Current execution excerpt:

- `agent/internal/runner.go:254-278`
  ```go
  findings := engine.AnalyzeDatabaseMetrics(*stats, delta)
  if activitySnapshot != nil {
      findings = append(findings, engine.AnalyzeDatabaseActivity(*activitySnapshot)...)
  }
  if tableStatsSnapshot != nil {
      findings = append(findings, engine.AnalyzeTableStats(*tableStatsSnapshot)...)
  }
  if queryStatsSnapshot != nil {
      findings = append(findings, engine.AnalyzeQueryStats(*queryStatsSnapshot)...)
  }
  if explainSnapshot != nil {
      findings = append(findings, engine.AnalyzeExplainPlans(*explainSnapshot)...)
  }
  _, err = apiClient.SendFindings(sendFindingsCtx, r.cfg.AgentID, r.cfg.DatabaseInstanceID, findings)
  ```

- `cloud/internal/api/handlers.go:176-204`
  ```go
  if len(req.Findings) > 0 {
      findings := make([]analysis.Finding, len(req.Findings))
      ...
      if err := repository.UpsertFindings(r.Context(), s.pool, findings); err != nil {
          http.Error(w, "failed to store findings", http.StatusInternalServerError)
          return
      }
  }
  ```

- `migrations/000003_table_query_stats.sql:1-44`
  ```sql
  CREATE TABLE IF NOT EXISTS table_stats ( ... );
  CREATE TABLE IF NOT EXISTS query_stats ( ... );
  ```

Repo conventions and constraints:

- Commit messages in recent history are short descriptive phrases, not
  conventional commits. Examples from `git log --oneline -8`:
  `direction changes`, `Dashboard changes`, `adds metrics setup`.
- The product direction doc says the agent should stay lightweight and the
  backend should own diagnosis logic. Match the vocabulary in
  `docs/architecture/diagnosis-platform.md`.
- Current backend handler style is simple net/http plus repository helpers; see
  `cloud/internal/api/findings.go` and `cloud/internal/api/dashboard_overview.go`.

## Commands you will need

| Purpose | Command | Expected on success |
|---------|---------|---------------------|
| Backend tests | `go test ./...` | exit 0 |
| Frontend build | `cd frontend && npm run build` | exit 0 |
| Frontend lint | `cd frontend && npm run lint` | exit 0 |

## Scope

**In scope**:

- `agent/internal/runner.go`
- `agent/internal/client/client.go`
- `cloud/internal/api/handlers.go`
- `cloud/internal/api/server.go`
- new backend diagnosis service files under `cloud/internal/api` or `cloud/internal/diagnosis`
- `cloud/internal/analysis/engine.go` usage sites
- `cloud/internal/storage/repository/*` files needed for new ingest/query paths
- new migrations for additional evidence tables
- collector payload types in `shared/metrics/*` only where required for API transport

**Out of scope**:

- frontend page rewrites beyond the minimum needed to keep existing flows working
- changing detector semantics yet; this plan is about who executes them
- health score design
- notification delivery

## Git workflow

- Branch: `advisor/001-backend-diagnosis-execution`
- Commit per logical unit, using the repo's short descriptive message style
- Do NOT push or open a PR unless the operator instructed it

## Steps

### Step 1: Design the backend-owned diagnosis cycle and freeze the agent contract

Create a short internal design note in the code comments or package docs for a
backend diagnosis cycle with these stages:

1. ingest evidence
2. persist evidence
3. load the latest diagnosis inputs for one database instance
4. run the analysis engine in the backend
5. upsert findings
6. resolve stale findings

Define the new agent contract:

- the agent sends raw evidence only
- the API is responsible for generating findings
- `POST /api/findings/ingest` becomes deprecated or transitional

**Verify**: `go test ./...` → exit 0

### Step 2: Add missing evidence ingestion for activity and EXPLAIN data

The backend cannot own diagnosis until it receives all rule inputs. Add new
schema and ingest paths for:

- `pg_stat_activity` snapshots
- EXPLAIN plan snapshots for selected queries

Recommended schema additions:

- `activity_snapshots`
  - `id BIGSERIAL`
  - `database_instance_id TEXT`
  - `agent_id TEXT`
  - `collected_at TIMESTAMPTZ`
  - `pid BIGINT`
  - `state TEXT`
  - `wait_event_type TEXT`
  - `wait_event TEXT`
  - `query_duration_ms DOUBLE PRECISION`
  - `blocking_pid BIGINT NULL`
  - `query TEXT`
  - `user_name TEXT`
  - `application_name TEXT`
- `explain_plans`
  - `id BIGSERIAL`
  - `database_instance_id TEXT`
  - `agent_id TEXT`
  - `collected_at TIMESTAMPTZ`
  - `query_id TEXT`
  - `query TEXT`
  - `plan_json JSONB`

Add agent client methods for these payloads, mirroring the existing table/query
stats transport pattern.

**Verify**: `go test ./...` → exit 0

### Step 3: Build a backend diagnosis service

Add a dedicated service layer, for example `cloud/internal/diagnosis/service.go`,
that:

- loads recent database stats and delta inputs
- loads the latest activity snapshot
- loads the latest table stats snapshot
- loads the latest query stats snapshot
- loads the latest explain plans
- runs `analysis.Engine`
- writes findings through `repository.UpsertFindings`
- resolves stale findings

Do not bury this in HTTP handlers. The handler should trigger a service call or
enqueue a diagnosis run per ingest cycle.

Important shape requirement:

- backend diagnosis should run per `database_instance_id`
- database stats ingest and evidence snapshot ingests should converge on one
  orchestration path
- diagnosis should not require the frontend to make multiple calls to assemble
  consistency

**Verify**: `go test ./...` → exit 0

### Step 4: Remove agent-side rule execution from the happy path

Change `agent/internal/runner.go` so the agent:

- still collects database stats, activity, table stats, query stats, and
  explain evidence
- sends raw evidence to the API
- no longer treats `SendFindings` as the source of truth for normal operation

If a transitional phase is needed, keep `SendFindings` behind an explicit
feature flag or debug path, but the default production path must be
backend-owned diagnosis.

**Verify**: `go test ./...` → exit 0

### Step 5: Deprecate findings ingest cleanly

After the backend diagnosis path is working:

- mark `POST /api/findings/ingest` as deprecated in comments
- remove ordinary runtime dependence on it from the agent
- keep it only if needed for temporary migration compatibility, otherwise
  delete it

If deleting it would break current local flows, keep it for one transitional
plan and document the removal in plan 002's maintenance notes.

**Verify**: `rg -n "SendFindings|/api/findings/ingest" internal` → only transitional compatibility references remain, or none if fully removed

## Test plan

- Add backend tests for the diagnosis service covering:
  - evidence ingest for one database instance triggers diagnosis
  - mixed evidence types produce one diagnosis pass
  - stale findings still resolve when no issue is re-detected
- Model new handler tests after the existing API handler style in
  `cloud/internal/api/findings.go` and related tests if present
- Add a regression test ensuring the agent can run one collection cycle without
  local rule execution on the default path
- Verification: `go test ./...` → all pass

## Done criteria

- [ ] The default agent runtime path does not run diagnosis rules locally
- [ ] Backend-owned diagnosis exists and runs against stored evidence
- [ ] Activity and EXPLAIN evidence are persisted through backend ingest paths
- [ ] `go test ./...` exits 0
- [ ] `cd frontend && npm run build` exits 0
- [ ] No files outside the in-scope list are modified (`git status`)
- [ ] `docs/plans/README.md` status row updated

## STOP conditions

- The live code no longer shows the agent as the diagnosis executor in the
  locations quoted above
- Activity or EXPLAIN collectors do not expose enough fields to reconstruct the
  existing rules without redesign
- Implementing the backend diagnosis cycle requires changing the public
  database-instance registration flow
- The migration strategy requires destructive rewrites of existing evidence
  tables instead of additive changes

## Maintenance notes

- Reviewers should focus on whether diagnosis now has a single source of truth
  in the backend rather than split ownership.
- If asynchronous diagnosis jobs are introduced later, keep the service API
  introduced here as the synchronous core and move only orchestration outward.
- Plan 002 assumes this plan delivers stable backend-owned findings generation.
