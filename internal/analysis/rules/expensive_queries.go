package rules

import (
	"github.com/okumujustine/postgresome/internal/analysis"
	"github.com/okumujustine/postgresome/internal/analysis/config"
	"github.com/okumujustine/postgresome/internal/metrics"
)

// ExpensiveQueryRule detects queries that consume a large amount of total
// database time across all their executions, which can indicate a query
// that runs too frequently, scans too much data, or both.
type ExpensiveQueryRule struct {
	Config config.RuleConfig
}

func (r ExpensiveQueryRule) Name() string {
	return config.RuleKeyExpensiveQuery
}

func (r ExpensiveQueryRule) Analyze(snapshot metrics.QueryStatsSnapshot) []analysis.Finding {
	if !r.Config.Enabled {
		return nil
	}

	warningThreshold := r.Config.Thresholds["warning_ms"]
	criticalThreshold := r.Config.Thresholds["critical_ms"]

	findings := make([]analysis.Finding, 0)

	for _, query := range snapshot.Queries {
		switch {
		case query.TotalExecutionTimeMs >= criticalThreshold:
			findings = append(findings, analysis.Finding{
				DetectedAt:         snapshot.CollectedAt,
				DatabaseInstanceID: snapshot.DatabaseInstanceID,
				Severity:           "critical",
				Category:           "queries",
				Title:              r.Config.Title,
				Message:            r.Config.Description,
				Recommendation:     r.Config.Recommendation,
				RuleKey:            r.Name(),
				ResourceType:       "query",
				ResourceName:       query.QueryID,
				CurrentValue:       query.TotalExecutionTimeMs,
				ThresholdValue:     criticalThreshold,
			})

		case query.TotalExecutionTimeMs >= warningThreshold:
			findings = append(findings, analysis.Finding{
				DetectedAt:         snapshot.CollectedAt,
				DatabaseInstanceID: snapshot.DatabaseInstanceID,
				Severity:           r.Config.Severity,
				Category:           "queries",
				Title:              r.Config.Title,
				Message:            r.Config.Description,
				Recommendation:     r.Config.Recommendation,
				RuleKey:            r.Name(),
				ResourceType:       "query",
				ResourceName:       query.QueryID,
				CurrentValue:       query.TotalExecutionTimeMs,
				ThresholdValue:     warningThreshold,
			})
		}
	}

	return findings
}
