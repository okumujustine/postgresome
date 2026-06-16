package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/okumujustine/postgresome/backend/internal/metrics"
	"github.com/okumujustine/postgresome/backend/internal/storage/repository"
)

type ingestExplainPlansRequest struct {
	AgentID            string                  `json:"agent_id"`
	DatabaseInstanceID string                  `json:"database_instance_id"`
	CollectedAt        time.Time               `json:"collected_at"`
	Plans              []ingestExplainPlanItem `json:"plans"`
}

type ingestExplainPlanItem struct {
	QueryID string           `json:"query_id"`
	Query   string           `json:"query"`
	Root    metrics.PlanNode `json:"root"`
}

func (s *Server) handleIngestExplainPlans(w http.ResponseWriter, r *http.Request) {
	var req ingestExplainPlansRequest
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

	plans := make([]repository.ExplainPlanRow, len(req.Plans))
	for i, plan := range req.Plans {
		plans[i] = repository.ExplainPlanRow{
			QueryID: plan.QueryID,
			Query:   plan.Query,
			Root:    plan.Root,
		}
	}

	if err := repository.ReplaceExplainPlans(r.Context(), s.pool, req.DatabaseInstanceID, req.AgentID, collectedAt, plans); err != nil {
		log.Printf("failed to store explain plans: %v", err)
		http.Error(w, "failed to store explain plans", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status": "accepted",
		"stored": len(plans),
	})
}
