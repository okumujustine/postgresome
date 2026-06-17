# Pricing and Packaging

This document defines how Postgresome should be packaged and sold as a premium
product.

## Pricing philosophy

Postgresome should not be priced like a dashboard.

It should be priced against:

- incident time saved
- faster root-cause analysis
- reduced cloud waste from unresolved database inefficiency
- delayed need for dedicated DBA hiring

The economic argument is simple:

```text
If one production database incident costs several engineer-hours,
the right diagnosis product pays for itself quickly.
```

## Packaging principles

1. Price on value, not raw chart count.
2. Make the diagnosis workflow available early.
3. Reserve collaboration, long history, and fleet workflows for higher tiers.
4. Keep direct setup and first value accessible.
5. Avoid too many confusing plan distinctions.

## Proposed plans

### Starter

For:

- individual developers
- very small teams
- initial production adoption

Includes:

- limited number of databases
- core diagnosis categories
- short historical retention
- diagnosis detail and evidence views
- manual diagnosis runs

Goal:

- drive conversion by proving value quickly

### Team

For:

- growing engineering teams
- multiple production databases
- ongoing weekly usage

Includes everything in Starter, plus:

- more databases
- longer history
- regression detection
- alerting
- collaboration states
- source-aware recommendations
- richer history and trend reasoning

Goal:

- become the default paid plan

### Enterprise

For:

- platform teams
- regulated environments
- companies with multiple teams or strict security requirements

Includes everything in Team, plus:

- SSO
- audit trail
- role controls
- private deployment options if needed later
- premium support
- longer retention and regional controls

Goal:

- support larger accounts without changing the core product story

## Feature gating guidance

### Keep in all paid plans

- diagnosis list and detail
- evidence-linked investigation
- query and table correlation
- core diagnosis categories

These are the product. They should not feel crippled.

### Best Team plan gates

- longer history
- alerts
- collaboration and ownership
- richer regression views
- more advanced recommendations

### Best Enterprise gates

- security and org controls
- deployment requirements
- support guarantees

## Metrics-based vs seat-based pricing

The likely best pricing structure is hybrid:

- database count or source count
- plus team/organization capability gates

Seat-only pricing would underprice infrastructure value.
Pure usage pricing may feel stressful or unpredictable.

## Pricing page narrative

The pricing page should emphasize:

- diagnosis quality
- incident response speed
- historical reasoning
- coverage across PostgreSQL providers

It should not emphasize:

- number of dashboards
- raw chart counts
- generic observability terminology

## Expansion logic

The natural expansion vectors are:

- more databases
- more teams
- longer retention
- more diagnosis automation

That aligns well with a premium product because value grows with operational
complexity.
