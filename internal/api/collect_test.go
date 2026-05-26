// internal/api/collect_test.go
package api

import (
	"net/http"
	"strings"
	"testing"
)

func TestCollect(t *testing.T) {
	_, ts := setupTestServer(t)
	token := generateTestToken(t, ts)

	// Add a site
	body := `{"domain":"example.com"}`
	req, _ := http.NewRequest("POST", ts.URL+"/api/sites", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	http.DefaultClient.Do(req)

	// Send tracking data
	collectBody := `{"origin":"https://example.com","path":"/test","referrer":"https://google.com","screen":"1920x1080"}`
	req, _ = http.NewRequest("POST", ts.URL+"/collect", strings.NewReader(collectBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/120.0.0.0")
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != 204 {
		t.Errorf("collect: expected 204, got %d", resp.StatusCode)
	}
}

func TestCollectIgnoredSite(t *testing.T) {
	_, ts := setupTestServer(t)

	// Send tracking for unregistered site
	collectBody := `{"origin":"https://unknown.com","path":"/","referrer":"","screen":"800x600"}`
	req, _ := http.NewRequest("POST", ts.URL+"/collect", strings.NewReader(collectBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://unknown.com")
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != 204 {
		t.Errorf("collect unknown site: expected 204, got %d", resp.StatusCode)
	}
}

func TestCollectIgnoredIP(t *testing.T) {
	_, ts := setupTestServer(t)
	token := generateTestToken(t, ts)

	// Add a site
	body := `{"domain":"example.com"}`
	req, _ := http.NewRequest("POST", ts.URL+"/api/sites", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	http.DefaultClient.Do(req)

	// Add ignored IP
	body = `{"ip":"127.0.0.1"}`
	req, _ = http.NewRequest("POST", ts.URL+"/api/ignore", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	http.DefaultClient.Do(req)

	// Send tracking from ignored IP
	collectBody := `{"origin":"https://example.com","path":"/","referrer":"","screen":"800x600"}`
	req, _ = http.NewRequest("POST", ts.URL+"/collect", strings.NewReader(collectBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://example.com")
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != 204 {
		t.Errorf("collect ignored IP: expected 204, got %d", resp.StatusCode)
	}
}

func TestTrackJS(t *testing.T) {
	_, ts := setupTestServer(t)
	resp, err := http.Get(ts.URL + "/track.js")
	if err != nil {
		t.Fatalf("get track.js: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if ct != "application/javascript" {
		t.Errorf("expected application/javascript, got %q", ct)
	}
}
