package parser

import (
	"fmt"
	"strings"

	"github.com/coderhaqi/rhui/ast"
)

// Parser parses a stream of tokens into an AST.
type Parser struct {
	lexer     *Lexer
	curToken  Token
	peekToken Token
	errors    []error
}

// NewParser creates a new Parser instance.
func NewParser(lexer *Lexer) *Parser {
	p := &Parser{
		lexer: lexer,
	}
	// Read two tokens so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.lexer.NextToken()
}

func (p *Parser) Errors() []error {
	return p.errors
}

func (p *Parser) addError(pos ast.Position, msg string) {
	p.errors = append(p.errors, fmt.Errorf("[%s] %s", pos.String(), msg))
}

// Parse parses the whole template into an *ast.RootNode.
func (p *Parser) Parse() (*ast.RootNode, error) {
	root := &ast.RootNode{
		Pos: p.curToken.Pos,
	}

	for p.curToken.Type != TokenEOF {
		node := p.parseNode()
		if node != nil {
			root.Children = append(root.Children, node)
		}
	}

	if len(p.errors) > 0 {
		return nil, p.errors[0]
	}

	return root, nil
}

func (p *Parser) parseNode() ast.Node {
	switch p.curToken.Type {
	case TokenText:
		text := &ast.TextNode{
			Content: p.curToken.Value,
			Pos:     p.curToken.Pos,
		}
		p.nextToken()
		return text

	case TokenComment:
		comment := &ast.CommentNode{
			Content: p.curToken.Value,
			Pos:     p.curToken.Pos,
		}
		p.nextToken()
		return comment

	case TokenInterpolation:
		interp := &ast.InterpolationNode{
			Expression: p.curToken.Value,
			Pos:        p.curToken.Pos,
		}
		p.nextToken()
		return interp

	case TokenTagOpen:
		return p.parseTagOrComponent()

	case TokenTagOpenSlash:
		// Unexpected closing tag at root level
		p.addError(p.curToken.Pos, fmt.Sprintf("unexpected closing tag '</%s>'", p.peekToken.Value))
		p.nextToken()
		return nil

	default:
		p.nextToken()
		return nil
	}
}

func (p *Parser) parseTagOrComponent() ast.Node {
	startPos := p.curToken.Pos
	p.nextToken() // consume '<'

	tagName := p.curToken.Value
	tokenType := p.curToken.Type

	switch tokenType {
	case TokenIfOpen:
		return p.parseIf(startPos)
	case TokenEachOpen:
		return p.parseEach(startPos)
	case TokenComponentOpen:
		return p.parseComponentDef(startPos)
	case TokenSlotOpen:
		return p.parseSlot(startPos, "slot")
	}

	// Check if this tag is slot:something
	if strings.HasPrefix(tagName, "slot:") {
		return p.parseSlot(startPos, tagName)
	}

	// Normal HTML tag or custom Component (<ui:...> or PascalCase)
	p.nextToken() // consume tagName

	attrs, isSelfClose := p.parseAttributes()

	// Check if this is a Component
	if isComponentTag(tagName) {
		comp := &ast.ComponentNode{
			Name:        tagName,
			Attributes:  attrs,
			NamedSlots:  make(map[string][]ast.Node),
			IsSelfClose: isSelfClose,
			Pos:         startPos,
		}

		if isSelfClose {
			return comp
		}

		// Parse children of component (could be regular children for default slot or <slot:name>)
		for p.curToken.Type != TokenEOF && !p.isClosingTag(tagName) {
			child := p.parseNode()
			if child != nil {
				// Check if child is a named slot block
				if slotNode, ok := child.(*namedSlotContainer); ok {
					comp.NamedSlots[slotNode.name] = append(comp.NamedSlots[slotNode.name], slotNode.children...)
				} else {
					comp.DefaultSlot = append(comp.DefaultSlot, child)
				}
			}
		}

		p.expectClosingTag(tagName)
		return comp
	}

	// Regular HTML Element
	elem := &ast.ElementNode{
		TagName:     tagName,
		Attributes:  attrs,
		IsSelfClose: isSelfClose,
		Pos:         startPos,
	}

	if isSelfClose || isVoidHTMLTag(tagName) {
		return elem
	}

	for p.curToken.Type != TokenEOF && !p.isClosingTag(tagName) {
		child := p.parseNode()
		if child != nil {
			elem.Children = append(elem.Children, child)
		}
	}

	p.expectClosingTag(tagName)
	return elem
}

func (p *Parser) parseAttributes() ([]ast.Attribute, bool) {
	var attrs []ast.Attribute

	for p.curToken.Type != TokenTagClose && p.curToken.Type != TokenTagSelfClose && p.curToken.Type != TokenEOF {
		switch p.curToken.Type {
		case TokenColonAttr:
			attrName := p.curToken.Value
			pos := p.curToken.Pos
			p.nextToken()

			val := ""
			if p.curToken.Type == TokenEqual {
				p.nextToken()
				if p.curToken.Type == TokenStringLiteral {
					val = p.curToken.Value
					p.nextToken()
				}
			}
			attrs = append(attrs, ast.Attribute{
				Name:    attrName,
				Value:   val,
				IsColon: true,
				Pos:     pos,
			})

		case TokenIdent, TokenAttrIdent:
			attrName := p.curToken.Value
			pos := p.curToken.Pos
			p.nextToken()

			val := ""
			if p.curToken.Type == TokenEqual {
				p.nextToken()
				if p.curToken.Type == TokenStringLiteral {
					val = p.curToken.Value
					p.nextToken()
				}
			} else {
				// Boolean attribute like <input disabled />
				val = "true"
			}
			attrs = append(attrs, ast.Attribute{
				Name:    attrName,
				Value:   val,
				IsColon: false,
				Pos:     pos,
			})

		default:
			p.nextToken()
		}
	}

	isSelfClose := (p.curToken.Type == TokenTagSelfClose)
	p.nextToken() // consume '>' or '/>'
	return attrs, isSelfClose
}

func (p *Parser) parseIf(pos ast.Position) *ast.IfNode {
	p.nextToken() // consume 'if'
	attrs, _ := p.parseAttributes()

	condition := ""
	for _, attr := range attrs {
		if attr.Name == "test" {
			condition = attr.Value
			break
		}
	}

	ifNode := &ast.IfNode{
		Condition: condition,
		Pos:       pos,
	}

	for p.curToken.Type != TokenEOF {
		if p.isClosingTag("if") {
			p.expectClosingTag("if")

			// Check if immediately followed by optional whitespace and <else>
			if p.curToken.Type == TokenText && strings.TrimSpace(p.curToken.Value) == "" {
				p.nextToken() // skip whitespace text between </if> and <else>
			}

			if p.curToken.Type == TokenTagOpen && (p.peekToken.Type == TokenElse || p.peekToken.Value == "else") {
				p.nextToken() // consume '<'
				p.nextToken() // consume 'else'
				_, isElseSelfClose := p.parseAttributes()

				if !isElseSelfClose {
					for p.curToken.Type != TokenEOF && !p.isClosingTag("else") {
						child := p.parseNode()
						if child != nil {
							ifNode.Else = append(ifNode.Else, child)
						}
					}
					p.expectClosingTag("else")
				}
			}
			break
		}

		if p.curToken.Type == TokenTagOpen && (p.peekToken.Type == TokenElse || p.peekToken.Value == "else") {
			// Found <else> inside <if>
			p.nextToken() // consume '<'
			p.nextToken() // consume 'else'
			_, isElseSelfClose := p.parseAttributes()

			if !isElseSelfClose {
				for p.curToken.Type != TokenEOF && !p.isClosingTag("else") {
					child := p.parseNode()
					if child != nil {
						ifNode.Else = append(ifNode.Else, child)
					}
				}
				p.expectClosingTag("else")
			}
			if p.isClosingTag("if") {
				p.expectClosingTag("if")
			}
			break
		}

		child := p.parseNode()
		if child != nil {
			ifNode.Then = append(ifNode.Then, child)
		}
	}

	return ifNode
}

func (p *Parser) parseEach(pos ast.Position) *ast.EachNode {
	p.nextToken() // consume 'each'
	attrs, _ := p.parseAttributes()

	eachNode := &ast.EachNode{
		Pos: pos,
	}

	for _, attr := range attrs {
		switch attr.Name {
		case "items":
			eachNode.Items = attr.Value
		case "as":
			eachNode.As = attr.Value
		case "index":
			eachNode.Index = attr.Value
		}
	}

	for p.curToken.Type != TokenEOF && !p.isClosingTag("each") {
		child := p.parseNode()
		if child != nil {
			eachNode.Body = append(eachNode.Body, child)
		}
	}

	p.expectClosingTag("each")
	return eachNode
}

func (p *Parser) parseComponentDef(pos ast.Position) *ast.ComponentDefNode {
	p.nextToken() // consume 'component'
	attrs, _ := p.parseAttributes()

	compDef := &ast.ComponentDefNode{
		Pos: pos,
	}
	for _, attr := range attrs {
		if attr.Name == "name" {
			compDef.Name = attr.Value
		}
	}

	for p.curToken.Type != TokenEOF && !p.isClosingTag("component") {
		// Look for <props> block
		if p.curToken.Type == TokenTagOpen && p.peekToken.Value == "props" {
			p.parsePropsBlock(compDef)
			continue
		}

		child := p.parseNode()
		if child != nil {
			compDef.Children = append(compDef.Children, child)
		}
	}

	p.expectClosingTag("component")
	return compDef
}

func (p *Parser) parsePropsBlock(compDef *ast.ComponentDefNode) {
	p.nextToken() // consume '<'
	p.nextToken() // consume 'props'
	_, _ = p.parseAttributes()

	for p.curToken.Type != TokenEOF && !p.isClosingTag("props") {
		if p.curToken.Type == TokenTagOpen && p.peekToken.Value == "prop" {
			propPos := p.curToken.Pos
			p.nextToken() // consume '<'
			p.nextToken() // consume 'prop'
			attrs, _ := p.parseAttributes()

			pDef := ast.PropDef{Pos: propPos, Type: "string"}
			for _, a := range attrs {
				switch a.Name {
				case "name":
					pDef.Name = a.Value
				case "type":
					pDef.Type = a.Value
				case "default":
					pDef.DefaultValue = a.Value
				}
			}
			compDef.Props = append(compDef.Props, pDef)
			continue
		}
		p.nextToken()
	}

	p.expectClosingTag("props")
}

type namedSlotContainer struct {
	name     string
	children []ast.Node
	pos      ast.Position
}

func (n *namedSlotContainer) Position() ast.Position { return n.pos }

func (p *Parser) parseSlot(pos ast.Position, rawName string) ast.Node {
	p.nextToken() // consume 'slot' or 'slot:name'
	_, isSelfClose := p.parseAttributes()

	slotName := ""
	if strings.HasPrefix(rawName, "slot:") {
		slotName = strings.TrimPrefix(rawName, "slot:")
	}

	if isSelfClose {
		return &ast.SlotNode{
			Name: slotName,
			Pos:  pos,
		}
	}

	// If it's a container like <slot:actions> ... </slot:actions>
	container := &namedSlotContainer{
		name: slotName,
		pos:  pos,
	}

	for p.curToken.Type != TokenEOF && !p.isClosingTag(rawName) {
		child := p.parseNode()
		if child != nil {
			container.children = append(container.children, child)
		}
	}

	p.expectClosingTag(rawName)
	return container
}

func (p *Parser) isClosingTag(name string) bool {
	return p.curToken.Type == TokenTagOpenSlash && p.peekToken.Value == name
}

func (p *Parser) expectClosingTag(name string) {
	if p.isClosingTag(name) {
		p.nextToken() // consume '</'
		p.nextToken() // consume name
		if p.curToken.Type == TokenTagClose {
			p.nextToken() // consume '>'
		}
	} else {
		p.addError(p.curToken.Pos, fmt.Sprintf("expected closing tag '</%s>', got %q", name, p.curToken.Value))
	}
}

func isComponentTag(name string) bool {
	if strings.HasPrefix(name, "ui:") {
		return true
	}
	// PascalCase check (e.g. Layout, Card, DataTable)
	if len(name) > 0 {
		first := name[0]
		return first >= 'A' && first <= 'Z'
	}
	return false
}

func isVoidHTMLTag(name string) bool {
	switch strings.ToLower(name) {
	case "area", "base", "br", "col", "embed", "hr", "img", "input", "link", "meta", "param", "source", "track", "wbr":
		return true
	default:
		return false
	}
}
