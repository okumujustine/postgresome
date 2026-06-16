package rules

import (
	"github.com/okumujustine/postgresome/backend/internal/analysis"
	"github.com/okumujustine/postgresome/backend/internal/analysis/config"
	"github.com/okumujustine/postgresome/backend/internal/metrics"
)

// BlockedQueryRule detects sessions waiting on locks held by other
// transactions, which can cause slow applications, request timeouts, and
// database pileups.
type BlockedQueryRule struct {
	Config config.RuleConfig
}

func (r BlockedQueryRule) Name() string {
	return config.RuleKeyBlockedQuery
}

func (r BlockedQueryRule) Analyze(snapshot metrics.DatabaseActivitySnapshot) []analysis.Finding {
	if !r.Config.Enabled {
		return nil
	}

	blocked := 0

	for _, a := range snapshot.Activities {
		if a.WaitEventType == "Lock" {
			blocked++
		}
	}

	warningThreshold := r.Config.Thresholds["warning"]
	criticalThreshold := r.Config.Thresholds["critical"]

	switch {
	case float64(blocked) > criticalThreshold:
		return []analysis.Finding{{
			DetectedAt:         snapshot.CollectedAt,
			DatabaseInstanceID: snapshot.DatabaseInstanceID,
			Severity:           "critical",
			Category:           "locks",
			Title:              r.Config.Title,
			Message:            r.Config.Description,
			Recommendation:     r.Config.Recommendation,
			RuleKey:            r.Name(),
			ResourceType:       "database",
			CurrentValue:       float64(blocked),
			ThresholdValue:     criticalThreshold,
		}}

	case float64(blocked) >= warningThreshold:
		return []analysis.Finding{{
			DetectedAt:         snapshot.CollectedAt,
			DatabaseInstanceID: snapshot.DatabaseInstanceID,
			Severity:           r.Config.Severity,
			Category:           "locks",
			Title:              r.Config.Title,
			Message:            r.Config.Description,
			Recommendation:     r.Config.Recommendation,
			RuleKey:            r.Name(),
			ResourceType:       "database",
			CurrentValue:       float64(blocked),
			ThresholdValue:     warningThreshold,
		}}

	default:
		return nil
	}
}
