package ls

import (
	"context"
	"strings"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/scanner"
)

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
