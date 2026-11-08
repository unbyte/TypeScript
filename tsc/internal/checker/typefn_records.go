package checker

import "github.com/microsoft/TypeScript/tsc/internal/ast"

type typefnRecordField struct {
	name  string
	value *Type
}

// Library descriptors are ordinary immutable native Types. Field order follows
// the library schema, and values retain native identities and pending checks.
func (v *typefnVM) record(node *ast.Node, fields ...typefnRecordField) *Type {
	if !v.allocate(node, len(fields)+1) {
		return v.c.errorType
	}
	members := ast.NewSymbolTable()
	for _, field := range fields {
		value := v.observe(field.value)
		if v.c.isErrorType(value) {
			v.failed = true
			return v.c.errorType
		}
		symbol := v.c.newSymbolEx(ast.SymbolFlagsProperty, field.name, ast.CheckFlagsNone)
		v.c.valueSymbolLinks.Get(symbol).resolvedType = value
		members.Set(field.name, symbol)
	}
	return v.c.newTypefnObject(members)
}
