// Package v1 provides the REST API skeleton for dbx.
package v1

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/itunified-io/dbx/pkg/pipeline"
)

// Server is the dbx REST API server.
type Server struct {
	pipeline *pipeline.Pipeline
	mux      *http.ServeMux
}

// NewServer creates a REST API server.
func NewServer(p *pipeline.Pipeline) *Server {
	s := &Server{
		pipeline: p,
		mux:      http.NewServeMux(),
	}
	s.routes()
	return s
}

// Handler returns the HTTP handler.
func (s *Server) Handler() http.Handler {
	return s.mux
}

// ListenAndServe starts the REST API server.
func (s *Server) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, s.mux)
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /health", s.handleHealth)
	s.mux.HandleFunc("GET /api/v1/version", s.handleVersion)
	s.mux.HandleFunc("GET /api/v1/targets", s.handleTargets)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleVersion(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"version": "dev"})
}

func (s *Server) handleTargets(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"message": "target list not yet implemented"})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		fmt.Fprintf(w, `{"error":%q}`, err.Error())
	}
}
