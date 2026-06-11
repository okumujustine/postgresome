package metrics

import "time"

type DatabaseMetricDelta struct {
	CollectedAt        time.Time
	DatabaseInstanceID string

	TransactionCommitsDelta   int64
	TransactionRollbacksDelta int64

	BlocksReadFromDiskDelta int64
	BlocksHitInCacheDelta   int64

	RowsScannedDelta int64
	RowsFetchedDelta int64

	RowsInsertedDelta int64
	RowsUpdatedDelta  int64
	RowsDeletedDelta  int64

	CacheHitRatio float64
	RollbackRate  float64
}
