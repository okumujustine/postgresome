package metrics

import "time"

// CalculateDatabaseDelta turns two cumulative DatabaseStats snapshots into a
// DatabaseMetricDelta describing what changed between them, along with
// derived rate metrics.
func CalculateDatabaseDelta(previous *DatabaseStats, current *DatabaseStats, collectedAt time.Time, databaseInstanceID string) DatabaseMetricDelta {
	transactionCommitsDelta := current.TransactionCommits - previous.TransactionCommits
	transactionRollbacksDelta := current.TransactionRollbacks - previous.TransactionRollbacks

	blocksReadFromDiskDelta := current.BlocksReadFromDisk - previous.BlocksReadFromDisk
	blocksHitInCacheDelta := current.BlocksHitInCache - previous.BlocksHitInCache

	rowsScannedDelta := current.RowsScanned - previous.RowsScanned
	rowsFetchedDelta := current.RowsFetched - previous.RowsFetched

	rowsInsertedDelta := current.RowsInserted - previous.RowsInserted
	rowsUpdatedDelta := current.RowsUpdated - previous.RowsUpdated
	rowsDeletedDelta := current.RowsDeleted - previous.RowsDeleted

	return DatabaseMetricDelta{
		CollectedAt:        collectedAt,
		DatabaseInstanceID: databaseInstanceID,

		TransactionCommitsDelta:   transactionCommitsDelta,
		TransactionRollbacksDelta: transactionRollbacksDelta,

		BlocksReadFromDiskDelta: blocksReadFromDiskDelta,
		BlocksHitInCacheDelta:   blocksHitInCacheDelta,

		RowsScannedDelta: rowsScannedDelta,
		RowsFetchedDelta: rowsFetchedDelta,

		RowsInsertedDelta: rowsInsertedDelta,
		RowsUpdatedDelta:  rowsUpdatedDelta,
		RowsDeletedDelta:  rowsDeletedDelta,

		CacheHitRatio: safeRatio(float64(blocksHitInCacheDelta), float64(blocksHitInCacheDelta+blocksReadFromDiskDelta)),
		RollbackRate:  safeRatio(float64(transactionRollbacksDelta), float64(transactionCommitsDelta+transactionRollbacksDelta)),
	}
}

func safeRatio(numerator, denominator float64) float64 {
	if denominator == 0 {
		return 0
	}

	return numerator / denominator
}
