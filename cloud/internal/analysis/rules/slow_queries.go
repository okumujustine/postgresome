package rules

import (
	"fmt"
	"strings"

	"github.com/okumujustine/postgresome/cloud/internal/analysis"
	"github.com/okumujustine/postgresome/cloud/internal/analysis/config"
	"github.com/okumujustine/postgresome/shared/metrics"
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
			queryPreview := previewQuery(query.Query, queryPreviewMaxLength)
			findings = append(findings, analysis.Finding{
				DetectedAt:         snapshot.CollectedAt,
				DatabaseInstanceID: snapshot.DatabaseInstanceID,
				Severity:           "critical",
				Category:           "queries",
				Title:              r.Config.Title,
				Message:            r.Config.Description,
				Recommendation:     r.Config.Recommendation,
				ProblemSummary:     fmt.Sprintf("A frequently executed query is averaging %.1f ms, which is above the critical %.1f ms threshold.", query.MeanExecutionTimeMs, criticalThreshold),
				EvidenceSummary:    fmt.Sprintf("Query %s is averaging %.1f ms across %d calls in the latest sample. Observed min/max latency was %.1f ms / %.1f ms.", queryPreview, query.MeanExecutionTimeMs, query.Calls, query.MinExecutionTimeMs, query.MaxExecutionTimeMs),
				ImpactSummary:      "Repeated executions of this query can dominate request latency and increase overall database load.",
				SuggestedAction:    "Review the query plan, check whether the filtering columns are indexed, and compare this plan with a faster historical baseline if one exists.",
				ConfidenceLabel:    "high",
				ConfidenceScore:    0.9,
				ChangeSummary:      "The latest collection crossed the critical slow-query threshold.",
				VerificationHint:   fmt.Sprintf("After tuning, confirm this query's mean execution time stays below %.1f ms.", warningThreshold),
				RuleKey:            r.Name(),
				ResourceType:       "query",
				ResourceName:       query.QueryID,
				CurrentValue:       query.MeanExecutionTimeMs,
				ThresholdValue:     criticalThreshold,
			})

		case query.MeanExecutionTimeMs >= warningThreshold:
			queryPreview := previewQuery(query.Query, queryPreviewMaxLength)
			findings = append(findings, analysis.Finding{
				DetectedAt:         snapshot.CollectedAt,
				DatabaseInstanceID: snapshot.DatabaseInstanceID,
				Severity:           r.Config.Severity,
				Category:           "queries",
				Title:              r.Config.Title,
				Message:            r.Config.Description,
				Recommendation:     r.Config.Recommendation,
				ProblemSummary:     fmt.Sprintf("A query is running slower than expected at %.1f ms on average, above the %.1f ms warning threshold.", query.MeanExecutionTimeMs, warningThreshold),
				EvidenceSummary:    fmt.Sprintf("Query %s averaged %.1f ms across %d calls in the latest sample.", queryPreview, query.MeanExecutionTimeMs, query.Calls),
				ImpactSummary:      "If this trend continues, the query can become a noticeable source of user-facing latency and background load.",
				SuggestedAction:    "Inspect the plan for unnecessary sequential scans, review indexes on the filtered columns, and compare with a faster previous version of the query.",
				ConfidenceLabel:    "medium",
				ConfidenceScore:    0.76,
				ChangeSummary:      "The latest collection moved this query above the slow-query warning threshold.",
				VerificationHint:   fmt.Sprintf("Track this query after changes and confirm its mean execution time returns below %.1f ms.", warningThreshold),
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
