package queries

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/okumujustine/postgresome/internal/metrics"
)

const maxQueryStats = 20

type QueryStatsCollector struct {
	pool *pgxpool.Pool
}

func NewQueryStatsCollector(pool *pgxpool.Pool) *QueryStatsCollector {
	return &QueryStatsCollector{
		pool: pool,
	}
}

func (c *QueryStatsCollector) GetQueryStats(ctx context.Context) (*metrics.QueryStatsSnapshot, error) {
	query := `
		SELECT
			s.queryid::text,
			s.query,
			d.datname,
			u.usename,
			s.calls,
			s.total_exec_time,
			s.mean_exec_time,
			s.min_exec_time,
			s.max_exec_time,
			s.rows,
			s.shared_blks_hit,
			s.shared_blks_read,
			s.shared_blks_dirtied,
			s.shared_blks_written,
			s.temp_blks_read,
			s.temp_blks_written
		FROM pg_stat_statements s
		LEFT JOIN pg_database d ON s.dbid = d.oid
		LEFT JOIN pg_user u ON s.userid = u.usesysid
		ORDER BY s.total_exec_time DESC
		LIMIT $1;
	`

	rows, err := c.pool.Query(ctx, query, maxQueryStats)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	queryStats := make([]metrics.QueryStats, 0)

	for rows.Next() {
		var (
			stat         metrics.QueryStats
			databaseName *string
			userName     *string
		)

		err := rows.Scan(
			&stat.QueryID,
			&stat.Query,
			&databaseName,
			&userName,
			&stat.Calls,
			&stat.TotalExecutionTimeMs,
			&stat.MeanExecutionTimeMs,
			&stat.MinExecutionTimeMs,
			&stat.MaxExecutionTimeMs,
			&stat.RowsReturned,
			&stat.SharedBlocksHit,
			&stat.SharedBlocksRead,
			&stat.SharedBlocksDirtied,
			&stat.SharedBlocksWritten,
			&stat.TempBlocksRead,
			&stat.TempBlocksWritten,
		)
		if err != nil {
			return nil, err
		}

		stat.DatabaseName = stringValue(databaseName)
		stat.UserName = stringValue(userName)

		queryStats = append(queryStats, stat)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &metrics.QueryStatsSnapshot{
		CollectedAt: time.Now(),
		Queries:     queryStats,
	}, nil
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}
