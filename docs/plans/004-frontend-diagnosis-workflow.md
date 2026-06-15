# Plan 004: Build diagnosis-first frontend workflows on the richer issue APIs

> **Executor instructions**: Follow this plan step by step. Run every
> verification command and confirm the expected result before moving to the
> next step. If anything in the "STOP conditions" section occurs, stop and
> report — do not improvise. When done, update the status row for this plan
> in `docs/plans/README.md` — unless a reviewer dispatched you and told you they
> maintain the index.
>
> **Drift check (run first)**: `git diff --stat bf5cb8f..HEAD -- frontend/src`
> If any in-scope file changed since this plan was written, compare the
> "Current state" excerpts against the live code before proceeding; on a
> mismatch, treat it as a STOP condition.

## Status

- **Priority**: P2
- **Effort**: L
- **Risk**: MED
- **Depends on**: docs/plans/002-diagnosis-issue-model.md, docs/plans/003-historical-reasoning-and-first-pack.md
- **Category**: direction
- **Planned at**: commit `bf5cb8f`, 2026-06-14

## Why this matters

The backend can only feel premium if the frontend presents the diagnosis in the
right order. The current UI has already moved toward health, issues, queries,
and tables, but it still uses a compact dashboard-era data model. This plan
updates the UI to make issue detail, evidence, and suggested fixes the default
experience.

## Current state

- `frontend/src/pages/HealthPage.tsx` — already frames the product as a queue
  of priority issues.
- `frontend/src/pages/IssueDetailPage.tsx` — currently renders problem,
  evidence, and recommended fix from a relatively thin issue shape.
- `frontend/src/types/dashboard.ts` — still uses dashboard-oriented naming and
  compact finding types.
- `frontend/src/pages/QueriesPage.tsx` and `frontend/src/pages/TablesPage.tsx`
  already act as supporting evidence views.

Current issue detail excerpt:

- `frontend/src/pages/IssueDetailPage.tsx:149-194`
  ```tsx
  <Card title="Problem">...</Card>
  <Card title="Evidence">...</Card>
  <Card title="Recommended fix">...</Card>
  ```

Current type excerpt:

- `frontend/src/types/dashboard.ts:24-38`
  ```ts
  export interface DashboardFinding {
    id: string;
    severity: string;
    category: string;
    title: string;
    message: string;
    recommendation: string;
    ...
  }
  ```

Design constraints from the product docs:

- problems first
- evidence second
- historical charts only when needed
- avoid reverting toward dashboard-builder patterns

## Commands you will need

| Purpose | Command | Expected on success |
|---------|---------|---------------------|
| Frontend build | `cd frontend && npm run build` | exit 0 |
| Frontend lint | `cd frontend && npm run lint` | exit 0 |
| Backend tests | `go test ./...` | exit 0 |

## Scope

**In scope**:

- `frontend/src/types/dashboard.ts` or its successors
- `frontend/src/api/*`
- `frontend/src/pages/HealthPage.tsx`
- `frontend/src/pages/IssuesPage.tsx`
- `frontend/src/pages/IssueDetailPage.tsx`
- `frontend/src/pages/QueriesPage.tsx`
- `frontend/src/pages/TablesPage.tsx`
- shared issue/evidence components under `frontend/src/components/*`

**Out of scope**:

- general dashboard customization
- a metrics explorer
- a major navigation overhaul
- notification center UI

## Git workflow

- Branch: `advisor/004-frontend-diagnosis-workflow`
- Commit messages should stay short and descriptive
- Do NOT push or open a PR unless instructed

## Steps

### Step 1: Split queue types from issue detail types

Stop forcing one frontend type to serve every use case. Introduce:

- queue item type for health/issues lists
- issue detail type for the investigation page
- optional evidence/history subtypes

Rename types away from `DashboardFinding` if that improves clarity and does not
cause unnecessary churn.

**Verify**: `cd frontend && npm run build` → exit 0

### Step 2: Make Health and Issues consume diagnosis summaries, not thin findings

Update `HealthPage.tsx` and `IssuesPage.tsx` so each issue card/row shows:

- problem headline
- one-line evidence summary
- impact hint
- suggested fix summary when present
- lifecycle hint if regressed or recurring becomes available

Do not overload these views with full evidence blocks. They are queue views.

**Verify**: `cd frontend && npm run build` → exit 0

### Step 3: Turn issue detail into the primary investigation workflow

Update `IssueDetailPage.tsx` so the default vertical flow is:

1. Problem
2. Why the backend thinks this is happening
3. Evidence
4. Impact
5. Suggested fix
6. Historical context
7. How to verify the fix worked

Historical charts should render only when the API provides diagnosis-relevant
series. Avoid generic chart sections with no explanatory job.

**Verify**: `cd frontend && npm run build` → exit 0

### Step 4: Reposition Queries and Tables as supporting evidence views

Queries and Tables pages should reinforce issue investigation, not compete with
the health queue as primary destinations. Update copy, callouts, and linking so
they clearly support diagnosis detail.

Examples:

- link from issue detail into filtered query evidence
- link from issue detail into affected table evidence
- show why a query/table row is relevant to an open issue

**Verify**: `cd frontend && npm run build` → exit 0

## Test plan

- Add frontend API/client tests if the repo already uses that pattern; if not,
  rely on build verification and component-level tests where existing patterns
  are available
- Manually verify that issue detail still loads when optional historical
  evidence is absent
- Verification: `cd frontend && npm run build` and `cd frontend && npm run lint` → both exit 0

## Done criteria

- [ ] Queue views use diagnosis summaries instead of thin dashboard-era fields
- [ ] Issue detail follows problem → evidence → impact → fix → historical context
- [ ] Queries and Tables clearly support diagnosis workflows
- [ ] `cd frontend && npm run build` exits 0
- [ ] `cd frontend && npm run lint` exits 0
- [ ] `go test ./...` exits 0
- [ ] No files outside the in-scope list are modified (`git status`)
- [ ] `docs/plans/README.md` status row updated

## STOP conditions

- Plans 002 or 003 have not landed
- The API still returns only the compact finding shape from the current excerpts
- The UI work starts requiring a new app shell or navigation redesign
- Historical sections cannot be implemented without inventing unsupported API data

## Maintenance notes

- Reviewers should check that every new frontend section has a clear diagnosis
  job and does not slip back into chart-first presentation.
- If the codebase later renames `dashboard` modules to `diagnosis`, do it as a
  separate cleanup after this workflow is stable.
