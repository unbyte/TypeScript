package printer

import (
	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/scanner"
)

// Printing DSL syntax must not route its nodes through the JavaScript emitter.
// Parenthesized operators deliberately preserve the DSL's angle-list grammar.
func (p *Printer) emitTypefn(node *ast.Node) {
	state := p.enterNode(node)
	switch node.Kind {
	case ast.KindTypeFunctionMatch:
		d := node.AsTypeFunctionMatch()
		p.writeKeyword("match")
		p.writeSpace()
		p.writePunctuation("(")
		p.emitTypefn(d.Expression)
		p.writePunctuation(") {")
		p.writeLine()
		p.increaseIndent()
		for _, arm := range d.Arms.Nodes {
			p.emitTypefn(arm)
			p.writeLine()
		}
		p.decreaseIndent()
		p.writePunctuation("}")
	case ast.KindTypeFunctionMatchArm:
		d := node.AsTypeFunctionMatchArm()
		p.emitTypefn(d.Pattern)
		if d.Guard != nil {
			p.writeSpace()
			p.writeKeyword("if")
			p.writeSpace()
			p.writePunctuation("(")
			p.emitTypefn(d.Guard)
			p.writePunctuation(")")
		}
		p.writeSpace()
		p.writePunctuation("=>")
		p.writeSpace()
		p.emitTypefn(d.Body)
	case ast.KindTypeFunctionWildcardPattern, ast.KindTypeFunctionCapturePattern,
		ast.KindTypeFunctionLiteralPattern, ast.KindTypeFunctionIsPattern,
		ast.KindTypeFunctionNativePattern, ast.KindTypeFunctionTuplePattern,
		ast.KindTypeFunctionRestPattern, ast.KindTypeFunctionObjectPattern,
		ast.KindTypeFunctionFieldPattern, ast.KindTypeFunctionTemplatePattern, ast.KindTypeFunctionPinPattern:
		d := node.AsTypeFunctionPattern()
		switch node.Kind {
		case ast.KindTypeFunctionWildcardPattern:
			p.writePunctuation("_")
		case ast.KindTypeFunctionCapturePattern:
			p.emitBindingIdentifier(node.Name().AsIdentifier())
		case ast.KindTypeFunctionTuplePattern, ast.KindTypeFunctionObjectPattern:
			open, close := "[", "]"
			if node.Kind == ast.KindTypeFunctionObjectPattern {
				open, close = "{", "}"
			}
			p.writePunctuation(open)
			for i, item := range d.Elements.Nodes {
				if i > 0 {
					p.writePunctuation(",")
					p.writeSpace()
				}
				p.emitTypefn(item)
			}
			p.writePunctuation(close)
		default:
			switch node.Kind {
			case ast.KindTypeFunctionIsPattern:
				p.writeKeyword("is")
				p.writeSpace()
			case ast.KindTypeFunctionRestPattern:
				p.writePunctuation("...")
			case ast.KindTypeFunctionPinPattern:
				p.writePunctuation("=")
			case ast.KindTypeFunctionFieldPattern:
				if ast.IsIdentifier(node.Name()) {
					p.emitBindingIdentifier(node.Name().AsIdentifier())
				} else {
					p.emitExpression(node.Name(), ast.OperatorPrecedenceLowest)
				}
				p.writePunctuation(":")
				p.writeSpace()
			}
			// Unary negative literal patterns are not parenthesized DSL expressions.
			if node.Kind == ast.KindTypeFunctionLiteralPattern && d.Expression.Kind == ast.KindTypeFunctionUnary {
				p.writeOperator("-")
				p.emitTypefn(d.Expression.AsTypeFunctionExpression().Operands.Nodes[0])
			} else {
				p.emitTypefn(d.Expression)
			}
		}
	case ast.KindTypeFunctionDeclaration, ast.KindTypeFunctionLambda:
		if ast.IsTypeFunctionDeclaration(node) {
			p.emitModifierList(node, node.Modifiers(), false)
		}
		p.writeKeyword("typefn")
		if ast.IsTypeFunctionDeclaration(node) {
			p.writeSpace()
			p.emitBindingIdentifier(node.Name().AsIdentifier())
		}
		p.writePunctuation("<")
		for i, parameter := range node.TypeParameters() {
			if i > 0 {
				p.writePunctuation(",")
				p.writeSpace()
			}
			p.emitTypefn(parameter)
		}
		p.writePunctuation(">")
		if kind := ast.TypeFunctionResultKind(node); kind != nil {
			p.writePunctuation(":")
			p.writeSpace()
			p.emitTypefn(kind)
		}
		p.writeSpace()
		p.emitTypefn(ast.TypeFunctionBody(node))
	case ast.KindTypeFunctionParameter:
		parameter := node.AsTypeFunctionParameter()
		p.emitBindingIdentifier(node.Name().AsIdentifier())
		if parameter.Slots != nil {
			p.writePunctuation("<")
			for i := range parameter.Slots.Nodes {
				if i > 0 {
					p.writePunctuation(",")
					p.writeSpace()
				}
				p.writePunctuation("_")
			}
			p.writePunctuation(">")
		}
		if parameter.Constraint != nil {
			p.writeSpace()
			p.writeKeyword("extends")
			p.writeSpace()
			p.emitTypefn(parameter.Constraint)
		}
	case ast.KindTypeFunctionKind:
		d := node.AsTypeFunctionKind()
		if d.KindToken.Kind == ast.KindIdentifier {
			p.writeKeyword("typefn")
			p.writePunctuation("<")
			if d.Slots != nil {
				for i := range d.Slots.Nodes {
					if i > 0 {
						p.writePunctuation(",")
						p.writeSpace()
					}
					p.writePunctuation("_")
				}
			}
			p.writePunctuation(">")
		} else {
			p.writeKeyword("type")
		}
	case ast.KindTypeFunctionTemplate:
		d := node.AsTypeFunctionTemplate()
		p.emitTemplateHead(d.Head.AsTemplateHead())
		for _, span := range d.Spans.Nodes {
			p.emitTypefn(span)
		}
	case ast.KindTypeFunctionTemplateSpan:
		d := node.AsTypeFunctionTemplateSpan()
		p.emitTypefn(d.Expression)
		p.emitTemplateMiddleTail(d.Literal)
	case ast.KindTypeFunctionBlock:
		p.writePunctuation("{")
		p.writeLine()
		p.increaseIndent()
		for _, statement := range node.AsTypeFunctionBlock().Statements.Nodes {
			p.emitTypefn(statement)
			p.writeLine()
		}
		p.decreaseIndent()
		p.writePunctuation("}")
	case ast.KindTypeFunctionBinding, ast.KindTypeFunctionAssignment, ast.KindTypeFunctionReturn, ast.KindTypeFunctionExpressionStatement, ast.KindTypeFunctionBreak, ast.KindTypeFunctionContinue:
		s := node.AsTypeFunctionStatement()
		switch node.Kind {
		case ast.KindTypeFunctionBinding:
			p.writeKeyword("type")
			p.writeSpace()
			fallthrough
		case ast.KindTypeFunctionAssignment:
			p.emitBindingIdentifier(node.Name().AsIdentifier())
			p.writeSpace()
			p.writeOperator("=")
			p.writeSpace()
		case ast.KindTypeFunctionReturn:
			p.writeKeyword("return")
			p.writeSpace()
		case ast.KindTypeFunctionBreak:
			p.writeKeyword("break")
		case ast.KindTypeFunctionContinue:
			p.writeKeyword("continue")
		}
		if s.Expression != nil {
			p.emitTypefn(s.Expression)
		}
		p.writePunctuation(";") // Mandatory even with OmitTrailingSemicolon.
	case ast.KindTypeFunctionIf, ast.KindTypeFunctionWhile, ast.KindTypeFunctionForOf, ast.KindTypeFunctionForIn:
		s := node.AsTypeFunctionStatement()
		switch node.Kind {
		case ast.KindTypeFunctionIf:
			p.writeKeyword("if")
		case ast.KindTypeFunctionWhile:
			p.writeKeyword("while")
		default:
			p.writeKeyword("for")
		}
		p.writeSpace()
		p.writePunctuation("(")
		if node.Kind == ast.KindTypeFunctionForOf || node.Kind == ast.KindTypeFunctionForIn {
			p.writeKeyword("type")
			p.writeSpace()
			p.emitBindingIdentifier(node.Name().AsIdentifier())
			p.writeSpace()
			if node.Kind == ast.KindTypeFunctionForOf {
				p.writeKeyword("of")
			} else {
				p.writeKeyword("in")
			}
			p.writeSpace()
		}
		p.emitTypefn(s.Expression)
		p.writePunctuation(")")
		p.writeSpace()
		p.emitTypefn(s.Body)
		if s.ElseBody != nil {
			p.writeSpace()
			p.writeKeyword("else")
			p.writeSpace()
			p.emitTypefn(s.ElseBody)
		}
	default:
		e := node.AsTypeFunctionExpression()
		list := func() {
			for i, item := range e.Operands.Nodes {
				if i > 0 {
					p.writePunctuation(",")
					p.writeSpace()
				}
				p.emitTypefn(item)
			}
		}
		switch node.Kind {
		case ast.KindTypeFunctionName:
			p.emitEntityName(node.Name())
		case ast.KindTypeFunctionCall:
			p.emitEntityName(node.Name())
			p.writePunctuation("<")
			list()
			p.writePunctuation(">")
		case ast.KindTypeFunctionLiteral:
			p.emitTypeNodeOutsideExtends(e.NativeType)
		case ast.KindTypeFunctionNativeType:
			p.writeKeyword("type")
			p.writePunctuation("(")
			p.emitTypeNodeOutsideExtends(e.NativeType)
			p.writePunctuation(")")
		case ast.KindTypeFunctionTuple:
			p.writePunctuation("[")
			list()
			p.writePunctuation("]")
		case ast.KindTypeFunctionObject:
			p.writePunctuation("{")
			list()
			p.writePunctuation("}")
		case ast.KindTypeFunctionSpread:
			p.writePunctuation("...")
			p.emitTypefn(e.Operands.Nodes[0])
		case ast.KindTypeFunctionIndex:
			p.writePunctuation("(")
			p.emitTypefn(e.Operands.Nodes[0])
			p.writePunctuation(")[")
			p.emitTypefn(e.Operands.Nodes[1])
			p.writePunctuation("]")
		case ast.KindTypeFunctionUnary:
			p.writePunctuation("(")
			p.writeOperator(scanner.TokenToString(e.OperatorToken.Kind))
			p.writeSpace()
			p.emitTypefn(e.Operands.Nodes[0])
			p.writePunctuation(")")
		case ast.KindTypeFunctionBinary:
			p.writePunctuation("(")
			p.emitTypefn(e.Operands.Nodes[0])
			p.writeSpace()
			p.writeOperator(scanner.TokenToString(e.OperatorToken.Kind))
			p.writeSpace()
			p.emitTypefn(e.Operands.Nodes[1])
			p.writePunctuation(")")
		case ast.KindTypeFunctionField:
			if e.Readonly {
				p.writeKeyword("readonly")
				p.writeSpace()
			}
			name := node.Name()
			if ast.IsIdentifier(name) {
				p.emitBindingIdentifier(name.AsIdentifier())
			} else if ast.IsStringLiteral(name) || ast.IsNumericLiteral(name) {
				p.emitExpression(name, ast.OperatorPrecedenceLowest)
			} else {
				p.writePunctuation("[")
				p.emitTypefn(name)
				p.writePunctuation("]")
			}
			if e.Optional {
				p.writePunctuation("?")
			}
			p.writePunctuation(":")
			p.writeSpace()
			p.emitTypefn(e.Operands.Nodes[0])
		}
	}
	p.exitNode(node, state)
}
