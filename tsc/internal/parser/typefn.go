package parser

import (
	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/diagnostics"
	"github.com/microsoft/TypeScript/tsc/internal/scanner"
)

// DSL bodies have their own statement and Pratt expression parser. Native parsing
// is entered only for declared bounds, literal leaves, and explicit type(...) islands.
type typefnCallParse struct {
	node *ast.Node
	end  scanner.ScannerState
}

func (p *Parser) isTypefnStart() bool {
	return p.opts.TypefnVersion == 1 && p.token == ast.KindIdentifier && p.scanner.TokenValue() == "typefn" &&
		p.lookAhead((*Parser).nextTokenIsIdentifierOnSameLine)
}

func (p *Parser) typefnError(message string) {
	p.parseErrorAtCurrentToken(diagnostics.NewAdHocMessage("typefn: " + message))
}

func (p *Parser) typefnStep() bool {
	if p.typefnAborted {
		return false
	}
	if p.typefnLimit == 0 {
		p.typefnLimit = p.opts.TypefnParserWork
		if p.typefnLimit == 0 {
			// Count lexical tokens once for the default work budget. This scan does
			// not create an AST, query symbols, or interpret template contents.
			s := scanner.NewScanner()
			s.SetText(p.sourceText)
			count := 0
			for s.Scan() != ast.KindEndOfFile {
				count++
			}
			p.typefnLimit = count*100 + 10000
		}
	}
	p.typefnWork++
	depthLimit := p.opts.TypefnSyntaxDepth
	if depthLimit == 0 {
		depthLimit = 256
	}
	if p.typefnWork > p.typefnLimit || p.typefnDepth > depthLimit {
		p.typefnAborted = true
		p.typefnError("parser work or nesting limit exceeded")
		// Resource recovery consumes the remaining file, never recursively.
		for p.token != ast.KindEndOfFile {
			p.nextToken()
		}
		return false
	}
	return true
}

func (p *Parser) parseTypefnDeclaration(pos int, jsdoc jsdocScannerInfo, modifiers *ast.ModifierList) *ast.Node {
	p.nextToken()
	name := p.parseIdentifier()
	if p.token != ast.KindLessThanToken {
		p.typefnError("expected an explicit parameter list")
	}
	parameters := p.parseTypeParameters()
	if parameters == nil {
		parameters = p.factory.NewNodeList(nil)
	}
	for _, parameter := range parameters.Nodes {
		if len(parameter.ModifierNodes()) != 0 {
			p.parseErrorAtRange(parameter.Loc, diagnostics.NewAdHocMessage("typefn: parameter modifiers are not supported"))
		}
		if parameter.AsTypeParameterDeclaration().DefaultType != nil {
			p.typefnError("parameter defaults are not supported")
		}
	}
	if p.token == ast.KindColonToken {
		p.nextToken()
		p.parseExpected(ast.KindTypeKeyword) // The initial Type-result profile.
	}
	body := p.parseTypefnBlock()
	node := p.finishNode(p.factory.NewTypeFunctionDeclaration(modifiers, name, parameters, body), pos)
	p.withJSDoc(node, jsdoc)
	p.checkJSSyntax(node)
	return node
}

func (p *Parser) parseTypefnBlock() *ast.Node {
	pos := p.nodePos()
	p.typefnDepth++
	defer func() { p.typefnDepth-- }()
	var statements []*ast.Node
	if p.typefnStep() && p.parseExpected(ast.KindOpenBraceToken) {
		for p.token != ast.KindCloseBraceToken && p.token != ast.KindEndOfFile {
			start := p.nodePos()
			statements = append(statements, p.parseTypefnStatement())
			if p.nodePos() == start {
				p.nextToken()
			}
		}
		p.parseExpected(ast.KindCloseBraceToken)
	}
	list := p.factory.NewNodeList(statements)
	list.Loc = core.NewTextRange(pos, p.nodePos())
	return p.finishNode(p.factory.NewTypeFunctionBlock(list), pos)
}

func (p *Parser) typefnStatement(kind ast.Kind, pos int, name, expression, body, alternative *ast.Node) *ast.Node {
	node := p.factory.NewTypeFunctionStatement(kind, name, expression, body, alternative)
	return p.finishNode(node, pos)
}

func (p *Parser) parseTypefnStatement() *ast.Node {
	p.typefnDepth++
	defer func() { p.typefnDepth-- }()
	pos := p.nodePos()
	if !p.typefnStep() {
		return p.typefnStatement(ast.KindTypeFunctionReturn, pos, nil, nil, nil, nil)
	}
	if p.isTypefnStart() {
		return p.parseTypefnDeclaration(pos, p.jsdocScannerInfo(), nil)
	}
	switch p.token {
	case ast.KindOpenBraceToken:
		return p.parseTypefnBlock()
	case ast.KindTypeKeyword:
		p.nextToken()
		name := p.parseIdentifier()
		p.parseExpected(ast.KindEqualsToken)
		expression := p.parseTypefnExpression(1)
		p.parseExpected(ast.KindSemicolonToken)
		return p.typefnStatement(ast.KindTypeFunctionBinding, pos, name, expression, nil, nil)
	case ast.KindReturnKeyword:
		p.nextToken()
		expression := p.parseTypefnExpression(1)
		p.parseExpected(ast.KindSemicolonToken)
		return p.typefnStatement(ast.KindTypeFunctionReturn, pos, nil, expression, nil, nil)
	case ast.KindIfKeyword, ast.KindWhileKeyword:
		kind := p.token
		p.nextToken()
		p.parseExpected(ast.KindOpenParenToken)
		condition := p.parseTypefnExpression(1)
		p.parseExpected(ast.KindCloseParenToken)
		body := p.parseTypefnBlock()
		var alternative *ast.Node
		if kind == ast.KindIfKeyword && p.token == ast.KindElseKeyword {
			p.nextToken()
			if p.token == ast.KindIfKeyword {
				alternative = p.parseTypefnStatement()
			} else {
				alternative = p.parseTypefnBlock()
			}
		}
		resultKind := ast.KindTypeFunctionIf
		if kind == ast.KindWhileKeyword {
			resultKind = ast.KindTypeFunctionWhile
		}
		return p.typefnStatement(resultKind, pos, nil, condition, body, alternative)
	case ast.KindForKeyword:
		p.nextToken()
		p.parseExpected(ast.KindOpenParenToken)
		p.parseExpected(ast.KindTypeKeyword)
		name := p.parseIdentifier()
		kind := ast.KindTypeFunctionForOf
		if p.token == ast.KindInKeyword {
			kind = ast.KindTypeFunctionForIn
			p.nextToken()
		} else {
			p.parseExpected(ast.KindOfKeyword)
		}
		subject := p.parseTypefnExpression(1)
		p.parseExpected(ast.KindCloseParenToken)
		return p.typefnStatement(kind, pos, name, subject, p.parseTypefnBlock(), nil)
	case ast.KindBreakKeyword, ast.KindContinueKeyword:
		kind := ast.KindTypeFunctionBreak
		if p.token == ast.KindContinueKeyword {
			kind = ast.KindTypeFunctionContinue
		}
		p.nextToken()
		p.parseExpected(ast.KindSemicolonToken)
		return p.typefnStatement(kind, pos, nil, nil, nil, nil)
	default:
		if p.isIdentifier() && p.lookAhead(func(p *Parser) bool { return p.nextToken() == ast.KindEqualsToken }) {
			name := p.parseIdentifier()
			p.nextToken()
			expression := p.parseTypefnExpression(1)
			p.parseExpected(ast.KindSemicolonToken)
			return p.typefnStatement(ast.KindTypeFunctionAssignment, pos, name, expression, nil, nil)
		}
		expression := p.parseTypefnExpression(1)
		if expression.Kind != ast.KindTypeFunctionCall {
			p.typefnError("only an application may be used as an expression statement")
		}
		p.parseExpected(ast.KindSemicolonToken)
		return p.typefnStatement(ast.KindTypeFunctionExpressionStatement, pos, nil, expression, nil, nil)
	}
}

func typefnPrecedence(kind ast.Kind) int {
	switch kind {
	case ast.KindBarBarToken:
		return 1
	case ast.KindAmpersandAmpersandToken:
		return 2
	case ast.KindIsKeyword, ast.KindEqualsEqualsToken, ast.KindExclamationEqualsToken, ast.KindLessThanToken, ast.KindLessThanEqualsToken, ast.KindGreaterThanToken, ast.KindGreaterThanEqualsToken:
		return 3
	case ast.KindBarToken:
		return 4
	case ast.KindAmpersandToken:
		return 5
	case ast.KindPlusToken, ast.KindMinusToken:
		return 6
	case ast.KindAsteriskToken, ast.KindSlashToken, ast.KindPercentToken:
		return 7
	}
	return 0
}

func (p *Parser) typefnExpression(kind ast.Kind, pos int, name *ast.Node, operands []*ast.Node, operator, nativeType *ast.Node) *ast.Node {
	var list *ast.NodeList
	if operands != nil {
		list = p.factory.NewNodeList(operands)
		list.Loc = core.NewTextRange(pos, p.nodePos())
	}
	node := p.factory.NewTypeFunctionExpression(kind, name, list, operator, nativeType, false, false)
	return p.finishNode(node, pos)
}

func (p *Parser) parseTypefnExpression(minimum int) *ast.Node {
	p.typefnDepth++
	defer func() { p.typefnDepth-- }()
	pos := p.nodePos()
	if !p.typefnStep() {
		return p.typefnExpression(ast.KindTypeFunctionName, pos, nil, nil, nil, nil)
	}
	left := p.parseTypefnPrefix()
	compared := false
	for {
		// The native scanner leaves > unsplit so Type grammars can close lists;
		// only full expressions rescan >= as a comparison.
		if minimum < 4 && p.token == ast.KindGreaterThanToken {
			p.token = p.scanner.ReScanGreaterThanToken()
		}
		power := typefnPrecedence(p.token)
		if power < minimum {
			break
		}
		if power == 3 && compared {
			p.typefnError("comparisons do not chain; use parentheses")
		}
		operator := p.parseTokenNode()
		right := p.parseTypefnExpression(power + 1)
		left = p.typefnExpression(ast.KindTypeFunctionBinary, pos, nil, []*ast.Node{left, right}, operator, nil)
		if power == 3 {
			compared = true
		}
	}
	return left
}

func (p *Parser) parseTypefnPrefix() *ast.Node {
	pos := p.nodePos()
	if p.token == ast.KindExclamationToken || p.token == ast.KindPlusToken || p.token == ast.KindMinusToken || p.token == ast.KindKeyOfKeyword {
		operator := p.parseTokenNode()
		return p.typefnExpression(ast.KindTypeFunctionUnary, pos, nil, []*ast.Node{p.parseTypefnExpression(8)}, operator, nil)
	}
	value := p.parseTypefnPrimary()
	for p.token == ast.KindOpenBracketToken {
		p.nextToken()
		index := p.parseTypefnExpression(1)
		p.parseExpected(ast.KindCloseBracketToken)
		value = p.typefnExpression(ast.KindTypeFunctionIndex, pos, nil, []*ast.Node{value, index}, nil, nil)
	}
	return value
}

func (p *Parser) parseTypefnPrimary() *ast.Node {
	pos := p.nodePos()
	switch p.token {
	case ast.KindOpenParenToken:
		p.nextToken()
		expression := p.parseTypefnExpression(1)
		p.parseExpected(ast.KindCloseParenToken)
		return expression
	case ast.KindTypeKeyword:
		p.nextToken()
		p.parseExpected(ast.KindOpenParenToken)
		native := p.parseType()
		p.parseExpected(ast.KindCloseParenToken)
		return p.typefnExpression(ast.KindTypeFunctionNativeType, pos, nil, nil, nil, native)
	case ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindBigIntLiteral, ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword, ast.KindNoSubstitutionTemplateLiteral:
		native := p.parseLiteralTypeNode(false)
		return p.typefnExpression(ast.KindTypeFunctionLiteral, pos, nil, nil, nil, native)
	case ast.KindAnyKeyword, ast.KindUnknownKeyword, ast.KindNeverKeyword, ast.KindStringKeyword, ast.KindNumberKeyword, ast.KindBigIntKeyword, ast.KindBooleanKeyword, ast.KindObjectKeyword, ast.KindVoidKeyword, ast.KindSymbolKeyword, ast.KindUndefinedKeyword:
		native := p.parseKeywordTypeNode()
		return p.typefnExpression(ast.KindTypeFunctionLiteral, pos, nil, nil, nil, native)
	case ast.KindOpenBracketToken:
		p.nextToken()
		items := make([]*ast.Node, 0)
		for p.token != ast.KindCloseBracketToken && p.token != ast.KindEndOfFile {
			start := p.nodePos()
			spread := p.token == ast.KindDotDotDotToken
			if spread {
				p.nextToken()
			}
			item := p.parseTypefnExpression(1)
			if spread {
				item = p.typefnExpression(ast.KindTypeFunctionSpread, start, nil, []*ast.Node{item}, nil, nil)
			}
			items = append(items, item)
			if p.token != ast.KindCommaToken {
				break
			}
			p.nextToken()
		}
		p.parseExpected(ast.KindCloseBracketToken)
		return p.typefnExpression(ast.KindTypeFunctionTuple, pos, nil, items, nil, nil)
	case ast.KindOpenBraceToken:
		return p.parseTypefnObject()
	default:
		// Bindings use parseIdentifier, which also accepts contextual keywords.
		// References must accept the same names (for example, out and async).
		if p.isIdentifier() {
			name := p.parseEntityName(true, false, nil)
			if p.token == ast.KindLessThanToken {
				if call := p.tryTypefnCall(name, pos); call != nil {
					return call
				}
			}
			return p.typefnExpression(ast.KindTypeFunctionName, pos, name, nil, nil, nil)
		}
		p.typefnError("expected a DSL expression; JavaScript expressions are not supported")
		if p.token != ast.KindEndOfFile {
			p.nextToken()
		}
		return p.typefnExpression(ast.KindTypeFunctionName, pos, nil, nil, nil, nil)
	}
}

func (p *Parser) parseTypefnObject() *ast.Node {
	pos := p.nodePos()
	p.nextToken()
	fields := make([]*ast.Node, 0)
	for p.token != ast.KindCloseBraceToken && p.token != ast.KindEndOfFile {
		start := p.nodePos()
		readonly := p.token == ast.KindReadonlyKeyword && p.lookAhead(func(p *Parser) bool {
			p.nextToken()
			return p.token != ast.KindColonToken && p.token != ast.KindQuestionToken
		})
		if readonly {
			p.nextToken()
		}
		var name *ast.Node
		if p.token == ast.KindOpenBracketToken {
			p.nextToken()
			name = p.parseTypefnExpression(1)
			p.parseExpected(ast.KindCloseBracketToken)
		} else if p.token == ast.KindStringLiteral || p.token == ast.KindNumericLiteral {
			name = p.parseLiteralExpression()
		} else {
			name = p.parseIdentifierName()
		}
		optional := p.token == ast.KindQuestionToken
		if optional {
			p.nextToken()
		}
		p.parseExpected(ast.KindColonToken)
		value := p.parseTypefnExpression(1)
		field := p.typefnExpression(ast.KindTypeFunctionField, start, name, []*ast.Node{value}, nil, nil)
		field.AsTypeFunctionExpression().Readonly = readonly
		field.AsTypeFunctionExpression().Optional = optional
		fields = append(fields, field)
		if p.token != ast.KindCommaToken && p.token != ast.KindSemicolonToken {
			break
		}
		p.nextToken()
	}
	p.parseExpected(ast.KindCloseBraceToken)
	return p.typefnExpression(ast.KindTypeFunctionObject, pos, nil, fields, nil, nil)
}

func (p *Parser) tryTypefnCall(name *ast.Node, pos int) *ast.Node {
	key := p.nodePos()
	if cached, ok := p.typefnCalls[key]; ok {
		if cached.node != nil {
			p.scanner.Rewind(cached.end)
			p.token = p.scanner.Token()
		}
		return cached.node
	}
	if p.typefnCalls == nil {
		p.typefnCalls = make(map[int]typefnCallParse)
	}
	state := p.mark()
	p.nextToken()
	arguments := make([]*ast.Node, 0)
	for p.token != ast.KindGreaterThanToken && p.token != ast.KindEndOfFile {
		arguments = append(arguments, p.parseTypefnExpression(4))
		if p.token != ast.KindCommaToken {
			break
		}
		p.nextToken()
	}
	if p.token == ast.KindGreaterThanToken && len(p.diagnostics) == state.diagnosticsLen {
		p.nextToken()
		if typefnCallFollower(p.token) {
			node := p.typefnExpression(ast.KindTypeFunctionCall, pos, name, arguments, nil, nil)
			p.typefnCalls[key] = typefnCallParse{node, p.scanner.Mark()}
			return node
		}
	}
	// A resource failure must survive speculation instead of being rewound away.
	if p.typefnAborted {
		return nil
	}
	p.rewind(state)
	p.typefnCalls[key] = typefnCallParse{}
	return nil
}

func typefnCallFollower(kind ast.Kind) bool {
	if typefnPrecedence(kind) != 0 {
		return true
	}
	switch kind {
	case ast.KindEndOfFile, ast.KindSemicolonToken, ast.KindCommaToken, ast.KindCloseParenToken, ast.KindCloseBracketToken, ast.KindCloseBraceToken, ast.KindOpenBracketToken:
		return true
	}
	return false
}
