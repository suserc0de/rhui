package tests

import (
	"embed"
	"strings"
	"testing"

	"github.com/coderhaqi/rhui"
	"github.com/coderhaqi/rhui/runtime"
)

//go:embed testdata/*
var testEmbedFS embed.FS

func TestEmbedFSLoader(t *testing.T) {
	engine, err := rhui.New(
		rhui.WithFS(testEmbedFS),
		rhui.WithDir("testdata"),
		rhui.WithExtension(".rhui"),
	)
	if err != nil {
		t.Fatalf("failed to initialize engine with embed.FS: %v", err)
	}

	out, err := engine.RenderString("pages/hello", map[string]any{
		"Name": "EmbedUser",
	})
	if err != nil {
		t.Fatalf("render embed template failed: %v", err)
	}

	if !strings.Contains(out, "Hello, EmbedUser!") {
		t.Fatalf("expected 'Hello, EmbedUser!', got: %s", out)
	}
}

func TestRenderWithLayout(t *testing.T) {
	engine, err := rhui.New()
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	layoutTmpl := `<!DOCTYPE html>
<html>
<head><title>{{ Title }}</title></head>
<body>
  <div class="sidebar">Nav</div>
  <main><slot /></main>
</body>
</html>`

	pageTmpl := `<div class="content"><h1>Welcome</h1><p>User page content</p></div>`

	err = engine.ParseTemplate("layouts/main", layoutTmpl)
	if err != nil {
		t.Fatalf("parse layout failed: %v", err)
	}
	err = engine.ParseTemplate("pages/home", pageTmpl)
	if err != nil {
		t.Fatalf("parse page failed: %v", err)
	}

	// Now test RenderWithLayout
	var buf strings.Builder
	err = engine.RenderWithLayout(&buf, "layouts/main", "pages/home", map[string]any{
		"Title": "Dashboard with Layout",
	})
	if err != nil {
		t.Fatalf("RenderWithLayout failed: %v", err)
	}

	res := buf.String()
	if !strings.Contains(res, "<title>Dashboard with Layout</title>") {
		t.Errorf("missing layout title: %s", res)
	}
	if !strings.Contains(res, "<main><div class=\"content\"><h1>Welcome</h1><p>User page content</p></div></main>") {
		t.Errorf("missing page content in layout slot: %s", res)
	}
}

func TestBufferPoolPerformance(t *testing.T) {
	buf := runtime.GetBuffer()
	buf.WriteString("Testing buffer pool")
	if buf.String() != "Testing buffer pool" {
		t.Fatalf("buffer content mismatch")
	}
	runtime.PutBuffer(buf)

	buf2 := runtime.GetBuffer()
	if buf2.Len() != 0 {
		t.Fatalf("buffer was not reset on retrieval from pool")
	}
	runtime.PutBuffer(buf2)
}
