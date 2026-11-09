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
		return false
	}
	if !needsABI {
		return true
	}
	for _, dependency := range tx.resolver.GetTypefnDeclarationDependencies(node, true) {
		if dependency.Accessibility != printer.SymbolAccessibilityAccessible {
			tx.typefnError(dependency.ErrorNode, "declaration recipe dependency cannot be named: "+dependency.ErrorSymbolName)
			valid = false
			continue
		}
		tx.tracker.handleSymbolAccessibilityError(dependency.SymbolAccessibilityResult)
		if failed, message := tx.typefnDependencyFailure(dependency.Declarations); failed != nil {
			tx.typefnError(dependency.Reference, message)
			diagnostic := tx.state.diagnostics[len(tx.state.diagnostics)-1]
			if ast.GetSourceFileOfNode(failed) != ast.GetSourceFileOfNode(dependency.Reference) {
				diagnostic.AddRelatedInfo(createDiagnosticForNode(failed, diagnostics.NewAdHocMessage("typefn: dependency declared here")))
			}
			valid = false
		}
	}
	return valid
}

// Dependency cycles are legal. Visit each declaration once per root and validate
// the whole lexical closure, without importing another file's visibility worklist.
func (tx *DeclarationTransformer) typefnDependencyFailure(roots []*ast.Node) (*ast.Node, string) {
	queue := append([]*ast.Node(nil), roots...)
	seen := map[*ast.Node]bool{}
	for index := 0; index < len(queue); index++ {
		declaration := queue[index]
		queue[index] = nil
		if seen[declaration] {
			continue
		}
		seen[declaration] = true
		source := ast.GetSourceFileOfNode(declaration)
		if source.IsDeclarationFile || !tx.host.SourceFileMayBeEmitted(source, false) {
			continue
		}
		if tx.state.stripInternal {
			for parent := declaration; parent != nil && !ast.IsSourceFile(parent); parent = parent.Parent {
				if tx.isInternalDeclaration(parent, source) {
					name := declaration.Name()
					if name == nil {
						name = declaration
					}
					return declaration, "stripped declaration dependency: " + scanner.GetTextOfNode(name)
				}
			}
		}
		if _, valid := tx.resolver.TypefnDeclarationStatus(declaration); !valid {
			return declaration, "cannot emit an invalid declaration recipe dependency"
		}
		for _, dependency := range tx.resolver.GetTypefnDeclarationDependencies(declaration, false) {
			if dependency.Accessibility != printer.SymbolAccessibilityAccessible {
				return dependency.Reference, "declaration recipe dependency cannot be named: " + dependency.ErrorSymbolName
			}
			queue = append(queue, dependency.Declarations...)
		}
	}
	return nil, ""
}

func (tx *DeclarationTransformer) transformTypeFunctionDeclaration(node *ast.Node) *ast.Node {
	if !tx.checkTypefnDeclaration(node) {
		// Recovered syntax is useful to editor services, never to recipe emission.
		return node
	}
	// A recipe is executable type syntax. Visiting its body as native declaration
	// syntax would erase control flow, mutable bindings and frozen closure creation.
	tx.needsDeclare = false
	d := node.AsTypeFunctionDeclaration()
	return tx.Factory().UpdateTypeFunctionDeclaration(d, tx.ensureModifiers(node), d.Name(), d.TypeParameters, d.ResultKind, d.Body)
}
