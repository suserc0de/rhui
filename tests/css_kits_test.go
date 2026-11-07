package tests

import (
	"strings"
	"testing"

	"github.com/suserc0de/rhui"
)

func TestTailwindAndBootstrapComponentKits(t *testing.T) {
	engine, err := rhui.New()
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	// 1. Tailwind UI Button Component
	tailwindBtn := `<component name="TailwindButton">
  <props>
    <prop name="variant" default="primary" />
    <prop name="size" default="md" />
    <prop name="customClass" default="" />
  </props>

  <button 
    class="inline-flex items-center justify-center font-medium transition-all duration-150 rounded-lg shadow-sm focus:outline-none focus:ring-2 focus:ring-offset-2 {{ customClass }}"
    :class="{'bg-indigo-600 hover:bg-indigo-700 text-white focus:ring-indigo-500': variant == 'primary', 'bg-slate-100 hover:bg-slate-200 text-slate-800 focus:ring-slate-400': variant == 'secondary', 'px-3 py-1.5 text-xs': size == 'sm', 'px-4 py-2 text-sm': size == 'md', 'px-6 py-3 text-base': size == 'lg'}">
    <slot />
  </button>
</component>`

	if err := engine.ParseTemplate("components/TailwindButton", tailwindBtn); err != nil {
		t.Fatalf("ParseTemplate TailwindButton failed: %v", err)
	}

	// 2. Bootstrap 5 Card Component
	bootstrapCard := `<component name="BootstrapCard">
  <props>
    <prop name="title" default="" />
    <prop name="headerClass" default="bg-light" />
  </props>

  <div class="card shadow-sm mb-4">
    <if test="title != ''">
      <div class="card-header d-flex justify-content-between align-items-center {{ headerClass }}">
        <h5 class="card-title mb-0">{{ title }}</h5>
        <slot:badge />
      </div>
    </if>
    <div class="card-body">
      <slot />
    </div>
    <if test="has_slot('footer')">
      <div class="card-footer text-muted small">
        <slot:footer />
      </div>
    </if>
  </div>
</component>`

	if err := engine.ParseTemplate("components/BootstrapCard", bootstrapCard); err != nil {
		t.Fatalf("ParseTemplate BootstrapCard failed: %v", err)
	}

	// 3. Page using both components with interpolated and dynamic attributes
	pageTmpl := `<div class="container mx-auto py-8">
  <ui:TailwindButton variant="primary" size="lg" customClass="w-full sm:w-auto">
    <span>Get Started</span>
  </ui:TailwindButton>

  <ui:BootstrapCard title="System Performance" headerClass="bg-dark text-white">
    <slot:badge>
      <span class="badge bg-success">99.9% Up</span>
    </slot:badge>
    <p class="card-text">All systems operational.</p>
    <slot:footer>
      Last checked 1 minute ago.
    </slot:footer>
  </ui:BootstrapCard>
</div>`

	if err := engine.ParseTemplate("pages/demo", pageTmpl); err != nil {
		t.Fatalf("ParseTemplate demo failed: %v", err)
	}

	out, err := engine.RenderString("pages/demo", nil)
	if err != nil {
		t.Fatalf("RenderString demo failed: %v", err)
	}

	// Verify Tailwind classes
	expectedTailwind := []string{
		"inline-flex items-center justify-center font-medium",
		"w-full sm:w-auto",
		"bg-indigo-600 hover:bg-indigo-700 text-white",
		"px-6 py-3 text-base",
		"<span>Get Started</span>",
	}
	for _, expected := range expectedTailwind {
		if !strings.Contains(out, expected) {
			t.Errorf("expected Tailwind class/element %q in output, got:\n%s", expected, out)
		}
	}

	// Verify Bootstrap classes
	expectedBootstrap := []string{
		"card shadow-sm mb-4",
		"card-header d-flex justify-content-between align-items-center bg-dark text-white",
		"card-title mb-0",
		"badge bg-success",
		"card-body",
		"card-footer text-muted small",
		"Last checked 1 minute ago.",
	}
	for _, expected := range expectedBootstrap {
		if !strings.Contains(out, expected) {
			t.Errorf("expected Bootstrap class/element %q in output, got:\n%s", expected, out)
		}
	}
}
