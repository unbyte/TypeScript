// Package typefunction describes the primitives in a type function body.
package typefunction

import (
	"slices"
	"strings"
)

type Builtin struct {
	Name, Parameters           string
	MinArguments, MaxArguments int
	Documentation              string
}

var builtins = []Builtin{
	{"array.compose", "P", 1, 1, "Builds an array type from an element type and a readonly flag."},
	{"array.decompose", "A", 1, 1, "Returns the element type and readonly flag of an Array or ReadonlyArray. Use tuple.decompose for tuples."},
	{"bigint.add", "A, B", 2, 2, "Returns the exact sum of two bigint literal types."},
	{"bigint.divide", "A, B", 2, 2, "Returns the bigint quotient of A divided by B, truncated toward zero. A zero divisor fails."},
	{"bigint.fromNumber", "N", 1, 1, "Converts a safe integer number literal to a bigint literal."},
	{"bigint.lessThan", "A, B", 2, 2, "Returns true if bigint literal A is less than B."},
	{"bigint.multiply", "A, B", 2, 2, "Returns the exact product of two bigint literal types."},
	{"bigint.negate", "A", 1, 1, "Returns the bigint literal with the opposite sign."},
	{"bigint.parse", "S", 1, 1, "Parses a decimal string literal as a bigint literal. An optional leading sign is allowed. Whitespace and the n suffix are invalid."},
	{"bigint.remainder", "A, B", 2, 2, "Returns the bigint remainder after dividing A by B. Its sign follows A. A zero divisor fails."},
	{"bigint.subtract", "A, B", 2, 2, "Returns the exact difference A minus B as a bigint literal."},
	{"enumType.members", "E", 1, 1, "Returns a record keyed by the declared names of an enum object. Each entry describes the member type and its constant or computed value."},
	{"fn.compose", "Parts", 1, 1, "Builds a function type from its members and ordered tuples of call and construct signatures."},
	{"fn.decompose", "F", 1, 1, "Returns a record with members, calls, and constructs. The signature tuples retain overload order."},
	{"fn.rebind", "Fragment, Pairs", 2, 2, "Replaces references to generic signature parameters in Fragment. Each entry in Pairs specifies a source reference and its replacement type."},
	{"fn.ref", "Key", 1, 1, "Creates a reference to a generic parameter identified by Key for use in a computed signature."},
	{"fn.specialize", "F, Args", 2, 2, "Applies a tuple of type arguments to a single generic call or construct signature, respecting its constraints and defaults."},
	{"intersection.fromTuple", "Items", 1, 1, "Intersects the types in Items in tuple order. An empty tuple yields unknown."},
	{"intersection.toTuple", "T", 1, 1, "Returns the constituents of a normalized intersection in checker order. A single type yields a one-element tuple."},
	{"iter.protocol", "T, Mode", 2, 2, "Returns the yield, return, and next types as a tuple. Mode selects the sync or async iteration protocol."},
	{"number.add", "A, B", 2, 2, "Returns the sum of two number literal types using JavaScript number precision."},
	{"number.divide", "A, B", 2, 2, "Returns A divided by B as a number literal. A non-finite result fails."},
	{"number.fromBigInt", "B", 1, 1, "Converts a bigint in the safe integer range to a number literal."},
	{"number.lessThan", "A, B", 2, 2, "Returns true if number literal A is less than B."},
	{"number.multiply", "A, B", 2, 2, "Returns the product of two number literal types. A non-finite result fails."},
	{"number.negate", "A", 1, 1, "Returns the number literal with the opposite sign."},
	{"number.parse", "S", 1, 1, "Parses an entire decimal string literal as a finite number literal. A sign, fractional part, and exponent are supported. Whitespace is invalid."},
	{"number.remainder", "A, B", 2, 2, "Returns the remainder after dividing A by B, with the sign of A. A zero divisor fails."},
	{"number.subtract", "A, B", 2, 2, "Returns the difference A minus B as a number literal. A non-finite result fails."},
	{"number.trunc", "A", 1, 1, "Truncates a number literal toward zero."},
	{"object.addField", "T, K, D", 3, 3, "Returns a new object type with field K described by D. An existing named field with the same key fails. A data descriptor supplies type and may set optional or readonly."},
	{"object.fields", "T", 1, 1, "Returns a record of named public field descriptors, including inherited fields. Each entry describes the field's type or accessors and its modifiers."},
	{"object.getField", "T, K", 2, 2, "Returns the descriptor of named field K in T. A missing named field fails."},
	{"object.hasField", "T, K", 2, 2, "Returns true if T has the named field K, including an optional field. An index signature alone does not count."},
	{"object.indexes", "T", 1, 1, "Returns the union of index signature descriptors, each with key, type, and readonly fields. An object with no index signatures yields never."},
	{"object.removeField", "T, K", 2, 2, "Returns a new object type without the named field K. An absent key leaves T unchanged."},
	{"object.setField", "T, K, D", 3, 3, "Returns a new object type with field K added or replaced using descriptor D, including its modifiers."},
	{"object.withFields", "T, FieldMap", 2, 2, "Returns an object type whose named public fields are taken from FieldMap, preserving the other representable members of T."},
	{"object.withIndexes", "T, IndexSet", 2, 2, "Returns an object type with its index signatures replaced by IndexSet. Pass never to remove all index signatures."},
	{"string.find", "S, Index, Needle", 3, 3, "Finds the first occurrence of Needle at or after Index. Returns a span with source, start, and exclusive end offsets in UTF-16 code units, or undefined on a miss."},
	{"string.fromCodePoint", "N", 1, 1, "Returns the string literal for an integer code point from 0 to 0x10FFFF."},
	{"string.length", "S", 1, 1, "Returns the length of a string literal in UTF-16 code units."},
	{"string.matchPrefix", "S, Index, Prefix", 3, 3, "Matches Prefix at the given UTF-16 index. Returns a tuple of the matched prefix and next index, or undefined on a miss."},
	{"string.read", "S, Index", 2, 2, "Reads one code point at a UTF-16 index. Returns a tuple of the text and next index, or undefined at the end of the string."},
	{"string.slice", "S, Start, End", 3, 3, "Returns the substring from Start up to, but excluding, End. Both UTF-16 offsets must be on code point boundaries."},
	{"string.split", "S, Separator", 2, 2, "Splits a string literal at each occurrence of a nonempty separator and returns a tuple of string literals, preserving empty pieces."},
	{"string.toCodePoint", "S", 1, 1, "Returns the numeric code point of a string literal containing exactly one code point."},
	{"template.compose", "P", 1, 1, "Interleaves texts and interpolation types to build a template literal type. Supply one more text than interpolation."},
	{"template.decompose", "T", 1, 1, "Returns the texts and interpolation types of a template literal type. A string literal has one text and no interpolations."},
	{"tuple.compose", "P", 1, 1, "Builds a tuple type from its slot descriptors and readonly flag, retaining labels and element kinds."},
	{"tuple.count", "T", 1, 1, "Returns the possible lengths of a tuple. Optional elements can yield a union of lengths, and an unbounded rest element yields number."},
	{"tuple.decompose", "T", 1, 1, "Returns a tuple's ordered slot descriptors and readonly flag. Each slot includes its element type, kind, and label when present."},
	{"type.isAssignableTo", "From, To", 2, 2, "Returns true if From is assignable to To according to the TypeScript checker."},
	{"type.isIdentical", "A, B", 2, 2, "Returns true if A and B are identical according to the TypeScript checker's type identity relation."},
	{"type.kind", "T", 1, 1, "Returns a string literal describing the kind of the whole normalized type. For example, a tuple yields \"tuple\"."},
	{"type.literalBase", "T", 1, 1, "Returns the base type of a literal, such as string for a string literal. Applies the same operation to each constituent of a union."},
	{"union.toTuple", "U, Key<_>", 1, 2, "Returns the normalized union members as a tuple. Literal members of a common supported kind are sorted by value. Supply a unary key constructor for other members to sort by unique literal keys."},
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
