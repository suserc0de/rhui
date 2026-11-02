package tests

import (
	"bytes"
	"strings"
	"sync"
	"testing"

	"github.com/suserc0de/rhui"
)

func TestSecurityXSSPrevention(t *testing.T) {
	engine, err := rhui.New()
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	tmpl := `<div class="user-comment">
  <p>{{ MaliciousComment }}</p>
  <a :href="MaliciousURL">Visit</a>
  <span :title="MaliciousAttr">Hover</span>
</div>`

	if err := engine.ParseTemplate("pages/comment", tmpl); err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	data := map[string]any{
		"MaliciousComment": "<script>alert('pwned')</script>",
		"MaliciousURL":     "javascript:alert('xss')",
		"MaliciousAttr":    `" onfocus="alert('xss')`,
	}

	out, err := engine.RenderString("pages/comment", data)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	// 1. Script tag must be escaped
	if strings.Contains(out, "<script>") {
		t.Errorf("XSS vulnerability: raw <script> found in output: %s", out)
	}
	if !strings.Contains(out, "&lt;script&gt;alert(&#39;pwned&#39;)&lt;/script&gt;") {
		t.Errorf("expected escaped HTML comment, got: %s", out)
	}

	// 2. javascript: URI scheme must be sanitized
	if strings.Contains(out, "href=\"javascript:") {
		t.Errorf("XSS vulnerability: dangerous javascript: URI scheme allowed: %s", out)
	}
	if !strings.Contains(out, "href=\"#unsafe-url\"") {
		t.Errorf("expected sanitized unsafe-url, got: %s", out)
	}

	// 3. Double quotes in attribute must be escaped
	if strings.Contains(out, "title=\"\" onfocus=") {
		t.Errorf("Attribute injection vulnerability: raw quotes broke out of attribute: %s", out)
	}
}

func TestConcurrentRenderingThreadSafety(t *testing.T) {
	engine, err := rhui.New()
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	tmpl := `<div><h3>{{ Title }}</h3><p>User #{{ ID }}</p></div>`
	if err := engine.ParseTemplate("item", tmpl); err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	var wg sync.WaitGroup
	workers := 100
	iterations := 50

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				var buf bytes.Buffer
				err := engine.Render(&buf, "item", map[string]any{
					"Title": "Concurrent Worker",
					"ID":    workerID*1000 + j,
				})
				if err != nil {
					t.Errorf("concurrent render failed: %v", err)
					return
				}
				if !strings.Contains(buf.String(), "Concurrent Worker") {
					t.Errorf("unexpected output in concurrency test")
				}
			}
		}(i)
	}

	wg.Wait()
}

func BenchmarkEngineRender(b *testing.B) {
	engine, _ := rhui.New()
	tmpl := `<div class="card"><h2>{{ Title }}</h2><p>Count: {{ Count }}</p></div>`
	_ = engine.ParseTemplate("bench", tmpl)

	data := map[string]any{
		"Title": "Performance Benchmark",
		"Count": 999,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		_ = engine.Render(&buf, "bench", data)
	}
}
