package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/adventurehound/beep/internal/db"
)

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	// Parse time range
	fromStr := q.Get("from")
	toStr := q.Get("to")
	lastStr := q.Get("last")

	var from, to time.Time
	var err error

	if lastStr != "" {
		to = time.Now()
		dur, err := parseDuration(lastStr)
		if err != nil {
			http.Error(w, "invalid last duration", http.StatusBadRequest)
			return
		}
		from = to.Add(-dur)
	} else if fromStr != "" && toStr != "" {
		from, err = time.Parse("2006-01-02", fromStr)
		if err != nil {
			http.Error(w, "invalid from date", http.StatusBadRequest)
			return
		}
		to, err = time.Parse("2006-01-02", toStr)
		if err != nil {
			http.Error(w, "invalid to date", http.StatusBadRequest)
			return
		}
		to = to.Add(24*time.Hour - time.Second) // End of day
	} else {
		// Default to last 24 hours
		to = time.Now()
		from = to.Add(-24 * time.Hour)
	}

	// Parse site filter
	siteDomain := q.Get("site")
	var siteID int64
	if siteDomain != "" {
		site, err := s.db.GetSiteByDomain(siteDomain)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if site == nil {
			http.Error(w, "site not found", http.StatusNotFound)
			return
		}
		siteID = site.ID
	}

	// Parse verbose flag
	verbose := q.Get("verbose") == "true"

	statsQuery := db.StatsQuery{
		SiteID: siteID,
		From:   from,
		To:     to,
	}

	if verbose {
		results, err := s.db.GetVerboseStats(statsQuery)
		if err != nil {
			http.Error(w, "failed to get stats", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
	} else {
		results, err := s.db.GetAggregateStats(statsQuery)
		if err != nil {
			http.Error(w, "failed to get stats", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
	}
}

func parseDuration(s string) (time.Duration, error) {
	switch s {
	case "24h":
		return 24 * time.Hour, nil
	case "7d":
		return 7 * 24 * time.Hour, nil
	case "30d":
		return 30 * 24 * time.Hour, nil
	default:
		// Try parsing as Go duration
		return time.ParseDuration(s)
	}
}
