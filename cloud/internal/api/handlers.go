package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/okumujustine/postgresome/cloud/internal/analysis"
	"github.com/okumujustine/postgresome/cloud/internal/storage/repository"
	"github.com/okumujustine/postgresome/shared/metrics"
)

func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func handleVersion(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"name":    "postgresome",
		"version": "dev",
	})
}

type ingestMetricsRequest struct {
	AgentID            string              `json:"agent_id"`
	DatabaseInstanceID string              `json:"database_instance_id"`
	Metrics            []ingestMetricPoint `json:"metrics"`
}

type ingestMetricPoint struct {
	Key         string            `json:"key"`
	Label       string            `json:"label"`
	Value       float64           `json:"value"`
	Unit        string            `json:"unit"`
	Category    string            `json:"category"`
	CollectedAt time.Time         `json:"collected_at"`
	Dimensions  map[string]string `json:"dimensions"`
}

func (s *Server) handleIngestMetrics(w http.ResponseWriter, r *http.Request) {
	var req ingestMetricsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.DatabaseInstanceID == "" || len(req.Metrics) == 0 {
		http.Error(w, "database_instance_id and metrics are required", http.StatusBadRequest)
		return
	}

	points := make([]metrics.MetricPoint, len(req.Metrics))

	for i, m := range req.Metrics {
		dimensions := m.Dimensions
		if dimensions == nil {
			dimensions = make(map[string]string)
		}

		if req.AgentID != "" {
			if _, ok := dimensions["agent_id"]; !ok {
				dimensions["agent_id"] = req.AgentID
			}
		}

		if _, ok := dimensions["database_instance_id"]; !ok {
			dimensions["database_instance_id"] = req.DatabaseInstanceID
		}

		points[i] = metrics.MetricPoint{
			Key:         m.Key,
			Label:       m.Label,
			Value:       m.Value,
			Unit:        m.Unit,
			Category:    m.Category,
			CollectedAt: m.CollectedAt,
			Dimensions:  dimensions,
		}
	}

	if err := repository.InsertMetricPoints(r.Context(), s.pool, points); err != nil {
		http.Error(w, "failed to store metrics", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status":   "accepted",
		"inserted": len(points),
	})
}

type ingestFindingItem struct {
	Severity       string `json:"severity"`
	Category       string `json:"category"`
	Title          string `json:"title"`
	Message        string `json:"message"`
	Recommendation string `json:"recommendation"`

	RuleKey      string `json:"rule_key"`
	ResourceType string `json:"resource_type"`
	ResourceName string `json:"resource_name"`

	CurrentValue   float64 `json:"current_value"`
	ThresholdValue float64 `json:"threshold_value"`
}

type ingestFindingsRequest struct {
	AgentID            string              `json:"agent_id"`
	DatabaseInstanceID string              `json:"database_instance_id"`
	Findings           []ingestFindingItem `json:"findings"`
}

const findingResolutionGracePeriod = 90 * time.Second

func (s *Server) handleIngestFindings(w http.ResponseWriter, r *http.Request) {
	var req ingestFindingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.DatabaseInstanceID == "" {
		http.Error(w, "database_instance_id is required", http.StatusBadRequest)
		return
	}

	detectedAt := time.Now()

	if len(req.Findings) > 0 {
		findings := make([]analysis.Finding, len(req.Findings))
		for i, f := range req.Findings {
			findings[i] = analysis.Finding{
				DetectedAt:         detectedAt,
				DatabaseInstanceID: req.DatabaseInstanceID,
				AgentID:            req.AgentID,
				Severity:           f.Severity,
				Category:           f.Category,
				Title:              f.Title,
				Message:            f.Message,
				Recommendation:     f.Recommendation,
				RuleKey:            f.RuleKey,
				ResourceType:       f.ResourceType,
				ResourceName:       f.ResourceName,
				CurrentValue:       f.CurrentValue,
				ThresholdValue:     f.ThresholdValue,
			}
		}

		if err := repository.UpsertFindings(r.Context(), s.pool, findings); err != nil {
			http.Error(w, "failed to store findings", http.StatusInternalServerError)
			return
		}
	}

	if err := repository.ResolveStaleFindings(r.Context(), s.pool, req.DatabaseInstanceID, detectedAt, findingResolutionGracePeriod); err != nil {
		http.Error(w, "failed to resolve stale findings", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status":   "accepted",
		"inserted": len(req.Findings),
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
