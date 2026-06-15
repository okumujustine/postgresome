package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/okumujustine/postgresome/cloud/internal/storage/repository"
)

type ingestQueryStatsRequest struct {
	AgentID            string            `json:"agent_id"`
	DatabaseInstanceID string            `json:"database_instance_id"`
	CollectedAt        time.Time         `json:"collected_at"`
	Queries            []ingestQueryStat `json:"queries"`
}

type ingestQueryStat struct {
	QueryID      string `json:"query_id"`
	DatabaseName string `json:"database_name"`
	UserName     string `json:"user_name"`
	Query        string `json:"query"`

	Calls int64 `json:"calls"`

	TotalExecTimeMs float64 `json:"total_exec_time_ms"`
	MeanExecTimeMs  float64 `json:"mean_exec_time_ms"`
	MinExecTimeMs   float64 `json:"min_exec_time_ms"`
	MaxExecTimeMs   float64 `json:"max_exec_time_ms"`

	RowsReturned int64 `json:"rows_returned"`

	SharedBlocksHit     int64 `json:"shared_blocks_hit"`
	SharedBlocksRead    int64 `json:"shared_blocks_read"`
	SharedBlocksDirtied int64 `json:"shared_blocks_dirtied"`
	SharedBlocksWritten int64 `json:"shared_blocks_written"`

	TempBlocksRead    int64 `json:"temp_blocks_read"`
	TempBlocksWritten int64 `json:"temp_blocks_written"`
}

// handleIngestQueryStats serves POST /api/queries/ingest, replacing the
// stored query statistics snapshot for a database instance.
func (s *Server) handleIngestQueryStats(w http.ResponseWriter, r *http.Request) {
	var req ingestQueryStatsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.DatabaseInstanceID == "" {
		http.Error(w, "database_instance_id is required", http.StatusBadRequest)
		return
	}

	collectedAt := req.CollectedAt
	if collectedAt.IsZero() {
		collectedAt = time.Now()
	}

	queries := make([]repository.QueryStatRow, len(req.Queries))
	for i, q := range req.Queries {
		queries[i] = repository.QueryStatRow{
			QueryID:             q.QueryID,
			DatabaseName:        q.DatabaseName,
			UserName:            q.UserName,
			Query:               q.Query,
			Calls:               q.Calls,
			TotalExecTimeMs:     q.TotalExecTimeMs,
			MeanExecTimeMs:      q.MeanExecTimeMs,
			MinExecTimeMs:       q.MinExecTimeMs,
			MaxExecTimeMs:       q.MaxExecTimeMs,
			RowsReturned:        q.RowsReturned,
			SharedBlocksHit:     q.SharedBlocksHit,
			SharedBlocksRead:    q.SharedBlocksRead,
			SharedBlocksDirtied: q.SharedBlocksDirtied,
			SharedBlocksWritten: q.SharedBlocksWritten,
			TempBlocksRead:      q.TempBlocksRead,
			TempBlocksWritten:   q.TempBlocksWritten,
		}
	}

	if err := repository.ReplaceQueryStats(r.Context(), s.pool, req.DatabaseInstanceID, req.AgentID, collectedAt, queries); err != nil {
		log.Printf("failed to store query stats: %v", err)
		http.Error(w, "failed to store query stats", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status": "accepted",
		"stored": len(queries),
	})
}

type queryStatsResponse struct {
	DatabaseInstance dashboardDatabaseInstanceDTO `json:"database_instance"`
	CollectedAt      *time.Time                   `json:"collected_at"`
	Queries          []queryStatDTO               `json:"queries"`
}

type queryStatDTO struct {
	QueryID      string `json:"query_id"`
	DatabaseName string `json:"database_name"`
	UserName     string `json:"user_name"`
	Query        string `json:"query"`

	Calls int64 `json:"calls"`

	TotalExecTimeMs float64 `json:"total_exec_time_ms"`
	MeanExecTimeMs  float64 `json:"mean_exec_time_ms"`
	MinExecTimeMs   float64 `json:"min_exec_time_ms"`
	MaxExecTimeMs   float64 `json:"max_exec_time_ms"`

	RowsReturned int64 `json:"rows_returned"`

	SharedBlocksRead int64 `json:"shared_blocks_read"`
	SharedBlocksHit  int64 `json:"shared_blocks_hit"`
}

// handleListQueryStats serves GET /api/queries, returning the latest query
// statistics snapshot for a database instance.
func (s *Server) handleListQueryStats(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	databaseInstanceID := query.Get("database_instance_id")
	if databaseInstanceID == "" {
		http.Error(w, "database_instance_id is required", http.StatusBadRequest)
		return
	}

	dbInstance, err := repository.GetDatabaseInstance(r.Context(), s.pool, databaseInstanceID)
	if err != nil {
		log.Printf("failed to load database instance: %v", err)
		http.Error(w, "failed to load query stats", http.StatusInternalServerError)
		return
	}
	if dbInstance == nil {
		http.Error(w, "database instance not found", http.StatusNotFound)
		return
	}

	ctx := r.Context()

	snapshot, err := repository.ListQueryStats(ctx, s.pool, databaseInstanceID)
	if err != nil {
		log.Printf("failed to list query stats: %v", err)
		http.Error(w, "failed to load query stats", http.StatusInternalServerError)
		return
	}

	severityCounts, err := repository.CountFindingsBySeverity(ctx, s.pool, databaseInstanceID, "")
	if err != nil {
		log.Printf("failed to count findings by severity: %v", err)
		http.Error(w, "failed to load query stats", http.StatusInternalServerError)
		return
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

	var collectedAt *time.Time
	if !snapshot.CollectedAt.IsZero() {
		collectedAt = &snapshot.CollectedAt
	}

	queries := make([]queryStatDTO, len(snapshot.Queries))
	for i, q := range snapshot.Queries {
		queries[i] = queryStatDTO{
			QueryID:         q.QueryID,
			DatabaseName:    q.DatabaseName,
			UserName:        q.UserName,
			Query:           q.Query,
			Calls:           q.Calls,
			TotalExecTimeMs: q.TotalExecTimeMs,
			MeanExecTimeMs:  q.MeanExecTimeMs,
			MinExecTimeMs:   q.MinExecTimeMs,
			MaxExecTimeMs:   q.MaxExecTimeMs,
			RowsReturned:    q.RowsReturned,

			SharedBlocksRead: q.SharedBlocksRead,
			SharedBlocksHit:  q.SharedBlocksHit,
		}
	}

	writeJSON(w, http.StatusOK, queryStatsResponse{
		DatabaseInstance: instanceDTO,
		CollectedAt:      collectedAt,
		Queries:          queries,
	})
}
