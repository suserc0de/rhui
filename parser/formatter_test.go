package parser_test

import (
	"strings"
	"testing"

	"github.com/suserc0de/rhui/parser"
)

func TestFormatterBasic(t *testing.T) {
	input := `<component name="Card"><props><prop name="title" default="Hello" /></props><div class="card"><h3>{{ title }}</h3><slot /></div></component>`
	formatter := parser.NewFormatter()

	out, err := formatter.Format("card.rhui", input)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	if !strings.Contains(out, `<component name="Card">`) {
		t.Errorf("expected component tag, got:\n%s", out)
	}
	if !strings.Contains(out, `  <props>`) {
		t.Errorf("expected indented props, got:\n%s", out)
	}
	if !strings.Contains(out, `    <prop name="title" default="Hello" />`) {
		t.Errorf("expected indented prop, got:\n%s", out)
	}
	if !strings.Contains(out, `  <div class="card">`) {
		t.Errorf("expected indented div, got:\n%s", out)
	}
	if !strings.Contains(out, `    <slot />`) {
		t.Errorf("expected slot, got:\n%s", out)
	}

	// Idempotency: formatting formatted output produces the exact same output
	out2, err := formatter.Format("card.rhui", out)
	if err != nil {
		t.Fatalf("second Format failed: %v", err)
	}
	if out != out2 {
		t.Errorf("formatter not idempotent!\nFirst:\n%s\nSecond:\n%s", out, out2)
	}
}

func TestFormatterConditionalsAndLoops(t *testing.T) {
	input := `<div class="container"><if test="isActive"><p>Active</p></if><else><p>Inactive</p></else><each items="Items" as="item"><span>{{ item.Name }}</span></each></div>`
	formatter := parser.NewFormatter()

	out, err := formatter.Format("test.rhui", input)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	if !strings.Contains(out, `  <if test="isActive">`) {
		t.Errorf("missing if tag in:\n%s", out)
	}
	if !strings.Contains(out, `  <else>`) {
		t.Errorf("missing else tag in:\n%s", out)
	}
	if !strings.Contains(out, `  <each items="Items" as="item">`) {
		t.Errorf("missing each tag in:\n%s", out)
	}

	// Verify idempotency
	out2, err := formatter.Format("test.rhui", out)
	if err != nil {
		t.Fatalf("second format failed: %v", err)
	}
	if out != out2 {
		t.Errorf("formatter not idempotent!\nOut1:\n%s\nOut2:\n%s", out, out2)
	}
}
