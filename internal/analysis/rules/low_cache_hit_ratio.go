package rules

import (
	"fmt"

	"github.com/okumujustine/postgresome/internal/analysis"
	"github.com/okumujustine/postgresome/internal/metrics"
)

const (
	lowCacheHitRatioWarningThreshold  = 0.95
	lowCacheHitRatioCriticalThreshold = 0.90
)

// LowCacheHitRatioRule detects poor PostgreSQL cache efficiency, which may
// indicate missing indexes, inefficient queries, or insufficient
// shared_buffers configuration.
type LowCacheHitRatioRule struct{}

func (r LowCacheHitRatioRule) Name() string {
	return "low_cache_hit_ratio"
}

func (r LowCacheHitRatioRule) Analyze(stats metrics.DatabaseStats, delta metrics.DatabaseMetricDelta) []analysis.Finding {
	// CacheHitRatio is 0 when there was no block activity in this interval
	// (safeRatio's divide-by-zero case), which is not a sign of poor cache
	// efficiency. Skip the rule until there is something to evaluate.
	if delta.BlocksHitInCacheDelta+delta.BlocksReadFromDiskDelta == 0 {
		return nil
	}

	ratio := delta.CacheHitRatio

	switch {
	case ratio < lowCacheHitRatioCriticalThreshold:
		return []analysis.Finding{{
			DetectedAt:         delta.CollectedAt,
			DatabaseInstanceID: delta.DatabaseInstanceID,
			Severity:           "critical",
			Category:           "cache",
			Title:              "Low cache hit ratio detected",
			Message:            fmt.Sprintf("PostgreSQL is reading frequently from disk. Current cache hit ratio is %.1f%%.", ratio*100),
			Recommendation:     "Investigate missing indexes, inefficient queries, and shared_buffers configuration.",
			RuleKey:            r.Name(),
			ResourceType:       "database",
			CurrentValue:       ratio,
			ThresholdValue:     lowCacheHitRatioCriticalThreshold,
		}}

	case ratio < lowCacheHitRatioWarningThreshold:
		return []analysis.Finding{{
			DetectedAt:         delta.CollectedAt,
			DatabaseInstanceID: delta.DatabaseInstanceID,
			Severity:           "warning",
			Category:           "cache",
			Title:              "Reduced cache efficiency detected",
			Message:            fmt.Sprintf("PostgreSQL cache hit ratio is currently %.1f%%.", ratio*100),
			Recommendation:     "Review query patterns, indexes, and memory configuration.",
			RuleKey:            r.Name(),
			ResourceType:       "database",
			CurrentValue:       ratio,
			ThresholdValue:     lowCacheHitRatioWarningThreshold,
		}}

	default:
		return nil
	}
}
