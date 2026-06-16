package api

import (
	"log"
	"net/http"

	"github.com/okumujustine/postgresome/backend/internal/storage/repository"
)

type listDatabaseInstancesResponse struct {
	DatabaseInstances []dashboardDatabaseInstanceDTO `json:"database_instances"`
}

// handleListDatabaseInstances serves GET /api/database-instances, returning
// every registered database instance with a coarse health status, for use in
// a database selector UI.
func (s *Server) handleListDatabaseInstances(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	instances, err := repository.ListDatabaseInstances(ctx, s.pool)
	if err != nil {
		log.Printf("failed to list database instances: %v", err)
		http.Error(w, "failed to list database instances", http.StatusInternalServerError)
		return
	}

	dtos := make([]dashboardDatabaseInstanceDTO, len(instances))
	for i, instance := range instances {
		severityCounts, err := repository.CountFindingsBySeverity(ctx, s.pool, instance.ID, "")
		if err != nil {
			log.Printf("failed to count findings for %s: %v", instance.ID, err)
			http.Error(w, "failed to list database instances", http.StatusInternalServerError)
			return
		}

		dtos[i] = dashboardDatabaseInstanceDTO{
			ID:           instance.ID,
			DatabaseName: instance.Name,
			Host:         instance.Host,
			SourceID:     instance.SourceID,
			SourceKind:   instance.SourceKind,
			Provider:     instance.Provider,
			Status:       dashboardInstanceStatus(severityCounts),
		}
	}

	writeJSON(w, http.StatusOK, listDatabaseInstancesResponse{DatabaseInstances: dtos})
}
