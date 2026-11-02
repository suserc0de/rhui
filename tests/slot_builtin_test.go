package tests

import (
	"bytes"
	"strings"
	"testing"

	"github.com/suserc0de/rhui/ast"
	"github.com/suserc0de/rhui/compiler"
	"github.com/suserc0de/rhui/parser"
	"github.com/suserc0de/rhui/runtime"
)

func TestHasSlotBuiltin(t *testing.T) {
	cardTmpl := `<component name="Modal">
  <div class="modal">
    <div class="modal-body"><slot /></div>
    <if test="has_slot('actions')">
      <div class="modal-actions"><slot:actions /></div>
    </if>
  </div>
</component>`

	p1 := parser.NewParser(parser.NewLexer("Modal.rhui", cardTmpl))
	cardRoot, err := p1.Parse()
	if err != nil {
		t.Fatalf("failed to parse modal component: %v", err)
	}

	registry := compiler.NewRegistry()
	for _, child := range cardRoot.Children {
		if compDef, ok := child.(*ast.ComponentDefNode); ok {
			registry.Register(compDef.Name, compDef)
		}
	}

	// Case 1: with actions slot
	pageTmpl1 := `<ui:Modal>
  <p>Are you sure?</p>
  <slot:actions><button>Confirm</button></slot:actions>
</ui:Modal>`

	p2 := parser.NewParser(parser.NewLexer("page1.rhui", pageTmpl1))
	root1, _ := p2.Parse()

	renderer := runtime.NewRenderer(registry, nil)
	var buf1 bytes.Buffer
	if err := renderer.RenderNode(&buf1, root1, runtime.NewScope(nil)); err != nil {
		t.Fatalf("render failed: %v", err)
	}

	if !strings.Contains(buf1.String(), "<div class=\"modal-actions\"><button>Confirm</button></div>") {
		t.Fatalf("expected modal-actions rendered when slot provided, got: %s", buf1.String())
	}

	// Case 2: without actions slot
	pageTmpl2 := `<ui:Modal><p>No actions</p></ui:Modal>`
	p3 := parser.NewParser(parser.NewLexer("page2.rhui", pageTmpl2))
	root2, _ := p3.Parse()

	var buf2 bytes.Buffer
	if err := renderer.RenderNode(&buf2, root2, runtime.NewScope(nil)); err != nil {
		t.Fatalf("render failed: %v", err)
	}

	if strings.Contains(buf2.String(), "modal-actions") {
		t.Fatalf("expected modal-actions NOT rendered when slot omitted, got: %s", buf2.String())
	}
}
