package tests

import (
	"bytes"
	"fmt"
	htmltmpl "html/template"
	"strconv"
	"strings"
	"testing"

	"github.com/suserc0de/rhui"
)

type BenchmarkItem struct {
	ID       int
	Name     string
	Price    int
	IsActive bool
}

func getBenchmarkData() map[string]any {
	items := make([]BenchmarkItem, 50)
	for i := 0; i < 50; i++ {
		items[i] = BenchmarkItem{
			ID:       i + 1,
			Name:     "Cloud Server Subscription",
			Price:    350000,
			IsActive: i%2 == 0,
		}
	}
	return map[string]any{
		"Title": "SaaS Billing Overview",
		"Items": items,
	}
}

// formatRupiahHelper formats integer to Indonesian Rupiah currency format "Rp 350.000",
// identical to RHUI's built-in `| rupiah` pipe filter.
func formatRupiahHelper(n int) string {
	in := strconv.Itoa(n)
	var out []byte
	l := len(in)
	for i := 0; i < l; i++ {
		if i > 0 && (l-i)%3 == 0 {
			out = append(out, '.')
		}
		out = append(out, in[i])
	}
	return fmt.Sprintf("Rp %s", string(out))
}

const (
	rhuiHeadToHeadTmpl = `<div class="container">
  <h1>{{ Title }}</h1>
  <table class="table">
    <each items="Items" as="item">
      <tr>
        <td>{{ item.ID }}</td>
        <td>{{ item.Name }}</td>
        <td>{{ item.Price | rupiah }}</td>
        <td>
          <if test="item.IsActive == true">
            <span class="badge-active">Active</span>
          </if>
          <else>
            <span class="badge-disabled">Disabled</span>
          </else>
        </td>
      </tr>
    </each>
  </table>
</div>`

	stdlibHeadToHeadTmpl = `<div class="container">
  <h1>{{ .Title }}</h1>
  <table class="table">
    {{ range .Items }}
      <tr>
        <td>{{ .ID }}</td>
        <td>{{ .Name }}</td>
        <td>{{ rupiah .Price }}</td>
        <td>
          {{ if .IsActive }}
            <span class="badge-active">Active</span>
          {{ else }}
            <span class="badge-disabled">Disabled</span>
          {{ end }}
        </td>
      </tr>
    {{ end }}
  </table>
</div>`
)

// normalizeHTML removes non-significant whitespace between tags to allow strict semantic HTML comparison
func normalizeHTML(s string) string {
	lines := strings.Split(s, "\n")
	var cleaned []string
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if trimmed != "" {
			cleaned = append(cleaned, trimmed)
		}
	}
	return strings.Join(cleaned, "")
}

// TestBenchmarkOutputParity proves that RHUI and html/template render 100% SEMANTICALLY IDENTICAL
// HTML output for the benchmark test case, ensuring zero bias and 100% fair head-to-head comparison.
func TestBenchmarkOutputParity(t *testing.T) {
	data := getBenchmarkData()

	// 1. Render with RHUI
	engine, err := rhui.New()
	if err != nil {
		t.Fatalf("failed to create RHUI engine: %v", err)
	}
	if err := engine.ParseTemplate("bench/table", rhuiHeadToHeadTmpl); err != nil {
		t.Fatalf("failed to parse RHUI template: %v", err)
	}
	var rhuiBuf bytes.Buffer
	if err := engine.Render(&rhuiBuf, "bench/table", data); err != nil {
		t.Fatalf("RHUI render failed: %v", err)
	}

	// 2. Render with html/template
	funcMap := htmltmpl.FuncMap{
		"rupiah": formatRupiahHelper,
	}
	stdlibTmpl, err := htmltmpl.New("table").Funcs(funcMap).Parse(stdlibHeadToHeadTmpl)
	if err != nil {
		t.Fatalf("failed to parse stdlib template: %v", err)
	}
	var stdlibBuf bytes.Buffer
	if err := stdlibTmpl.Execute(&stdlibBuf, data); err != nil {
		t.Fatalf("stdlib render failed: %v", err)
	}

	// 3. Verify 100% parity
	rhuiNorm := normalizeHTML(rhuiBuf.String())
	stdlibNorm := normalizeHTML(stdlibBuf.String())

	if rhuiNorm != stdlibNorm {
		t.Fatalf("Benchmark output disparity detected!\nRHUI Output:\n%s\nStdlib Output:\n%s", rhuiNorm, stdlibNorm)
	}
}

// BenchmarkRHUI_Render measures RHUI's AST evaluation + sync.Pool speed
func BenchmarkRHUI_Render(b *testing.B) {
	engine, err := rhui.New()
	if err != nil {
		b.Fatalf("failed to create engine: %v", err)
	}

	if err := engine.ParseTemplate("bench/table", rhuiHeadToHeadTmpl); err != nil {
		b.Fatalf("ParseTemplate failed: %v", err)
	}

	data := getBenchmarkData()
	var buf discardWriter

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if err := engine.Render(&buf, "bench/table", data); err != nil {
			b.Fatalf("Render failed: %v", err)
		}
	}
}

// BenchmarkStdlib_HtmlTemplate measures Go standard library html/template
func BenchmarkStdlib_HtmlTemplate(b *testing.B) {
	funcMap := htmltmpl.FuncMap{
		"rupiah": formatRupiahHelper,
	}
	tmpl, err := htmltmpl.New("table").Funcs(funcMap).Parse(stdlibHeadToHeadTmpl)
	if err != nil {
		b.Fatalf("failed to parse stdlib template: %v", err)
	}

	data := getBenchmarkData()
	var buf discardWriter

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if err := tmpl.Execute(&buf, data); err != nil {
			b.Fatalf("Execute failed: %v", err)
		}
	}
}
