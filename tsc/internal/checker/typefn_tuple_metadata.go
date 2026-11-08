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
