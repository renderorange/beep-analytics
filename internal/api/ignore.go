package api

import (
	"encoding/json"
	"net/http"
)

func (s *Server) handleAddIgnore(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IP string `json:"ip"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.IP == "" {
		http.Error(w, "ip required", http.StatusBadRequest)
		return
	}

	if err := s.db.AddIgnoredIP(req.IP); err != nil {
		http.Error(w, "failed to add ignored ip", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (s *Server) handleRemoveIgnore(w http.ResponseWriter, r *http.Request) {
	ip := r.PathValue("ip")
	if ip == "" {
		http.Error(w, "ip required", http.StatusBadRequest)
		return
	}

	if err := s.db.RemoveIgnoredIP(ip); err != nil {
		http.Error(w, "failed to remove ignored ip", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleListIgnore(w http.ResponseWriter, r *http.Request) {
	ips, err := s.db.ListIgnoredIPs()
	if err != nil {
		http.Error(w, "failed to list ignored ips", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ips)
}
