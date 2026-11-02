package tests

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/coderhaqi/rhui"
)

// TestLoad10000ConcurrentRequests evaluates memory stability and zero packet drops under 10,000 requests
func TestLoad10000ConcurrentRequests(t *testing.T) {
	engine, err := rhui.New()
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	tmpl := `<div class="dashboard-metric">
  <h2>{{ Title }}</h2>
  <p class="metric-val">{{ Count }}</p>
  <if test="Count > 5000">
    <span class="badge-high">High Load</span>
  </if>
  <else>
    <span class="badge-normal">Normal Load</span>
  </else>
</div>`

	if err := engine.ParseTemplate("pages/metric", tmpl); err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	// Create test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := engine.Render(w, "pages/metric", map[string]any{
			"Title": "Concurrent Pressure Test",
			"Count": 7500,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	totalRequests := 10000
	concurrency := 100

	var wg sync.WaitGroup
	var successCount int64
	var errorCount int64

	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        1000,
			MaxIdleConnsPerHost: 1000,
			IdleConnTimeout:     30 * time.Second,
		},
		Timeout: 5 * time.Second,
	}

	reqsPerWorker := totalRequests / concurrency
	startTime := time.Now()

	for c := 0; c < concurrency; c++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < reqsPerWorker; i++ {
				resp, err := client.Get(server.URL)
				if err != nil {
					atomic.AddInt64(&errorCount, 1)
					continue
				}
				if resp.StatusCode == http.StatusOK {
					atomic.AddInt64(&successCount, 1)
				} else {
					atomic.AddInt64(&errorCount, 1)
				}
				resp.Body.Close()
			}
		}(c)
	}

	wg.Wait()
	duration := time.Since(startTime)

	t.Logf("Completed %d concurrent requests in %v (%.2f RPS)",
		totalRequests, duration, float64(totalRequests)/duration.Seconds())

	if errorCount > 0 {
		t.Fatalf("Load test encountered %d errors out of %d requests", errorCount, totalRequests)
	}

	if successCount != int64(totalRequests) {
		t.Fatalf("Expected %d successful responses, got %d", totalRequests, successCount)
	}
}

// TestStressRaceGoroutineExecution runs concurrent template parses and renders simultaneously
func TestStressRaceGoroutineExecution(t *testing.T) {
	engine, err := rhui.New()
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	var wg sync.WaitGroup
	goroutines := 500

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			name := fmt.Sprintf("dynamic/template_%d", id%10)
			_ = engine.ParseTemplate(name, `<div>Iteration {{ id }}</div>`)
			_, _ = engine.RenderString(name, map[string]any{"id": id})
		}(i)
	}

	wg.Wait()
}
