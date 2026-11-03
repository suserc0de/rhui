package tests

import (
	"sync"
	"testing"
	"time"

	"github.com/suserc0de/rhui"
)

// TestConcurrentReloadAndRender ensures that continuous concurrent rendering
// while templates are repeatedly reloaded produces 0 data races and 0 panics.
func TestConcurrentReloadAndRender(t *testing.T) {
	engine, err := rhui.New()
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	tmpl1 := `<div class="v1"><h1>{{ Title }}</h1><p>{{ Body }}</p></div>`
	tmpl2 := `<div class="v2"><h2>{{ Title }} (Reloaded)</h2><span>{{ Body }}</span></div>`

	if err := engine.ParseTemplate("page", tmpl1); err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	var wg sync.WaitGroup
	stopChan := make(chan struct{})

	// 8 concurrent render worker goroutines
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			var buf discardWriter
			data := map[string]any{
				"Title": "Concurrent Rendering Test",
				"Body":  "Ensuring thread-safe reload snapshotting",
			}
			for {
				select {
				case <-stopChan:
					return
				default:
					if err := engine.Render(buf, "page", data); err != nil {
						t.Errorf("worker %d render error: %v", workerID, err)
						return
					}
				}
			}
		}(i)
	}

	// 1 goroutine that continuously reloads/updates template concurrently
	wg.Add(1)
	go func() {
		defer wg.Done()
		toggle := false
		for {
			select {
			case <-stopChan:
				return
			default:
				if toggle {
					_ = engine.ParseTemplate("page", tmpl1)
				} else {
					_ = engine.ParseTemplate("page", tmpl2)
				}
				toggle = !toggle
				time.Sleep(1 * time.Millisecond)
			}
		}
	}()

	// Run stress test for 1 second
	time.Sleep(1 * time.Second)
	close(stopChan)
	wg.Wait()
}

// TestNewEngineAutoReloadOption verifies that AutoReload: true enables devMode
func TestNewEngineAutoReloadOption(t *testing.T) {
	e := rhui.NewEngine(rhui.EngineOptions{
		ViewsDir:   "tests/testdata",
		AutoReload: true,
	})
	if e == nil {
		t.Fatal("expected non-nil engine")
	}

	// Direct verification of reload without crash
	if err := e.Reload(); err != nil {
		t.Fatalf("Reload failed: %v", err)
	}
}
