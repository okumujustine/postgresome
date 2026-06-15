package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/okumujustine/postgresome/cloud/internal/storage"
	"github.com/okumujustine/postgresome/cloud/internal/storage/repository"
)

var allowedSourceKinds = map[string]struct{}{
	"direct":              {},
	"managed_integration": {},
}

var allowedSourceProviders = map[string]struct{}{
	"postgres": {},
	"supabase": {},
	"rds":      {},
	"cloudsql": {},
	"neon":     {},
}

type sourceDTO struct {
	ID       string `json:"id"`
	Kind     string `json:"kind"`
	Provider string `json:"provider"`
	Name     string `json:"name"`
}

type sourceDatabaseDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Host string `json:"host"`
}

type sourceRecordDTO struct {
	Source   sourceDTO                    `json:"source"`
	Database sourceDatabaseDTO            `json:"database"`
	Instance dashboardDatabaseInstanceDTO `json:"instance"`
}

type listSourcesResponse struct {
	Sources []sourceRecordDTO `json:"sources"`
}

type createSourceRequest struct {
	Source struct {
		ID       string `json:"id"`
		Kind     string `json:"kind"`
		Provider string `json:"provider"`
		Name     string `json:"name"`
	} `json:"source"`
	Database struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Host string `json:"host"`
	} `json:"database"`
}

func (s *Server) handleListSources(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	sources, err := repository.ListSources(ctx, s.pool)
	if err != nil {
		log.Printf("failed to list sources: %v", err)
		http.Error(w, "failed to list sources", http.StatusInternalServerError)
		return
	}

	instances, err := repository.ListDatabaseInstances(ctx, s.pool)
	if err != nil {
		log.Printf("failed to list database instances for sources: %v", err)
		http.Error(w, "failed to list sources", http.StatusInternalServerError)
		return
	}

	instancesBySource := make(map[string]repository.DatabaseInstance, len(instances))
	for _, instance := range instances {
		instancesBySource[instance.SourceID] = instance
	}

	items := make([]sourceRecordDTO, 0, len(sources))
	for _, source := range sources {
		instance, ok := instancesBySource[source.ID]
		if !ok {
			continue
		}

		severityCounts, err := repository.CountFindingsBySeverity(ctx, s.pool, instance.ID, "")
		if err != nil {
			log.Printf("failed to count findings for source %s: %v", source.ID, err)
			http.Error(w, "failed to list sources", http.StatusInternalServerError)
			return
		}

		items = append(items, sourceRecordDTO{
			Source: sourceDTO{
				ID:       source.ID,
				Kind:     source.Kind,
				Provider: source.Provider,
				Name:     source.Name,
			},
			Database: sourceDatabaseDTO{
				ID:   instance.ID,
				Name: instance.Name,
				Host: instance.Host,
			},
			Instance: dashboardDatabaseInstanceDTO{
				ID:           instance.ID,
				DatabaseName: instance.Name,
				Host:         instance.Host,
				SourceID:     instance.SourceID,
				SourceKind:   instance.SourceKind,
				Provider:     instance.Provider,
				Status:       dashboardInstanceStatus(severityCounts),
			},
		})
	}

	writeJSON(w, http.StatusOK, listSourcesResponse{Sources: items})
}

func (s *Server) handleCreateSource(w http.ResponseWriter, r *http.Request) {
	var req createSourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Source.ID == "" || req.Source.Kind == "" || req.Source.Provider == "" || req.Source.Name == "" {
		http.Error(w, "source.id, source.kind, source.provider, and source.name are required", http.StatusBadRequest)
		return
	}

	if req.Database.ID == "" || req.Database.Name == "" || req.Database.Host == "" {
		http.Error(w, "database.id, database.name, and database.host are required", http.StatusBadRequest)
		return
	}

	if _, ok := allowedSourceKinds[req.Source.Kind]; !ok {
		http.Error(w, "invalid source.kind", http.StatusBadRequest)
		return
	}

	if _, ok := allowedSourceProviders[req.Source.Provider]; !ok {
		http.Error(w, "invalid source.provider", http.StatusBadRequest)
		return
	}

	if err := repository.UpsertSource(r.Context(), s.pool, req.Source.ID, req.Source.Kind, req.Source.Provider, req.Source.Name, ""); err != nil {
		http.Error(w, "failed to create source", http.StatusInternalServerError)
		return
	}

	if err := storage.UpsertDatabaseInstance(r.Context(), s.pool, req.Database.ID, req.Source.ID, "", req.Database.Name, req.Database.Host); err != nil {
		http.Error(w, "failed to create database instance", http.StatusInternalServerError)
		return
	}

	instance, err := repository.GetDatabaseInstance(r.Context(), s.pool, req.Database.ID)
	if err != nil || instance == nil {
		http.Error(w, "failed to load created source", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, sourceRecordDTO{
		Source: sourceDTO{
			ID:       req.Source.ID,
			Kind:     req.Source.Kind,
			Provider: req.Source.Provider,
			Name:     req.Source.Name,
		},
		Database: sourceDatabaseDTO{
			ID:   instance.ID,
			Name: instance.Name,
			Host: instance.Host,
		},
		Instance: dashboardDatabaseInstanceDTO{
			ID:           instance.ID,
			DatabaseName: instance.Name,
			Host:         instance.Host,
			SourceID:     instance.SourceID,
			SourceKind:   instance.SourceKind,
			Provider:     instance.Provider,
			Status:       "unknown",
		},
	})
}
