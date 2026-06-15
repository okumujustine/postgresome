package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/okumujustine/postgresome/cloud/internal/analysis"
)

type FindingEvidenceRow struct {
	ID string

	ObservedAt time.Time

	EvidenceType string
	Role         string
	Label        string
	Summary      string
	MetricKey    string
	ReferenceID  string

	CurrentValue    float64
	BaselineValue   float64
	ChangePercent   float64
	ConfidenceScore float64

	Metadata map[string]any
}

const deleteFindingEvidenceSQL = `
	DELETE FROM finding_evidence
	WHERE finding_id = $1::uuid
`

const insertFindingEvidenceSQL = `
	INSERT INTO finding_evidence (
		finding_id, observed_at, evidence_type, role, label, summary,
		metric_key, current_value, baseline_value, change_percent,
		confidence_score, reference_id, position, metadata
	) VALUES (
		$1::uuid, $2, $3, $4, $5, $6,
		$7, $8, $9, $10,
		$11, $12, $13, $14::jsonb
	)
`

func ReplaceFindingEvidence(ctx context.Context, pool *pgxpool.Pool, findingID string, evidence []analysis.FindingEvidence) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin finding evidence transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, deleteFindingEvidenceSQL, findingID); err != nil {
		return fmt.Errorf("failed to clear finding evidence for %s: %w", findingID, err)
	}

	for i, item := range evidence {
		metadata := item.Metadata
		if metadata == nil {
			metadata = map[string]any{}
		}

		payload, err := json.Marshal(metadata)
		if err != nil {
			return fmt.Errorf("failed to encode finding evidence metadata for %s: %w", findingID, err)
		}

		_, err = tx.Exec(ctx, insertFindingEvidenceSQL,
			findingID,
			item.ObservedAt,
			item.EvidenceType,
			item.Role,
			item.Label,
			item.Summary,
			item.MetricKey,
			item.CurrentValue,
			item.BaselineValue,
			item.ChangePercent,
			item.ConfidenceScore,
			item.ReferenceID,
			i,
			payload,
		)
		if err != nil {
			return fmt.Errorf("failed to insert finding evidence for %s: %w", findingID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit finding evidence for %s: %w", findingID, err)
	}

	return nil
}

const listFindingEvidenceSQL = `
	SELECT
		id::text,
		observed_at,
		evidence_type,
		role,
		label,
		summary,
		metric_key,
		current_value,
		baseline_value,
		change_percent,
		confidence_score,
		reference_id,
		metadata
	FROM finding_evidence
	WHERE finding_id = $1::uuid
	ORDER BY position ASC, observed_at DESC
`

func ListFindingEvidence(ctx context.Context, pool *pgxpool.Pool, findingID string) ([]FindingEvidenceRow, error) {
	rows, err := pool.Query(ctx, listFindingEvidenceSQL, findingID)
	if err != nil {
		return nil, fmt.Errorf("failed to query finding evidence for %s: %w", findingID, err)
	}
	defer rows.Close()

	evidence := make([]FindingEvidenceRow, 0)
	for rows.Next() {
		var (
			item        FindingEvidenceRow
			metadataRaw []byte
		)

		if err := rows.Scan(
			&item.ID,
			&item.ObservedAt,
			&item.EvidenceType,
			&item.Role,
			&item.Label,
			&item.Summary,
			&item.MetricKey,
			&item.CurrentValue,
			&item.BaselineValue,
			&item.ChangePercent,
			&item.ConfidenceScore,
			&item.ReferenceID,
			&metadataRaw,
		); err != nil {
			return nil, fmt.Errorf("failed to scan finding evidence for %s: %w", findingID, err)
		}

		item.Metadata = map[string]any{}
		if len(metadataRaw) > 0 {
			if err := json.Unmarshal(metadataRaw, &item.Metadata); err != nil {
				return nil, fmt.Errorf("failed to decode finding evidence metadata for %s: %w", findingID, err)
			}
		}

		evidence = append(evidence, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to read finding evidence for %s: %w", findingID, err)
	}

	return evidence, nil
}
