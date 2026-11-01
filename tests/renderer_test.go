package tests

import (
	"bytes"
	"strings"
	"testing"

	"github.com/coderhaqi/rhui/ast"
	"github.com/coderhaqi/rhui/compiler"
	"github.com/coderhaqi/rhui/parser"
	"github.com/coderhaqi/rhui/runtime"
)

func TestRendererBasicElementAndInterpolation(t *testing.T) {
	tmpl := `<div class="profile"><h1>{{ Title }}</h1><p>Welcome, {{ User.Name }}!</p></div>`
	lex := parser.NewLexer("page.rhui", tmpl)
	p := parser.NewParser(lex)
	root, err := p.Parse()
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	scope := runtime.NewScope(map[string]any{
		"Title": "Dashboard",
		"User": map[string]any{
			"Name": "Haqi",
		},
	})

	renderer := runtime.NewRenderer(nil, nil)
	var buf bytes.Buffer
	if err := renderer.RenderNode(&buf, root, scope); err != nil {
		t.Fatalf("render failed: %v", err)
	}

	expected := `<div class="profile"><h1>Dashboard</h1><p>Welcome, Haqi!</p></div>`
	if buf.String() != expected {
		t.Fatalf("expected:\n%s\ngot:\n%s", expected, buf.String())
	}
}

func TestRendererBooleanAttributeHandling(t *testing.T) {
	tmpl := `<button type="button" :disabled="IsDisabled">Click Me</button>`
	lex := parser.NewLexer("btn.rhui", tmpl)
	p := parser.NewParser(lex)
	root, _ := p.Parse()

	renderer := runtime.NewRenderer(nil, nil)

	// Case 1: IsDisabled = false -> disabled attribute must be omitted entirely
	var buf1 bytes.Buffer
	scope1 := runtime.NewScope(map[string]any{"IsDisabled": false})
	_ = renderer.RenderNode(&buf1, root, scope1)
	if strings.Contains(buf1.String(), "disabled") {
		t.Fatalf("expected 'disabled' attribute to be omitted when false, got: %s", buf1.String())
	}

	// Case 2: IsDisabled = true -> disabled attribute rendered
	var buf2 bytes.Buffer
	scope2 := runtime.NewScope(map[string]any{"IsDisabled": true})
	_ = renderer.RenderNode(&buf2, root, scope2)
	if !strings.Contains(buf2.String(), "disabled") {
		t.Fatalf("expected 'disabled' attribute to be present when true, got: %s", buf2.String())
	}
}

func TestRendererComponentWithSlots(t *testing.T) {
	// 1. Define Component "Card"
	cardTmpl := `<component name="Card">
  <props>
    <prop name="title" default="Default Title" />
  </props>
  <div class="card">
    <div class="card-header"><h3>{{ title }}</h3></div>
    <div class="card-body"><slot /></div>
    <div class="card-footer"><slot:footer /></div>
  </div>
</component>`

	p1 := parser.NewParser(parser.NewLexer("Card.rhui", cardTmpl))
	cardRoot, err := p1.Parse()
	if err != nil {
		t.Fatalf("failed to parse card component: %v", err)
	}

	registry := compiler.NewRegistry()
	for _, child := range cardRoot.Children {
		if compDef, ok := child.(*ast.ComponentDefNode); ok {
			registry.Register(compDef.Name, compDef)
		}
	}

	// 2. Page invoking <ui:Card>
	pageTmpl := `<ui:Card title="User List">
  <p>Default slot body content: {{ Count }} items</p>
  <slot:footer>
    <button>Close</button>
  </slot:footer>
</ui:Card>`

	p2 := parser.NewParser(parser.NewLexer("page.rhui", pageTmpl))
	pageRoot, err := p2.Parse()
	if err != nil {
		t.Fatalf("failed to parse page: %v", err)
	}

	renderer := runtime.NewRenderer(registry, nil)
	scope := runtime.NewScope(map[string]any{
		"Count": 42,
	})

	var buf bytes.Buffer
	if err := renderer.RenderNode(&buf, pageRoot, scope); err != nil {
		t.Fatalf("failed to render component: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "<h3>User List</h3>") {
		t.Errorf("expected prop 'User List' in header, got: %s", out)
	}
	if !strings.Contains(out, "Default slot body content: 42 items") {
		t.Errorf("expected default slot with parent scope data '42 items', got: %s", out)
	}
	if !strings.Contains(out, "<button>Close</button>") {
		t.Errorf("expected named slot footer to render button, got: %s", out)
	}
}
