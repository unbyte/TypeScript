package checker

import (
	"fmt"
	"slices"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/binder"
)

// Native Types are created only after the complete reachable scratch graph is
// validated. Generated recursive aliases use explicit leaf environments, so native
// instantiation sees every symbolic dependency (including lazy member Types).
// No native fake type parameter stands in for an unsealed Self handle.
func (v *typefnVM) sealConstruction(node *ast.Node, arena *typefnConstructionArena, roots []*Type, group bool) *Type {
	edges := make([][]typefnGraphEdge, len(arena.nodes))
	for index, t := range arena.nodes {
		if !v.step(node) {
			return v.c.errorType
		}
		graph := constructionOf(t)
		inputs := graph.inputs
		if graph.kind == ast.KindTypeReference {
			inputs = []*Type{arena.definitions[graph.reference]}
		}
		guarded := graph.kind == ast.KindTypeLiteral || graph.kind == ast.KindTupleType || graph.kind == ast.KindArrayType || graph.kind == ast.KindFunctionType || graph.kind == ast.KindConstructorType
		for i, input := range inputs {
			if input == nil {
				return v.error(node, "unsealed recursive definition")
			}
			if next := constructionOf(input); next != nil {
				if next.arena != arena {
					return v.error(node, "recursive reference belongs to another construction")
				}
				edgeGuarded := guarded
				if graph.kind == ast.KindTypeLiteral && i >= len(graph.properties)+len(graph.indexers) {
					// Forwarding a signature handle adds no structural guard of its own.
					edgeGuarded = false
				}
				if graph.kind == ast.KindTupleType && graph.elements[i].flags == ElementFlagsVariadic {
					edgeGuarded = false
				}
				edges[index] = append(edges[index], typefnGraphEdge{next.index, edgeGuarded})
			}
		}
	}
	indices := []int{}
	for _, root := range append(slices.Clone(roots), arena.checks...) {
		if graph := constructionOf(root); graph != nil {
			indices = append(indices, graph.index)
		}
	}
	validation := []int{}
	for _, root := range append(slices.Clone(arena.validationRoots), arena.checks...) {
		if graph := constructionOf(root); graph != nil {
			validation = append(validation, graph.index)
		}
	}
	status, reachable := validateTypefnGraph(edges, validation, func() bool { return v.step(node) })
	if status == typefnGraphCanceled {
		return v.c.errorType
	}
	if status != typefnGraphValid {
		return v.error(node, "recursive cycle must cross a structural edge")
	}
	// An ordinary result that ignores Self retains its original native identity.
	if len(indices) == 0 {
		if group {
			return v.c.createTupleType(roots)
		}
		return roots[0]
	}
	// Discarded valid graphs must not add phantom leaf parameters to the
	// returned native Types. Materialize output shapes and reached checks; each
	// dependency component keeps its own leaf environment.
	if len(arena.validationRoots) != len(roots) || len(arena.checks) != 0 {
		status, reachable = reachableTypefnGraph(edges, indices, func() bool { return v.step(node) })
		if status != typefnGraphValid {
			return v.c.errorType
		}
	}
	plan := v.constructionEnvironment(node, arena, edges, reachable)
	defer func() { v.liveBytes -= plan.bytes }()
	if v.failed {
		return v.c.errorType
	}
	// One native tuple parameter carries each component environment. Recursive
	// edges pass it unchanged; DAG edges project only the child's leaf slots.
	for _, index := range reachable {
		graph := constructionOf(arena.nodes[index])
		inputs := graph.inputs
		if graph.kind == ast.KindTypeReference {
			inputs = []*Type{arena.definitions[graph.reference]}
		}
		if !v.allocate(node, 16+12*len(inputs)) {
			return v.c.errorType
		}
		for _, input := range inputs {
			if child := constructionOf(input); child != nil && !slices.Equal(plan.parameters[index], plan.parameters[child.index]) {
				if !v.allocate(node, 5*len(plan.parameters[child.index])) {
					return v.c.errorType
				}
			}
		}
	}
	f := ast.NewNodeFactory(ast.NodeFactoryHooks{OnCreate: func(created *ast.Node) { created.Loc = node.Loc }})
	aliasName := func(index int) string { return fmt.Sprintf("__TypefnRecursive%d", index) }
	const environmentName = "__TypefnEnvironment"
	environmentReference := func() *ast.Node { return f.NewTypeReferenceNode(f.NewIdentifier(environmentName), nil) }
	parameters := func(index int) *ast.NodeList {
		if len(plan.parameters[index]) == 0 {
			return nil
		}
		bound := f.NewTypeOperatorNode(ast.KindReadonlyKeyword, f.NewArrayTypeNode(f.NewKeywordTypeNode(ast.KindUnknownKeyword)))
		return f.NewNodeList([]*ast.Node{f.NewTypeParameterDeclaration(nil, f.NewIdentifier(environmentName), bound, nil, nil)})
	}
	leafReference := func(index, id int) *ast.Node {
		offset, found := slices.BinarySearch(plan.parameters[index], id)
		if !found {
			panic("missing recursive environment dependency")
		}
		return f.NewIndexedAccessTypeNode(environmentReference(), f.NewLiteralTypeNode(f.NewNumericLiteral(fmt.Sprint(offset), ast.TokenFlagsNone)))
	}
	reference := func(index int, t *Type) *ast.Node {
		if graph := constructionOf(t); graph != nil {
			var arguments *ast.NodeList
			child := plan.parameters[graph.index]
			if len(child) != 0 {
				environment := environmentReference()
				if !slices.Equal(plan.parameters[index], child) {
					elements := make([]*ast.Node, len(child))
					for i, id := range child {
						elements[i] = leafReference(index, id)
					}
					environment = f.NewTupleTypeNode(f.NewNodeList(elements))
				}
				arguments = f.NewNodeList([]*ast.Node{environment})
			}
			return f.NewTypeReferenceNode(f.NewIdentifier(aliasName(graph.index)), arguments)
		}
		return leafReference(index, plan.leafIndex[t])
	}
	statements := []*ast.Node{}
	domainBindings := map[*Type]*ast.Node{}
	domainReference := func(key *Type) *ast.Node {
		binding := domainBindings[key]
		if binding == nil {
			if !v.allocate(node, 4) {
				return f.NewKeywordTypeNode(ast.KindUnknownKeyword)
			}
			name := fmt.Sprintf("__TypefnDomain%d", len(domainBindings))
			binding = f.NewTypeAliasDeclaration(nil, f.NewIdentifier(name), nil, f.NewKeywordTypeNode(ast.KindUnknownKeyword))
			domainBindings[key] = binding
			statements = append(statements, binding)
		}
		return f.NewTypeReferenceNode(f.NewIdentifier(binding.Name().Text()), nil)
	}
	keyBindings := map[*Type]*ast.Node{}
	keyName := func(key *Type) *ast.Node {
		if binding := keyBindings[key]; binding != nil {
			return f.NewIdentifier(binding.Name().Text())
		}
		if !v.allocate(node, 6) {
			return f.NewIdentifier("__typefnError")
		}
		name := fmt.Sprintf("__TypefnKey%d", len(keyBindings))
		declaration := f.NewVariableDeclaration(f.NewIdentifier(name), nil, f.NewKeywordTypeNode(ast.KindUnknownKeyword), nil)
		statement := f.NewVariableStatement(f.NewModifierList([]*ast.Node{f.NewToken(ast.KindDeclareKeyword)}), f.NewVariableDeclarationList(f.NewNodeList([]*ast.Node{declaration}), ast.NodeFlagsConst))
		keyBindings[key] = declaration
		statements = append(statements, statement)
		return f.NewIdentifier(name)
	}
	aliases := map[int]*ast.Node{}
	membersPlans := map[*ast.Node]*typefnConstructionMembers{}
	for _, index := range reachable {
		if !v.step(node) {
			return v.c.errorType
		}
		graph := constructionOf(arena.nodes[index])
		inputs := make([]*ast.Node, len(graph.inputs))
		for i, input := range graph.inputs {
			inputs[i] = reference(index, input)
		}
		var body *ast.Node
		switch graph.kind {
		case ast.KindTypeReference:
			body = reference(index, arena.definitions[graph.reference])
		case ast.KindBarToken:
			body = f.NewUnionTypeNode(f.NewNodeList(inputs))
		case ast.KindAmpersandToken:
			body = f.NewIntersectionTypeNode(f.NewNodeList(inputs))
		case ast.KindTupleType:
			for i, info := range graph.elements {
				var optional, rest *ast.Node
				if info.flags == ElementFlagsOptional {
					optional = f.NewToken(ast.KindQuestionToken)
				}
				if info.flags&ElementFlagsVariable != 0 {
					rest = f.NewToken(ast.KindDotDotDotToken)
				}
				if info.flags == ElementFlagsRest {
					inputs[i] = f.NewArrayTypeNode(inputs[i])
				}
				if info.labeledDeclaration != nil {
					inputs[i] = f.NewNamedTupleMember(rest, f.NewIdentifier(info.labeledDeclaration.Name().Text()), optional, inputs[i])
				} else {
					if optional != nil {
						inputs[i] = f.NewOptionalTypeNode(inputs[i])
					}
					if rest != nil {
						inputs[i] = f.NewRestTypeNode(inputs[i])
					}
				}
			}
			body = f.NewTupleTypeNode(f.NewNodeList(inputs))
		case ast.KindFunctionType, ast.KindConstructorType:
			parameters := f.NewNodeList([]*ast.Node{f.NewParameterDeclaration(nil, f.NewToken(ast.KindDotDotDotToken), f.NewIdentifier("args"), nil, inputs[0], nil)})
			if graph.kind == ast.KindFunctionType {
				body = f.NewFunctionTypeNode(nil, parameters, inputs[1])
			} else {
				var modifiers *ast.ModifierList
				if graph.abstract {
					modifiers = f.NewModifierList([]*ast.Node{f.NewToken(ast.KindAbstractKeyword)})
				}
				body = f.NewConstructorTypeNode(modifiers, nil, parameters, inputs[1])
			}
		case ast.KindArrayType:
			body = f.NewArrayTypeNode(inputs[0])
		case ast.KindTypeLiteral:
			members := make([]*ast.Node, len(graph.properties))
			for i, p := range graph.properties {
				name := f.NewStringLiteral(p.name, ast.TokenFlagsNone)
				if p.key != nil {
					name = f.NewComputedPropertyName(keyName(p.key))
					if v.failed {
						return v.c.errorType
					}
				}
				var modifiers *ast.ModifierList
				if p.readonly {
					modifiers = f.NewModifierList([]*ast.Node{f.NewToken(ast.KindReadonlyKeyword)})
				}
				var optional *ast.Node
				if p.optional {
					optional = f.NewToken(ast.KindQuestionToken)
				}
				members[i] = f.NewPropertySignatureDeclaration(modifiers, name, optional, inputs[i], nil)
			}
			for i, info := range graph.indexers {
				var modifiers *ast.ModifierList
				if info.readonly {
					modifiers = f.NewModifierList([]*ast.Node{f.NewToken(ast.KindReadonlyKeyword)})
				}
				parameter := f.NewParameterDeclaration(nil, nil, f.NewIdentifier("key"), nil, domainReference(info.key), nil)
				members = append(members, f.NewIndexSignatureDeclaration(modifiers, f.NewNodeList([]*ast.Node{parameter}), inputs[len(graph.properties)+i]))
			}
			if v.failed {
				return v.c.errorType
			}
			body = f.NewTypeLiteralNode(f.NewNodeList(members))
			if graph.calls+graph.constructs != 0 {
				start := len(graph.properties) + len(graph.indexers)
				sources := inputs[start:]
				for _, source := range sources {
					ast.SetParentInChildren(source)
					source.Parent = body
				}
				membersPlans[body] = &typefnConstructionMembers{calls: sources[:graph.calls], constructs: sources[graph.calls:]}
			}
		default:
			panic("unknown recursive construction shape")
		}
		if graph.readonly {
			body = f.NewTypeOperatorNode(ast.KindReadonlyKeyword, body)
		}
		alias := f.NewTypeAliasDeclaration(nil, f.NewIdentifier(aliasName(index)), parameters(index), body)
		aliases[index] = alias
		statements = append(statements, alias)
	}
	source := ast.GetSourceFileOfNode(node)
	scope := f.NewModuleDeclaration(nil, ast.KindNamespaceKeyword, f.NewIdentifier("__TypefnConstruction"), nil, f.NewModuleBlock(f.NewNodeList(statements)))
	file := f.NewSourceFile(source.ParseOptions(), source.Text(), f.NewNodeList([]*ast.Node{scope}), f.NewToken(ast.KindEndOfFile)).AsSourceFile()
	file.ScriptKind = source.ScriptKind
	file.Loc = source.Loc
	ast.SetParentInChildren(file.AsNode())
	binder.BindSourceFile(file)
	scope.Parent = source.AsNode()
	if len(membersPlans) != 0 {
		state := v.c.getTypefnState()
		if state.constructionMembers == nil {
			state.constructionMembers = map[*ast.Symbol]*typefnConstructionMembers{}
		}
		for body, members := range membersPlans {
			state.constructionMembers[body.Symbol()] = members
		}
	}
	for key, binding := range domainBindings {
		v.c.typeAliasLinks.Get(binding.Symbol()).declaredType = key
	}
	for key, binding := range keyBindings {
		v.c.valueSymbolLinks.Get(binding.Symbol()).resolvedType = key
	}
	// All edges are now native syntax and all leaf arguments are ordinary Types.
	// Release the private construction context before native evaluation/replay.
	v.construction = nil
	savedSite := v.nativeIslandSite
	v.nativeIslandSite = node
	defer func() { v.nativeIslandSite = savedSite }()
	environments := map[int]*Type{}
	resolve := func(root *Type) *Type {
		var result *Type
		if !v.step(node) {
			return v.c.errorType
		}
		if graph := constructionOf(root); graph != nil {
			declared := v.c.getDeclaredTypeOfSymbol(aliases[graph.index].Symbol())
			if v.c.isErrorType(declared) {
				v.failed = true
				return v.c.errorType
			}
			if len(plan.parameters[graph.index]) == 0 {
				result = declared
			} else {
				owner := plan.owners[graph.index]
				environment := environments[owner]
				if environment == nil {
					parameters := plan.parameters[graph.index]
					if !plan.reserve(v, node, 16*len(parameters)) || !v.allocate(node, len(parameters)+1) {
						return v.c.errorType
					}
					values := make([]*Type, len(parameters))
					for i, id := range parameters {
						values[i] = plan.leaves[id]
					}
					environment = v.c.createTupleType(values)
					environments[owner] = environment
				}
				result = v.c.getTypeAliasInstantiation(aliases[graph.index].Symbol(), []*Type{environment}, nil)
			}
		} else {
			result = root
		}
		if v.c.isErrorType(result) {
			v.failed = true
			return v.c.errorType
		}
		return result
	}
	result := make([]*Type, len(roots))
	for i, root := range roots {
		result[i] = resolve(root)
		if v.failed {
			return v.c.errorType
		}
	}
	// Reached construction checks survive even when their result was discarded.
	// Each component has its own environment, so these checks cannot introduce
	// phantom type parameters into the published result.
	for _, checked := range arena.checks {
		value := resolve(checked)
		if v.failed {
			return v.c.errorType
		}
		v.validateObjectIndexers(node, value, v.c.getIndexInfosOfType(value))
		if v.failed || v.blocked {
			return v.c.errorType
		}
	}
	if group {
		return v.c.createTupleType(result)
	}
	return result[0]
}
