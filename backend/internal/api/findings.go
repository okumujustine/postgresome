package api

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/okumujustine/postgresome/backend/internal/diagnosis"
	"github.com/okumujustine/postgresome/backend/internal/storage/repository"
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
	RelatedQuery      *relatedQueryDTO             `json:"related_query,omitempty"`
	RelatedTable      *relatedTableDTO             `json:"related_table,omitempty"`
	AlertPayload      *diagnosis.AlertPayload      `json:"alert_payload,omitempty"`
}

type relatedQueryDTO struct {
	CollectedAt      time.Time `json:"collected_at"`
	QueryID          string    `json:"query_id"`
	DatabaseName     string    `json:"database_name"`
	UserName         string    `json:"user_name"`
	Query            string    `json:"query"`
	Calls            int64     `json:"calls"`
	TotalExecTimeMs  float64   `json:"total_exec_time_ms"`
	MeanExecTimeMs   float64   `json:"mean_exec_time_ms"`
	MinExecTimeMs    float64   `json:"min_exec_time_ms"`
	MaxExecTimeMs    float64   `json:"max_exec_time_ms"`
	RowsReturned     int64     `json:"rows_returned"`
	SharedBlocksRead int64     `json:"shared_blocks_read"`
	SharedBlocksHit  int64     `json:"shared_blocks_hit"`
}

type relatedTableDTO struct {
	CollectedAt        time.Time  `json:"collected_at"`
	SchemaName         string     `json:"schema_name"`
	TableName          string     `json:"table_name"`
	LiveRows           int64      `json:"live_rows"`
	DeadRows           int64      `json:"dead_rows"`
	SequentialScans    int64      `json:"sequential_scans"`
	SequentialRowsRead int64      `json:"sequential_rows_read"`
	IndexScans         int64      `json:"index_scans"`
	IndexRowsFetched   int64      `json:"index_rows_fetched"`
	RowsInserted       int64      `json:"rows_inserted"`
	RowsUpdated        int64      `json:"rows_updated"`
	RowsDeleted        int64      `json:"rows_deleted"`
	LastVacuumAt       *time.Time `json:"last_vacuum_at,omitempty"`
	LastAutovacuumAt   *time.Time `json:"last_autovacuum_at,omitempty"`
	LastAnalyzeAt      *time.Time `json:"last_analyze_at,omitempty"`
	LastAutoAnalyzeAt  *time.Time `json:"last_autoanalyze_at,omitempty"`
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

	relatedQuery, relatedTable, err := s.buildRelatedFindingObjects(ctx, *finding)
	if err != nil {
		log.Printf("failed to load related objects for finding %q: %v", id, err)
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
		RelatedQuery:      relatedQuery,
		RelatedTable:      relatedTable,
		AlertPayload:      alertPayloadPtr(diagnosis.BuildAlertPayload(*finding)),
	})
}

func (s *Server) buildRelatedFindingObjects(ctx context.Context, finding repository.RecentFinding) (*relatedQueryDTO, *relatedTableDTO, error) {
	var (
		queryID         string
		tableSchemaName string
		tableName       string
	)

	switch finding.ResourceType {
	case "query":
		queryID = finding.ResourceName
	case "table":
		tableSchemaName, tableName = splitTableFindingResource(finding.ResourceName)
	case "query_plan":
		parts := strings.SplitN(finding.ResourceName, ":", 2)
		if len(parts) > 0 {
			queryID = parts[0]
		}
		if len(parts) == 2 && strings.Contains(parts[1], ".") {
			tableSchemaName, tableName = splitTableFindingResource(parts[1])
		}
	}

	var relatedQuery *relatedQueryDTO
	if queryID != "" {
		queryStat, err := repository.GetQueryStatAtOrBefore(ctx, s.pool, finding.DatabaseInstanceID, queryID, finding.LastSeenAt)
		if err != nil {
			return nil, nil, err
		}
		if queryStat != nil {
			relatedQuery = &relatedQueryDTO{
				CollectedAt:      queryStat.CollectedAt,
				QueryID:          queryStat.Query.QueryID,
				DatabaseName:     queryStat.Query.DatabaseName,
				UserName:         queryStat.Query.UserName,
				Query:            queryStat.Query.Query,
				Calls:            queryStat.Query.Calls,
				TotalExecTimeMs:  queryStat.Query.TotalExecTimeMs,
				MeanExecTimeMs:   queryStat.Query.MeanExecTimeMs,
				MinExecTimeMs:    queryStat.Query.MinExecTimeMs,
				MaxExecTimeMs:    queryStat.Query.MaxExecTimeMs,
				RowsReturned:     queryStat.Query.RowsReturned,
				SharedBlocksRead: queryStat.Query.SharedBlocksRead,
				SharedBlocksHit:  queryStat.Query.SharedBlocksHit,
			}
		}
	}

	var relatedTable *relatedTableDTO
	if tableName != "" {
		tableStat, err := repository.GetTableStatAtOrBefore(ctx, s.pool, finding.DatabaseInstanceID, tableSchemaName, tableName, finding.LastSeenAt)
		if err != nil {
			return nil, nil, err
		}
		if tableStat != nil {
			relatedTable = &relatedTableDTO{
				CollectedAt:        tableStat.CollectedAt,
				SchemaName:         tableStat.Table.SchemaName,
				TableName:          tableStat.Table.TableName,
				LiveRows:           tableStat.Table.LiveRows,
				DeadRows:           tableStat.Table.DeadRows,
				SequentialScans:    tableStat.Table.SequentialScans,
				SequentialRowsRead: tableStat.Table.SequentialRowsRead,
				IndexScans:         tableStat.Table.IndexScans,
				IndexRowsFetched:   tableStat.Table.IndexRowsFetched,
				RowsInserted:       tableStat.Table.RowsInserted,
				RowsUpdated:        tableStat.Table.RowsUpdated,
				RowsDeleted:        tableStat.Table.RowsDeleted,
				LastVacuumAt:       tableStat.Table.LastVacuumAt,
				LastAutovacuumAt:   tableStat.Table.LastAutoVacuumAt,
				LastAnalyzeAt:      tableStat.Table.LastAnalyzeAt,
				LastAutoAnalyzeAt:  tableStat.Table.LastAutoAnalyzeAt,
			}
		}
	}

	return relatedQuery, relatedTable, nil
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
