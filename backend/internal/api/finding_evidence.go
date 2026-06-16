package api

import (
	"context"
	"strings"
	"time"

	"github.com/okumujustine/postgresome/backend/internal/analysis/config"
	"github.com/okumujustine/postgresome/backend/internal/storage/repository"
)

const (
	findingEvidenceLookback   = 7 * 24 * time.Hour
	findingEvidencePointLimit = 8
)

func (s *Server) buildFindingEvidence(ctx context.Context, finding repository.RecentFinding) (*findingHistoricalContextDTO, []findingEvidencePointDTO, error) {
	switch finding.RuleKey {
	case config.RuleKeyHighDeadRows, config.RuleKeyAutovacuumLag:
		schemaName, tableName := splitTableFindingResource(finding.ResourceName)
		history, err := repository.GetTableStatHistoricalContext(
			ctx,
			s.pool,
			finding.DatabaseInstanceID,
			schemaName,
			tableName,
			finding.LastSeenAt,
			findingEvidenceLookback,
			finding.ThresholdValue,
		)
		if err != nil {
			return nil, nil, err
		}
		if history.Current == nil {
			return nil, nil, nil
		}

		points, err := repository.ListTableStatHistory(
			ctx,
			s.pool,
			finding.DatabaseInstanceID,
			schemaName,
			tableName,
			finding.LastSeenAt.Add(-findingEvidenceLookback),
			finding.LastSeenAt,
			findingEvidencePointLimit,
		)
		if err != nil {
			return nil, nil, err
		}

		baselineValue := finding.BaselineValue
		if baselineValue == 0 {
			baselineValue = history.BaselineDeadRows
		}

		evidencePoints := make([]findingEvidencePointDTO, 0, len(points)*2)
		for _, point := range points {
			evidencePoints = append(evidencePoints,
				findingEvidencePointDTO{Time: point.CollectedAt, Series: "dead_rows", Value: float64(point.DeadRows)},
				findingEvidencePointDTO{Time: point.CollectedAt, Series: "sequential_scans", Value: float64(point.SequentialScans)},
			)
		}

		return &findingHistoricalContextDTO{
			CurrentValue:  float64(history.Current.DeadRows),
			PreviousValue: previousTableValue(history.Previous),
			BaselineValue: baselineValue,
			ChangePercent: findingEvidenceChangePercent(float64(history.Current.DeadRows), baselineValue),
			TrendWindow:   "7d",
			BaselineLabel: firstNonEmpty(finding.BaselineLabel, "Baseline dead tuples"),
		}, evidencePoints, nil

	case config.RuleKeySlowQuery:
		history, err := repository.GetQueryStatHistoricalContext(
			ctx,
			s.pool,
			finding.DatabaseInstanceID,
			finding.ResourceName,
			finding.LastSeenAt,
			findingEvidenceLookback,
			finding.ThresholdValue,
		)
		if err != nil {
			return nil, nil, err
		}
		if history.Current == nil {
			return nil, nil, nil
		}

		points, err := repository.ListQueryStatHistory(
			ctx,
			s.pool,
			finding.DatabaseInstanceID,
			finding.ResourceName,
			finding.LastSeenAt.Add(-findingEvidenceLookback),
			finding.LastSeenAt,
			findingEvidencePointLimit,
		)
		if err != nil {
			return nil, nil, err
		}

		baselineValue := finding.BaselineValue
		if baselineValue == 0 {
			baselineValue = history.BaselineMeanExecMs
		}

		evidencePoints := make([]findingEvidencePointDTO, 0, len(points)*2)
		for _, point := range points {
			evidencePoints = append(evidencePoints,
				findingEvidencePointDTO{Time: point.CollectedAt, Series: "mean_exec_time_ms", Value: point.MeanExecTimeMs},
				findingEvidencePointDTO{Time: point.CollectedAt, Series: "shared_blocks_read", Value: float64(point.SharedBlocksRead)},
			)
		}

		return &findingHistoricalContextDTO{
			CurrentValue:  history.Current.MeanExecTimeMs,
			PreviousValue: previousQueryValue(history.Previous),
			BaselineValue: baselineValue,
			ChangePercent: findingEvidenceChangePercent(history.Current.MeanExecTimeMs, baselineValue),
			TrendWindow:   "7d",
			BaselineLabel: firstNonEmpty(finding.BaselineLabel, "Baseline mean latency (ms)"),
		}, evidencePoints, nil
	}

	return nil, nil, nil
}

func splitTableFindingResource(resource string) (string, string) {
	parts := strings.SplitN(resource, ".", 2)
	if len(parts) != 2 {
		return "public", resource
	}
	return parts[0], parts[1]
}

func previousTableValue(previous *repository.TableStatHistoryPoint) float64 {
	if previous == nil {
		return 0
	}
	return float64(previous.DeadRows)
}

func previousQueryValue(previous *repository.QueryStatHistoryPoint) float64 {
	if previous == nil {
		return 0
	}
	return previous.MeanExecTimeMs
}

func findingEvidenceChangePercent(current, baseline float64) float64 {
	if baseline == 0 {
		if current == 0 {
			return 0
		}
		return 100
	}
	return ((current - baseline) / baseline) * 100
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
