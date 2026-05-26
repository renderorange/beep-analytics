package api

import (
	"log"
	"net/http"

	"github.com/adventurehound/beep/internal/db"
	"github.com/adventurehound/beep/internal/geoip"
)

type Server struct {
	db    *db.DB
	geoip *geoip.Lookup
	addr  string
	mux   *http.ServeMux
}

func NewServer(database *db.DB, geo *geoip.Lookup, addr string) *Server {
	s := &Server{
		db:    database,
		geoip: geo,
		addr:  addr,
		mux:   http.NewServeMux(),
	}
	s.registerRoutes()
	return s
}

func (s *Server) registerRoutes() {
	// Public endpoints
	s.mux.HandleFunc("POST /collect", s.handleCollect)
	s.mux.HandleFunc("GET /track.js", s.handleTrackJS)

	// API endpoints (auth required)
	s.mux.HandleFunc("POST /api/sites", s.requireAuth(s.handleAddSite))
	s.mux.HandleFunc("DELETE /api/sites/{domain}", s.requireAuth(s.handleRemoveSite))
	s.mux.HandleFunc("GET /api/sites", s.requireAuth(s.handleListSites))

	s.mux.HandleFunc("POST /api/ignore", s.requireAuth(s.handleAddIgnore))
	s.mux.HandleFunc("DELETE /api/ignore/{ip}", s.requireAuth(s.handleRemoveIgnore))
	s.mux.HandleFunc("GET /api/ignore", s.requireAuth(s.handleListIgnore))

	// Token generation allows bootstrap (no auth needed if no tokens exist)
	s.mux.HandleFunc("POST /api/tokens/generate", s.requireAuthUnlessNoTokens(s.handleGenerateToken))
	s.mux.HandleFunc("DELETE /api/tokens/{id}", s.requireAuth(s.handleRevokeToken))
	s.mux.HandleFunc("GET /api/tokens", s.requireAuth(s.handleListTokens))

	s.mux.HandleFunc("GET /api/stats", s.requireAuth(s.handleStats))
}

func (s *Server) ListenAndServe() error {
	log.Printf("beep listening on %s", s.addr)
	return http.ListenAndServe(s.addr, s.mux)
}
