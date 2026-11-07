package tests

import (
	"strings"
	"testing"

	"github.com/suserc0de/rhui"
)

func TestAttributeInterpolation(t *testing.T) {
	engine, err := rhui.New()
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	// 1. Component with static attribute interpolation in element tags
	alertTmpl := `<component name="Alert">
  <props>
    <prop name="type" default="info" />
    <prop name="customClass" default="" />
  </props>
  <div class="alert alert-{{ type }} {{ customClass }}" data-type="{{ type }}" id="alert-box-{{ type }}">
    <slot />
  </div>
</component>

<ui:Alert type="danger" customClass="shadow-lg font-bold">
  <span>System Error!</span>
</ui:Alert>`

	if err := engine.ParseTemplate("alert_test", alertTmpl); err != nil {
		t.Fatalf("ParseTemplate failed: %v", err)
	}

	out, err := engine.RenderString("alert_test", nil)
	if err != nil {
		t.Fatalf("RenderString failed: %v", err)
	}

	expectedParts := []string{
		`class="alert alert-danger shadow-lg font-bold"`,
		`data-type="danger"`,
		`id="alert-box-danger"`,
		`<span>System Error!</span>`,
	}
	for _, part := range expectedParts {
		if !strings.Contains(out, part) {
			t.Errorf("expected output to contain %q, but got:\n%s", part, out)
		}
	}

	// 2. Element level attribute interpolation with Scope variables
	pageTmpl := `<div class="btn btn-{{ Variant }} px-{{ Padding }}" title="User: {{ User.Name }}">
  Click Me
</div>`

	if err := engine.ParseTemplate("page_test", pageTmpl); err != nil {
		t.Fatalf("ParseTemplate failed: %v", err)
	}

	outPage, err := engine.RenderString("page_test", map[string]any{
		"Variant": "success",
		"Padding": 4,
		"User": map[string]string{
			"Name": "Haqi",
		},
	})
	if err != nil {
		t.Fatalf("RenderString failed: %v", err)
	}

	expectedPageParts := []string{
		`class="btn btn-success px-4"`,
		`title="User: Haqi"`,
	}
	for _, part := range expectedPageParts {
		if !strings.Contains(outPage, part) {
			t.Errorf("expected page output to contain %q, but got:\n%s", part, outPage)
		}
	}

	// 3. Attribute interpolation quote breakout safety test
	attackTmpl := `<div class="prefix-{{ Injected }}-suffix">Safe</div>`
	if err := engine.ParseTemplate("attack_test", attackTmpl); err != nil {
		t.Fatalf("ParseTemplate failed: %v", err)
	}

	outAttack, err := engine.RenderString("attack_test", map[string]any{
		"Injected": `foo" onfocus="alert(1)`,
	})
	if err != nil {
		t.Fatalf("RenderString failed: %v", err)
	}

	// Must be escaped, preventing breakout quotes
	if strings.Contains(outAttack, `onfocus="alert(1)`) {
		t.Fatalf("Attribute quote breakout succeeded! Output:\n%s", outAttack)
	}
	if !strings.Contains(outAttack, `&#34;`) && !strings.Contains(outAttack, `&quot;`) {
		t.Fatalf("Expected escaped quote, got:\n%s", outAttack)
	}
}
