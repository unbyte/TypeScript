package parser

import (
	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/diagnostics"
)

func (p *Parser) isTypeFunctionStart() bool {
	return p.token == ast.KindIdentifier && p.scanner.TokenText() == "typefn" && p.lookAhead(func(p *Parser) bool {
		return p.nextTokenIsIdentifierOnSameLine() && p.nextToken() == ast.KindLessThanToken
	})
}

func (p *Parser) parseTypeFunction(pos int, doc jsdocScannerInfo, modifiers *ast.ModifierList) *ast.Node {
	p.nextToken()
	name := p.parseIdentifier()
	parameters := p.parseBracketedList(PCTypeParameters, (*Parser).parseTypeFunctionParameter, ast.KindLessThanToken, ast.KindGreaterThanToken)
	saved := p.inTypeFunction
	p.inTypeFunction = true
	body := p.parseTypeFunctionBlock()
	p.inTypeFunction = saved
	result := p.finishNode(p.factory.NewTypeFunctionDeclaration(modifiers, name, parameters, body), pos)
	p.withJSDoc(result, doc)
	p.checkJSSyntax(result)
	return result
}

func (p *Parser) parseTypeFunctionBlock() *ast.Node {
	pos := p.nodePos()
	statements := p.createMissingList()
	if p.parseExpected(ast.KindOpenBraceToken) {
		statements = p.parseList(PCBlockStatements, (*Parser).parseTypeFunctionStatement)
		p.parseExpected(ast.KindCloseBraceToken)
	}
	return p.finishNode(p.factory.NewTypeFunctionBlock(statements), pos)
}

func (p *Parser) parseTypeFunctionBindingName() *ast.Node {
	pos := p.nodePos()
	return p.finishNode(p.factory.NewTypeParameterDeclaration(nil, p.parseIdentifier(), nil, nil, nil), pos)
}

func (p *Parser) parseTypeFunctionStatement() *ast.Node {
	pos := p.nodePos()
	var result *ast.Node
	switch p.token {
	case ast.KindTypeKeyword:
		if p.lookAhead(func(p *Parser) bool { return p.nextToken() == ast.KindDotToken }) {
			break
		}
		p.nextToken()
		listPos := p.nodePos()
		destructured := p.token == ast.KindOpenBracketToken
		var bindings *ast.NodeList
		if destructured {
			bindings = p.parseBracketedList(PCArrayBindingElements, (*Parser).parseTypeFunctionBindingName, ast.KindOpenBracketToken, ast.KindCloseBracketToken)
		} else {
			name := p.parseTypeFunctionBindingName()
			bindings = p.newNodeList(core.NewTextRange(listPos, p.nodePos()), []*ast.Node{name})
		}
		p.parseExpected(ast.KindEqualsToken)
		result = p.factory.NewTypeFunctionBinding(bindings, destructured, p.parseTypeFunctionOperand())
	case ast.KindReturnKeyword:
		p.nextToken()
		if p.hasPrecedingLineBreak() {
			p.parseErrorAtCurrentToken(diagnostics.Line_break_not_permitted_here)
		}
		result = p.factory.NewTypeFunctionReturn(p.parseTypeFunctionOperand())
	case ast.KindIfKeyword, ast.KindWhileKeyword:
		kind := p.token
		p.nextToken()
		p.parseExpected(ast.KindOpenParenToken)
		condition := p.parseTypeFunctionOperand()
		p.parseExpected(ast.KindCloseParenToken)
		body := p.parseTypeFunctionBlock()
		if kind == ast.KindWhileKeyword {
			return p.finishNode(p.factory.NewTypeFunctionWhile(condition, body), pos)
		}
		var alternative *ast.Node
		if p.parseOptional(ast.KindElseKeyword) {
			if p.token == ast.KindIfKeyword {
				alternative = p.parseTypeFunctionStatement()
			} else {
				alternative = p.parseTypeFunctionBlock()
			}
		}
		return p.finishNode(p.factory.NewTypeFunctionIf(condition, body, alternative), pos)
	case ast.KindForKeyword:
		p.nextToken()
		p.parseExpected(ast.KindOpenParenToken)
		p.parseExpected(ast.KindTypeKeyword)
		binding := p.parseTypeFunctionBindingName()
		p.parseExpected(ast.KindOfKeyword)
		sequence := p.parseTypeFunctionOperand()
		p.parseExpected(ast.KindCloseParenToken)
		return p.finishNode(p.factory.NewTypeFunctionFor(binding, sequence, p.parseTypeFunctionBlock()), pos)
	case ast.KindBreakKeyword:
		p.nextToken()
		result = p.factory.NewBreakStatement(nil)
	case ast.KindContinueKeyword:
		p.nextToken()
		result = p.factory.NewContinueStatement(nil)
	case ast.KindIdentifier:
		if p.scanner.TokenText() == "match" && !p.lookAhead(func(p *Parser) bool { return p.nextToken() == ast.KindLessThanToken }) {
			p.nextToken()
			value := p.parseTypeFunctionOperand()
			p.parseExpected(ast.KindOpenBraceToken)
			start := p.nodePos()
			var arms []*ast.Node
			for p.token == ast.KindCaseKeyword || p.token == ast.KindElseKeyword {
				armPos := p.nodePos()
				tag := p.token
				p.nextToken()
				var pattern *ast.Node
				if tag == ast.KindCaseKeyword {
					pattern = p.parseType()
				}
				p.parseOptional(ast.KindColonToken)
				arms = append(arms, p.finishNode(p.factory.NewTypeFunctionArm(pattern, p.parseTypeFunctionBlock()), armPos))
			}
			list := p.newNodeList(core.NewTextRange(start, p.nodePos()), arms)
			p.parseExpected(ast.KindCloseBraceToken)
			return p.finishNode(p.factory.NewTypeFunctionMatch(value, list), pos)
		}
		if p.lookAhead(func(p *Parser) bool { return p.nextToken() == ast.KindEqualsToken }) {
			target := p.parseIdentifier()
			p.nextToken()
			result = p.factory.NewTypeFunctionAssignment(target, p.parseTypeFunctionOperand())
		}
	}
	if result == nil {
		result = p.factory.NewTypeFunctionCallStatement(p.parseTypeFunctionOperand())
		if p.nodePos() == pos && p.token != ast.KindEndOfFile {
			p.nextToken()
		}
	}
	p.parseSemicolon()
	return p.finishNode(result, pos)
}

func (p *Parser) parseTypeFunctionOperand() *ast.Node {
	return p.parseTypeFunctionBinary(ast.TypePrecedenceTypeFunctionOr)
}

func (p *Parser) parseTypeFunctionBinary(minPrecedence ast.TypePrecedence) *ast.Node {
	pos := p.nodePos()
	var left *ast.Node
	switch {
	case p.parseOptional(ast.KindExclamationToken):
		left = p.finishNode(p.factory.NewTypeFunctionNot(p.parseTypeFunctionBinary(ast.TypePrecedenceTypeFunctionPrefix)), pos)
	case p.token == ast.KindMinusToken && !p.lookAhead((*Parser).nextTokenIsNumericOrBigIntLiteral):
		p.nextToken()
		left = p.finishNode(p.factory.NewTypeFunctionNegate(p.parseTypeFunctionBinary(ast.TypePrecedenceTypeFunctionPrefix)), pos)
	default:
		// Signed literal types are native operands, including in unions such as -1 | 1.
		left = p.parseNativeType()
	}
	for precedence := ast.GetTypeFunctionOperatorPrecedence(p.token); precedence >= minPrecedence; precedence = ast.GetTypeFunctionOperatorPrecedence(p.token) {
		if precedence == ast.TypePrecedenceTypeFunctionComparison && ast.IsTypeFunctionBinary(left) &&
			ast.GetTypeFunctionOperatorPrecedence(left.AsTypeFunctionBinary().OperatorToken.Kind) == ast.TypePrecedenceTypeFunctionComparison {
			p.parseErrorAtCurrentToken(diagnostics.Type_function_Colon_0, "parenthesize chained comparisons")
		}
		op := p.parseTokenNode()
		right := p.parseTypeFunctionBinary(precedence + 1)
		left = p.finishNode(p.factory.NewTypeFunctionBinary(left, op, right), pos)
	}
	return left
}

func (p *Parser) parseTypeFunctionParameter() *ast.Node {
	if !p.isIdentifier() || !p.lookAhead(func(p *Parser) bool { return p.nextToken() == ast.KindLessThanToken }) {
		return p.parseTypeParameter()
	}
	pos := p.nodePos()
	name := p.parseIdentifier()
	start := p.nodePos()
	holes := p.parseBracketedList(PCTypeParameters, (*Parser).parseIdentifier, ast.KindLessThanToken, ast.KindGreaterThanToken)
	slots := p.finishNode(p.factory.NewTypeFunctionConstructorSlots(holes), start)
	var defaultType *ast.TypeNode
	if p.parseOptional(ast.KindEqualsToken) {
		defaultType = p.parseType()
	}
	return p.finishNode(p.factory.NewTypeParameterDeclaration(nil, name, nil, slots, defaultType), pos)
}
