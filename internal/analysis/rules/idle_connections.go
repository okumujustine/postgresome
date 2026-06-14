package rules

import (
	"github.com/okumujustine/postgresome/internal/analysis"
	"github.com/okumujustine/postgresome/internal/analysis/config"
	"github.com/okumujustine/postgresome/internal/metrics"
)

// IdleConnectionRule detects when a large share of database sessions are
// idle, which can indicate connection leaks or oversized application
// connection pools.
type IdleConnectionRule struct {
	Config config.RuleConfig
}

func (r IdleConnectionRule) Name() string {
	return config.RuleKeyIdleConnections
}

func (r IdleConnectionRule) Analyze(snapshot metrics.DatabaseActivitySnapshot) []analysis.Finding {
	if !r.Config.Enabled {
		return nil
	}

	total := 0
	idle := 0
	idleInTransaction := 0

	for _, a := range snapshot.Activities {
		// Background workers (checkpointer, autovacuum launcher, etc.) have
		// no database name and aren't client sessions.
		if a.DatabaseName == "" {
			continue
		}

		total++

		switch a.State {
		case "idle":
			idle++
		case "idle in transaction", "idle in transaction (aborted)":
			idleInTransaction++
		}
	}

	if total == 0 {
		return nil
	}

	idlePercentage := float64(idle) / float64(total)
	warningRatio := r.Config.Thresholds["warning_ratio"]
	criticalRatio := r.Config.Thresholds["critical_ratio"]
	idleInTransactionCritical := r.Config.Thresholds["idle_in_transaction_critical"]

	switch {
	case idlePercentage >= criticalRatio || float64(idleInTransaction) >= idleInTransactionCritical:
		return []analysis.Finding{{
			DetectedAt:         snapshot.CollectedAt,
			DatabaseInstanceID: snapshot.DatabaseInstanceID,
			Severity:           "critical",
			Category:           "connections",
			Title:              r.Config.Title,
			Message:            r.Config.Description,
			Recommendation:     r.Config.Recommendation,
			RuleKey:            r.Name(),
			ResourceType:       "database",
			CurrentValue:       idlePercentage,
			ThresholdValue:     criticalRatio,
		}}

	case idlePercentage >= warningRatio || idleInTransaction > 0:
		return []analysis.Finding{{
			DetectedAt:         snapshot.CollectedAt,
			DatabaseInstanceID: snapshot.DatabaseInstanceID,
			Severity:           r.Config.Severity,
			Category:           "connections",
			Title:              r.Config.Title,
			Message:            r.Config.Description,
			Recommendation:     r.Config.Recommendation,
			RuleKey:            r.Name(),
			ResourceType:       "database",
			CurrentValue:       idlePercentage,
			ThresholdValue:     warningRatio,
		}}

	default:
		return nil
	}
}
