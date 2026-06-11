package metrics

import "time"

type TableStats struct {
	SchemaName string
	TableName  string

	LiveRows int64
	DeadRows int64

	SequentialScans    int64
	SequentialRowsRead int64

	IndexScans       int64
	IndexRowsFetched int64

	RowsInserted int64
	RowsUpdated  int64
	RowsDeleted  int64

	LastVacuumAt      *time.Time
	LastAutoVacuumAt  *time.Time
	LastAnalyzeAt     *time.Time
	LastAutoAnalyzeAt *time.Time

	VacuumCount      int64
	AutoVacuumCount  int64
	AnalyzeCount     int64
	AutoAnalyzeCount int64
}

type TableStatsSnapshot struct {
	CollectedAt time.Time

	DatabaseInstanceID string

	Tables []TableStats
}
