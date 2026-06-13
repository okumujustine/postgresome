package explain

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/okumujustine/postgresome/internal/metrics"
)

const maxExplainedQueries = 5

type ExplainCollector struct {
	pool *pgxpool.Pool
}

func NewExplainCollector(pool *pgxpool.Pool) *ExplainCollector {
	return &ExplainCollector{
		pool: pool,
	}
}

// ProbeGenericPlanSupport checks whether EXPLAIN (GENERIC_PLAN) is supported
// by the connected PostgreSQL server. GENERIC_PLAN was added in PostgreSQL
// 16 and allows planning a parameterized statement without supplying
// parameter values.
func (c *ExplainCollector) ProbeGenericPlanSupport(ctx context.Context) bool {
	_, err := c.pool.Exec(ctx, "EXPLAIN (GENERIC_PLAN, FORMAT JSON) SELECT 1")
	return err == nil
}

type explainResult struct {
	Plan metrics.PlanNode `json:"Plan"`
}

// GetExplainPlans runs EXPLAIN (FORMAT JSON, GENERIC_PLAN) for up to
// maxExplainedQueries queries from snapshot that belong to currentDatabase,
// in the order they appear (snapshot.Queries is sorted by total execution
// time descending). Queries that fail to plan - utility statements,
// truncated query text, dropped objects, etc. - are skipped.
func (c *ExplainCollector) GetExplainPlans(ctx context.Context, snapshot metrics.QueryStatsSnapshot, currentDatabase string) *metrics.ExplainSnapshot {
	plans := make([]metrics.ExplainPlan, 0)

	conn, err := c.pool.Acquire(ctx)
	if err != nil {
		return &metrics.ExplainSnapshot{CollectedAt: snapshot.CollectedAt, Plans: plans}
	}
	defer conn.Release()

	pgConn := conn.Conn().PgConn()

	count := 0
	for _, q := range snapshot.Queries {
		if q.DatabaseName != currentDatabase {
			continue
		}

		if count >= maxExplainedQueries {
			break
		}

		// q.Query is the normalized pg_stat_statements text and retains $1,
		// $2, ... placeholders, which only PostgreSQL's GENERIC_PLAN-aware
		// planner can interpret. pgx's normal Query/Exec methods always
		// treat $N as bind parameters and fail when no values are supplied,
		// so the statement is sent via the raw simple query protocol
		// instead - the same way psql sends it.
		results, err := pgConn.Exec(ctx, "EXPLAIN (FORMAT JSON, GENERIC_PLAN) "+q.Query).ReadAll()
		if err != nil || len(results) == 0 || len(results[0].Rows) == 0 || len(results[0].Rows[0]) == 0 {
			continue
		}

		count++

		var parsed []explainResult
		if err := json.Unmarshal(results[0].Rows[0][0], &parsed); err != nil || len(parsed) == 0 {
			continue
		}

		plans = append(plans, metrics.ExplainPlan{
			QueryID: q.QueryID,
			Query:   q.Query,
			Root:    parsed[0].Plan,
		})
	}

	return &metrics.ExplainSnapshot{
		CollectedAt: snapshot.CollectedAt,
		Plans:       plans,
	}
}
