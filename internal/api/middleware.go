package api

import (
	"net/http"
	"strings"
)

func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(auth, "Bearer ")
		if token == auth {
			// No "Bearer " prefix found
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		valid, err := s.db.ValidateToken(token)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if !valid {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}

// requireAuthUnlessNoTokens allows access if no tokens exist yet (bootstrap)
func (s *Server) requireAuthUnlessNoTokens(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hasTokens, err := s.db.HasAnyToken()
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		if !hasTokens {
			// No tokens exist yet, allow bootstrap
			next(w, r)
			return
		}

		s.requireAuth(next)(w, r)
	}
}
