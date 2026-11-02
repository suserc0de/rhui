package chi

import (
	"net/http"

	"github.com/suserc0de/rhui"
)

// Render renders a RHUI template to an http.ResponseWriter within a Chi handler.
func Render(w http.ResponseWriter, status int, engine *rhui.Engine, name string, data any) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	return engine.Render(w, name, data)
}

// RenderWithLayout renders a page inside a layout directly within a Chi handler.
func RenderWithLayout(w http.ResponseWriter, status int, engine *rhui.Engine, layoutName, pageName string, data any) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	return engine.RenderWithLayout(w, layoutName, pageName, data)
}
