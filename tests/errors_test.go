package tests

import (
	"strings"
	"testing"

	"github.com/suserc0de/rhui"
	"github.com/suserc0de/rhui/ast"
)

func TestPositionString(t *testing.T) {
	pos := ast.Position{File: "components/Card.rhui", Line: 10, Column: 4}
	expected := "components/Card.rhui:10:4"
	if pos.String() != expected {
		t.Fatalf("expected %s, got %s", expected, pos.String())
	}

	noFile := ast.Position{Line: 5, Column: 2}
	if noFile.String() != "5:2" {
		t.Fatalf("expected 5:2, got %s", noFile.String())
	}
}

func TestTemplateErrorSnippet(t *testing.T) {
	source := `<Layout>
  <ui:Card :title="invalid
</Layout>`

	pos := ast.Position{File: "index.rhui", Line: 2, Column: 12}
	err := rhui.NewTemplateErrorWithSnippet(pos, "unterminated expression string", source)

	errStr := err.Error()
	if !strings.Contains(errStr, "[index.rhui:2:12] unterminated expression string") {
		t.Errorf("error string missing header position: %s", errStr)
	}

	if !strings.Contains(errStr, "<ui:Card :title=\"invalid") {
		t.Errorf("error snippet missing target line: %s", errStr)
	}

	if !strings.Contains(errStr, "^") {
		t.Errorf("error snippet missing caret pointer: %s", errStr)
	}
}
