package rules

import (
	"fmt"
	"strings"

	"github.com/okumujustine/postgresome/internal/analysis"
	"github.com/okumujustine/postgresome/internal/metrics"
)

const (
	slowQueryWarningMeanExecutionTimeMs  = 100
	slowQueryCriticalMeanExecutionTimeMs = 1000

	queryPreviewMaxLength = 80
)

// SlowQueryRule detects individual queries with a high mean execution time,
// which can indicate missing indexes, inefficient query plans, or queries
// that need to be rewritten.
type SlowQueryRule struct{}

func (r SlowQueryRule) Name() string {
	return "slow_query"
}

func (r SlowQueryRule) Analyze(snapshot metrics.QueryStatsSnapshot) []analysis.Finding {
	findings := make([]analysis.Finding, 0)

	for _, query := range snapshot.Queries {
		switch {
		case query.MeanExecutionTimeMs >= slowQueryCriticalMeanExecutionTimeMs:
			findings = append(findings, analysis.Finding{
				DetectedAt:         snapshot.CollectedAt,
				DatabaseInstanceID: snapshot.DatabaseInstanceID,
				Severity:           "critical",
				Category:           "queries",
				Title:              "Critical slow query detected",
				Message:            fmt.Sprintf("Query %q has a mean execution time of %.1f ms.", previewQuery(query.Query, queryPreviewMaxLength), query.MeanExecutionTimeMs),
				Recommendation:     "Review query execution plan, indexes, joins, and filters.",
			})

		case query.MeanExecutionTimeMs >= slowQueryWarningMeanExecutionTimeMs:
			findings = append(findings, analysis.Finding{
				DetectedAt:         snapshot.CollectedAt,
				DatabaseInstanceID: snapshot.DatabaseInstanceID,
				Severity:           "warning",
				Category:           "queries",
				Title:              "Slow query detected",
				Message:            fmt.Sprintf("Query %q has a mean execution time of %.1f ms.", previewQuery(query.Query, queryPreviewMaxLength), query.MeanExecutionTimeMs),
				Recommendation:     "Review query execution plan, indexes, joins, and filters.",
			})
		}
	}

	return findings
}

// previewQuery collapses whitespace in a SQL query and truncates it to a
// safe length so findings never include full, potentially large query text.
func previewQuery(query string, maxLength int) string {
	query = strings.Join(strings.Fields(query), " ")

	if len(query) <= maxLength {
		return query
	}

	return query[:maxLength] + "..."
}
