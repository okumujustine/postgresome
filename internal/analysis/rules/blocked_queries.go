package rules

import (
	"fmt"

	"github.com/okumujustine/postgresome/internal/analysis"
	"github.com/okumujustine/postgresome/internal/metrics"
)

const (
	blockedQueryWarningThreshold  = 1
	blockedQueryCriticalThreshold = 5
)

// BlockedQueryRule detects sessions waiting on locks held by other
// transactions, which can cause slow applications, request timeouts, and
// database pileups.
type BlockedQueryRule struct{}

func (r BlockedQueryRule) Name() string {
	return "blocked_queries"
}

func (r BlockedQueryRule) Analyze(snapshot metrics.DatabaseActivitySnapshot) []analysis.Finding {
	blocked := 0

	for _, a := range snapshot.Activities {
		if a.WaitEventType == "Lock" {
			blocked++
		}
	}

	switch {
	case blocked > blockedQueryCriticalThreshold:
		return []analysis.Finding{{
			DetectedAt:         snapshot.CollectedAt,
			DatabaseInstanceID: snapshot.DatabaseInstanceID,
			Severity:           "critical",
			Category:           "locks",
			Title:              "Critical database lock contention detected",
			Message:            "Multiple database sessions are blocked waiting for locks.",
			Recommendation:     "Identify blocking queries and investigate transaction handling immediately.",
			RuleKey:            r.Name(),
			ResourceType:       "database",
			CurrentValue:       float64(blocked),
			ThresholdValue:     blockedQueryCriticalThreshold,
		}}

	case blocked >= blockedQueryWarningThreshold:
		return []analysis.Finding{{
			DetectedAt:         snapshot.CollectedAt,
			DatabaseInstanceID: snapshot.DatabaseInstanceID,
			Severity:           "warning",
			Category:           "locks",
			Title:              "Blocked database queries detected",
			Message:            fmt.Sprintf("%d database sessions are waiting on locks.", blocked),
			Recommendation:     "Review long-running transactions and queries holding locks.",
			RuleKey:            r.Name(),
			ResourceType:       "database",
			CurrentValue:       float64(blocked),
			ThresholdValue:     blockedQueryWarningThreshold,
		}}

	default:
		return nil
	}
}
