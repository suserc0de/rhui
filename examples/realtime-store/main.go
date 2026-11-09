package main

import (
	"embed"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"

	"github.com/suserc0de/rhui"
	"github.com/suserc0de/rhui/sse"
)

//go:embed templates/*
var templateFS embed.FS

type Product struct {
	ID       int
	Title    string
	Price    int
	Stock    int
	Category string
}

type StoreState struct {
	mu        sync.RWMutex
	cartCount int
	products  []Product
}

var store = &StoreState{
	cartCount: 0,
	products: []Product{
		{ID: 1, Title: "Compute Engine Pro (4 vCPU, 16GB RAM)", Price: 450000, Stock: 8, Category: "Compute"},
		{ID: 2, Title: "Managed Postgres Database Cluster", Price: 720000, Stock: 4, Category: "Database"},
		{ID: 3, Title: "Global CDN Edge Accelerator", Price: 180000, Stock: 15, Category: "Networking"},
		{ID: 4, Title: "Serverless Container Runner", Price: 290000, Stock: 2, Category: "Compute"},
		{ID: 5, Title: "Persistent Block Storage (1 TB SSD)", Price: 350000, Stock: 11, Category: "Storage"},
		{ID: 6, Title: "DDoS Mitigation & WAF Shield", Price: 600000, Stock: 6, Category: "Security"},
	},
}

func main() {
	// Initialize RHUI Engine with idiomatic rhui.New constructor
	engine, err := rhui.New(
		rhui.WithFS(templateFS),
		rhui.WithDir("templates"),
		rhui.WithExtension(".rhui"),
	)
	if err != nil {
		log.Fatalf("failed to initialize RHUI engine: %v", err)
	}

	broker := sse.NewBroker()

	// 1. Homepage / Store View (Render with Layout)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		store.mu.RLock()
		data := map[string]any{
			"Title":     "Cloud Marketplace",
			"CartCount": store.cartCount,
			"Products":  store.products,
		}
		store.mu.RUnlock()

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := engine.RenderWithLayout(w, "layouts/store", "pages/store", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	// 2. Add to Cart (HTMX Partial Swap using RenderFragment)
	http.HandleFunc("/cart/add", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		idStr := r.URL.Query().Get("id")
		id, _ := strconv.Atoi(idStr)

		store.mu.Lock()
		store.cartCount++
		var updatedProduct *Product
		for i := range store.products {
			if store.products[i].ID == id && store.products[i].Stock > 0 {
				store.products[i].Stock--
				updatedProduct = &store.products[i]
				break
			}
		}
		currentCount := store.cartCount
		store.mu.Unlock()

		// Broadcast stock change to all active SSE browser connections
		if updatedProduct != nil {
			broadcastStockUpdate(engine, broker, updatedProduct)
		}

		// Return ONLY the cart-badge fragment to HTMX
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		err := engine.RenderFragment(w, "layouts/store", "cart-badge", map[string]any{
			"CartCount": currentCount,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	// 3. Simulate Remote Order (Triggers SSE Stock Broadcast)
	http.HandleFunc("/simulate/purchase", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		store.mu.Lock()
		idx := rand.Intn(len(store.products))
		if store.products[idx].Stock > 0 {
			store.products[idx].Stock--
		} else {
			store.products[idx].Stock = 10 // restock simulation
		}
		p := store.products[idx]
		store.mu.Unlock()

		broadcastStockUpdate(engine, broker, &p)
		w.WriteHeader(http.StatusNoContent)
	})

	// 4. SSE Endpoint for live stock updates
	http.HandleFunc("/sse/live-stock", func(w http.ResponseWriter, r *http.Request) {
		stream, err := broker.Connect(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer stream.Close()

		notify := r.Context().Done()
		go func() {
			<-notify
			stream.Close()
		}()

		stream.Wait()
	})

	port := 8081
	fmt.Printf("🚀 RHUI Realtime Store Showcase running at: http://localhost:%d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

func broadcastStockUpdate(engine *rhui.Engine, broker *sse.Broker, p *Product) {
	// Parse inline component definition or render template
	componentTmpl := fmt.Sprintf(`<ui:StockBadge stock="%d" />`, p.Stock)
	_ = engine.ParseTemplate("inline/badge", componentTmpl)

	_ = broker.BroadcastComponent(engine, "inline/badge", map[string]any{"stock": p.Stock}, sse.EventOptions{
		EventName: fmt.Sprintf("stock-update-%d", p.ID),
	})
}
