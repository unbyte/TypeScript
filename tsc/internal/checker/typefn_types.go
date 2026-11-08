package checker

import "github.com/microsoft/TypeScript/tsc/internal/ast"

// Typefn terms are checker Types, not placeholder type parameters. A term owns
// immutable operands from this checker snapshot. Substitution either reuses the
// term or evaluates it with the substituted operands; it never mutates a frame.
type typefnTermKind uint8

const (
	typefnBinaryTerm typefnTermKind = iota
	typefnUnaryTerm
	typefnCallTerm
	typefnExpressionTerm
	typefnNativeCallTerm
	typefnGuardTerm
)

type typefnTerm struct {
	ConstrainedType
	kind        typefnTermKind
	node        *ast.Node
	declaration *ast.Node
	symbol      *ast.Symbol
	expression  *typefnExpression
	operator    ast.Kind
	args        []*Type
	shape       *Type // Only guards have an already-known result shape.
}

type typefnTermKey struct {
	kind        typefnTermKind
	declaration *ast.Node
	symbol      *ast.Symbol
	expression  *typefnExpression
	operator    ast.Kind
	operands    CacheHashKey
	shape       *Type
}
