package tests

import (
	"strings"
	"testing"

	"github.com/suserc0de/rhui"
)

func TestComponentRecursionDepthLimit(t *testing.T) {
	// Configure engine with a low max depth of 5
	engine, err := rhui.New(rhui.WithMaxRenderDepth(5))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	// Define a self-referencing component (infinite recursion)
	loopTmpl := `<component name="InfiniteLoop">
  <div>
    <span>Recursion</span>
    <ui:InfiniteLoop />
  </div>
</component>

<div>
  <ui:InfiniteLoop />
</div>`

	if err := engine.ParseTemplate("recursive_page", loopTmpl); err != nil {
		t.Fatalf("ParseTemplate failed: %v", err)
	}

	_, err = engine.RenderString("recursive_page", nil)
	if err == nil {
		t.Fatal("expected error due to max recursion depth limit, got nil")
	}

	if !strings.Contains(err.Error(), "maximum component recursion depth (5) exceeded") {
		t.Fatalf("unexpected error message: %v", err)
	}
}
