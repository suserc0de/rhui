package tests

import (
	"testing"

	"github.com/suserc0de/rhui/parser"
	"github.com/suserc0de/rhui/runtime"
)

// FuzzParser ensures that the lexer and parser never crash or panic on arbitrary input.
func FuzzParser(f *testing.F) {
	// Seed corpus with representative templates
	seeds := []string{
		`<div>Hello World</div>`,
		`<ui:Card title="Title"><p>Content</p></ui:Card>`,
		`<if test="len(Users) > 0"><each items="Users" as="u">{{ u.Name }}</each></if>`,
		`<button :disabled="User.IsActive == false" :attrs="extra_attrs">Click</button>`,
		`<slot:actions><ui:Button>Confirm</ui:Button></slot:actions>`,
		`{{ Price | rupiah | upper }}`,
		`<!-- Unclosed comment`,
		`<div class="broken" <p>`,
		`<tag attr=value >`,
		`{{ unclosed interpolation`,
		`<if test="invalid ==">`,
		`<each items="Users"`,
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// Test Lexer
		lex := parser.NewLexer("fuzz.rhui", input)
		p := parser.NewParser(lex)

		// Must not panic!
		root, err := p.Parse()
		if err != nil {
			return
		}

		// If parsed successfully into AST, renderer must also never panic
		if root != nil {
			renderer := runtime.NewRenderer(nil, nil)
			scope := runtime.NewScope(map[string]any{
				"Users": []string{"A", "B"},
				"Price": 50000,
			})
			var buf discardWriter
			_ = renderer.RenderNode(buf, root, scope)
		}
	})
}

type discardWriter struct{}

func (d discardWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}
