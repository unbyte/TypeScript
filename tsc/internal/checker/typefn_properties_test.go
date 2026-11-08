package checker

import (
	"context"
	"testing"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/core"
)

func TestTypefnPropertyMetadataInstantiation(t *testing.T) {
	c, types := newCompositeTestChecker()
	c.ctx = context.Background()
	c.initializeClosures()
	c.compilerOptions = &core.CompilerOptions{Typefn: &core.TypefnOptions{Version: 1}}
	parameter := c.newType(TypeFlagsTypeParameter, ObjectFlagsNone, (&TypeParameter{}).AsType(), typeKindTypeParameter)
	property := c.newSymbolEx(ast.SymbolFlagsProperty|ast.SymbolFlagsOptional, "value", ast.CheckFlagsNone)
	// The read Type can stay unchanged even when declared-value metadata changes.
	c.valueSymbolLinks.Get(property).resolvedType = c.unknownType
	c.rememberTypefnProperty(property, parameter)
	members := ast.NewSymbolTable()
	members.Set("value", property)
	original := c.newTypefnObject(members)
	unchanged := c.instantiateTypefnObject(original, newSimpleTypeMapper(parameter, parameter))
	if unchanged != original || len(c.typefn.propertyValues) != 1 {
		t.Fatal("unchanged mapping retained redundant property metadata")
	}
	changed := c.instantiateTypefnObject(original, newSimpleTypeMapper(parameter, types[0]))
	if changed == original {
		t.Fatal("declared-value-only substitution was lost")
	}
	p := c.getPropertyOfType(changed, "value")
	if c.getTypeOfSymbol(p) != c.unknownType || c.typefn.propertyValues[p] != types[0] {
		t.Fatal("read and declared Types did not remain independent")
	}
	if c.typefn.propertyValues[property] != parameter {
		t.Fatal("substitution mutated the original snapshot")
	}
}
