package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/okumujustine/postgresome/cloud/internal/analysis"
)

const upsertFindingSQL = `
	INSERT INTO findings (
		severity, category, title, message, recommendation,
		problem_summary, evidence_summary, impact_summary, suggested_action,
		confidence_label, confidence_score, baseline_value, baseline_label,
		change_summary, verification_hint, verification_status, verification_summary,
		database_instance_id, agent_id,
		rule_key, resource_type, resource_name, fingerprint,
		current_value, threshold_value,
		status, occurrence_count, regression_count, first_seen_at, last_seen_at, created_at
	) VALUES (
		$1, $2, $3, $4, $5,
		$6, $7, $8, $9,
		$10, $11, $12, $13,
		$14, $15, COALESCE(NULLIF($16, ''), 'pending'), $17,
		$18, NULLIF($19, ''),
		$20, $21, $22, $23,
		$24, $25,
		'open', 1, 0, $26, $26, $26
	)
	ON CONFLICT (database_instance_id, fingerprint) DO UPDATE SET
		severity = EXCLUDED.severity,
		category = EXCLUDED.category,
		title = EXCLUDED.title,
		message = EXCLUDED.message,
		recommendation = EXCLUDED.recommendation,
		problem_summary = EXCLUDED.problem_summary,
		evidence_summary = EXCLUDED.evidence_summary,
		impact_summary = EXCLUDED.impact_summary,
		suggested_action = EXCLUDED.suggested_action,
		confidence_label = EXCLUDED.confidence_label,
		confidence_score = EXCLUDED.confidence_score,
		baseline_value = EXCLUDED.baseline_value,
		baseline_label = EXCLUDED.baseline_label,
		change_summary = EXCLUDED.change_summary,
		verification_hint = EXCLUDED.verification_hint,
		verification_status = CASE
			WHEN findings.status = 'resolved' THEN 'regressed'
			WHEN EXCLUDED.verification_status <> '' THEN EXCLUDED.verification_status
			ELSE 'pending'
		END,
		verification_summary = CASE
			WHEN findings.status = 'resolved' THEN 'This issue recurred after previously resolving.'
			WHEN EXCLUDED.verification_summary <> '' THEN EXCLUDED.verification_summary
			ELSE findings.verification_summary
		END,
		agent_id = EXCLUDED.agent_id,
		current_value = EXCLUDED.current_value,
		threshold_value = EXCLUDED.threshold_value,
		status = 'open',
		resolved_at = NULL,
		occurrence_count = findings.occurrence_count + 1,
		regression_count = CASE
			WHEN findings.status = 'resolved' THEN findings.regression_count + 1
			ELSE findings.regression_count
		END,
		last_regressed_at = CASE
			WHEN findings.status = 'resolved' THEN EXCLUDED.last_seen_at
			ELSE findings.last_regressed_at
		END,
		improving_since = CASE
			WHEN findings.status = 'resolved' THEN NULL
			WHEN EXCLUDED.verification_status = 'improving' THEN COALESCE(findings.improving_since, EXCLUDED.last_seen_at)
			ELSE NULL
		END,
		verified_fixed_at = NULL,
		last_seen_at = EXCLUDED.last_seen_at
	RETURNING id::text
`

// UpsertFindings persists analyzer findings, deduplicating on
// (database_instance_id, fingerprint). A finding whose fingerprint matches
// an existing row bumps occurrence_count and last_seen_at, and reopens the
// finding (status='open', resolved_at=NULL) if it had previously been
// resolved.
func UpsertFindings(ctx context.Context, pool *pgxpool.Pool, findings []analysis.Finding) error {
	for i, finding := range findings {
		err := pool.QueryRow(ctx, upsertFindingSQL,
			finding.Severity,
			finding.Category,
			finding.Title,
			finding.Message,
			finding.Recommendation,
			finding.ProblemSummary,
			finding.EvidenceSummary,
			finding.ImpactSummary,
			finding.SuggestedAction,
			finding.ConfidenceLabel,
			finding.ConfidenceScore,
			finding.BaselineValue,
			finding.BaselineLabel,
			finding.ChangeSummary,
			finding.VerificationHint,
			finding.VerificationStatus,
			finding.VerificationSummary,
			finding.DatabaseInstanceID,
			finding.AgentID,
			finding.RuleKey,
			finding.ResourceType,
			finding.ResourceName,
			finding.Fingerprint(),
			finding.CurrentValue,
			finding.ThresholdValue,
			finding.DetectedAt,
		).Scan(&findings[i].ID)
		if err != nil {
			return fmt.Errorf("failed to upsert finding %q: %w", finding.Title, err)
		}
	}

	return nil
}

const resolveStaleFindingsSQL = `
	UPDATE findings
	SET status = 'resolved', resolved_at = $2
	WHERE database_instance_id = $1
	  AND status = 'open'
	  AND last_seen_at < $3
`

// ResolveStaleFindings marks open findings for the given database instance as
// resolved if they were not re-detected within the grace period ending at
// now (last_seen_at < now - gracePeriod). It should be run once per ingest
// cycle, after upserting that cycle's findings, so issues that stop
// recurring are automatically closed.
func ResolveStaleFindings(ctx context.Context, pool *pgxpool.Pool, databaseInstanceID string, now time.Time, gracePeriod time.Duration) error {
	cutoff := now.Add(-gracePeriod)

	_, err := pool.Exec(ctx, resolveStaleFindingsSQL, databaseInstanceID, now, cutoff)
	if err != nil {
		return fmt.Errorf("failed to resolve stale findings: %w", err)
	}

	return nil
}

const countFindingsBySeveritySQL = `
	SELECT severity, count(*)
	FROM findings
	WHERE status = 'open'
	  AND ($1 = '' OR database_instance_id = $1)
	  AND ($2 = '' OR agent_id = $2)
	GROUP BY severity
`

// FindingSeverityCounts holds the number of currently open findings, grouped
// by severity.
type FindingSeverityCounts struct {
	Critical int
	Warning  int
	Info     int
}

// CountFindingsBySeverity counts currently open findings, grouped by
// severity.
func CountFindingsBySeverity(ctx context.Context, pool *pgxpool.Pool, databaseInstanceID, agentID string) (FindingSeverityCounts, error) {
	rows, err := pool.Query(ctx, countFindingsBySeveritySQL, databaseInstanceID, agentID)
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

const findingSelectColumns = `
	id::text, severity, category, title, message, recommendation, status,
	problem_summary, evidence_summary, impact_summary, suggested_action,
	confidence_label, confidence_score, baseline_value, baseline_label,
	change_summary, verification_hint, verification_status, verification_summary,
	rule_key, resource_type, resource_name, current_value, threshold_value,
	occurrence_count, regression_count, first_seen_at, last_seen_at,
	last_regressed_at, improving_since, verified_fixed_at,
	COALESCE(database_instance_id::text, ''), COALESCE(agent_id::text, '')
`

const getRecentFindingsSQL = `
	SELECT ` + findingSelectColumns + `
	FROM findings
	WHERE status = 'open'
	  AND last_seen_at >= $1
	  AND ($2 = '' OR database_instance_id = $2)
	  AND ($3 = '' OR agent_id = $3)
	ORDER BY last_seen_at DESC
	LIMIT $4
`

// RecentFinding is a finding row returned for dashboard summaries and
// findings list views.
type RecentFinding struct {
	ID             string
	Severity       string
	Category       string
	Title          string
	Message        string
	Recommendation string
	Status         string

	ProblemSummary      string
	EvidenceSummary     string
	ImpactSummary       string
	SuggestedAction     string
	ConfidenceLabel     string
	ConfidenceScore     float64
	BaselineValue       float64
	BaselineLabel       string
	ChangeSummary       string
	VerificationHint    string
	VerificationStatus  string
	VerificationSummary string

	RuleKey      string
	ResourceType string
	ResourceName string

	CurrentValue   float64
	ThresholdValue float64

	OccurrenceCount    int
	RegressionCount    int
	FirstSeenAt        time.Time
	LastSeenAt         time.Time
	LastRegressedAt    *time.Time
	ImprovingSince     *time.Time
	VerifiedFixedAt    *time.Time
	DatabaseInstanceID string
	AgentID            string
}

func scanRecentFinding(scanner interface {
	Scan(dest ...any) error
}) (RecentFinding, error) {
	var f RecentFinding
	err := scanner.Scan(
		&f.ID,
		&f.Severity,
		&f.Category,
		&f.Title,
		&f.Message,
		&f.Recommendation,
		&f.Status,
		&f.ProblemSummary,
		&f.EvidenceSummary,
		&f.ImpactSummary,
		&f.SuggestedAction,
		&f.ConfidenceLabel,
		&f.ConfidenceScore,
		&f.BaselineValue,
		&f.BaselineLabel,
		&f.ChangeSummary,
		&f.VerificationHint,
		&f.VerificationStatus,
		&f.VerificationSummary,
		&f.RuleKey,
		&f.ResourceType,
		&f.ResourceName,
		&f.CurrentValue,
		&f.ThresholdValue,
		&f.OccurrenceCount,
		&f.RegressionCount,
		&f.FirstSeenAt,
		&f.LastSeenAt,
		&f.LastRegressedAt,
		&f.ImprovingSince,
		&f.VerifiedFixedAt,
		&f.DatabaseInstanceID,
		&f.AgentID,
	)
	if err != nil {
		return RecentFinding{}, err
	}

	return f, nil
}

// GetRecentFindings returns open findings last seen at or after the given
// time, most recently active first, up to limit results.
func GetRecentFindings(ctx context.Context, pool *pgxpool.Pool, databaseInstanceID, agentID string, since time.Time, limit int) ([]RecentFinding, error) {
	rows, err := pool.Query(ctx, getRecentFindingsSQL, since, databaseInstanceID, agentID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent findings: %w", err)
	}
	defer rows.Close()

	findings := make([]RecentFinding, 0)
	for rows.Next() {
		f, err := scanRecentFinding(rows)
		if err != nil {
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
	SELECT ` + findingSelectColumns + `
	FROM findings
	WHERE last_seen_at >= $1
	  AND ($2 = '' OR database_instance_id = $2)
	  AND ($3 = '' OR agent_id = $3)
	  AND ($4 = '' OR severity = $4)
	  AND ($5 = '' OR category = $5)
	  AND ($6 = '' OR status = $6)
	ORDER BY last_seen_at DESC
	LIMIT $7 OFFSET $8
`

// ListFindingsParams filters and paginates findings returned by
// ListFindings and CountFindings. Since filters on last_seen_at, and Status
// filters on the finding's lifecycle status ("open" or "resolved"); an empty
// Status matches both.
type ListFindingsParams struct {
	DatabaseInstanceID string
	AgentID            string
	Severity           string
	Category           string
	Status             string
	Since              time.Time
	Limit              int
	Offset             int
}

// ListFindings returns a page of findings matching the given filters, most
// recently active first.
func ListFindings(ctx context.Context, pool *pgxpool.Pool, params ListFindingsParams) ([]RecentFinding, error) {
	rows, err := pool.Query(ctx, listFindingsSQL,
		params.Since,
		params.DatabaseInstanceID,
		params.AgentID,
		params.Severity,
		params.Category,
		params.Status,
		params.Limit,
		params.Offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query findings: %w", err)
	}
	defer rows.Close()

	findings := make([]RecentFinding, 0)
	for rows.Next() {
		f, err := scanRecentFinding(rows)
		if err != nil {
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
	WHERE last_seen_at >= $1
	  AND ($2 = '' OR database_instance_id = $2)
	  AND ($3 = '' OR agent_id = $3)
	  AND ($4 = '' OR severity = $4)
	  AND ($5 = '' OR category = $5)
	  AND ($6 = '' OR status = $6)
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
		params.Status,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count findings: %w", err)
	}

	return count, nil
}

const getFindingByIDSQL = `
	SELECT ` + findingSelectColumns + `
	FROM findings
	WHERE id::text = $1
	  AND ($2 = '' OR database_instance_id = $2)
	  AND ($3 = '' OR agent_id = $3)
	LIMIT 1
`

// GetFindingByID returns a single finding by id, optionally scoped to a
// database instance and legacy agent id. It returns nil when no matching finding
// exists.
func GetFindingByID(ctx context.Context, pool *pgxpool.Pool, id, databaseInstanceID, agentID string) (*RecentFinding, error) {
	f, err := scanRecentFinding(pool.QueryRow(ctx, getFindingByIDSQL, id, databaseInstanceID, agentID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to load finding %q: %w", id, err)
	}

	return &f, nil
}

const listStaleOpenFindingsSQL = `
	SELECT ` + findingSelectColumns + `
	FROM findings
	WHERE database_instance_id = $1
	  AND status = 'open'
	  AND last_seen_at < $2
	ORDER BY last_seen_at ASC
`

// ListStaleOpenFindings returns open findings whose last_seen_at is older
// than the provided cutoff.
func ListStaleOpenFindings(ctx context.Context, pool *pgxpool.Pool, databaseInstanceID string, cutoff time.Time) ([]RecentFinding, error) {
	rows, err := pool.Query(ctx, listStaleOpenFindingsSQL, databaseInstanceID, cutoff)
	if err != nil {
		return nil, fmt.Errorf("failed to query stale open findings: %w", err)
	}
	defer rows.Close()

	findings := make([]RecentFinding, 0)
	for rows.Next() {
		f, err := scanRecentFinding(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stale open finding: %w", err)
		}
		findings = append(findings, f)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to read stale open findings: %w", err)
	}

	return findings, nil
}

const resolveFindingSQL = `
	UPDATE findings
	SET status = 'resolved',
	    resolved_at = $2,
	    verification_status = COALESCE(NULLIF($3, ''), verification_status),
	    verification_summary = COALESCE($4, verification_summary),
	    verified_fixed_at = $5
	WHERE id::text = $1
`

// ResolveFinding applies lifecycle metadata to a single finding as it
// transitions from open to resolved.
func ResolveFinding(ctx context.Context, pool *pgxpool.Pool, findingID string, resolvedAt time.Time, verificationStatus, verificationSummary string, verifiedFixedAt *time.Time) error {
	_, err := pool.Exec(ctx, resolveFindingSQL, findingID, resolvedAt, verificationStatus, verificationSummary, verifiedFixedAt)
	if err != nil {
		return fmt.Errorf("failed to resolve finding %q: %w", findingID, err)
	}
	return nil
}

const countFindingsByCategoryTitleSQL = `
	SELECT count(*)
	FROM findings
	WHERE category = $1
	  AND title ILIKE $2
	  AND first_seen_at >= $3
	  AND first_seen_at < $4
	  AND ($5 = '' OR database_instance_id = $5)
	  AND ($6 = '' OR agent_id = $6)
`

// CountFindingsByCategoryTitle counts findings first detected in [start, end)
// matching the given category and a title pattern (using SQL ILIKE
// wildcards).
func CountFindingsByCategoryTitle(ctx context.Context, pool *pgxpool.Pool, category, titlePattern, databaseInstanceID, agentID string, start, end time.Time) (int, error) {
	var count int

	err := pool.QueryRow(ctx, countFindingsByCategoryTitleSQL, category, titlePattern, start, end, databaseInstanceID, agentID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count findings by category and title: %w", err)
	}

	return count, nil
}
