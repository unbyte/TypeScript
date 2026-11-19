package printer

import (
	"github.com/microsoft/TypeScript/tsc/internal/ast"
)

func (p *Printer) emitTypeFunctionDeclaration(node *ast.TypeFunctionDeclaration) {
	state := p.enterNode(node.AsNode())
	p.emitModifierList(node.AsNode(), node.Modifiers(), false)
	p.writeKeyword("typefn")
	p.writeSpace()
	p.emitBindingIdentifier(node.Name().AsIdentifier())
	if len(node.TypeParameters.Nodes) == 0 {
		p.writePunctuation("<>")
	} else {
		p.emitTypeParameters(node.AsNode(), node.TypeParameters)
	}
	p.writeSpace()
	p.emitTypeFunctionBlock(node.Body)
	p.exitNode(node.AsNode(), state)
}

func (p *Printer) emitTypeFunctionBlock(node *ast.Node) {
	state := p.enterNode(node)
	statements := node.AsTypeFunctionBlock().Statements
	p.emitToken(ast.KindOpenBraceToken, node.Pos(), WriteKindPunctuation, node)
	p.emitList((*Printer).emitTypeFunctionStatement, node, statements, LFMultiLineBlockStatements)
	p.emitTokenEx(ast.KindCloseBraceToken, statements.End(), WriteKindPunctuation, node, tefIndentLeadingComments)
	p.exitNode(node, state)
}

func (p *Printer) emitTypeFunctionStatement(node *ast.Node) {
	state := p.enterNode(node)
	switch node.Kind {
	case ast.KindTypeFunctionBinding:
		binding := node.AsTypeFunctionBinding()
		p.writeKeyword("type")
		p.writeSpace()
		if binding.Destructured {
			p.writePunctuation("[")
		}
		for i, b := range binding.Bindings.Nodes {
			if i != 0 {
				p.writePunctuation(",")
				p.writeSpace()
			}
			p.emitBindingIdentifier(b.Name().AsIdentifier())
		}
		if binding.Destructured {
			p.writePunctuation("]")
		}
		p.writeSpace()
		p.writePunctuation("=")
		p.writeSpace()
		p.emitTypeNodeOutsideExtends(binding.Initializer)
		p.writeTrailingSemicolon()
	case ast.KindTypeFunctionAssignment:
		assignment := node.AsTypeFunctionAssignment()
		p.emitBindingIdentifier(assignment.Target.AsIdentifier())
		p.writeSpace()
		p.writePunctuation("=")
		p.writeSpace()
		p.emitTypeNodeOutsideExtends(assignment.Value)
		p.writeTrailingSemicolon()
	case ast.KindTypeFunctionReturn, ast.KindTypeFunctionCallStatement:
		if node.Kind == ast.KindTypeFunctionReturn {
			p.writeKeyword("return")
			p.writeSpace()
			p.emitTypeNodeOutsideExtends(node.AsTypeFunctionReturn().Operand)
		} else {
			p.emitTypeNodeOutsideExtends(node.AsTypeFunctionCallStatement().Operand)
		}
		p.writeTrailingSemicolon()
	case ast.KindBreakStatement, ast.KindContinueStatement:
		if node.Kind == ast.KindBreakStatement {
			p.writeKeyword("break")
		} else {
			p.writeKeyword("continue")
		}
		p.writeTrailingSemicolon()
	case ast.KindTypeFunctionIf, ast.KindTypeFunctionWhile:
		condition, block, alternative := ast.TypeFunctionBranchParts(node)
		if node.Kind == ast.KindTypeFunctionIf {
			p.writeKeyword("if")
		} else {
			p.writeKeyword("while")
		}
		p.writeSpace()
		p.writePunctuation("(")
		p.emitTypeNodeOutsideExtends(condition)
		p.writePunctuation(")")
		p.writeSpace()
		p.emitTypeFunctionBlock(block)
		if alternative != nil {
			p.writeSpace()
			p.writeKeyword("else")
			p.writeSpace()
			if ast.IsTypeFunctionBlock(alternative) {
				p.emitTypeFunctionBlock(alternative)
			} else {
				p.emitTypeFunctionStatement(alternative)
			}
		}
	case ast.KindTypeFunctionFor:
		loop := node.AsTypeFunctionFor()
		p.writeKeyword("for")
		p.writeSpace()
		p.writePunctuation("(")
		p.writeKeyword("type")
		p.writeSpace()
		p.emitBindingIdentifier(loop.Binding.Name().AsIdentifier())
		p.writeSpace()
		p.writeKeyword("of")
		p.writeSpace()
		p.emitTypeNodeOutsideExtends(loop.Sequence)
		p.writePunctuation(")")
		p.writeSpace()
		p.emitTypeFunctionBlock(loop.Block)
	case ast.KindTypeFunctionMatch:
		match := node.AsTypeFunctionMatch()
		p.writeKeyword("match")
		p.writeSpace()
		p.emitTypeNodeOutsideExtends(match.Operand)
		p.writeSpace()
		p.writePunctuation("{")
		p.emitList((*Printer).emitTypeFunctionStatement, node, match.Arms, LFMultiLineBlockStatements)
		p.writePunctuation("}")
	case ast.KindTypeFunctionArm:
		arm := node.AsTypeFunctionArm()
		if arm.Pattern != nil {
			p.writeKeyword("case")
			p.writeSpace()
			p.emitTypeNodeOutsideExtends(arm.Pattern)
		} else {
			p.writeKeyword("else")
		}
		p.writePunctuation(":")
		p.writeSpace()
		p.emitTypeFunctionBlock(arm.Block)
	}
	p.exitNode(node, state)
}

func (p *Printer) emitTypeFunctionExpression(node *ast.Node) {
	state := p.enterNode(node)
	switch node.Kind {
	case ast.KindTypeFunctionNot:
		p.writePunctuation("!")
		p.emitTypeNode(node.AsTypeFunctionNot().Operand, ast.TypePrecedenceTypeFunctionPrefix)
	case ast.KindTypeFunctionNegate:
		operand := node.AsTypeFunctionNegate().Operand
		p.writePunctuation("-")
		if ast.IsTypeFunctionNegate(operand) || ast.IsLiteralTypeNode(operand) && ast.IsPrefixUnaryExpression(operand.AsLiteralTypeNode().Literal) {
			p.writeSpace() // Keep the two minus tokens separate.
		}
		p.emitTypeNode(operand, ast.TypePrecedenceTypeFunctionPrefix)
	case ast.KindTypeFunctionBinary:
		binary := node.AsTypeFunctionBinary()
		precedence := ast.GetTypeFunctionOperatorPrecedence(binary.OperatorToken.Kind)
		leftPrecedence := precedence
		if precedence == ast.TypePrecedenceTypeFunctionComparison {
			leftPrecedence++ // Comparisons require explicit grouping on either side.
		}
		if ast.IsFunctionTypeNode(binary.Left) || ast.IsConstructorTypeNode(binary.Left) {
			leftPrecedence = ast.TypePrecedenceUnion // A signature's return type consumes the following operator.
		}
		p.emitTypeNode(binary.Left, leftPrecedence)
		p.writeSpace()
		p.emitTokenNode(binary.OperatorToken)
		p.writeSpace()
		p.emitTypeNode(binary.Right, precedence+1)
	}
	p.exitNode(node, state)
}
