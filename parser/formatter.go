package parser

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/suserc0de/rhui/ast"
)

// Formatter formats a RHUI AST back into clean, canonical .rhui template source.
type Formatter struct {
	indentStr string
}

// NewFormatter creates a Formatter with standard 2-space indentation.
func NewFormatter() *Formatter {
	return &Formatter{
		indentStr: "  ",
	}
}

// Format formats the given template content and returns the formatted string.
func (f *Formatter) Format(name, content string) (string, error) {
	p := NewParser(NewLexer(name, content))
	root, err := p.Parse()
	if err != nil {
		return "", err
	}
	if len(p.Errors()) > 0 {
		return "", p.Errors()[0]
	}

	var buf bytes.Buffer
	f.formatNodes(&buf, root.Children, 0)
	res := strings.TrimRight(buf.String(), " \t\r\n") + "\n"
	return res, nil
}

func (f *Formatter) formatNodes(buf *bytes.Buffer, nodes []ast.Node, level int) {
	for i, node := range nodes {
		f.formatNode(buf, node, level)
		// Add newline between block-level components/elements if needed
		_ = i
	}
}

func (f *Formatter) indent(level int) string {
	return strings.Repeat(f.indentStr, level)
}

func (f *Formatter) formatNode(buf *bytes.Buffer, node ast.Node, level int) {
	if node == nil {
		return
	}

	switch n := node.(type) {
	case *ast.TextNode:
		// Preserve non-empty text
		trimmed := strings.TrimSpace(n.Content)
		if trimmed == "" {
			return
		}
		buf.WriteString(f.indent(level))
		buf.WriteString(trimmed)
		buf.WriteString("\n")

	case *ast.CommentNode:
		buf.WriteString(f.indent(level))
		buf.WriteString("<!--")
		buf.WriteString(n.Content)
		buf.WriteString("-->\n")

	case *ast.InterpolationNode:
		buf.WriteString(f.indent(level))
		buf.WriteString("{{ ")
		buf.WriteString(strings.TrimSpace(n.Expression))
		buf.WriteString(" }}\n")

	case *ast.ComponentDefNode:
		buf.WriteString(f.indent(level))
		buf.WriteString(fmt.Sprintf(`<component name="%s">`, n.Name))
		buf.WriteString("\n")

		if len(n.Props) > 0 {
			buf.WriteString(f.indent(level + 1))
			buf.WriteString("<props>\n")
			for _, prop := range n.Props {
				buf.WriteString(f.indent(level + 2))
				if prop.DefaultValue != "" {
					buf.WriteString(fmt.Sprintf(`<prop name="%s" default="%s" />`, prop.Name, prop.DefaultValue))
				} else {
					buf.WriteString(fmt.Sprintf(`<prop name="%s" />`, prop.Name))
				}
				buf.WriteString("\n")
			}
			buf.WriteString(f.indent(level + 1))
			buf.WriteString("</props>\n")
		}

		f.formatNodes(buf, n.Children, level+1)
		buf.WriteString(f.indent(level))
		buf.WriteString("</component>\n")

	case *ast.ElementNode:
		f.formatElement(buf, n, level)

	case *ast.ComponentNode:
		f.formatComponent(buf, n, level)

	case *ast.SlotNode:
		buf.WriteString(f.indent(level))
		if n.Name == "" {
			if len(n.Fallback) == 0 {
				buf.WriteString("<slot />\n")
			} else {
				buf.WriteString("<slot>\n")
				f.formatNodes(buf, n.Fallback, level+1)
				buf.WriteString(f.indent(level))
				buf.WriteString("</slot>\n")
			}
		} else {
			if len(n.Fallback) == 0 {
				buf.WriteString(fmt.Sprintf("<slot:%s />\n", n.Name))
			} else {
				buf.WriteString(fmt.Sprintf("<slot:%s>\n", n.Name))
				f.formatNodes(buf, n.Fallback, level+1)
				buf.WriteString(f.indent(level))
				buf.WriteString(fmt.Sprintf("</slot:%s>\n", n.Name))
			}
		}

	case *ast.IfNode:
		buf.WriteString(f.indent(level))
		buf.WriteString(fmt.Sprintf(`<if test="%s">`, n.Condition))
		buf.WriteString("\n")
		f.formatNodes(buf, n.Then, level+1)
		if len(n.Else) > 0 {
			buf.WriteString(f.indent(level))
			buf.WriteString("<else>\n")
			f.formatNodes(buf, n.Else, level+1)
			buf.WriteString(f.indent(level))
			buf.WriteString("</else>\n")
		}
		buf.WriteString(f.indent(level))
		buf.WriteString("</if>\n")

	case *ast.EachNode:
		buf.WriteString(f.indent(level))
		if n.Index != "" {
			buf.WriteString(fmt.Sprintf(`<each items="%s" as="%s" index="%s">`, n.Items, n.As, n.Index))
		} else {
			buf.WriteString(fmt.Sprintf(`<each items="%s" as="%s">`, n.Items, n.As))
		}
		buf.WriteString("\n")
		f.formatNodes(buf, n.Body, level+1)
		buf.WriteString(f.indent(level))
		buf.WriteString("</each>\n")
	}
}

func (f *Formatter) formatElement(buf *bytes.Buffer, n *ast.ElementNode, level int) {
	buf.WriteString(f.indent(level))
	buf.WriteString("<")
	buf.WriteString(n.TagName)
	f.formatAttributes(buf, n.Attributes)

	if n.IsSelfClose && len(n.Children) == 0 {
		buf.WriteString(" />\n")
		return
	}

	buf.WriteString(">\n")
	f.formatNodes(buf, n.Children, level+1)
	buf.WriteString(f.indent(level))
	buf.WriteString("</")
	buf.WriteString(n.TagName)
	buf.WriteString(">\n")
}

func (f *Formatter) formatComponent(buf *bytes.Buffer, n *ast.ComponentNode, level int) {
	buf.WriteString(f.indent(level))
	buf.WriteString("<")
	buf.WriteString(n.Name)
	f.formatAttributes(buf, n.Attributes)

	hasChildren := len(n.DefaultSlot) > 0 || len(n.NamedSlots) > 0
	if n.IsSelfClose && !hasChildren {
		buf.WriteString(" />\n")
		return
	}

	buf.WriteString(">\n")
	if len(n.DefaultSlot) > 0 {
		f.formatNodes(buf, n.DefaultSlot, level+1)
	}
	for name, slotNodes := range n.NamedSlots {
		buf.WriteString(f.indent(level + 1))
		buf.WriteString(fmt.Sprintf("<slot:%s>\n", name))
		f.formatNodes(buf, slotNodes, level+2)
		buf.WriteString(f.indent(level + 1))
		buf.WriteString(fmt.Sprintf("</slot:%s>\n", name))
	}
	buf.WriteString(f.indent(level))
	buf.WriteString("</")
	buf.WriteString(n.Name)
	buf.WriteString(">\n")
}

func (f *Formatter) formatAttributes(buf *bytes.Buffer, attrs []ast.Attribute) {
	for _, attr := range attrs {
		buf.WriteString(" ")
		if attr.IsColon {
			buf.WriteString(":")
			buf.WriteString(attr.Name)
			buf.WriteString(fmt.Sprintf(`="%s"`, attr.Value))
		} else {
			buf.WriteString(attr.Name)
			if attr.Value != "" {
				buf.WriteString(fmt.Sprintf(`="%s"`, attr.Value))
			}
		}
	}
}
