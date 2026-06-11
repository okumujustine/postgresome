package analysis

import "github.com/okumujustine/postgresome/internal/metrics"

// Rule inspects database metrics and reports zero or more findings.
// Concrete rules are implemented in internal/analysis/rules and satisfy
// this interface structurally.
type Rule interface {
	Name() string
	Analyze(stats metrics.DatabaseStats, delta metrics.DatabaseMetricDelta) []Finding
}

// Engine runs a set of rules against collected metrics to produce findings.
type Engine struct {
	rules []Rule
}

func NewEngine(rules ...Rule) *Engine {
	return &Engine{rules: rules}
}

func (e *Engine) AnalyzeDatabaseMetrics(stats metrics.DatabaseStats, delta metrics.DatabaseMetricDelta) []Finding {
	findings := make([]Finding, 0)

	for _, rule := range e.rules {
		findings = append(findings, rule.Analyze(stats, delta)...)
	}

	return findings
}
