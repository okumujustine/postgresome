# Product Requirements Document

This PRD defines the first premium version of Postgresome as a diagnosis-first
PostgreSQL product.

## Product summary

Postgresome helps engineering teams understand PostgreSQL problems quickly by
turning raw signals into findings, evidence, and recommended next steps.

## Problem

Teams often have PostgreSQL metrics, logs, and traces but still struggle to
answer:

- what is wrong
- why it is happening
- what changed
- which query or table is involved
- what should be done next

This leads to:

- longer incidents
- repeated tuning guesswork
- dependence on a few senior engineers
- poor operational confidence

## Goal

Reduce time to understanding and time to action for PostgreSQL performance and
reliability issues.

## Non-goals

This version of Postgresome is not trying to become:

- a generic metrics dashboard
- a broad infrastructure monitoring suite
- a full SQL IDE
- a complete autonomous DBA platform

## Primary user journey

```text
Setup
-> run diagnosis
-> review diagnosis list
-> select one diagnosis
-> inspect evidence
-> take the recommended next step
```

## User stories

### Setup

- As a developer, I want to connect a PostgreSQL source quickly so I can start a diagnosis run without learning a complex collection system.

### Diagnosis list

- As an engineer, I want a prioritized list of findings so I know what to investigate first.

### Diagnosis detail

- As an engineer, I want one finding expanded into a clear report so I can understand the issue without interpreting many raw metrics.

### Evidence

- As an engineer, I want the evidence connected to the finding so I can trust the diagnosis and verify the recommendation.

### Query drill-down

- As an engineer, I want to inspect related queries when a finding points toward SQL behavior.

### History

- As an engineer, I want to see when the issue started, regressed, or improved so I can connect it to workload changes or releases.

## Core screens

### 1. Setup

Must include:

- source connection form
- source status
- run diagnosis action
- recent diagnosis result summary

### 2. Diagnosis list

Must include:

- prioritized queue
- severity
- title
- affected object
- time detected
- clear selection state

### 3. Diagnosis detail

Must include:

- problem
- explanation
- evidence
- impact
- recommended next step
- verification guidance

### 4. Query explorer

Must remain subordinate to diagnosis.

Must include:

- query list
- query detail
- relation to findings where applicable

### 5. History

Must include:

- finding opened
- regressed
- improving
- verified fixed

## MVP diagnosis categories

1. Slow query regressions
2. Lock contention
3. Vacuum problems
4. Table bloat
5. Missing indexes
6. Cache problems
7. Connection saturation/issues

## Backend requirements

The backend must:

- accept evidence from PostgreSQL sources
- persist raw evidence in TimescaleDB
- maintain historical context
- run diagnosis rules
- produce finding records
- produce evidence records linked to findings
- expose diagnosis-first APIs

## Diagnosis quality requirements

Every finding should answer:

- What is wrong?
- Why does Postgresome believe it?
- What changed?
- How confident are we?
- What should the user do next?

## V1 success criteria

- A new user can connect a source and run diagnosis with low friction.
- The diagnosis list helps the user identify a useful first investigation.
- The diagnosis detail page feels like a report, not a raw metric dump.
- Findings link meaningfully to related queries or tables where relevant.
- Historical context explains whether the problem is worsening, stable, or improving.

## Product metrics

### Activation

- time to first connected source
- time to first diagnosis run
- time to first viewed finding

### Value

- findings opened per active database
- query drill-down rate from findings
- recommended action engagement
- return usage per week

### Quality

- false positive rate
- missing-major-issue rate
- recommendation usefulness feedback

## Release criteria for first premium version

- diagnosis-first setup and UI are stable
- core diagnosis categories are implemented end-to-end
- findings have evidence, impact, and recommended action
- history is available for active findings
- direct connection flow is reliable
- the product feels clearly different from a dashboard
