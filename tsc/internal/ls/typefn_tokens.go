package ls

import (
	"context"
	"strings"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/checker"
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
		templateEnds, intrinsicTokens := typefnTokenMetadata(file, node)
		base := scanner.GetTokenPosOfNode(node, file, false)
		scan := scanner.NewScanner()
		scan.SetText(file.Text()[base:node.End()])
		for kind := scan.Scan(); kind != ast.KindEndOfFile; kind = scan.Scan() {
			if ctx.Err() != nil {
				return true
			}
			lo, hi := base+scan.TokenStart(), base+scan.TokenEnd()
			if kind == ast.KindCloseBraceToken && templateEnds[lo] {
				kind = scan.ReScanTemplateToken(false)
				hi = base + scan.TokenEnd()
			}
			if lo < start || hi > end {
				continue
			}
			category := tokenType(-1)
			switch {
			case kind == ast.KindIdentifier && scan.TokenValue() == "typefn" && lo < node.Name().Pos():
				category = tokenTypeKeyword
			case kind == ast.KindIdentifier:
				if token, ok := intrinsicTokens[lo]; ok {
					category = token
				}
			case ast.IsKeywordKind(kind):
				category = tokenTypeKeyword
			case kind == ast.KindStringLiteral || kind == ast.KindNoSubstitutionTemplateLiteral || kind == ast.KindTemplateHead || kind == ast.KindTemplateMiddle || kind == ast.KindTemplateTail:
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

// One tree walk supplies scanner mode transitions and intrinsic colors. Avoid
// searching the AST afresh for every identifier in a large DSL body.
func typefnTokenMetadata(file *ast.SourceFile, root *ast.Node) (map[int]bool, map[int]tokenType) {
	ends := map[int]bool{}
	intrinsics := map[int]tokenType{}
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindTypeFunctionKind {
			kind := node.AsTypeFunctionKind()
			intrinsics[scanner.GetTokenPosOfNode(kind.KindToken, file, false)] = tokenTypeKeyword
			if kind.Slots != nil {
				for _, slot := range kind.Slots.Nodes {
					intrinsics[scanner.GetTokenPosOfNode(slot, file, false)] = tokenTypeOperator
				}
			}
		}
		if node.Kind == ast.KindTypeFunctionParameter {
			if slots := node.AsTypeFunctionParameter().Slots; slots != nil {
				for _, slot := range slots.Nodes {
					intrinsics[scanner.GetTokenPosOfNode(slot, file, false)] = tokenTypeOperator
				}
			}
		}
		if ast.IsTypeFunctionLike(node) || node.Kind == ast.KindTypeFunctionMatch {
			intrinsics[scanner.GetTokenPosOfNode(node, file, false)] = tokenTypeKeyword
		}
		if node.Kind == ast.KindTemplateMiddle || node.Kind == ast.KindTemplateTail {
			ends[scanner.GetTokenPosOfNode(node, file, false)] = true
		}
		if ast.IsIdentifier(node) && checker.TypeFunctionIntrinsicAtLocation(node) != nil {
			kind := tokenTypeFunction
			if node.Parent != nil && ast.IsQualifiedName(node.Parent) && node.Parent.AsQualifiedName().Left == node {
				kind = tokenTypeNamespace
			}
			intrinsics[scanner.GetTokenPosOfNode(node, file, false)] = kind
		}
		return node.ForEachChild(visit)
	}
	visit(root)
	return ends, intrinsics
}
