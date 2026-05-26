package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestAuthMiddleware(t *testing.T) {
	_, ts := setupTestServer(t)

	// Generate a token first (bootstrap mode, no auth needed)
	resp, err := http.Post(ts.URL+"/api/tokens/generate", "application/json", nil)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var result struct {
		Token string `json:"token"`
		ID    int64  `json:"id"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	resp.Body.Close()

	// Now try to access without token
	resp, _ = http.Get(ts.URL + "/api/sites")
	if resp.StatusCode != 401 {
		t.Errorf("expected 401 without token, got %d", resp.StatusCode)
	}

	// Try with invalid token
	req, _ := http.NewRequest("GET", ts.URL+"/api/sites", nil)
	req.Header.Set("Authorization", "Bearer invalid")
	resp, _ = http.DefaultClient.Do(req)
	if resp.StatusCode != 401 {
		t.Errorf("expected 401 with invalid token, got %d", resp.StatusCode)
	}

	// Try with valid token
	req, _ = http.NewRequest("GET", ts.URL+"/api/sites", nil)
	req.Header.Set("Authorization", "Bearer "+result.Token)
	resp, _ = http.DefaultClient.Do(req)
	if resp.StatusCode != 200 {
		t.Errorf("expected 200 with valid token, got %d", resp.StatusCode)
	}

	// Revoke token, then try again
	req, _ = http.NewRequest("DELETE", ts.URL+"/api/tokens/1", nil)
	req.Header.Set("Authorization", "Bearer "+result.Token)
	resp, _ = http.DefaultClient.Do(req)
	if resp.StatusCode != 204 {
		t.Errorf("expected 204 revoking token, got %d", resp.StatusCode)
	}

	req, _ = http.NewRequest("GET", ts.URL+"/api/sites", nil)
	req.Header.Set("Authorization", "Bearer "+result.Token)
	resp, _ = http.DefaultClient.Do(req)
	if resp.StatusCode != 401 {
		t.Errorf("expected 401 after token revoked, got %d", resp.StatusCode)
	}
}

func TestAuthWithoutBearerPrefix(t *testing.T) {
	_, ts := setupTestServer(t)
	token := generateTestToken(t, ts)

	// Try with auth but no Bearer prefix
	req, _ := http.NewRequest("GET", ts.URL+"/api/sites", nil)
	req.Header.Set("Authorization", token)
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != 401 {
		t.Errorf("expected 401 without Bearer prefix, got %d", resp.StatusCode)
	}
}

func TestListTokensEndpoint(t *testing.T) {
	_, ts := setupTestServer(t)
	token := generateTestToken(t, ts)

	// List tokens
	req, _ := http.NewRequest("GET", ts.URL+"/api/tokens", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != 200 {
		t.Fatalf("list tokens: expected 200, got %d", resp.StatusCode)
	}

	var tokens []struct {
		ID int64 `json:"id"`
	}
	json.NewDecoder(resp.Body).Decode(&tokens)
	resp.Body.Close()

	// Should have at least the token we generated
	if len(tokens) < 1 {
		t.Errorf("expected at least 1 token, got %d", len(tokens))
	}
}

func TestRevokeTokenInvalidID(t *testing.T) {
	_, ts := setupTestServer(t)
	token := generateTestToken(t, ts)

	// Try to revoke with invalid ID
	req, _ := http.NewRequest("DELETE", ts.URL+"/api/tokens/invalid", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != 400 {
		t.Errorf("invalid token id: expected 400, got %d", resp.StatusCode)
	}
}

func TestAddSiteInvalidJSON(t *testing.T) {
	_, ts := setupTestServer(t)
	token := generateTestToken(t, ts)

	req, _ := http.NewRequest("POST", ts.URL+"/api/sites", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != 400 {
		t.Errorf("invalid json: expected 400, got %d", resp.StatusCode)
	}
}

func TestAddSiteEmptyDomain(t *testing.T) {
	_, ts := setupTestServer(t)
	token := generateTestToken(t, ts)

	body := `{"domain":""}`
	req, _ := http.NewRequest("POST", ts.URL+"/api/sites", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != 400 {
		t.Errorf("empty domain: expected 400, got %d", resp.StatusCode)
	}
}

func TestAddIgnoreInvalidJSON(t *testing.T) {
	_, ts := setupTestServer(t)
	token := generateTestToken(t, ts)

	req, _ := http.NewRequest("POST", ts.URL+"/api/ignore", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != 400 {
		t.Errorf("invalid json: expected 400, got %d", resp.StatusCode)
	}
}

func TestAddIgnoreEmptyIP(t *testing.T) {
	_, ts := setupTestServer(t)
	token := generateTestToken(t, ts)

	body := `{"ip":""}`
	req, _ := http.NewRequest("POST", ts.URL+"/api/ignore", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != 400 {
		t.Errorf("empty ip: expected 400, got %d", resp.StatusCode)
	}
}

func TestCollectInvalidJSON(t *testing.T) {
	_, ts := setupTestServer(t)
	token := generateTestToken(t, ts)

	// Add a site
	body := `{"domain":"example.com"}`
	req, _ := http.NewRequest("POST", ts.URL+"/api/sites", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	http.DefaultClient.Do(req)

	// Send invalid JSON to collect
	req, _ = http.NewRequest("POST", ts.URL+"/collect", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://example.com")
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != 400 {
		t.Errorf("invalid json: expected 400, got %d", resp.StatusCode)
	}
}

func TestCollectWithReferer(t *testing.T) {
	_, ts := setupTestServer(t)
	token := generateTestToken(t, ts)

	// Add a site
	body := `{"domain":"example.com"}`
	req, _ := http.NewRequest("POST", ts.URL+"/api/sites", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	http.DefaultClient.Do(req)

	// Send tracking data with Referer instead of Origin
	collectBody := `{"origin":"https://example.com","path":"/","referrer":"","screen":"1920x1080"}`
	req, _ = http.NewRequest("POST", ts.URL+"/collect", strings.NewReader(collectBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Referer", "https://example.com/page")
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != 204 {
		t.Errorf("collect with referer: expected 204, got %d", resp.StatusCode)
	}
}

func TestCollectWithXForwardedFor(t *testing.T) {
	_, ts := setupTestServer(t)
	token := generateTestToken(t, ts)

	// Add a site
	body := `{"domain":"example.com"}`
	req, _ := http.NewRequest("POST", ts.URL+"/api/sites", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	http.DefaultClient.Do(req)

	// Send tracking data with X-Forwarded-For
	collectBody := `{"origin":"https://example.com","path":"/","referrer":"","screen":"1920x1080"}`
	req, _ = http.NewRequest("POST", ts.URL+"/collect", strings.NewReader(collectBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("X-Forwarded-For", "203.0.113.1, 70.41.3.18")
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != 204 {
		t.Errorf("collect with X-Forwarded-For: expected 204, got %d", resp.StatusCode)
	}
}

func TestCollectWithXRealIP(t *testing.T) {
	_, ts := setupTestServer(t)
	token := generateTestToken(t, ts)

	// Add a site
	body := `{"domain":"example.com"}`
	req, _ := http.NewRequest("POST", ts.URL+"/api/sites", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	http.DefaultClient.Do(req)

	// Send tracking data with X-Real-IP
	collectBody := `{"origin":"https://example.com","path":"/","referrer":"","screen":"1920x1080"}`
	req, _ = http.NewRequest("POST", ts.URL+"/collect", strings.NewReader(collectBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("X-Real-IP", "203.0.113.2")
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != 204 {
		t.Errorf("collect with X-Real-IP: expected 204, got %d", resp.StatusCode)
	}
}

func TestExtractDomain(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"https://example.com", "example.com"},
		{"http://example.com", "example.com"},
		{"https://example.com:8080", "example.com"},
		{"https://example.com/path", "example.com"},
		{"https://example.com:8080/path", "example.com"},
	}

	for _, tt := range tests {
		result := extractDomain(tt.input)
		if result != tt.expected {
			t.Errorf("extractDomain(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestParseScreen(t *testing.T) {
	tests := []struct {
		input string
		w     int
		h     int
	}{
		{"1920x1080", 1920, 1080},
		{"800x600", 800, 600},
		{"invalid", 0, 0},
		{"1920", 0, 0},
		{"1920xabc", 1920, 0},
	}

	for _, tt := range tests {
		w, h := parseScreen(tt.input)
		if w != tt.w || h != tt.h {
			t.Errorf("parseScreen(%q) = (%d, %d), want (%d, %d)", tt.input, w, h, tt.w, tt.h)
		}
	}
}
