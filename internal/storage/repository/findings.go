package repository

import (
	"context"
	"fmt"
	"time"

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

const countFindingsBySeveritySQL = `
	SELECT severity, count(*)
	FROM findings
	WHERE created_at >= $1
	  AND ($2 = '' OR database_instance_id = $2)
	  AND ($3 = '' OR agent_id = $3)
	GROUP BY severity
`

// FindingSeverityCounts holds the number of findings detected at or after a
// given time, grouped by severity.
type FindingSeverityCounts struct {
	Critical int
	Warning  int
	Info     int
}

// CountFindingsBySeverity counts findings detected since the given time,
// grouped by severity.
func CountFindingsBySeverity(ctx context.Context, pool *pgxpool.Pool, databaseInstanceID, agentID string, since time.Time) (FindingSeverityCounts, error) {
	rows, err := pool.Query(ctx, countFindingsBySeveritySQL, since, databaseInstanceID, agentID)
	if err != nil {
		return FindingSeverityCounts{}, fmt.Errorf("failed to count findings by severity: %w", err)
	}
	defer rows.Close()

	var counts FindingSeverityCounts
	for rows.Next() {
		var severity string
		var count int
		if err := rows.Scan(&severity, &count); err != nil {
			return FindingSeverityCounts{}, fmt.Errorf("failed to scan finding severity count: %w", err)
		}

		switch severity {
		case "critical":
			counts.Critical = count
		case "warning":
			counts.Warning = count
		case "info":
			counts.Info = count
		}
	}

	if err := rows.Err(); err != nil {
		return FindingSeverityCounts{}, fmt.Errorf("failed to read finding severity counts: %w", err)
	}

	return counts, nil
}

const getRecentFindingsSQL = `
	SELECT id::text, severity, category, title, message, recommendation, created_at
	FROM findings
	WHERE created_at >= $1
	  AND ($2 = '' OR database_instance_id = $2)
	  AND ($3 = '' OR agent_id = $3)
	ORDER BY created_at DESC
	LIMIT $4
`

// RecentFinding is a finding row returned for dashboard summaries.
type RecentFinding struct {
	ID             string
	Severity       string
	Category       string
	Title          string
	Message        string
	Recommendation string
	DetectedAt     time.Time
}

// GetRecentFindings returns the most recent findings detected since the given
// time, newest first, up to limit results.
func GetRecentFindings(ctx context.Context, pool *pgxpool.Pool, databaseInstanceID, agentID string, since time.Time, limit int) ([]RecentFinding, error) {
	rows, err := pool.Query(ctx, getRecentFindingsSQL, since, databaseInstanceID, agentID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent findings: %w", err)
	}
	defer rows.Close()

	findings := make([]RecentFinding, 0)
	for rows.Next() {
		var f RecentFinding
		if err := rows.Scan(&f.ID, &f.Severity, &f.Category, &f.Title, &f.Message, &f.Recommendation, &f.DetectedAt); err != nil {
			return nil, fmt.Errorf("failed to scan recent finding: %w", err)
		}
		findings = append(findings, f)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to read recent findings: %w", err)
	}

	return findings, nil
}

const listFindingsSQL = `
	SELECT id::text, severity, category, title, message, recommendation, created_at
	FROM findings
	WHERE created_at >= $1
	  AND ($2 = '' OR database_instance_id = $2)
	  AND ($3 = '' OR agent_id = $3)
	  AND ($4 = '' OR severity = $4)
	  AND ($5 = '' OR category = $5)
	ORDER BY created_at DESC
	LIMIT $6 OFFSET $7
`

// ListFindingsParams filters and paginates findings returned by
// ListFindings and CountFindings.
type ListFindingsParams struct {
	DatabaseInstanceID string
	AgentID            string
	Severity           string
	Category           string
	Since              time.Time
	Limit              int
	Offset             int
}

// ListFindings returns a page of findings matching the given filters,
// newest first.
func ListFindings(ctx context.Context, pool *pgxpool.Pool, params ListFindingsParams) ([]RecentFinding, error) {
	rows, err := pool.Query(ctx, listFindingsSQL,
		params.Since,
		params.DatabaseInstanceID,
		params.AgentID,
		params.Severity,
		params.Category,
		params.Limit,
		params.Offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query findings: %w", err)
	}
	defer rows.Close()

	findings := make([]RecentFinding, 0)
	for rows.Next() {
		var f RecentFinding
		if err := rows.Scan(&f.ID, &f.Severity, &f.Category, &f.Title, &f.Message, &f.Recommendation, &f.DetectedAt); err != nil {
			return nil, fmt.Errorf("failed to scan finding: %w", err)
		}
		findings = append(findings, f)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to read findings: %w", err)
	}

	return findings, nil
}

const countFindingsSQL = `
	SELECT count(*)
	FROM findings
	WHERE created_at >= $1
	  AND ($2 = '' OR database_instance_id = $2)
	  AND ($3 = '' OR agent_id = $3)
	  AND ($4 = '' OR severity = $4)
	  AND ($5 = '' OR category = $5)
`

// CountFindings returns the total number of findings matching the given
// filters, ignoring Limit and Offset.
func CountFindings(ctx context.Context, pool *pgxpool.Pool, params ListFindingsParams) (int, error) {
	var count int

	err := pool.QueryRow(ctx, countFindingsSQL,
		params.Since,
		params.DatabaseInstanceID,
		params.AgentID,
		params.Severity,
		params.Category,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count findings: %w", err)
	}

	return count, nil
}

const countFindingsByCategoryTitleSQL = `
	SELECT count(*)
	FROM findings
	WHERE category = $1
	  AND title ILIKE $2
	  AND created_at >= $3
	  AND created_at < $4
	  AND ($5 = '' OR database_instance_id = $5)
	  AND ($6 = '' OR agent_id = $6)
`

// CountFindingsByCategoryTitle counts findings in [start, end) matching the
// given category and a title pattern (using SQL ILIKE wildcards).
func CountFindingsByCategoryTitle(ctx context.Context, pool *pgxpool.Pool, category, titlePattern, databaseInstanceID, agentID string, start, end time.Time) (int, error) {
	var count int

	err := pool.QueryRow(ctx, countFindingsByCategoryTitleSQL, category, titlePattern, start, end, databaseInstanceID, agentID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count findings by category and title: %w", err)
	}

	return count, nil
}
