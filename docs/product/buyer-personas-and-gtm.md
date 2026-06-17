# Buyer Personas and Go-To-Market

This document describes who should buy Postgresome first, why they buy, and
how Postgresome should be introduced to the market.

## Buyer personas

### 1. Senior backend engineer

Profile:

- owns application performance and reliability
- understands SQL and PostgreSQL basics
- does not want to become a full-time DBA

Pain:

- database issues take too long to isolate
- metrics and logs exist, but root cause still takes time
- repeated incidents erode confidence

Buying motivation:

- faster incident triage
- better evidence during debugging
- less time lost reading raw PostgreSQL signals

Best message:

```text
Understand what is wrong with Postgres in minutes, not hours.
```

### 2. Engineering manager or tech lead

Profile:

- responsible for delivery and reliability
- needs a repeatable way for teams to handle Postgres issues

Pain:

- too much senior-engineer time goes into incident debugging
- knowledge is trapped in a few people
- there is no consistent workflow for database investigations

Buying motivation:

- shorter incidents
- repeatable diagnosis process
- fewer escalations to specialists

Best message:

```text
Give every backend engineer a DBA-style investigation workflow.
```

### 3. Platform or SRE lead

Profile:

- supports multiple services or databases
- needs a shared operational language across teams

Pain:

- every team debugs Postgres differently
- incidents create noisy handoffs between app and platform teams
- provider-specific tooling creates fragmentation

Buying motivation:

- standardized diagnosis
- one tool across multiple Postgres sources
- better fleet visibility later

Best message:

```text
One diagnosis layer across RDS, Cloud SQL, Supabase, Neon, and self-hosted Postgres.
```

### 4. CTO at early scale

Profile:

- still close to engineering operations
- wants reliability without premature platform complexity

Pain:

- database issues block growth
- hiring DBA expertise too early is expensive
- generic tooling still leaves critical questions unanswered

Buying motivation:

- reduce operational risk
- postpone specialist hiring
- improve engineering leverage

Best message:

```text
Scale your Postgres operations without scaling firefighting.
```

## Go-to-market focus

### Best first segment

The best first customers are likely:

- startup and scale-up SaaS teams
- product-led engineering cultures
- teams already using managed Postgres
- companies that ship often and feel performance regressions quickly

### Why this segment first

- the pain is real
- the buyer is easy to identify
- setup can be simpler
- the value can be shown quickly
- PostgreSQL is often mission-critical for these teams

## Positioning angles

### Primary angle

Database doctor for developers.

### Secondary angle

PostgreSQL diagnosis and root-cause analysis for fast-moving teams.

### Supporting angle

Historical evidence and actionable findings across any PostgreSQL environment.

## Go-to-market motion

### Motion 1: self-serve trial

Strong for:

- startups
- individual backend engineers
- engineering leads evaluating tools

Requirements:

- fast setup
- first diagnosis in minutes
- meaningful seeded examples if live findings are not immediately available

### Motion 2: team expansion

Strong for:

- multi-engineer teams
- multiple production databases

Expansion hooks:

- more history
- collaboration
- alerts
- fleet support

### Motion 3: founder-led / technical sales

Strong for:

- higher-value teams
- companies with reliability pressure
- mixed estates or self-hosted environments

Helpful proof points:

- fewer incident hours
- clearer root-cause analysis
- reduced tuning guesswork

## Website and sales narrative

The story should be:

1. Your app slowed down.
2. Postgresome tells you which database issue matters.
3. It shows the evidence and what changed.
4. It recommends the next step.
5. Your team fixes the right thing faster.

Avoid telling the story through:

- generic observability language
- dashboards
- infrastructure buzzwords
- vague AI claims

## Proof points to pursue

The strongest proof points will be:

- time to first useful finding
- reduction in time to root cause
- reduction in repeated incidents
- number of useful findings per week
- recommendation acceptance or action rates

## Early objections and responses

### “We already have monitoring”

Response:

Monitoring shows signals. Postgresome explains what those signals mean for a
real PostgreSQL problem and what to do next.

### “Our cloud provider already gives us metrics”

Response:

Cloud metrics help surface symptoms. Postgresome connects symptoms to findings,
evidence, history, and recommended action.

### “We only need this occasionally”

Response:

That is acceptable. The cost of even occasional database incidents can exceed
the cost of a diagnosis product very quickly.

### “We have strong engineers already”

Response:

Strong engineers still lose time translating raw PostgreSQL signals into
decisions. Postgresome increases leverage; it does not replace expertise.
