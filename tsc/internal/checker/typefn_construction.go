package checker

import (
	"slices"
	"unsafe"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/jsnum"
)

// Scratch handles have the VM's Type kind, but never enter native normalization,
// relation, alias or closure caches. Every container that carries a scratch handle
// is itself scratch. Thus membership is an O(1) test, not a native graph walk.
type typefnConstruction struct {
	arena      *typefnConstructionArena
	index      int
	kind       ast.Kind
	inputs     []*Type
	properties []typefnConstructionProperty
	elements   []TupleElementInfo
	readonly   bool
	abstract   bool
	reference  int
}
type typefnConstructionProperty struct {
	name               string
	key                *Type
	optional, readonly bool
}
type typefnConstructionArena struct {
	nodes       []*Type
	definitions []*Type
	blocked     bool
	bytes       int
}

func constructionOf(t *Type) *typefnConstruction {
	if t != nil && t.flags&TypeFlagsTypefn != 0 && t.asTypefnTerm().kind == typefnConstructionTerm {
		return t.asTypefnTerm().construction
	}
	return nil
}
func (v *typefnVM) rejectConstruction(node *ast.Node, values ...*Type) bool {
	for _, value := range values {
		if constructionOf(value) != nil {
			v.error(node, "cannot observe an unsealed recursive reference")
			return true
		}
	}
	return false
}
func (v *typefnVM) construct(node *ast.Node, data *typefnConstruction) *Type {
	arena := v.construction
	if arena == nil {
		panic("recursive construction outside arena")
	}
	size := int(unsafe.Sizeof(typefnConstruction{})) + int(unsafe.Sizeof(typefnTerm{})) + len(data.inputs)*int(unsafe.Sizeof((*Type)(nil))) + len(data.properties)*int(unsafe.Sizeof(typefnConstructionProperty{})) + len(data.elements)*int(unsafe.Sizeof(TupleElementInfo{}))
	if !v.allocate(node, len(data.inputs)+1) || !v.reserveValues(node, size) {
		return v.c.errorType
	}
	arena.bytes += size
	for _, input := range data.inputs {
		if other := constructionOf(input); other != nil && other.arena != arena {
			return v.error(node, "recursive reference belongs to another construction")
		}
	}
	data.arena, data.index = arena, len(arena.nodes)
	term := &typefnTerm{kind: typefnConstructionTerm, node: node, construction: data}
	t := v.c.newType(TypeFlagsTypefn, ObjectFlagsNone, term.AsType(), typeKindTypefn)
	arena.nodes = append(arena.nodes, t)
	return t
}
func (v *typefnVM) constructionObject(node *ast.Node, properties []typefnConstructionProperty, values []*Type) *Type {
	for _, value := range values {
		if constructionOf(value) != nil {
			return v.construct(node, &typefnConstruction{kind: ast.KindTypeLiteral, inputs: values, properties: properties})
		}
	}
	if !v.allocate(node, len(properties)+1) {
		return v.c.errorType
	}
	members := ast.NewSymbolTable()
	for i, p := range properties {
		flags := ast.SymbolFlagsProperty
		if p.optional {
			flags |= ast.SymbolFlagsOptional
		}
		checks := ast.CheckFlagsNone
		if p.readonly {
			checks |= ast.CheckFlagsReadonly
		}
		symbol := v.c.newSymbolEx(flags, p.name, checks)
		links := v.c.valueSymbolLinks.Get(symbol)
		links.resolvedType = v.c.addOptionalityEx(values[i], true, p.optional)
		if p.optional {
			v.c.rememberTypefnProperty(symbol, values[i])
		}
		links.setNameType(v.c, p.key)
		members.Set(p.name, symbol)
	}
	return v.c.newTypefnObject(members)
}
func requiredConstructionTuple(graph *typefnConstruction) bool {
	if graph == nil || graph.kind != ast.KindTupleType {
		return false
	}
	for _, info := range graph.elements {
		if info.flags != ElementFlagsRequired {
			return false
		}
	}
	return true
}
func (v *typefnVM) constructionIndex(node *ast.Node, object, index *Type) *Type {
	graph := constructionOf(object)
	if requiredConstructionTuple(graph) && index.flags == TypeFlagsNumberLiteral {
		n := float64(index.AsLiteralType().value.(jsnum.Number))
		if n >= 0 && n < float64(len(graph.inputs)) && n == float64(int(n)) {
			return graph.inputs[int(n)]
		}
		return v.error(node, "recursive construction tuple index is out of bounds")
	}
	return v.error(node, "cannot observe an unsealed recursive reference")
}
func (v *typefnVM) recursiveConstruction(e *typefnExpression, slots []typefnValue) *Type {
	count := int64(1)
	group := e.intrinsic == "types.recursiveGroup"
	if group {
		value := v.expression(e.args[0], slots)
		if v.failed || v.blocked {
			return v.c.errorType
		}
		if v.rejectConstruction(e.node, value) {
			return v.c.errorType
		}
		if typefnUnresolved(value) {
			return v.deferExpression(e, slots)
		}
		var ok bool
		count, ok = v.intrinsicOffset(e.node, value)
		if !ok {
			return v.c.errorType
		}
		if count == 0 {
			return v.error(e.node, "recursive group count must be positive")
		}
	}
	builder := v.resultExpression(e.args[len(e.args)-1], slots)
	if v.failed || v.blocked {
		return v.c.errorType
	}
	if builder.closure == nil || !builder.closure.acceptsArity(1) {
		return v.error(e.node, "recursive builder requires a one-slot Type constructor")
	}
	if count > int64(v.limits.NativeNodes-v.nodes) {
		return v.error(e.node, "native construction limit exceeded")
	}
	parent := v.construction
	if parent == nil {
		v.construction = &typefnConstructionArena{}
	}
	arena := v.construction
	if parent == nil {
		defer func() { v.construction = nil; v.liveBytes -= arena.bytes }()
	}
	storage := int(count) * (2*int(unsafe.Sizeof((*Type)(nil))) + int(unsafe.Sizeof(TupleElementInfo{})))
	if !v.reserveValues(e.node, storage) {
		return v.c.errorType
	}
	arena.bytes += storage
	base := len(arena.definitions)
	arena.definitions = append(arena.definitions, make([]*Type, int(count))...)
	references := make([]*Type, int(count))
	for i := range references {
		if !v.step(e.node) {
			return v.c.errorType
		}
		references[i] = v.construct(e.node, &typefnConstruction{kind: ast.KindTypeReference, reference: base + i})
		if v.failed {
			return v.c.errorType
		}
	}
	argument := references[0]
	if group {
		argument = v.construct(e.node, &typefnConstruction{kind: ast.KindTupleType, inputs: references, elements: slices.Repeat([]TupleElementInfo{{flags: ElementFlagsRequired}}, int(count))})
	}
	result := v.applyConstructor(e.node, builder.closure, []*Type{argument})
	roots := []*Type{result}
	if !v.failed && !v.blocked && !arena.blocked && group {
		if graph := constructionOf(result); graph != nil {
			if !requiredConstructionTuple(graph) {
				return v.error(e.node, "recursive group must return a required tuple of roots")
			}
			roots = graph.inputs
		} else {
			var ok bool
			roots, ok = v.requiredTuple(e.node, result)
			if !ok && !v.blocked {
				return v.c.errorType
			}
		}
		if !v.blocked && len(roots) != int(count) {
			return v.error(e.node, "recursive group must return one root per reference")
		}
	}
	if v.failed {
		return v.c.errorType
	}
	if v.blocked || arena.blocked {
		if parent != nil {
			v.blocked = true
			return v.c.errorType
		}
		v.construction = nil
		v.blocked = false
		return v.deferExpression(e, slots)
	}
	copy(arena.definitions[base:], roots)
	if parent != nil {
		return result
	}
	return v.sealConstruction(e.node, arena, roots, group)
}
