package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const shutdownTimeout = 5 * time.Second

// Server is the Postgresome HTTP API server. It runs independently of the
// agent and will eventually serve collected metrics and findings to a
// frontend dashboard.
type Server struct {
	addr       string
	pool       *pgxpool.Pool
	httpServer *http.Server
}

func NewServer(addr string, pool *pgxpool.Pool) *Server {
	s := &Server{
		addr: addr,
		pool: pool,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("GET /version", handleVersion)
	mux.HandleFunc("POST /api/agents/register", s.handleRegisterAgent)
	mux.HandleFunc("POST /api/metrics/ingest", s.handleIngestMetrics)
	mux.HandleFunc("GET /api/metrics/query", s.handleQueryMetrics)
	mux.HandleFunc("POST /api/findings/ingest", s.handleIngestFindings)
	mux.HandleFunc("GET /api/dashboard/overview", s.handleDashboardOverview)

	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: withCORS(mux),
	}

	return s
}

// withCORS allows browser-based frontends (e.g. the dashboard dev server) to
// call this read-only, unauthenticated API from a different origin.
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Start runs the HTTP server until ctx is cancelled, then shuts it down
// gracefully.
func (s *Server) Start(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}

		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("failed to shut down API server: %w", err)
		}

		return ctx.Err()

	case err := <-errCh:
		return err
	}
}
