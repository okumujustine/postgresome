package api

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/okumujustine/postgresome/cloud/internal/diagnosis"
	"github.com/okumujustine/postgresome/cloud/internal/storage/repository"
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

type findingDetailResponse struct {
	DatabaseInstance  dashboardDatabaseInstanceDTO `json:"database_instance"`
	Finding           dashboardFindingDTO          `json:"finding"`
	Evidence          []findingEvidenceItemDTO     `json:"evidence,omitempty"`
	HistoricalContext *findingHistoricalContextDTO `json:"historical_context,omitempty"`
	EvidencePoints    []findingEvidencePointDTO    `json:"evidence_points,omitempty"`
	AlertPayload      *diagnosis.AlertPayload      `json:"alert_payload,omitempty"`
}

type findingEvidenceItemDTO struct {
	ID string `json:"id"`

	ObservedAt time.Time `json:"observed_at"`

	EvidenceType string `json:"evidence_type"`
	Role         string `json:"role"`
	Label        string `json:"label"`
	Summary      string `json:"summary"`
	MetricKey    string `json:"metric_key"`
	ReferenceID  string `json:"reference_id"`

	CurrentValue    float64        `json:"current_value"`
	BaselineValue   float64        `json:"baseline_value"`
	ChangePercent   float64        `json:"change_percent"`
	ConfidenceScore float64        `json:"confidence_score"`
	Metadata        map[string]any `json:"metadata,omitempty"`
}

type findingHistoricalContextDTO struct {
	CurrentValue  float64 `json:"current_value"`
	PreviousValue float64 `json:"previous_value"`
	BaselineValue float64 `json:"baseline_value"`
	ChangePercent float64 `json:"change_percent"`
	TrendWindow   string  `json:"trend_window"`
	BaselineLabel string  `json:"baseline_label"`
}

type findingEvidencePointDTO struct {
	Time   time.Time `json:"time"`
	Series string    `json:"series"`
	Value  float64   `json:"value"`
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

	status := query.Get("status")
	switch status {
	case "":
		status = "open"
	case "all":
		status = ""
	case "open", "resolved":
		// pass through as-is
	default:
		http.Error(w, "invalid status", http.StatusBadRequest)
		return
	}

	since := time.Now().Add(-rangeDuration)
	ctx := r.Context()

	severityCounts, err := repository.CountFindingsBySeverity(ctx, s.pool, databaseInstanceID, "")
	if err != nil {
		log.Printf("failed to count findings by severity: %v", err)
		http.Error(w, "failed to load findings", http.StatusInternalServerError)
		return
	}

	listParams := repository.ListFindingsParams{
		DatabaseInstanceID: databaseInstanceID,
		AgentID:            "",
		Severity:           query.Get("severity"),
		Category:           query.Get("category"),
		Status:             status,
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
		findingDTOs[i] = toDashboardFindingDTO(f)
	}

	instanceDTO := dashboardDatabaseInstanceDTO{
		ID:           dbInstance.ID,
		DatabaseName: dbInstance.Name,
		Host:         dbInstance.Host,
		SourceID:     dbInstance.SourceID,
		SourceKind:   dbInstance.SourceKind,
		Provider:     dbInstance.Provider,
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

// handleGetFinding serves GET /api/findings/{id}, returning a single issue
// with diagnosis-oriented explanation fields for the detail page.
func (s *Server) handleGetFinding(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "finding id is required", http.StatusBadRequest)
		return
	}

	query := r.URL.Query()
	databaseInstanceID := query.Get("database_instance_id")

	ctx := r.Context()
	finding, err := repository.GetFindingByID(ctx, s.pool, id, databaseInstanceID, "")
	if err != nil {
		log.Printf("failed to load finding %q: %v", id, err)
		http.Error(w, "failed to load finding", http.StatusInternalServerError)
		return
	}
	if finding == nil {
		http.Error(w, "finding not found", http.StatusNotFound)
		return
	}

	if databaseInstanceID == "" {
		databaseInstanceID = finding.DatabaseInstanceID
	}
	if databaseInstanceID == "" {
		http.Error(w, "finding is not associated with a database instance", http.StatusNotFound)
		return
	}

	dbInstance, err := repository.GetDatabaseInstance(ctx, s.pool, databaseInstanceID)
	if err != nil {
		log.Printf("failed to load database instance for finding %q: %v", id, err)
		http.Error(w, "failed to load finding", http.StatusInternalServerError)
		return
	}
	if dbInstance == nil {
		http.Error(w, "database instance not found", http.StatusNotFound)
		return
	}

	severityCounts, err := repository.CountFindingsBySeverity(ctx, s.pool, dbInstance.ID, "")
	if err != nil {
		log.Printf("failed to count findings by severity for finding %q: %v", id, err)
		http.Error(w, "failed to load finding", http.StatusInternalServerError)
		return
	}

	historicalContext, evidencePoints, err := s.buildFindingEvidence(ctx, *finding)
	if err != nil {
		log.Printf("failed to load historical evidence for finding %q: %v", id, err)
		http.Error(w, "failed to load finding", http.StatusInternalServerError)
		return
	}

	evidence, err := repository.ListFindingEvidence(ctx, s.pool, id)
	if err != nil {
		log.Printf("failed to load linked evidence for finding %q: %v", id, err)
		http.Error(w, "failed to load finding", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, findingDetailResponse{
		DatabaseInstance: dashboardDatabaseInstanceDTO{
			ID:           dbInstance.ID,
			DatabaseName: dbInstance.Name,
			Host:         dbInstance.Host,
			SourceID:     dbInstance.SourceID,
			SourceKind:   dbInstance.SourceKind,
			Provider:     dbInstance.Provider,
			Status:       dashboardInstanceStatus(severityCounts),
		},
		Finding:           toDashboardFindingDTO(*finding),
		Evidence:          toFindingEvidenceDTOs(evidence),
		HistoricalContext: historicalContext,
		EvidencePoints:    evidencePoints,
		AlertPayload:      alertPayloadPtr(diagnosis.BuildAlertPayload(*finding)),
	})
}

func alertPayloadPtr(payload diagnosis.AlertPayload) *diagnosis.AlertPayload {
	return &payload
}

func toFindingEvidenceDTOs(items []repository.FindingEvidenceRow) []findingEvidenceItemDTO {
	if len(items) == 0 {
		return nil
	}

	result := make([]findingEvidenceItemDTO, len(items))
	for i, item := range items {
		result[i] = findingEvidenceItemDTO{
			ID:              item.ID,
			ObservedAt:      item.ObservedAt,
			EvidenceType:    item.EvidenceType,
			Role:            item.Role,
			Label:           item.Label,
			Summary:         item.Summary,
			MetricKey:       item.MetricKey,
			ReferenceID:     item.ReferenceID,
			CurrentValue:    item.CurrentValue,
			BaselineValue:   item.BaselineValue,
			ChangePercent:   item.ChangePercent,
			ConfidenceScore: item.ConfidenceScore,
			Metadata:        item.Metadata,
		}
	}

	return result
}
