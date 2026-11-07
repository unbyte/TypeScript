package printer

import (
	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/scanner"
)

// Printing DSL syntax must not route its nodes through the JavaScript emitter.
// Parenthesized operators deliberately preserve the DSL's angle-list grammar.
func (p *Printer) emitTypefn(node *ast.Node) {
	switch node.Kind {
	case ast.KindTypeFunctionDeclaration:
		d := node.AsTypeFunctionDeclaration()
		p.emitModifierList(node, node.Modifiers(), false)
		p.writeKeyword("typefn")
		p.writeSpace()
		p.emitBindingIdentifier(node.Name().AsIdentifier())
		p.writePunctuation("<")
		for i, parameter := range d.TypeParameters.Nodes {
			if i > 0 {
				p.writePunctuation(",")
				p.writeSpace()
			}
			p.emitTypeParameter(parameter.AsTypeParameterDeclaration())
		}
		p.writePunctuation(">")
		p.writeSpace()
		p.emitTypefn(d.Body)
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
}
