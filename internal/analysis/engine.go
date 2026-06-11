package analysis

import "github.com/okumujustine/postgresome/internal/metrics"

// Rule inspects database metrics and reports zero or more findings.
// Concrete rules are implemented in internal/analysis/rules and satisfy
// this interface structurally.
type Rule interface {
	Name() string
	Analyze(stats metrics.DatabaseStats, delta metrics.DatabaseMetricDelta) []Finding
}

// ActivityRule inspects a snapshot of pg_stat_activity sessions and reports
// zero or more findings. Concrete rules are implemented in
// internal/analysis/rules and satisfy this interface structurally.
type ActivityRule interface {
	Name() string
	Analyze(snapshot metrics.DatabaseActivitySnapshot) []Finding
}

// Engine runs a set of rules against collected metrics to produce findings.
type Engine struct {
	rules         []Rule
	activityRules []ActivityRule
}

func NewEngine(rules ...Rule) *Engine {
	return &Engine{rules: rules}
}

// RegisterActivityRules adds rules that analyze pg_stat_activity snapshots.
func (e *Engine) RegisterActivityRules(activityRules ...ActivityRule) {
	e.activityRules = append(e.activityRules, activityRules...)
}

func (e *Engine) AnalyzeDatabaseMetrics(stats metrics.DatabaseStats, delta metrics.DatabaseMetricDelta) []Finding {
	findings := make([]Finding, 0)

	for _, rule := range e.rules {
		findings = append(findings, rule.Analyze(stats, delta)...)
	}

	return findings
}

// AnalyzeDatabaseActivity runs all registered activity rules against a
// pg_stat_activity snapshot.
func (e *Engine) AnalyzeDatabaseActivity(snapshot metrics.DatabaseActivitySnapshot) []Finding {
	findings := make([]Finding, 0)

	for _, rule := range e.activityRules {
		findings = append(findings, rule.Analyze(snapshot)...)
	}

	return findings
}
