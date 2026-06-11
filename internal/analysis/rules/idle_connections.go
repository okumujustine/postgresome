package rules

import (
	"github.com/okumujustine/postgresome/internal/analysis"
	"github.com/okumujustine/postgresome/internal/metrics"
)

const (
	idleConnectionsWarningRatio  = 0.50
	idleConnectionsCriticalRatio = 0.80

	idleInTransactionCriticalThreshold = 5
)

// IdleConnectionRule detects when a large share of database sessions are
// idle, which can indicate connection leaks or oversized application
// connection pools.
type IdleConnectionRule struct{}

func (r IdleConnectionRule) Name() string {
	return "idle_connections"
}

func (r IdleConnectionRule) Analyze(snapshot metrics.DatabaseActivitySnapshot) []analysis.Finding {
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

	switch {
	case idlePercentage >= idleConnectionsCriticalRatio || idleInTransaction >= idleInTransactionCriticalThreshold:
		return []analysis.Finding{{
			DetectedAt:         snapshot.CollectedAt,
			DatabaseInstanceID: snapshot.DatabaseInstanceID,
			Severity:           "critical",
			Category:           "connections",
			Title:              "Critical idle connection pressure detected",
			Message:            "Most database sessions are idle or several sessions are idle in transaction.",
			Recommendation:     "Investigate connection leaks, idle transactions, and application pool configuration immediately.",
		}}

	case idlePercentage >= idleConnectionsWarningRatio || idleInTransaction > 0:
		return []analysis.Finding{{
			DetectedAt:         snapshot.CollectedAt,
			DatabaseInstanceID: snapshot.DatabaseInstanceID,
			Severity:           "warning",
			Category:           "connections",
			Title:              "High idle connection usage detected",
			Message:            "A large percentage of database sessions are idle.",
			Recommendation:     "Review application connection pooling and check whether connections are being held longer than necessary.",
		}}

	default:
		return nil
	}
}
