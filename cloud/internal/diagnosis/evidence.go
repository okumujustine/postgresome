package diagnosis

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/okumujustine/postgresome/cloud/internal/analysis"
	"github.com/okumujustine/postgresome/cloud/internal/analysis/config"
	"github.com/okumujustine/postgresome/cloud/internal/storage/repository"
)

func (s *Service) attachFindingEvidence(ctx context.Context, detectedAt time.Time, activitySnapshot repository.ActivityStatsSnapshot, findings []analysis.Finding) error {
	for i := range findings {
		if len(findings[i].Evidence) == 0 {
			switch findings[i].RuleKey {
			case config.RuleKeyBlockedQuery:
				findings[i].Evidence = buildBlockedQueryEvidence(findings[i], activitySnapshot)
			}
		}

		enrichFindingDefaults(&findings[i], detectedAt)
	}

	for i := range findings {
		if findings[i].ID == "" {
			continue
		}
		if err := repository.ReplaceFindingEvidence(ctx, s.pool, findings[i].ID, findings[i].Evidence); err != nil {
			return err
		}
	}

	return nil
}

func enrichFindingDefaults(finding *analysis.Finding, detectedAt time.Time) {
	if finding.DetectedAt.IsZero() {
		finding.DetectedAt = detectedAt
	}

	if finding.ProblemSummary == "" {
		finding.ProblemSummary = defaultProblemSummary(*finding)
	}
	if finding.EvidenceSummary == "" {
		finding.EvidenceSummary = defaultEvidenceSummary(*finding)
	}
	if finding.ImpactSummary == "" {
		finding.ImpactSummary = defaultImpactSummary(*finding)
	}
	if finding.SuggestedAction == "" {
		finding.SuggestedAction = finding.Recommendation
	}
	if finding.ChangeSummary == "" {
		finding.ChangeSummary = fmt.Sprintf("Observed %.1f against a threshold of %.1f.", finding.CurrentValue, finding.ThresholdValue)
	}
	if finding.ConfidenceScore == 0 {
		finding.ConfidenceScore = 0.68
	}
	if finding.ConfidenceLabel == "" {
		finding.ConfidenceLabel = confidenceLabel(finding.ConfidenceScore)
	}
	if finding.VerificationHint == "" {
		finding.VerificationHint = "Confirm the supporting evidence trends back toward baseline and the issue stops recurring."
	}
	if finding.VerificationStatus == "" {
		finding.VerificationStatus = "pending"
	}

	if len(finding.Evidence) == 0 {
		finding.Evidence = []analysis.FindingEvidence{{
			ObservedAt:      finding.DetectedAt,
			EvidenceType:    normalizeEvidenceType(finding.ResourceType),
			Role:            "trigger",
			Label:           defaultEvidenceLabel(*finding),
			Summary:         finding.EvidenceSummary,
			MetricKey:       defaultMetricKey(*finding),
			ReferenceID:     finding.ResourceName,
			CurrentValue:    finding.CurrentValue,
			BaselineValue:   finding.BaselineValue,
			ChangePercent:   percentChange(finding.CurrentValue, finding.BaselineValue),
			ConfidenceScore: finding.ConfidenceScore,
			Metadata: map[string]any{
				"rule_key": finding.RuleKey,
				"severity": finding.Severity,
			},
		}}
	}
}

func buildTableEvidence(
	finding *analysis.Finding,
	table repository.TableStatRow,
	history repository.TableStatHistoricalContext,
	baselineDeadRows float64,
	baselineSeqScans float64,
	deadRowsChange float64,
	seqScanChange float64,
	now time.Time,
) []analysis.FindingEvidence {
	lastMaintenanceHours := 0.0
	lastMaintenanceLabel := "No recent vacuum activity recorded."
	if last := latestMaintenanceTime(table.LastAutoVacuumAt, table.LastVacuumAt); last != nil {
		lastMaintenanceHours = now.Sub(*last).Hours()
		lastMaintenanceLabel = maintenanceAge(now, table.LastAutoVacuumAt, table.LastVacuumAt)
	}

	return []analysis.FindingEvidence{
		{
			ObservedAt:      finding.DetectedAt,
			EvidenceType:    "table",
			Role:            "trigger",
			Label:           "Dead tuples",
			Summary:         fmt.Sprintf("%s is carrying %d dead tuples compared with a baseline of %.0f.", finding.ResourceName, table.DeadRows, baselineDeadRows),
			MetricKey:       "dead_rows",
			ReferenceID:     finding.ResourceName,
			CurrentValue:    float64(table.DeadRows),
			BaselineValue:   baselineDeadRows,
			ChangePercent:   deadRowsChange,
			ConfidenceScore: finding.ConfidenceScore,
			Metadata: map[string]any{
				"schema_table": finding.ResourceName,
				"live_rows":    table.LiveRows,
			},
		},
		{
			ObservedAt:      finding.DetectedAt,
			EvidenceType:    "table",
			Role:            "supporting",
			Label:           "Sequential scans",
			Summary:         fmt.Sprintf("%s is now at %d sequential scans, which is %s versus baseline.", finding.ResourceName, history.Current.SequentialScans, formatPercent(seqScanChange)),
			MetricKey:       "sequential_scans",
			ReferenceID:     finding.ResourceName,
			CurrentValue:    float64(history.Current.SequentialScans),
			BaselineValue:   baselineSeqScans,
			ChangePercent:   seqScanChange,
			ConfidenceScore: finding.ConfidenceScore,
			Metadata: map[string]any{
				"schema_table": finding.ResourceName,
			},
		},
		{
			ObservedAt:      finding.DetectedAt,
			EvidenceType:    "maintenance",
			Role:            "supporting",
			Label:           "Vacuum recency",
			Summary:         lastMaintenanceLabel,
			MetricKey:       "vacuum_age_hours",
			ReferenceID:     finding.ResourceName,
			CurrentValue:    lastMaintenanceHours,
			BaselineValue:   0,
			ChangePercent:   0,
			ConfidenceScore: finding.ConfidenceScore,
			Metadata: map[string]any{
				"last_autovacuum_at": timePtrRFC3339(table.LastAutoVacuumAt),
				"last_vacuum_at":     timePtrRFC3339(table.LastVacuumAt),
			},
		},
	}
}

func buildSlowQueryEvidence(
	finding *analysis.Finding,
	query repository.QueryStatRow,
	history repository.QueryStatHistoricalContext,
	baselineExecMs float64,
	baselineCalls float64,
	baselineBlocksRead float64,
	latencyChange float64,
	callChange float64,
	diskReadChange float64,
) []analysis.FindingEvidence {
	return []analysis.FindingEvidence{
		{
			ObservedAt:      finding.DetectedAt,
			EvidenceType:    "query",
			Role:            "trigger",
			Label:           "Mean execution time",
			Summary:         fmt.Sprintf("Query %s is now averaging %.1f ms, up from a baseline of %.1f ms.", finding.ResourceName, history.Current.MeanExecTimeMs, baselineExecMs),
			MetricKey:       "mean_exec_time_ms",
			ReferenceID:     finding.ResourceName,
			CurrentValue:    history.Current.MeanExecTimeMs,
			BaselineValue:   baselineExecMs,
			ChangePercent:   latencyChange,
			ConfidenceScore: finding.ConfidenceScore,
			Metadata: map[string]any{
				"query_preview": previewEvidenceQuery(query.Query),
				"user_name":     query.UserName,
			},
		},
		{
			ObservedAt:      finding.DetectedAt,
			EvidenceType:    "query",
			Role:            "supporting",
			Label:           "Call volume",
			Summary:         fmt.Sprintf("Call volume moved to %d executions in the latest sample, which is %s versus baseline.", history.Current.Calls, formatPercent(callChange)),
			MetricKey:       "calls",
			ReferenceID:     finding.ResourceName,
			CurrentValue:    float64(history.Current.Calls),
			BaselineValue:   baselineCalls,
			ChangePercent:   callChange,
			ConfidenceScore: finding.ConfidenceScore,
			Metadata: map[string]any{
				"query_preview": previewEvidenceQuery(query.Query),
			},
		},
		{
			ObservedAt:      finding.DetectedAt,
			EvidenceType:    "query",
			Role:            "supporting",
			Label:           "Shared block reads",
			Summary:         fmt.Sprintf("Disk reads changed to %d shared blocks, which is %s versus baseline.", history.Current.SharedBlocksRead, formatPercent(diskReadChange)),
			MetricKey:       "shared_blocks_read",
			ReferenceID:     finding.ResourceName,
			CurrentValue:    float64(history.Current.SharedBlocksRead),
			BaselineValue:   baselineBlocksRead,
			ChangePercent:   diskReadChange,
			ConfidenceScore: finding.ConfidenceScore,
			Metadata: map[string]any{
				"query_preview": previewEvidenceQuery(query.Query),
			},
		},
	}
}

func buildBlockedQueryEvidence(finding analysis.Finding, activitySnapshot repository.ActivityStatsSnapshot) []analysis.FindingEvidence {
	blockedSessions := make([]repository.ActivityStatRow, 0)
	for _, activity := range activitySnapshot.Activities {
		if activity.WaitEventType == "Lock" {
			blockedSessions = append(blockedSessions, activity)
		}
	}

	if len(blockedSessions) == 0 {
		return nil
	}

	items := []analysis.FindingEvidence{{
		ObservedAt:      finding.DetectedAt,
		EvidenceType:    "activity",
		Role:            "trigger",
		Label:           "Blocked sessions",
		Summary:         fmt.Sprintf("%d sessions are waiting on locks, above the configured threshold of %.0f.", len(blockedSessions), finding.ThresholdValue),
		MetricKey:       "blocked_sessions",
		ReferenceID:     finding.ResourceName,
		CurrentValue:    float64(len(blockedSessions)),
		BaselineValue:   finding.ThresholdValue,
		ChangePercent:   percentChange(float64(len(blockedSessions)), finding.ThresholdValue),
		ConfidenceScore: finding.ConfidenceScore,
		Metadata: map[string]any{
			"wait_event_type": "Lock",
		},
	}}

	firstBlocked := blockedSessions[0]
	items = append(items, analysis.FindingEvidence{
		ObservedAt:      finding.DetectedAt,
		EvidenceType:    "activity",
		Role:            "supporting",
		Label:           "Representative blocked query",
		Summary:         fmt.Sprintf("PID %d (%s) is waiting on %s.", firstBlocked.ProcessID, firstBlocked.UserName, firstNonEmpty(firstBlocked.WaitEvent, "a lock")),
		MetricKey:       "blocked_query_pid",
		ReferenceID:     strconv.Itoa(firstBlocked.ProcessID),
		CurrentValue:    float64(firstBlocked.ProcessID),
		BaselineValue:   0,
		ChangePercent:   0,
		ConfidenceScore: finding.ConfidenceScore,
		Metadata: map[string]any{
			"query_preview": previewEvidenceQuery(firstBlocked.Query),
			"application":   firstBlocked.ApplicationName,
		},
	})

	return items
}

func defaultProblemSummary(finding analysis.Finding) string {
	resource := finding.ResourceName
	if resource == "" {
		resource = "the database"
	}

	switch finding.RuleKey {
	case config.RuleKeyBlockedQuery:
		return fmt.Sprintf("%s is experiencing lock waits that are blocking active work.", resource)
	default:
		return fmt.Sprintf("%s crossed the %s diagnosis threshold.", resource, finding.RuleKey)
	}
}

func defaultEvidenceSummary(finding analysis.Finding) string {
	return fmt.Sprintf(
		"Observed %.1f for %s against a threshold of %.1f.",
		finding.CurrentValue,
		firstNonEmpty(finding.ResourceName, "the observed signal"),
		finding.ThresholdValue,
	)
}

func defaultImpactSummary(finding analysis.Finding) string {
	switch finding.ResourceType {
	case "query":
		return "This behavior can increase request latency and make the workload less predictable."
	case "table":
		return "This behavior can slow reads, increase maintenance cost, and amplify bloat pressure."
	default:
		return "This behavior can degrade database performance if it continues unchecked."
	}
}

func defaultEvidenceLabel(finding analysis.Finding) string {
	switch finding.ResourceType {
	case "query":
		return "Query regression"
	case "table":
		return "Table health signal"
	case "database":
		return "Database signal"
	default:
		return "Observed signal"
	}
}

func defaultMetricKey(finding analysis.Finding) string {
	if finding.BaselineLabel != "" {
		return strings.ToLower(strings.ReplaceAll(finding.BaselineLabel, " ", "_"))
	}
	return strings.ReplaceAll(finding.RuleKey, ".", "_")
}

func normalizeEvidenceType(resourceType string) string {
	if resourceType == "" {
		return "metric"
	}
	return resourceType
}

func latestMaintenanceTime(lastAutoVacuumAt, lastVacuumAt *time.Time) *time.Time {
	if lastAutoVacuumAt != nil {
		return lastAutoVacuumAt
	}
	return lastVacuumAt
}

func timePtrRFC3339(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.Format(time.RFC3339)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func previewEvidenceQuery(query string) string {
	query = strings.Join(strings.Fields(query), " ")
	if len(query) <= 80 {
		return query
	}
	return query[:80] + "..."
}
