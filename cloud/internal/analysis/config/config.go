// Package config holds the tunable configuration for analyzer rules.
//
// Rule implementations (cloud/internal/analysis/rules) own PostgreSQL
// knowledge:
// how to calculate a metric, evaluate it, build evidence, and generate a
// fingerprint. This package owns product behavior: whether a rule runs, the
// thresholds it evaluates against, and the severity/title/description/
// recommendation text shown to users.
package config

// RuleConfig controls the tunable behavior of a single analyzer rule.
type RuleConfig struct {
	// Enabled controls whether the rule runs at all. Disabled rules must
	// return no findings.
	Enabled bool

	// Severity is the severity assigned to findings at the rule's base
	// (lowest) threshold tier. Rules with an escalated tier always use
	// "critical" for that tier, since crossing the critical threshold is
	// critical by definition regardless of configuration.
	Severity string

	// Thresholds holds the named numeric thresholds this rule evaluates
	// against. Key names are rule-specific; see DefaultAnalysisConfig for
	// the keys each rule expects.
	Thresholds map[string]float64

	// Title, Description, and Recommendation are the user-facing finding
	// content for this rule. Description fills Finding.Message.
	Title          string
	Description    string
	Recommendation string
}

// AnalysisConfig holds the configuration for every analyzer rule, keyed by
// rule key (format: "category.problem").
type AnalysisConfig struct {
	Rules map[string]RuleConfig
}

// Rule returns the configuration for the given rule key. If no
// configuration exists for that key, it returns a disabled RuleConfig so an
// unconfigured rule safely produces no findings instead of running with
// undefined thresholds.
func (c AnalysisConfig) Rule(key string) RuleConfig {
	if cfg, ok := c.Rules[key]; ok {
		return cfg
	}

	return RuleConfig{Enabled: false, Thresholds: map[string]float64{}}
}
