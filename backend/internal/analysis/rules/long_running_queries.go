package rules

import (
	"fmt"
	"time"

	"github.com/okumujustine/postgresome/backend/internal/analysis"
	"github.com/okumujustine/postgresome/backend/internal/analysis/config"
	"github.com/okumujustine/postgresome/backend/internal/metrics"
)

// LongRunningQueryRule detects active queries that have been running for an
// unusually long time, which can indicate inefficient queries, missing
// indexes, large table scans, or stuck application requests.
type LongRunningQueryRule struct {
	Config config.RuleConfig
}

func (r LongRunningQueryRule) Name() string {
	return config.RuleKeyLongRunningQuery
}

func (r LongRunningQueryRule) Analyze(snapshot metrics.DatabaseActivitySnapshot) []analysis.Finding {
	if !r.Config.Enabled {
		return nil
	}

	warningThreshold := time.Duration(r.Config.Thresholds["warning_minutes"] * float64(time.Minute))
	criticalThreshold := time.Duration(r.Config.Thresholds["critical_minutes"] * float64(time.Minute))

	findings := make([]analysis.Finding, 0)

	for _, a := range snapshot.Activities {
		if a.State != "active" || a.QueryStartedAt == nil {
			continue
		}

		duration := snapshot.CollectedAt.Sub(*a.QueryStartedAt)
		if duration < warningThreshold {
			continue
		}

		minutes := duration.Minutes()
		session := describeSession(a)

		if duration >= criticalThreshold {
			findings = append(findings, analysis.Finding{
				DetectedAt:         snapshot.CollectedAt,
				DatabaseInstanceID: snapshot.DatabaseInstanceID,
				Severity:           "critical",
				Category:           "queries",
				Title:              r.Config.Title,
				Message:            r.Config.Description,
				Recommendation:     r.Config.Recommendation,
				RuleKey:            r.Name(),
				ResourceType:       "session",
				ResourceName:       session,
				CurrentValue:       minutes,
				ThresholdValue:     criticalThreshold.Minutes(),
			})
			continue
		}

		findings = append(findings, analysis.Finding{
			DetectedAt:         snapshot.CollectedAt,
			DatabaseInstanceID: snapshot.DatabaseInstanceID,
			Severity:           r.Config.Severity,
			Category:           "queries",
			Title:              r.Config.Title,
			Message:            r.Config.Description,
			Recommendation:     r.Config.Recommendation,
			RuleKey:            r.Name(),
			ResourceType:       "session",
			ResourceName:       session,
			CurrentValue:       minutes,
			ThresholdValue:     warningThreshold.Minutes(),
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
