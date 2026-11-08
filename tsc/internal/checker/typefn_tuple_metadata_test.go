package checker

import (
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/jsnum"
	"testing"
)

func TestTypefnTupleMetadataViews(t *testing.T) {
	c, types := newCompositeTestChecker()
	target := c.newObjectType(ObjectFlagsTuple|ObjectFlagsReference, nil)
	target.AsTupleType().elementInfos = []TupleElementInfo{{flags: ElementFlagsOptional}}
	target.AsTupleType().combinedFlags = ElementFlagsOptional
	target.AsTupleType().instantiations = map[CacheHashKey]*Type{}
	base := c.createTypeReference(target, []*Type{c.unknownType})
	values := []*Type{types[0]}
	first := c.withTypefnTupleDeclared(base, values)
	values[0] = types[1]
	second := c.withTypefnTupleDeclared(base, values)
	if first == second || c.typefnTupleDeclared(first)[0] != types[0] {
		t.Fatal("distinct declared values lost identity or snapshot isolation")
	}
	if c.getTypeArguments(first)[0] != c.unknownType || first.Target() != base.Target() {
		t.Fatal("metadata changed native semantics")
	}
	if c.withTypefnTupleDeclared(base, []*Type{types[0]}) != first {
		t.Fatal("equivalent metadata views were not interned")
	}
	if c.withTypefnTupleDeclared(base, []*Type{nil}) != base {
		t.Fatal("unknown provenance should not allocate a metadata view")
	}
	other, _ := newCompositeTestChecker()
	if other.typefn != nil {
		t.Fatal("tuple metadata escaped its checker snapshot")
	}
}

func TestTypefnTupleNormalizationRetainsDeclaredValues(t *testing.T) {
	c, types := newCompositeTestChecker()
	c.compilerOptions = &core.CompilerOptions{Typefn: &core.TypefnOptions{Version: 1}}
	c.tupleTypes = map[CacheHashKey]*Type{}
	c.numberLiteralTypes = map[jsnum.Number]*Type{}
	c.undefinedType = c.newIntrinsicType(TypeFlagsUndefined, "undefined")
	c.undefinedOrMissingType = c.undefinedType
	optionalRead := c.getUnionType([]*Type{types[0], c.undefinedType})
	target := c.getTupleTargetType([]TupleElementInfo{{flags: ElementFlagsOptional}}, false)
	implicit := c.createNormalizedTupleTypeWithDeclared(target, []*Type{optionalRead}, []*Type{types[0]}, ObjectFlagsNone)
	explicit := c.createNormalizedTupleTypeWithDeclared(target, []*Type{optionalRead}, []*Type{optionalRead}, ObjectFlagsNone)
	if implicit == explicit {
		t.Fatal("native optional-read interning erased declared undefined")
	}
	spreadTarget := c.getTupleTargetType([]TupleElementInfo{{flags: ElementFlagsVariadic}}, false)
	spread := c.createNormalizedTupleTypeWithDeclared(spreadTarget, []*Type{implicit}, nil, ObjectFlagsNone)
	if c.typefnTupleDeclared(spread)[0] != types[0] || c.getElementTypes(spread)[0] != optionalRead {
		t.Fatal("spread lost declared/read distinction")
	}
	requiredTarget := c.getTupleTargetType([]TupleElementInfo{{flags: ElementFlagsVariadic}, {flags: ElementFlagsRequired}}, false)
	required := c.createNormalizedTupleTypeWithDeclared(requiredTarget, []*Type{implicit, types[1]}, nil, ObjectFlagsNone)
	if required.TargetTupleType().elementInfos[0].flags != ElementFlagsRequired || c.typefnTupleDeclared(required)[0] != optionalRead {
		t.Fatal("optional-to-required normalization lost native undefined")
	}
}
