// Package typefunction describes the primitives in a type function body.
package typefunction

import (
	"slices"
	"strings"
)

type Builtin struct {
	Name                       string
	MinArguments, MaxArguments int
}

var builtins = []Builtin{
	{"array.compose", 1, 1},
	{"array.decompose", 1, 1},
	{"bigint.add", 2, 2},
	{"bigint.divide", 2, 2},
	{"bigint.fromDecimal", 1, 1},
	{"bigint.fromNumber", 1, 1},
	{"bigint.lessThan", 2, 2},
	{"bigint.multiply", 2, 2},
	{"bigint.negate", 1, 1},
	{"bigint.remainder", 2, 2},
	{"bigint.subtract", 2, 2},
	{"enumType.members", 1, 1},
	{"fn.compose", 1, 1},
	{"fn.decompose", 1, 1},
	{"fn.rebind", 2, 2},
	{"fn.ref", 1, 1},
	{"fn.specialize", 2, 2},
	{"intersection.fromTuple", 1, 1},
	{"intersection.parts", 1, 1},
	{"iter.protocol", 2, 2},
	{"number.add", 2, 2},
	{"number.divide", 2, 2},
	{"number.fromBigInt", 1, 1},
	{"number.fromDecimal", 1, 1},
	{"number.lessThan", 2, 2},
	{"number.multiply", 2, 2},
	{"number.negate", 1, 1},
	{"number.remainder", 2, 2},
	{"number.subtract", 2, 2},
	{"number.trunc", 1, 1},
	{"object.addField", 3, 3},
	{"object.fields", 1, 1},
	{"object.getField", 2, 2},
	{"object.hasField", 2, 2},
	{"object.indexes", 1, 1},
	{"object.removeField", 2, 2},
	{"object.setField", 3, 3},
	{"object.withFields", 2, 2},
	{"object.withIndexes", 2, 2},
	{"string.codePoint", 1, 1},
	{"string.find", 2, 2},
	{"string.fromCodePoint", 1, 1},
	{"string.length", 1, 1},
	{"string.match", 2, 2},
	{"string.read", 1, 1},
	{"string.slice", 3, 3},
	{"string.split", 2, 2},
	{"template.compose", 1, 1},
	{"template.decompose", 1, 1},
	{"tuple.compose", 1, 1},
	{"tuple.count", 1, 1},
	{"tuple.decompose", 1, 1},
	{"type.baseOfLiteral", 1, 1},
	{"type.isAssignableTo", 2, 2},
	{"type.isIdentical", 2, 2},
	{"type.kind", 1, 1},
	{"union.toTuple", 1, 2},
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
