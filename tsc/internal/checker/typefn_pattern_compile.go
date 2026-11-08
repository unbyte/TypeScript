package checker

import (
	"github.com/microsoft/TypeScript/tsc/internal/ast"
)

type typefnPattern struct {
	node     *ast.Node
	expr     *typefnExpression
	children []*typefnPattern
	slot     int
	key      string
}

func (b *typefnCompiler) match(node *ast.Node) {
	d := node.AsTypeFunctionMatch()
	slot := b.p.slots
	b.p.slots++
	b.emit(typefnInstruction{op: tfStore, node: node, expr: b.typeExpression(d.Expression), slot: slot})
	var ends []int
	for index, armNode := range d.Arms.Nodes {
		arm := armNode.AsTypeFunctionMatchArm()
		if arm.Pattern.Kind == ast.KindTypeFunctionWildcardPattern {
			if index != len(d.Arms.Nodes)-1 || arm.Guard != nil {
				b.error(arm.Pattern, "the wildcard arm must be last and unguarded")
			}
		} else if index == len(d.Arms.Nodes)-1 {
			b.error(arm.Pattern, "match requires a final unguarded wildcard arm")
		}
		// Compile all pin/Type expressions in the outer environment, before publishing
		// any captures. Native name resolution follows this same scope boundary.
		pattern := b.pattern(arm.Pattern)
		b.scopes = append(b.scopes, map[string]typefnBinding{})
		b.patternBindings(pattern)
		attempt := b.emit(typefnInstruction{op: tfMatch, node: armNode, slot: slot, pattern: pattern})
		guard := -1
		if arm.Guard != nil {
			guard = b.emit(typefnInstruction{op: tfBranch, node: arm.Guard, expr: b.typeExpression(arm.Guard)})
		}
		b.block(arm.Body, true)
		ends = append(ends, b.emit(typefnInstruction{op: tfJump, node: armNode}))
		next := len(b.p.code)
		b.p.code[attempt].target = next
		if guard >= 0 {
			b.p.code[guard].target = next
		}
		b.scopes = b.scopes[:len(b.scopes)-1]
	}
	for _, jump := range ends {
		b.p.code[jump].target = len(b.p.code)
	}
}

func (b *typefnCompiler) pattern(node *ast.Node) *typefnPattern {
	p := &typefnPattern{node: node, slot: -1}
	d := node.AsTypeFunctionPattern()
	switch node.Kind {
	case ast.KindTypeFunctionLiteralPattern, ast.KindTypeFunctionIsPattern, ast.KindTypeFunctionNativePattern, ast.KindTypeFunctionPinPattern:
		p.expr = b.typeExpression(d.Expression)
	case ast.KindTypeFunctionFieldPattern:
		p.key = node.Name().Text()
		p.children = append(p.children, b.pattern(d.Expression))
	case ast.KindTypeFunctionRestPattern:
		p.children = append(p.children, b.pattern(d.Expression))
	case ast.KindTypeFunctionTemplatePattern:
		if d.Expression.Kind == ast.KindTypeFunctionTemplate {
			for _, span := range d.Expression.AsTypeFunctionTemplate().Spans.Nodes {
				p.children = append(p.children, b.pattern(span.AsTypeFunctionTemplateSpan().Expression))
			}
		}
	case ast.KindTypeFunctionTuplePattern, ast.KindTypeFunctionObjectPattern:
		keys := map[string]bool{}
		for _, element := range d.Elements.Nodes {
			child := b.pattern(element)
			if node.Kind == ast.KindTypeFunctionObjectPattern {
				if keys[child.key] {
					b.error(element.Name(), "duplicate object pattern key "+child.key)
				}
				keys[child.key] = true
			}
			p.children = append(p.children, child)
		}
	}
	return p
}

func (b *typefnCompiler) patternBindings(p *typefnPattern) {
	if p.node.Kind == ast.KindTypeFunctionCapturePattern {
		p.slot = b.bind(p.node.Name(), false)
	}
	for _, child := range p.children {
		b.patternBindings(child)
	}
}
