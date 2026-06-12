package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/okumujustine/postgresome/internal/analysis"
)

const insertFindingSQL = `
	INSERT INTO findings (id, severity, category, title, message, recommendation, database_instance_id, agent_id, created_at)
	VALUES (COALESCE(NULLIF($1, '')::uuid, gen_random_uuid()), $2, $3, $4, $5, $6, $7, NULLIF($8, ''), $9)
`

// InsertFindings persists analyzer findings into the findings table.
func InsertFindings(ctx context.Context, pool *pgxpool.Pool, findings []analysis.Finding) error {
	for _, finding := range findings {
		_, err := pool.Exec(ctx, insertFindingSQL,
			finding.ID,
			finding.Severity,
			finding.Category,
			finding.Title,
			finding.Message,
			finding.Recommendation,
			finding.DatabaseInstanceID,
			finding.AgentID,
			finding.DetectedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert finding %q: %w", finding.Title, err)
		}
	}

	return nil
}
