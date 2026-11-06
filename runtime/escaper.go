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

// IsValidAttributeName validates HTML attribute names according to W3C HTML5 recommendations.
// Disallows whitespace, control characters, quotes, equals, slashes, and characters that can break attribute boundaries.
// Security: Automatically rejects all inline event handlers (attributes starting with 'on', e.g. onclick, onerror, onload).
func IsValidAttributeName(name string) bool {
	if len(name) == 0 {
		return false
	}

	// Security Hardening: Block all inline event handlers (case-insensitive prefix 'on')
	lower := strings.ToLower(name)
	if strings.HasPrefix(lower, "on") {
		return false
	}

	for i := 0; i < len(name); i++ {
		c := name[i]
		// Disallow control characters, whitespace, quotes, slashes, equals, angle brackets, and non-ASCII
		if c <= 0x20 || c == 0x7F || c == '"' || c == '\'' || c == '>' || c == '/' || c == '=' || c > 0x7E {
			return false
		}
	}
	return true
}

// SanitizeURL validates and sanitizes URLs for href and src attributes.
// Security: Uses a strict allowlist approach. Only safe schemes (http, https, mailto, tel, sms)
// and relative/fragment URLs (#, /, ./, ../, ?) are allowed.
// All pseudo-protocols (javascript:, vbscript:, data:, blob:, file:) and protocol-relative URLs (//)
// are strictly neutralized to prevent XSS and open redirect vectors.
func SanitizeURL(urlStr string) string {
	trimmed := strings.TrimSpace(urlStr)
	if trimmed == "" {
		return ""
	}

	// Strip internal whitespace and control characters that attackers use for obfuscation
	cleanURL := strings.Map(func(r rune) rune {
		if r <= 0x20 || r == 0x7F {
			return -1
		}
		return r
	}, trimmed)

	if cleanURL == "" {
		return "#unsafe-url"
	}

	// Reject protocol-relative URLs (e.g. "//evil.com") which can lead to open redirects or unexpected protocol inheritances
	if strings.HasPrefix(cleanURL, "//") {
		return "#unsafe-url"
	}

	// Safe relative paths and fragments
	if strings.HasPrefix(cleanURL, "/") ||
		strings.HasPrefix(cleanURL, "#") ||
		strings.HasPrefix(cleanURL, "./") ||
		strings.HasPrefix(cleanURL, "../") ||
		strings.HasPrefix(cleanURL, "?") {
		return EscapeHTML(trimmed)
	}

	lower := strings.ToLower(cleanURL)

	// Safe scheme allowlist
	if strings.HasPrefix(lower, "http://") ||
		strings.HasPrefix(lower, "https://") ||
		strings.HasPrefix(lower, "mailto:") ||
		strings.HasPrefix(lower, "tel:") ||
		strings.HasPrefix(lower, "sms:") {
		return EscapeHTML(trimmed)
	}

	// If the URL contains a colon before any path separator, it is an untrusted or dangerous scheme (e.g., javascript:, data:, vbscript:, file:, blob:)
	colonIdx := strings.IndexByte(lower, ':')
	if colonIdx != -1 {
		return "#unsafe-url"
	}

	// Standard relative path without leading slash (e.g. "profile/view", "image.png")
	return EscapeHTML(trimmed)
}
