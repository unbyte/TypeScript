package declarations

import (
	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/diagnostics"
	"github.com/microsoft/TypeScript/tsc/internal/printer"
	"github.com/microsoft/TypeScript/tsc/internal/scanner"
)

func (tx *DeclarationTransformer) typefnError(node *ast.Node, message string) {
	tx.state.invalidTypefnRecipe = true
	tx.state.addDiagnostic(createDiagnosticForNode(node, diagnostics.NewAdHocMessage("typefn: "+message)))
}

func (tx *DeclarationTransformer) checkTypefnDeclaration(node *ast.Node) bool {
	if tx.compilerOptions.GetTypefnVersion() != 1 {
		return true
	}
	needsABI, valid := tx.resolver.TypefnDeclarationStatus(node)
	tx.needsTypefnABI = tx.needsTypefnABI || needsABI
	if !valid {
		tx.typefnError(node, "cannot emit an invalid declaration recipe")
	}
	return valid
}

func (tx *DeclarationTransformer) transformTypeFunctionDeclaration(node *ast.Node) *ast.Node {
	if !tx.checkTypefnDeclaration(node) {
		// Recovered syntax is useful to editor services, never to recipe emission.
		return node
	}
	for _, dependency := range tx.resolver.GetTypefnDeclarationDependencies(node) {
		if dependency.Accessibility == printer.SymbolAccessibilityAccessible {
			tx.tracker.handleSymbolAccessibilityError(dependency.SymbolAccessibilityResult)
			if tx.state.stripInternal {
				for _, declaration := range dependency.Declarations {
					source := ast.GetSourceFileOfNode(declaration)
					if source.IsDeclarationFile || !tx.host.SourceFileMayBeEmitted(source, false) {
						continue
					}
					for parent := declaration; parent != nil && !ast.IsSourceFile(parent); parent = parent.Parent {
						if tx.isInternalDeclaration(parent, source) {
							tx.typefnError(dependency.Reference, "stripped declaration dependency: "+scanner.GetTextOfNode(dependency.Reference))
							break
						}
					}
				}
			}
		} else {
			tx.typefnError(dependency.ErrorNode, "declaration recipe dependency cannot be named: "+dependency.ErrorSymbolName)
		}
	}
	// A recipe is executable type syntax. Visiting its body as native declaration
	// syntax would erase control flow, mutable bindings and frozen closure creation.
	tx.needsDeclare = false
	d := node.AsTypeFunctionDeclaration()
	return tx.Factory().UpdateTypeFunctionDeclaration(d, tx.ensureModifiers(node), d.Name(), d.TypeParameters, d.ResultKind, d.Body)
}
