package checker

import (
	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/printer"
)

// GetTypefnDeclarationDependencies resolves dependencies of a retained source recipe
// using native declaration visibility. Walk the lexical syntax, including dead
// branches and nested helpers; executing the recipe cannot discover this closure.
// Local DSL and native island binders are resolved in their original scopes.
// Transitive validation passes makeVisible=false: it must not schedule declarations
// from a dependency module for emission into the current module.
func (r *EmitResolver) GetTypefnDeclarationDependencies(node *ast.Node, makeVisible bool) []printer.TypefnDeclarationDependency {
	r.checkerMu.Lock()
	defer r.checkerMu.Unlock()
	return r.checker.typefnDeclarationDependencies(node, makeVisible)
}

func (c *Checker) typefnDeclarationDependencies(node *ast.Node, makeVisible bool) []printer.TypefnDeclarationDependency {
	// Bundled native library declarations are terminal dependencies. User
	// augmentations have separate declarations and are still traversed.
	if c.program.IsSourceFileDefaultLibrary(ast.GetSourceFileOfNode(node).PathKey()) {
		return nil
	}
	var results []printer.TypefnDeclarationDependency
	type referenceKey struct{ root, target *ast.Symbol }
	seen := map[referenceKey]bool{}
	var visit func(*ast.Node) bool
	visit = func(n *ast.Node) bool {
		// An explicitly annotated native declaration exposes its written Type.
		// Its implementation/default initializer is not part of that surface.
		// DSL bodies have different node kinds and are always retained in full.
		if parent := n.Parent; parent != nil && parent.Type() != nil &&
			((ast.IsVariableDeclaration(parent) || ast.IsPropertyDeclaration(parent) || ast.IsParameterDeclaration(parent)) && parent.Initializer() == n ||
				ast.IsFunctionLike(parent) && parent.Body() == n) {
			return false
		}
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
		case ast.KindExpressionWithTypeArguments:
			if ast.IsEntityName(n.Expression()) || ast.IsEntityNameExpression(n.Expression()) {
				name = n.Expression()
			}
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
							visible = c.hasVisibleDeclarations(root, makeVisible)
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

// GetTypefnReferenceDependencies keeps ordinary native aliases/annotations in the
// recipe failure boundary, without changing native-only declaration emit policy.
func (r *EmitResolver) GetTypefnReferenceDependencies(node *ast.Node) []*ast.Node {
	r.checkerMu.Lock()
	defer r.checkerMu.Unlock()
	c := r.checker
	if c.compilerOptions.GetTypefnVersion() != 1 || !ast.IsTypeReferenceNode(node) || !ast.IsParseTreeNode(node) {
		return nil
	}
	symbol := c.resolveEntityName(node.AsTypeReferenceNode().TypeName, ast.SymbolFlagsType, true, false, nil)
	if symbol == nil || symbol == c.unknownSymbol || symbol.Flags()&ast.SymbolFlagsTypeParameter != 0 {
		return nil
	}
	roots := symbol.Declarations()
	if c.typefnDeclarationsHaveRecipe(roots) {
		return roots
	}
	return nil
}

// Reachability is syntax/symbol based, not inferred from an evaluated result:
// a completed literal Type may still have an executable source recipe. Cache
// negatives only after closing the entire reachable graph, so cycles cannot hide
// a positive edge discovered later. Positive results propagate along a proven path.
func (c *Checker) typefnDeclarationsHaveRecipe(roots []*ast.Node) bool {
	if len(roots) == 0 {
		return false
	}
	state := c.getTypefnState()
	if state.emitRecipeReachability == nil {
		state.emitRecipeReachability = map[*ast.Node]bool{}
	}
	known := state.emitRecipeReachability
	allKnown := true
	for _, root := range roots {
		value, cached := known[root]
		if cached && value {
			return true
		}
		allKnown = allKnown && cached
	}
	if allKnown {
		return false
	}
	queue := make([]*ast.Node, 0, len(roots))
	parents := map[*ast.Node]*ast.Node{}
	for _, root := range roots {
		if _, seen := parents[root]; !seen {
			parents[root] = nil
			queue = append(queue, root)
		}
	}
	for index := 0; index < len(queue); index++ {
		declaration := queue[index]
		value, cached := known[declaration]
		if ast.IsTypeFunctionDeclaration(declaration) || cached && value {
			for current := declaration; current != nil; current = parents[current] {
				known[current] = true
			}
			return true
		}
		if cached {
			continue
		}
		// The bundled native libraries contain no executable DSL. Merged user
		// augmentations have their own declarations and remain in the worklist.
		if c.program.IsSourceFileDefaultLibrary(ast.GetSourceFileOfNode(declaration).PathKey()) {
			continue
		}
		for _, dependency := range c.typefnDeclarationDependencies(declaration, false) {
			for _, target := range dependency.Declarations {
				if _, seen := parents[target]; !seen {
					parents[target] = declaration
					queue = append(queue, target)
				}
			}
		}
	}
	for declaration := range parents {
		known[declaration] = false
	}
	return false
}
