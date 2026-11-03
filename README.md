# RHUI (RH Template Engine)

<p align="center">
  <a href="https://github.com/suserc0de/rhui/actions"><img src="https://img.shields.io/github/actions/workflow/status/suserc0de/rhui/ci.yml?branch=master&style=flat-square&logo=github&label=CI" alt="CI Status" /></a>
  <a href="https://pkg.go.dev/github.com/suserc0de/rhui"><img src="https://img.shields.io/badge/Go-Reference-007d9c?style=flat-square&logo=go&logoColor=white" alt="Go Reference" /></a>
  <a href="https://goreportcard.com/report/github.com/suserc0de/rhui"><img src="https://img.shields.io/badge/Go%20Report-A%2B-brightgreen?style=flat-square" alt="Go Report Card" /></a>
  <a href="https://github.com/suserc0de/rhui/releases"><img src="https://img.shields.io/badge/Release-v1.0.2-blue?style=flat-square" alt="Latest Release" /></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-green.svg?style=flat-square" alt="License" /></a>
</p>

<p align="center">
  <b>A component-based, high-performance Server-Side Rendering (SSR) template engine for Go.</b><br>
  Zero JavaScript build-step &bull; Safe by default &bull; Reusable components with slots &bull; Event-Driven SSE
</p>

---

## 🚀 Overview

**RHUI** is a modern, declarative component template engine for Go. It brings the developer experience of Vue, Svelte, or Astro directly to Go backends with **pure Server-Side Rendering (SSR)**.

With RHUI, you can organize your user interfaces into reusable components (`.rhui` files) with typed props, default values, and named slots without needing Node.js, npm, Webpack, or hydration overhead.

### Key Features

- **🧩 Declarative Reusable Components:** Write components like `<ui:Card>`, `<ui:Button>`, and layouts with clean HTML syntax.
- **⚡ Zero Build Step:** Templates are parsed and compiled in-memory during application startup. No `go generate` or CLI transpile step required.
- **🏷️ Signature `.rhui` Extension:** Completely distinct from Go's standard `html/template` (`.tmpl`), preventing editor lint conflicts while remaining 100% compatible with HTML tooling.
- **🛡️ Safe by Default:** Contextual HTML escaping and a sandboxed expression evaluator with strict allowlists (no arbitrary code execution).
- **📡 Event-Driven & Realtime SSE:** Native Server-Sent Events package (`rhui/sse`) to stream live rendered component updates directly to browsers (seamless with HTMX SSE and Datastar).
- **🔌 Multi-Framework Support:** Official adapters for **Gin**, **Echo**, **Fiber**, and standard `net/http` / **Chi**.
- **🌊 Low Heap Allocations:** Pre-compiled expression AST caching, pooled evaluation contexts, zero-allocation fast-path escaper, and sub-scope recycling with `sync.Pool`.
- **⚡ Supercharged Engine Speed:** Sub-microsecond single component render (~3,000 ns) and outperforms `html/template` on complex loop tables.

---

## 📦 Installation

```bash
go get github.com/suserc0de/rhui
```

---

## 💡 Quick Start

### 1. Configure VS Code Syntax Highlighting

To get automatic HTML syntax highlighting for `.rhui` files, add this to your VS Code `settings.json`:

```json
{
  "files.associations": {
    "*.rhui": "html"
  }
}
```

### 2. Define a Reusable Component (`templates/components/Card.rhui`)

```html
<component name="Card">
  <props>
    <prop name="title" default="Default Title" />
    <prop name="variant" default="default" />
  </props>

  <div class="card card-{{ variant }}">
    <div class="card-header">
      <h3>{{ title }}</h3>
    </div>
    <div class="card-body">
      <!-- Default slot -->
      <slot />
    </div>
    <div class="card-footer">
      <!-- Named slot -->
      <slot:footer />
    </div>
  </div>
</component>
```

### 3. Compose Components in a Page (`templates/pages/dashboard.rhui`)

```html
<div class="container">
  <h1>{{ Heading }}</h1>

  <ui:Card title="User Overview" variant="elevated">
    <if test="len(Users) > 0">
      <each items="Users" as="user">
        <p>{{ user.Name }} - Status: {{ user.Role }}</p>
      </each>
    </if>
    <else>
      <p>No active users found.</p>
    </else>

    <slot:footer>
      <small>Total: {{ len(Users) }} registered users</small>
    </slot:footer>
  </ui:Card>
</div>
```

### 4. Render in Go (`main.go`)

```go
package main

import (
	"log"
	"net/http"

	"github.com/suserc0de/rhui"
)

func main() {
	engine, err := rhui.New(
		rhui.WithDir("templates"),
		rhui.WithExtension(".rhui"),
	)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		data := map[string]any{
			"Heading": "Analytics Dashboard",
			"Users": []map[string]any{
				{"Name": "Haqi", "Role": "Admin"},
				{"Name": "Alice", "Role": "Developer"},
			},
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := engine.Render(w, "pages/dashboard", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	log.Println("Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

---

## 🌐 Framework Integrations

### Gin Framework
```go
import (
    "github.com/gin-gonic/gin"
    "github.com/suserc0de/rhui"
    rhuigin "github.com/suserc0de/rhui/adapter/gin"
)

func main() {
    engine, _ := rhui.New(rhui.WithDir("templates"))
    r := gin.Default()
    r.HTMLRender = rhuigin.New(engine)

    r.GET("/users", func(c *gin.Context) {
        c.HTML(http.StatusOK, "pages/users", gin.H{"Title": "Users"})
    })
    r.Run(":8080")
}
```

### Echo Framework
```go
import (
    "github.com/labstack/echo/v4"
    "github.com/suserc0de/rhui"
    rhuiecho "github.com/suserc0de/rhui/adapter/echo"
)

func main() {
    engine, _ := rhui.New(rhui.WithDir("templates"))
    e := echo.New()
    e.Renderer = rhuiecho.New(engine)

    e.GET("/users", func(c echo.Context) error {
        return c.Render(http.StatusOK, "pages/users", map[string]any{"Title": "Users"})
    })
    e.Start(":8080")
}
```

### Chi Router & Standard `net/http`
```go
import (
    "github.com/go-chi/chi/v5"
    "github.com/suserc0de/rhui"
    rhuichi "github.com/suserc0de/rhui/adapter/chi"
)

func main() {
    engine, _ := rhui.New(rhui.WithDir("templates"))
    r := chi.NewRouter()

    r.Get("/dashboard", func(w http.ResponseWriter, r *http.Request) {
        rhuichi.RenderWithLayout(w, http.StatusOK, engine, "layouts/main", "pages/dashboard", map[string]any{
            "Title": "Overview",
        })
    })
    http.ListenAndServe(":8080", r)
}
```

---

## ⚡ Live Component Streaming (Server-Sent Events)

RHUI makes realtime updates effortless with native SSE:

```go
var broker = sse.NewBroker()

// 1. Client connects via SSE (e.g. hx-ext="sse" in HTMX)
http.HandleFunc("/events/notifications", func(w http.ResponseWriter, r *http.Request) {
    stream, _ := broker.Connect(w, r)
    defer stream.Close()
    stream.Wait()
})

// 2. Broadcast rendered component updates to all connected users
func OnOrderCompleted(order Order) {
    broker.BroadcastComponent(engine, "components/NotificationAlert", map[string]any{
        "Order": order,
    }, sse.EventOptions{
        EventName: "NewNotification",
    })
}
```

---

## 📖 Syntax Guide

| Syntax | Description | Example |
| :--- | :--- | :--- |
| `<ui:Name>` | Invoking a component | `<ui:Button variant="primary">Submit</ui:Button>` |
| `:attr="expr"` | Dynamic expression binding | `<button :disabled="User.IsActive == false" />` |
| `{{ expr }}` | Text interpolation | `Hello, {{ User.Name }}!` |
| `<slot>fallback</slot>` | Slot with fallback content | `<slot><p>Default content</p></slot>` |
| `<slot:name>` | Named slot | `<slot:footer>...</slot:footer>` |
| `{{ val \| filter }}` | Pipe filter transformation | `{{ Price \| rupiah }}`, `{{ Name \| upper }}` |
| `@first`, `@last`, `@index` | Loop iteration metadata | `<tr :class="@even ? 'even' : 'odd'">` |
| `fragment="id"` | Partial fragment for HTMX | `engine.RenderFragment(w, "page", "id", data)` |
| `:attrs="map"` | Dynamic attribute spreading | `<button :attrs="extra_attrs">` |
| `<if test="expr">` | Conditional block | `<if test="len(Items) > 0">...</if><else>...</else>` |
| `has_slot("name")` | Check if named slot is provided | `<if test="has_slot('actions')">...</if>` |
| `<each items="x" as="item">` | Iteration block | `<each items="Users" as="user">...</each>` |

---

## 💖 Support & Donation

If you find **RHUI** useful for your projects and want to support its ongoing development and maintenance, you can buy me a coffee or donate via PayPal:

<p align="center">
  <a href="https://paypal.me/ridwanhaqi" target="_blank">
    <img src="https://img.shields.io/badge/Donate-PayPal-blue.svg?style=for-the-badge&logo=paypal" alt="Donate with PayPal" />
  </a>
</p>

- **PayPal Email:** `ridwan.haqi@gmail.com`
- **PayPal Me:** [paypal.me/ridwanhaqi](https://paypal.me/ridwanhaqi)

Every contribution is greatly appreciated! Thank you for your support. 🙏

---

## 📄 License

Licensed under the [MIT License](LICENSE).
