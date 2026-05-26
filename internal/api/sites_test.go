package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSitesCRUD(t *testing.T) {
	_, ts := setupTestServer(t)

	// Generate token
	token := generateTestToken(t, ts)

	// List sites (empty)
	sites := listSites(t, ts, token)
	if len(sites) != 0 {
		t.Errorf("expected 0 sites, got %d", len(sites))
	}

	// Add site
	body := `{"domain":"example.com"}`
	req, _ := http.NewRequest("POST", ts.URL+"/api/sites", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != 200 {
		t.Fatalf("add site: expected 200, got %d", resp.StatusCode)
	}

	// List sites (1)
	sites = listSites(t, ts, token)
	if len(sites) != 1 {
		t.Errorf("expected 1 site, got %d", len(sites))
	}
	if sites[0].Domain != "example.com" {
		t.Errorf("expected example.com, got %s", sites[0].Domain)
	}

	// Remove site
	req, _ = http.NewRequest("DELETE", ts.URL+"/api/sites/example.com", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ = http.DefaultClient.Do(req)
	if resp.StatusCode != 204 {
		t.Errorf("remove site: expected 204, got %d", resp.StatusCode)
	}

	// List sites (empty)
	sites = listSites(t, ts, token)
	if len(sites) != 0 {
		t.Errorf("expected 0 sites after removal, got %d", len(sites))
	}
}

func generateTestToken(t *testing.T, ts *httptest.Server) string {
	t.Helper()
	resp, _ := http.Post(ts.URL+"/api/tokens/generate", "application/json", nil)
	var result struct {
		Token string `json:"token"`
		ID    int64  `json:"id"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	resp.Body.Close()
	return result.Token
}

func listSites(t *testing.T, ts *httptest.Server, token string) []SiteResponse {
	t.Helper()
	req, _ := http.NewRequest("GET", ts.URL+"/api/sites", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ := http.DefaultClient.Do(req)
	var sites []SiteResponse
	json.NewDecoder(resp.Body).Decode(&sites)
	resp.Body.Close()
	return sites
}
