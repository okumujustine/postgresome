package rules

import (
	"fmt"

	"github.com/okumujustine/postgresome/internal/analysis"
	"github.com/okumujustine/postgresome/internal/metrics"
)

const (
	highRollbackRateWarningThreshold  = 0.05
	highRollbackRateCriticalThreshold = 0.15
)

// HighRollbackRateRule detects unhealthy transaction behavior where a large
// share of transactions are being rolled back instead of committed.
type HighRollbackRateRule struct{}

func (r HighRollbackRateRule) Name() string {
	return "high_rollback_rate"
}

func (r HighRollbackRateRule) Analyze(stats metrics.DatabaseStats, delta metrics.DatabaseMetricDelta) []analysis.Finding {
	rate := delta.RollbackRate

	switch {
	case rate > highRollbackRateCriticalThreshold:
		return []analysis.Finding{{
			DetectedAt:         delta.CollectedAt,
			DatabaseInstanceID: delta.DatabaseInstanceID,
			Severity:           "critical",
			Category:           "transactions",
			Title:              "High transaction rollback rate detected",
			Message:            "A large percentage of database work is being discarded due to rollbacks.",
			Recommendation:     "Investigate application failures, constraint violations, deadlocks, or timeout issues.",
		}}

	case rate >= highRollbackRateWarningThreshold:
		return []analysis.Finding{{
			DetectedAt:         delta.CollectedAt,
			DatabaseInstanceID: delta.DatabaseInstanceID,
			Severity:           "warning",
			Category:           "transactions",
			Title:              "Elevated transaction rollback rate detected",
			Message:            fmt.Sprintf("%.1f%% of recent database transactions are rolling back.", rate*100),
			Recommendation:     "Review application errors, failed queries, and transaction handling.",
		}}

	default:
		return nil
	}
}
