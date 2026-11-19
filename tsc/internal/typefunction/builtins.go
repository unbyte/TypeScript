// Package typefunction describes the primitives in a type function body.
package typefunction

import (
	"slices"
	"strings"
)

type Builtin struct {
	Name, Parameters           string
	MinArguments, MaxArguments int
}

var builtins = []Builtin{
	{"array.compose", "P", 1, 1},
	{"array.decompose", "A", 1, 1},
	{"bigint.add", "A, B", 2, 2},
	{"bigint.divide", "A, B", 2, 2},
	{"bigint.fromDecimal", "S", 1, 1},
	{"bigint.fromNumber", "N", 1, 1},
	{"bigint.lessThan", "A, B", 2, 2},
	{"bigint.multiply", "A, B", 2, 2},
	{"bigint.negate", "A", 1, 1},
	{"bigint.remainder", "A, B", 2, 2},
	{"bigint.subtract", "A, B", 2, 2},
	{"enumType.members", "E", 1, 1},
	{"fn.compose", "Parts", 1, 1},
	{"fn.decompose", "F", 1, 1},
	{"fn.rebind", "Fragment, Pairs", 2, 2},
	{"fn.ref", "Key", 1, 1},
	{"fn.specialize", "F, Args", 2, 2},
	{"intersection.fromTuple", "Items", 1, 1},
	{"intersection.parts", "T", 1, 1},
	{"iter.protocol", "T, Mode", 2, 2},
	{"number.add", "A, B", 2, 2},
	{"number.divide", "A, B", 2, 2},
	{"number.fromBigInt", "B", 1, 1},
	{"number.fromDecimal", "S", 1, 1},
	{"number.lessThan", "A, B", 2, 2},
	{"number.multiply", "A, B", 2, 2},
	{"number.negate", "A", 1, 1},
	{"number.remainder", "A, B", 2, 2},
	{"number.subtract", "A, B", 2, 2},
	{"number.trunc", "A", 1, 1},
	{"object.addField", "T, K, D", 3, 3},
	{"object.fields", "T", 1, 1},
	{"object.getField", "T, K", 2, 2},
	{"object.hasField", "T, K", 2, 2},
	{"object.indexes", "T", 1, 1},
	{"object.removeField", "T, K", 2, 2},
	{"object.setField", "T, K, D", 3, 3},
	{"object.withFields", "T, FieldMap", 2, 2},
	{"object.withIndexes", "T, IndexSet", 2, 2},
	{"string.codePoint", "S", 1, 1},
	{"string.find", "C, Needle", 2, 2},
	{"string.fromCodePoint", "N", 1, 1},
	{"string.length", "S", 1, 1},
	{"string.match", "C, Pattern", 2, 2},
	{"string.read", "C", 1, 1},
	{"string.slice", "S, Start, End", 3, 3},
	{"string.split", "S, Separator", 2, 2},
	{"template.compose", "P", 1, 1},
	{"template.decompose", "T", 1, 1},
	{"tuple.compose", "P", 1, 1},
	{"tuple.count", "T", 1, 1},
	{"tuple.decompose", "T", 1, 1},
	{"type.baseOfLiteral", "T", 1, 1},
	{"type.isAssignableTo", "From, To", 2, 2},
	{"type.isIdentical", "A, B", 2, 2},
	{"type.kind", "T", 1, 1},
	{"union.toTuple", "U, Key<_>", 1, 2},
}

func Lookup(name string) (Builtin, bool) {
	i, ok := slices.BinarySearchFunc(builtins, name, func(a Builtin, b string) int { return strings.Compare(a.Name, b) })
	if !ok {
		return Builtin{}, false
	}
	return builtins[i], true
}
func Members(namespace string) []string {
	prefix := namespace + "."
	var result []string
	for _, b := range builtins {
		if member, ok := strings.CutPrefix(b.Name, prefix); ok {
			result = append(result, member)
		}
	}
	return result
}
func IsNamespace(name string) bool {
	prefix := name + "."
	i, _ := slices.BinarySearchFunc(builtins, prefix, func(a Builtin, b string) int { return strings.Compare(a.Name, b) })
	return i < len(builtins) && strings.HasPrefix(builtins[i].Name, prefix)
}
