package rules

import (
	"github.com/okumujustine/postgresome/cloud/internal/analysis"
	"github.com/okumujustine/postgresome/cloud/internal/analysis/config"
	"github.com/okumujustine/postgresome/shared/metrics"
)

// LowCacheHitRatioRule detects poor PostgreSQL cache efficiency, which may
// indicate missing indexes, inefficient queries, or insufficient
// shared_buffers configuration.
type LowCacheHitRatioRule struct {
	Config config.RuleConfig
}

func (r LowCacheHitRatioRule) Name() string {
	return config.RuleKeyLowCacheHitRatio
}

func (r LowCacheHitRatioRule) Analyze(stats metrics.DatabaseStats, delta metrics.DatabaseMetricDelta) []analysis.Finding {
	if !r.Config.Enabled {
		return nil
	}

	// CacheHitRatio is 0 when there was no block activity in this interval
	// (safeRatio's divide-by-zero case), which is not a sign of poor cache
	// efficiency. Skip the rule until there is something to evaluate.
	if delta.BlocksHitInCacheDelta+delta.BlocksReadFromDiskDelta == 0 {
		return nil
	}

	ratio := delta.CacheHitRatio
	warningThreshold := r.Config.Thresholds["warning"]
	criticalThreshold := r.Config.Thresholds["critical"]

	switch {
	case ratio < criticalThreshold:
		return []analysis.Finding{{
			DetectedAt:         delta.CollectedAt,
			DatabaseInstanceID: delta.DatabaseInstanceID,
			Severity:           "critical",
			Category:           "cache",
			Title:              r.Config.Title,
			Message:            r.Config.Description,
			Recommendation:     r.Config.Recommendation,
			RuleKey:            r.Name(),
			ResourceType:       "database",
			CurrentValue:       ratio,
			ThresholdValue:     criticalThreshold,
		}}

	case ratio < warningThreshold:
		return []analysis.Finding{{
			DetectedAt:         delta.CollectedAt,
			DatabaseInstanceID: delta.DatabaseInstanceID,
			Severity:           r.Config.Severity,
			Category:           "cache",
			Title:              r.Config.Title,
			Message:            r.Config.Description,
			Recommendation:     r.Config.Recommendation,
			RuleKey:            r.Name(),
			ResourceType:       "database",
			CurrentValue:       ratio,
			ThresholdValue:     warningThreshold,
		}}

	default:
		return nil
	}
}
