package rhui_test

import (
	"fmt"
	"log"
	"os"
	"strings"
	"testing/fstest"

	"github.com/suserc0de/rhui"
)

// Example demonstrates standard initialization using rhui.New and rendering a template.
func Example_basic() {
	fsys := fstest.MapFS{
		"views/hello.rhui": &fstest.MapFile{
			Data: []byte(`<h1>Hello {{ Name }}!</h1>`),
		},
	}

	engine, err := rhui.New(
		rhui.WithFS(fsys),
		rhui.WithDir("views"),
	)
	if err != nil {
		log.Fatalf("failed to create engine: %v", err)
	}

	out, err := engine.RenderString("hello", map[string]any{
		"Name": "World",
	})
	if err != nil {
		log.Fatalf("failed to render: %v", err)
	}

	fmt.Println(out)
	// Output:
	// <h1>Hello World!</h1>
}

// ExampleMust demonstrates using rhui.Must for package-level or panic-on-failure initialization.
func ExampleMust() {
	fsys := fstest.MapFS{
		"templates/index.rhui": &fstest.MapFile{
			Data: []byte(`<p>Welcome {{ User }}</p>`),
		},
	}

	// Must wraps New and panics if an error occurs.
	engine := rhui.Must(rhui.New(rhui.WithFS(fsys), rhui.WithDir("templates")))

	_ = engine.Render(os.Stdout, "index", map[string]any{
		"User": "Gopher",
	})
	// Output:
	// <p>Welcome Gopher</p>
}

// ExampleWithFS demonstrates using an embedded or in-memory filesystem (io/fs).
func ExampleWithFS() {
	fsys := fstest.MapFS{
		"templates/card.rhui": &fstest.MapFile{
			Data: []byte(`<div class="card"><h3>{{ Title }}</h3><p>{{ Body }}</p></div>`),
		},
	}

	engine, err := rhui.New(
		rhui.WithFS(fsys),
		rhui.WithDir("templates"),
	)
	if err != nil {
		log.Fatal(err)
	}

	out, err := engine.RenderString("card", map[string]any{
		"Title": "RHUI Engine",
		"Body":  "Fast and expressive Go component templates.",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(out)
	// Output:
	// <div class="card"><h3>RHUI Engine</h3><p>Fast and expressive Go component templates.</p></div>
}

// ExampleEngine_RenderFragment demonstrates targeted partial rendering for HTMX or Turbo streams.
func ExampleEngine_RenderFragment() {
	fsys := fstest.MapFS{
		"feed.rhui": &fstest.MapFile{
			Data: []byte(`<div id="feed">
  <div fragment="status-badge" class="badge active">{{ Status }}</div>
  <p>Main feed content</p>
</div>`),
		},
	}

	engine, err := rhui.New(rhui.WithFS(fsys), rhui.WithDir("."))
	if err != nil {
		log.Fatal(err)
	}

	err = engine.RenderFragment(os.Stdout, "feed", "status-badge", map[string]any{
		"Status": "Online",
	})
	if err != nil {
		log.Fatal(err)
	}
	// Output:
	// <div fragment="status-badge" class="badge active">Online</div>
}

// ExampleEngine_RegisterFilter demonstrates registering custom transformation pipe filters.
func ExampleEngine_RegisterFilter() {
	fsys := fstest.MapFS{
		"item.rhui": &fstest.MapFile{
			Data: []byte(`<span>{{ Name | slugify }}</span>`),
		},
	}

	engine, err := rhui.New(rhui.WithFS(fsys), rhui.WithDir("."))
	if err != nil {
		log.Fatal(err)
	}

	engine.RegisterFilter("slugify", func(val any, args ...any) any {
		s, ok := val.(string)
		if !ok {
			return val
		}
		return strings.ToLower(strings.ReplaceAll(s, " ", "-"))
	})

	_ = engine.Render(os.Stdout, "item", map[string]any{
		"Name": "Go Component Template Engine",
	})
	// Output:
	// <span>go-component-template-engine</span>
}

// ExampleEngine_RegisterFunc demonstrates registering custom helper functions callable in expressions.
func ExampleEngine_RegisterFunc() {
	fsys := fstest.MapFS{
		"calc.rhui": &fstest.MapFile{
			Data: []byte(`<p>Total: {{ add(Price, Tax) }}</p>`),
		},
	}

	engine, err := rhui.New(rhui.WithFS(fsys), rhui.WithDir("."))
	if err != nil {
		log.Fatal(err)
	}

	engine.RegisterFunc("add", func(a, b int) int {
		return a + b
	})

	_ = engine.Render(os.Stdout, "calc", map[string]any{
		"Price": 100,
		"Tax":   11,
	})
	// Output:
	// <p>Total: 111</p>
}

// ExampleEngine_RenderWithLayout demonstrates rendering content inside a master layout template.
func ExampleEngine_RenderWithLayout() {
	fsys := fstest.MapFS{
		"layout.rhui": &fstest.MapFile{
			Data: []byte(`<html><head><title>{{ Title }}</title></head><body><slot /></body></html>`),
		},
		"home.rhui": &fstest.MapFile{
			Data: []byte(`<main><h1>{{ Heading }}</h1></main>`),
		},
	}

	engine, err := rhui.New(rhui.WithFS(fsys), rhui.WithDir("."))
	if err != nil {
		log.Fatal(err)
	}

	_ = engine.RenderWithLayout(os.Stdout, "layout", "home", map[string]any{
		"Title":   "My App",
		"Heading": "Welcome to RHUI",
	})
	// Output:
	// <html><head><title>My App</title></head><body><main><h1>Welcome to RHUI</h1></main></body></html>
}
