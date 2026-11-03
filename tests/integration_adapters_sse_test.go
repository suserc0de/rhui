package tests

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gofiber/fiber/v2"
	"github.com/labstack/echo/v4"
	"github.com/suserc0de/rhui"
	rhuichi "github.com/suserc0de/rhui/adapter/chi"
	rhuiecho "github.com/suserc0de/rhui/adapter/echo"
	rhuifiber "github.com/suserc0de/rhui/adapter/fiber"
	rhuigin "github.com/suserc0de/rhui/adapter/gin"
	"github.com/suserc0de/rhui/sse"
)

// TestAdaptersFullStackLifecycle verifies each router can serve real HTTP traffic through RHUI
func TestAdaptersFullStackLifecycle(t *testing.T) {
	engine, err := rhui.New()
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	tmpl := `<div class="user-card"><h3>{{ Title }}</h3><p>User: {{ Name }}</p></div>`
	if err := engine.ParseTemplate("user", tmpl); err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	// 1. Gin Framework Test
	t.Run("Gin Router Integration", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		r := gin.New()
		r.HTMLRender = rhuigin.New(engine)
		r.GET("/user", func(c *gin.Context) {
			c.HTML(http.StatusOK, "user", gin.H{"Title": "Profile", "Name": "Alice"})
		})

		req := httptest.NewRequest("GET", "/user", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "User: Alice") {
			t.Fatalf("Gin fullstack failed: status %d, body: %s", rec.Code, rec.Body.String())
		}
	})

	// 2. Echo Framework Test
	t.Run("Echo Router Integration", func(t *testing.T) {
		e := echo.New()
		e.Renderer = rhuiecho.New(engine)
		e.GET("/user", func(c echo.Context) error {
			return c.Render(http.StatusOK, "user", map[string]any{"Title": "Profile", "Name": "Bob"})
		})

		req := httptest.NewRequest("GET", "/user", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "User: Bob") {
			t.Fatalf("Echo fullstack failed: status %d, body: %s", rec.Code, rec.Body.String())
		}
	})

	// 3. Fiber Framework Test
	t.Run("Fiber Router Integration", func(t *testing.T) {
		app := fiber.New(fiber.Config{
			Views: rhuifiber.New(engine),
		})
		app.Get("/user", func(c *fiber.Ctx) error {
			return c.Render("user", fiber.Map{"Title": "Profile", "Name": "Charlie"})
		})

		req := httptest.NewRequest("GET", "/user", nil)
		resp, err := app.Test(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatalf("Fiber fullstack request failed: %v", err)
		}
	})

	// 4. Chi / net/http Router Test
	t.Run("Chi & NetHTTP Router Integration", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = rhuichi.Render(w, http.StatusOK, engine, "user", map[string]any{
				"Title": "Profile",
				"Name":  "David",
			})
		})

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/user", nil)
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "User: David") {
			t.Fatalf("Chi fullstack failed: status %d, body: %s", rec.Code, rec.Body.String())
		}
	})
}

// TestSSEConcurrentMultiClientBroadcast tests SSE with 20 concurrent connected client streams
func TestSSEConcurrentMultiClientBroadcast(t *testing.T) {
	engine, _ := rhui.New()
	_ = engine.ParseTemplate("badge", `<span class="badge">{{ Count }}</span>`)

	broker := sse.NewBroker()
	clientCount := 20
	var wg sync.WaitGroup
	streams := make([]*sse.Stream, clientCount)
	recorders := make([]*httptest.ResponseRecorder, clientCount)

	for i := 0; i < clientCount; i++ {
		rec := httptest.NewRecorder()
		recorders[i] = rec
		req := httptest.NewRequest("GET", "/events", nil)

		stream, err := broker.Connect(rec, req)
		if err != nil {
			t.Fatalf("client %d connect failed: %v", i, err)
		}
		streams[i] = stream

		wg.Add(1)
		go func(s *sse.Stream) {
			defer wg.Done()
			s.Wait()
		}(stream)
	}

	time.Sleep(30 * time.Millisecond)

	// Broadcast live component update to all 20 clients simultaneously
	err := broker.BroadcastComponent(engine, "badge", map[string]any{
		"Count": 99,
	}, sse.EventOptions{
		EventName: "BadgeUpdate",
	})
	if err != nil {
		t.Fatalf("broadcast failed: %v", err)
	}

	// Close all streams cleanly
	for _, s := range streams {
		s.Close()
	}
	wg.Wait()

	// Verify every single client received the exact rendered HTML
	for i, rec := range recorders {
		body := rec.Body.String()
		expectedData := `data: <span class="badge">99</span>`
		if !strings.Contains(body, expectedData) {
			t.Fatalf("client %d did not receive broadcast, body: %s", i, body)
		}
	}
}
