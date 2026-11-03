package compiler

import (
	"fmt"

	"github.com/suserc0de/rhui/ast"
)

// ComponentMeta holds compiled metadata and AST representation of a component.
type ComponentMeta struct {
	Name     string
	Props    map[string]ast.PropDef
	Children []ast.Node
}

// Registry manages registered reusable components.
type Registry struct {
	components map[string]*ComponentMeta
}

// NewRegistry creates a new component Registry.
func NewRegistry() *Registry {
	return &Registry{
		components: make(map[string]*ComponentMeta),
	}
}

// Register registers a component definition into the registry.
func (r *Registry) Register(name string, def *ast.ComponentDefNode) {
	meta := &ComponentMeta{
		Name:     name,
		Props:    make(map[string]ast.PropDef),
		Children: def.Children,
	}
	for _, p := range def.Props {
		meta.Props[p.Name] = p
	}
	r.components[name] = meta
}

// RegisterMeta directly registers a ComponentMeta.
func (r *Registry) RegisterMeta(meta *ComponentMeta) {
	r.components[meta.Name] = meta
}

// Get returns component metadata if registered.
func (r *Registry) Get(name string) (*ComponentMeta, bool) {
	meta, ok := r.components[name]
	return meta, ok
}

// All returns all registered ComponentMeta instances.
func (r *Registry) All() []*ComponentMeta {
	list := make([]*ComponentMeta, 0, len(r.components))
	for _, m := range r.components {
		list = append(list, m)
	}
	return list
}

// List returns all registered component names.
func (r *Registry) List() []string {
	var list []string
	for k := range r.components {
		list = append(list, k)
	}
	return list
}

func (r *Registry) String() string {
	return fmt.Sprintf("ComponentRegistry(%d components)", len(r.components))
}
