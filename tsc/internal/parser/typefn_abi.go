package parser

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/diagnostics"
	"github.com/microsoft/TypeScript/tsc/internal/scanner"
)

// TypefnDeclarationABIDiagnostic validates a declaration file known to contain
// executable DSL definitions or constructor aliases. Definitions are known during
// parsing; constructor-valued native aliases are discovered by the checker.
// This check is independent of skipLibCheck and precedes VM execution.
func TypefnDeclarationABIDiagnostic(file *ast.SourceFile) *ast.Diagnostic {
	if !file.IsDeclarationFile {
		return nil
	}
	text := file.Text()
	factory := ast.NewNodeFactory(ast.NodeFactoryHooks{})
	version := ""
	found := false
	loc := core.NewTextRange(0, 0)
	failure := func(at core.TextRange, message string) *ast.Diagnostic {
		return ast.NewDiagnostic(file, at, diagnostics.NewAdHocMessage("typefn: "+message))
	}
	for comment := range scanner.GetLeadingCommentRanges(factory, text, 0) {
		if comment.Kind != ast.KindSingleLineCommentTrivia {
			continue
		}
		body := strings.TrimSpace(text[comment.Pos()+2 : comment.End()])
		raw, ok := strings.CutPrefix(body, "@typefn-abi:")
		if !ok {
			continue
		}
		next := strings.TrimSpace(raw)
		if found && next != version {
			return failure(comment.TextRange, "conflicting declaration ABI headers")
		}
		version, found, loc = next, true, comment.TextRange
	}
	if !found {
		return failure(loc, "declaration file requires '// @typefn-abi: 1'")
	}
	if version != "1" {
		return failure(loc, "unsupported declaration ABI "+version+"; expected 1")
	}
	return nil
}
