package tests

import (
	"strings"
	"testing"

	"github.com/suserc0de/rhui/runtime"
)

func FuzzXSSPenetration(f *testing.F) {
	seeds := []string{
		`<script>alert(1)</script>`,
		`"><svg onload=alert(1)>`,
		`' onfocus='alert(1)`,
		`javascript:alert(1)`,
		`java\0script:evil()`,
		`<iframe src="data:text/html,<script>alert(1)</script>">`,
		`<a href="javascript:void(0)">link</a>`,
		`foo" onmouseover="alert(1)`,
		`"><img src=x onerror=alert(1)>`,
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, payload string) {
		// 1. Escaping must never panic
		escapedHTML := runtime.EscapeHTML(payload)
		escapedAttr := runtime.EscapeAttribute(payload)
		sanitizedURL := runtime.SanitizeURL(payload)

		// 2. Dangerous raw HTML brackets must never pass through EscapeHTML
		if strings.Contains(escapedHTML, "<script>") || strings.Contains(escapedHTML, "</script>") {
			t.Fatalf("Unescaped script tag in HTML escape: %s", escapedHTML)
		}

		// 3. Double quotes must never exist unescaped in EscapeAttribute
		if strings.Contains(escapedAttr, `"`) {
			t.Fatalf("Unescaped quote in attribute escape: %s", escapedAttr)
		}

		// 4. Sanitized URLs must never allow active javascript/vbscript protocols
		lowerURL := strings.ToLower(sanitizedURL)
		if strings.HasPrefix(lowerURL, "javascript:") || strings.HasPrefix(lowerURL, "vbscript:") {
			t.Fatalf("Dangerous script URL allowed: %s", sanitizedURL)
		}
	})
}
