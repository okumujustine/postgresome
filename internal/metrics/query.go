package metrics

import "time"

type QueryStats struct {
	QueryID      string
	DatabaseName string
	UserName     string
	Query        string

	Calls int64

	TotalExecutionTimeMs float64
	MeanExecutionTimeMs  float64
	MinExecutionTimeMs   float64
	MaxExecutionTimeMs   float64

	RowsReturned int64

	SharedBlocksHit     int64
	SharedBlocksRead    int64
	SharedBlocksDirtied int64
	SharedBlocksWritten int64

	TempBlocksRead    int64
	TempBlocksWritten int64
}

type QueryStatsSnapshot struct {
	CollectedAt time.Time

	DatabaseInstanceID string

	Queries []QueryStats
}
