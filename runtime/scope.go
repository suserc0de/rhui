package runtime

import (
	"fmt"
	"reflect"
	"sync"
)

var scopePool = sync.Pool{
	New: func() any {
		return &Scope{
			values: make(map[string]any, 8),
		}
	},
}

// Scope represents a generic evaluated value in the template scope.
type Scope struct {
	parent *Scope
	values map[string]any
}

// NewScope creates a new root data scope.
func NewScope(data any) *Scope {
	s := acquireScope(nil)
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

func acquireScope(parent *Scope) *Scope {
	s := scopePool.Get().(*Scope)
	s.parent = parent
	for k := range s.values {
		delete(s.values, k)
	}
	return s
}

// ReleaseScope returns a Scope to the pool.
func ReleaseScope(s *Scope) {
	if s == nil {
		return
	}
	s.parent = nil
	if len(s.values) > 64 {
		return
	}
	scopePool.Put(s)
}

// SubScope creates a child scope for local variables (e.g. loops or slot scopes).
func (s *Scope) SubScope() *Scope {
	return acquireScope(s)
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

// Parent returns the parent scope.
func (s *Scope) Parent() *Scope {
	return s.parent
}
