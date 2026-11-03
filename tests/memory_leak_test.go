package tests

import (
	"runtime"
	"testing"

	"github.com/suserc0de/rhui"
)

// TestMemoryLeakScopeLifecycle ensures 100,000 iterations of nested components and loops
// do not cause memory leaks or linear heap growth.
func TestMemoryLeakScopeLifecycle(t *testing.T) {
	engine, err := rhui.New()
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	cardTmpl := `<component name="TestCard">
  <props>
    <prop name="title" default="Default" />
  </props>
  <div class="card">
    <h4>{{ title }}</h4>
    <slot />
  </div>
</component>`

	pageTmpl := `<div class="container">
  <each items="Items" as="item">
    <ui:TestCard title="Card Title">
      <p>Item {{ item.ID }}: {{ item.Name }}</p>
    </ui:TestCard>
  </each>
</div>`

	if err := engine.ParseTemplate("components/TestCard", cardTmpl); err != nil {
		t.Fatalf("ParseTemplate Card failed: %v", err)
	}
	if err := engine.ParseTemplate("pages/leak_test", pageTmpl); err != nil {
		t.Fatalf("ParseTemplate Page failed: %v", err)
	}

	data := map[string]any{
		"Items": []map[string]any{
			{"ID": 1, "Name": "A"},
			{"ID": 2, "Name": "B"},
			{"ID": 3, "Name": "C"},
			{"ID": 4, "Name": "D"},
		},
	}

	var buf discardWriter

	// Warmup
	for i := 0; i < 1000; i++ {
		_ = engine.Render(buf, "pages/leak_test", data)
	}
	runtime.GC()

	var memStart runtime.MemStats
	runtime.ReadMemStats(&memStart)

	// 100,000 iterations of rendering nested components and loops
	for i := 0; i < 100000; i++ {
		if err := engine.Render(buf, "pages/leak_test", data); err != nil {
			t.Fatalf("Render failed: %v", err)
		}
	}

	runtime.GC()
	var memEnd runtime.MemStats
	runtime.ReadMemStats(&memEnd)

	// Heap growth should be minimal and bounded by the pool max capacity (well under 5MB)
	growth := int64(memEnd.HeapAlloc) - int64(memStart.HeapAlloc)
	if growth > 5*1024*1024 {
		t.Fatalf("Potential memory leak detected! Heap grew by %d bytes over 100,000 renders", growth)
	}
}
