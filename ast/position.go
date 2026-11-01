package ast

import "fmt"

// Position tracks the location of a node or token in a source file.
type Position struct {
	File   string
	Line   int
	Column int
}

// String returns a human-readable representation of Position (e.g., "page.rhui:12:5").
func (p Position) String() string {
	if p.File == "" {
		return fmt.Sprintf("%d:%d", p.Line, p.Column)
	}
	return fmt.Sprintf("%s:%d:%d", p.File, p.Line, p.Column)
}

// IsValid checks if position contains non-zero line and column.
func (p Position) IsValid() bool {
	return p.Line > 0 && p.Column > 0
}
