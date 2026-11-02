package tests

import (
	"bytes"
	htmltmpl "html/template"
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

// BenchmarkRHUI_Render measures RHUI's AST evaluation + sync.Pool speed
func BenchmarkRHUI_Render(b *testing.B) {
	engine, err := rhui.New()
	if err != nil {
		b.Fatalf("failed to create engine: %v", err)
	}

	tmpl := `<div class="container">
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

	if err := engine.ParseTemplate("bench/table", tmpl); err != nil {
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
	rawTmpl := `<div class="container">
  <h1>{{ .Title }}</h1>
  <table class="table">
    {{ range .Items }}
      <tr>
        <td>{{ .ID }}</td>
        <td>{{ .Name }}</td>
        <td>Rp {{ .Price }}</td>
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

	tmpl, err := htmltmpl.New("table").Parse(rawTmpl)
	if err != nil {
		b.Fatalf("failed to parse stdlib template: %v", err)
	}

	data := getBenchmarkData()
	var buf bytes.Buffer

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		if err := tmpl.Execute(&buf, data); err != nil {
			b.Fatalf("Execute failed: %v", err)
		}
	}
}
