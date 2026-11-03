package expr

// Expr represents a pre-compiled expression AST node that can be evaluated against a Context.
type Expr interface {
	Eval(ctx Context) (any, error)
}

// Context provides variable resolution, functions, and filters for expression evaluation.
type Context interface {
	Get(key string) (any, bool)
	GetFunc(name string) (any, bool)
	GetFilter(name string) (any, bool)
}
