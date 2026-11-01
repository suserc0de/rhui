package rhui

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/coderhaqi/rhui/ast"
	"github.com/coderhaqi/rhui/compiler"
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

// Engine represents the top-level RHUI template engine.
type Engine struct {
	dir       string
	extension string
	registry  *compiler.Registry
	evaluator *runtime.Evaluator
	renderer  *runtime.Renderer
	templates map[string]*ast.RootNode
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

	// If directory exists, automatically load and compile components & templates
	if fi, err := os.Stat(e.dir); err == nil && fi.IsDir() {
		if err := e.loadAll(); err != nil {
			return nil, err
		}
	}

	return e, nil
}

// RegisterFunc registers a custom function to the engine evaluator.
func (e *Engine) RegisterFunc(name string, fn any) {
	e.evaluator.RegisterFunc(name, fn)
}

// Render renders a compiled template by name with data into w.
func (e *Engine) Render(w io.Writer, name string, data any) error {
	cleanName := filepath.Clean(name)
	root, ok := e.templates[cleanName]
	if !ok {
		// Try appending extension if omitted
		cleanNameWithExt := cleanName + e.extension
		root, ok = e.templates[cleanNameWithExt]
		if !ok {
			return fmt.Errorf("template %q not found", name)
		}
	}

	scope := runtime.NewScope(data)
	return e.renderer.RenderNode(w, root, scope)
}

// RenderString renders a template to string.
func (e *Engine) RenderString(name string, data any) (string, error) {
	var sb strings.Builder
	err := e.Render(&sb, name, data)
	return sb.String(), err
}

// ParseTemplate parses and registers a raw template string directly under a given name.
func (e *Engine) ParseTemplate(name, content string) error {
	p := parser.NewParser(parser.NewLexer(name, content))
	root, err := p.Parse()
	if err != nil {
		return err
	}

	// Register any <component name="..."> definitions found inside
	for _, child := range root.Children {
		if compDef, ok := child.(*ast.ComponentDefNode); ok {
			e.registry.Register(compDef.Name, compDef)
		}
	}

	e.templates[filepath.Clean(name)] = root
	return nil
}

func (e *Engine) loadAll() error {
	return filepath.Walk(e.dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) == e.extension {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			relPath, _ := filepath.Rel(e.dir, path)
			// e.g. "pages/users/index" or "components/Card"
			key := strings.TrimSuffix(relPath, e.extension)
			return e.ParseTemplate(key, string(content))
		}
		return nil
	})
}
