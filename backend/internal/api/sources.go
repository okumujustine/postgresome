package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
	"unicode"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/okumujustine/postgresome/backend/internal/checkup"
	"github.com/okumujustine/postgresome/backend/internal/storage"
	"github.com/okumujustine/postgresome/backend/internal/storage/repository"
)

var allowedSourceKinds = map[string]struct{}{
	"direct":              {},
	"managed_integration": {},
}

var allowedSourceProviders = map[string]struct{}{
	"postgres": {},
	"supabase": {},
	"rds":      {},
	"cloudsql": {},
	"neon":     {},
}

type sourceDTO struct {
	ID                   string     `json:"id"`
	Kind                 string     `json:"kind"`
	Provider             string     `json:"provider"`
	Name                 string     `json:"name"`
	Configured           bool       `json:"configured"`
	SetupState           string     `json:"setup_state"`
	LastCheckStatus      string     `json:"last_check_status"`
	LastCheckStartedAt   *time.Time `json:"last_check_started_at,omitempty"`
	LastCheckCompletedAt *time.Time `json:"last_check_completed_at,omitempty"`
	LastCheckError       string     `json:"last_check_error,omitempty"`
}

type sourceDatabaseDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Host string `json:"host"`
}

type sourceRecordDTO struct {
	Source   sourceDTO                    `json:"source"`
	Database sourceDatabaseDTO            `json:"database"`
	Instance dashboardDatabaseInstanceDTO `json:"instance"`
}

type listSourcesResponse struct {
	Sources []sourceRecordDTO `json:"sources"`
}

type createSourceRequest struct {
	Source struct {
		ID       string `json:"id"`
		Kind     string `json:"kind"`
		Provider string `json:"provider"`
		Name     string `json:"name"`
	} `json:"source"`
	Database struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Host string `json:"host"`
	} `json:"database"`
	Connection struct {
		URI string `json:"uri"`
	} `json:"connection"`
}

type runSourceCheckupResponse struct {
	Status             string                `json:"status"`
	DatabaseInstanceID string                `json:"database_instance_id"`
	DetectedAt         string                `json:"detected_at"`
	Warnings           []string              `json:"warnings"`
	Findings           []dashboardFindingDTO `json:"findings"`
}

func (s *Server) handleListSources(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	sources, err := repository.ListSources(ctx, s.pool)
	if err != nil {
		log.Printf("failed to list sources: %v", err)
		http.Error(w, "failed to list sources", http.StatusInternalServerError)
		return
	}

	instances, err := repository.ListDatabaseInstances(ctx, s.pool)
	if err != nil {
		log.Printf("failed to list database instances for sources: %v", err)
		http.Error(w, "failed to list sources", http.StatusInternalServerError)
		return
	}

	instancesBySource := make(map[string]repository.DatabaseInstance, len(instances))
	for _, instance := range instances {
		instancesBySource[instance.SourceID] = instance
	}

	items := make([]sourceRecordDTO, 0, len(sources))
	for _, source := range sources {
		instance, ok := instancesBySource[source.ID]
		if !ok {
			continue
		}

		severityCounts, err := repository.CountFindingsBySeverity(ctx, s.pool, instance.ID, "")
		if err != nil {
			log.Printf("failed to count findings for source %s: %v", source.ID, err)
			http.Error(w, "failed to list sources", http.StatusInternalServerError)
			return
		}

		items = append(items, sourceRecordDTO{
			Source: sourceDTO{
				ID:                   source.ID,
				Kind:                 source.Kind,
				Provider:             source.Provider,
				Name:                 source.Name,
				Configured:           source.Configured,
				SetupState:           sourceSetupState(source.Kind, source.Configured),
				LastCheckStatus:      source.LastCheckStatus,
				LastCheckStartedAt:   source.LastCheckStartedAt,
				LastCheckCompletedAt: source.LastCheckCompletedAt,
				LastCheckError:       source.LastCheckError,
			},
			Database: sourceDatabaseDTO{
				ID:   instance.ID,
				Name: instance.Name,
				Host: instance.Host,
			},
			Instance: dashboardDatabaseInstanceDTO{
				ID:           instance.ID,
				DatabaseName: instance.Name,
				Host:         instance.Host,
				SourceID:     instance.SourceID,
				SourceKind:   instance.SourceKind,
				Provider:     instance.Provider,
				Status:       dashboardInstanceStatus(severityCounts),
			},
		})
	}

	writeJSON(w, http.StatusOK, listSourcesResponse{Sources: items})
}

func (s *Server) handleCreateSource(w http.ResponseWriter, r *http.Request) {
	var req createSourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Source.Kind == "" || req.Source.Provider == "" || req.Source.Name == "" {
		http.Error(w, "source.kind, source.provider, and source.name are required", http.StatusBadRequest)
		return
	}

	if req.Database.Name == "" || req.Database.Host == "" {
		http.Error(w, "database.name and database.host are required", http.StatusBadRequest)
		return
	}

	if _, ok := allowedSourceKinds[req.Source.Kind]; !ok {
		http.Error(w, "invalid source.kind", http.StatusBadRequest)
		return
	}

	if _, ok := allowedSourceProviders[req.Source.Provider]; !ok {
		http.Error(w, "invalid source.provider", http.StatusBadRequest)
		return
	}

	connectionURI := strings.TrimSpace(req.Connection.URI)
	if connectionURI == "" {
		http.Error(w, "connection.uri is required", http.StatusBadRequest)
		return
	}
	if _, err := pgxpool.ParseConfig(connectionURI); err != nil {
		http.Error(w, "invalid connection.uri", http.StatusBadRequest)
		return
	}

	if req.Source.ID == "" {
		req.Source.ID = generatedResourceID("src", req.Source.Name)
	}

	if req.Database.ID == "" {
		req.Database.ID = generatedResourceID("db", req.Database.Name)
	}

	if err := repository.UpsertSource(r.Context(), s.pool, req.Source.ID, req.Source.Kind, req.Source.Provider, req.Source.Name, ""); err != nil {
		http.Error(w, "failed to create source", http.StatusInternalServerError)
		return
	}
	connectionURIEncrypted, err := s.connectionProtector.Encrypt(connectionURI)
	if err != nil {
		log.Printf("failed to encrypt source connection URI for %s: %v", req.Source.ID, err)
		http.Error(w, "failed to store source connection", http.StatusInternalServerError)
		return
	}
	if err := repository.UpsertSourceConnectionProfile(r.Context(), s.pool, req.Source.ID, connectionURIEncrypted); err != nil {
		http.Error(w, "failed to store source connection", http.StatusInternalServerError)
		return
	}

	if err := storage.UpsertDatabaseInstance(r.Context(), s.pool, req.Database.ID, req.Source.ID, "", req.Database.Name, req.Database.Host); err != nil {
		http.Error(w, "failed to create database instance", http.StatusInternalServerError)
		return
	}

	instance, err := repository.GetDatabaseInstance(r.Context(), s.pool, req.Database.ID)
	if err != nil || instance == nil {
		http.Error(w, "failed to load created source", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, sourceRecordDTO{
		Source: sourceDTO{
			ID:              req.Source.ID,
			Kind:            req.Source.Kind,
			Provider:        req.Source.Provider,
			Name:            req.Source.Name,
			Configured:      true,
			SetupState:      sourceSetupState(req.Source.Kind, true),
			LastCheckStatus: "not_run",
		},
		Database: sourceDatabaseDTO{
			ID:   instance.ID,
			Name: instance.Name,
			Host: instance.Host,
		},
		Instance: dashboardDatabaseInstanceDTO{
			ID:           instance.ID,
			DatabaseName: instance.Name,
			Host:         instance.Host,
			SourceID:     instance.SourceID,
			SourceKind:   instance.SourceKind,
			Provider:     instance.Provider,
			Status:       "unknown",
		},
	})
}

func (s *Server) handleRunSourceCheckup(w http.ResponseWriter, r *http.Request) {
	sourceID := r.PathValue("id")
	if sourceID == "" {
		http.Error(w, "source id is required", http.StatusBadRequest)
		return
	}

	service := checkup.NewService(s.pool, s.connectionProtector)
	result, err := service.RunForSource(r.Context(), sourceID)
	if err != nil {
		log.Printf("failed to run source checkup for %s: %v", sourceID, err)
		http.Error(w, "failed to run source checkup", http.StatusInternalServerError)
		return
	}

	findings := make([]dashboardFindingDTO, len(result.Findings))
	for i, finding := range result.Findings {
		findings[i] = dashboardFindingDTO{
			Severity:            finding.Severity,
			Category:            finding.Category,
			Title:               finding.Title,
			Message:             finding.Message,
			Recommendation:      finding.Recommendation,
			Status:              "open",
			ProblemSummary:      finding.ProblemSummary,
			EvidenceSummary:     finding.EvidenceSummary,
			ImpactSummary:       finding.ImpactSummary,
			SuggestedAction:     finding.SuggestedAction,
			ConfidenceLabel:     finding.ConfidenceLabel,
			ConfidenceScore:     finding.ConfidenceScore,
			BaselineValue:       finding.BaselineValue,
			BaselineLabel:       finding.BaselineLabel,
			ChangeSummary:       finding.ChangeSummary,
			VerificationHint:    finding.VerificationHint,
			VerificationStatus:  finding.VerificationStatus,
			VerificationSummary: finding.VerificationSummary,
			RuleKey:             finding.RuleKey,
			ResourceType:        finding.ResourceType,
			ResourceName:        finding.ResourceName,
			CurrentValue:        finding.CurrentValue,
			ThresholdValue:      finding.ThresholdValue,
			DetectedAt:          finding.DetectedAt,
			FirstSeenAt:         finding.DetectedAt,
			LastSeenAt:          finding.DetectedAt,
		}
	}

	writeJSON(w, http.StatusOK, runSourceCheckupResponse{
		Status:             "succeeded",
		DatabaseInstanceID: result.DatabaseInstanceID,
		DetectedAt:         result.CollectedAt.Format(time.RFC3339Nano),
		Warnings:           result.Warnings,
		Findings:           findings,
	})
}

func sourceSetupState(kind string, configured bool) string {
	if configured {
		return "ready"
	}
	return "credentials_required"
}

func generatedResourceID(prefix, value string) string {
	base := slugify(value)
	if base == "" {
		base = prefix
	}
	return fmt.Sprintf("%s-%s-%s", prefix, base, randomHex(3))
}

func slugify(value string) string {
	var b strings.Builder
	lastDash := false

	for _, r := range strings.ToLower(strings.TrimSpace(value)) {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
			lastDash = false
		case !lastDash:
			b.WriteByte('-')
			lastDash = true
		}
	}

	return strings.Trim(b.String(), "-")
}

func randomHex(size int) string {
	bytes := make([]byte, size)
	if _, err := rand.Read(bytes); err != nil {
		return "local"
	}
	return hex.EncodeToString(bytes)
}
