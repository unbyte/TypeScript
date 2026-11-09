package checker

import (
	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/nodebuilder"
)

// A source recipe belongs to an expression and its lexical instantiation, not to
// the resulting Type. Literal Types are shared by unrelated native expressions.
type typefnSourceRecipe struct {
	node                *ast.Node
	mapper              *TypeMapper
	arguments           map[*Type]typefnRecipeArgument
	call                *ast.Node
	signature           *Signature
	applicationRequired bool
}

type typefnRecipeArgument struct {
	node      *ast.Node
	defaulted bool
}

// Explicit arguments belong to the caller's scope. Defaults belong to the
// callee's scope and can refer to preceding argument bindings.
type typefnRecipeEnvironment struct {
	arguments    map[*Type]typefnRecipeArgument
	parent       *typefnRecipeEnvironment
	callerMapper *TypeMapper
	active       map[*Type]bool
}

func (c *Checker) typefnNodeHasRecipe(node *ast.Node) bool {
	for _, dependency := range c.typefnDeclarationDependencies(node, false) {
		if c.typefnDeclarationsHaveRecipe(dependency.Declarations) {
			return true
		}
	}
	return false
}

func (c *Checker) typefnExpressionRecipe(expression *ast.Node) *typefnSourceRecipe {
	if expression == nil || c.compilerOptions.GetTypefnVersion() != 1 {
		return nil
	}
	expression = ast.SkipParentheses(expression)
	if !ast.IsCallExpression(expression) {
		return nil
	}
	signature := c.getResolvedSignature(expression, nil, CheckModeNormal)
	if signature == nil || signature.declaration == nil {
		return nil
	}
	annotation := signature.declaration.Type()
	if annotation == nil || ast.IsTypePredicateNode(annotation) {
		return nil
	}
	recipe := &typefnSourceRecipe{node: annotation, mapper: signature.mapper, call: expression, signature: signature}
	hasRecipe := c.typefnNodeHasRecipe(annotation)
	for _, parameter := range signature.declaration.TypeParameters() {
		if constraint := parameter.AsTypeParameterDeclaration().Constraint; constraint != nil {
			recipe.applicationRequired = true
			hasRecipe = hasRecipe || c.typefnNodeHasRecipe(constraint)
		}
	}
	for _, parameter := range signature.declaration.Parameters() {
		if parameter.Type() != nil && c.typefnParameterHasRecipe(parameter.Type()) {
			recipe.applicationRequired = true
			hasRecipe = true
		}
	}
	arguments := expression.TypeArguments()
	defaults := c.typefnCallDefaultOrigin(expression)
	if defaults != nil && defaults.signature != signature {
		defaults = nil
	}
	if len(arguments) != 0 || defaults != nil {
		recipe.arguments = map[*Type]typefnRecipeArgument{}
		for i, parameter := range signature.declaration.TypeParameters() {
			if len(arguments) == 0 && (i >= len(defaults.used) || !defaults.used[i]) {
				continue
			}
			argument := parameter.AsTypeParameterDeclaration().DefaultType
			defaulted := i >= len(arguments)
			if !defaulted {
				argument = arguments[i]
			}
			if argument != nil {
				parameterType := c.getDeclaredTypeOfTypeParameter(parameter.Symbol())
				recipe.arguments[parameterType] = typefnRecipeArgument{argument, defaulted}
				hasRecipe = hasRecipe || c.typefnNodeHasRecipe(argument)
			}
		}
	}
	if hasRecipe {
		return recipe
	}
	return nil
}

func (c *Checker) typefnParameterHasRecipe(node *ast.Node) bool {
	for _, dependency := range c.typefnDeclarationDependencies(node, false) {
		for _, declaration := range dependency.Declarations {
			// Referring to T does not execute T's unused default. Its selected
			// argument and native constraint are tracked separately.
			if !ast.IsTypeParameterDeclaration(declaration) && c.typefnDeclarationsHaveRecipe([]*ast.Node{declaration}) {
				return true
			}
		}
	}
	return false
}

func (b *NodeBuilderImpl) tryTypefnExpressionRecipe(expression *ast.Node, inferred *Type) *ast.Node {
	if b.ctx.internalFlags&nodebuilder.InternalFlagsTypefnDeclaration == 0 {
		return nil
	}
	recipe := b.ch.typefnExpressionRecipe(expression)
	if recipe == nil {
		return nil
	}
	oldMapper, oldFlags := b.ctx.mapper, b.ctx.internalFlags
	oldEnvironment := b.ctx.typefnRecipeEnvironment
	b.ctx.mapper = recipe.mapper
	if oldMapper != nil {
		b.ctx.mapper = b.ch.combineTypeMappers(recipe.mapper, oldMapper)
	}
	b.ctx.internalFlags |= nodebuilder.InternalFlagsTypefnRecipe
	b.ctx.typefnRecipeEnvironment = &typefnRecipeEnvironment{arguments: recipe.arguments, parent: oldEnvironment, callerMapper: oldMapper}
	defer func() {
		b.ctx.mapper, b.ctx.internalFlags = oldMapper, oldFlags
		b.ctx.typefnRecipeEnvironment = oldEnvironment
	}()
	// Native widening or a surrounding operation may change the inferred Type.
	// Only reuse a signature recipe when it describes this expression's Type.
	if !b.ch.isTypeIdenticalTo(b.getTypeFromTypeNode(recipe.node, false), inferred) {
		return nil
	}
	if result := b.typefnCallRecipeNode(recipe); result != nil {
		return result
	}
	if !recipe.applicationRequired {
		if result := b.tryReuseExistingNodeHelper(recipe.node); result != nil {
			return result
		}
	}
	// Failure to name a known recipe must not silently publish its current value.
	b.ctx.tracker.ReportTypefnRecipeError()
	b.ctx.encounteredError = true
	return b.f.NewKeywordTypeNode(ast.KindUnknownKeyword)
}

// Reusing the callable's instantiated Type retains native bounds and inaccessible
// private return helpers without copying their definitions into another module.
// A conditional over multiple overloads would select the last one, so this path
// requires one signature with the same declaration as the selected call.
func (b *NodeBuilderImpl) typefnCallRecipeNode(recipe *typefnSourceRecipe) *ast.Node {
	name := ast.SkipParentheses(recipe.call.Expression())
	if !ast.IsIdentifier(name) {
		return nil
	}
	signatures := b.ch.getSignaturesOfType(b.ch.getTypeOfExpression(name), SignatureKindCall)
	if len(signatures) != 1 || signatures[0].declaration != recipe.signature.declaration ||
		len(signatures[0].TypeParameters()) != len(recipe.signature.declaration.TypeParameters()) {
		return nil
	}
	bound := b.createRecoveryBoundary()
	var arguments []*ast.Node
	for _, parameter := range recipe.signature.declaration.TypeParameters() {
		t := b.ch.getDeclaredTypeOfTypeParameter(parameter.Symbol())
		argument, found := b.typefnRecipeArgumentNode(t)
		if !found {
			argument = b.typeToTypeNode(b.ch.instantiateType(t, b.ctx.mapper))
		}
		if argument == nil {
			bound.markError(nil)
			break
		}
		arguments = append(arguments, argument)
	}
	var typeArguments *ast.NodeList
	if len(arguments) != 0 {
		typeArguments = b.f.NewNodeList(arguments)
	}
	callable := b.serializeTypeName(name, true, nil)
	if callable == nil {
		bound.markError(nil)
	} else if typeArguments != nil {
		switch callable.Kind {
		case ast.KindTypeQuery:
			callable = b.f.UpdateTypeQueryNode(callable.AsTypeQueryNode(), callable.AsTypeQueryNode().ExprName, typeArguments)
		case ast.KindImportType:
			d := callable.AsImportTypeNode()
			callable = b.f.UpdateImportTypeNode(d, d.IsTypeOf, d.Argument, d.Attributes, d.Qualifier, typeArguments)
		default:
			bound.markError(nil)
		}
	}
	if !b.finalizeBoundary(bound) {
		return nil
	}
	infer := func(name string) *ast.Node {
		return b.f.NewInferTypeNode(b.f.NewTypeParameterDeclaration(nil, b.f.NewIdentifier(name), nil, nil, nil))
	}
	parameters := b.f.NewNodeList([]*ast.Node{b.f.NewParameterDeclaration(nil, b.f.NewToken(ast.KindDotDotDotToken), b.f.NewIdentifier("__args"), nil, infer("__typefn_args"), nil)})
	pattern := b.f.NewFunctionTypeNode(nil, parameters, infer("__typefn_result"))
	result := b.f.NewTypeReferenceNode(b.f.NewIdentifier("__typefn_result"), nil)
	return b.f.NewConditionalTypeNode(callable, pattern, result, b.f.NewKeywordTypeNode(ast.KindNeverKeyword))
}

func (b *NodeBuilderImpl) typefnRecipeArgumentNode(parameter *Type) (node *ast.Node, found bool) {
	environment := b.ctx.typefnRecipeEnvironment
	if environment == nil {
		return nil, false
	}
	argument, found := environment.arguments[parameter]
	if !found {
		return nil, false
	}
	if environment.active[parameter] {
		return nil, true
	}
	if environment.active == nil {
		environment.active = map[*Type]bool{}
	}
	environment.active[parameter] = true
	oldMapper := b.ctx.mapper
	if !argument.defaulted {
		b.ctx.mapper = environment.callerMapper
		b.ctx.typefnRecipeEnvironment = environment.parent
	}
	defer func() {
		delete(environment.active, parameter)
		b.ctx.mapper = oldMapper
		b.ctx.typefnRecipeEnvironment = environment
	}()
	return b.tryReuseExistingNodeHelper(argument.node), true
}
