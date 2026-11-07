package expr

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

var (
	exprCache = newLRUCache(DefaultLRUCapacity)
)

// SetCacheCapacity reconfigures the LRU cache capacity for parsed expressions.
func SetCacheCapacity(capacity int) {
	exprCache = newLRUCache(capacity)
}

// ClearCache flushes the expression parse cache.
func ClearCache() {
	exprCache.clear()
}

// Parse compiles an expression string into an AST Expr.
// Results are thread-safe and cached in a bounded LRU cache for instant subsequent lookups.
func Parse(input string) (Expr, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return nil, nil
	}

	if cached, ok := exprCache.get(trimmed); ok {
		return cached, nil
	}

	p := &parser{
		input: trimmed,
		pos:   0,
	}
	node, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	exprCache.put(trimmed, node)
	return node, nil
}

type parser struct {
	input string
	pos   int
}

func (p *parser) parseExpression() (Expr, error) {
	target, err := p.parseOr()
	if err != nil {
		return nil, err
	}

	// Handle pipe filter chaining: expr | filter1 | filter2("arg")
	var filters []FilterInvocation
	for {
		p.skipWhitespace()
		if p.pos < len(p.input) && p.input[p.pos] == '|' {
			if p.pos+1 < len(p.input) && p.input[p.pos+1] == '|' {
				break
			}
			p.pos++ // consume '|'
			p.skipWhitespace()
			filterName := p.parseIdentifier()

			var filterArgs []Expr
			p.skipWhitespace()
			if p.pos < len(p.input) && p.input[p.pos] == '(' {
				p.pos++ // consume '('
				p.skipWhitespace()
				if p.pos < len(p.input) && p.input[p.pos] == ')' {
					p.pos++
				} else {
					for {
						arg, err := p.parseExpression()
						if err != nil {
							return nil, err
						}
						filterArgs = append(filterArgs, arg)
						p.skipWhitespace()
						if p.pos < len(p.input) && p.input[p.pos] == ',' {
							p.pos++
							continue
						}
						if p.pos < len(p.input) && p.input[p.pos] == ')' {
							p.pos++
							break
						}
						break
					}
				}
			}

			filters = append(filters, FilterInvocation{
				Name: filterName,
				Args: filterArgs,
			})
		} else {
			break
		}
	}

	if len(filters) > 0 {
		return &PipeExpr{
			Target:  target,
			Filters: filters,
		}, nil
	}

	return target, nil
}

func (p *parser) parseOr() (Expr, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}

	for {
		p.skipWhitespace()
		if p.match("||") {
			right, err := p.parseAnd()
			if err != nil {
				return nil, err
			}
			if right == nil {
				return nil, fmt.Errorf("missing right-hand operand for ||")
			}
			left = &BinaryExpr{
				Op:    "||",
				Left:  left,
				Right: right,
			}
		} else {
			break
		}
	}
	return left, nil
}

func (p *parser) parseAnd() (Expr, error) {
	left, err := p.parseComparison()
	if err != nil {
		return nil, err
	}

	for {
		p.skipWhitespace()
		if p.match("&&") {
			right, err := p.parseComparison()
			if err != nil {
				return nil, err
			}
			if right == nil {
				return nil, fmt.Errorf("missing right-hand operand for &&")
			}
			left = &BinaryExpr{
				Op:    "&&",
				Left:  left,
				Right: right,
			}
		} else {
			break
		}
	}
	return left, nil
}

func (p *parser) parseComparison() (Expr, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}

	p.skipWhitespace()
	for _, op := range []string{"==", "!=", "<=", ">=", "<", ">"} {
		if p.match(op) {
			right, err := p.parseUnary()
			if err != nil {
				return nil, err
			}
			if right == nil {
				return nil, fmt.Errorf("missing right-hand operand for operator %s", op)
			}
			return &BinaryExpr{
				Op:    op,
				Left:  left,
				Right: right,
			}, nil
		}
	}

	return left, nil
}

func (p *parser) parseUnary() (Expr, error) {
	p.skipWhitespace()
	if p.match("!") {
		inner, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &UnaryExpr{
			Op:   "!",
			Expr: inner,
		}, nil
	}
	return p.parsePrimary()
}

func (p *parser) parsePrimary() (Expr, error) {
	p.skipWhitespace()
	if p.pos >= len(p.input) {
		return nil, nil
	}

	ch := p.input[p.pos]

	// Parentheses
	if ch == '(' {
		p.pos++
		val, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		p.skipWhitespace()
		if p.pos < len(p.input) && p.input[p.pos] == ')' {
			p.pos++
		}
		return val, nil
	}

	// Map literal: { 'key': expr, ... }
	if ch == '{' {
		return p.parseMapLiteral()
	}

	// Slice/Array literal: [ 'val1', expr, ... ]
	if ch == '[' {
		return p.parseSliceLiteral()
	}

	// String literal
	if ch == '"' || ch == '\'' {
		str, err := p.parseStringLiteral(ch)
		if err != nil {
			return nil, err
		}
		return &LiteralExpr{Value: str}, nil
	}

	// Number literal
	if unicode.IsDigit(rune(ch)) {
		num, err := p.parseNumber()
		if err != nil {
			return nil, err
		}
		return &LiteralExpr{Value: num}, nil
	}

	// Identifier or function call or field access or loop variable (@index, @first, etc.)
	if unicode.IsLetter(rune(ch)) || ch == '_' || ch == '@' {
		ident := p.parseIdentifier()

		p.skipWhitespace()
		// Function call: ident(...)
		if p.pos < len(p.input) && p.input[p.pos] == '(' {
			return p.parseFunctionCall(ident)
		}

		// Field access: ident.Sub.Field
		var fields []string
		for p.pos < len(p.input) && p.input[p.pos] == '.' {
			p.pos++ // consume '.'
			field := p.parseIdentifier()
			if field != "" {
				fields = append(fields, field)
			}
		}

		identNode := &IdentExpr{Name: ident}
		if len(fields) > 0 {
			return &FieldAccessExpr{
				Target: identNode,
				Fields: fields,
			}, nil
		}
		return identNode, nil
	}

	return nil, fmt.Errorf("unexpected character in expression: %c", ch)
}

func (p *parser) parseStringLiteral(quote byte) (string, error) {
	p.pos++ // consume quote
	start := p.pos
	for p.pos < len(p.input) {
		if p.input[p.pos] == quote {
			res := p.input[start:p.pos]
			p.pos++ // consume closing quote
			return res, nil
		}
		p.pos++
	}
	return "", fmt.Errorf("unclosed string literal in expression")
}

func (p *parser) parseNumber() (any, error) {
	start := p.pos
	isFloat := false
	for p.pos < len(p.input) {
		c := p.input[p.pos]
		if unicode.IsDigit(rune(c)) {
			p.pos++
		} else if c == '.' && !isFloat {
			isFloat = true
			p.pos++
		} else {
			break
		}
	}
	numStr := p.input[start:p.pos]
	if isFloat {
		return strconv.ParseFloat(numStr, 64)
	}
	return strconv.ParseInt(numStr, 10, 64)
}

func (p *parser) parseIdentifier() string {
	start := p.pos
	for p.pos < len(p.input) {
		c := rune(p.input[p.pos])
		if unicode.IsLetter(c) || unicode.IsDigit(c) || c == '_' || c == '@' {
			p.pos++
		} else {
			break
		}
	}
	return p.input[start:p.pos]
}

func (p *parser) parseFunctionCall(funcName string) (Expr, error) {
	p.pos++ // consume '('
	var args []Expr

	p.skipWhitespace()
	if p.pos < len(p.input) && p.input[p.pos] == ')' {
		p.pos++ // consume ')'
	} else {
		for {
			arg, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			args = append(args, arg)

			p.skipWhitespace()
			if p.pos < len(p.input) && p.input[p.pos] == ',' {
				p.pos++
				continue
			}
			if p.pos < len(p.input) && p.input[p.pos] == ')' {
				p.pos++
				break
			}
			break
		}
	}

	return &CallExpr{
		Name: funcName,
		Args: args,
	}, nil
}

func (p *parser) parseMapLiteral() (Expr, error) {
	p.pos++ // consume '{'
	var entries []MapEntry

	for {
		p.skipWhitespace()
		if p.pos >= len(p.input) {
			return nil, fmt.Errorf("unclosed map literal '{'")
		}
		if p.input[p.pos] == '}' {
			p.pos++ // consume '}'
			break
		}

		// Key: can be string literal or identifier
		key, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}

		p.skipWhitespace()
		if p.pos >= len(p.input) || p.input[p.pos] != ':' {
			return nil, fmt.Errorf("expected ':' after map key")
		}
		p.pos++ // consume ':'

		// Value
		val, err := p.parseExpression()
		if err != nil {
			return nil, err
		}

		entries = append(entries, MapEntry{
			Key:   key,
			Value: val,
		})

		p.skipWhitespace()
		if p.pos < len(p.input) && p.input[p.pos] == ',' {
			p.pos++
			continue
		}
		if p.pos < len(p.input) && p.input[p.pos] == '}' {
			p.pos++
			break
		}
		break
	}

	return &MapExpr{Entries: entries}, nil
}

func (p *parser) parseSliceLiteral() (Expr, error) {
	p.pos++ // consume '['
	var elements []Expr

	for {
		p.skipWhitespace()
		if p.pos >= len(p.input) {
			return nil, fmt.Errorf("unclosed slice literal '['")
		}
		if p.input[p.pos] == ']' {
			p.pos++ // consume ']'
			break
		}

		elem, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		elements = append(elements, elem)

		p.skipWhitespace()
		if p.pos < len(p.input) && p.input[p.pos] == ',' {
			p.pos++
			continue
		}
		if p.pos < len(p.input) && p.input[p.pos] == ']' {
			p.pos++
			break
		}
		break
	}

	return &SliceExpr{Elements: elements}, nil
}

func (p *parser) skipWhitespace() {
	for p.pos < len(p.input) && unicode.IsSpace(rune(p.input[p.pos])) {
		p.pos++
	}
}

func (p *parser) match(op string) bool {
	if strings.HasPrefix(p.input[p.pos:], op) {
		p.pos += len(op)
		return true
	}
	return false
}
