# Plan 005: Add issue regression, fix verification, and diagnosis-first alerts

> **Executor instructions**: Follow this plan step by step. Run every
> verification command and confirm the expected result before moving to the
> next step. If anything in the "STOP conditions" section occurs, stop and
> report — do not improvise. When done, update the status row for this plan
> in `docs/plans/README.md` — unless a reviewer dispatched you and told you they
> maintain the index.
>
> **Drift check (run first)**: `git diff --stat bf5cb8f..HEAD -- cloud/internal/storage/repository/findings.go cloud/internal/api frontend/src migrations`
> If any in-scope file changed since this plan was written, compare the
> "Current state" excerpts against the live code before proceeding; on a
> mismatch, treat it as a STOP condition.

## Status

- **Priority**: P2
- **Effort**: L
- **Risk**: MED
- **Depends on**: docs/plans/002-diagnosis-issue-model.md, docs/plans/003-historical-reasoning-and-first-pack.md, docs/plans/004-frontend-diagnosis-workflow.md
- **Category**: direction
- **Planned at**: commit `bf5cb8f`, 2026-06-14

## Why this matters

Premium diagnosis software should not stop at detection. It should remember
whether an issue is recurring, tell the user whether things are improving, and
confirm whether a fix worked. That same lifecycle intelligence is what makes
alerts useful instead of noisy threshold spam.

## Current state

- `migrations/000004_findings_lifecycle.sql` — current lifecycle only supports
  `open` and `resolved`.
- `cloud/internal/storage/repository/findings.go` — stale findings resolve, and
  recurring detections reopen the same row, but there is no explicit
  `regressed`, `improving`, or `verified_fixed` state.
- frontend issue views show open/resolved but do not yet present recurrence or
  verification.

Current lifecycle excerpt:

- `migrations/000004_findings_lifecycle.sql:27-34`
  ```sql
  ALTER TABLE findings
      ADD CONSTRAINT findings_status_check CHECK (status IN ('open', 'resolved'));
  ```

- `cloud/internal/storage/repository/findings.go:18-34`
  ```sql
  ON CONFLICT (database_instance_id, fingerprint) DO UPDATE SET
      ...
      status = 'open',
      resolved_at = NULL,
      occurrence_count = findings.occurrence_count + 1,
      last_seen_at = EXCLUDED.last_seen_at
  ```

## Commands you will need

| Purpose | Command | Expected on success |
|---------|---------|---------------------|
| Backend tests | `go test ./...` | exit 0 |
| Frontend build | `cd frontend && npm run build` | exit 0 |
| Frontend lint | `cd frontend && npm run lint` | exit 0 |

## Scope

**In scope**:

- `migrations/*` affecting finding lifecycle
- `cloud/internal/storage/repository/findings.go`
- issue detail/list API files under `cloud/internal/api`
- frontend issue views/components that surface lifecycle and verification state
- alert integration scaffolding if it already exists; otherwise add backend-only
  alert payload generation hooks without full provider integration

**Out of scope**:

- full Slack/email delivery if no notification plumbing exists yet
- unrelated account/team/billing work

## Git workflow

- Branch: `advisor/005-regression-verification-alerts`
- Use the repo's short descriptive commit message style
- Do NOT push or open a PR unless instructed

## Steps

### Step 1: Expand lifecycle states

Add explicit lifecycle states or flags to support:

- recurring/regressed
- improving
- verified fixed

Recommended approach:

- keep `status` for coarse open/resolved state
- add supplementary fields such as:
  - `regression_count`
  - `last_regressed_at`
  - `improving_since`
  - `verified_fixed_at`
  - `verification_status`

This avoids overloading one status enum with too many meanings.

**Verify**: `go test ./...` → exit 0

### Step 2: Add fix verification rules

For the first diagnosis pack, define verification checks that can confirm a fix:

- table bloat / vacuum:
  - dead rows down
  - vacuum recency improved
  - sequential scan pressure reduced
- slow queries:
  - mean or total execution time reduced
  - disk reads reduced

Store the verification result on the issue so it can be surfaced in detail
views and alert payloads.

**Verify**: `go test ./...` → exit 0

### Step 3: Surface regression and verification in the UI

Update issue queue and issue detail so users can tell:

- this issue recurred
- this issue is improving
- this fix appears verified

Do not bury this in small metadata text. It should be prominent enough to
change operator behavior.

**Verify**: `cd frontend && npm run build` → exit 0

### Step 4: Generate diagnosis-first alert payloads

Add backend alert payload generation that emits:

- problem headline
- affected resource
- evidence summary
- impact summary
- suggested next step
- lifecycle hint (`new`, `regressed`, `resolved`, `verified fixed`)

If notification transport does not exist yet, stop at payload generation plus
tests and documentation hooks. Do not build a half-finished notification stack
in this plan.

**Verify**: `go test ./...` → exit 0

## Test plan

- Add lifecycle tests for recurring, resolving, regressing, and verifying flows
- Add verification tests for the first diagnosis pack categories
- Add API tests for lifecycle fields in queue and detail responses
- Add frontend build/lint coverage to ensure lifecycle badges/details compile
- Verification: `go test ./...`, `cd frontend && npm run build`, and `cd frontend && npm run lint` → all pass

## Done criteria

- [ ] Findings support recurrence/regression and verification semantics
- [ ] The first diagnosis pack can mark fixes as verified when evidence improves
- [ ] UI surfaces lifecycle and verification clearly
- [ ] Diagnosis-first alert payload generation exists
- [ ] `go test ./...` exits 0
- [ ] `cd frontend && npm run build` exits 0
- [ ] `cd frontend && npm run lint` exits 0
- [ ] No files outside the in-scope list are modified (`git status`)
- [ ] `docs/plans/README.md` status row updated

## STOP conditions

- Plans 002 through 004 have not landed
- Verification logic cannot be implemented without historical helpers from plan 003
- Alert delivery infrastructure is absent and stakeholders expect full provider
  integration in this same plan
- Lifecycle semantics become ambiguous enough that a product decision is needed

## Maintenance notes

- Reviewers should focus on whether recurrence and verification are backed by
  evidence, not just timestamps.
- Once notification plumbing exists, reuse the alert payload generator here
  rather than duplicating diagnosis summary logic inside integrations.
