package diagnosis

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/okumujustine/postgresome/backend/internal/analysis"
	"github.com/okumujustine/postgresome/backend/internal/analysis/config"
	"github.com/okumujustine/postgresome/backend/internal/storage/repository"
)

const premiumHistoryLookback = 7 * 24 * time.Hour

func (s *Service) enrichHistoricalDiagnosis(ctx context.Context, detectedAt time.Time, databaseInstanceID string, tableSnapshot repository.TableStatsSnapshot, querySnapshot repository.QueryStatsSnapshot, findings []analysis.Finding) error {
	tableRows := make(map[string]repository.TableStatRow, len(tableSnapshot.Tables))
	for _, table := range tableSnapshot.Tables {
		key := fmt.Sprintf("%s.%s", table.SchemaName, table.TableName)
		tableRows[key] = table
	}

	queryRows := make(map[string]repository.QueryStatRow, len(querySnapshot.Queries))
	for _, query := range querySnapshot.Queries {
		queryRows[query.QueryID] = query
	}

	for i := range findings {
		switch findings[i].RuleKey {
		case config.RuleKeyHighDeadRows, config.RuleKeyAutovacuumLag:
			table, ok := tableRows[findings[i].ResourceName]
			if !ok {
				continue
			}
			schemaName, tableName := splitTableResource(findings[i].ResourceName)
			history, err := repository.GetTableStatHistoricalContext(ctx, s.pool, databaseInstanceID, schemaName, tableName, detectedAt, premiumHistoryLookback, findings[i].ThresholdValue)
			if err != nil {
				return err
			}
			enrichTableFinding(&findings[i], table, history, detectedAt)

		case config.RuleKeySlowQuery:
			query, ok := queryRows[findings[i].ResourceName]
			if !ok {
				continue
			}
			history, err := repository.GetQueryStatHistoricalContext(ctx, s.pool, databaseInstanceID, findings[i].ResourceName, detectedAt, premiumHistoryLookback, findings[i].ThresholdValue)
			if err != nil {
				return err
			}
			enrichSlowQueryFinding(&findings[i], query, history)
		}
	}

	return nil
}

func enrichTableFinding(finding *analysis.Finding, table repository.TableStatRow, history repository.TableStatHistoricalContext, now time.Time) {
	if history.Current == nil {
		return
	}

	baselineDeadRows := history.BaselineDeadRows
	if baselineDeadRows == 0 && history.Previous != nil {
		baselineDeadRows = float64(history.Previous.DeadRows)
	}
	deadRowsChange := percentChange(float64(history.Current.DeadRows), baselineDeadRows)

	baselineSeqScans := history.BaselineSeqScans
	if baselineSeqScans == 0 && history.Previous != nil {
		baselineSeqScans = float64(history.Previous.SequentialScans)
	}
	seqScanChange := percentChange(float64(history.Current.SequentialScans), baselineSeqScans)

	lastMaintenance := maintenanceAge(now, table.LastAutoVacuumAt, table.LastVacuumAt)
	firstAbnormal := ""
	if history.FirstAbnormalAt != nil {
		firstAbnormal = fmt.Sprintf(" The abnormal pattern first appeared around %s.", history.FirstAbnormalAt.Format("2006-01-02 15:04 MST"))
	}

	switch finding.RuleKey {
	case config.RuleKeyHighDeadRows:
		finding.Title = "Table bloat risk detected"
		finding.ProblemSummary = fmt.Sprintf("Table %s is accumulating dead tuples faster than cleanup, which puts it at risk of bloat.", finding.ResourceName)
		finding.EvidenceSummary = fmt.Sprintf(
			"Dead tuples on %s increased from a %.0f baseline to %d now (%s). Sequential scans are %d now (%s vs baseline).%s",
			finding.ResourceName,
			baselineDeadRows,
			history.Current.DeadRows,
			formatPercent(deadRowsChange),
			history.Current.SequentialScans,
			formatPercent(seqScanChange),
			firstAbnormal,
		)
		finding.ImpactSummary = "Queries touching this table may scan more heap pages, become slower over time, and make maintenance work more expensive."
		finding.SuggestedAction = fmt.Sprintf("Review autovacuum settings for %s and plan manual vacuum work if dead tuples keep climbing or reads remain scan-heavy.", finding.ResourceName)
		finding.BaselineValue = baselineDeadRows
		finding.BaselineLabel = "Baseline dead tuples"
		finding.ChangeSummary = fmt.Sprintf("Dead tuples are %s versus the recent baseline; %s.", formatPercent(deadRowsChange), lastMaintenance)
		finding.VerificationHint = fmt.Sprintf("After remediation, confirm dead tuples on %s trend back toward the %.0f baseline and sequential scans stop rising.", finding.ResourceName, baselineDeadRows)
		finding.ConfidenceScore = tableFindingConfidence(deadRowsChange, seqScanChange, table.LastAutoVacuumAt, table.LastVacuumAt, now)
		finding.ConfidenceLabel = confidenceLabel(finding.ConfidenceScore)
		assignImprovingLifecycle(finding, history.Previous != nil && history.Current.DeadRows < history.Previous.DeadRows, "Dead tuples are moving down, but the table is still above the diagnosis threshold.")
		finding.Evidence = buildTableEvidence(finding, table, history, baselineDeadRows, baselineSeqScans, deadRowsChange, seqScanChange, now)

	case config.RuleKeyAutovacuumLag:
		deadRowRatio := 0.0
		if totalRows := table.LiveRows + table.DeadRows; totalRows > 0 {
			deadRowRatio = (float64(table.DeadRows) / float64(totalRows)) * 100
		}
		finding.Title = "Vacuum lag detected"
		finding.ProblemSummary = fmt.Sprintf("Vacuum maintenance is not keeping up on %s, allowing cleanup debt to accumulate.", finding.ResourceName)
		finding.EvidenceSummary = fmt.Sprintf(
			"%s now has %d dead tuples (%.1f%% dead-row ratio) versus a %.0f baseline. Sequential scans are %d now (%s vs baseline), and %s.%s",
			finding.ResourceName,
			table.DeadRows,
			deadRowRatio,
			baselineDeadRows,
			history.Current.SequentialScans,
			formatPercent(seqScanChange),
			lastMaintenance,
			firstAbnormal,
		)
		finding.ImpactSummary = "When vacuum falls behind, bloat compounds, visibility maps stay stale, and both reads and future cleanup cycles get more expensive."
		finding.SuggestedAction = fmt.Sprintf("Inspect autovacuum thresholds, scale factors, and worker availability for %s; schedule manual maintenance if the table needs immediate relief.", finding.ResourceName)
		finding.BaselineValue = baselineDeadRows
		finding.BaselineLabel = "Baseline dead tuples"
		finding.ChangeSummary = fmt.Sprintf("Dead tuples are %s versus the recent baseline, while vacuum recency suggests cleanup is lagging.", formatPercent(deadRowsChange))
		finding.VerificationHint = fmt.Sprintf("Confirm a fresh vacuum/autovacuum runs on %s and dead tuples start falling toward the %.0f baseline.", finding.ResourceName, baselineDeadRows)
		finding.ConfidenceScore = tableFindingConfidence(deadRowsChange, seqScanChange, table.LastAutoVacuumAt, table.LastVacuumAt, now)
		finding.ConfidenceLabel = confidenceLabel(finding.ConfidenceScore)
		assignImprovingLifecycle(finding, history.Previous != nil && history.Current.DeadRows < history.Previous.DeadRows, "Vacuum pressure is easing, but the table still needs more cleanup before the issue is considered fixed.")
		finding.Evidence = buildTableEvidence(finding, table, history, baselineDeadRows, baselineSeqScans, deadRowsChange, seqScanChange, now)
	}
}

func enrichSlowQueryFinding(finding *analysis.Finding, query repository.QueryStatRow, history repository.QueryStatHistoricalContext) {
	if history.Current == nil {
		return
	}

	baselineExecMs := history.BaselineMeanExecMs
	if baselineExecMs == 0 && history.Previous != nil {
		baselineExecMs = history.Previous.MeanExecTimeMs
	}
	baselineCalls := history.BaselineCalls
	if baselineCalls == 0 && history.Previous != nil {
		baselineCalls = float64(history.Previous.Calls)
	}
	baselineBlocksRead := history.BaselineBlocksRead
	if baselineBlocksRead == 0 && history.Previous != nil {
		baselineBlocksRead = float64(history.Previous.SharedBlocksRead)
	}

	latencyChange := percentChange(history.Current.MeanExecTimeMs, baselineExecMs)
	callChange := percentChange(float64(history.Current.Calls), baselineCalls)
	diskReadChange := percentChange(float64(history.Current.SharedBlocksRead), baselineBlocksRead)

	firstAbnormal := ""
	if history.FirstAbnormalAt != nil {
		firstAbnormal = fmt.Sprintf(" The query first crossed the threshold around %s.", history.FirstAbnormalAt.Format("2006-01-02 15:04 MST"))
	}

	finding.Title = "Slow query regression detected"
	finding.ProblemSummary = fmt.Sprintf("Query %s is materially slower than its recent baseline.", finding.ResourceName)
	finding.EvidenceSummary = fmt.Sprintf(
		"Mean execution time rose from %.1f ms to %.1f ms (%s). Calls changed by %s and shared block reads changed by %s.%s",
		baselineExecMs,
		history.Current.MeanExecTimeMs,
		formatPercent(latencyChange),
		formatPercent(callChange),
		formatPercent(diskReadChange),
		firstAbnormal,
	)
	finding.ImpactSummary = "If this query remains on the hot path, it can increase request latency, amplify disk pressure, and crowd out healthier work."
	finding.SuggestedAction = "Inspect the execution plan, check whether index usage regressed, and compare the plan with an earlier faster version of the same query."
	finding.BaselineValue = baselineExecMs
	finding.BaselineLabel = "Baseline mean latency (ms)"
	finding.ChangeSummary = fmt.Sprintf("Latency is %s versus baseline, with call volume %s and disk reads %s.", formatPercent(latencyChange), formatPercent(callChange), formatPercent(diskReadChange))
	finding.VerificationHint = fmt.Sprintf("After tuning, confirm mean execution time for query %s trends back toward %.1f ms and shared block reads stop climbing.", finding.ResourceName, baselineExecMs)
	finding.ConfidenceScore = queryFindingConfidence(latencyChange, diskReadChange, query.Calls)
	finding.ConfidenceLabel = confidenceLabel(finding.ConfidenceScore)
	assignImprovingLifecycle(
		finding,
		history.Previous != nil && history.Current.MeanExecTimeMs < history.Previous.MeanExecTimeMs && history.Current.SharedBlocksRead <= history.Previous.SharedBlocksRead,
		"Latency and disk-read pressure are moving in the right direction, but the query is still above the diagnosis threshold.",
	)
	finding.Evidence = buildSlowQueryEvidence(finding, query, history, baselineExecMs, baselineCalls, baselineBlocksRead, latencyChange, callChange, diskReadChange)
}

func splitTableResource(resource string) (string, string) {
	parts := strings.SplitN(resource, ".", 2)
	if len(parts) != 2 {
		return "public", resource
	}
	return parts[0], parts[1]
}

func percentChange(current, baseline float64) float64 {
	if baseline == 0 {
		if current == 0 {
			return 0
		}
		return 100
	}
	return ((current - baseline) / baseline) * 100
}

func formatPercent(value float64) string {
	switch {
	case math.IsNaN(value), math.IsInf(value, 0):
		return "0%"
	case value >= 0:
		return fmt.Sprintf("+%.0f%%", value)
	default:
		return fmt.Sprintf("%.0f%%", value)
	}
}

func maintenanceAge(now time.Time, lastAutoVacuumAt, lastVacuumAt *time.Time) string {
	last := lastAutoVacuumAt
	if last == nil {
		last = lastVacuumAt
	}
	if last == nil {
		return "no recent vacuum has been recorded"
	}

	age := now.Sub(*last)
	if age >= 48*time.Hour {
		return fmt.Sprintf("the last vacuum activity was %.1f days ago", age.Hours()/24)
	}

	return fmt.Sprintf("the last vacuum activity was %.1f hours ago", age.Hours())
}

func tableFindingConfidence(deadRowsChange, seqScanChange float64, lastAutoVacuumAt, lastVacuumAt *time.Time, now time.Time) float64 {
	score := 0.62
	if deadRowsChange >= 100 {
		score += 0.12
	}
	if seqScanChange >= 25 {
		score += 0.1
	}
	last := lastAutoVacuumAt
	if last == nil {
		last = lastVacuumAt
	}
	if last == nil || now.Sub(*last) >= 48*time.Hour {
		score += 0.1
	}
	if score > 0.95 {
		return 0.95
	}
	return score
}

func queryFindingConfidence(latencyChange, diskReadChange float64, calls int64) float64 {
	score := 0.66
	if latencyChange >= 100 {
		score += 0.12
	}
	if diskReadChange >= 25 {
		score += 0.08
	}
	if calls >= 25 {
		score += 0.06
	}
	if score > 0.94 {
		return 0.94
	}
	return score
}

func confidenceLabel(score float64) string {
	switch {
	case score >= 0.85:
		return "high"
	case score >= 0.65:
		return "medium"
	default:
		return "low"
	}
}

func assignImprovingLifecycle(finding *analysis.Finding, improving bool, summary string) {
	if improving {
		finding.VerificationStatus = "improving"
		finding.VerificationSummary = summary
		return
	}

	finding.VerificationStatus = "pending"
	finding.VerificationSummary = ""
}
