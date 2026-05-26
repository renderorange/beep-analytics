// internal/api/collect.go
package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/adventurehound/beep/internal/models"
	"github.com/adventurehound/beep/internal/ua"
)

type CollectRequest struct {
	Origin string `json:"origin"`
	Path   string `json:"path"`
	Referrer string `json:"referrer"`
	Screen string `json:"screen"`
}

func (s *Server) handleCollect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get origin from header (set by browser)
	origin := r.Header.Get("Origin")
	if origin == "" {
		// Fallback to referer
		origin = r.Header.Get("Referer")
	}
	if origin == "" {
		http.Error(w, "origin required", http.StatusBadRequest)
		return
	}

	// Parse origin to get domain
	domain := extractDomain(origin)
	if domain == "" {
		http.Error(w, "invalid origin", http.StatusBadRequest)
		return
	}

	// Look up site
	site, err := s.db.GetSiteByDomain(domain)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if site == nil {
		// Site not registered, silently ignore
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Parse request body
	var req CollectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	// Get client IP
	ip := getClientIP(r)

	// Check if IP is ignored
	ignored, err := s.db.MatchesIgnoredIP(ip)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if ignored {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Parse user agent
	uaInfo := ua.Parse(r.UserAgent())

	// Parse screen size
	screenWidth, screenHeight := parseScreen(req.Screen)

	// GeoIP lookup
	var country, region, city, locality string
	if s.geoip.Enabled() {
		geo := s.geoip.LookupIP(ip)
		country = geo.Country
		region = geo.Region
		city = geo.City
		locality = geo.Locality
	}

	// Store pageview
	pv := models.PageviewInput{
		SiteID:       site.ID,
		Path:         req.Path,
		Referrer:     req.Referrer,
		Browser:      uaInfo.Browser,
		OS:           uaInfo.OS,
		ScreenWidth:  screenWidth,
		ScreenHeight: screenHeight,
		Country:      country,
		Region:       region,
		City:         city,
		Locality:     locality,
		IP:           ip,
		UserAgent:    r.UserAgent(),
	}

	if err := s.db.InsertPageview(pv); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func extractDomain(origin string) string {
	// Remove protocol
	origin = strings.TrimPrefix(origin, "https://")
	origin = strings.TrimPrefix(origin, "http://")
	// Remove port
	if idx := strings.Index(origin, ":"); idx != -1 {
		origin = origin[:idx]
	}
	// Remove path
	if idx := strings.Index(origin, "/"); idx != -1 {
		origin = origin[:idx]
	}
	return origin
}

func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// First IP in the list is the client
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

func parseScreen(screen string) (int, int) {
	parts := strings.Split(screen, "x")
	if len(parts) != 2 {
		return 0, 0
	}
	w, _ := strconv.Atoi(parts[0])
	h, _ := strconv.Atoi(parts[1])
	return w, h
}
