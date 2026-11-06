package tests

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/suserc0de/rhui"
	"github.com/suserc0de/rhui/sse"
)

func TestSSESecurityHeadersAndSanitization(t *testing.T) {
	// 1. Configurable CORS Origin
	broker := sse.NewBroker(sse.WithAllowedOrigin("https://trusted.domain.com"))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/events", nil)

	stream, err := broker.Connect(rec, req)
	if err != nil {
		t.Fatalf("broker.Connect failed: %v", err)
	}
	defer stream.Close()

	if origin := rec.Header().Get("Access-Control-Allow-Origin"); origin != "https://trusted.domain.com" {
		t.Fatalf("expected origin 'https://trusted.domain.com', got %q", origin)
	}

	// 2. SSE Injection Prevention in Event Name and ID
	engine, err := rhui.New()
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	if err := engine.ParseTemplate("badge", `<span>Active</span>`); err != nil {
		t.Fatalf("ParseTemplate failed: %v", err)
	}

	// Inject CR/LF into EventName and ID
	maliciousOpts := sse.EventOptions{
		EventName: "alert\r\ndata: injected-fake-event\r\n\r\n",
		ID:        "id123\ncustom-header: evil",
	}

	if err := broker.BroadcastComponent(engine, "badge", nil, maliciousOpts); err != nil {
		t.Fatalf("BroadcastComponent failed: %v", err)
	}

	// Give broadcast channel time to flush
	time.Sleep(10 * time.Millisecond)
}
