package ls

import (
	"context"
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/astnav"
	"github.com/microsoft/TypeScript/tsc/internal/checker"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/lsp/lsproto"
	"github.com/microsoft/TypeScript/tsc/internal/scanner"
)

func typefnSymbolInfo(c *checker.Checker, symbol *ast.Symbol, node *ast.Node, vs bool) (symbolDisplayInfo, bool) {
	if symbol == nil {
		return symbolDisplayInfo{}, false
	}
	declaration := c.GetTypeFunctionDeclaration(symbol)
	text := ""
	if declaration != nil {
		text = c.TypeFunctionSignature(symbol)
	} else {
		declaration = core.Find(symbol.Declarations(), ast.IsTypeFunctionLocalDeclaration)
		if declaration == nil {
			return symbolDisplayInfo{}, false
		}
		text = "type " + symbol.Name() + ": type"
		if ast.IsTypeFunctionIteration(declaration) {
			text += " (iteration binding)"
		} else {
			text += " (mutable binding)"
		}
	}
	writer := newDisplayPartsWriter(vs)
	writer.Write(text)
	return symbolDisplayInfo{displayParts: writer, declaration: declaration}, true
}

// Locate the DSL scope even between tokens and in incomplete source.
func typefnNodeAtPosition(file *ast.SourceFile, position int) *ast.Node {
	// Recovered nodes end before trailing EOF trivia. Keep an unfinished
	// declaration's scope available while the user types in that trivia.
	if len(file.Diagnostics()) > 0 && position > file.EndOfFileToken.Pos() {
		end := file.EndOfFileToken.Pos()
		if end > 0 && file.Text()[end-1] != '}' {
			position = end
		}
	}
	var found *ast.Node
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node.Pos() > position || node.End() < position {
			return false
		}
		if ast.IsTypeFunctionNode(node) {
			found = node
		}
		node.ForEachChild(visit)
		return false
	}
	visit(file.AsNode())
	return found
}

func (l *LanguageService) typefnCompletions(ctx context.Context, file *ast.SourceFile, position int) *lsproto.CompletionList {
	scope := typefnNodeAtPosition(file, position)
	if scope == nil || scope.Kind == ast.KindTypeFunctionNativeType || ast.FindAncestor(scope, ast.IsTypeFunctionDeclaration) == nil {
		return nil
	}
	c, done := l.program.GetTypeCheckerForFile(ctx, file)
	defer done()
	symbols := c.GetSymbolsInScope(scope, ast.SymbolFlagsType|ast.SymbolFlagsNamespace|ast.SymbolFlagsAlias)
	memberCompletion := false
	for name := scope.Name(); name != nil && ast.IsQualifiedName(name); name = name.AsQualifiedName().Left {
		qualifier := name.AsQualifiedName().Left
		if position <= qualifier.End() {
			continue
		}
		memberCompletion = true
		symbols = nil
		symbol := c.GetSymbolAtLocation(qualifier)
		if symbol != nil && symbol.Flags()&ast.SymbolFlagsAlias != 0 {
			symbol = c.GetAliasedSymbol(symbol)
		}
		if symbol != nil && symbol.Flags()&ast.SymbolFlagsNamespace != 0 {
			symbols = c.GetExportsOfModule(symbol)
		}
		break
	}
	var items []*lsproto.CompletionItem
	seen := map[string]bool{}
	for _, symbol := range symbols {
		target := symbol
		if target.Flags()&ast.SymbolFlagsAlias != 0 {
			target = c.GetAliasedSymbol(target)
		}
		if memberCompletion && (target == nil || target.Flags()&(ast.SymbolFlagsType|ast.SymbolFlagsNamespace) == 0) {
			continue
		}
		if ctx.Err() != nil {
			return &lsproto.CompletionList{Items: items}
		}
		name := symbol.Name()
		if seen[name] || strings.HasPrefix(name, "__") {
			continue
		}
		decl := core.Find(symbol.Declarations(), ast.IsTypeFunctionLocalDeclaration)
		if decl != nil && decl.End() > position && decl.Kind == ast.KindTypeFunctionBinding {
			continue
		}
		kind := lsproto.CompletionItemKindClass
		detail := ""
		if fn := c.GetTypeFunctionDeclaration(symbol); fn != nil {
			kind = lsproto.CompletionItemKindFunction
			detail = c.TypeFunctionSignature(symbol)
		}
		if decl != nil {
			kind = lsproto.CompletionItemKindTypeParameter
			detail = "type " + name + ": type"
		}
		item := &lsproto.CompletionItem{Label: name, Kind: &kind}
		if detail != "" {
			item.Detail = &detail
		}
		items = append(items, item)
		seen[name] = true
	}
	keywords := []string{"type", "return", "if", "else", "while", "for", "break", "continue", "keyof", "never", "unknown", "any", "string", "number", "bigint", "boolean", "true", "false"}
	if memberCompletion {
		keywords = nil
	}
	for _, word := range keywords {
		if !seen[word] {
			items = append(items, &lsproto.CompletionItem{Label: word, Kind: new(lsproto.CompletionItemKindKeyword)})
		}
	}
	slices.SortFunc(items, func(a, b *lsproto.CompletionItem) int { return strings.Compare(a.Label, b.Label) })
	return &lsproto.CompletionList{Items: items}
}

// Scan balanced delimiters up to the cursor, including recovered/incomplete
// input. An angle opener is a call candidate only if its preceding name resolves
// to a typefn. Thus a comparison of local Types does not acquire a call signature.
func (l *LanguageService) typefnSignatureHelp(ctx context.Context, file *ast.SourceFile, position int) *lsproto.SignatureHelp {
	start := -1
	if scope := typefnNodeAtPosition(file, position); scope != nil {
		if fn := ast.FindAncestor(scope, ast.IsTypeFunctionDeclaration); fn != nil && fn.AsTypeFunctionDeclaration().Body != nil {
			start = fn.AsTypeFunctionDeclaration().Body.Pos()
		}
	}
	if start < 0 {
		lookupPosition := position
		if len(file.Diagnostics()) > 0 && lookupPosition > file.EndOfFileToken.Pos() {
			lookupPosition = file.EndOfFileToken.Pos()
		}
		var visit func(*ast.Node) bool
		visit = func(node *ast.Node) bool {
			if node.Pos() > lookupPosition || node.End() < lookupPosition {
				return false
			}
			if ast.IsTypeReferenceNode(node) && position > node.AsTypeReferenceNode().TypeName.End() && start < 0 {
				start = node.Pos()
			}
			node.ForEachChild(visit)
			return false
		}
		visit(file.AsNode())
	}
	if start < 0 || start > position {
		return nil
	}
	type frame struct {
		kind     ast.Kind
		namePos  int
		argument uint32
	}
	var stack []frame
	scan := scanner.NewScanner()
	scan.SetText(file.Text()[start:position])
	previous, previousPos := ast.KindUnknown, 0
	for kind := scan.Scan(); kind != ast.KindEndOfFile; kind = scan.Scan() {
		if ctx.Err() != nil {
			return nil
		}
		switch kind {
		case ast.KindLessThanToken, ast.KindOpenParenToken, ast.KindOpenBracketToken, ast.KindOpenBraceToken:
			namePos := -1
			if kind == ast.KindLessThanToken && previous == ast.KindIdentifier {
				namePos = previousPos
			}
			stack = append(stack, frame{kind: kind, namePos: namePos})
		case ast.KindGreaterThanToken, ast.KindCloseParenToken, ast.KindCloseBracketToken, ast.KindCloseBraceToken:
			opener := ast.KindLessThanToken
			switch kind {
			case ast.KindCloseParenToken:
				opener = ast.KindOpenParenToken
			case ast.KindCloseBracketToken:
				opener = ast.KindOpenBracketToken
			case ast.KindCloseBraceToken:
				opener = ast.KindOpenBraceToken
			}
			if opener != ast.KindLessThanToken {
				// A '<' within an expression group may be a comparison. Its
				// tentative angle frame must not survive the group's closing
				// delimiter and steal commas from the enclosing application.
				for len(stack) > 0 && stack[len(stack)-1].kind == ast.KindLessThanToken {
					stack = stack[:len(stack)-1]
				}
			}
			if len(stack) > 0 && stack[len(stack)-1].kind == opener {
				stack = stack[:len(stack)-1]
			}
		case ast.KindCommaToken:
			if len(stack) > 0 {
				stack[len(stack)-1].argument++
			}
		case ast.KindSemicolonToken:
			// A recovered angle expression cannot continue into the next statement.
			for len(stack) > 0 && stack[len(stack)-1].kind == ast.KindLessThanToken {
				stack = stack[:len(stack)-1]
			}
		}
		previous, previousPos = kind, start+scan.TokenStart()
	}
	c, done := l.program.GetTypeCheckerForFile(ctx, file)
	defer done()
	for i := len(stack) - 1; i >= 0; i-- {
		entry := stack[i]
		if entry.kind != ast.KindLessThanToken || entry.namePos < 0 {
			continue
		}
		name := astnav.GetTokenAtPosition(file, entry.namePos)
		symbol := c.GetSymbolAtLocation(name)
		declaration := c.GetTypeFunctionDeclaration(symbol)
		if declaration == nil {
			continue
		}
		var parameters []*lsproto.ParameterInformation
		for _, parameter := range declaration.TypeParameters() {
			label := scanner.GetTextOfNode(parameter)
			parameters = append(parameters, &lsproto.ParameterInformation{Label: lsproto.StringOrTuple{String: &label}})
		}
		return &lsproto.SignatureHelp{Signatures: []*lsproto.SignatureInformation{{Label: c.TypeFunctionSignature(symbol), Parameters: &parameters}}, ActiveSignature: new(uint32(0)), ActiveParameter: &lsproto.UintegerOrNull{Uinteger: &entry.argument}}
	}
	return nil
}

// Use the native scanner for lexical coloring, including contextual `typefn`.
// Identifier colors and modifiers still come from bound symbols. Source ranges
// are converted by the standard UTF-16/span-mapping pipeline.
func (l *LanguageService) typefnLexicalTokens(ctx context.Context, file *ast.SourceFile, start, end int) []semanticToken {
	var result []semanticToken
	factory := ast.NodeFactory{}
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if ctx.Err() != nil {
			return true
		}
		if !ast.IsTypeFunctionDeclaration(node) {
			return node.ForEachChild(visit)
		}
		if node.End() <= start || node.Pos() >= end {
			return false
		}
		base := scanner.GetTokenPosOfNode(node, file, false)
		scan := scanner.NewScanner()
		scan.SetText(file.Text()[base:node.End()])
		for kind := scan.Scan(); kind != ast.KindEndOfFile; kind = scan.Scan() {
			if ctx.Err() != nil {
				return true
			}
			lo, hi := base+scan.TokenStart(), base+scan.TokenEnd()
			if lo < start || hi > end {
				continue
			}
			category := tokenType(-1)
			switch {
			case kind == ast.KindIdentifier && scan.TokenValue() == "typefn" && lo < node.Name().Pos():
				category = tokenTypeKeyword
			case ast.IsKeywordKind(kind):
				category = tokenTypeKeyword
			case kind == ast.KindStringLiteral || kind == ast.KindNoSubstitutionTemplateLiteral:
				category = tokenTypeString
			case kind == ast.KindNumericLiteral || kind == ast.KindBigIntLiteral:
				category = tokenTypeNumber
			case ast.IsPunctuationKind(kind):
				category = tokenTypeOperator
			}
			// Multiline strings are handled by the editor's lexical grammar; LSP tokens
			// must occupy one line unless the client explicitly advertises otherwise.
			if category < 0 || strings.ContainsAny(file.Text()[lo:hi], "\r\n") {
				continue
			}
			token := factory.NewToken(kind)
			token.Loc = core.NewTextRange(lo, hi)
			token.Parent = file.AsNode()
			result = append(result, semanticToken{node: token, file: file, tokenType: category})
		}
		return false
	}
	visit(file.AsNode())
	return result
}
