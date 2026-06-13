package rules

import (
	"fmt"

	"github.com/okumujustine/postgresome/internal/analysis"
	"github.com/okumujustine/postgresome/internal/metrics"
)

const (
	highConnectionsWarningThreshold  = 100
	highConnectionsCriticalThreshold = 500
)

// HighConnectionRule detects unusually high active connection counts,
// which may indicate connection leaks, missing connection pooling, poor
// pool configuration, or traffic spikes.
type HighConnectionRule struct{}

func (r HighConnectionRule) Name() string {
	return "high_connections"
}

func (r HighConnectionRule) Analyze(stats metrics.DatabaseStats, delta metrics.DatabaseMetricDelta) []analysis.Finding {
	connections := stats.ActiveConnections

	switch {
	case connections > highConnectionsCriticalThreshold:
		return []analysis.Finding{{
			DetectedAt:         delta.CollectedAt,
			DatabaseInstanceID: delta.DatabaseInstanceID,
			Severity:           "critical",
			Category:           "connections",
			Title:              "Critical database connection pressure",
			Message:            fmt.Sprintf("Database has %d active connections which may affect performance.", connections),
			Recommendation:     "Investigate connection spikes, idle connections, and connection pooling configuration.",
			RuleKey:            r.Name(),
			ResourceType:       "database",
			CurrentValue:       float64(connections),
			ThresholdValue:     highConnectionsCriticalThreshold,
		}}

	case connections > highConnectionsWarningThreshold:
		return []analysis.Finding{{
			DetectedAt:         delta.CollectedAt,
			DatabaseInstanceID: delta.DatabaseInstanceID,
			Severity:           "warning",
			Category:           "connections",
			Title:              "High database connections detected",
			Message:            fmt.Sprintf("Database currently has %d active connections.", connections),
			Recommendation:     "Review application connection pooling and check for connection leaks.",
			RuleKey:            r.Name(),
			ResourceType:       "database",
			CurrentValue:       float64(connections),
			ThresholdValue:     highConnectionsWarningThreshold,
		}}

	default:
		return nil
	}
}
