package runtime

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/suserc0de/rhui/ast"
	"github.com/suserc0de/rhui/compiler"
)

// Renderer renders AST nodes directly to an io.Writer.
type Renderer struct {
	registry  *compiler.Registry
	evaluator *Evaluator
}

// NewRenderer creates a new Renderer instance.
func NewRenderer(registry *compiler.Registry, evaluator *Evaluator) *Renderer {
	if evaluator == nil {
		evaluator = NewEvaluator()
	}
	return &Renderer{
		registry:  registry,
		evaluator: evaluator,
	}
}

// RenderNode renders any AST Node to w using the given Scope.
func (r *Renderer) RenderNode(w io.Writer, node ast.Node, scope *Scope) error {
	switch n := node.(type) {
	case *ast.RootNode:
		for _, child := range n.Children {
			if err := r.RenderNode(w, child, scope); err != nil {
				return err
			}
		}
		return nil

	case *ast.TextNode:
		_, err := io.WriteString(w, n.Content)
		return err

	case *ast.CommentNode:
		// Optional: write comment or skip
		_, err := io.WriteString(w, n.Content)
		return err

	case *ast.InterpolationNode:
		val, err := r.evaluator.Eval(n.Expression, scope)
		if err != nil {
			return err
		}
		if val != nil {
			strVal := fmt.Sprint(val)
			escaped := EscapeHTML(strVal)
			_, err = io.WriteString(w, escaped)
			return err
		}
		return nil

	case *ast.ElementNode:
		return r.renderElement(w, n, scope)

	case *ast.ComponentNode:
		return r.renderComponent(w, n, scope)

	case *ast.SlotNode:
		// When slot placeholder is encountered directly in component rendering
		return r.renderSlot(w, n, scope)

	case *ast.IfNode:
		cond, err := r.evaluator.EvalBool(n.Condition, scope)
		if err != nil {
			return err
		}
		if cond {
			for _, child := range n.Then {
				if err := r.RenderNode(w, child, scope); err != nil {
					return err
				}
			}
		} else {
			for _, child := range n.Else {
				if err := r.RenderNode(w, child, scope); err != nil {
					return err
				}
			}
		}
		return nil

	case *ast.EachNode:
		return r.renderEach(w, n, scope)

	default:
		return nil
	}
}

func (r *Renderer) renderElement(w io.Writer, elem *ast.ElementNode, scope *Scope) error {
	if _, err := fmt.Fprintf(w, "<%s", elem.TagName); err != nil {
		return err
	}

	// Render attributes
	for _, attr := range elem.Attributes {
		if attr.IsColon {
			// Dynamic expression binding: e.g. :disabled="false" or :attrs="myAttrs"
			val, err := r.evaluator.Eval(attr.Value, scope)
			if err != nil {
				return err
			}
			attrName := strings.TrimPrefix(attr.Name, ":")

			// Handle :attrs dynamic attribute spreading
			if attrName == "attrs" {
				if m, ok := val.(map[string]any); ok {
					for k, v := range m {
						if b, isBool := v.(bool); isBool {
							if b {
								if _, err := fmt.Fprintf(w, " %s", EscapeHTML(k)); err != nil {
									return err
								}
							}
						} else if v != nil {
							vStr := fmt.Sprint(v)
							if k == "href" || k == "src" {
								vStr = SanitizeURL(vStr)
							} else {
								vStr = EscapeHTML(vStr)
							}
							if _, err := fmt.Fprintf(w, " %s=\"%s\"", EscapeHTML(k), vStr); err != nil {
								return err
							}
						}
					}
				}
				continue
			}

			// Boolean attribute handling:
			// If false or nil, omit the attribute entirely!
			if b, ok := val.(bool); ok {
				if b {
					if _, err := fmt.Fprintf(w, " %s", attrName); err != nil {
						return err
					}
				}
			} else if val != nil {
				valStr := fmt.Sprint(val)
				if attrName == "href" || attrName == "src" {
					valStr = SanitizeURL(valStr)
				} else {
					valStr = EscapeHTML(valStr)
				}
				if _, err := fmt.Fprintf(w, " %s=\"%s\"", attrName, valStr); err != nil {
					return err
				}
			}
		} else {
			// Literal attribute
			if attr.Value == "true" {
				if _, err := fmt.Fprintf(w, " %s", attr.Name); err != nil {
					return err
				}
			} else {
				valStr := attr.Value
				if attr.Name == "href" || attr.Name == "src" {
					valStr = SanitizeURL(valStr)
				} else {
					valStr = EscapeHTML(valStr)
				}
				if _, err := fmt.Fprintf(w, " %s=\"%s\"", attr.Name, valStr); err != nil {
					return err
				}
			}
		}
	}

	if elem.IsSelfClose {
		_, err := io.WriteString(w, " />")
		return err
	}

	if _, err := io.WriteString(w, ">"); err != nil {
		return err
	}

	for _, child := range elem.Children {
		if err := r.RenderNode(w, child, scope); err != nil {
			return err
		}
	}

	_, err := fmt.Fprintf(w, "</%s>", elem.TagName)
	return err
}

func (r *Renderer) renderComponent(w io.Writer, comp *ast.ComponentNode, scope *Scope) error {
	compName := comp.Name
	// Strip "ui:" prefix if any for matching registry name
	lookupName := strings.TrimPrefix(compName, "ui:")

	var meta *compiler.ComponentMeta
	if r.registry != nil {
		if m, ok := r.registry.Get(compName); ok {
			meta = m
		} else if m, ok := r.registry.Get(lookupName); ok {
			meta = m
		}
	}

	if meta == nil {
		return fmt.Errorf("component %q is not registered", comp.Name)
	}

	// Create an isolated subscope for this component invocation
	compScope := NewScope(nil)

	// Populate props with default values
	for _, propDef := range meta.Props {
		compScope.Set(propDef.Name, propDef.DefaultValue)
	}

	// Evaluate passed attributes/props and collect undeclared ones into extra_attrs
	extraAttrs := make(map[string]any)
	for _, attr := range comp.Attributes {
		propName := attr.Name
		isDynamic := attr.IsColon
		if isDynamic {
			propName = strings.TrimPrefix(attr.Name, ":")
		}

		var propVal any
		if isDynamic {
			val, err := r.evaluator.Eval(attr.Value, scope)
			if err != nil {
				return err
			}
			propVal = val
		} else {
			propVal = attr.Value
		}

		// Check if declared in component props
		if _, isDeclared := meta.Props[propName]; isDeclared {
			compScope.Set(propName, propVal)
		} else {
			extraAttrs[propName] = propVal
		}
	}
	compScope.Set("extra_attrs", extraAttrs)

	// Attach slots into component scope
	compScope.Set("__default_slot__", comp.DefaultSlot)
	compScope.Set("__parent_scope__", scope)
	for slotName, nodes := range comp.NamedSlots {
		compScope.Set("__slot_"+slotName+"__", nodes)
	}

	// Render the component's internal children/template
	for _, child := range meta.Children {
		if err := r.RenderNode(w, child, compScope); err != nil {
			return err
		}
	}

	return nil
}

func (r *Renderer) renderSlot(w io.Writer, slot *ast.SlotNode, compScope *Scope) error {
	parentScopeRaw, _ := compScope.Get("__parent_scope__")
	parentScope, _ := parentScopeRaw.(*Scope)
	if parentScope == nil {
		parentScope = compScope
	}

	var slotNodes []ast.Node
	if slot.Name == "" {
		// Default slot
		if nodesRaw, ok := compScope.Get("__default_slot__"); ok {
			slotNodes, _ = nodesRaw.([]ast.Node)
		}
	} else {
		// Named slot
		if nodesRaw, ok := compScope.Get("__slot_" + slot.Name + "__"); ok {
			slotNodes, _ = nodesRaw.([]ast.Node)
		}
	}

	// If no caller slot provided, render fallback children
	if len(slotNodes) == 0 && len(slot.Fallback) > 0 {
		for _, node := range slot.Fallback {
			if err := r.RenderNode(w, node, compScope); err != nil {
				return err
			}
		}
		return nil
	}

	// Render slot contents with the parent/caller's scope
	for _, node := range slotNodes {
		if err := r.RenderNode(w, node, parentScope); err != nil {
			return err
		}
	}

	return nil
}

func (r *Renderer) renderEach(w io.Writer, each *ast.EachNode, scope *Scope) error {
	itemsVal, err := r.evaluator.Eval(each.Items, scope)
	if err != nil {
		return err
	}
	if itemsVal == nil {
		return nil
	}

	rv := reflect.ValueOf(itemsVal)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return nil
	}

	totalLen := rv.Len()
	for i := 0; i < totalLen; i++ {
		item := rv.Index(i).Interface()
		loopScope := scope.SubScope()
		if each.As != "" {
			loopScope.Set(each.As, item)
		}
		if each.Index != "" {
			loopScope.Set(each.Index, i)
		}

		// Inject loop metadata variables
		loopScope.Set("@index", i)
		loopScope.Set("@count", i+1)
		loopScope.Set("@first", i == 0)
		loopScope.Set("@last", i == totalLen-1)
		loopScope.Set("@even", (i+1)%2 == 0)
		loopScope.Set("@odd", (i+1)%2 != 0)

		for _, child := range each.Body {
			if err := r.RenderNode(w, child, loopScope); err != nil {
				return err
			}
		}
	}

	return nil
}
