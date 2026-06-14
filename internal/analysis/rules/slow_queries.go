package rules

import (
	"strings"

	"github.com/okumujustine/postgresome/internal/analysis"
	"github.com/okumujustine/postgresome/internal/analysis/config"
	"github.com/okumujustine/postgresome/internal/metrics"
)

const queryPreviewMaxLength = 80

// SlowQueryRule detects individual queries with a high mean execution time,
// which can indicate missing indexes, inefficient query plans, or queries
// that need to be rewritten.
type SlowQueryRule struct {
	Config config.RuleConfig
}

func (r SlowQueryRule) Name() string {
	return config.RuleKeySlowQuery
}

func (r SlowQueryRule) Analyze(snapshot metrics.QueryStatsSnapshot) []analysis.Finding {
	if !r.Config.Enabled {
		return nil
	}

	warningThreshold := r.Config.Thresholds["warning_ms"]
	criticalThreshold := r.Config.Thresholds["critical_ms"]

	findings := make([]analysis.Finding, 0)

	for _, query := range snapshot.Queries {
		switch {
		case query.MeanExecutionTimeMs >= criticalThreshold:
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
				CurrentValue:       query.MeanExecutionTimeMs,
				ThresholdValue:     criticalThreshold,
			})

		case query.MeanExecutionTimeMs >= warningThreshold:
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
				CurrentValue:       query.MeanExecutionTimeMs,
				ThresholdValue:     warningThreshold,
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
