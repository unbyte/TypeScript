package ls

import (
	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/checker"
	"github.com/microsoft/TypeScript/tsc/internal/core"
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
		if ast.FindAncestor(node, ast.IsTypeFunctionLike) != ast.FindAncestor(declaration, ast.IsTypeFunctionLike) {
			text += " (capture, readonly)"
		} else if ast.IsTypeFunctionIteration(declaration) {
			text += " (iteration binding)"
		} else if declaration.Kind == ast.KindTypeFunctionCapturePattern {
			text += " (pattern binding)"
		} else {
			text += " (mutable binding)"
		}
	}
	writer := newDisplayPartsWriter(vs)
	writer.Write(text)
	return symbolDisplayInfo{displayParts: writer, declaration: declaration}, true
}
