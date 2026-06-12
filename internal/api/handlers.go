package api

import (
	"encoding/json"
	"net/http"

	"github.com/okumujustine/postgresome/internal/storage"
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

type registerAgentRequest struct {
	Agent struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Environment string `json:"environment"`
	} `json:"agent"`
	Database struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Host    string `json:"host"`
		Version string `json:"version"`
	} `json:"database"`
}

func (s *Server) handleRegisterAgent(w http.ResponseWriter, r *http.Request) {
	var req registerAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Agent.ID == "" || req.Agent.Name == "" || req.Database.ID == "" || req.Database.Name == "" {
		http.Error(w, "agent.id, agent.name, database.id, and database.name are required", http.StatusBadRequest)
		return
	}

	if err := storage.UpsertAgent(r.Context(), s.pool, req.Agent.ID, req.Agent.Name, req.Agent.Environment); err != nil {
		http.Error(w, "failed to register agent", http.StatusInternalServerError)
		return
	}

	if err := storage.UpsertDatabaseInstance(r.Context(), s.pool, req.Database.ID, req.Agent.ID, req.Database.Name, req.Database.Host); err != nil {
		http.Error(w, "failed to register database instance", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"agent_id":             req.Agent.ID,
		"database_instance_id": req.Database.ID,
		"status":               "registered",
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}
