package diagnosis

import (
	"context"
	"fmt"
	"time"

	"github.com/okumujustine/postgresome/backend/internal/analysis/config"
	"github.com/okumujustine/postgresome/backend/internal/storage/repository"
)

func (s *Service) resolveStaleFindingsWithVerification(ctx context.Context, databaseInstanceID string, detectedAt time.Time, tableSnapshot repository.TableStatsSnapshot, querySnapshot repository.QueryStatsSnapshot) error {
	cutoff := detectedAt.Add(-FindingResolutionGracePeriod)
	staleFindings, err := repository.ListStaleOpenFindings(ctx, s.pool, databaseInstanceID, cutoff)
	if err != nil {
		return err
	}

	tableRows := make(map[string]repository.TableStatRow, len(tableSnapshot.Tables))
	for _, table := range tableSnapshot.Tables {
		tableRows[fmt.Sprintf("%s.%s", table.SchemaName, table.TableName)] = table
	}

	queryRows := make(map[string]repository.QueryStatRow, len(querySnapshot.Queries))
	for _, query := range querySnapshot.Queries {
		queryRows[query.QueryID] = query
	}

	for _, finding := range staleFindings {
		verificationStatus, verificationSummary, verifiedFixedAt, err := s.evaluateResolvedFinding(ctx, detectedAt, finding, tableRows, queryRows)
		if err != nil {
			return err
		}
		if err := repository.ResolveFinding(ctx, s.pool, finding.ID, detectedAt, verificationStatus, verificationSummary, verifiedFixedAt); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) evaluateResolvedFinding(
	ctx context.Context,
	detectedAt time.Time,
	finding repository.RecentFinding,
	tableRows map[string]repository.TableStatRow,
	queryRows map[string]repository.QueryStatRow,
) (string, string, *time.Time, error) {
	switch finding.RuleKey {
	case config.RuleKeyHighDeadRows, config.RuleKeyAutovacuumLag:
		table, ok := tableRows[finding.ResourceName]
		if !ok {
			return "pending", "The issue stopped recurring, but no fresh table evidence was available to verify the fix.", nil, nil
		}

		schemaName, tableName := splitTableResource(finding.ResourceName)
		history, err := repository.GetTableStatHistoricalContext(ctx, s.pool, finding.DatabaseInstanceID, schemaName, tableName, detectedAt, premiumHistoryLookback, finding.ThresholdValue)
		if err != nil {
			return "", "", nil, err
		}

		improvedDeadRows := float64(table.DeadRows) < finding.CurrentValue && float64(table.DeadRows) < finding.ThresholdValue
		improvedVacuum := newerMaintenanceObserved(table, finding.LastSeenAt)
		improvedSeqScans := history.Current != nil && history.Previous != nil && history.Current.SequentialScans <= history.Previous.SequentialScans
		if improvedDeadRows && (improvedVacuum || improvedSeqScans) {
			verifiedAt := detectedAt
			return "verified_fixed", "The issue stopped recurring and the supporting table evidence improved enough to verify the fix.", &verifiedAt, nil
		}

		return "pending", "The issue stopped recurring, but the supporting table evidence is not yet strong enough to verify the fix.", nil, nil

	case config.RuleKeySlowQuery:
		query, ok := queryRows[finding.ResourceName]
		if !ok {
			return "pending", "The issue stopped recurring, but no fresh query evidence was available to verify the fix.", nil, nil
		}

		history, err := repository.GetQueryStatHistoricalContext(ctx, s.pool, finding.DatabaseInstanceID, finding.ResourceName, detectedAt, premiumHistoryLookback, finding.ThresholdValue)
		if err != nil {
			return "", "", nil, err
		}

		improvedLatency := query.MeanExecTimeMs < finding.CurrentValue && query.MeanExecTimeMs < finding.ThresholdValue
		improvedReads := history.Current != nil && history.Previous != nil && history.Current.SharedBlocksRead <= history.Previous.SharedBlocksRead
		if improvedLatency && improvedReads {
			verifiedAt := detectedAt
			return "verified_fixed", "The issue stopped recurring and the query's latency and disk-read pressure improved enough to verify the fix.", &verifiedAt, nil
		}

		return "pending", "The issue stopped recurring, but the query evidence is not yet strong enough to verify the fix.", nil, nil
	}

	return "pending", "The issue stopped recurring, but fix verification is not available for this diagnosis yet.", nil, nil
}

func newerMaintenanceObserved(table repository.TableStatRow, seenAt time.Time) bool {
	if table.LastAutoVacuumAt != nil && table.LastAutoVacuumAt.After(seenAt) {
		return true
	}
	return table.LastVacuumAt != nil && table.LastVacuumAt.After(seenAt)
}
