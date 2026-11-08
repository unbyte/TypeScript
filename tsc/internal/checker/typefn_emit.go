package checker

import (
	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/printer"
)

// GetTypefnDeclarationDependencies marks dependencies of a retained source recipe
// using native declaration visibility. Walk the lexical syntax, including dead
// branches and nested helpers; executing the recipe cannot discover this closure.
// Local DSL and native island binders are resolved in their original scopes.
func (r *EmitResolver) GetTypefnDeclarationDependencies(node *ast.Node) []printer.SymbolAccessibilityResult {
	r.checkerMu.Lock()
	defer r.checkerMu.Unlock()
	c := r.checker
	var results []printer.SymbolAccessibilityResult
	seen := map[*ast.Symbol]bool{}
	var visit func(*ast.Node) bool
	visit = func(n *ast.Node) bool {
		var name *ast.Node
		switch n.Kind {
		case ast.KindTypeFunctionName, ast.KindTypeFunctionCall, ast.KindTypeFunctionAssignment:
			name = n.Name()
			if typefnReserved(ast.GetFirstIdentifier(name).Text()) {
				name = nil
			}
		case ast.KindTypeReference:
			name = n.AsTypeReferenceNode().TypeName
		case ast.KindTypeQuery:
			name = n.AsTypeQueryNode().ExprName
		case ast.KindComputedPropertyName:
			if ast.IsEntityNameExpression(n.Expression()) {
				name = n.Expression()
			}
		}
		if name != nil {
			first := ast.GetFirstIdentifier(name)
			symbol := c.resolveName(first, first.Text(), getMeaningOfEntityNameReference(name), nil, false, false)
			if symbol == nil {
				results = append(results, printer.SymbolAccessibilityResult{Accessibility: printer.SymbolAccessibilityNotResolved, ErrorSymbolName: first.Text(), ErrorNode: first})
			} else if !seen[symbol] {
				seen[symbol] = true
				local := false
				for _, declaration := range symbol.Declarations() {
					for parent := declaration; parent != nil; parent = parent.Parent {
						if parent == node {
							local = true
							break
						}
					}
					if local {
						break
					}
				}
				if !local {
					// Native declaration emit normally considers every Type parameter
					// visible. An outer parameter is a free capture here: moving
					// this recipe requires explicit lifting, not that shortcut.
					var visible *printer.SymbolAccessibilityResult
					if symbol.Flags()&ast.SymbolFlagsTypeParameter == 0 {
						visible = c.hasVisibleDeclarations(symbol, true)
					}
					if visible != nil {
						results = append(results, *visible)
					} else {
						results = append(results, printer.SymbolAccessibilityResult{Accessibility: printer.SymbolAccessibilityNotAccessible, ErrorSymbolName: first.Text(), ErrorNode: first})
					}
				}
			}
		}
		return n.ForEachChild(visit)
	}
	visit(node)
	return results
}

// TypefnDeclarationStatus validates a retained definition or constructor alias.
// Ordinary aliases keep their native source syntax and do not require an ABI.
func (r *EmitResolver) TypefnDeclarationStatus(node *ast.Node) (needsABI, valid bool) {
	r.checkerMu.Lock()
	defer r.checkerMu.Unlock()
	c := r.checker
	if c.compilerOptions.GetTypefnVersion() != 1 {
		return false, true
	}
	if ast.IsTypeFunctionDeclaration(node) {
		return true, c.compileTypefn(node).valid
	}
	if ast.IsTypeAliasDeclaration(node) {
		if alias := c.typefnConstructorAlias(node.Symbol()); alias != nil {
			result := c.runTypefn(node, func(v *typefnVM) *Type {
				v.constructorAlias(alias)
				if v.failed {
					return c.errorType
				}
				return c.unknownType
			})
			return true, result != c.errorType
		}
	}
	return false, true
}
