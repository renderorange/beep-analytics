package ua

import "testing"

func TestParse(t *testing.T) {
	tests := []struct {
		ua      string
		browser string
		os      string
	}{
		{
			ua:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			browser: "Chrome 120",
			os:      "Windows 10",
		},
		{
			ua:      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Safari/605.1.15",
			browser: "Safari 17",
			os:      "macOS 10",
		},
		{
			ua:      "Mozilla/5.0 (X11; Linux x86_64; rv:121.0) Gecko/20100101 Firefox/121.0",
			browser: "Firefox 121",
			os:      "Linux",
		},
		{
			ua:      "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
			browser: "Safari 17",
			os:      "iOS 17",
		},
		{
			ua:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0",
			browser: "Edge 120",
			os:      "Windows 10",
		},
		{
			ua:      "Mozilla/5.0 (Linux; Android 14) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.6099.230 Mobile Safari/537.36",
			browser: "Chrome 120",
			os:      "Android 14",
		},
		{
			ua:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 OPR/106.0.0.0",
			browser: "Opera 106",
			os:      "Windows 10",
		},
		{
			ua:      "Mozilla/5.0 (iPad; CPU OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
			browser: "Safari 17",
			os:      "iPadOS 17",
		},
		{
			ua:      "Mozilla/5.0 (X11; CrOS x86_64 14541.0.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			browser: "Chrome 120",
			os:      "ChromeOS",
		},
		{
			ua:      "curl/7.68.0",
			browser: "Other",
			os:      "Other",
		},
		{
			ua:      "Googlebot/2.1 (+http://www.google.com/bot.html)",
			browser: "Other",
			os:      "Other",
		},
		{
			ua:      "",
			browser: "Other",
			os:      "Other",
		},
		{
			ua:      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 edge/",
			browser: "Edge",
			os:      "macOS 10",
		},
		{
			ua:      "Mozilla/5.0 (Linux; Android 14) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.6099.230 Mobile Safari/537.36 CriOS/120.0.6099.230",
			browser: "Chrome 120",
			os:      "Android 14",
		},
		{
			ua:      "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) FxiOS/121.0 Mobile/15E148 Safari/605.1.15",
			browser: "Firefox 121",
			os:      "iOS 17",
		},
		{
			ua:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Opera/106.0.0.0",
			browser: "Opera 106",
			os:      "Windows 10",
		},
	}

	for _, tt := range tests {
		info := Parse(tt.ua)
		if info.Browser != tt.browser {
			t.Errorf("UA %q: expected browser %q, got %q", tt.ua, tt.browser, info.Browser)
		}
		if info.OS != tt.os {
			t.Errorf("UA %q: expected OS %q, got %q", tt.ua, tt.os, info.OS)
		}
	}
}

func TestParseBrowser(t *testing.T) {
	tests := []struct {
		ua       string
		expected string
	}{
		{"chrome/120", "Chrome"},
		{"firefox/121", "Firefox"},
		{"safari/605", "Safari"},
		{"edg/120", "Edge"},
		{"edge/120", "Edge"},
		{"opera/106", "Opera"},
		{"opr/106", "Opera"},
		{"crios/120", "Chrome"},
		{"fxios/121", "Firefox"},
		{"unknown", "Other"},
	}

	for _, tt := range tests {
		result := parseBrowser(tt.ua)
		if result != tt.expected {
			t.Errorf("parseBrowser(%q) = %q, want %q", tt.ua, result, tt.expected)
		}
	}
}

func TestParseOS(t *testing.T) {
	tests := []struct {
		ua       string
		expected string
	}{
		{"android 14", "Android"},
		{"iphone", "iOS"},
		{"ipod", "iOS"},
		{"ipad", "iPadOS"},
		{"windows nt", "Windows"},
		{"mac os x", "macOS"},
		{"macos", "macOS"},
		{"linux", "Linux"},
		{"chrome os", "ChromeOS"},
		{"cros", "ChromeOS"},
		{"unknown", "Other"},
	}

	for _, tt := range tests {
		result := parseOS(tt.ua)
		if result != tt.expected {
			t.Errorf("parseOS(%q) = %q, want %q", tt.ua, result, tt.expected)
		}
	}
}
