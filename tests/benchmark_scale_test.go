package tests

import (
	"bytes"
	"fmt"
	htmltmpl "html/template"
	"testing"

	"github.com/suserc0de/rhui"
)

func generateBenchmarkItems(count int) map[string]any {
	items := make([]BenchmarkItem, count)
	for i := 0; i < count; i++ {
		items[i] = BenchmarkItem{
			ID:       i + 1,
			Name:     "Cloud Subscription Instance",
			Price:    350000,
			IsActive: i%2 == 0,
		}
	}
	return map[string]any{
		"Title": fmt.Sprintf("Billing for %d Customers", count),
		"Items": items,
	}
}

const benchmarkRHUITableTmpl = `<div class="container">
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

const benchmarkStdlibTableTmpl = `<div class="container">
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

// BenchmarkRHUI_Table_100Rows measures RHUI rendering a 100-row table
func BenchmarkRHUI_Table_100Rows(b *testing.B) {
	engine, err := rhui.New()
	if err != nil {
		b.Fatal(err)
	}
	if err := engine.ParseTemplate("bench/table100", benchmarkRHUITableTmpl); err != nil {
		b.Fatal(err)
	}
	data := generateBenchmarkItems(100)
	var buf discardWriter

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := engine.Render(&buf, "bench/table100", data); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkStdlib_Table_100Rows measures html/template rendering a 100-row table
func BenchmarkStdlib_Table_100Rows(b *testing.B) {
	tmpl, err := htmltmpl.New("table100").Parse(benchmarkStdlibTableTmpl)
	if err != nil {
		b.Fatal(err)
	}
	data := generateBenchmarkItems(100)
	var buf bytes.Buffer

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		if err := tmpl.Execute(&buf, data); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkRHUI_Table_1000Rows measures RHUI rendering a 1000-row table
func BenchmarkRHUI_Table_1000Rows(b *testing.B) {
	engine, err := rhui.New()
	if err != nil {
		b.Fatal(err)
	}
	if err := engine.ParseTemplate("bench/table1000", benchmarkRHUITableTmpl); err != nil {
		b.Fatal(err)
	}
	data := generateBenchmarkItems(1000)
	var buf discardWriter

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := engine.Render(&buf, "bench/table1000", data); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkStdlib_Table_1000Rows measures html/template rendering a 1000-row table
func BenchmarkStdlib_Table_1000Rows(b *testing.B) {
	tmpl, err := htmltmpl.New("table1000").Parse(benchmarkStdlibTableTmpl)
	if err != nil {
		b.Fatal(err)
	}
	data := generateBenchmarkItems(1000)
	var buf bytes.Buffer

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		if err := tmpl.Execute(&buf, data); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkRHUI_NestedComponents measures multi-level nested component trees
func BenchmarkRHUI_NestedComponents(b *testing.B) {
	engine, err := rhui.New()
	if err != nil {
		b.Fatal(err)
	}

	btnTmpl := `<component name="Button">
  <props>
    <prop name="variant" default="primary" />
  </props>
  <button class="btn btn-{{ variant }}"><slot /></button>
</component>`

	cardTmpl := `<component name="Card">
  <props>
    <prop name="title" default="Title" />
  </props>
  <div class="card">
    <div class="card-header">{{ title }}</div>
    <div class="card-body"><slot /></div>
    <div class="card-actions"><slot:actions /></div>
  </div>
</component>`

	layoutTmpl := `<component name="Layout">
  <div class="app-layout">
    <header class="navbar"><slot:nav /></header>
    <main><slot /></main>
  </div>
</component>`

	pageTmpl := `<ui:Layout>
  <slot:nav>
    <nav><span>Dashboard</span></nav>
  </slot:nav>
  <ui:Card title="Project Statistics">
    <p>CPU and memory utilization metrics</p>
    <slot:actions>
      <ui:Button variant="secondary">Cancel</ui:Button>
      <ui:Button variant="primary">Submit Report</ui:Button>
    </slot:actions>
  </ui:Card>
</ui:Layout>`

	if err := engine.ParseTemplate("components/Button", btnTmpl); err != nil {
		b.Fatal(err)
	}
	if err := engine.ParseTemplate("components/Card", cardTmpl); err != nil {
		b.Fatal(err)
	}
	if err := engine.ParseTemplate("components/Layout", layoutTmpl); err != nil {
		b.Fatal(err)
	}
	if err := engine.ParseTemplate("pages/nested", pageTmpl); err != nil {
		b.Fatal(err)
	}

	data := map[string]any{"User": "Admin"}
	var buf discardWriter

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := engine.Render(&buf, "pages/nested", data); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkRHUI_ParallelConcurrency tests concurrent requests scaling across multiple goroutines
func BenchmarkRHUI_ParallelConcurrency(b *testing.B) {
	engine, err := rhui.New()
	if err != nil {
		b.Fatal(err)
	}
	if err := engine.ParseTemplate("bench/table50", benchmarkRHUITableTmpl); err != nil {
		b.Fatal(err)
	}
	data := generateBenchmarkItems(50)

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		var buf discardWriter
		for pb.Next() {
			if err := engine.Render(&buf, "bench/table50", data); err != nil {
				b.Fatal(err)
			}
		}
	})
}
