package main

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/suserc0de/rhui"
	"github.com/suserc0de/rhui/sse"
)

//go:embed templates/*
var templateFS embed.FS

type User struct {
	ID    int
	Name  string
	Email string
	Role  string
}

func main() {
	// Initialize RHUI Engine with embedded templates
	engine, err := rhui.New(
		rhui.WithFS(templateFS),
		rhui.WithDir("templates"),
		rhui.WithExtension(".rhui"),
	)
	if err != nil {
		log.Fatalf("failed to initialize RHUI engine: %v", err)
	}

	broker := sse.NewBroker()

	// Initial data
	users := []User{
		{ID: 1, Name: "Raden Haqi", Email: "haqi@example.com", Role: "Administrator"},
		{ID: 2, Name: "Alice Smith", Email: "alice@example.com", Role: "Developer"},
		{ID: 3, Name: "Bob Wilson", Email: "bob@example.com", Role: "Manager"},
	}

	// 1. Dashboard Page Handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := map[string]any{
			"Title":   "Overview",
			"Heading": "System Dashboard",
			"Users":   users,
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := engine.RenderWithLayout(w, "layouts/main", "pages/dashboard", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	// 2. Realtime SSE Stream Endpoint
	http.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		stream, err := broker.Connect(w, r)
		if err != nil {
			http.Error(w, "SSE unsupported", http.StatusBadRequest)
			return
		}
		defer stream.Close()
		stream.Wait()
	})

	// 3. Trigger alert event (HTMX button)
	http.HandleFunc("/api/trigger-alert", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		alertMsg := fmt.Sprintf("Live Server Notification: Task completed successfully at %s!", time.Now().Format("15:04:05"))

		// Broadcast rendered Alert component directly via SSE
		_ = broker.BroadcastComponent(engine, "components/Alert", map[string]any{
			"type":    "info",
			"message": alertMsg,
		}, sse.EventOptions{
			EventName: "NewNotification",
		})

		w.WriteHeader(http.StatusOK)
	})

	log.Println("🚀 RHUI Admin Dashboard running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
