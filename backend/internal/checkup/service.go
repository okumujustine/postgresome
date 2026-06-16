package checkup

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/okumujustine/postgresome/backend/internal/analysis"
	"github.com/okumujustine/postgresome/backend/internal/diagnosis"
	"github.com/okumujustine/postgresome/backend/internal/secrets"
	"github.com/okumujustine/postgresome/backend/internal/storage/repository"
)

type Service struct {
	pool                *pgxpool.Pool
	collector           *Collector
	diagnosis           *diagnosis.Service
	connectionProtector *secrets.ConnectionProtector
}

type RunResult struct {
	SourceID           string
	DatabaseInstanceID string
	CollectedAt        time.Time
	Findings           []analysis.Finding
	Warnings           []string
}

func NewService(pool *pgxpool.Pool, connectionProtector *secrets.ConnectionProtector) *Service {
	return &Service{
		pool:                pool,
		collector:           NewCollector(),
		diagnosis:           diagnosis.NewService(pool),
		connectionProtector: connectionProtector,
	}
}

func (s *Service) RunForSource(ctx context.Context, sourceID string) (*RunResult, error) {
	source, err := repository.GetSource(ctx, s.pool, sourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to load source: %w", err)
	}
	if source == nil {
		return nil, fmt.Errorf("source %q not found", sourceID)
	}

	instance, err := repository.GetDatabaseInstanceBySourceID(ctx, s.pool, sourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to load database instance for source: %w", err)
	}
	if instance == nil {
		return nil, fmt.Errorf("source %q has no database instance", sourceID)
	}

	profile, err := repository.GetSourceConnectionProfile(ctx, s.pool, sourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to load source connection profile: %w", err)
	}
	if profile == nil {
		return nil, fmt.Errorf("source %q has no saved connection string", sourceID)
	}

	connectionURI, err := s.resolveConnectionURI(profile)
	if err != nil {
		return nil, err
	}

	startedAt := time.Now().UTC()
	if err := repository.MarkSourceCheckRunning(ctx, s.pool, sourceID, startedAt); err != nil {
		return nil, err
	}

	evidence, err := s.collector.Collect(ctx, sourceID, instance.ID, connectionURI)
	if err != nil {
		_ = repository.MarkSourceCheckFinished(ctx, s.pool, sourceID, time.Now().UTC(), "failed", err.Error())
		return nil, err
	}

	if err := repository.InsertMetricPoints(ctx, s.pool, evidence.MetricPoints); err != nil {
		_ = repository.MarkSourceCheckFinished(ctx, s.pool, sourceID, time.Now().UTC(), "failed", err.Error())
		return nil, err
	}

	if len(evidence.Activities) > 0 {
		if err := repository.ReplaceActivityStats(ctx, s.pool, instance.ID, "", evidence.CollectedAt, evidence.Activities); err != nil {
			_ = repository.MarkSourceCheckFinished(ctx, s.pool, sourceID, time.Now().UTC(), "failed", err.Error())
			return nil, err
		}
	}

	if len(evidence.Tables) > 0 {
		if err := repository.ReplaceTableStats(ctx, s.pool, instance.ID, "", evidence.CollectedAt, evidence.Tables); err != nil {
			_ = repository.MarkSourceCheckFinished(ctx, s.pool, sourceID, time.Now().UTC(), "failed", err.Error())
			return nil, err
		}
	}

	if len(evidence.Queries) > 0 {
		if err := repository.ReplaceQueryStats(ctx, s.pool, instance.ID, "", evidence.CollectedAt, evidence.Queries); err != nil {
			_ = repository.MarkSourceCheckFinished(ctx, s.pool, sourceID, time.Now().UTC(), "failed", err.Error())
			return nil, err
		}
	}

	if len(evidence.Plans) > 0 {
		if err := repository.ReplaceExplainPlans(ctx, s.pool, instance.ID, "", evidence.CollectedAt, evidence.Plans); err != nil {
			_ = repository.MarkSourceCheckFinished(ctx, s.pool, sourceID, time.Now().UTC(), "failed", err.Error())
			return nil, err
		}
	}

	diagnosisResult, err := s.diagnosis.RunForDatabase(ctx, instance.ID)
	if err != nil {
		_ = repository.MarkSourceCheckFinished(ctx, s.pool, sourceID, time.Now().UTC(), "failed", err.Error())
		return nil, err
	}

	if err := repository.MarkSourceCheckFinished(ctx, s.pool, sourceID, time.Now().UTC(), "succeeded", ""); err != nil {
		return nil, err
	}

	return &RunResult{
		SourceID:           sourceID,
		DatabaseInstanceID: instance.ID,
		CollectedAt:        evidence.CollectedAt,
		Findings:           diagnosisResult.Findings,
		Warnings:           evidence.Warnings,
	}, nil
}

func (s *Service) resolveConnectionURI(profile *repository.SourceConnectionProfile) (string, error) {
	if strings.TrimSpace(profile.ConnectionURIEncrypted) != "" {
		connectionURI, err := s.connectionProtector.Decrypt(profile.ConnectionURIEncrypted)
		if err != nil {
			return "", fmt.Errorf("failed to decrypt saved source connection string: %w", err)
		}
		return connectionURI, nil
	}

	if strings.TrimSpace(profile.ConnectionURI) != "" {
		return profile.ConnectionURI, nil
	}

	return "", fmt.Errorf("source %q has no saved connection string", profile.SourceID)
}
