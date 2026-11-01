package parser

import (
	"fmt"

	"github.com/coderhaqi/rhui/ast"
)

// TokenType represents the category of lexical tokens in .rhui files.
type TokenType int

const (
	TokenEOF TokenType = iota
	TokenError

	// Text & Comments
	TokenText       // HTML text / literals between tags
	TokenComment    // <!-- ... -->

	// Delimiters
	TokenTagOpen       // <
	TokenTagOpenSlash  // </
	TokenTagClose      // >
	TokenTagSelfClose  // />

	// Identifiers & Attributes
	TokenIdent         // div, ui:Button, slot:actions, h1
	TokenAttrIdent     // href, title, class, hx-post
	TokenColonAttr     // :rows, :disabled, :title (Dynamic expression binding)
	TokenEqual         // =
	TokenStringLiteral // "..." or '...'

	// Directives / Control structures
	TokenIfOpen        // <if
	TokenIfClose       // </if>
	TokenElse          // <else> or <else/>
	TokenEachOpen      // <each
	TokenEachClose     // </each>

	// Special Component & Slot tags
	TokenComponentOpen // <component
	TokenComponentClose// </component>
	TokenSlotOpen      // <slot
	TokenSlotClose     // </slot>

	// Interpolations
	TokenInterpolation // {{ Expression }}
)

func (t TokenType) String() string {
	switch t {
	case TokenEOF:
		return "EOF"
	case TokenError:
		return "Error"
	case TokenText:
		return "Text"
	case TokenComment:
		return "Comment"
	case TokenTagOpen:
		return "<"
	case TokenTagOpenSlash:
		return "</"
	case TokenTagClose:
		return ">"
	case TokenTagSelfClose:
		return "/>"
	case TokenIdent:
		return "Identifier"
	case TokenAttrIdent:
		return "AttributeIdentifier"
	case TokenColonAttr:
		return "ColonAttribute"
	case TokenEqual:
		return "="
	case TokenStringLiteral:
		return "StringLiteral"
	case TokenIfOpen:
		return "<if"
	case TokenIfClose:
		return "</if>"
	case TokenElse:
		return "<else>"
	case TokenEachOpen:
		return "<each"
	case TokenEachClose:
		return "</each>"
	case TokenComponentOpen:
		return "<component"
	case TokenComponentClose:
		return "</component>"
	case TokenSlotOpen:
		return "<slot"
	case TokenSlotClose:
		return "</slot>"
	case TokenInterpolation:
		return "Interpolation"
	default:
		return fmt.Sprintf("Token(%d)", t)
	}
}

// Token represents a single parsed token with its position.
type Token struct {
	Type  TokenType
	Value string
	Pos   ast.Position
}

func (t Token) String() string {
	return fmt.Sprintf("[%s %s %q]", t.Pos.String(), t.Type.String(), t.Value)
}
