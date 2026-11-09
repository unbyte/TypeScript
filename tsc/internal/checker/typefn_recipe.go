package checker

import (
	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/nodebuilder"
)

// A source recipe belongs to an expression and its lexical instantiation, not to
// the resulting Type. Literal Types are shared by unrelated native expressions.
type typefnSourceRecipe struct {
	node      *ast.Node
	mapper    *TypeMapper
	arguments map[*Type]typefnRecipeArgument
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
	recipe := &typefnSourceRecipe{node: annotation, mapper: signature.mapper}
	hasRecipe := c.typefnNodeHasRecipe(annotation)
	arguments := expression.TypeArguments()
	if len(arguments) != 0 {
		recipe.arguments = map[*Type]typefnRecipeArgument{}
		for i, parameter := range signature.declaration.TypeParameters() {
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
	if result := b.tryReuseExistingNodeHelper(recipe.node); result != nil {
		return result
	}
	// Failure to name a known recipe must not silently publish its current value.
	b.ctx.tracker.ReportTypefnRecipeError()
	b.ctx.encounteredError = true
	return b.f.NewKeywordTypeNode(ast.KindUnknownKeyword)
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
