package rhui

import (
	"fmt"
	"strings"

	"github.com/coderhaqi/rhui/ast"
)

// TemplateError represents a detailed error during tokenization, parsing, compilation, or rendering.
type TemplateError struct {
	Pos     ast.Position
	Message string
	Snippet string
}

// Error implements the standard error interface.
func (e *TemplateError) Error() string {
	var sb strings.Builder
	if e.Pos.IsValid() {
		sb.WriteString(fmt.Sprintf("[%s] ", e.Pos.String()))
	}
	sb.WriteString(e.Message)

	if e.Snippet != "" {
		sb.WriteString("\n\n" + e.Snippet)
	}
	return sb.String()
}

// NewTemplateError creates a new TemplateError with source position.
func NewTemplateError(pos ast.Position, msg string) *TemplateError {
	return &TemplateError{
		Pos:     pos,
		Message: msg,
	}
}

// NewTemplateErrorWithSnippet creates a TemplateError with a visual source snippet pointer.
func NewTemplateErrorWithSnippet(pos ast.Position, msg string, source string) *TemplateError {
	snippet := generateSnippet(pos, source)
	return &TemplateError{
		Pos:     pos,
		Message: msg,
		Snippet: snippet,
	}
}

func generateSnippet(pos ast.Position, source string) string {
	if pos.Line <= 0 || source == "" {
		return ""
	}

	lines := strings.Split(source, "\n")
	if pos.Line > len(lines) {
		return ""
	}

	targetLine := lines[pos.Line-1]
	indent := ""
	if pos.Column > 1 {
		indent = strings.Repeat(" ", pos.Column-1)
	}

	return fmt.Sprintf("  %4d | %s\n       | %s^", pos.Line, targetLine, indent)
}
