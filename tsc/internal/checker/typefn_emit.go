package checker

import (
	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/printer"
)

// GetTypefnDeclarationDependencies marks dependencies of a retained source recipe
// using native declaration visibility. Walk the lexical syntax, including dead
// branches and nested helpers; executing the recipe cannot discover this closure.
// Local DSL and native island binders are resolved in their original scopes.
func (r *EmitResolver) GetTypefnDeclarationDependencies(node *ast.Node) []printer.TypefnDeclarationDependency {
	r.checkerMu.Lock()
	defer r.checkerMu.Unlock()
	c := r.checker
	var results []printer.TypefnDeclarationDependency
	type referenceKey struct{ root, target *ast.Symbol }
	seen := map[referenceKey]bool{}
	var visit func(*ast.Node) bool
	visit = func(n *ast.Node) bool {
		var name *ast.Node
		meaning := ast.SymbolFlagsType
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
			meaning = ast.SymbolFlagsValue | ast.SymbolFlagsExportValue
		case ast.KindComputedPropertyName:
			if ast.IsEntityNameExpression(n.Expression()) {
				name = n.Expression()
			}
			meaning = ast.SymbolFlagsValue | ast.SymbolFlagsExportValue
		}
		if name != nil {
			first := ast.GetFirstIdentifier(name)
			root := c.resolveName(first, first.Text(), getMeaningOfEntityNameReference(name), nil, false, false)
			target := c.resolveEntityName(name, meaning, true, false, nil)
			key := referenceKey{root, target}
			if !seen[key] {
				seen[key] = true
				local := false
				if root != nil {
					for _, declaration := range root.Declarations() {
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
				}
				if !local {
					dependency := printer.TypefnDeclarationDependency{Reference: name}
					dependency.ErrorNode, dependency.ErrorSymbolName = first, first.Text()
					if root == nil || target == nil || target == c.unknownSymbol {
						dependency.Accessibility = printer.SymbolAccessibilityNotResolved
					} else {
						dependency.Declarations = target.Declarations()
						// Native visibility treats all Type parameters as visible. An outer
						// parameter is a free recipe capture and needs explicit lifting instead.
						var visible *printer.SymbolAccessibilityResult
						if root.Flags()&ast.SymbolFlagsTypeParameter == 0 {
							visible = c.hasVisibleDeclarations(root, true)
						}
						if visible != nil {
							dependency.SymbolAccessibilityResult = *visible
						} else {
							dependency.Accessibility = printer.SymbolAccessibilityNotAccessible
						}
					}
					results = append(results, dependency)
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
