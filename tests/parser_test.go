package tests

import (
	"testing"

	"github.com/suserc0de/rhui/ast"
	"github.com/suserc0de/rhui/parser"
)

func TestParserBasicElement(t *testing.T) {
	input := `<div class="container"><p>Hello World</p></div>`
	lex := parser.NewLexer("test.rhui", input)
	p := parser.NewParser(lex)

	root, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if len(root.Children) != 1 {
		t.Fatalf("expected 1 root child, got %d", len(root.Children))
	}

	div, ok := root.Children[0].(*ast.ElementNode)
	if !ok {
		t.Fatalf("expected ElementNode, got %T", root.Children[0])
	}
	if div.TagName != "div" {
		t.Fatalf("expected tag 'div', got %q", div.TagName)
	}
	if len(div.Attributes) != 1 || div.Attributes[0].Name != "class" || div.Attributes[0].Value != "container" {
		t.Fatalf("unexpected attributes: %+v", div.Attributes)
	}
	if len(div.Children) != 1 {
		t.Fatalf("expected 1 child in div, got %d", len(div.Children))
	}
}

func TestParserComponentWithNamedSlot(t *testing.T) {
	input := `<ui:Card title="My Card">
  <p>Default Content</p>
  <slot:footer>
    <button>Submit</button>
  </slot:footer>
</ui:Card>`

	lex := parser.NewLexer("card.rhui", input)
	p := parser.NewParser(lex)

	root, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if len(root.Children) != 1 {
		t.Fatalf("expected 1 root child, got %d", len(root.Children))
	}

	comp, ok := root.Children[0].(*ast.ComponentNode)
	if !ok {
		t.Fatalf("expected ComponentNode, got %T", root.Children[0])
	}
	if comp.Name != "ui:Card" {
		t.Fatalf("expected component 'ui:Card', got %q", comp.Name)
	}

	// Verify default slot
	if len(comp.DefaultSlot) == 0 {
		t.Fatalf("expected default slot to have content")
	}

	// Verify named slot "footer"
	footerSlot, exists := comp.NamedSlots["footer"]
	if !exists {
		t.Fatalf("expected named slot 'footer' to exist")
	}
	if len(footerSlot) == 0 {
		t.Fatalf("expected named slot 'footer' to have content")
	}
}

func TestParserConditionalAndEach(t *testing.T) {
	input := `<if test="len(Users) > 0">
  <each items="Users" as="u">
    <span>{{ u.Name }}</span>
  </each>
</if>
<else>
  <p>No Users</p>
</else>`

	lex := parser.NewLexer("cond.rhui", input)
	p := parser.NewParser(lex)

	root, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if len(root.Children) != 1 {
		t.Fatalf("expected 1 root child, got %d", len(root.Children))
	}

	ifNode, ok := root.Children[0].(*ast.IfNode)
	if !ok {
		t.Fatalf("expected IfNode, got %T", root.Children[0])
	}

	if ifNode.Condition != "len(Users) > 0" {
		t.Fatalf("expected condition 'len(Users) > 0', got %q", ifNode.Condition)
	}

	if len(ifNode.Then) == 0 {
		t.Fatalf("expected then branch to have children")
	}

	if len(ifNode.Else) == 0 {
		t.Fatalf("expected else branch to have children")
	}
}
