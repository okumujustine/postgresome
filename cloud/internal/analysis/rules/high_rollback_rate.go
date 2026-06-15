package rules

import (
	"github.com/okumujustine/postgresome/cloud/internal/analysis"
	"github.com/okumujustine/postgresome/cloud/internal/analysis/config"
	"github.com/okumujustine/postgresome/shared/metrics"
)

// HighRollbackRateRule detects unhealthy transaction behavior where a large
// share of transactions are being rolled back instead of committed.
type HighRollbackRateRule struct {
	Config config.RuleConfig
}

func (r HighRollbackRateRule) Name() string {
	return config.RuleKeyHighRollbackRate
}

func (r HighRollbackRateRule) Analyze(stats metrics.DatabaseStats, delta metrics.DatabaseMetricDelta) []analysis.Finding {
	if !r.Config.Enabled {
		return nil
	}

	rate := delta.RollbackRate
	warningThreshold := r.Config.Thresholds["warning"]
	criticalThreshold := r.Config.Thresholds["critical"]

	switch {
	case rate > criticalThreshold:
		return []analysis.Finding{{
			DetectedAt:         delta.CollectedAt,
			DatabaseInstanceID: delta.DatabaseInstanceID,
			Severity:           "critical",
			Category:           "transactions",
			Title:              r.Config.Title,
			Message:            r.Config.Description,
			Recommendation:     r.Config.Recommendation,
			RuleKey:            r.Name(),
			ResourceType:       "database",
			CurrentValue:       rate,
			ThresholdValue:     criticalThreshold,
		}}

	case rate >= warningThreshold:
		return []analysis.Finding{{
			DetectedAt:         delta.CollectedAt,
			DatabaseInstanceID: delta.DatabaseInstanceID,
			Severity:           r.Config.Severity,
			Category:           "transactions",
			Title:              r.Config.Title,
			Message:            r.Config.Description,
			Recommendation:     r.Config.Recommendation,
			RuleKey:            r.Name(),
			ResourceType:       "database",
			CurrentValue:       rate,
			ThresholdValue:     warningThreshold,
		}}

	default:
		return nil
	}
}
