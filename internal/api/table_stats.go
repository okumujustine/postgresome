package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/okumujustine/postgresome/internal/storage/repository"
)

type ingestTableStatsRequest struct {
	AgentID            string            `json:"agent_id"`
	DatabaseInstanceID string            `json:"database_instance_id"`
	CollectedAt        time.Time         `json:"collected_at"`
	Tables             []ingestTableStat `json:"tables"`
}

type ingestTableStat struct {
	SchemaName string `json:"schema_name"`
	TableName  string `json:"table_name"`

	LiveRows int64 `json:"live_rows"`
	DeadRows int64 `json:"dead_rows"`

	SequentialScans    int64 `json:"sequential_scans"`
	SequentialRowsRead int64 `json:"sequential_rows_read"`

	IndexScans       int64 `json:"index_scans"`
	IndexRowsFetched int64 `json:"index_rows_fetched"`

	RowsInserted int64 `json:"rows_inserted"`
	RowsUpdated  int64 `json:"rows_updated"`
	RowsDeleted  int64 `json:"rows_deleted"`

	LastVacuumAt      *time.Time `json:"last_vacuum_at"`
	LastAutoVacuumAt  *time.Time `json:"last_autovacuum_at"`
	LastAnalyzeAt     *time.Time `json:"last_analyze_at"`
	LastAutoAnalyzeAt *time.Time `json:"last_autoanalyze_at"`

	VacuumCount      int64 `json:"vacuum_count"`
	AutoVacuumCount  int64 `json:"autovacuum_count"`
	AnalyzeCount     int64 `json:"analyze_count"`
	AutoAnalyzeCount int64 `json:"autoanalyze_count"`
}

// handleIngestTableStats serves POST /api/tables/ingest, replacing the
// stored table statistics snapshot for a database instance.
func (s *Server) handleIngestTableStats(w http.ResponseWriter, r *http.Request) {
	var req ingestTableStatsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.AgentID == "" || req.DatabaseInstanceID == "" {
		http.Error(w, "agent_id and database_instance_id are required", http.StatusBadRequest)
		return
	}

	collectedAt := req.CollectedAt
	if collectedAt.IsZero() {
		collectedAt = time.Now()
	}

	tables := make([]repository.TableStatRow, len(req.Tables))
	for i, t := range req.Tables {
		tables[i] = repository.TableStatRow{
			SchemaName:         t.SchemaName,
			TableName:          t.TableName,
			LiveRows:           t.LiveRows,
			DeadRows:           t.DeadRows,
			SequentialScans:    t.SequentialScans,
			SequentialRowsRead: t.SequentialRowsRead,
			IndexScans:         t.IndexScans,
			IndexRowsFetched:   t.IndexRowsFetched,
			RowsInserted:       t.RowsInserted,
			RowsUpdated:        t.RowsUpdated,
			RowsDeleted:        t.RowsDeleted,
			LastVacuumAt:       t.LastVacuumAt,
			LastAutoVacuumAt:   t.LastAutoVacuumAt,
			LastAnalyzeAt:      t.LastAnalyzeAt,
			LastAutoAnalyzeAt:  t.LastAutoAnalyzeAt,
			VacuumCount:        t.VacuumCount,
			AutoVacuumCount:    t.AutoVacuumCount,
			AnalyzeCount:       t.AnalyzeCount,
			AutoAnalyzeCount:   t.AutoAnalyzeCount,
		}
	}

	if err := repository.ReplaceTableStats(r.Context(), s.pool, req.DatabaseInstanceID, req.AgentID, collectedAt, tables); err != nil {
		log.Printf("failed to store table stats: %v", err)
		http.Error(w, "failed to store table stats", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status": "accepted",
		"stored": len(tables),
	})
}

type tableStatsResponse struct {
	DatabaseInstance dashboardDatabaseInstanceDTO `json:"database_instance"`
	CollectedAt      *time.Time                   `json:"collected_at"`
	Tables           []tableStatDTO               `json:"tables"`
}

type tableStatDTO struct {
	SchemaName string `json:"schema_name"`
	TableName  string `json:"table_name"`

	LiveRows int64 `json:"live_rows"`
	DeadRows int64 `json:"dead_rows"`

	SequentialScans    int64 `json:"sequential_scans"`
	SequentialRowsRead int64 `json:"sequential_rows_read"`

	IndexScans       int64 `json:"index_scans"`
	IndexRowsFetched int64 `json:"index_rows_fetched"`

	RowsInserted int64 `json:"rows_inserted"`
	RowsUpdated  int64 `json:"rows_updated"`
	RowsDeleted  int64 `json:"rows_deleted"`

	LastVacuumAt      *time.Time `json:"last_vacuum_at"`
	LastAutoVacuumAt  *time.Time `json:"last_autovacuum_at"`
	LastAnalyzeAt     *time.Time `json:"last_analyze_at"`
	LastAutoAnalyzeAt *time.Time `json:"last_autoanalyze_at"`
}

// handleListTableStats serves GET /api/tables, returning the latest table
// statistics snapshot for a database instance.
func (s *Server) handleListTableStats(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	agentID := query.Get("agent_id")
	databaseInstanceID := query.Get("database_instance_id")
	if databaseInstanceID == "" {
		http.Error(w, "database_instance_id is required", http.StatusBadRequest)
		return
	}

	dbInstance, err := repository.GetDatabaseInstance(r.Context(), s.pool, databaseInstanceID)
	if err != nil {
		log.Printf("failed to load database instance: %v", err)
		http.Error(w, "failed to load table stats", http.StatusInternalServerError)
		return
	}
	if dbInstance == nil {
		http.Error(w, "database instance not found", http.StatusNotFound)
		return
	}

	if agentID == "" {
		agentID = dbInstance.AgentID
	}

	ctx := r.Context()

	snapshot, err := repository.ListTableStats(ctx, s.pool, databaseInstanceID)
	if err != nil {
		log.Printf("failed to list table stats: %v", err)
		http.Error(w, "failed to load table stats", http.StatusInternalServerError)
		return
	}

	severityCounts, err := repository.CountFindingsBySeverity(ctx, s.pool, databaseInstanceID, agentID)
	if err != nil {
		log.Printf("failed to count findings by severity: %v", err)
		http.Error(w, "failed to load table stats", http.StatusInternalServerError)
		return
	}

	instanceDTO := dashboardDatabaseInstanceDTO{
		ID:           dbInstance.ID,
		DatabaseName: dbInstance.Name,
		Host:         dbInstance.Host,
		Status:       dashboardInstanceStatus(severityCounts),
	}

	var collectedAt *time.Time
	if !snapshot.CollectedAt.IsZero() {
		collectedAt = &snapshot.CollectedAt
	}

	tables := make([]tableStatDTO, len(snapshot.Tables))
	for i, t := range snapshot.Tables {
		tables[i] = tableStatDTO{
			SchemaName:         t.SchemaName,
			TableName:          t.TableName,
			LiveRows:           t.LiveRows,
			DeadRows:           t.DeadRows,
			SequentialScans:    t.SequentialScans,
			SequentialRowsRead: t.SequentialRowsRead,
			IndexScans:         t.IndexScans,
			IndexRowsFetched:   t.IndexRowsFetched,
			RowsInserted:       t.RowsInserted,
			RowsUpdated:        t.RowsUpdated,
			RowsDeleted:        t.RowsDeleted,
			LastVacuumAt:       t.LastVacuumAt,
			LastAutoVacuumAt:   t.LastAutoVacuumAt,
			LastAnalyzeAt:      t.LastAnalyzeAt,
			LastAutoAnalyzeAt:  t.LastAutoAnalyzeAt,
		}
	}

	writeJSON(w, http.StatusOK, tableStatsResponse{
		DatabaseInstance: instanceDTO,
		CollectedAt:      collectedAt,
		Tables:           tables,
	})
}
