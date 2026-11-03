package expr

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// LiteralExpr represents static constants: int, float, string, bool, nil.
type LiteralExpr struct {
	Value any
}

func (e *LiteralExpr) Eval(ctx Context) (any, error) {
	return e.Value, nil
}

// IdentExpr represents variable lookups, e.g. "user", "Title", "@index".
type IdentExpr struct {
	Name string
}

func (e *IdentExpr) Eval(ctx Context) (any, error) {
	switch e.Name {
	case "true":
		return true, nil
	case "false":
		return false, nil
	case "nil", "null":
		return nil, nil
	}
	if val, ok := ctx.Get(e.Name); ok {
		return val, nil
	}
	return nil, nil
}

// FieldAccessExpr represents accessing properties/keys, e.g. "user.Profile.Name".
type FieldAccessExpr struct {
	Target Expr
	Fields []string
}

func (e *FieldAccessExpr) Eval(ctx Context) (any, error) {
	val, err := e.Target.Eval(ctx)
	if err != nil || val == nil {
		return nil, err
	}
	for _, f := range e.Fields {
		val = resolveField(val, f)
		if val == nil {
			return nil, nil
		}
	}
	return val, nil
}

// UnaryExpr represents unary operators like !val.
type UnaryExpr struct {
	Op   string // "!"
	Expr Expr
}

func (e *UnaryExpr) Eval(ctx Context) (any, error) {
	val, err := e.Expr.Eval(ctx)
	if err != nil {
		return nil, err
	}
	if e.Op == "!" {
		return !IsTruthy(val), nil
	}
	return val, nil
}

// BinaryExpr represents binary operations: ==, !=, <, <=, >, >=, &&, ||.
type BinaryExpr struct {
	Op    string
	Left  Expr
	Right Expr
}

func (e *BinaryExpr) Eval(ctx Context) (any, error) {
	switch e.Op {
	case "||":
		lVal, err := e.Left.Eval(ctx)
		if err != nil {
			return nil, err
		}
		if IsTruthy(lVal) {
			return true, nil
		}
		rVal, err := e.Right.Eval(ctx)
		if err != nil {
			return nil, err
		}
		return IsTruthy(rVal), nil

	case "&&":
		lVal, err := e.Left.Eval(ctx)
		if err != nil {
			return nil, err
		}
		if !IsTruthy(lVal) {
			return false, nil
		}
		rVal, err := e.Right.Eval(ctx)
		if err != nil {
			return nil, err
		}
		return IsTruthy(rVal), nil

	case "==":
		lVal, err := e.Left.Eval(ctx)
		if err != nil {
			return nil, err
		}
		rVal, err := e.Right.Eval(ctx)
		if err != nil {
			return nil, err
		}
		return compareValues(lVal, rVal) == 0, nil

	case "!=":
		lVal, err := e.Left.Eval(ctx)
		if err != nil {
			return nil, err
		}
		rVal, err := e.Right.Eval(ctx)
		if err != nil {
			return nil, err
		}
		return compareValues(lVal, rVal) != 0, nil

	case "<":
		lVal, err := e.Left.Eval(ctx)
		if err != nil {
			return nil, err
		}
		rVal, err := e.Right.Eval(ctx)
		if err != nil {
			return nil, err
		}
		return compareValues(lVal, rVal) < 0, nil

	case "<=":
		lVal, err := e.Left.Eval(ctx)
		if err != nil {
			return nil, err
		}
		rVal, err := e.Right.Eval(ctx)
		if err != nil {
			return nil, err
		}
		return compareValues(lVal, rVal) <= 0, nil

	case ">":
		lVal, err := e.Left.Eval(ctx)
		if err != nil {
			return nil, err
		}
		rVal, err := e.Right.Eval(ctx)
		if err != nil {
			return nil, err
		}
		return compareValues(lVal, rVal) > 0, nil

	case ">=":
		lVal, err := e.Left.Eval(ctx)
		if err != nil {
			return nil, err
		}
		rVal, err := e.Right.Eval(ctx)
		if err != nil {
			return nil, err
		}
		return compareValues(lVal, rVal) >= 0, nil
	}
	return nil, fmt.Errorf("unsupported binary operator: %s", e.Op)
}

// CallExpr represents function calls, e.g. len(items), has_slot("header"), customFunc(arg1).
type CallExpr struct {
	Name string
	Args []Expr
}

func (e *CallExpr) Eval(ctx Context) (any, error) {
	// Evaluate arguments
	args := make([]any, len(e.Args))
	for i, argExpr := range e.Args {
		val, err := argExpr.Eval(ctx)
		if err != nil {
			return nil, err
		}
		args[i] = val
	}

	// Built-in functions
	switch e.Name {
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
	case "has_slot":
		if len(args) == 0 || args[0] == nil {
			return false, nil
		}
		slotName := fmt.Sprint(args[0])
		var slotKey string
		if slotName == "" || slotName == "default" {
			slotKey = "__default_slot__"
		} else {
			slotKey = "__slot_" + slotName + "__"
		}
		if val, ok := ctx.Get(slotKey); ok && val != nil {
			if rv := reflect.ValueOf(val); rv.Kind() == reflect.Slice {
				return rv.Len() > 0, nil
			}
			return true, nil
		}
		return false, nil
	}

	// Custom registered functions
	if fn, ok := ctx.GetFunc(e.Name); ok {
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

	return nil, fmt.Errorf("undefined function %q in expression", e.Name)
}

// FilterInvocation represents a pipe filter call: | upper or | dateFormat("2006-01-02").
type FilterInvocation struct {
	Name string
	Args []Expr // additional arguments after the piped target
}

// PipeExpr represents chained filter operations: Target | Filter1 | Filter2.
type PipeExpr struct {
	Target  Expr
	Filters []FilterInvocation
}

func (e *PipeExpr) Eval(ctx Context) (any, error) {
	val, err := e.Target.Eval(ctx)
	if err != nil {
		return nil, err
	}

	for _, filter := range e.Filters {
		// Evaluate arguments with target as first argument
		evaledArgs := make([]any, 1+len(filter.Args))
		evaledArgs[0] = val
		for i, argExpr := range filter.Args {
			argVal, err := argExpr.Eval(ctx)
			if err != nil {
				return nil, err
			}
			evaledArgs[i+1] = argVal
		}

		filtered, err := applyFilter(ctx, filter.Name, evaledArgs)
		if err != nil {
			return nil, err
		}
		val = filtered
	}
	return val, nil
}

func applyFilter(ctx Context, name string, args []any) (any, error) {
	if len(args) == 0 {
		return nil, nil
	}
	target := args[0]

	// Built-in filters
	switch name {
	case "upper":
		return strings.ToUpper(fmt.Sprint(target)), nil
	case "lower":
		return strings.ToLower(fmt.Sprint(target)), nil
	case "trim":
		return strings.TrimSpace(fmt.Sprint(target)), nil
	case "rupiah":
		val, _ := toFloat(target)
		return fmt.Sprintf("Rp %s", formatRupiahNumber(int64(val))), nil
	case "default":
		if target == nil || fmt.Sprint(target) == "" {
			if len(args) > 1 {
				return args[1], nil
			}
			return "", nil
		}
		return target, nil
	case "dateFormat":
		if len(args) > 1 {
			layout := fmt.Sprint(args[1])
			if t, ok := target.(interface{ Format(string) string }); ok {
				return t.Format(layout), nil
			}
		}
		return target, nil
	case "json":
		bytes, err := json.Marshal(target)
		if err != nil {
			return "", err
		}
		return string(bytes), nil
	}

	// Custom registered filters
	if fn, ok := ctx.GetFilter(name); ok {
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

	return nil, fmt.Errorf("unknown filter %q", name)
}

func formatRupiahNumber(n int64) string {
	in := strconv.FormatInt(n, 10)
	var out []byte
	l := len(in)
	for i := 0; i < l; i++ {
		if i > 0 && (l-i)%3 == 0 {
			out = append(out, '.')
		}
		out = append(out, in[i])
	}
	return string(out)
}

func resolveField(obj any, field string) any {
	if obj == nil {
		return nil
	}

	// Fast path: map[string]any
	if m, ok := obj.(map[string]any); ok {
		return m[field]
	}

	// Reflection for struct or struct pointer
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
