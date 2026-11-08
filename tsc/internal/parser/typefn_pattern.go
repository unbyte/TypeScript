package parser

import (
	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/core"
)

func (p *Parser) typefnPattern(kind ast.Kind, pos int, name, expression *ast.Node, elements []*ast.Node) *ast.Node {
	var list *ast.NodeList
	if elements != nil {
		list = p.newNodeList(core.NewTextRange(pos, p.nodePos()), elements)
	}
	return p.finishNode(p.factory.NewTypeFunctionPattern(kind, name, expression, list), pos)
}

func (p *Parser) parseTypefnMatch() *ast.Node {
	pos := p.nodePos()
	p.nextToken()
	p.parseExpected(ast.KindOpenParenToken)
	subject := p.parseTypefnExpression(1)
	p.parseExpected(ast.KindCloseParenToken)
	p.parseExpected(ast.KindOpenBraceToken)
	start := p.nodePos()
	var arms []*ast.Node
	for p.token != ast.KindCloseBraceToken && p.token != ast.KindEndOfFile {
		armPos := p.nodePos()
		pattern := p.parseTypefnPattern()
		var guard *ast.Node
		if p.token == ast.KindIfKeyword {
			p.nextToken()
			p.parseExpected(ast.KindOpenParenToken)
			guard = p.parseTypefnExpression(1)
			p.parseExpected(ast.KindCloseParenToken)
		}
		p.parseExpected(ast.KindEqualsGreaterThanToken)
		body := p.parseTypefnBlock()
		arms = append(arms, p.finishNode(p.factory.NewTypeFunctionMatchArm(pattern, guard, body), armPos))
		if p.nodePos() == armPos {
			p.nextToken()
		}
	}
	if len(arms) == 0 {
		p.typefnError("match requires at least one arm")
	}
	list := p.newNodeList(core.NewTextRange(start, p.nodePos()), arms)
	p.parseExpected(ast.KindCloseBraceToken)
	return p.finishNode(p.factory.NewTypeFunctionMatch(subject, list), pos)
}

func (p *Parser) parseTypefnCapture() *ast.Node {
	pos := p.nodePos()
	name := p.parseIdentifier()
	kind := ast.KindTypeFunctionCapturePattern
	if name.Text() == "_" {
		kind = ast.KindTypeFunctionWildcardPattern
	}
	return p.typefnPattern(kind, pos, name, nil, nil)
}

func (p *Parser) parseTypefnPattern() *ast.Node {
	p.typefnDepth++
	defer func() { p.typefnDepth-- }()
	pos := p.nodePos()
	if !p.typefnStep() {
		return p.typefnPattern(ast.KindTypeFunctionWildcardPattern, pos, nil, nil, nil)
	}
	switch p.token {
	case ast.KindIsKeyword:
		p.nextToken()
		saved := p.typefnPatternFollow
		p.typefnPatternFollow = true
		expression := p.parseTypefnExpression(4)
		p.typefnPatternFollow = saved
		return p.typefnPattern(ast.KindTypeFunctionIsPattern, pos, nil, expression, nil)
	case ast.KindTypeKeyword:
		expression := p.parseTypefnPrimary()
		return p.typefnPattern(ast.KindTypeFunctionNativePattern, pos, nil, expression, nil)
	case ast.KindMinusToken:
		operator := p.parseTokenNode()
		if p.token != ast.KindNumericLiteral && p.token != ast.KindBigIntLiteral {
			p.typefnError("negative pattern requires a numeric literal")
		}
		value := p.parseTypefnPrimary()
		expression := p.typefnExpression(ast.KindTypeFunctionUnary, pos, nil, []*ast.Node{value}, operator, nil)
		return p.typefnPattern(ast.KindTypeFunctionLiteralPattern, pos, nil, expression, nil)
	case ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindBigIntLiteral, ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword, ast.KindUndefinedKeyword:
		return p.typefnPattern(ast.KindTypeFunctionLiteralPattern, pos, nil, p.parseTypefnPrimary(), nil)
	case ast.KindNoSubstitutionTemplateLiteral:
		return p.typefnPattern(ast.KindTypeFunctionTemplatePattern, pos, nil, p.parseTypefnPrimary(), nil)
	case ast.KindTemplateHead:
		head := p.parseTemplateHead(false)
		start := p.nodePos()
		var spans []*ast.Node
		for {
			if !p.typefnStep() {
				break
			}
			spanPos := p.nodePos()
			var expression *ast.Node
			if p.token == ast.KindEqualsToken {
				p.nextToken()
				expression = p.typefnPattern(ast.KindTypeFunctionPinPattern, spanPos, nil, p.parseTypefnExpression(1), nil)
			} else {
				expression = p.parseTypefnCapture()
			}
			literal := p.parseLiteralOfTemplateSpan(false)
			spans = append(spans, p.finishNode(p.factory.NewTypeFunctionTemplateSpan(expression, literal), spanPos))
			if literal.Kind != ast.KindTemplateMiddle {
				break
			}
		}
		template := p.finishNode(p.factory.NewTypeFunctionTemplate(head, p.newNodeList(core.NewTextRange(start, p.nodePos()), spans)), pos)
		return p.typefnPattern(ast.KindTypeFunctionTemplatePattern, pos, nil, template, nil)
	case ast.KindOpenBracketToken:
		p.nextToken()
		elements := make([]*ast.Node, 0)
		for p.token != ast.KindCloseBracketToken && p.token != ast.KindEndOfFile {
			rest := p.token == ast.KindDotDotDotToken
			start := p.nodePos()
			var element *ast.Node
			if rest {
				p.nextToken()
				element = p.typefnPattern(ast.KindTypeFunctionRestPattern, start, nil, p.parseTypefnCapture(), nil)
			} else {
				element = p.parseTypefnPattern()
			}
			elements = append(elements, element)
			if p.token != ast.KindCommaToken {
				break
			}
			p.nextToken()
			if rest {
				p.typefnError("tuple rest capture must be last")
				break
			}
		}
		p.parseExpected(ast.KindCloseBracketToken)
		return p.typefnPattern(ast.KindTypeFunctionTuplePattern, pos, nil, nil, elements)
	case ast.KindOpenBraceToken:
		p.nextToken()
		fields := make([]*ast.Node, 0)
		for p.token != ast.KindCloseBraceToken && p.token != ast.KindEndOfFile {
			start := p.nodePos()
			var name *ast.Node
			if p.token == ast.KindStringLiteral || p.token == ast.KindNumericLiteral {
				name = p.parseLiteralExpression()
			} else {
				name = p.parseIdentifierName()
			}
			p.parseExpected(ast.KindColonToken)
			field := p.typefnPattern(ast.KindTypeFunctionFieldPattern, start, name, p.parseTypefnPattern(), nil)
			fields = append(fields, field)
			if p.token != ast.KindCommaToken {
				break
			}
			p.nextToken()
		}
		p.parseExpected(ast.KindCloseBraceToken)
		return p.typefnPattern(ast.KindTypeFunctionObjectPattern, pos, nil, nil, fields)
	default:
		return p.parseTypefnCapture()
	}
}
