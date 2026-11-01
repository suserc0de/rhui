package nethttp

import (
	"net/http"

	"github.com/coderhaqi/rhui"
)

// Render renders a RHUI template to a standard http.ResponseWriter.
func Render(w http.ResponseWriter, status int, engine *rhui.Engine, name string, data any) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	return engine.Render(w, name, data)
}
