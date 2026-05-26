package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestIgnoreCRUD(t *testing.T) {
	_, ts := setupTestServer(t)
	token := generateTestToken(t, ts)

	// List (empty)
	req, _ := http.NewRequest("GET", ts.URL+"/api/ignore", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ := http.DefaultClient.Do(req)
	var ips []string
	json.NewDecoder(resp.Body).Decode(&ips)
	resp.Body.Close()
	if len(ips) != 0 {
		t.Errorf("expected 0 ignored IPs, got %d", len(ips))
	}

	// Add
	body := `{"ip":"1.2.3.4"}`
	req, _ = http.NewRequest("POST", ts.URL+"/api/ignore", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ = http.DefaultClient.Do(req)
	if resp.StatusCode != 201 {
		t.Errorf("add ignore: expected 201, got %d", resp.StatusCode)
	}

	// List (1)
	req, _ = http.NewRequest("GET", ts.URL+"/api/ignore", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ = http.DefaultClient.Do(req)
	json.NewDecoder(resp.Body).Decode(&ips)
	resp.Body.Close()
	if len(ips) != 1 {
		t.Errorf("expected 1 ignored IP, got %d", len(ips))
	}

	// Remove
	req, _ = http.NewRequest("DELETE", ts.URL+"/api/ignore/1.2.3.4", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ = http.DefaultClient.Do(req)
	if resp.StatusCode != 204 {
		t.Errorf("remove ignore: expected 204, got %d", resp.StatusCode)
	}

	// List (empty)
	req, _ = http.NewRequest("GET", ts.URL+"/api/ignore", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ = http.DefaultClient.Do(req)
	json.NewDecoder(resp.Body).Decode(&ips)
	resp.Body.Close()
	if len(ips) != 0 {
		t.Errorf("expected 0 after removal, got %d", len(ips))
	}
}
