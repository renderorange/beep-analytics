package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestStatsEndpoint(t *testing.T) {
	_, ts := setupTestServer(t)
	token := generateTestToken(t, ts)

	// Add a site
	body := `{"domain":"example.com"}`
	req, _ := http.NewRequest("POST", ts.URL+"/api/sites", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	http.DefaultClient.Do(req)

	// Collect some data
	collectBody := `{"origin":"https://example.com","path":"/","referrer":"https://google.com","screen":"1920x1080"}`
	req, _ = http.NewRequest("POST", ts.URL+"/collect", strings.NewReader(collectBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://example.com")
	http.DefaultClient.Do(req)

	// Get aggregate stats
	req, _ = http.NewRequest("GET", ts.URL+"/api/stats?site=example.com&last=24h", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != 200 {
		t.Fatalf("stats: expected 200, got %d", resp.StatusCode)
	}

	var stats []struct {
		Site  string `json:"site"`
		IP    string `json:"ip"`
		Path  string `json:"path"`
		Count int    `json:"count"`
	}
	json.NewDecoder(resp.Body).Decode(&stats)
	resp.Body.Close()

	if len(stats) != 1 {
		t.Errorf("expected 1 stat row, got %d", len(stats))
	}

	// Get verbose stats
	req, _ = http.NewRequest("GET", ts.URL+"/api/stats?site=example.com&last=24h&verbose=true", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ = http.DefaultClient.Do(req)
	if resp.StatusCode != 200 {
		t.Fatalf("verbose stats: expected 200, got %d", resp.StatusCode)
	}

	var verbose []struct {
		Site     string `json:"site"`
		IP       string `json:"ip"`
		Path     string `json:"path"`
		Browser  string `json:"browser"`
		OS       string `json:"os"`
		Referrer string `json:"referrer"`
	}
	json.NewDecoder(resp.Body).Decode(&verbose)
	resp.Body.Close()

	if len(verbose) != 1 {
		t.Errorf("expected 1 verbose row, got %d", len(verbose))
	}
}

func TestStatsDefaultTimeRange(t *testing.T) {
	_, ts := setupTestServer(t)
	token := generateTestToken(t, ts)

	// Get stats without specifying time range (should default to last 24h)
	req, _ := http.NewRequest("GET", ts.URL+"/api/stats", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != 200 {
		t.Fatalf("default time range: expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestStatsInvalidDuration(t *testing.T) {
	_, ts := setupTestServer(t)
	token := generateTestToken(t, ts)

	// Try with invalid duration
	req, _ := http.NewRequest("GET", ts.URL+"/api/stats?last=invalid", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != 400 {
		t.Errorf("invalid duration: expected 400, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestStatsSiteNotFound(t *testing.T) {
	_, ts := setupTestServer(t)
	token := generateTestToken(t, ts)

	// Try with non-existent site
	req, _ := http.NewRequest("GET", ts.URL+"/api/stats?site=nonexistent.com", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != 404 {
		t.Errorf("site not found: expected 404, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestStatsCustomDateRange(t *testing.T) {
	_, ts := setupTestServer(t)
	token := generateTestToken(t, ts)

	// Get stats with custom date range
	req, _ := http.NewRequest("GET", ts.URL+"/api/stats?from=2024-01-01&to=2024-01-31", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != 200 {
		t.Fatalf("custom date range: expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestStatsInvalidDateRange(t *testing.T) {
	_, ts := setupTestServer(t)
	token := generateTestToken(t, ts)

	// Try with invalid from date
	req, _ := http.NewRequest("GET", ts.URL+"/api/stats?from=invalid&to=2024-01-31", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != 400 {
		t.Errorf("invalid from date: expected 400, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Try with invalid to date
	req, _ = http.NewRequest("GET", ts.URL+"/api/stats?from=2024-01-01&to=invalid", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ = http.DefaultClient.Do(req)
	if resp.StatusCode != 400 {
		t.Errorf("invalid to date: expected 400, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestStats7DayRange(t *testing.T) {
	_, ts := setupTestServer(t)
	token := generateTestToken(t, ts)

	req, _ := http.NewRequest("GET", ts.URL+"/api/stats?last=7d", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != 200 {
		t.Errorf("7d range: expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestStats30DayRange(t *testing.T) {
	_, ts := setupTestServer(t)
	token := generateTestToken(t, ts)

	req, _ := http.NewRequest("GET", ts.URL+"/api/stats?last=30d", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != 200 {
		t.Errorf("30d range: expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestStatsAllSites(t *testing.T) {
	_, ts := setupTestServer(t)
	token := generateTestToken(t, ts)

	// Add two sites
	body := `{"domain":"example.com"}`
	req, _ := http.NewRequest("POST", ts.URL+"/api/sites", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	http.DefaultClient.Do(req)

	body = `{"domain":"other.com"}`
	req, _ = http.NewRequest("POST", ts.URL+"/api/sites", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	http.DefaultClient.Do(req)

	// Collect data for both
	collectBody := `{"origin":"https://example.com","path":"/","referrer":"","screen":"1920x1080"}`
	req, _ = http.NewRequest("POST", ts.URL+"/collect", strings.NewReader(collectBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://example.com")
	http.DefaultClient.Do(req)

	collectBody = `{"origin":"https://other.com","path":"/blog","referrer":"","screen":"800x600"}`
	req, _ = http.NewRequest("POST", ts.URL+"/collect", strings.NewReader(collectBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://other.com")
	http.DefaultClient.Do(req)

	// Get stats without site filter
	req, _ = http.NewRequest("GET", ts.URL+"/api/stats?last=24h", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != 200 {
		t.Fatalf("all sites stats: expected 200, got %d", resp.StatusCode)
	}

	var stats []struct {
		Site string `json:"site"`
	}
	json.NewDecoder(resp.Body).Decode(&stats)
	resp.Body.Close()

	// Should have data from both sites
	sites := make(map[string]bool)
	for _, s := range stats {
		sites[s.Site] = true
	}
	if !sites["example.com"] || !sites["other.com"] {
		t.Errorf("expected both sites in results, got %v", sites)
	}
}
