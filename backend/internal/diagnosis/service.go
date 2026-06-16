package diagnosis

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/okumujustine/postgresome/backend/internal/analysis"
	"github.com/okumujustine/postgresome/backend/internal/metrics"
	"github.com/okumujustine/postgresome/backend/internal/storage/repository"
)

const FindingResolutionGracePeriod = 90 * time.Second

type Service struct {
	pool   *pgxpool.Pool
	engine *analysis.Engine
}

type RunResult struct {
	DetectedAt time.Time
	Findings   []analysis.Finding
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{
		pool:   pool,
		engine: NewDefaultEngine(),
	}
}

func (s *Service) RunForDatabase(ctx context.Context, databaseInstanceID string) (*RunResult, error) {
	dbInstance, err := repository.GetDatabaseInstance(ctx, s.pool, databaseInstanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to load database instance: %w", err)
	}
	if dbInstance == nil {
		return nil, fmt.Errorf("database instance %q not found", databaseInstanceID)
	}

	statsPair, err := repository.GetLatestDatabaseStatsPair(ctx, s.pool, databaseInstanceID, "", dbInstance.Name)
	if err != nil {
		return nil, err
	}
	if statsPair.Current == nil {
		return nil, fmt.Errorf("no metric snapshots found for database instance %q", databaseInstanceID)
	}

	detectedAt := statsPair.CollectedAt
	delta := metrics.DatabaseMetricDelta{
		CollectedAt:        detectedAt,
		DatabaseInstanceID: databaseInstanceID,
	}
	if statsPair.Previous != nil {
		delta = metrics.CalculateDatabaseDelta(statsPair.Previous, statsPair.Current, detectedAt, databaseInstanceID)
	}

	findings := s.engine.AnalyzeDatabaseMetrics(*statsPair.Current, delta)

	activitySnapshot, err := repository.ListActivityStatsAt(ctx, s.pool, databaseInstanceID, detectedAt)
	if err != nil {
		return nil, err
	}
	if !activitySnapshot.CollectedAt.IsZero() {
		findings = append(findings, s.engine.AnalyzeDatabaseActivity(*activitySnapshot.ToMetrics(databaseInstanceID))...)
	}

	tableSnapshot, err := repository.ListTableStatsAt(ctx, s.pool, databaseInstanceID, detectedAt)
	if err != nil {
		return nil, err
	}
	if !tableSnapshot.CollectedAt.IsZero() {
		findings = append(findings, s.engine.AnalyzeTableStats(tableSnapshot.ToMetrics(databaseInstanceID))...)
	}

	querySnapshot, err := repository.ListQueryStatsAt(ctx, s.pool, databaseInstanceID, detectedAt)
	if err != nil {
		return nil, err
	}
	if !querySnapshot.CollectedAt.IsZero() {
		findings = append(findings, s.engine.AnalyzeQueryStats(querySnapshot.ToMetrics(databaseInstanceID))...)
	}

	explainSnapshot, err := repository.ListExplainPlansAt(ctx, s.pool, databaseInstanceID, detectedAt)
	if err != nil {
		return nil, err
	}
	if !explainSnapshot.CollectedAt.IsZero() {
		findings = append(findings, s.engine.AnalyzeExplainPlans(*explainSnapshot.ToMetrics(databaseInstanceID))...)
	}

	for i := range findings {
		findings[i].AgentID = ""
		findings[i].DatabaseInstanceID = databaseInstanceID
		if findings[i].DetectedAt.IsZero() {
			findings[i].DetectedAt = detectedAt
		}
	}

	if err := s.enrichHistoricalDiagnosis(ctx, detectedAt, databaseInstanceID, tableSnapshot, querySnapshot, findings); err != nil {
		return nil, err
	}

	if err := repository.UpsertFindings(ctx, s.pool, findings); err != nil {
		return nil, err
	}

	if err := s.attachFindingEvidence(ctx, detectedAt, activitySnapshot, findings); err != nil {
		return nil, err
	}

	if err := s.resolveStaleFindingsWithVerification(ctx, databaseInstanceID, detectedAt, tableSnapshot, querySnapshot); err != nil {
		return nil, err
	}

	return &RunResult{
		DetectedAt: detectedAt,
		Findings:   findings,
	}, nil
}
