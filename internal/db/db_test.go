package db

import (
	"errors"
	"testing"
	"time"

	"github.com/adventurehound/beep-analytics/internal/models"
)

func setupTestDB(t *testing.T) *DB {
	t.Helper()
	path := t.TempDir() + "/test.db"
	db, err := Open(path)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestMigrate(t *testing.T) {
	db := setupTestDB(t)
	// If we got here without error, migrations ran
	_ = db
}

func TestAddSite(t *testing.T) {
	db := setupTestDB(t)
	site, err := db.AddSite("example.com")
	if err != nil {
		t.Fatalf("add site: %v", err)
	}
	if site.Domain != "example.com" {
		t.Errorf("expected domain example.com, got %s", site.Domain)
	}
	if site.ID == 0 {
		t.Error("expected non-zero ID")
	}
}

func TestRemoveSite(t *testing.T) {
	db := setupTestDB(t)
	_, _ = db.AddSite("example.com")
	if err := db.RemoveSite("example.com"); err != nil {
		t.Fatalf("remove site: %v", err)
	}
	site, err := db.GetSiteByDomain("example.com")
	if err != nil {
		t.Fatalf("get site: %v", err)
	}
	if site != nil {
		t.Error("expected site to be removed")
	}
}

func TestListSites(t *testing.T) {
	db := setupTestDB(t)
	_, _ = db.AddSite("example.com")
	_, _ = db.AddSite("blog.example.com")

	sites, err := db.ListSites()
	if err != nil {
		t.Fatalf("list sites: %v", err)
	}
	if len(sites) != 2 {
		t.Errorf("expected 2 sites, got %d", len(sites))
	}
}

func TestIgnoredIPs(t *testing.T) {
	db := setupTestDB(t)

	if err := db.AddIgnoredIP("1.2.3.4"); err != nil {
		t.Fatalf("add ignored ip: %v", err)
	}

	ignored, err := db.IsIPIgnored("1.2.3.4")
	if err != nil {
		t.Fatalf("is ip ignored: %v", err)
	}
	if !ignored {
		t.Error("expected 1.2.3.4 to be ignored")
	}

	ignored, _ = db.IsIPIgnored("5.6.7.8")
	if ignored {
		t.Error("expected 5.6.7.8 to not be ignored")
	}

	ips, _ := db.ListIgnoredIPs()
	if len(ips) != 1 {
		t.Errorf("expected 1 ignored ip, got %d", len(ips))
	}

	_ = db.RemoveIgnoredIP("1.2.3.4")
	ignored, _ = db.IsIPIgnored("1.2.3.4")
	if ignored {
		t.Error("expected 1.2.3.4 to not be ignored after removal")
	}
}

func TestTokens(t *testing.T) {
	db := setupTestDB(t)

	// No tokens initially
	hasAny, _ := db.HasAnyToken()
	if hasAny {
		t.Error("expected no tokens initially")
	}

	// Generate token
	token, id, err := db.GenerateToken()
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}
	if id == 0 {
		t.Error("expected non-zero id")
	}

	// Validate token
	valid, _ := db.ValidateToken(token)
	if !valid {
		t.Error("expected token to be valid")
	}

	// Invalid token
	valid, _ = db.ValidateToken("invalid")
	if valid {
		t.Error("expected invalid token to fail validation")
	}

	// Has tokens now
	hasAny, _ = db.HasAnyToken()
	if !hasAny {
		t.Error("expected tokens to exist")
	}

	// Revoke
	_ = db.RevokeToken(id)
	valid, _ = db.ValidateToken(token)
	if valid {
		t.Error("expected token to be invalid after revocation")
	}
}

func TestGetSiteByID(t *testing.T) {
	db := setupTestDB(t)

	// Add a site to get its ID
	site, err := db.AddSite("example.com")
	if err != nil {
		t.Fatalf("add site: %v", err)
	}

	// Test found case
	found, err := db.GetSiteByID(site.ID)
	if err != nil {
		t.Fatalf("get site by id: %v", err)
	}
	if found == nil {
		t.Fatal("expected to find site")
	}
	if found.Domain != "example.com" {
		t.Errorf("expected domain example.com, got %s", found.Domain)
	}
	if found.ID != site.ID {
		t.Errorf("expected ID %d, got %d", site.ID, found.ID)
	}

	// Test not found case
	notFound, err := db.GetSiteByID(99999)
	if err != nil {
		t.Fatalf("get site by id (not found): %v", err)
	}
	if notFound != nil {
		t.Error("expected nil for non-existent site ID")
	}
}

func TestListTokens(t *testing.T) {
	db := setupTestDB(t)

	// Empty initially
	tokens, err := db.ListTokens()
	if err != nil {
		t.Fatalf("list tokens: %v", err)
	}
	if len(tokens) != 0 {
		t.Errorf("expected 0 tokens, got %d", len(tokens))
	}

	// Generate some tokens
	_, id1, _ := db.GenerateToken()
	_, id2, _ := db.GenerateToken()

	tokens, err = db.ListTokens()
	if err != nil {
		t.Fatalf("list tokens: %v", err)
	}
	if len(tokens) != 2 {
		t.Errorf("expected 2 tokens, got %d", len(tokens))
	}

	// Verify token info has expected fields
	for _, token := range tokens {
		if token.ID == 0 {
			t.Error("expected non-zero token ID")
		}
		if token.CreatedAt.IsZero() {
			t.Error("expected non-zero created_at")
		}
	}

	// Verify we got the right IDs
	ids := map[int64]bool{id1: false, id2: false}
	for _, token := range tokens {
		ids[token.ID] = true
	}
	for id, found := range ids {
		if !found {
			t.Errorf("expected token ID %d in list", id)
		}
	}
}

func TestInsertPageview(t *testing.T) {
	db := setupTestDB(t)

	// Need a site first
	site, err := db.AddSite("example.com")
	if err != nil {
		t.Fatalf("add site: %v", err)
	}

	// Insert a pageview
	pv := models.PageviewInput{
		SiteID:       site.ID,
		Path:         "/test",
		Referrer:     "https://google.com",
		Browser:      "Chrome",
		OS:           "Linux",
		ScreenWidth:  1920,
		ScreenHeight: 1080,
		Country:      "US",
		Region:       "CA",
		City:         "San Francisco",
		Locality:     "",
		IP:           "1.2.3.4",
		UserAgent:    "Mozilla/5.0",
	}
	err = db.InsertPageview(pv)
	if err != nil {
		t.Fatalf("insert pageview: %v", err)
	}

	// Insert another pageview with minimal fields
	pv2 := models.PageviewInput{
		SiteID: site.ID,
		Path:   "/another",
	}
	err = db.InsertPageview(pv2)
	if err != nil {
		t.Fatalf("insert pageview (minimal): %v", err)
	}
}

func TestMatchesIgnoredIP(t *testing.T) {
	db := setupTestDB(t)

	// Add an ignored IP
	if err := db.AddIgnoredIP("10.0.0.1"); err != nil {
		t.Fatalf("add ignored ip: %v", err)
	}

	// Test match
	matches, err := db.MatchesIgnoredIP("10.0.0.1")
	if err != nil {
		t.Fatalf("matches ignored ip: %v", err)
	}
	if !matches {
		t.Error("expected 10.0.0.1 to match ignored IP")
	}

	// Test no match
	matches, err = db.MatchesIgnoredIP("192.168.1.1")
	if err != nil {
		t.Fatalf("matches ignored ip (non-matching): %v", err)
	}
	if matches {
		t.Error("expected 192.168.1.1 to not match any ignored IP")
	}

	// Test with multiple ignored IPs
	if err := db.AddIgnoredIP("172.16.0.1"); err != nil {
		t.Fatalf("add second ignored ip: %v", err)
	}
	matches, err = db.MatchesIgnoredIP("172.16.0.1")
	if err != nil {
		t.Fatalf("matches second ignored ip: %v", err)
	}
	if !matches {
		t.Error("expected 172.16.0.1 to match ignored IP")
	}
}

func TestGetPageviewsBySite(t *testing.T) {
	db := setupTestDB(t)

	site, _ := db.AddSite("example.com")

	// No pageviews initially
	count, err := db.GetPageviewsBySite(site.ID)
	if err != nil {
		t.Fatalf("get pageviews by site: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 pageviews, got %d", count)
	}

	// Add some pageviews
	db.InsertPageview(models.PageviewInput{SiteID: site.ID, Path: "/"})
	db.InsertPageview(models.PageviewInput{SiteID: site.ID, Path: "/about"})
	db.InsertPageview(models.PageviewInput{SiteID: site.ID, Path: "/"})

	count, err = db.GetPageviewsBySite(site.ID)
	if err != nil {
		t.Fatalf("get pageviews by site: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 pageviews, got %d", count)
	}

	// Different site should have 0
	site2, _ := db.AddSite("other.com")
	count, _ = db.GetPageviewsBySite(site2.ID)
	if count != 0 {
		t.Errorf("expected 0 pageviews for other site, got %d", count)
	}
}

func TestGetAggregateStats(t *testing.T) {
	db := setupTestDB(t)

	site, _ := db.AddSite("example.com")

	// Add pageviews with different IPs
	db.InsertPageview(models.PageviewInput{SiteID: site.ID, IP: "1.2.3.4", Path: "/"})
	db.InsertPageview(models.PageviewInput{SiteID: site.ID, IP: "1.2.3.4", Path: "/"})
	db.InsertPageview(models.PageviewInput{SiteID: site.ID, IP: "5.6.7.8", Path: "/about"})

	now := time.Now()
	q := StatsQuery{
		SiteID: site.ID,
		From:   now.Add(-24 * time.Hour),
		To:     now.Add(time.Hour),
	}

	rows, err := db.GetAggregateStats(q)
	if err != nil {
		t.Fatalf("get aggregate stats: %v", err)
	}

	// Should have 2 groups: (1.2.3.4, /) with count 2, and (5.6.7.8, /about) with count 1
	if len(rows) != 2 {
		t.Fatalf("expected 2 stat rows, got %d", len(rows))
	}

	// Verify counts (order may vary)
	totalCount := 0
	for _, r := range rows {
		totalCount += r.Count
		if r.Site != "example.com" {
			t.Errorf("expected site example.com, got %s", r.Site)
		}
	}
	if totalCount != 3 {
		t.Errorf("expected total count 3, got %d", totalCount)
	}
}

func TestGetAggregateStatsAllSites(t *testing.T) {
	db := setupTestDB(t)

	site1, _ := db.AddSite("example.com")
	site2, _ := db.AddSite("other.com")

	db.InsertPageview(models.PageviewInput{SiteID: site1.ID, IP: "1.2.3.4", Path: "/"})
	db.InsertPageview(models.PageviewInput{SiteID: site2.ID, IP: "5.6.7.8", Path: "/"})

	now := time.Now()
	q := StatsQuery{
		SiteID: 0, // All sites
		From:   now.Add(-24 * time.Hour),
		To:     now.Add(time.Hour),
	}

	rows, err := db.GetAggregateStats(q)
	if err != nil {
		t.Fatalf("get aggregate stats all sites: %v", err)
	}

	if len(rows) != 2 {
		t.Errorf("expected 2 stat rows, got %d", len(rows))
	}
}

func TestGetVerboseStats(t *testing.T) {
	db := setupTestDB(t)

	site, _ := db.AddSite("example.com")

	db.InsertPageview(models.PageviewInput{
		SiteID:    site.ID,
		IP:        "1.2.3.4",
		Path:      "/",
		Referrer:  "https://google.com",
		Browser:   "Chrome",
		OS:        "Windows",
		Country:   "US",
		UserAgent: "Mozilla/5.0",
	})

	now := time.Now()
	q := StatsQuery{
		SiteID: site.ID,
		From:   now.Add(-24 * time.Hour),
		To:     now.Add(time.Hour),
	}

	rows, err := db.GetVerboseStats(q)
	if err != nil {
		t.Fatalf("get verbose stats: %v", err)
	}

	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}

	r := rows[0]
	if r.Site != "example.com" {
		t.Errorf("expected site example.com, got %s", r.Site)
	}
	if r.IP != "1.2.3.4" {
		t.Errorf("expected IP 1.2.3.4, got %s", r.IP)
	}
	if r.Browser != "Chrome" {
		t.Errorf("expected browser Chrome, got %s", r.Browser)
	}
	if r.OS != "Windows" {
		t.Errorf("expected OS Windows, got %s", r.OS)
	}
	if r.Path != "/" {
		t.Errorf("expected path /, got %s", r.Path)
	}
}

func TestGetVerboseStatsAllSites(t *testing.T) {
	db := setupTestDB(t)

	site1, _ := db.AddSite("example.com")
	site2, _ := db.AddSite("other.com")

	db.InsertPageview(models.PageviewInput{SiteID: site1.ID, Path: "/"})
	db.InsertPageview(models.PageviewInput{SiteID: site2.ID, Path: "/blog"})

	now := time.Now()
	q := StatsQuery{
		SiteID: 0,
		From:   now.Add(-24 * time.Hour),
		To:     now.Add(time.Hour),
	}

	rows, err := db.GetVerboseStats(q)
	if err != nil {
		t.Fatalf("get verbose stats all sites: %v", err)
	}

	if len(rows) != 2 {
		t.Errorf("expected 2 rows, got %d", len(rows))
	}
}

func TestAddSiteDuplicate(t *testing.T) {
	db := setupTestDB(t)

	_, err := db.AddSite("example.com")
	if err != nil {
		t.Fatalf("add site: %v", err)
	}

	// Duplicate should fail
	_, err = db.AddSite("example.com")
	if err == nil {
		t.Error("expected error for duplicate site")
	}
}

func TestAddIgnoredIPDuplicate(t *testing.T) {
	db := setupTestDB(t)

	err := db.AddIgnoredIP("1.2.3.4")
	if err != nil {
		t.Fatalf("add ignored ip: %v", err)
	}

	// Duplicate should not error (INSERT OR IGNORE)
	err = db.AddIgnoredIP("1.2.3.4")
	if err != nil {
		t.Fatalf("add duplicate ignored ip should not error: %v", err)
	}

	// Should still only have 1
	ips, _ := db.ListIgnoredIPs()
	if len(ips) != 1 {
		t.Errorf("expected 1 ignored ip after duplicate add, got %d", len(ips))
	}
}

func TestRemoveNonexistentSite(t *testing.T) {
	db := setupTestDB(t)

	// Removing non-existent site should not error
	err := db.RemoveSite("nonexistent.com")
	if err != nil {
		t.Fatalf("remove nonexistent site: %v", err)
	}
}

func TestRemoveNonexistentIP(t *testing.T) {
	db := setupTestDB(t)

	err := db.RemoveIgnoredIP("1.2.3.4")
	if err != nil {
		t.Fatalf("remove nonexistent ip: %v", err)
	}
}

func TestListEmptySites(t *testing.T) {
	db := setupTestDB(t)

	sites, err := db.ListSites()
	if err != nil {
		t.Fatalf("list sites: %v", err)
	}
	if len(sites) != 0 {
		t.Errorf("expected 0 sites, got %d", len(sites))
	}
}

func TestListEmptyIgnoredIPs(t *testing.T) {
	db := setupTestDB(t)

	ips, err := db.ListIgnoredIPs()
	if err != nil {
		t.Fatalf("list ignored ips: %v", err)
	}
	if len(ips) != 0 {
		t.Errorf("expected 0 ignored ips, got %d", len(ips))
	}
}

func TestGetSiteByDomainNotFound(t *testing.T) {
	db := setupTestDB(t)

	site, err := db.GetSiteByDomain("nonexistent.com")
	if err != nil {
		t.Fatalf("get site by domain: %v", err)
	}
	if site != nil {
		t.Error("expected nil for non-existent domain")
	}
}

func TestIsIPIgnoredNotFound(t *testing.T) {
	db := setupTestDB(t)

	ignored, err := db.IsIPIgnored("1.2.3.4")
	if err != nil {
		t.Fatalf("is ip ignored: %v", err)
	}
	if ignored {
		t.Error("expected IP to not be ignored")
	}
}

func TestRevokeNonexistentToken(t *testing.T) {
	db := setupTestDB(t)

	err := db.RevokeToken(999)
	if err == nil {
		t.Fatal("expected error for nonexistent token")
	}
	if !errors.Is(err, ErrTokenNotFound) {
		t.Errorf("expected ErrTokenNotFound, got %v", err)
	}
}
