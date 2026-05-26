// integration_test.go
package main

import (
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestIntegration(t *testing.T) {
	// Build binary
	binaryPath := filepath.Join(t.TempDir(), "beep-test")
	cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/beep")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}

	// Start server
	dbPath := t.TempDir() + "/test.db"
	server := exec.Command(binaryPath, "serve", "--port", "18080", "--db", dbPath)
	server.Stdout = os.Stdout
	server.Stderr = os.Stderr
	if err := server.Start(); err != nil {
		t.Fatalf("start server: %v", err)
	}
	defer server.Process.Kill()

	// Wait for server to start
	time.Sleep(500 * time.Millisecond)

	baseURL := "http://localhost:18080"

	// Generate token (bootstrap mode)
	resp, err := http.Post(baseURL+"/api/tokens/generate", "application/json", nil)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	var tokenResp struct {
		Token string `json:"token"`
		ID    int64  `json:"id"`
	}
	json.NewDecoder(resp.Body).Decode(&tokenResp)
	resp.Body.Close()

	// Add site
	body := `{"domain":"test.example.com"}`
	req, _ := http.NewRequest("POST", baseURL+"/api/sites", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokenResp.Token)
	resp, _ = http.DefaultClient.Do(req)
	if resp.StatusCode != 200 {
		t.Fatalf("add site: expected 200, got %d", resp.StatusCode)
	}

	// Send tracking data
	collectBody := `{"origin":"https://test.example.com","path":"/test","referrer":"https://google.com","screen":"1920x1080"}`
	req, _ = http.NewRequest("POST", baseURL+"/collect", strings.NewReader(collectBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://test.example.com")
	resp, _ = http.DefaultClient.Do(req)
	if resp.StatusCode != 204 {
		t.Fatalf("collect: expected 204, got %d", resp.StatusCode)
	}

	// Get stats
	req, _ = http.NewRequest("GET", baseURL+"/api/stats?site=test.example.com&last=24h", nil)
	req.Header.Set("Authorization", "Bearer "+tokenResp.Token)
	resp, _ = http.DefaultClient.Do(req)
	if resp.StatusCode != 200 {
		t.Fatalf("stats: expected 200, got %d", resp.StatusCode)
	}

	var stats []struct {
		Site  string `json:"site"`
		Count int    `json:"count"`
	}
	json.NewDecoder(resp.Body).Decode(&stats)
	resp.Body.Close()

	if len(stats) != 1 {
		t.Errorf("expected 1 stat row, got %d", len(stats))
	}
	if stats[0].Count != 1 {
		t.Errorf("expected count 1, got %d", stats[0].Count)
	}

	// Test track.js endpoint
	resp, _ = http.Get(baseURL + "/track.js")
	if resp.StatusCode != 200 {
		t.Errorf("track.js: expected 200, got %d", resp.StatusCode)
	}
}
