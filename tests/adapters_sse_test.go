package tests

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coderhaqi/rhui"
	"github.com/coderhaqi/rhui/adapter/chi"
	"github.com/coderhaqi/rhui/adapter/echo"
	"github.com/coderhaqi/rhui/adapter/fiber"
	"github.com/coderhaqi/rhui/adapter/gin"
	"github.com/coderhaqi/rhui/adapter/nethttp"
	"github.com/coderhaqi/rhui/sse"
)

func TestNetHTTPAdapter(t *testing.T) {
	engine, _ := rhui.New()
	_ = engine.ParseTemplate("home", "<h1>Welcome</h1>")

	rec := httptest.NewRecorder()
	err := nethttp.Render(rec, http.StatusOK, engine, "home", nil)
	if err != nil {
		t.Fatalf("nethttp adapter render failed: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "<h1>Welcome</h1>") {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestChiAdapter(t *testing.T) {
	engine, _ := rhui.New()
	_ = engine.ParseTemplate("chi-page", "<h2>Chi Route</h2>")

	rec := httptest.NewRecorder()
	err := chi.Render(rec, http.StatusOK, engine, "chi-page", nil)
	if err != nil {
		t.Fatalf("chi adapter render failed: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "<h2>Chi Route</h2>") {
		t.Fatalf("unexpected chi output: %s", rec.Body.String())
	}
}

func TestGinAdapter(t *testing.T) {
	engine, _ := rhui.New()
	_ = engine.ParseTemplate("dashboard", "<p>Gin Page</p>")

	ginAdapter := gin.New(engine)
	renderInstance := ginAdapter.Instance("dashboard", nil)

	rec := httptest.NewRecorder()
	err := renderInstance.Render(rec)
	if err != nil {
		t.Fatalf("gin adapter render failed: %v", err)
	}

	if !strings.Contains(rec.Body.String(), "<p>Gin Page</p>") {
		t.Fatalf("unexpected gin render output: %s", rec.Body.String())
	}
}

func TestEchoAdapter(t *testing.T) {
	engine, _ := rhui.New()
	_ = engine.ParseTemplate("echo-page", "<span>Echo Content</span>")

	echoAdapter := echo.New(engine)
	rec := httptest.NewRecorder()
	err := echoAdapter.Render(rec, "echo-page", nil, nil)
	if err != nil {
		t.Fatalf("echo adapter render failed: %v", err)
	}

	if !strings.Contains(rec.Body.String(), "<span>Echo Content</span>") {
		t.Fatalf("unexpected echo render output: %s", rec.Body.String())
	}
}

func TestFiberAdapter(t *testing.T) {
	engine, _ := rhui.New()
	_ = engine.ParseTemplate("fiber-page", "<div>Fiber View</div>")

	fiberAdapter := fiber.New(engine)
	rec := httptest.NewRecorder()
	err := fiberAdapter.Render(rec, "fiber-page", nil)
	if err != nil {
		t.Fatalf("fiber adapter render failed: %v", err)
	}

	if !strings.Contains(rec.Body.String(), "<div>Fiber View</div>") {
		t.Fatalf("unexpected fiber render output: %s", rec.Body.String())
	}
}

func TestSSELiveComponentBroadcast(t *testing.T) {
	engine, _ := rhui.New()
	_ = engine.ParseTemplate("components/Alert", "<div class=\"alert\">{{ Message }}</div>")

	broker := sse.NewBroker()

	req := httptest.NewRequest("GET", "/events", nil)
	rec := httptest.NewRecorder()

	stream, err := broker.Connect(rec, req)
	if err != nil {
		t.Fatalf("broker.Connect failed: %v", err)
	}
	defer stream.Close()

	// Wait briefly for client registration
	time.Sleep(20 * time.Millisecond)

	// Broadcast component
	err = broker.BroadcastComponent(engine, "components/Alert", map[string]any{
		"Message": "Live Update Successful!",
	}, sse.EventOptions{
		EventName: "NewAlert",
		ID:        "msg-1",
	})
	if err != nil {
		t.Fatalf("broadcast component failed: %v", err)
	}

	// Flush event in background
	go func() {
		time.Sleep(50 * time.Millisecond)
		stream.Close()
	}()

	stream.Wait()

	body := rec.Body.String()
	if !strings.Contains(body, "event: NewAlert") {
		t.Errorf("expected SSE event name 'NewAlert', got: %s", body)
	}
	if !strings.Contains(body, "id: msg-1") {
		t.Errorf("expected SSE ID 'msg-1', got: %s", body)
	}
	if !strings.Contains(body, "data: <div class=\"alert\">Live Update Successful!</div>") {
		t.Errorf("expected SSE rendered component data, got: %s", body)
	}
}
