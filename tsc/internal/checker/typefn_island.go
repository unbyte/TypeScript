package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/binder"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/scanner"
)

type typefnIsland struct {
	names []string
	nodes int
}

// Resolve captures against the original native lexical scopes. Native mapped,
// infer and signature binders shadow DSL bindings using their actual symbols.
func (b *typefnCompiler) island(e *typefnExpression) {
	plan := &typefnIsland{}
	e.island = plan
	captured := map[*ast.Symbol]bool{}
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		plan.nodes++
		if ast.IsTypeQueryNode(node) {
			name := node.AsTypeQueryNode().ExprName
			if !ast.IsThisIdentifier(name) {
				symbol := b.c.resolveEntityName(name, ast.SymbolFlagsValue, false, false, nil)
				if symbol == nil || symbol == b.c.unknownSymbol {
					b.p.valid = false
				}
			}
		}
		if ast.IsTypeReferenceNode(node) {
			name := node.AsTypeReferenceNode().TypeName
			symbol := b.c.resolveEntityName(name, ast.SymbolFlagsType, false, false, nil)
			if symbol == nil || symbol == b.c.unknownSymbol {
				b.p.valid = false
				return false
			}
			parameters := b.c.getLocalTypeParametersOfClassOrInterfaceOrTypeAlias(symbol)
			count, minimum := len(parameters), b.c.getMinTypeArgumentCount(parameters)
			if declaration := core.Find(symbol.Declarations(), ast.IsTypeFunctionDeclaration); declaration != nil {
				count, minimum = len(declaration.TypeParameters()), len(declaration.TypeParameters())
				if node.TypeArgumentList() == nil {
					b.error(node, "typefn requires an explicit application in a native island")
				}
			}
			actual := len(node.TypeArguments())
			if actual < minimum || actual > count {
				b.error(node, fmt.Sprintf("native application requires %d..%d Type arguments", minimum, count))
			}
			if ast.IsIdentifier(name) {
				if binding, ok := b.lookup(name.Text()); ok && binding.symbol == symbol {
					if binding.slot < 0 {
						b.error(name, "binding read before initialization")
						return false
					}
					if node.TypeArgumentList() != nil {
						b.error(node, "a captured Type is not a constructor")
						return false
					}
					if !captured[symbol] {
						captured[symbol] = true
						plan.names = append(plan.names, name.Text())
						e.args = append(e.args, &typefnExpression{node: name, slot: binding.slot})
					}
				}
			}
		}
		node.ForEachChild(visit)
		return false
	}
	visit(e.node.AsTypeFunctionExpression().NativeType)
}

// Each native AST is bound in an immutable lexical snapshot. Captured values are
// real native alias symbols, not text substitutions or fake deferred type
// parameters. No source AST cache is mutated or reused across environments.
func (v *typefnVM) nativeIsland(e *typefnExpression, slots []*Type) *Type {
	values := make([]*Type, len(e.args))
	for i, argument := range e.args {
		values[i] = v.expression(argument, slots)
		if v.failed {
			return v.c.errorType
		}
		// Until native cache contexts can retain symbolic capture dependencies,
		// conservatively keep the original island and frozen locals for replay.
		if v.unresolvedGraph(e.node, values[i]) {
			if v.failed {
				return v.c.errorType
			}
			return v.deferExpression(e, slots)
		}
	}
	plan := e.island
	if !v.allocate(e.node, plan.nodes+len(values)*4+7) {
		return v.c.errorType
	}
	source := ast.GetSourceFileOfNode(e.node)
	factory := ast.NewNodeFactory(ast.NodeFactoryHooks{})
	body := factory.DeepCloneReparse(e.node.AsTypeFunctionExpression().NativeType)
	statements := make([]*ast.Node, 0, len(values)+1)
	for _, name := range plan.names {
		alias := factory.NewTypeAliasDeclaration(nil, factory.NewIdentifier(name), nil, factory.NewKeywordTypeNode(ast.KindUnknownKeyword))
		alias.Loc = e.node.Loc
		alias.Name().Loc = e.node.Loc
		statements = append(statements, alias)
	}
	result := factory.NewTypeAliasDeclaration(nil, factory.NewIdentifier("(typefn result)"), nil, body)
	result.Loc, result.Name().Loc = e.node.Loc, e.node.Loc
	statements = append(statements, result)
	block := factory.NewModuleBlock(factory.NewNodeList(statements))
	scope := factory.NewModuleDeclaration(nil, ast.KindNamespaceKeyword, factory.NewIdentifier("(typefn captures)"), nil, block)
	scope.Loc, scope.Name().Loc, block.Loc = e.node.Loc, e.node.Loc, e.node.Loc
	file := factory.NewSourceFile(source.ParseOptions(), source.Text(), factory.NewNodeList([]*ast.Node{scope}), factory.NewToken(ast.KindEndOfFile)).AsSourceFile()
	file.ScriptKind = source.ScriptKind
	file.Loc = source.Loc
	ast.SetParentInChildren(file.AsNode())
	binder.BindSourceFile(file)
	// Only this detached subtree changes parents. Its outward lexical lookup uses
	// the original source scope and source spans after its local binders are installed.
	scope.Parent = e.node
	for i, value := range values {
		v.c.typeAliasLinks.Get(statements[i].Symbol()).declaredType = value
	}
	saved := v.nativeIslandSite
	v.nativeIslandSite = e.node
	defer func() { v.nativeIslandSite = saved }()
	for _, diagnostic := range file.BindDiagnostics() {
		v.c.addDiagnostic(diagnostic)
	}
	v.c.checkSourceElement(body)
	if v.failed {
		return v.c.errorType
	}
	return v.c.getTypeFromTypeNode(body)
}

func (v *typefnVM) islandDiagnostic(diagnostic *ast.Diagnostic) *ast.Diagnostic {
	v.failed = true
	site := core.Coalesce(v.replaySite, core.Coalesce(v.rootSite, v.nativeIslandSite))
	source := ast.GetSourceFileOfNode(site)
	loc := scanner.GetErrorRangeForNode(source, site)
	if diagnostic.File() == source && diagnostic.Loc() == loc {
		return diagnostic
	}
	result := diagnostic.Clone()
	result.SetFile(source)
	result.SetLocation(loc)
	result.AddRelatedInfo(diagnostic)
	return result
}
