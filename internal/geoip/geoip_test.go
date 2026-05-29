package geoip

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func waitReady(t *testing.T, l *Lookup, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for !l.Ready() {
		if time.Now().After(deadline) {
			t.Fatal("timeout waiting for GeoIP to be ready")
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestDisabledLookup(t *testing.T) {
	l, err := New("")
	if err != nil {
		t.Fatalf("new lookup: %v", err)
	}
	if l.Enabled() {
		t.Error("expected lookup to be disabled")
	}
	info := l.LookupIP("1.2.3.4")
	if info.Country != "" {
		t.Errorf("expected empty country for disabled lookup, got %q", info.Country)
	}
}

func TestInvalidIP(t *testing.T) {
	l, err := New("")
	if err != nil {
		t.Fatalf("new lookup: %v", err)
	}
	info := l.LookupIP("not-an-ip")
	if info.Country != "" {
		t.Errorf("expected empty country for invalid IP, got %q", info.Country)
	}
}

func createTestCSVFiles(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()

	// Create test locations CSV (14 columns matching GeoLite2 format)
	// Columns: geoname_id,locale_code,continent_code,continent_name,country_iso_code,country_name,subdivision_1_iso_code,subdivision_1_name,subdivision_2_iso_code,subdivision_2_name,city_name,metro_code,time_zone,is_in_european_union
	// Code uses: row[0]=id, row[4]=Country, row[7]=Region, row[10]=City
	locationsCSV := "geoname_id,locale_code,continent_code,continent_name,country_iso_code,country_name,subdivision_1_iso_code,subdivision_1_name,subdivision_2_iso_code,subdivision_2_name,city_name,metro_code,time_zone,is_in_european_union\n" +
		"100,en,NA,North America,US,United States,CA,California,,,San Francisco,,,0\n" +
		"200,en,EU,Europe,GB,United Kingdom,ENG,England,,,London,,Europe/London,0\n" +
		"300,en,EU,Europe,DE,Germany,,,BE,Berlin,Berlin,,Europe/Berlin,0\n"
	err := os.WriteFile(filepath.Join(dir, "GeoLite2-City-Locations-en.csv"), []byte(locationsCSV), 0644)
	if err != nil {
		t.Fatalf("write locations csv: %v", err)
	}

	// Create test blocks CSV (needs at least 2 columns)
	blocksCSV := "network,geoname_id\n" +
		"192.168.1.0/24,100\n" +
		"10.0.0.0/8,200\n" +
		"172.16.0.0/12,300\n"
	err = os.WriteFile(filepath.Join(dir, "GeoLite2-City-Blocks.csv"), []byte(blocksCSV), 0644)
	if err != nil {
		t.Fatalf("write blocks csv: %v", err)
	}

	return filepath.Join(dir, "dummy.mmdb")
}

func TestEnabledLookup(t *testing.T) {
	mmdbPath := createTestCSVFiles(t)

	l, err := New(mmdbPath)
	if err != nil {
		t.Fatalf("new lookup: %v", err)
	}
	if !l.Enabled() {
		t.Error("expected lookup to be enabled")
	}
	waitReady(t, l, 5*time.Second)
}

func TestLookupIPFound(t *testing.T) {
	mmdbPath := createTestCSVFiles(t)

	l, err := New(mmdbPath)
	if err != nil {
		t.Fatalf("new lookup: %v", err)
	}
	waitReady(t, l, 5*time.Second)

	tests := []struct {
		ip      string
		country string
		region  string
		city    string
	}{
		{"192.168.1.100", "US", "California", "San Francisco"},
		{"10.50.60.70", "GB", "England", "London"},
		{"172.20.30.40", "DE", "", "Berlin"},
	}

	for _, tt := range tests {
		info := l.LookupIP(tt.ip)
		if info.Country != tt.country {
			t.Errorf("LookupIP(%q).Country = %q, want %q", tt.ip, info.Country, tt.country)
		}
		if info.Region != tt.region {
			t.Errorf("LookupIP(%q).Region = %q, want %q", tt.ip, info.Region, tt.region)
		}
		if info.City != tt.city {
			t.Errorf("LookupIP(%q).City = %q, want %q", tt.ip, info.City, tt.city)
		}
	}
}

func TestLookupIPNotFound(t *testing.T) {
	mmdbPath := createTestCSVFiles(t)

	l, err := New(mmdbPath)
	if err != nil {
		t.Fatalf("new lookup: %v", err)
	}
	waitReady(t, l, 5*time.Second)

	// IP not in any block
	info := l.LookupIP("8.8.8.8")
	if info.Country != "" {
		t.Errorf("expected empty country for unknown IP, got %q", info.Country)
	}
}

func TestLookupIPInvalidIP(t *testing.T) {
	mmdbPath := createTestCSVFiles(t)

	l, err := New(mmdbPath)
	if err != nil {
		t.Fatalf("new lookup: %v", err)
	}
	waitReady(t, l, 5*time.Second)

	info := l.LookupIP("not-an-ip")
	if info.Country != "" {
		t.Errorf("expected empty country for invalid IP, got %q", info.Country)
	}
}

func TestNewMissingLocationsFile(t *testing.T) {
	dir := t.TempDir()

	// Only create blocks file, not locations
	blocksCSV := `network,geoname_id
192.168.1.0/24,100
`
	os.WriteFile(filepath.Join(dir, "GeoLite2-City-Blocks.csv"), []byte(blocksCSV), 0644)

	_, err := New(filepath.Join(dir, "dummy.mmdb"))
	if err == nil {
		t.Error("expected error for missing locations file")
	}
}

func TestNewMissingBlocksFile(t *testing.T) {
	dir := t.TempDir()

	// Only create locations file, not blocks
	locationsCSV := `geoname_id,country_name
100,United States
`
	os.WriteFile(filepath.Join(dir, "GeoLite2-City-Locations-en.csv"), []byte(locationsCSV), 0644)

	_, err := New(filepath.Join(dir, "dummy.mmdb"))
	if err == nil {
		t.Error("expected error for missing blocks file")
	}
}

func TestLookupIPNoMatchingLocation(t *testing.T) {
	dir := t.TempDir()

	// Create locations that don't match blocks
	locationsCSV := `geoname_id,country_name
200,United Kingdom
`
	os.WriteFile(filepath.Join(dir, "GeoLite2-City-Locations-en.csv"), []byte(locationsCSV), 0644)

	// Block references geoID 100 which doesn't exist
	blocksCSV := `network,geoname_id
192.168.1.0/24,100
`
	os.WriteFile(filepath.Join(dir, "GeoLite2-City-Blocks.csv"), []byte(blocksCSV), 0644)

	l, err := New(filepath.Join(dir, "dummy.mmdb"))
	if err != nil {
		t.Fatalf("new lookup: %v", err)
	}
	waitReady(t, l, 5*time.Second)

	// Should return empty since location doesn't exist
	info := l.LookupIP("192.168.1.100")
	if info.Country != "" {
		t.Errorf("expected empty country for missing location, got %q", info.Country)
	}
}
