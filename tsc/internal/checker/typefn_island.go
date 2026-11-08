package checker

import (
	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/binder"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/scanner"
)

type typefnIslandCapture struct {
	name         string
	helper       *typefnProgram
	nativeDirect bool
	value        bool
	context      bool
}
type typefnIsland struct {
	captures []typefnIslandCapture
	contexts map[*ast.Node]int
	nodes    int
}

// Each native AST is bound in an immutable lexical snapshot. Captured values are
// real native alias symbols, not text substitutions or fake deferred type
// parameters. No source AST cache is mutated or reused across environments.
func (v *typefnVM) nativeIsland(e *typefnExpression, slots []typefnValue) *Type {
	values := make([]typefnValue, len(e.args))
	for i, argument := range e.args {
		values[i] = v.resultExpression(argument, slots)
		if v.failed {
			return v.c.errorType
		}
		if values[i].native != nil && v.rejectConstruction(e.node, values[i].native) {
			return v.c.errorType
		}
		// Until native cache contexts can retain symbolic capture dependencies,
		// conservatively keep the original island and frozen locals for replay.
		if v.unresolvedValue(e.node, values[i]) {
			if v.failed {
				return v.c.errorType
			}
			return v.deferExpression(e, slots)
		}
	}
	plan := e.island
	captureNodes := 0
	for _, capture := range plan.captures {
		if capture.context && !capture.value {
			continue
		}
		if capture.value {
			captureNodes += 6
		} else {
			captureNodes += 4
		}
	}
	if !v.allocate(e.node, plan.nodes+captureNodes+7) {
		return v.c.errorType
	}
	source := ast.GetSourceFileOfNode(e.node)
	factory := ast.NewNodeFactory(ast.NodeFactoryHooks{OnClone: func(clone, original *ast.Node) {
		if i, ok := plan.contexts[original]; ok {
			if plan.captures[i].value {
				// Only the cloned this-root is renamed; member lookup and native query
				// type arguments still run in the frozen capture scope.
				clone.AsIdentifier().Text = plan.captures[i].name
			} else {
				v.c.typeNodeLinks.Get(clone).resolvedType = values[i].native
			}
		}
	}})
	body := factory.DeepCloneReparse(e.node.AsTypeFunctionExpression().NativeType)
	statements := make([]*ast.Node, 0, len(values)+1)
	declarations := make([]*ast.Node, len(values))
	for i, capture := range plan.captures {
		if capture.context && !capture.value {
			continue
		}
		name := factory.NewIdentifier(capture.name)
		name.Loc = e.node.Loc
		var declaration, statement *ast.Node
		if capture.value {
			declaration = factory.NewVariableDeclaration(name, nil, factory.NewKeywordTypeNode(ast.KindUnknownKeyword), nil)
			statement = factory.NewVariableStatement(factory.NewModifierList([]*ast.Node{factory.NewToken(ast.KindDeclareKeyword)}), factory.NewVariableDeclarationList(factory.NewNodeList([]*ast.Node{declaration}), ast.NodeFlagsConst))
		} else {
			declaration = factory.NewTypeAliasDeclaration(nil, name, nil, factory.NewKeywordTypeNode(ast.KindUnknownKeyword))
			statement = declaration
		}
		declaration.Loc = e.node.Loc
		statement.Loc = e.node.Loc
		declarations[i] = declaration
		statements = append(statements, statement)
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
		capture := plan.captures[i]
		if capture.context && !capture.value {
			continue
		}
		symbol := declarations[i].Symbol()
		if capture.value {
			v.c.valueSymbolLinks.Get(symbol).resolvedType = value.native
		} else if value.native != nil {
			v.c.typeAliasLinks.Get(symbol).declaredType = value.native
		} else {
			state := v.c.getTypefnState()
			if state.aliases == nil {
				state.aliases = map[*ast.Symbol]*typefnAlias{}
			}
			state.aliases[symbol] = &typefnAlias{declaration: declarations[i], kind: e.args[i].kind, value: value.closure, helper: capture.helper, nativeDirect: capture.nativeDirect}
		}
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
