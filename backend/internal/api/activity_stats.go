package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/okumujustine/postgresome/backend/internal/storage/repository"
)

type ingestActivityStatsRequest struct {
	AgentID            string                  `json:"agent_id"`
	DatabaseInstanceID string                  `json:"database_instance_id"`
	CollectedAt        time.Time               `json:"collected_at"`
	Activities         []ingestActivityStatRow `json:"activities"`
}

type ingestActivityStatRow struct {
	DatabaseName string `json:"database_name"`

	ProcessID int `json:"process_id"`

	UserName        string `json:"user_name"`
	ApplicationName string `json:"application_name"`
	ClientAddress   string `json:"client_address"`

	State string `json:"state"`
	Query string `json:"query"`

	WaitEventType string `json:"wait_event_type"`
	WaitEvent     string `json:"wait_event"`

	BackendStartedAt time.Time  `json:"backend_started_at"`
	QueryStartedAt   *time.Time `json:"query_started_at"`
	StateChangedAt   *time.Time `json:"state_changed_at"`
}

func (s *Server) handleIngestActivityStats(w http.ResponseWriter, r *http.Request) {
	var req ingestActivityStatsRequest
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

	activities := make([]repository.ActivityStatRow, len(req.Activities))
	for i, activity := range req.Activities {
		activities[i] = repository.ActivityStatRow{
			DatabaseName:     activity.DatabaseName,
			ProcessID:        activity.ProcessID,
			UserName:         activity.UserName,
			ApplicationName:  activity.ApplicationName,
			ClientAddress:    activity.ClientAddress,
			State:            activity.State,
			Query:            activity.Query,
			WaitEventType:    activity.WaitEventType,
			WaitEvent:        activity.WaitEvent,
			BackendStartedAt: activity.BackendStartedAt,
			QueryStartedAt:   activity.QueryStartedAt,
			StateChangedAt:   activity.StateChangedAt,
		}
	}

	if err := repository.ReplaceActivityStats(r.Context(), s.pool, req.DatabaseInstanceID, req.AgentID, collectedAt, activities); err != nil {
		log.Printf("failed to store activity stats: %v", err)
		http.Error(w, "failed to store activity stats", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status": "accepted",
		"stored": len(activities),
	})
}
