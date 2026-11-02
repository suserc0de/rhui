package tests

import (
	"testing"

	"github.com/suserc0de/rhui/runtime"
)

// FuzzEvaluator verifies that arbitrary comparisons, logic, and expressions
// never panic the sandboxed expression evaluator.
func FuzzEvaluator(f *testing.F) {
	seeds := []string{
		"10 > 5",
		"price * quantity",
		"user.IsActive && count > 0",
		"status == 'active'",
		"title == 'test'",
		"1 == 0",
		"true && false || !true",
		"nested.a.b.c.d.e",
		"''''''''",
		"((((1 + 2))))",
		"unknown_var",
		"price > 10000 && status != 'inactive'",
		"-99999999999999999999999999999999",
	}

	for _, s := range seeds {
		f.Add(s)
	}

	scope := runtime.NewScope(map[string]any{
		"price":    15000,
		"quantity": 3,
		"count":    10,
		"status":   "active",
		"user": map[string]any{
			"Name":     "Haqi",
			"IsActive": true,
		},
		"title": "rhui engine",
	})

	f.Fuzz(func(t *testing.T, exprStr string) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("Evaluator panicked on input: %q, panic: %v", exprStr, r)
			}
		}()

		evaluator := runtime.NewEvaluator()
		_, _ = evaluator.Eval(exprStr, scope)
	})
}
