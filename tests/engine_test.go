package tests

import (
	"strings"
	"testing"

	"github.com/suserc0de/rhui"
)

func TestEngineEndToEnd(t *testing.T) {
	engine, err := rhui.New()
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	// 1. Register a Card component
	cardTmpl := `<component name="Card">
  <props>
    <prop name="title" default="No Title" />
  </props>
  <div class="card">
    <div class="header"><h3>{{ title }}</h3></div>
    <div class="body"><slot /></div>
  </div>
</component>`
	if err := engine.ParseTemplate("components/Card", cardTmpl); err != nil {
		t.Fatalf("failed to register card component: %v", err)
	}

	// 2. Register page template
	pageTmpl := `<div class="container">
  <h1>{{ Heading }}</h1>
  <ui:Card title="Stats">
    <p>Total items: {{ Count }}</p>
  </ui:Card>
</div>`
	if err := engine.ParseTemplate("pages/stats", pageTmpl); err != nil {
		t.Fatalf("failed to register page template: %v", err)
	}

	// 3. Render
	out, err := engine.RenderString("pages/stats", map[string]any{
		"Heading": "Analytics",
		"Count":   100,
	})
	if err != nil {
		t.Fatalf("engine.RenderString failed: %v", err)
	}

	if !strings.Contains(out, "<h1>Analytics</h1>") {
		t.Errorf("expected heading Analytics, got: %s", out)
	}
	if !strings.Contains(out, "<h3>Stats</h3>") {
		t.Errorf("expected card title Stats, got: %s", out)
	}
	if !strings.Contains(out, "<p>Total items: 100</p>") {
		t.Errorf("expected body with total items 100, got: %s", out)
	}
}
