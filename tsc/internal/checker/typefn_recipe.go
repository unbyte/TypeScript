package checker

import (
	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/nodebuilder"
	"github.com/microsoft/TypeScript/tsc/internal/pseudochecker"
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
	reference           *ast.Node
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

// Native entity-name serialization preserves namespace/import visibility. An
// ordinary object member needs a receiver-relative recipe instead; it must not
// be mistaken for a namespace export just because it uses dotted syntax.
func (c *Checker) typefnRecipeValueName(name *ast.Node) bool {
	if !ast.IsEntityNameExpression(name) || ast.IsOptionalChain(name) {
		return false
	}
	if ast.IsIdentifier(name) {
		return true
	}
	symbol := c.resolveEntityName(name, ast.SymbolFlagsValue, true, false, nil)
	return symbol != nil && symbol != c.unknownSymbol
}

func (c *Checker) typefnExpressionRecipe(expression *ast.Node) *typefnSourceRecipe {
	if expression == nil || c.compilerOptions.GetTypefnVersion() != 1 {
		return nil
	}
	expression = ast.SkipParentheses(expression)
	if c.typefnRecipeValueName(expression) {
		symbol := c.resolveEntityName(expression, ast.SymbolFlagsValue, true, false, nil)
		if len(c.typefnSymbolDependencies(symbol)) != 0 {
			return &typefnSourceRecipe{reference: expression}
		}
		return nil
	}
	if ast.IsExpressionWithTypeArguments(expression) && c.typefnRecipeValueName(ast.SkipParentheses(expression.Expression())) {
		if c.typefnNodeHasRecipe(expression) {
			return &typefnSourceRecipe{reference: expression}
		}
		return nil
	}
	if !ast.IsCallExpression(expression) {
		return nil
	}
	signature := c.getResolvedSignature(expression, nil, CheckModeNormal)
	if signature == nil {
		return nil
	}
	var annotation *ast.Node
	var typeParameters, parameters []*ast.Node
	if declaration := signature.declaration; declaration != nil {
		annotation = declaration.Type()
		typeParameters, parameters = declaration.TypeParameters(), declaration.Parameters()
	}
	if annotation != nil && ast.IsTypePredicateNode(annotation) {
		return nil
	}
	recipe := &typefnSourceRecipe{node: annotation, mapper: signature.mapper, call: expression, signature: signature}
	hasRecipe := c.typefnNodeHasRecipe(expression.Expression())
	if annotation != nil {
		hasRecipe = hasRecipe || c.typefnNodeHasRecipe(annotation)
	}
	for _, parameter := range typeParameters {
		if constraint := parameter.AsTypeParameterDeclaration().Constraint; constraint != nil {
			recipe.applicationRequired = true
			hasRecipe = hasRecipe || c.typefnNodeHasRecipe(constraint)
		}
	}
	for _, parameter := range parameters {
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
		for i, parameter := range typeParameters {
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
	if hasRecipe && c.getTypePredicateOfSignature(signature) == nil {
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
	if recipe.reference != nil {
		// A value query denotes the declared Type, not a control-flow-narrowed
		// use. Keep native widening/narrowing when those Types differ.
		reference := recipe.reference
		name := reference
		if ast.IsExpressionWithTypeArguments(reference) {
			name = ast.SkipParentheses(reference.Expression())
		}
		symbol := b.ch.resolveEntityName(name, ast.SymbolFlagsValue, true, false, nil)
		t := b.ch.getTypeOfSymbol(symbol)
		if reference != name {
			if !b.ch.isTypeIdenticalTo(t, b.ch.getTypeOfExpression(name)) {
				return nil
			}
			t = b.ch.getTypeOfExpression(reference)
		}
		t = b.ch.instantiateType(t, b.ctx.mapper)
		if !b.ch.isTypeIdenticalTo(t, inferred) {
			return nil
		}
		if result := b.typefnValueRecipeNode(reference); result != nil {
			return result
		}
		b.ctx.tracker.ReportTypefnRecipeError()
		b.ctx.encounteredError = true
		return b.f.NewKeywordTypeNode(ast.KindUnknownKeyword)
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
	var resultType *Type
	if recipe.node != nil {
		resultType = b.getTypeFromTypeNode(recipe.node, false)
	} else {
		// The native signature's inferred return is already instantiated. Only
		// apply the enclosing builder mapper, not the call mapper a second time.
		resultType = b.ch.instantiateType(b.ch.getReturnTypeOfSignature(recipe.signature), oldMapper)
	}
	if !b.ch.isTypeIdenticalTo(resultType, inferred) {
		return nil
	}
	if result := b.typefnCallRecipeNode(recipe); result != nil {
		return result
	}
	if !recipe.applicationRequired && recipe.node != nil {
		if result := b.tryReuseExistingNodeHelper(recipe.node); result != nil {
			return result
		}
	}
	// Failure to name a known recipe must not silently publish its current value.
	b.ctx.tracker.ReportTypefnRecipeError()
	b.ctx.encounteredError = true
	return b.f.NewKeywordTypeNode(ast.KindUnknownKeyword)
}

func (b *NodeBuilderImpl) typefnValueRecipeNode(expression *ast.Node) *ast.Node {
	bound := b.createRecoveryBoundary()
	oldFlags := b.ctx.internalFlags
	b.ctx.internalFlags |= nodebuilder.InternalFlagsTypefnRecipe
	defer func() { b.ctx.internalFlags = oldFlags }()
	name := expression
	var arguments []*ast.Node
	if ast.IsExpressionWithTypeArguments(expression) {
		name = ast.SkipParentheses(expression.Expression())
		for _, argument := range expression.TypeArguments() {
			node := b.tryReuseExistingNodeHelper(argument)
			if node == nil {
				bound.markError(nil)
				break
			}
			arguments = append(arguments, node)
		}
	}
	result := b.typefnValueQuery(name, arguments)
	if result == nil {
		bound.markError(nil)
	}
	if !b.finalizeBoundary(bound) {
		return nil
	}
	return result
}

func (b *NodeBuilderImpl) typefnValueQuery(name *ast.Node, arguments []*ast.Node) *ast.Node {
	result := b.serializeTypeName(name, true, nil)
	if result == nil || len(arguments) == 0 {
		return result
	}
	typeArguments := b.f.NewNodeList(arguments)
	switch result.Kind {
	case ast.KindTypeQuery:
		return b.f.UpdateTypeQueryNode(result.AsTypeQueryNode(), result.AsTypeQueryNode().ExprName, typeArguments)
	case ast.KindImportType:
		d := result.AsImportTypeNode()
		return b.f.UpdateImportTypeNode(d, d.IsTypeOf, d.Argument, d.Attributes, d.Qualifier, typeArguments)
	default:
		return nil
	}
}

func (b *NodeBuilderImpl) tryTypefnReturnRecipe(signature *Signature, inferred *Type) *ast.Node {
	if b.ctx.internalFlags&nodebuilder.InternalFlagsTypefnDeclaration == 0 ||
		b.ch.compilerOptions.GetTypefnVersion() != 1 || signature.declaration == nil || signature.declaration.Type() != nil {
		return nil
	}
	// Use the native single-return analysis; multiple returns, generators and
	// async transformations require their own source-equivalent recipes.
	pt := b.pc.GetReturnTypeOfSignature(signature.declaration)
	if pt.Kind == pseudochecker.PseudoTypeKindInferred {
		expression := pt.AsPseudoTypeInferred()
		if !expression.IsSignatureReturn {
			return b.tryTypefnExpressionRecipe(expression.Expression, inferred)
		}
	}
	return nil
}

// Reusing the callable's instantiated Type retains native bounds and inaccessible
// private return helpers without copying their definitions into another module.
// A conditional over multiple overloads would select the last one, so this path
// requires one signature with the same declaration as the selected call.
func (b *NodeBuilderImpl) typefnCallRecipeNode(recipe *typefnSourceRecipe) *ast.Node {
	name := ast.SkipParentheses(recipe.call.Expression())
	valueName := name
	if ast.IsExpressionWithTypeArguments(name) {
		valueName = ast.SkipParentheses(name.Expression())
	}
	if !b.ch.typefnRecipeValueName(valueName) {
		return nil
	}
	if !ast.IsIdentifier(valueName) && recipe.signature.ThisParameter() != nil {
		// Explicit receiver obligations need a call recipe that validates the
		// receiver as well as the callable's generic and parameter bounds.
		return nil
	}
	signatures := b.ch.getSignaturesOfType(b.ch.getTypeOfExpression(name), SignatureKindCall)
	if len(signatures) != 1 || signatures[0].declaration != recipe.signature.declaration {
		return nil
	}
	var parameters []*ast.Node
	if recipe.signature.declaration != nil {
		parameters = recipe.signature.declaration.TypeParameters()
	} else if signatures[0] != recipe.signature {
		// Constructed nongeneric signatures have no source declaration. Their
		// identity, rather than two nil declaration pointers, selects the call.
		return nil
	}
	if len(signatures[0].TypeParameters()) == 0 {
		parameters = nil // The callee already carries its specialization recipe.
	} else if len(signatures[0].TypeParameters()) != len(parameters) {
		return nil
	}
	bound := b.createRecoveryBoundary()
	var arguments []*ast.Node
	for _, parameter := range parameters {
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
	var callable *ast.Node
	if ast.IsExpressionWithTypeArguments(name) {
		// A specialization expression was written in the caller's scope, before
		// applying the selected signature's mapper.
		mapper, environment := b.ctx.mapper, b.ctx.typefnRecipeEnvironment
		b.ctx.mapper, b.ctx.typefnRecipeEnvironment = environment.callerMapper, environment.parent
		callable = b.typefnValueRecipeNode(name)
		b.ctx.mapper, b.ctx.typefnRecipeEnvironment = mapper, environment
	} else {
		callable = b.typefnValueQuery(name, arguments)
	}
	if callable == nil {
		bound.markError(nil)
	}
	if !b.finalizeBoundary(bound) {
		return nil
	}
	infer := func(name string) *ast.Node {
		return b.f.NewInferTypeNode(b.f.NewTypeParameterDeclaration(nil, b.f.NewIdentifier(name), nil, nil, nil))
	}
	patternParameters := b.f.NewNodeList([]*ast.Node{b.f.NewParameterDeclaration(nil, b.f.NewToken(ast.KindDotDotDotToken), b.f.NewIdentifier("__args"), nil, infer("__typefn_args"), nil)})
	pattern := b.f.NewFunctionTypeNode(nil, patternParameters, infer("__typefn_result"))
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
