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
				if graph.kind == ast.KindTupleType && graph.elements[i].flags == ElementFlagsVariadic {
					edgeGuarded = false
				}
				edges[index] = append(edges[index], typefnGraphEdge{next.index, edgeGuarded})
			}
		}
	}
	indices := []int{}
	for _, root := range roots {
		if graph := constructionOf(root); graph != nil {
			indices = append(indices, graph.index)
		}
	}
	validation := []int{}
	for _, root := range arena.validationRoots {
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
	// returned native Types. Only reachable output shapes are materialized.
	if len(arena.validationRoots) != len(roots) {
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
	aliases := map[int]*ast.Node{}
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
				if p.key != nil {
					return v.error(node, "symbol-keyed recursive construction is not implemented yet")
				}
				var modifiers *ast.ModifierList
				if p.readonly {
					modifiers = f.NewModifierList([]*ast.Node{f.NewToken(ast.KindReadonlyKeyword)})
				}
				var optional *ast.Node
				if p.optional {
					optional = f.NewToken(ast.KindQuestionToken)
				}
				members[i] = f.NewPropertySignatureDeclaration(modifiers, f.NewStringLiteral(p.name, ast.TokenFlagsNone), optional, inputs[i], nil)
			}
			body = f.NewTypeLiteralNode(f.NewNodeList(members))
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
	// All edges are now native syntax and all leaf arguments are ordinary Types.
	// Release the private construction context before native evaluation/replay.
	v.construction = nil
	savedSite := v.nativeIslandSite
	v.nativeIslandSite = node
	defer func() { v.nativeIslandSite = savedSite }()
	environments := map[int]*Type{}
	result := make([]*Type, len(roots))
	for i, root := range roots {
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
				result[i] = declared
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
				result[i] = v.c.getTypeAliasInstantiation(aliases[graph.index].Symbol(), []*Type{environment}, nil)
			}
		} else {
			result[i] = root
		}
		if v.c.isErrorType(result[i]) {
			v.failed = true
			return v.c.errorType
		}
	}
	if group {
		return v.c.createTupleType(result)
	}
	return result[0]
}
