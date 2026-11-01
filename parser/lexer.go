package parser

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/coderhaqi/rhui/ast"
)

// Lexer scans a .rhui source string and produces a stream of Tokens.
type Lexer struct {
	filename string
	source   string
	start    int // start of current token
	pos      int // current read position
	width    int // width of last read rune

	line   int // current line (1-based)
	column int // current column (1-based)

	tokenLine   int // line of token start
	tokenColumn int // column of token start

	inTag bool // whether the scanner is currently inside an open tag (<...>)
}

// NewLexer creates and initializes a new Lexer for the given filename and source.
func NewLexer(filename, source string) *Lexer {
	return &Lexer{
		filename:    filename,
		source:      source,
		line:        1,
		column:      1,
		tokenLine:   1,
		tokenColumn: 1,
	}
}

// NextToken scans and returns the next Token.
func (l *Lexer) NextToken() Token {
	for {
		if l.isEOF() {
			return l.emit(TokenEOF, "")
		}

		if !l.inTag {
			// Scan content outside tags: comments, interpolation, tag open, or text.
			if l.hasPrefix("<!--") {
				return l.lexComment()
			}
			if l.hasPrefix("{{") {
				return l.lexInterpolation()
			}
			if l.peek() == '<' {
				return l.lexTagOpen()
			}
			return l.lexText()
		}

		// Inside an open tag (<...>): skip whitespace, scan attributes, symbols, or tag close
		l.skipWhitespace()
		if l.isEOF() {
			return l.emit(TokenEOF, "")
		}

		ch := l.peek()

		// Tag closing
		if ch == '>' {
			l.next()
			l.inTag = false
			return l.emit(TokenTagClose, ">")
		}
		if l.hasPrefix("/>") {
			l.next()
			l.next()
			l.inTag = false
			return l.emit(TokenTagSelfClose, "/>")
		}

		// Equal sign
		if ch == '=' {
			l.next()
			return l.emit(TokenEqual, "=")
		}

		// String literals for attribute values
		if ch == '"' || ch == '\'' {
			return l.lexStringLiteral(ch)
		}

		// Colon-prefixed dynamic attributes (e.g. :rows, :disabled)
		if ch == ':' {
			return l.lexColonAttribute()
		}

		// Attribute identifiers or tag names
		if isIdentStart(ch) {
			return l.lexIdent()
		}

		// Any unexpected character inside tag
		r := l.next()
		return l.emit(TokenError, string(r))
	}
}

func (l *Lexer) lexComment() Token {
	l.markStart()
	// Skip <!--
	l.pos += 4
	l.column += 4

	for !l.isEOF() {
		if l.hasPrefix("-->") {
			l.pos += 3
			l.column += 3
			val := l.source[l.start:l.pos]
			return l.emit(TokenComment, val)
		}
		r := l.next()
		if r == '\n' {
			l.line++
			l.column = 1
		}
	}
	return l.emit(TokenError, "unclosed HTML comment")
}

func (l *Lexer) lexInterpolation() Token {
	l.markStart()
	// Skip {{
	l.pos += 2
	l.column += 2

	for !l.isEOF() {
		if l.hasPrefix("}}") {
			l.pos += 2
			l.column += 2
			// Content inside {{ ... }}
			content := l.source[l.start+2 : l.pos-2]
			return l.emit(TokenInterpolation, strings.TrimSpace(content))
		}
		r := l.next()
		if r == '\n' {
			l.line++
			l.column = 1
		}
	}
	return l.emit(TokenError, "unclosed interpolation '}}'")
}

func (l *Lexer) lexTagOpen() Token {
	l.markStart()

	// Check for closing tag </
	if l.hasPrefix("</") {
		l.pos += 2
		l.column += 2
		l.inTag = true
		return l.emit(TokenTagOpenSlash, "</")
	}

	// Normal opening tag <
	l.next()
	l.inTag = true
	return l.emit(TokenTagOpen, "<")
}

func (l *Lexer) lexText() Token {
	l.markStart()

	for !l.isEOF() {
		if l.peek() == '<' || l.hasPrefix("{{") {
			break
		}
		r := l.next()
		if r == '\n' {
			l.line++
			l.column = 1
		}
	}

	val := l.source[l.start:l.pos]
	return l.emit(TokenText, val)
}

func (l *Lexer) lexStringLiteral(quote rune) Token {
	l.markStart()
	l.next() // consume opening quote

	var sb strings.Builder
	for !l.isEOF() {
		r := l.next()
		if r == quote {
			return l.emit(TokenStringLiteral, sb.String())
		}
		if r == '\\' {
			if !l.isEOF() {
				escaped := l.next()
				sb.WriteRune(escaped)
				continue
			}
		}
		if r == '\n' {
			l.line++
			l.column = 1
		}
		sb.WriteRune(r)
	}
	return l.emit(TokenError, "unclosed string literal")
}

func (l *Lexer) lexColonAttribute() Token {
	l.markStart()
	l.next() // consume ':'

	for !l.isEOF() {
		ch := l.peek()
		if isIdentPart(ch) {
			l.next()
		} else {
			break
		}
	}

	val := l.source[l.start:l.pos]
	return l.emit(TokenColonAttr, val)
}

func (l *Lexer) lexIdent() Token {
	l.markStart()

	for !l.isEOF() {
		ch := l.peek()
		if isIdentPart(ch) {
			l.next()
		} else {
			break
		}
	}

	val := l.source[l.start:l.pos]

	// Check for special tags / directives
	switch val {
	case "if":
		return l.emit(TokenIfOpen, val)
	case "else":
		return l.emit(TokenElse, val)
	case "each":
		return l.emit(TokenEachOpen, val)
	case "component":
		return l.emit(TokenComponentOpen, val)
	case "slot":
		return l.emit(TokenSlotOpen, val)
	default:
		return l.emit(TokenIdent, val)
	}
}

func (l *Lexer) markStart() {
	l.start = l.pos
	l.tokenLine = l.line
	l.tokenColumn = l.column
}

func (l *Lexer) emit(t TokenType, val string) Token {
	tok := Token{
		Type:  t,
		Value: val,
		Pos: ast.Position{
			File:   l.filename,
			Line:   l.tokenLine,
			Column: l.tokenColumn,
		},
	}
	return tok
}

func (l *Lexer) next() rune {
	if l.pos >= len(l.source) {
		l.width = 0
		return 0
	}
	r, w := utf8.DecodeRuneInString(l.source[l.pos:])
	l.width = w
	l.pos += w
	l.column++
	return r
}

func (l *Lexer) peek() rune {
	if l.pos >= len(l.source) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(l.source[l.pos:])
	return r
}

func (l *Lexer) hasPrefix(prefix string) bool {
	return strings.HasPrefix(l.source[l.pos:], prefix)
}

func (l *Lexer) isEOF() bool {
	return l.pos >= len(l.source)
}

func (l *Lexer) skipWhitespace() {
	for !l.isEOF() {
		r := l.peek()
		if unicode.IsSpace(r) {
			l.next()
			if r == '\n' {
				l.line++
				l.column = 1
			}
		} else {
			break
		}
	}
}

func isIdentStart(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_' || ch == '@'
}

func isIdentPart(ch rune) bool {
	return unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' || ch == '-' || ch == ':' || ch == '.' || ch == '@'
}
