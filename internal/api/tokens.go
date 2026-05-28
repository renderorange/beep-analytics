package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/adventurehound/beep-analytics/internal/db"
)

func (s *Server) handleGenerateToken(w http.ResponseWriter, r *http.Request) {
	token, id, err := s.db.GenerateToken()
	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token": token,
		"id":    id,
	})
}

func (s *Server) handleRevokeToken(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid token id", http.StatusBadRequest)
		return
	}

	if err := s.db.RevokeToken(id); err != nil {
		if errors.Is(err, db.ErrTokenNotFound) {
			http.Error(w, "token not found", http.StatusNotFound)
		} else {
			http.Error(w, "failed to revoke token", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleListTokens(w http.ResponseWriter, r *http.Request) {
	tokens, err := s.db.ListTokens()
	if err != nil {
		http.Error(w, "failed to list tokens", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokens)
}
