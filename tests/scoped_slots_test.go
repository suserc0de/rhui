package tests

import (
	"strings"
	"testing"

	"github.com/suserc0de/rhui"
)

func TestScopedSlotsAndSelfClosingTags(t *testing.T) {
	engine, err := rhui.New()
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	// 1. Component with Scoped Slot: <slot item="row" index="idx" />
	tableTmpl := `<component name="DataTable">
  <props>
    <prop name="items" default="[]" />
  </props>
  <div class="table-container">
    <each items="items" as="row" index="idx">
      <div class="row">
        <slot item="row" index="idx" />
      </div>
    </each>
  </div>
</component>`

	if err := engine.ParseTemplate("components/DataTable", tableTmpl); err != nil {
		t.Fatalf("failed to parse DataTable component: %v", err)
	}

	// 2. Void/self-closing component: <ui:Icon name="check" />
	iconTmpl := `<component name="Icon">
  <props>
    <prop name="name" default="" />
  </props>
  <i class="icon icon-{{ name }}"></i>
</component>`

	if err := engine.ParseTemplate("components/Icon", iconTmpl); err != nil {
		t.Fatalf("failed to parse Icon component: %v", err)
	}

	// 3. Page using DataTable with Scoped Slot and Self-closing Icon
	pageTmpl := `<div class="page">
  <ui:DataTable :items="Users">
    <span class="user-item">#{{ index }}: {{ item.Name }} <ui:Icon name="user" /></span>
  </ui:DataTable>
</div>`

	if err := engine.ParseTemplate("pages/scoped_test", pageTmpl); err != nil {
		t.Fatalf("failed to parse page template: %v", err)
	}

	// 4. Render
	data := map[string]any{
		"Users": []map[string]any{
			{"Name": "Raden Haqi"},
			{"Name": "Alice Developer"},
		},
	}

	out, err := engine.RenderString("pages/scoped_test", data)
	if err != nil {
		t.Fatalf("RenderString failed: %v", err)
	}

	// 5. Assertions
	if !strings.Contains(out, `#0: Raden Haqi`) || !strings.Contains(out, `<i class="icon icon-user"></i>`) {
		t.Errorf("expected first user with scoped slot props and self-closing icon, got:\n%s", out)
	}
	if !strings.Contains(out, `#1: Alice Developer`) {
		t.Errorf("expected second user with scoped slot props and self-closing icon, got:\n%s", out)
	}
}
