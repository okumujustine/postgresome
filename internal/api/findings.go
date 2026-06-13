package api

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/okumujustine/postgresome/internal/storage/repository"
)

const (
	defaultFindingsLimit = 20
	maxFindingsLimit     = 100
)

type findingsListResponse struct {
	DatabaseInstance dashboardDatabaseInstanceDTO `json:"database_instance"`
	SeverityCounts   findingsSeverityCountsDTO    `json:"severity_counts"`
	Total            int                          `json:"total"`
	Limit            int                          `json:"limit"`
	Offset           int                          `json:"offset"`
	Findings         []dashboardFindingDTO        `json:"findings"`
}

type findingsSeverityCountsDTO struct {
	Critical int `json:"critical"`
	Warning  int `json:"warning"`
	Info     int `json:"info"`
}

// handleListFindings serves GET /api/findings, returning a paginated,
// filterable list of findings for a database instance along with severity
// counts over the selected time range.
func (s *Server) handleListFindings(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	rangeParam := query.Get("range")
	if rangeParam == "" {
		rangeParam = defaultMetricQueryRange
	}

	rangeDuration, ok := metricQueryRanges[rangeParam]
	if !ok {
		http.Error(w, "invalid range", http.StatusBadRequest)
		return
	}

	limit := defaultFindingsLimit
	if limitParam := query.Get("limit"); limitParam != "" {
		parsed, err := strconv.Atoi(limitParam)
		if err != nil || parsed <= 0 {
			http.Error(w, "invalid limit", http.StatusBadRequest)
			return
		}
		if parsed > maxFindingsLimit {
			parsed = maxFindingsLimit
		}
		limit = parsed
	}

	offset := 0
	if offsetParam := query.Get("offset"); offsetParam != "" {
		parsed, err := strconv.Atoi(offsetParam)
		if err != nil || parsed < 0 {
			http.Error(w, "invalid offset", http.StatusBadRequest)
			return
		}
		offset = parsed
	}

	agentID := query.Get("agent_id")
	databaseInstanceID := query.Get("database_instance_id")
	if databaseInstanceID == "" {
		http.Error(w, "database_instance_id is required", http.StatusBadRequest)
		return
	}

	dbInstance, err := repository.GetDatabaseInstance(r.Context(), s.pool, databaseInstanceID)
	if err != nil {
		log.Printf("failed to load database instance: %v", err)
		http.Error(w, "failed to load findings", http.StatusInternalServerError)
		return
	}
	if dbInstance == nil {
		http.Error(w, "database instance not found", http.StatusNotFound)
		return
	}

	if agentID == "" {
		agentID = dbInstance.AgentID
	}

	since := time.Now().Add(-rangeDuration)
	ctx := r.Context()

	severityCounts, err := repository.CountFindingsBySeverity(ctx, s.pool, databaseInstanceID, agentID, since)
	if err != nil {
		log.Printf("failed to count findings by severity: %v", err)
		http.Error(w, "failed to load findings", http.StatusInternalServerError)
		return
	}

	listParams := repository.ListFindingsParams{
		DatabaseInstanceID: databaseInstanceID,
		AgentID:            agentID,
		Severity:           query.Get("severity"),
		Category:           query.Get("category"),
		Since:              since,
		Limit:              limit,
		Offset:             offset,
	}

	findings, err := repository.ListFindings(ctx, s.pool, listParams)
	if err != nil {
		log.Printf("failed to list findings: %v", err)
		http.Error(w, "failed to load findings", http.StatusInternalServerError)
		return
	}

	total, err := repository.CountFindings(ctx, s.pool, listParams)
	if err != nil {
		log.Printf("failed to count findings: %v", err)
		http.Error(w, "failed to load findings", http.StatusInternalServerError)
		return
	}

	findingDTOs := make([]dashboardFindingDTO, len(findings))
	for i, f := range findings {
		findingDTOs[i] = dashboardFindingDTO{
			ID:             f.ID,
			Severity:       f.Severity,
			Category:       f.Category,
			Title:          f.Title,
			Message:        f.Message,
			Recommendation: f.Recommendation,
			DetectedAt:     f.DetectedAt,
		}
	}

	instanceDTO := dashboardDatabaseInstanceDTO{
		ID:           dbInstance.ID,
		DatabaseName: dbInstance.Name,
		Host:         dbInstance.Host,
		Status:       dashboardInstanceStatus(severityCounts),
	}

	writeJSON(w, http.StatusOK, findingsListResponse{
		DatabaseInstance: instanceDTO,
		SeverityCounts: findingsSeverityCountsDTO{
			Critical: severityCounts.Critical,
			Warning:  severityCounts.Warning,
			Info:     severityCounts.Info,
		},
		Total:    total,
		Limit:    limit,
		Offset:   offset,
		Findings: findingDTOs,
	})
}
