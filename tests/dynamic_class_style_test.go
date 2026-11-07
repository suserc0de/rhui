package tests

import (
	"strings"
	"testing"

	"github.com/suserc0de/rhui"
)

func TestDynamicClassAndStyleBindings(t *testing.T) {
	engine, err := rhui.New()
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	// 1. :class with map[string]bool
	tmpl1 := `<button :class="ClassMap">Button</button>`
	if err := engine.ParseTemplate("tmpl1", tmpl1); err != nil {
		t.Fatalf("failed to parse tmpl1: %v", err)
	}

	out1, err := engine.RenderString("tmpl1", map[string]any{
		"ClassMap": map[string]bool{
			"btn":          true,
			"btn-primary":  true,
			"btn-disabled": false,
		},
	})
	if err != nil {
		t.Fatalf("render tmpl1 failed: %v", err)
	}
	if !strings.Contains(out1, "btn") || !strings.Contains(out1, "btn-primary") {
		t.Errorf("expected btn and btn-primary in out1, got: %s", out1)
	}
	if strings.Contains(out1, "btn-disabled") {
		t.Errorf("btn-disabled should be omitted, got: %s", out1)
	}

	// 2. :class with []string
	tmpl2 := `<div :class="ClassList">Card</div>`
	if err := engine.ParseTemplate("tmpl2", tmpl2); err != nil {
		t.Fatalf("failed to parse tmpl2: %v", err)
	}

	out2, err := engine.RenderString("tmpl2", map[string]any{
		"ClassList": []string{"card", "shadow-sm", "p-4"},
	})
	if err != nil {
		t.Fatalf("render tmpl2 failed: %v", err)
	}
	if !strings.Contains(out2, `class="card shadow-sm p-4"`) {
		t.Errorf("expected joined class list, got: %s", out2)
	}

	// 3. :style with map[string]string and map[string]any
	tmpl3 := `<div :style="StyleMap">Styled Box</div>`
	if err := engine.ParseTemplate("tmpl3", tmpl3); err != nil {
		t.Fatalf("failed to parse tmpl3: %v", err)
	}

	out3, err := engine.RenderString("tmpl3", map[string]any{
		"StyleMap": map[string]string{
			"background-color": "red",
			"font-size":        "16px",
		},
	})
	if err != nil {
		t.Fatalf("render tmpl3 failed: %v", err)
	}
	if !strings.Contains(out3, "background-color: red") || !strings.Contains(out3, "font-size: 16px") {
		t.Errorf("expected dynamic styles in out3, got: %s", out3)
	}

	// 4. Component passing :class to child component props
	compTmpl := `<component name="Badge">
  <props>
    <prop name="class" default="" />
  </props>
  <span class="badge {{ class }}"><slot /></span>
</component>

<ui:Badge :class="DynamicBadge">New</ui:Badge>`

	if err := engine.ParseTemplate("comp_badge", compTmpl); err != nil {
		t.Fatalf("failed to parse comp_badge: %v", err)
	}

	out4, err := engine.RenderString("comp_badge", map[string]any{
		"DynamicBadge": map[string]bool{
			"badge-success": true,
			"badge-outline": true,
			"badge-ghost":   false,
		},
	})
	if err != nil {
		t.Fatalf("render comp_badge failed: %v", err)
	}

	if !strings.Contains(out4, "badge-success") || !strings.Contains(out4, "badge-outline") {
		t.Errorf("expected badge-success & badge-outline in out4, got: %s", out4)
	}
	if strings.Contains(out4, "badge-ghost") {
		t.Errorf("badge-ghost should not be present in out4, got: %s", out4)
	}
}
