# Plan 002: Expand findings into a diagnosis-grade issue model

> **Executor instructions**: Follow this plan step by step. Run every
> verification command and confirm the expected result before moving to the
> next step. If anything in the "STOP conditions" section occurs, stop and
> report — do not improvise. When done, update the status row for this plan
> in `docs/plans/README.md` — unless a reviewer dispatched you and told you they
> maintain the index.
>
> **Drift check (run first)**: `git diff --stat bf5cb8f..HEAD -- backend/internal/analysis/finding.go backend/internal/storage/repository/findings.go backend/internal/api/findings.go backend/internal/api/dashboard_overview.go frontend/src/types/dashboard.ts migrations`
> If any in-scope file changed since this plan was written, compare the
> "Current state" excerpts against the live code before proceeding; on a
> mismatch, treat it as a STOP condition.

## Status

- **Priority**: P1
- **Effort**: L
- **Risk**: HIGH
- **Depends on**: docs/plans/001-backend-diagnosis-execution.md
- **Category**: direction
- **Planned at**: commit `bf5cb8f`, 2026-06-14

## Why this matters

The current findings model is enough for a queue, but not enough for a premium
diagnosis workflow. It can say "what tripped a rule," but it cannot cleanly
express impact, confidence, baseline comparison, evidence summaries, or fix
verification guidance. This plan upgrades findings into first-class diagnosis
issues that the backend and frontend can both depend on.

## Current state

- `backend/internal/analysis/finding.go` — current finding fields stop at title,
  message, recommendation, rule identity, and current/threshold values.
- `backend/internal/storage/repository/findings.go` — persistence only upserts the
  current compact finding shape.
- `backend/internal/api/dashboard_overview.go` and `backend/internal/api/findings.go` — API DTOs
  mirror the same compact shape.
- `frontend/src/types/dashboard.ts` — frontend types still use
  `DashboardFinding` with only title/message/recommendation plus lifecycle and
  threshold fields.

Current shape excerpt:

- `backend/internal/analysis/finding.go:8-31`
  ```go
  type Finding struct {
      ID string
      DetectedAt time.Time
      DatabaseInstanceID string
      AgentID string
      Severity string
      Category string
      Title string
      Message string
      Recommendation string
      RuleKey string
      ResourceType string
      ResourceName string
      CurrentValue float64
      ThresholdValue float64
  }
  ```

- `frontend/src/types/dashboard.ts:24-38`
  ```ts
  export interface DashboardFinding {
    id: string;
    severity: string;
    category: string;
    title: string;
    message: string;
    recommendation: string;
    status: string;
    rule_key: string;
    resource_type: string;
    resource_name: string;
    current_value: number;
    threshold_value: number;
  }
  ```

Product vocabulary to honor from `docs/architecture/diagnosis-platform.md`:

- `Problem`
- `Evidence`
- `Impact`
- `Suggested action`

Use that language in field names, comments, and DTOs when it improves clarity.

## Commands you will need

| Purpose | Command | Expected on success |
|---------|---------|---------------------|
| Backend tests | `cd backend && go test ./...` | exit 0 |
| Frontend build | `cd frontend && npm run build` | exit 0 |
| Frontend lint | `cd frontend && npm run lint` | exit 0 |

## Scope

**In scope**:

- `backend/internal/analysis/finding.go`
- `backend/internal/storage/repository/findings.go`
- `backend/internal/api/findings.go`
- `backend/internal/api/dashboard_overview.go`
- new issue detail API file(s) under `backend/internal/api`
- `frontend/src/types/dashboard.ts`
- any frontend API client files needed to consume the richer issue detail shape
- new migrations that alter `findings`

**Out of scope**:

- detector logic changes beyond emitting the richer structure
- major frontend layout rewrites
- alert delivery

## Git workflow

- Branch: `advisor/002-diagnosis-issue-model`
- Commit messages should stay short and descriptive, matching existing history
- Do NOT push or open a PR unless instructed

## Steps

### Step 1: Define the richer diagnosis issue model

Extend `analysis.Finding` into a diagnosis-grade issue shape. Keep existing
fields for compatibility, but add:

- `ProblemSummary string`
- `ImpactSummary string`
- `Confidence string` or `ConfidenceScore float64`
- `BaselineValue float64`
- `BaselineLabel string`
- `ChangeSummary string`
- `EvidenceSummary string`
- `SuggestedAction string`
- `VerificationHint string`
- `ObservedAt time.Time` if distinct from `DetectedAt`

Recommended compatibility strategy:

- keep `Title`, `Message`, and `Recommendation` as compatibility aliases for
  the old API contract during transition
- define one canonical meaning for each:
  - `Title` → short issue headline
  - `Message` → problem statement
  - `Recommendation` → suggested action

**Verify**: `cd backend && go test ./...` → exit 0

### Step 2: Add schema support for diagnosis-grade findings

Add a migration after `000004_findings_lifecycle.sql` that introduces new
columns to `findings`, such as:

- `impact_summary TEXT`
- `confidence_score DOUBLE PRECISION`
- `confidence_label TEXT`
- `baseline_value DOUBLE PRECISION`
- `baseline_label TEXT`
- `change_summary TEXT`
- `evidence_summary TEXT`
- `suggested_action TEXT`
- `verification_hint TEXT`
- `metadata JSONB DEFAULT '{}'::jsonb`

Use additive schema changes only. Do not remove the existing columns in this
plan. Backfill old rows conservatively where possible.

**Verify**: `cd backend && go test ./...` → exit 0

### Step 3: Update repository read/write paths

Update `backend/internal/storage/repository/findings.go` so:

- upserts write the new columns
- list queries return enough fields for queue-level diagnosis summaries
- one new repository method returns full issue detail by `id`

Preferred repository split:

- queue/list projection for findings pages
- detail projection for issue detail pages

Do not overload one DTO for both if it creates a large partial-response type.

**Verify**: `cd backend && go test ./...` → exit 0

### Step 4: Add a dedicated issue detail endpoint

Add an endpoint such as:

- `GET /api/issues/{id}`
  or
- `GET /api/findings/{id}`

Return a payload that includes:

- issue identity and lifecycle fields
- problem summary
- evidence summary
- impact summary
- suggested action
- confidence
- current/baseline/threshold values
- first seen / last seen / occurrence count
- any historical notes available so far

Keep `GET /api/findings` optimized for queue views. Do not turn the list
endpoint into a heavy detail payload.

**Verify**: `cd backend && go test ./...` → exit 0

### Step 5: Update frontend types and clients for the richer issue model

Update `frontend/src/types/dashboard.ts` and related client files so the
frontend can consume:

- a lightweight queue item shape
- a rich issue detail shape

If the current `DashboardFinding` name becomes misleading, rename the type
while preserving compatibility at call sites in a controlled way.

**Verify**: `cd frontend && npm run build` → exit 0

## Test plan

- Add repository tests for finding upsert/list/detail behavior covering:
  - new diagnosis fields persist correctly
  - old lifecycle behavior still works
  - detail fetch returns richer issue information
- Add API handler tests for the new detail endpoint
- Add frontend type-level/build coverage by ensuring the issue detail client
  compiles against the new DTO
- Verification: `cd backend && go test ./...` and `cd frontend && npm run build` → all pass

## Done criteria

- [ ] The findings schema stores diagnosis-grade issue fields
- [ ] The backend exposes a dedicated issue detail response
- [ ] Queue/list and detail projections are separated cleanly
- [ ] `cd backend && go test ./...` exits 0
- [ ] `cd frontend && npm run build` exits 0
- [ ] No files outside the in-scope list are modified (`git status`)
- [ ] `docs/plans/README.md` status row updated

## STOP conditions

- Plan 001 has not landed and the backend does not yet own normal findings generation
- The richer fields require removing existing compatibility fields immediately
- The frontend is already using a renamed issue detail API shape not reflected
  in the current excerpts
- Schema changes require rewriting historical findings rows destructively

## Maintenance notes

- Reviewers should check that the issue detail endpoint is optimized for
  diagnosis workflow, not overloaded with raw metrics.
- Plan 003 will populate the historical and confidence fields with better data;
  this plan only establishes the shape.
