package gin

import (
	"net/http"

	"github.com/coderhaqi/rhui"
	"github.com/gin-gonic/gin/render"
)

// Render implements gin's render.HTMLRender interface.
type Render struct {
	engine *rhui.Engine
}

// New creates a new Gin HTMLRender adapter for RHUI.
func New(engine *rhui.Engine) *Render {
	return &Render{
		engine: engine,
	}
}

// Instance returns an instance of gin's render.Render.
func (r *Render) Instance(name string, data any) render.Render {
	return &HTML{
		engine: r.engine,
		name:   name,
		data:   data,
	}
}

// HTML implements gin's render.Render interface.
type HTML struct {
	engine *rhui.Engine
	name   string
	data   any
}

// Render writes HTML to the response writer.
func (h *HTML) Render(w http.ResponseWriter) error {
	h.WriteContentType(w)
	return h.engine.Render(w, h.name, h.data)
}

// WriteContentType sets the Content-Type header.
func (h *HTML) WriteContentType(w http.ResponseWriter) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = []string{"text/html; charset=utf-8"}
	}
}
