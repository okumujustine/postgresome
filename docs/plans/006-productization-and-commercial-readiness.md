# 006 Productization and Commercial Readiness

## Objective

Turn the diagnosis-first product direction into a commercially coherent first
premium release.

## Why this plan exists

The current codebase already reflects the diagnosis-first reset technically, but
the product still needs a stronger path from architecture to buyer value.

This plan aligns:

- product strategy
- diagnosis feature scope
- pricing and packaging
- messaging
- implementation priorities

## Scope

### In scope

- diagnosis-first premium release definition
- packaging and pricing model
- buyer and messaging alignment
- product requirements for the first premium version
- implementation sequencing for the next delivery phase

### Out of scope

- final marketing site build
- billing implementation
- enterprise procurement workflows
- full remediation automation

## Deliverables

1. Product strategy document
2. Buyer personas and GTM document
3. Pricing and packaging document
4. Premium PRD
5. Messaging document
6. Roadmap alignment update

## Recommended execution order

### Step 1: Confirm diagnosis-first premium scope

Task:

- lock the first premium release around the seven core diagnosis categories
- keep setup, diagnosis, evidence, and history as the primary product flow

Success criteria:

- the team can describe the premium version in one paragraph
- all major stakeholders use the same product language

### Step 2: Strengthen backend finding quality

Task:

- improve finding confidence, correlation, and evidence summaries
- ensure each rule outputs impact and recommended next step consistently

Files likely touched:

- `backend/internal/analysis/`
- `backend/internal/diagnosis/`
- `backend/internal/findingcontract/`

Success criteria:

- all core diagnosis rules emit diagnosis-grade findings

### Step 3: Strengthen source and provider abstraction

Task:

- make direct connection and managed source setup feel like one coherent product
- refine source metadata and provider-aware defaults

Files likely touched:

- `backend/internal/api/sources.go`
- `backend/internal/storage/repository/sources.go`
- `frontend/src/pages/SetupPage.tsx`

Success criteria:

- setup supports the product story for managed PostgreSQL users

### Step 4: Improve diagnosis and history trust surfaces

Task:

- improve evidence presentation
- improve confidence and verification language
- make history more obviously useful during investigations

Files likely touched:

- `frontend/src/components/finding-detail.tsx`
- `frontend/src/pages/HistoryPage.tsx`
- supporting API contracts if needed

Success criteria:

- a developer can understand a finding and trust why it exists

### Step 5: Add collaboration and lifecycle workflow

Task:

- add states like investigating, resolved, ignored
- support lightweight notes or ownership later if feasible

Files likely touched:

- finding APIs
- persistence layer
- diagnosis UI

Success criteria:

- findings can move through a basic operational workflow

### Step 6: Add commercial readiness surfaces

Task:

- create pricing page copy
- create product page copy
- define trial activation moments

Success criteria:

- the repo contains a coherent commercial story tied to the shipped product

## Risks

### Risk: product drifts back into dashboard thinking

Mitigation:

- use the finding-first workflow as a release gate

### Risk: too many weak findings reduce trust

Mitigation:

- prioritize quality and suppression over breadth

### Risk: commercial story overpromises before diagnosis quality is ready

Mitigation:

- keep marketing claims anchored to implemented diagnosis capabilities

## Completion criteria

This plan is complete when:

- the premium release is clearly defined
- product docs, pricing logic, and messaging are in the repo
- implementation sequencing reflects buyer value, not just technical order
