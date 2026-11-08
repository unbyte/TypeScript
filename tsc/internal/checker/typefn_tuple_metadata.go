package checker

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/core"
)

type typefnTupleKey struct {
	base     *Type
	declared CacheHashKey
}

func (c *Checker) typefnTupleNodeDeclared(node *ast.Node) []*Type {
	return core.Map(node.Elements(), func(element *ast.Node) *Type {
		if element.Kind == ast.KindOptionalType || element.Kind == ast.KindNamedTupleMember && element.QuestionToken() != nil {
			return c.getTypeFromTypeNode(element.Type())
		}
		return c.getTypeFromTypeNode(element)
	})
}

// A nil optional slot means native construction did not retain its provenance.
// It must not be guessed by stripping undefined, which can be explicitly declared.
func (c *Checker) typefnTupleDeclared(t *Type) []*Type {
	if c.typefn != nil {
		if values := c.typefn.tupleDeclared[t]; values != nil {
			return values
		}
	}
	reference := t.AsTypeReference()
	if reference.node != nil && reference.node.Kind == ast.KindTupleType {
		return c.instantiateTypes(c.typefnTupleNodeDeclared(reference.node), reference.mapper)
	}
	values := c.getElementTypes(t)
	if t.TargetTupleType().combinedFlags&ElementFlagsOptional == 0 {
		return values
	}
	values = slices.Clone(values)
	for i, info := range t.TargetTupleType().elementInfos {
		if info.flags&ElementFlagsOptional != 0 {
			values[i] = nil
		}
	}
	return values
}

// Metadata-bearing views have unchanged native targets and semantic arguments.
// Separate interning retains declared undefined distinctions that ordinary native
// optional-read interning intentionally erases in the non-exact compiler mode.
func (c *Checker) withTypefnTupleDeclared(base *Type, declared []*Type) *Type {
	if declared == nil || !isTupleType(base) {
		return base
	}
	known := false
	for i, info := range base.TargetTupleType().elementInfos {
		if info.flags&ElementFlagsOptional != 0 && i < len(declared) && declared[i] != nil {
			known = true
			break
		}
	}
	if !known {
		return base
	}
	state := c.getTypefnState()
	var key keyBuilder
	key.writeInt(len(declared))
	for _, value := range declared {
		if value == nil {
			key.writeUint32(0)
		} else {
			key.writeType(value)
		}
	}
	id := typefnTupleKey{base, key.hash()}
	if t := state.tupleViews[id]; t != nil {
		return t
	}
	if state.tupleViews == nil {
		state.tupleViews = map[typefnTupleKey]*Type{}
	}
	if state.tupleDeclared == nil {
		state.tupleDeclared = map[*Type][]*Type{}
	}
	result := c.cloneTypeReference(base)
	state.tupleDeclared[result] = slices.Clone(declared)
	state.tupleViews[id] = result
	return result
}

func (c *Checker) instantiateTypefnTuple(t *Type, mapper *TypeMapper) *Type {
	old := c.getTypeArguments(t)
	values := c.instantiateTypes(old, mapper)
	declared := c.typefn.tupleDeclared[t]
	mapped := core.SameMap(declared, func(value *Type) *Type {
		if value == nil {
			return nil
		}
		return c.instantiateType(value, mapper)
	})
	if core.Same(old, values) && core.Same(declared, mapped) {
		return t
	}
	return c.createNormalizedTupleTypeWithDeclared(t.Target(), values, mapped, ObjectFlagsNone)
}

// Mapped optionality changes read Types, but not the declared template value.
// Identity projections also retain the source's declared optional value, just as
// native property mappings do. Other templates use their native computed value.
func (c *Checker) typefnMappedTupleDeclared(source, sourceValue, mapped, key *Type, mapper *TypeMapper) *Type {
	base := core.OrElse(mapped.AsMappedType().target, mapped)
	template := ast.SkipParentheses(base.AsMappedType().declaration.Type)
	if template == nil {
		return nil
	}
	templateMapper := appendTypeMapping(mapper, c.getTypeParameterFromMappedType(mapped), key)
	instantiate := func(node *ast.Node) *Type {
		return c.instantiateType(c.instantiateType(c.getTypeFromTypeNode(node), base.AsMappedType().mapper), templateMapper)
	}
	if ast.IsIndexedAccessTypeNode(template) {
		access := template.AsIndexedAccessTypeNode()
		if c.isTypeIdenticalTo(instantiate(access.ObjectType), source) && c.isTypeIdenticalTo(instantiate(access.IndexType), key) {
			return sourceValue
		}
	}
	return c.removeMissingType(instantiate(template), true)
}

// Only tuple-producing paths request provenance. Disabled native projects do not
// allocate side metadata or inspect optional declarations.
func (c *Checker) typefnTupleDeclaredIfEnabled(t *Type) []*Type {
	if c.compilerOptions.GetTypefnVersion() != 1 || !isTupleType(t) || t.TargetTupleType().combinedFlags&ElementFlagsOptional == 0 {
		return nil
	}
	return c.typefnTupleDeclared(t)
}

func (c *Checker) copyTupleTypeRange(t *Type, start, end int, readonly bool) *Type {
	declared := c.typefnTupleDeclaredIfEnabled(t)
	if declared != nil {
		declared = declared[start:end]
	}
	return c.createTupleTypeWithDeclared(c.getTypeArguments(t)[start:end], t.TargetTupleType().elementInfos[start:end], readonly, declared)
}

func (c *Checker) copyTypefnTupleMetadata(source, result *Type) {
	if c.compilerOptions.GetTypefnVersion() != 1 || !isTupleType(source) {
		return
	}
	var declared []*Type
	if c.typefn != nil {
		declared = c.typefn.tupleDeclared[source]
	}
	if declared == nil && source.AsTypeReference().node != nil {
		declared = c.typefnTupleDeclaredIfEnabled(source)
	}
	if declared != nil {
		state := c.getTypefnState()
		if state.tupleDeclared == nil {
			state.tupleDeclared = map[*Type][]*Type{}
		}
		// Provenance slices are immutable and checker-local, just like native arguments.
		state.tupleDeclared[result] = declared
	}
}
