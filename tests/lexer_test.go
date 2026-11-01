package tests

import (
	"testing"

	"github.com/coderhaqi/rhui/parser"
)

func TestLexerBasicTagAndText(t *testing.T) {
	input := `<div class="card">Hello, {{ User.Name }}</div>`
	lex := parser.NewLexer("test.rhui", input)

	expected := []struct {
		expectedType parser.TokenType
		expectedVal  string
	}{
		{parser.TokenTagOpen, "<"},
		{parser.TokenIdent, "div"},
		{parser.TokenIdent, "class"},
		{parser.TokenEqual, "="},
		{parser.TokenStringLiteral, "card"},
		{parser.TokenTagClose, ">"},
		{parser.TokenText, "Hello, "},
		{parser.TokenInterpolation, "User.Name"},
		{parser.TokenTagOpenSlash, "</"},
		{parser.TokenIdent, "div"},
		{parser.TokenTagClose, ">"},
		{parser.TokenEOF, ""},
	}

	for i, exp := range expected {
		tok := lex.NextToken()
		if tok.Type != exp.expectedType {
			t.Fatalf("[%d] expected token type %v (%s), got %v (%s) with value %q",
				i, exp.expectedType, exp.expectedType, tok.Type, tok.Type, tok.Value)
		}
		if tok.Value != exp.expectedVal {
			t.Fatalf("[%d] expected token value %q, got %q", i, exp.expectedVal, tok.Value)
		}
	}
}

func TestLexerComponentAndDynamicAttributes(t *testing.T) {
	input := `<ui:Button variant="primary" :disabled="User.IsActive == false" />`
	lex := parser.NewLexer("btn.rhui", input)

	expected := []struct {
		expectedType parser.TokenType
		expectedVal  string
	}{
		{parser.TokenTagOpen, "<"},
		{parser.TokenIdent, "ui:Button"},
		{parser.TokenIdent, "variant"},
		{parser.TokenEqual, "="},
		{parser.TokenStringLiteral, "primary"},
		{parser.TokenColonAttr, ":disabled"},
		{parser.TokenEqual, "="},
		{parser.TokenStringLiteral, "User.IsActive == false"},
		{parser.TokenTagSelfClose, "/>"},
		{parser.TokenEOF, ""},
	}

	for i, exp := range expected {
		tok := lex.NextToken()
		if tok.Type != exp.expectedType {
			t.Fatalf("[%d] expected token type %v, got %v with value %q", i, exp.expectedType, tok.Type, tok.Value)
		}
		if tok.Value != exp.expectedVal {
			t.Fatalf("[%d] expected token value %q, got %q", i, exp.expectedVal, tok.Value)
		}
	}
}

func TestLexerConditionalsAndSlots(t *testing.T) {
	input := `<if test="len(Users) > 0">
  <slot:actions>
    <ui:Card />
  </slot:actions>
</if>
<else>
  <p>Kosong</p>
</else>`

	lex := parser.NewLexer("cond.rhui", input)

	var tokens []parser.Token
	for {
		tok := lex.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == parser.TokenEOF || tok.Type == parser.TokenError {
			break
		}
	}

	// Verify we parsed <if, :actions named slot, and <else
	hasIf := false
	hasSlot := false
	hasElse := false

	for _, tok := range tokens {
		if tok.Type == parser.TokenIfOpen {
			hasIf = true
		}
		if tok.Type == parser.TokenIdent && tok.Value == "slot:actions" {
			hasSlot = true
		}
		if tok.Type == parser.TokenElse {
			hasElse = true
		}
	}

	if !hasIf {
		t.Errorf("expected TokenIfOpen in token stream")
	}
	if !hasSlot {
		t.Errorf("expected slot:actions in token stream")
	}
	if !hasElse {
		t.Errorf("expected TokenElse in token stream")
	}
}
