package printer

import (
	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/scanner"
)

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
		}
		p.emitTypeNodeOutsideExtends(node.AsTypeFunctionOperandStatement().Operand)
		p.writeTrailingSemicolon()
	case ast.KindBreakStatement, ast.KindContinueStatement:
		if node.Kind == ast.KindBreakStatement {
			p.writeKeyword("break")
		} else {
			p.writeKeyword("continue")
		}
		p.writeTrailingSemicolon()
	case ast.KindTypeFunctionIf, ast.KindTypeFunctionWhile:
		branch := node.AsTypeFunctionBranch()
		if node.Kind == ast.KindTypeFunctionIf {
			p.writeKeyword("if")
		} else {
			p.writeKeyword("while")
		}
		p.writeSpace()
		p.writePunctuation("(")
		p.emitTypeNodeOutsideExtends(branch.Condition)
		p.writePunctuation(")")
		p.writeSpace()
		p.emitTypeFunctionBlock(branch.Block)
		if branch.ElseBranch != nil {
			p.writeSpace()
			p.writeKeyword("else")
			p.writeSpace()
			if ast.IsTypeFunctionBlock(branch.ElseBranch) {
				p.emitTypeFunctionBlock(branch.ElseBranch)
			} else {
				p.emitTypeFunctionStatement(branch.ElseBranch)
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
		p.writeSpace()
		p.emitTypeFunctionBlock(arm.Block)
	}
	p.exitNode(node, state)
}

func (p *Printer) emitTypeFunctionExpression(node *ast.Node) {
	state := p.enterNode(node)
	if ast.IsTypeFunctionNot(node) {
		p.writePunctuation("!")
		p.emitTypeNodeOutsideExtends(node.AsTypeFunctionNot().Operand)
	} else {
		binary := node.AsTypeFunctionBinary()
		p.emitTypeNodeOutsideExtends(binary.Left)
		p.writeSpace()
		p.writePunctuation(scanner.TokenToString(binary.OperatorToken.Kind))
		p.writeSpace()
		p.emitTypeNodeOutsideExtends(binary.Right)
	}
	p.exitNode(node, state)
}
