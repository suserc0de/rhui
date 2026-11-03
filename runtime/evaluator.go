package runtime

import (
	"strings"
	"sync"

	"github.com/suserc0de/rhui/expr"
)

// Evaluator evaluates expressions and pipe filters.
type Evaluator struct {
	customFuncs   map[string]any
	customFilters map[string]any
}

// NewEvaluator creates a new Evaluator.
func NewEvaluator() *Evaluator {
	return &Evaluator{
		customFuncs:   make(map[string]any),
		customFilters: make(map[string]any),
	}
}

// RegisterFunc registers a custom function to the evaluator.
func (e *Evaluator) RegisterFunc(name string, fn any) {
	e.customFuncs[name] = fn
}

// RegisterFilter registers a custom filter into the evaluator.
func (e *Evaluator) RegisterFilter(name string, fn any) {
	e.customFilters[name] = fn
}

// Eval evaluates an expression string within a given Scope, utilizing expression AST caching.
func (e *Evaluator) Eval(input string, scope *Scope) (any, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return nil, nil
	}
	compiled, err := expr.Parse(trimmed)
	if err != nil {
		return nil, err
	}
	return e.EvalCompiled(compiled, scope)
}

var evalCtxPool = sync.Pool{
	New: func() any {
		return &evalContext{}
	},
}

// EvalCompiled evaluates a pre-compiled Expr AST against the scope without re-parsing.
func (e *Evaluator) EvalCompiled(compiled expr.Expr, scope *Scope) (any, error) {
	if compiled == nil {
		return nil, nil
	}
	ctx := evalCtxPool.Get().(*evalContext)
	ctx.scope = scope
	ctx.customFuncs = e.customFuncs
	ctx.customFilters = e.customFilters
	val, err := compiled.Eval(ctx)
	ctx.scope = nil
	ctx.customFuncs = nil
	ctx.customFilters = nil
	evalCtxPool.Put(ctx)
	return val, err
}

// EvalBool evaluates an expression and returns its truthiness.
func (e *Evaluator) EvalBool(input string, scope *Scope) (bool, error) {
	res, err := e.Eval(input, scope)
	if err != nil {
		return false, err
	}
	return expr.IsTruthy(res), nil
}

// EvalBoolCompiled evaluates a pre-compiled Expr AST and returns its truthiness.
func (e *Evaluator) EvalBoolCompiled(compiled expr.Expr, scope *Scope) (bool, error) {
	res, err := e.EvalCompiled(compiled, scope)
	if err != nil {
		return false, err
	}
	return expr.IsTruthy(res), nil
}

type evalContext struct {
	scope         *Scope
	customFuncs   map[string]any
	customFilters map[string]any
}

func (c *evalContext) Get(key string) (any, bool) {
	if c.scope == nil {
		return nil, false
	}
	return c.scope.Get(key)
}

func (c *evalContext) GetFunc(name string) (any, bool) {
	fn, ok := c.customFuncs[name]
	return fn, ok
}

func (c *evalContext) GetFilter(name string) (any, bool) {
	fn, ok := c.customFilters[name]
	return fn, ok
}

// IsTruthy determines whether an arbitrary value evaluates to true in conditional contexts.
func IsTruthy(val any) bool {
	return expr.IsTruthy(val)
}
