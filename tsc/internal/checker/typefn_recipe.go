package checker

import (
	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/nodebuilder"
)

// A source recipe belongs to an expression and its lexical instantiation, not to
// the resulting Type. Literal Types are shared by unrelated native expressions.
type typefnSourceRecipe struct {
	node   *ast.Node
	mapper *TypeMapper
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
	for _, dependency := range c.typefnDeclarationDependencies(annotation, false) {
		if c.typefnDeclarationsHaveRecipe(dependency.Declarations) {
			return &typefnSourceRecipe{annotation, signature.mapper}
		}
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
	b.ctx.mapper = recipe.mapper
	if oldMapper != nil {
		b.ctx.mapper = b.ch.combineTypeMappers(recipe.mapper, oldMapper)
	}
	b.ctx.internalFlags |= nodebuilder.InternalFlagsTypefnRecipe
	defer func() { b.ctx.mapper, b.ctx.internalFlags = oldMapper, oldFlags }()
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
