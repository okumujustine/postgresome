package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/okumujustine/postgresome/shared/metrics"
)

type ExplainPlanRow struct {
	QueryID string
	Query   string
	Root    metrics.PlanNode
}

type ExplainPlansSnapshot struct {
	CollectedAt time.Time
	Plans       []ExplainPlanRow
}

const deleteExplainPlansSQL = `
	DELETE FROM explain_plans WHERE database_instance_id = $1
`

const insertExplainPlanSQL = `
	INSERT INTO explain_plans (
		database_instance_id, agent_id, collected_at, query_id, query, plan_json
	) VALUES (
		$1, $2, $3, $4, $5, $6
	)
`

func ReplaceExplainPlans(ctx context.Context, pool *pgxpool.Pool, databaseInstanceID, agentID string, collectedAt time.Time, plans []ExplainPlanRow) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, deleteExplainPlansSQL, databaseInstanceID); err != nil {
		return fmt.Errorf("failed to delete existing explain plans: %w", err)
	}

	for _, plan := range plans {
		encodedPlan, err := json.Marshal(plan.Root)
		if err != nil {
			return fmt.Errorf("failed to encode explain plan for %s: %w", plan.QueryID, err)
		}

		_, err = tx.Exec(ctx, insertExplainPlanSQL, databaseInstanceID, agentID, collectedAt, plan.QueryID, plan.Query, encodedPlan)
		if err != nil {
			return fmt.Errorf("failed to insert explain plan for %s: %w", plan.QueryID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit explain plans transaction: %w", err)
	}

	return nil
}

const listExplainPlansSQL = `
	SELECT collected_at, query_id, query, plan_json
	FROM explain_plans
	WHERE database_instance_id = $1
	ORDER BY query_id ASC
`

func ListExplainPlans(ctx context.Context, pool *pgxpool.Pool, databaseInstanceID string) (ExplainPlansSnapshot, error) {
	rows, err := pool.Query(ctx, listExplainPlansSQL, databaseInstanceID)
	if err != nil {
		return ExplainPlansSnapshot{}, fmt.Errorf("failed to query explain plans: %w", err)
	}
	defer rows.Close()

	snapshot := ExplainPlansSnapshot{Plans: make([]ExplainPlanRow, 0)}

	for rows.Next() {
		var (
			plan        ExplainPlanRow
			collectedAt time.Time
			rawPlan     []byte
		)

		if err := rows.Scan(&collectedAt, &plan.QueryID, &plan.Query, &rawPlan); err != nil {
			return ExplainPlansSnapshot{}, fmt.Errorf("failed to scan explain plan: %w", err)
		}

		if err := json.Unmarshal(rawPlan, &plan.Root); err != nil {
			return ExplainPlansSnapshot{}, fmt.Errorf("failed to decode explain plan for %s: %w", plan.QueryID, err)
		}

		snapshot.CollectedAt = collectedAt
		snapshot.Plans = append(snapshot.Plans, plan)
	}

	if err := rows.Err(); err != nil {
		return ExplainPlansSnapshot{}, fmt.Errorf("failed to read explain plans: %w", err)
	}

	return snapshot, nil
}

func (s ExplainPlansSnapshot) ToMetrics(databaseInstanceID string) *metrics.ExplainSnapshot {
	plans := make([]metrics.ExplainPlan, len(s.Plans))
	for i, plan := range s.Plans {
		plans[i] = metrics.ExplainPlan{
			QueryID: plan.QueryID,
			Query:   plan.Query,
			Root:    plan.Root,
		}
	}

	return &metrics.ExplainSnapshot{
		CollectedAt:        s.CollectedAt,
		DatabaseInstanceID: databaseInstanceID,
		Plans:              plans,
	}
}
