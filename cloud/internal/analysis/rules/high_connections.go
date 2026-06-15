package rules

import (
	"github.com/okumujustine/postgresome/cloud/internal/analysis"
	"github.com/okumujustine/postgresome/cloud/internal/analysis/config"
	"github.com/okumujustine/postgresome/shared/metrics"
)

// HighConnectionRule detects unusually high active connection counts,
// which may indicate connection leaks, missing connection pooling, poor
// pool configuration, or traffic spikes.
type HighConnectionRule struct {
	Config config.RuleConfig
}

func (r HighConnectionRule) Name() string {
	return config.RuleKeyHighConnections
}

func (r HighConnectionRule) Analyze(stats metrics.DatabaseStats, delta metrics.DatabaseMetricDelta) []analysis.Finding {
	if !r.Config.Enabled {
		return nil
	}

	connections := float64(stats.ActiveConnections)
	warningThreshold := r.Config.Thresholds["warning"]
	criticalThreshold := r.Config.Thresholds["critical"]

	switch {
	case connections > criticalThreshold:
		return []analysis.Finding{{
			DetectedAt:         delta.CollectedAt,
			DatabaseInstanceID: delta.DatabaseInstanceID,
			Severity:           "critical",
			Category:           "connections",
			Title:              r.Config.Title,
			Message:            r.Config.Description,
			Recommendation:     r.Config.Recommendation,
			RuleKey:            r.Name(),
			ResourceType:       "database",
			CurrentValue:       connections,
			ThresholdValue:     criticalThreshold,
		}}

	case connections > warningThreshold:
		return []analysis.Finding{{
			DetectedAt:         delta.CollectedAt,
			DatabaseInstanceID: delta.DatabaseInstanceID,
			Severity:           r.Config.Severity,
			Category:           "connections",
			Title:              r.Config.Title,
			Message:            r.Config.Description,
			Recommendation:     r.Config.Recommendation,
			RuleKey:            r.Name(),
			ResourceType:       "database",
			CurrentValue:       connections,
			ThresholdValue:     warningThreshold,
		}}

	default:
		return nil
	}
}
