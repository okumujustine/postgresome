package metrics

import "time"

type DatabaseMetricSample struct {
	CollectedAt time.Time

	DatabaseInstanceID string

	ActiveConnections int

	TransactionCommits   int64
	TransactionRollbacks int64

	BlocksReadFromDisk int64
	BlocksHitInCache   int64

	RowsScanned int64
	RowsFetched int64

	RowsInserted int64
	RowsUpdated  int64
	RowsDeleted  int64
}
