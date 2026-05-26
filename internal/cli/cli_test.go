package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adventurehound/beep/internal/models"
)

func TestNewClient(t *testing.T) {
	c := NewClient("http://localhost:9999", "token123")
	if c.BaseURL != "http://localhost:9999" {
		t.Errorf("expected BaseURL http://localhost:9999, got %s", c.BaseURL)
	}
	if c.Token != "token123" {
		t.Errorf("expected Token token123, got %s", c.Token)
	}
}

func TestNewClientDefaultServer(t *testing.T) {
	c := NewClient("", "")
	if c.BaseURL != "http://localhost:8080" {
		t.Errorf("expected default BaseURL http://localhost:8080, got %s", c.BaseURL)
	}
}

func TestLoadTokenFromEnv(t *testing.T) {
	os.Setenv("BEEP_TOKEN", "env-token")
	defer os.Unsetenv("BEEP_TOKEN")

	token := LoadToken()
	if token != "env-token" {
		t.Errorf("expected env-token, got %s", token)
	}
}

func TestLoadTokenFromFile(t *testing.T) {
	os.Unsetenv("BEEP_TOKEN")

	// Create temp config dir
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "beep")
	os.MkdirAll(configDir, 0755)

	// Write token file
	tokenFile := filepath.Join(configDir, "token")
	os.WriteFile(tokenFile, []byte("file-token\n"), 0644)

	// Override home dir for test
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	token := LoadToken()
	if token != "file-token" {
		t.Errorf("expected file-token, got %s", token)
	}
}

func TestLoadTokenNone(t *testing.T) {
	os.Unsetenv("BEEP_TOKEN")

	// Use empty temp dir
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	token := LoadToken()
	if token != "" {
		t.Errorf("expected empty token, got %s", token)
	}
}

func TestClientGet(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/sites" {
			t.Errorf("expected /api/sites, got %s", r.URL.Path)
		}
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("expected Bearer test-token, got %s", auth)
		}
		json.NewEncoder(w).Encode([]Site{{ID: 1, Domain: "example.com"}})
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "test-token")
	data, err := c.Get("/api/sites")
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	var sites []Site
	json.Unmarshal(data, &sites)
	if len(sites) != 1 || sites[0].Domain != "example.com" {
		t.Errorf("unexpected sites: %v", sites)
	}
}

func TestClientPost(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		ct := r.Header.Get("Content-Type")
		if ct != "application/json" {
			t.Errorf("expected application/json content type, got %s", ct)
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "test-token")
	_, err := c.Post("/api/sites", map[string]string{"domain": "example.com"})
	if err != nil {
		t.Fatalf("post: %v", err)
	}
}

func TestClientDelete(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "test-token")
	_, err := c.Delete("/api/sites/example.com")
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestClientError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "test-token")
	_, err := c.Get("/api/nonexistent")
	if err == nil {
		t.Error("expected error for 404 response")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("expected 404 in error, got %v", err)
	}
}

func TestClientNoToken(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "" {
			t.Errorf("expected no auth header, got %s", auth)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "")
	c.Get("/api/sites")
}

func TestParseGlobalFlags(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedServer string
		expectedToken  string
	}{
		{
			name:           "default",
			args:           []string{},
			expectedServer: "http://localhost:8080",
			expectedToken:  "",
		},
		{
			name:           "custom server",
			args:           []string{"--server", "http://example.com:9000"},
			expectedServer: "http://example.com:9000",
		},
		{
			name:          "custom token",
			args:          []string{"--token", "my-token"},
			expectedToken: "my-token",
		},
		{
			name:           "both custom",
			args:           []string{"--server", "http://example.com", "--token", "token123"},
			expectedServer: "http://example.com",
			expectedToken:  "token123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, token, remaining := ParseGlobalFlags(tt.args)
			if tt.expectedServer != "" && server != tt.expectedServer {
				t.Errorf("expected server %s, got %s", tt.expectedServer, server)
			}
			if tt.expectedToken != "" && token != tt.expectedToken {
				t.Errorf("expected token %s, got %s", tt.expectedToken, token)
			}
			if len(remaining) != 0 {
				t.Errorf("expected no remaining args, got %v", remaining)
			}
		})
	}
}

func TestParseGlobalFlagsWithRemaining(t *testing.T) {
	_, _, remaining := ParseGlobalFlags([]string{"--server", "http://example.com", "example.com"})
	if len(remaining) != 1 || remaining[0] != "example.com" {
		t.Errorf("expected remaining [example.com], got %v", remaining)
	}
}

func TestParseGlobalFlagsServerFromEnv(t *testing.T) {
	os.Setenv("BEEP_SERVER", "http://env-server:9090")
	defer os.Unsetenv("BEEP_SERVER")

	server, _, _ := ParseGlobalFlags([]string{})
	if server != "http://env-server:9090" {
		t.Errorf("expected server from env, got %s", server)
	}
}

func TestParseGlobalFlagsFlagOverridesEnv(t *testing.T) {
	os.Setenv("BEEP_SERVER", "http://env-server:9090")
	defer os.Unsetenv("BEEP_SERVER")

	server, _, _ := ParseGlobalFlags([]string{"--server", "http://flag-server:8080"})
	if server != "http://flag-server:8080" {
		t.Errorf("expected flag to override env, got %s", server)
	}
}

func TestTruncateReferrer(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "(direct)"},
		{"https://google.com", "https://google.com"},
		{"https://very-long-referrer-url-that-exceeds-twenty-characters.com/page", "https://very-long..."},
	}

	for _, tt := range tests {
		result := truncateReferrer(tt.input)
		if result != tt.expected {
			t.Errorf("truncateReferrer(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestDisplayAggregateStatsEmpty(t *testing.T) {
	// Just test that it doesn't panic with empty data
	displayAggregateStats([]byte("[]"), "")
}

func TestDisplayAggregateStatsWithData(t *testing.T) {
	stats := []models.StatsRow{
		{Site: "example.com", IP: "1.2.3.4", Path: "/", Count: 10},
		{Site: "example.com", IP: "5.6.7.8", Path: "/about", Count: 5},
	}
	data, _ := json.Marshal(stats)
	displayAggregateStats(data, "example.com")
}

func TestDisplayAggregateStatsMultipleSites(t *testing.T) {
	stats := []models.StatsRow{
		{Site: "example.com", IP: "1.2.3.4", Path: "/", Count: 10},
		{Site: "other.com", IP: "5.6.7.8", Path: "/blog", Count: 3},
	}
	data, _ := json.Marshal(stats)
	displayAggregateStats(data, "")
}

func TestDisplayVerboseStatsEmpty(t *testing.T) {
	displayVerboseStats([]byte("[]"), "")
}

func TestDisplayVerboseStatsWithData(t *testing.T) {
	stats := []models.StatsRow{
		{Site: "example.com", IP: "1.2.3.4", Path: "/", Browser: "Chrome", OS: "Windows", Country: "US", Referrer: "https://google.com", Time: "2024-01-15T14:23:00Z"},
	}
	data, _ := json.Marshal(stats)
	displayVerboseStats(data, "example.com")
}

func TestDisplayVerboseStatsLongTime(t *testing.T) {
	stats := []models.StatsRow{
		{Site: "example.com", IP: "1.2.3.4", Path: "/", Time: "2024-01-15T14:23:00+00:00"},
	}
	data, _ := json.Marshal(stats)
	displayVerboseStats(data, "example.com")
}

func TestDisplayVerboseStatsMultipleSites(t *testing.T) {
	stats := []models.StatsRow{
		{Site: "example.com", IP: "1.2.3.4", Path: "/"},
		{Site: "other.com", IP: "5.6.7.8", Path: "/blog"},
	}
	data, _ := json.Marshal(stats)
	displayVerboseStats(data, "")
}
