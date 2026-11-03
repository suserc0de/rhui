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
func WithDir(dir string) Option {
	return func(e *Engine) {
		e.dir = dir
	}
}

// WithExtension sets the template file extension (defaults to ".rhui").
func WithExtension(ext string) Option {
	return func(e *Engine) {
		e.extension = ext
	}
}

// WithFS sets an fs.FS (such as embed.FS) as the source of templates.
func WithFS(fsys fs.FS) Option {
	return func(e *Engine) {
		e.fsys = fsys
	}
}

// WithDevMode enables automatic reloading of templates on render.
func WithDevMode(enabled bool) Option {
	return func(e *Engine) {
		e.devMode = enabled
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
	dir       string
	extension string
	fsys      fs.FS
	devMode   bool
	loader    loader.Loader

	evaluator *runtime.Evaluator
	state     atomic.Pointer[engineState]
	reloadMu  sync.Mutex // serializes Reload() and ParseTemplate()
}

// New creates and initializes a new RHUI Engine.
func New(opts ...Option) (*Engine, error) {
	e := &Engine{
		dir:       "templates",
		extension: ".rhui",
		evaluator: runtime.NewEvaluator(),
	}

	initialRegistry := compiler.NewRegistry()
	initialState := &engineState{
		templates: make(map[string]*ast.RootNode),
		registry:  initialRegistry,
		renderer:  runtime.NewRenderer(initialRegistry, e.evaluator),
	}
	e.state.Store(initialState)

	for _, opt := range opts {
		opt(e)
	}

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

// EngineOptions configures the Engine when using NewEngine.
type EngineOptions struct {
	ViewsDir   string
	AutoReload bool
}

// NewEngine is a convenience constructor that creates an Engine from options struct, panicking on fatal initialization error.
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
	e, err := New(optList...)
	if err != nil {
		panic(err)
	}
	return e
}

// RegisterFunc registers a custom function to the engine evaluator.
func (e *Engine) RegisterFunc(name string, fn any) {
	e.evaluator.RegisterFunc(name, fn)
}

// RegisterFilter registers a custom filter into the engine evaluator.
func (e *Engine) RegisterFilter(name string, fn any) {
	e.evaluator.RegisterFilter(name, fn)
}

// Reload compiles and refreshes all templates and components from the loader.
// Thread-safe and atomic: active renders proceed on the previous snapshot without race conditions.
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

	newState := &engineState{
		templates: newTemplates,
		registry:  newRegistry,
		renderer:  runtime.NewRenderer(newRegistry, e.evaluator),
	}
	e.state.Store(newState)
	return nil
}

// RenderFragment renders only a specific node with fragment="fragmentID" from the template.
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

// Render renders a compiled template by name with data into w.
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

// RenderWithLayout renders a content template inside a layout template.
// The content template output is injected as the default slot of the layout.
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

// RenderString renders a template to string using buffer pool.
func (e *Engine) RenderString(name string, data any) (string, error) {
	buf := runtime.GetBuffer()
	defer runtime.PutBuffer(buf)

	err := e.Render(buf, name, data)
	return buf.String(), err
}

// ParseTemplate parses and registers a raw template string directly under a given name.
// Thread-safe and atomic: active renders proceed on previous snapshots.
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
	newState := &engineState{
		templates: newTemplates,
		registry:  newRegistry,
		renderer:  runtime.NewRenderer(newRegistry, e.evaluator),
	}
	e.state.Store(newState)
	return nil
}
