package runtime

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

// Evaluator evaluates safe allowlisted expressions.
type Evaluator struct {
	customFuncs map[string]any
}

// NewEvaluator creates a new Evaluator.
func NewEvaluator() *Evaluator {
	return &Evaluator{
		customFuncs: make(map[string]any),
	}
}

// RegisterFunc registers a custom function to the evaluator.
func (e *Evaluator) RegisterFunc(name string, fn any) {
	e.customFuncs[name] = fn
}

// Eval evaluates an expression string within a given Scope.
func (e *Evaluator) Eval(expr string, scope *Scope) (any, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return nil, nil
	}
	p := &exprParser{
		input:       expr,
		pos:         0,
		scope:       scope,
		customFuncs: e.customFuncs,
	}
	return p.parseExpression()
}

// EvalBool evaluates an expression and returns its truthiness.
func (e *Evaluator) EvalBool(expr string, scope *Scope) (bool, error) {
	res, err := e.Eval(expr, scope)
	if err != nil {
		return false, err
	}
	return IsTruthy(res), nil
}

// IsTruthy determines whether an arbitrary value evaluates to true in conditional contexts.
func IsTruthy(val any) bool {
	if val == nil {
		return false
	}
	switch v := val.(type) {
	case bool:
		return v
	case string:
		return v != ""
	case int:
		return v != 0
	case int8:
		return v != 0
	case int16:
		return v != 0
	case int32:
		return v != 0
	case int64:
		return v != 0
	case uint:
		return v != 0
	case uint8:
		return v != 0
	case uint16:
		return v != 0
	case uint32:
		return v != 0
	case uint64:
		return v != 0
	case float32:
		return v != 0
	case float64:
		return v != 0
	}

	rv := reflect.ValueOf(val)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map, reflect.String:
		return rv.Len() > 0
	case reflect.Ptr, reflect.Interface:
		return !rv.IsNil()
	default:
		return true
	}
}

type exprParser struct {
	input       string
	pos         int
	scope       *Scope
	customFuncs map[string]any
}

func (p *exprParser) parseExpression() (any, error) {
	return p.parseOr()
}

func (p *exprParser) parseOr() (any, error) {
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
			left = IsTruthy(left) || IsTruthy(right)
		} else {
			break
		}
	}
	return left, nil
}

func (p *exprParser) parseAnd() (any, error) {
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
			left = IsTruthy(left) && IsTruthy(right)
		} else {
			break
		}
	}
	return left, nil
}

func (p *exprParser) parseComparison() (any, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}

	p.skipWhitespace()
	if p.match("==") {
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return compareValues(left, right) == 0, nil
	}
	if p.match("!=") {
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return compareValues(left, right) != 0, nil
	}
	if p.match("<=") {
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return compareValues(left, right) <= 0, nil
	}
	if p.match(">=") {
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return compareValues(left, right) >= 0, nil
	}
	if p.match("<") {
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return compareValues(left, right) < 0, nil
	}
	if p.match(">") {
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return compareValues(left, right) > 0, nil
	}

	return left, nil
}

func (p *exprParser) parseUnary() (any, error) {
	p.skipWhitespace()
	if p.match("!") {
		val, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return !IsTruthy(val), nil
	}
	return p.parsePrimary()
}

func (p *exprParser) parsePrimary() (any, error) {
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

	// String literal
	if ch == '"' || ch == '\'' {
		return p.parseStringLiteral(ch)
	}

	// Number literal
	if unicode.IsDigit(rune(ch)) {
		return p.parseNumber()
	}

	// Identifier or function call or field access
	if unicode.IsLetter(rune(ch)) || ch == '_' {
		ident := p.parseIdentifier()

		p.skipWhitespace()
		// Function call: ident(...)
		if p.pos < len(p.input) && p.input[p.pos] == '(' {
			return p.parseFunctionCall(ident)
		}

		// Field access: ident.Sub.Field
		val := p.resolveIdentifier(ident)
		for p.pos < len(p.input) && p.input[p.pos] == '.' {
			p.pos++ // consume '.'
			field := p.parseIdentifier()
			val = resolveField(val, field)
		}
		return val, nil
	}

	return nil, fmt.Errorf("unexpected character in expression: %c", ch)
}

func (p *exprParser) parseStringLiteral(quote byte) (string, error) {
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

func (p *exprParser) parseNumber() (any, error) {
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

func (p *exprParser) parseIdentifier() string {
	start := p.pos
	for p.pos < len(p.input) {
		c := rune(p.input[p.pos])
		if unicode.IsLetter(c) || unicode.IsDigit(c) || c == '_' {
			p.pos++
		} else {
			break
		}
	}
	return p.input[start:p.pos]
}

func (p *exprParser) parseFunctionCall(funcName string) (any, error) {
	p.pos++ // consume '('
	var args []any

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

	// Check built-in functions
	switch funcName {
	case "len":
		if len(args) == 0 || args[0] == nil {
			return 0, nil
		}
		rv := reflect.ValueOf(args[0])
		switch rv.Kind() {
		case reflect.Slice, reflect.Array, reflect.Map, reflect.String:
			return rv.Len(), nil
		default:
			return 0, nil
		}
	}

	// Check custom functions
	if fn, ok := p.customFuncs[funcName]; ok {
		rv := reflect.ValueOf(fn)
		if rv.Kind() == reflect.Func {
			var in []reflect.Value
			for _, a := range args {
				if a == nil {
					in = append(in, reflect.Zero(reflect.TypeOf((*any)(nil)).Elem()))
				} else {
					in = append(in, reflect.ValueOf(a))
				}
			}
			out := rv.Call(in)
			if len(out) > 0 {
				return out[0].Interface(), nil
			}
			return nil, nil
		}
	}

	return nil, fmt.Errorf("undefined function %q in expression", funcName)
}

func (p *exprParser) resolveIdentifier(ident string) any {
	switch ident {
	case "true":
		return true
	case "false":
		return false
	case "nil", "null":
		return nil
	}

	if val, ok := p.scope.Get(ident); ok {
		return val
	}
	return nil
}

func resolveField(obj any, field string) any {
	if obj == nil {
		return nil
	}

	// If map
	if m, ok := obj.(map[string]any); ok {
		return m[field]
	}

	// Reflection for struct
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() == reflect.Struct {
		fieldVal := val.FieldByName(field)
		if fieldVal.IsValid() && fieldVal.CanInterface() {
			return fieldVal.Interface()
		}
	}
	return nil
}

func compareValues(a, b any) int {
	if a == nil && b == nil {
		return 0
	}
	if a == nil {
		return -1
	}
	if b == nil {
		return 1
	}

	// Convert ints / floats to float64
	numA, isNumA := toFloat(a)
	numB, isNumB := toFloat(b)
	if isNumA && isNumB {
		if numA < numB {
			return -1
		}
		if numA > numB {
			return 1
		}
		return 0
	}

	strA := fmt.Sprint(a)
	strB := fmt.Sprint(b)
	return strings.Compare(strA, strB)
}

func toFloat(val any) (float64, bool) {
	switch v := val.(type) {
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	case float32:
		return float64(v), true
	case float64:
		return v, true
	default:
		return 0, false
	}
}

func (p *exprParser) skipWhitespace() {
	for p.pos < len(p.input) && unicode.IsSpace(rune(p.input[p.pos])) {
		p.pos++
	}
}

func (p *exprParser) match(op string) bool {
	if strings.HasPrefix(p.input[p.pos:], op) {
		p.pos += len(op)
		return true
	}
	return false
}
