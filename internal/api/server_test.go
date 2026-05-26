package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/adventurehound/beep/internal/db"
	"github.com/adventurehound/beep/internal/geoip"
)

func setupTestServer(t *testing.T) (*Server, *httptest.Server) {
	t.Helper()
	path := t.TempDir() + "/test.db"
	database, err := db.Open(path)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { database.Close() })

	geo, _ := geoip.New("")
	srv := NewServer(database, geo, ":0")
	ts := httptest.NewServer(srv.mux)
	t.Cleanup(ts.Close)
	return srv, ts
}

func TestTrackJSRoute(t *testing.T) {
	_, ts := setupTestServer(t)
	resp, err := http.Get(ts.URL + "/track.js")
	if err != nil {
		t.Fatalf("get track.js: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestCollectRequiresOrigin(t *testing.T) {
	_, ts := setupTestServer(t)
	resp, err := http.Post(ts.URL+"/collect", "application/json", nil)
	if err != nil {
		t.Fatalf("post collect: %v", err)
	}
	// Should fail without origin header
	if resp.StatusCode == 200 {
		t.Error("expected non-200 without origin")
	}
}

func TestAPIRequiresAuth(t *testing.T) {
	_, ts := setupTestServer(t)
	resp, err := http.Get(ts.URL + "/api/sites")
	if err != nil {
		t.Fatalf("get sites: %v", err)
	}
	if resp.StatusCode != 401 {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}
