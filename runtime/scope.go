package runtime

import (
	"fmt"
	"reflect"
)

// Value represents a generic evaluated value in the template scope.
type Scope struct {
	parent *Scope
	values map[string]any
}

// NewScope creates a new root data scope.
func NewScope(data any) *Scope {
	s := &Scope{
		values: make(map[string]any),
	}
	if data == nil {
		return s
	}

	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Map {
		for _, k := range val.MapKeys() {
			s.values[fmt.Sprint(k.Interface())] = val.MapIndex(k).Interface()
		}
	} else if val.Kind() == reflect.Struct {
		typ := val.Type()
		for i := 0; i < val.NumField(); i++ {
			s.values[typ.Field(i).Name] = val.Field(i).Interface()
		}
		s.values["."] = data
	} else {
		s.values["."] = data
	}
	return s
}

// SubScope creates a child scope for local variables (e.g. loops or slot scopes).
func (s *Scope) SubScope() *Scope {
	return &Scope{
		parent: s,
		values: make(map[string]any),
	}
}

// Set sets a variable in the current local scope.
func (s *Scope) Set(key string, val any) {
	s.values[key] = val
}

// Get resolves a variable by key, searching outwards through parent scopes.
func (s *Scope) Get(key string) (any, bool) {
	if val, ok := s.values[key]; ok {
		return val, true
	}
	if s.parent != nil {
		return s.parent.Get(key)
	}
	return nil, false
}
