package api

import (
	"encoding/json"
	"net/http"
)

type SiteResponse struct {
	ID     int64  `json:"id"`
	Domain string `json:"domain"`
}

func (s *Server) handleAddSite(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Domain string `json:"domain"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.Domain == "" {
		http.Error(w, "domain required", http.StatusBadRequest)
		return
	}

	site, err := s.db.AddSite(req.Domain)
	if err != nil {
		http.Error(w, "failed to add site", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SiteResponse{
		ID:     site.ID,
		Domain: site.Domain,
	})
}

func (s *Server) handleRemoveSite(w http.ResponseWriter, r *http.Request) {
	domain := r.PathValue("domain")
	if domain == "" {
		http.Error(w, "domain required", http.StatusBadRequest)
		return
	}

	if err := s.db.RemoveSite(domain); err != nil {
		http.Error(w, "failed to remove site", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleListSites(w http.ResponseWriter, r *http.Request) {
	sites, err := s.db.ListSites()
	if err != nil {
		http.Error(w, "failed to list sites", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sites)
}
