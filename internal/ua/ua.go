package ua

import "strings"

type Info struct {
	Browser string
	OS      string
}

func Parse(userAgent string) Info {
	ua := strings.ToLower(userAgent)
	browser := parseBrowser(ua)
	os := parseOS(ua)

	if v := browserVersion(ua, browser); v != "" {
		browser = browser + " " + v
	}
	if v := osVersion(ua, os); v != "" {
		os = os + " " + v
	}

	return Info{Browser: browser, OS: os}
}

func extractMajor(ua, prefix string) string {
	idx := strings.Index(ua, prefix)
	if idx < 0 {
		return ""
	}
	var s strings.Builder
	for _, c := range ua[idx+len(prefix):] {
		if c >= '0' && c <= '9' {
			s.WriteRune(c)
		} else if s.Len() > 0 {
			break
		}
	}
	return s.String()
}

func browserVersion(ua, browser string) string {
	switch browser {
	case "Chrome":
		if v := extractMajor(ua, "chrome/"); v != "" {
			return v
		}
		return extractMajor(ua, "crios/")
	case "Firefox":
		if v := extractMajor(ua, "firefox/"); v != "" {
			return v
		}
		return extractMajor(ua, "fxios/")
	case "Safari":
		return extractMajor(ua, "version/")
	case "Edge":
		if v := extractMajor(ua, "edg/"); v != "" {
			return v
		}
		return extractMajor(ua, "edge/")
	case "Opera":
		if v := extractMajor(ua, "opr/"); v != "" {
			return v
		}
		return extractMajor(ua, "opera/")
	}
	return ""
}

func osVersion(ua, os string) string {
	switch os {
	case "Windows":
		return extractMajor(ua, "windows nt ")
	case "macOS":
		if v := extractMajor(ua, "mac os x "); v != "" {
			return v
		}
		return extractMajor(ua, "macos ")
	case "iOS":
		if v := extractMajor(ua, "iphone os "); v != "" {
			return v
		}
		if v := extractMajor(ua, "ipod os "); v != "" {
			return v
		}
		return extractMajor(ua, "cpu os ")
	case "iPadOS":
		if v := extractMajor(ua, "ipad os "); v != "" {
			return v
		}
		return extractMajor(ua, "cpu os ")
	case "Android":
		return extractMajor(ua, "android ")
	}
	return ""
}

func parseBrowser(ua string) string {
	if strings.Contains(ua, "edg/") || strings.Contains(ua, "edge/") {
		return "Edge"
	}
	if strings.Contains(ua, "opera") || strings.Contains(ua, "opr/") {
		return "Opera"
	}
	if strings.Contains(ua, "chrome") || strings.Contains(ua, "crios") {
		return "Chrome"
	}
	if strings.Contains(ua, "firefox") || strings.Contains(ua, "fxios") {
		return "Firefox"
	}
	if strings.Contains(ua, "safari") && !strings.Contains(ua, "chrome") {
		return "Safari"
	}
	return "Other"
}

func parseOS(ua string) string {
	if strings.Contains(ua, "android") {
		return "Android"
	}
	if strings.Contains(ua, "iphone") || strings.Contains(ua, "ipod") {
		return "iOS"
	}
	if strings.Contains(ua, "ipad") {
		return "iPadOS"
	}
	if strings.Contains(ua, "windows") {
		return "Windows"
	}
	if strings.Contains(ua, "mac os x") || strings.Contains(ua, "macos") {
		return "macOS"
	}
	if strings.Contains(ua, "linux") && !strings.Contains(ua, "android") {
		return "Linux"
	}
	if strings.Contains(ua, "chrome os") || strings.Contains(ua, "cros") {
		return "ChromeOS"
	}
	return "Other"
}
