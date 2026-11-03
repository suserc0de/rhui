package runtime

import (
	"html"
	"strings"
)

// EscapeHTML escapes standard HTML special characters (&, <, >, ", ').
// Fast path: returns original string with 0 allocations if no escaping is needed.
func EscapeHTML(s string) string {
	if strings.IndexAny(s, "&<>'\"") == -1 {
		return s
	}
	return html.EscapeString(s)
}

// EscapeAttribute escapes HTML attribute values preventing quote breakouts and event injections.
// Fast path: returns original string with 0 allocations if no quotes or sensitive characters exist.
func EscapeAttribute(s string) string {
	if strings.IndexAny(s, "&<>'\"") == -1 {
		return s
	}
	escaped := html.EscapeString(s)
	// Extra safety replacement for attribute delimiters
	escaped = strings.ReplaceAll(escaped, `"`, "&#34;")
	escaped = strings.ReplaceAll(escaped, `'`, "&#39;")
	return escaped
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
