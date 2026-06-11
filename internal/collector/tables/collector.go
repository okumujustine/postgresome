package tables

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/okumujustine/postgresome/internal/metrics"
)

type TableStatsCollector struct {
	pool *pgxpool.Pool
}

func NewTableStatsCollector(pool *pgxpool.Pool) *TableStatsCollector {
	return &TableStatsCollector{
		pool: pool,
	}
}

func (c *TableStatsCollector) GetTableStats(ctx context.Context) (*metrics.TableStatsSnapshot, error) {
	query := `
		SELECT
			schemaname,
			relname,
			n_live_tup,
			n_dead_tup,
			seq_scan,
			seq_tup_read,
			idx_scan,
			idx_tup_fetch,
			n_tup_ins,
			n_tup_upd,
			n_tup_del,
			last_vacuum,
			last_autovacuum,
			last_analyze,
			last_autoanalyze,
			vacuum_count,
			autovacuum_count,
			analyze_count,
			autoanalyze_count
		FROM pg_stat_user_tables;
	`

	rows, err := c.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tables := make([]metrics.TableStats, 0)

	for rows.Next() {
		var (
			table             metrics.TableStats
			lastVacuumAt      *time.Time
			lastAutoVacuumAt  *time.Time
			lastAnalyzeAt     *time.Time
			lastAutoAnalyzeAt *time.Time
		)

		err := rows.Scan(
			&table.SchemaName,
			&table.TableName,
			&table.LiveRows,
			&table.DeadRows,
			&table.SequentialScans,
			&table.SequentialRowsRead,
			&table.IndexScans,
			&table.IndexRowsFetched,
			&table.RowsInserted,
			&table.RowsUpdated,
			&table.RowsDeleted,
			&lastVacuumAt,
			&lastAutoVacuumAt,
			&lastAnalyzeAt,
			&lastAutoAnalyzeAt,
			&table.VacuumCount,
			&table.AutoVacuumCount,
			&table.AnalyzeCount,
			&table.AutoAnalyzeCount,
		)
		if err != nil {
			return nil, err
		}

		table.LastVacuumAt = lastVacuumAt
		table.LastAutoVacuumAt = lastAutoVacuumAt
		table.LastAnalyzeAt = lastAnalyzeAt
		table.LastAutoAnalyzeAt = lastAutoAnalyzeAt

		tables = append(tables, table)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &metrics.TableStatsSnapshot{
		CollectedAt: time.Now(),
		Tables:      tables,
	}, nil
}
