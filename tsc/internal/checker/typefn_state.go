package checker

import "github.com/microsoft/TypeScript/tsc/internal/ast"

// All DSL-owned state belongs to one checker snapshot. Disabled-profile checking does
// not allocate it. The VM is scoped to a root evaluation; programs, immutable
// terms and syntax-dependency tables live for the owning checker lifetime.
type typefnState struct {
	// Native resolution can outlive the root VM that reported a failure. Cache
	// publishers compare this generation to avoid retaining its recovery Type.
	failureEpoch uint64

	// Native frontend checks may return recovery Types after reporting an error.
	// A scoped definition check must reject that program before VM execution.
	definitionCheck        *typefnProgram
	declarationABI         map[*ast.SourceFile]*ast.Diagnostic
	emitRecipeReachability map[*ast.Node]bool
	callResolutions        []typefnCallResolution
	callDefaults           map[*ast.Node]*typefnCallDefaults

	constructionMembers map[*ast.Symbol]*typefnConstructionMembers
	programs            map[*ast.Node]*typefnProgram
	kindStack           []*typefnProgram
	nextKindIndex       int
	closures            map[typefnClosureKey]*typefnClosure
	nativeConstructors  map[*ast.Symbol]*typefnNativeConstructor
	aliases             map[*ast.Symbol]*typefnAlias
	nextClosureID       uint32
	vm                  *typefnVM
	terms               map[typefnTermKey]*Type
	syntaxChecks        map[*ast.Node]bool
	objects             map[*Type]bool
	propertyValues      map[*ast.Symbol]*Type
	tupleDeclared       map[*Type][]*Type
	tupleViews          map[typefnTupleKey]*Type
	tupleLabels         map[typefnTupleLabelKey]*ast.Node
}

func (c *Checker) getTypefnState() *typefnState {
	if c.typefn == nil {
		c.typefn = &typefnState{}
	}
	return c.typefn
}
