package ua

import "strings"

type Info struct {
	Browser string
	OS      string
}

func Parse(userAgent string) Info {
	ua := strings.ToLower(userAgent)
	return Info{
		Browser: parseBrowser(ua),
		OS:      parseOS(ua),
	}
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
