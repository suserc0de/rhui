package fiber

import (
	"io"

	"github.com/suserc0de/rhui"
)

// Engine implements Fiber's fiber.Views interface.
type Engine struct {
	engine *rhui.Engine
}

// New creates a new Fiber Views adapter for RHUI.
func New(engine *rhui.Engine) *Engine {
	return &Engine{
		engine: engine,
	}
}

// Load loads templates (handled automatically by RHUI engine).
func (e *Engine) Load() error {
	return nil
}

// Render renders a RHUI template with given data and layout.
func (e *Engine) Render(out io.Writer, template string, binding interface{}, layout ...string) error {
	return e.engine.Render(out, template, binding)
}
