package rhui

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"github.com/coderhaqi/rhui/ast"
	"github.com/coderhaqi/rhui/compiler"
	"github.com/coderhaqi/rhui/loader"
	"github.com/coderhaqi/rhui/parser"
	"github.com/coderhaqi/rhui/runtime"
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

// Engine represents the top-level RHUI template engine.
type Engine struct {
	dir       string
	extension string
	fsys      fs.FS
	devMode   bool
	loader    loader.Loader

	registry  *compiler.Registry
	evaluator *runtime.Evaluator
	renderer  *runtime.Renderer

	templates map[string]*ast.RootNode
	mu        sync.RWMutex
}

// New creates and initializes a new RHUI Engine.
func New(opts ...Option) (*Engine, error) {
	e := &Engine{
		dir:       "templates",
		extension: ".rhui",
		registry:  compiler.NewRegistry(),
		evaluator: runtime.NewEvaluator(),
		templates: make(map[string]*ast.RootNode),
	}

	for _, opt := range opts {
		opt(e)
	}

	e.renderer = runtime.NewRenderer(e.registry, e.evaluator)

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

// RegisterFunc registers a custom function to the engine evaluator.
func (e *Engine) RegisterFunc(name string, fn any) {
	e.evaluator.RegisterFunc(name, fn)
}

// Reload compiles and refreshes all templates and components from the loader.
func (e *Engine) Reload() error {
	if e.loader == nil {
		return nil
	}

	e.mu.Lock()
	defer e.mu.Unlock()

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

	e.templates = newTemplates
	e.registry = newRegistry
	e.renderer = runtime.NewRenderer(e.registry, e.evaluator)
	return nil
}

// Render renders a compiled template by name with data into w.
func (e *Engine) Render(w io.Writer, name string, data any) error {
	if e.devMode {
		_ = e.Reload()
	}

	cleanName := filepath.Clean(name)

	e.mu.RLock()
	root, ok := e.templates[cleanName]
	if !ok {
		// Try fallback with extension
		cleanNameWithExt := cleanName + e.extension
		root, ok = e.templates[cleanNameWithExt]
	}
	e.mu.RUnlock()

	if !ok {
		return fmt.Errorf("template %q not found", name)
	}

	scope := runtime.NewScope(data)
	return e.renderer.RenderNode(w, root, scope)
}

// RenderWithLayout renders a content template inside a layout template.
// The content template output is injected as the default slot of the layout.
func (e *Engine) RenderWithLayout(w io.Writer, layoutName, pageName string, data any) error {
	// 1. Render page content using buffer pool
	buf := runtime.GetBuffer()
	defer runtime.PutBuffer(buf)

	if err := e.Render(buf, pageName, data); err != nil {
		return fmt.Errorf("render page %q failed: %w", pageName, err)
	}

	// 2. Prepare layout data
	layoutData := make(map[string]any)
	if m, ok := data.(map[string]any); ok {
		for k, v := range m {
			layoutData[k] = v
		}
	}
	pageHTML := buf.String()

	cleanLayout := filepath.Clean(layoutName)
	e.mu.RLock()
	layoutRoot, ok := e.templates[cleanLayout]
	if !ok {
		layoutRoot, ok = e.templates[cleanLayout+e.extension]
	}
	e.mu.RUnlock()

	if !ok {
		return fmt.Errorf("layout %q not found", layoutName)
	}

	// Create scope and pass pageHTML into default slot
	scope := runtime.NewScope(layoutData)
	scope.Set("__default_slot__", []ast.Node{
		&ast.TextNode{Content: pageHTML},
	})

	return e.renderer.RenderNode(w, layoutRoot, scope)
}

// RenderString renders a template to string using buffer pool.
func (e *Engine) RenderString(name string, data any) (string, error) {
	buf := runtime.GetBuffer()
	defer runtime.PutBuffer(buf)

	err := e.Render(buf, name, data)
	return buf.String(), err
}

// ParseTemplate parses and registers a raw template string directly under a given name.
func (e *Engine) ParseTemplate(name, content string) error {
	p := parser.NewParser(parser.NewLexer(name, content))
	root, err := p.Parse()
	if err != nil {
		return err
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	// Register any <component name="..."> definitions found inside
	for _, child := range root.Children {
		if compDef, ok := child.(*ast.ComponentDefNode); ok {
			e.registry.Register(compDef.Name, compDef)
		}
	}

	e.templates[filepath.Clean(name)] = root
	return nil
}
