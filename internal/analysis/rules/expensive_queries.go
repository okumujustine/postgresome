package rules

import (
	"fmt"

	"github.com/okumujustine/postgresome/internal/analysis"
	"github.com/okumujustine/postgresome/internal/metrics"
)

const (
	expensiveQueryWarningTotalExecutionTimeMs  = 10000
	expensiveQueryCriticalTotalExecutionTimeMs = 60000
)

// ExpensiveQueryRule detects queries that consume a large amount of total
// database time across all their executions, which can indicate a query
// that runs too frequently, scans too much data, or both.
type ExpensiveQueryRule struct{}

func (r ExpensiveQueryRule) Name() string {
	return "expensive_query"
}

func (r ExpensiveQueryRule) Analyze(snapshot metrics.QueryStatsSnapshot) []analysis.Finding {
	findings := make([]analysis.Finding, 0)

	for _, query := range snapshot.Queries {
		switch {
		case query.TotalExecutionTimeMs >= expensiveQueryCriticalTotalExecutionTimeMs:
			findings = append(findings, analysis.Finding{
				DetectedAt:         snapshot.CollectedAt,
				DatabaseInstanceID: snapshot.DatabaseInstanceID,
				Severity:           "critical",
				Category:           "queries",
				Title:              "Critical query cost detected",
				Message:            fmt.Sprintf("Query %q has a total execution time of %.0f ms across %d calls.", previewQuery(query.Query, queryPreviewMaxLength), query.TotalExecutionTimeMs, query.Calls),
				Recommendation:     "This query consumes significant database time. Review frequency, indexing, and whether results can be cached or optimized.",
				RuleKey:            r.Name(),
				ResourceType:       "query",
				ResourceName:       query.QueryID,
				CurrentValue:       query.TotalExecutionTimeMs,
				ThresholdValue:     expensiveQueryCriticalTotalExecutionTimeMs,
			})

		case query.TotalExecutionTimeMs >= expensiveQueryWarningTotalExecutionTimeMs:
			findings = append(findings, analysis.Finding{
				DetectedAt:         snapshot.CollectedAt,
				DatabaseInstanceID: snapshot.DatabaseInstanceID,
				Severity:           "warning",
				Category:           "queries",
				Title:              "High total query cost detected",
				Message:            fmt.Sprintf("Query %q has a total execution time of %.0f ms across %d calls.", previewQuery(query.Query, queryPreviewMaxLength), query.TotalExecutionTimeMs, query.Calls),
				Recommendation:     "This query consumes significant database time. Review frequency, indexing, and whether results can be cached or optimized.",
				RuleKey:            r.Name(),
				ResourceType:       "query",
				ResourceName:       query.QueryID,
				CurrentValue:       query.TotalExecutionTimeMs,
				ThresholdValue:     expensiveQueryWarningTotalExecutionTimeMs,
			})
		}
	}

	return findings
}
