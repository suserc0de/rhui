package runtime

import (
	"html"
	"strings"
)

// EscapeHTML escapes standard HTML special characters (&, <, >, ", ').
func EscapeHTML(s string) string {
	return html.EscapeString(s)
}

// SanitizeURL checks if a URL is safe for href and src attributes, preventing javascript: pseudo-protocol attacks.
func SanitizeURL(urlStr string) string {
	trimmed := strings.TrimSpace(urlStr)
	lower := strings.ToLower(trimmed)

	// Block dangerous URI schemes
	if strings.HasPrefix(lower, "javascript:") ||
		strings.HasPrefix(lower, "vbscript:") ||
		strings.HasPrefix(lower, "data:text/html") {
		return "#unsafe-url"
	}

	return EscapeHTML(trimmed)
}
