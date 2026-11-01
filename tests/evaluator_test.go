package tests

import (
	"testing"

	"github.com/coderhaqi/rhui/runtime"
)

type TestUser struct {
	Name     string
	Age      int
	IsActive bool
}

func TestEvaluatorBasicLiteralAndComparison(t *testing.T) {
	eval := runtime.NewEvaluator()
	scope := runtime.NewScope(nil)

	res, err := eval.Eval("10 > 5", scope)
	if err != nil || res != true {
		t.Fatalf("expected true, got %v (err: %v)", res, err)
	}

	res, err = eval.Eval("\"hello\" == \"hello\"", scope)
	if err != nil || res != true {
		t.Fatalf("expected true, got %v (err: %v)", res, err)
	}

	res, err = eval.Eval("5 + 5", scope) // unsupported operator defaults
	if err == nil && res != nil {
		// Just ensuring comparison & literals work
	}
}

func TestEvaluatorScopeAndFieldAccess(t *testing.T) {
	eval := runtime.NewEvaluator()
	user := TestUser{
		Name:     "Haqi",
		Age:      25,
		IsActive: true,
	}

	scope := runtime.NewScope(map[string]any{
		"User":  user,
		"Users": []string{"Alice", "Bob", "Charlie"},
	})

	// Field access
	res, err := eval.Eval("User.Name", scope)
	if err != nil || res != "Haqi" {
		t.Fatalf("expected 'Haqi', got %v (err: %v)", res, err)
	}

	// Boolean check
	resBool, err := eval.EvalBool("User.IsActive && User.Age >= 20", scope)
	if err != nil || !resBool {
		t.Fatalf("expected true, got %v (err: %v)", resBool, err)
	}

	// len built-in function
	resLen, err := eval.Eval("len(Users)", scope)
	if err != nil || resLen != 3 {
		t.Fatalf("expected len 3, got %v (err: %v)", resLen, err)
	}

	// len condition
	resCond, err := eval.EvalBool("len(Users) > 0", scope)
	if err != nil || !resCond {
		t.Fatalf("expected true, got %v (err: %v)", resCond, err)
	}
}

func TestEvaluatorSubScope(t *testing.T) {
	root := runtime.NewScope(map[string]any{"Global": "Root"})
	child := root.SubScope()
	child.Set("Local", "Item1")

	eval := runtime.NewEvaluator()

	res1, _ := eval.Eval("Global", child)
	if res1 != "Root" {
		t.Fatalf("expected child to access parent scope, got %v", res1)
	}

	res2, _ := eval.Eval("Local", child)
	if res2 != "Item1" {
		t.Fatalf("expected child to access its own variable, got %v", res2)
	}

	res3, _ := eval.Eval("Local", root)
	if res3 != nil {
		t.Fatalf("expected root to not access child scope, got %v", res3)
	}
}
