package rhui

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/suserc0de/rhui/ast"
	"github.com/suserc0de/rhui/compiler"
	"github.com/suserc0de/rhui/loader"
	"github.com/suserc0de/rhui/parser"
	"github.com/suserc0de/rhui/runtime"
)

// Option defines a functional option for configuring the Engine.
type Option func(*Engine)

// WithDir sets the root directory for template and component files.
// For disk-based templates, dir can be a relative or absolute path on the host filesystem.
// When used in conjunction with WithFS, dir specifies the sub-directory within the given fs.FS.
func WithDir(dir string) Option {
	return func(e *Engine) {
		e.dir = dir
	}
}

// WithExtension sets the template file extension (defaults to ".rhui").
// Only files matching this extension will be compiled by the engine.
func WithExtension(ext string) Option {
	return func(e *Engine) {
		e.extension = ext
	}
}

// WithFS sets an fs.FS (such as Go standard embed.FS) as the source of templates.
// This enables self-contained binary distributions without requiring external template folders.
//
// Example:
//
//	//go:embed templates/*
//	var viewsFS embed.FS
//
//	engine, err := rhui.New(
//	    rhui.WithFS(viewsFS),
//	    rhui.WithDir("templates"),
//	)
func WithFS(fsys fs.FS) Option {
	return func(e *Engine) {
		e.fsys = fsys
	}
}

// WithDevMode enables automatic hot-reloading of templates on every render.
// When enabled, the engine checks for changes and reloads templates seamlessly.
// Active rendering requests continue without race conditions via atomic snapshot swapping.
// Recommended only for local development environments.
func WithDevMode(enabled bool) Option {
	return func(e *Engine) {
		e.devMode = enabled
	}
}

// WithMaxRenderDepth sets the maximum allowable recursion and nesting depth for components.
// Defaults to 64. Guards against stack overflow and CPU starvation caused by circular component references.
func WithMaxRenderDepth(depth int) Option {
	return func(e *Engine) {
		e.maxRenderDepth = depth
	}
}

// engineState holds an atomic, immutable snapshot of the compiled templates and renderer.
type engineState struct {
	templates map[string]*ast.RootNode
	registry  *compiler.Registry
	renderer  *runtime.Renderer
}

// Engine represents the top-level RHUI template engine.
type Engine struct {
	dir            string
	extension      string
	fsys           fs.FS
	devMode        bool
	maxRenderDepth int
	loader         loader.Loader

	evaluator *runtime.Evaluator
	state     atomic.Pointer[engineState]
	reloadMu  sync.Mutex // serializes Reload() and ParseTemplate()
}

// New creates and initializes a new RHUI Engine.
func New(opts ...Option) (*Engine, error) {
	e := &Engine{
		dir:            "templates",
		extension:      ".rhui",
		maxRenderDepth: runtime.DefaultMaxRenderDepth,
		evaluator:      runtime.NewEvaluator(),
	}

	for _, opt := range opts {
		opt(e)
	}

	initialRegistry := compiler.NewRegistry()
	initialRenderer := runtime.NewRenderer(initialRegistry, e.evaluator)
	initialRenderer.SetMaxRenderDepth(e.maxRenderDepth)

	initialState := &engineState{
		templates: make(map[string]*ast.RootNode),
		registry:  initialRegistry,
		renderer:  initialRenderer,
	}
	e.state.Store(initialState)

	// Configure appropriate Loader (embed.FS or Disk)
	if e.fsys != nil {
		e.loader = loader.NewFSLoader(e.fsys, e.dir)
	} else if fi, err := os.Stat(e.dir); err == nil && fi.IsDir() {
		e.loader = loader.NewDiskLoader(e.dir)
	}

	if e.loader != nil {
		if err := e.Reload(); err != nil {
			return nil, err
		}
	}

	return e, nil
}

// Must is a convenience helper that wraps a call to New (or any function returning (*Engine, error))
// and panics if the error is non-nil. It follows idiomatic Go patterns (similar to template.Must)
// and is intended for use in variable initializations or quickstart bootstraps.
//
// Example:
//
//	var engine = rhui.Must(rhui.New(rhui.WithDir("templates")))
func Must(e *Engine, err error) *Engine {
	if err != nil {
		panic(err)
	}
	return e
}

// EngineOptions configures the Engine when using NewEngine.
// Deprecated: Configure the Engine using functional options with New() instead.
type EngineOptions struct {
	ViewsDir   string
	AutoReload bool
}

// NewEngine is a legacy convenience constructor that creates an Engine from an options struct,
// panicking on fatal initialization error.
//
// Deprecated: Use New() which returns a clear error instead of panicking, or wrap with Must(New(...))
// if fatal panic on startup failure is explicitly desired.
func NewEngine(opts ...EngineOptions) *Engine {
	var optList []Option
	if len(opts) > 0 {
		if opts[0].ViewsDir != "" {
			optList = append(optList, WithDir(opts[0].ViewsDir))
		}
		if opts[0].AutoReload {
			optList = append(optList, WithDevMode(true))
		}
	}
	return Must(New(optList...))
}

// RegisterFunc registers a custom helper function to the engine's expression evaluator.
// The function can then be invoked inside template expressions such as conditionals:
//
//	<if test="myHelper(user.Role, 'admin')">
//	  <span>Superuser</span>
//	</if>
func (e *Engine) RegisterFunc(name string, fn any) {
	e.evaluator.RegisterFunc(name, fn)
}

// RegisterFilter registers a custom pipe filter into the engine evaluator.
// Filters can be chained using the pipe '|' syntax in interpolations and expressions:
//
//	{{ user.Balance | rupiah }}
//	{{ user.Name | upper }}
//
// Supported filter signatures include:
//   - func(val T) R
//   - func(val T, arg A) R
func (e *Engine) RegisterFilter(name string, fn any) {
	e.evaluator.RegisterFilter(name, fn)
}

// Reload compiles and refreshes all templates and components from the configured loader.
// Reload is thread-safe and atomic: active rendering requests proceed on the existing immutable
// snapshot without race conditions, and swap to the new snapshot upon successful compilation.
func (e *Engine) Reload() error {
	if e.loader == nil {
		return nil
	}

	e.reloadMu.Lock()
	defer e.reloadMu.Unlock()

	newTemplates := make(map[string]*ast.RootNode)
	newRegistry := compiler.NewRegistry()

	err := e.loader.Walk(e.extension, func(name string, content []byte) error {
		p := parser.NewParser(parser.NewLexer(name, string(content)))
		root, err := p.Parse()
		if err != nil {
			return err
		}

		// Register component definitions
		for _, child := range root.Children {
			if compDef, ok := child.(*ast.ComponentDefNode); ok {
				newRegistry.Register(compDef.Name, compDef)
			}
		}

		newTemplates[filepath.Clean(name)] = root
		return nil
	})

	if err != nil {
		return err
	}

	newRenderer := runtime.NewRenderer(newRegistry, e.evaluator)
	newRenderer.SetMaxRenderDepth(e.maxRenderDepth)
	newState := &engineState{
		templates: newTemplates,
		registry:  newRegistry,
		renderer:  newRenderer,
	}
	e.state.Store(newState)
	return nil
}

// RenderFragment renders only a specific DOM node bearing fragment="fragmentID" from the template.
// This is exceptionally useful for HTMX partial swaps and targeted component re-renders
// without having to re-render the entire enclosing document or layout.
//
// Example in Template (pages/users.rhui):
//
//	<table fragment="users-table">
//	  <each in="Users" as="user">
//	    <tr><td>{{ user.Name }}</td></tr>
//	  </each>
//	</table>
//
// Handler invocation:
//
//	err := engine.RenderFragment(w, "pages/users", "users-table", map[string]any{"Users": users})
func (e *Engine) RenderFragment(w io.Writer, name, fragmentID string, data any) error {
	if e.devMode {
		_ = e.Reload()
	}

	snapshot := e.state.Load()
	cleanName := filepath.Clean(name)
	root, ok := snapshot.templates[cleanName]
	if !ok {
		root, ok = snapshot.templates[cleanName+e.extension]
	}

	if !ok {
		return fmt.Errorf("template %q not found", name)
	}

	fragmentNode := findFragment(root, fragmentID)
	if fragmentNode == nil {
		return fmt.Errorf("fragment %q not found in template %q", fragmentID, name)
	}

	scope := runtime.NewScope(data)
	defer runtime.ReleaseScope(scope)
	return snapshot.renderer.RenderNode(w, fragmentNode, scope)
}

func findFragment(node ast.Node, id string) ast.Node {
	switch n := node.(type) {
	case *ast.RootNode:
		for _, child := range n.Children {
			if match := findFragment(child, id); match != nil {
				return match
			}
		}
	case *ast.ElementNode:
		for _, attr := range n.Attributes {
			if attr.Name == "fragment" && attr.Value == id {
				return n
			}
		}
		for _, child := range n.Children {
			if match := findFragment(child, id); match != nil {
				return match
			}
		}
	case *ast.ComponentNode:
		for _, child := range n.DefaultSlot {
			if match := findFragment(child, id); match != nil {
				return match
			}
		}
		for _, slotNodes := range n.NamedSlots {
			for _, child := range slotNodes {
				if match := findFragment(child, id); match != nil {
					return match
				}
			}
		}
	}
	return nil
}

// Render renders a compiled template by name with data into the provided io.Writer.
// In dev mode (enabled via WithDevMode(true)), the filesystem is re-scanned and templates
// reloaded automatically on each render call.
//
// Example:
//
//	err := engine.Render(w, "index.rhui", map[string]any{"Title": "Home"})
//	if err != nil {
//	    http.Error(w, err.Error(), http.StatusInternalServerError)
//	}
func (e *Engine) Render(w io.Writer, name string, data any) error {
	if e.devMode {
		_ = e.Reload()
	}

	snapshot := e.state.Load()
	cleanName := filepath.Clean(name)
	root, ok := snapshot.templates[cleanName]
	if !ok {
		root, ok = snapshot.templates[cleanName+e.extension]
	}

	if !ok {
		return fmt.Errorf("template %q not found", name)
	}

	scope := runtime.NewScope(data)
	defer runtime.ReleaseScope(scope)
	return snapshot.renderer.RenderNode(w, root, scope)
}

// RenderWithLayout renders a content template inside an outer layout template.
// The rendered content template output is injected into the layout's default slot `<slot />`.
// Both layout and page have access to the data passed in data.
//
// Example:
//
//	// layout.rhui:
//	// <!DOCTYPE html><html><body><slot /></body></html>
//	// page.rhui:
//	// <h1>{{ Title }}</h1>
//	err := engine.RenderWithLayout(w, "layout.rhui", "page.rhui", map[string]any{
//	    "Title": "Dashboard",
//	})
func (e *Engine) RenderWithLayout(w io.Writer, layoutName, pageName string, data any) error {
	buf := runtime.GetBuffer()
	defer runtime.PutBuffer(buf)

	if err := e.Render(buf, pageName, data); err != nil {
		return fmt.Errorf("render page %q failed: %w", pageName, err)
	}

	layoutData := make(map[string]any)
	if m, ok := data.(map[string]any); ok {
		for k, v := range m {
			layoutData[k] = v
		}
	}
	pageHTML := buf.String()

	snapshot := e.state.Load()
	cleanLayout := filepath.Clean(layoutName)
	layoutRoot, ok := snapshot.templates[cleanLayout]
	if !ok {
		layoutRoot, ok = snapshot.templates[cleanLayout+e.extension]
	}

	if !ok {
		return fmt.Errorf("layout %q not found", layoutName)
	}

	scope := runtime.NewScope(layoutData)
	defer runtime.ReleaseScope(scope)
	scope.Set("__default_slot__", []ast.Node{
		&ast.TextNode{Content: pageHTML},
	})

	return snapshot.renderer.RenderNode(w, layoutRoot, scope)
}

// RenderString renders a template to an in-memory string using RHUI's zero-allocation buffer pool.
//
// Example:
//
//	html, err := engine.RenderString("greeting.rhui", map[string]any{"Name": "Alice"})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(html)
func (e *Engine) RenderString(name string, data any) (string, error) {
	buf := runtime.GetBuffer()
	defer runtime.PutBuffer(buf)

	err := e.Render(buf, name, data)
	return buf.String(), err
}

// ParseTemplate parses and registers a raw template string directly under a given name.
// This is thread-safe and atomic: active renders proceed on previous snapshots while the
// new template AST is registered.
//
// Example:
//
//	err := engine.ParseTemplate("inline.rhui", "<div>Hello {{ Name }}!</div>")
//	if err != nil {
//	    log.Fatalf("failed to parse template: %v", err)
//	}
func (e *Engine) ParseTemplate(name, content string) error {
	p := parser.NewParser(parser.NewLexer(name, content))
	root, err := p.Parse()
	if err != nil {
		return err
	}

	e.reloadMu.Lock()
	defer e.reloadMu.Unlock()

	current := e.state.Load()
	newTemplates := make(map[string]*ast.RootNode, len(current.templates)+1)
	for k, v := range current.templates {
		newTemplates[k] = v
	}
	newRegistry := compiler.NewRegistry()
	if current.registry != nil {
		for _, comp := range current.registry.All() {
			newRegistry.RegisterMeta(comp)
		}
	}

	for _, child := range root.Children {
		if compDef, ok := child.(*ast.ComponentDefNode); ok {
			newRegistry.Register(compDef.Name, compDef)
		}
	}

	newTemplates[filepath.Clean(name)] = root
	newRenderer := runtime.NewRenderer(newRegistry, e.evaluator)
	newRenderer.SetMaxRenderDepth(e.maxRenderDepth)
	newState := &engineState{
		templates: newTemplates,
		registry:  newRegistry,
		renderer:  newRenderer,
	}
	e.state.Store(newState)
	return nil
}
