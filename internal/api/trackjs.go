// internal/api/trackjs.go
package api

import (
	"net/http"

	"github.com/adventurehound/beep-analytics/internal/web"
)

func (s *Server) handleTrackJS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Write(web.TrackJS)
}
