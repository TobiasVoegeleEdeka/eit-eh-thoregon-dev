package server

import (
	"bounceservice/internal/bouncer"
	"encoding/json"
	"net/http"
)

type Server struct {
	bouncer *bouncer.Bouncer
}

func New(b *bouncer.Bouncer) *Server {
	return &Server{bouncer: b}
}

func (s *Server) RegisterRoutes() {
	http.HandleFunc("/bounces", s.handleBounces)
	http.HandleFunc("/health", s.handleHealth)
}

func (s *Server) handleBounces(w http.ResponseWriter, r *http.Request) {
	s.bouncer.Mu.Lock()
	defer s.bouncer.Mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.bouncer.Bounces)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
