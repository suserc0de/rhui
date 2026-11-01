package echo

import (
	"io"

	"github.com/coderhaqi/rhui"
	"github.com/labstack/echo/v4"
)

// Renderer implements echo.Renderer interface.
type Renderer struct {
	engine *rhui.Engine
}

// New creates an Echo Renderer adapter for RHUI.
func New(engine *rhui.Engine) *Renderer {
	return &Renderer{
		engine: engine,
	}
}

// Render renders a RHUI template view.
func (r *Renderer) Render(w io.Writer, name string, data any, c echo.Context) error {
	return r.engine.Render(w, name, data)
}
