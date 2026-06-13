package rules

import (
	"fmt"
	"time"

	"github.com/okumujustine/postgresome/internal/analysis"
	"github.com/okumujustine/postgresome/internal/metrics"
)

const (
	longRunningQueryWarningThreshold  = 1 * time.Minute
	longRunningQueryCriticalThreshold = 10 * time.Minute
)

// LongRunningQueryRule detects active queries that have been running for an
// unusually long time, which can indicate inefficient queries, missing
// indexes, large table scans, or stuck application requests.
type LongRunningQueryRule struct{}

func (r LongRunningQueryRule) Name() string {
	return "long_running_queries"
}

func (r LongRunningQueryRule) Analyze(snapshot metrics.DatabaseActivitySnapshot) []analysis.Finding {
	findings := make([]analysis.Finding, 0)

	for _, a := range snapshot.Activities {
		if a.State != "active" || a.QueryStartedAt == nil {
			continue
		}

		duration := snapshot.CollectedAt.Sub(*a.QueryStartedAt)
		if duration < longRunningQueryWarningThreshold {
			continue
		}

		minutes := duration.Minutes()
		session := describeSession(a)

		if duration >= longRunningQueryCriticalThreshold {
			findings = append(findings, analysis.Finding{
				DetectedAt:         snapshot.CollectedAt,
				DatabaseInstanceID: snapshot.DatabaseInstanceID,
				Severity:           "critical",
				Category:           "queries",
				Title:              "Critical long running query detected",
				Message:            fmt.Sprintf("A query has been executing for more than %.1f minutes on %s.", minutes, session),
				Recommendation:     "Investigate immediately for inefficient queries, locks, or missing indexes.",
				RuleKey:            r.Name(),
				ResourceType:       "session",
				ResourceName:       session,
				CurrentValue:       minutes,
				ThresholdValue:     longRunningQueryCriticalThreshold.Minutes(),
			})
			continue
		}

		findings = append(findings, analysis.Finding{
			DetectedAt:         snapshot.CollectedAt,
			DatabaseInstanceID: snapshot.DatabaseInstanceID,
			Severity:           "warning",
			Category:           "queries",
			Title:              "Long running query detected",
			Message:            fmt.Sprintf("A query has been running for %.1f minutes on %s.", minutes, session),
			Recommendation:     "Review query execution plan, indexes, and query complexity.",
			RuleKey:            r.Name(),
			ResourceType:       "session",
			ResourceName:       session,
			CurrentValue:       minutes,
			ThresholdValue:     longRunningQueryWarningThreshold.Minutes(),
		})
	}

	return findings
}

// describeSession summarizes the database, user, and (if available)
// application name for a session, without including any query text.
func describeSession(a metrics.DatabaseActivity) string {
	description := fmt.Sprintf("database %q (user: %s", a.DatabaseName, a.UserName)

	if a.ApplicationName != "" {
		description += fmt.Sprintf(", application: %s", a.ApplicationName)
	}

	return description + ")"
}
