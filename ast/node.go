package ast

// Node represents an element in the Abstract Syntax Tree.
type Node interface {
	Position() Position
}

// Attribute represents an HTML attribute or a component prop.
type Attribute struct {
	Name         string   // e.g. "class", ":disabled", "variant"
	Value        string   // e.g. "btn", "User.IsActive", "primary"
	IsColon      bool     // true if binding with ':' e.g. :disabled="expr"
	CompiledExpr any      // Pre-compiled expression AST
	Pos          Position
}

// RootNode is the top-level container of a template AST.
type RootNode struct {
	Children []Node
	Pos      Position
}

func (n *RootNode) Position() Position { return n.Pos }

// TextNode represents raw static text / HTML.
type TextNode struct {
	Content string
	Pos     Position
}

func (n *TextNode) Position() Position { return n.Pos }

// CommentNode represents an HTML comment.
type CommentNode struct {
	Content string
	Pos     Position
}

func (n *CommentNode) Position() Position { return n.Pos }

// InterpolationNode represents a dynamic expression interpolation {{ ... }}.
type InterpolationNode struct {
	Expression   string
	CompiledExpr any // Pre-compiled expression AST
	Pos          Position
}

func (n *InterpolationNode) Position() Position { return n.Pos }

// ElementNode represents a standard HTML tag (e.g. <div>, <span>, <input>).
type ElementNode struct {
	TagName     string
	Attributes  []Attribute
	Children    []Node
	IsSelfClose bool
	Pos         Position
}

func (n *ElementNode) Position() Position { return n.Pos }

// ComponentNode represents an invoked custom component (e.g. <ui:Button>, <ui:Card>, <Layout>).
type ComponentNode struct {
	Name        string              // e.g. "ui:Button", "Layout", "Card"
	Attributes  []Attribute         // Props passed to component
	DefaultSlot []Node              // Children inside default <slot />
	NamedSlots  map[string][]Node   // Children mapped to named slots like <slot:actions>
	IsSelfClose bool
	Pos         Position
}

func (n *ComponentNode) Position() Position { return n.Pos }

// PropDef represents a declared prop inside a component definition.
type PropDef struct {
	Name         string
	Type         string // "string", "bool", "int", "any"
	DefaultValue string
	Pos          Position
}

// ComponentDefNode represents a component definition block <component name="...">.
type ComponentDefNode struct {
	Name     string
	Props    []PropDef
	Children []Node
	Pos      Position
}

func (n *ComponentDefNode) Position() Position { return n.Pos }

// SlotNode represents a slot placeholder in a component (<slot /> or <slot:header />) with optional fallback children.
type SlotNode struct {
	Name     string // empty for default slot, otherwise name of slot e.g. "header"
	Fallback []Node // fallback nodes to render if caller doesn't provide content
	Pos      Position
}

func (n *SlotNode) Position() Position { return n.Pos }

// IfNode represents a conditional block <if test="..."> ... </if> <else> ... </else>.
type IfNode struct {
	Condition    string
	CompiledExpr any // Pre-compiled expression AST
	Then         []Node
	Else         []Node
	Pos          Position
}

func (n *IfNode) Position() Position { return n.Pos }

// EachNode represents an iteration block <each items="..." as="..."> ... </each>.
type EachNode struct {
	Items        string // expression for items, e.g. "Users"
	CompiledExpr any    // Pre-compiled expression AST
	As           string // item variable name, e.g. "user"
	Index        string // optional index variable name, e.g. "idx"
	Body         []Node
	Pos          Position
}

func (n *EachNode) Position() Position { return n.Pos }
