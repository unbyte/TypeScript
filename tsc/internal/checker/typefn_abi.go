package checker

import (
	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/parser"
)

func (c *Checker) checkTypefnDeclarationABI(file *ast.SourceFile) bool {
	if !file.IsDeclarationFile {
		return true
	}
	state := c.getTypefnState()
	diagnostic, known := state.declarationABI[file]
	if !known {
		if state.declarationABI == nil {
			state.declarationABI = map[*ast.SourceFile]*ast.Diagnostic{}
		}
		diagnostic = parser.TypefnDeclarationABIDiagnostic(file)
		state.declarationABI[file] = diagnostic
	}
	if diagnostic == nil {
		return true
	}
	// Declaration-file diagnostics can be omitted by skipLibCheck. Execution
	// failures belong to the consumer, with the ABI header as related provenance.
	// Cache the producer diagnostic, never a consumer-specific remapping.
	if vm := state.vm; vm != nil && (vm.rootSite != nil || vm.replaySite != nil) {
		diagnostic = vm.sourceDiagnostic(diagnostic)
	}
	c.addDiagnostic(diagnostic)
	return false
}
