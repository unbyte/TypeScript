package checker

import (
	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/jsnum"
)

// Prefix tables keep searches linear even on adversarial repeated prefixes.
// The table is temporary root-owned storage and is released by the caller.
func (v *typefnVM) textSearcher(node *ast.Node, needle []uint16) ([]int, func(), bool) {
	const wordBytes = 8 // conservative on 32-bit hosts
	if len(needle) > (v.limits.LiveBytes-v.liveBytes)/wordBytes {
		v.error(node, "live VM allocation limit exceeded")
		return nil, func() {}, false
	}
	size := len(needle) * wordBytes
	v.liveBytes += size
	release := func() { v.liveBytes -= size }
	table := make([]int, len(needle))
	for i, j := 1, 0; i < len(needle); i++ {
		for j > 0 {
			if !v.stringWork(node, 1) {
				release()
				return nil, func() {}, false
			}
			if needle[i] == needle[j] {
				break
			}
			j = table[j-1]
		}
		if !v.stringWork(node, 1) {
			release()
			return nil, func() {}, false
		}
		if needle[i] == needle[j] {
			j++
		}
		table[i] = j
	}
	return table, release, true
}

func (v *typefnVM) textIndex(node *ast.Node, text, needle []uint16, table []int, start int) int {
	if len(needle) == 0 {
		return start
	}
	for i, j := start, 0; i < len(text); i++ {
		for j > 0 {
			if !v.stringWork(node, 1) {
				return -1
			}
			if text[i] == needle[j] {
				break
			}
			j = table[j-1]
		}
		if !v.stringWork(node, 1) {
			return -1
		}
		if text[i] == needle[j] {
			j++
		}
		if j == len(needle) {
			return i + 1 - j
		}
	}
	return -1
}

func (v *typefnVM) textBuffer(node *ast.Node, length int) ([]uint16, func(), bool) {
	if length > (v.limits.LiveBytes-v.liveBytes)/2 {
		v.error(node, "live VM allocation limit exceeded")
		return nil, func() {}, false
	}
	if length > v.limits.StringUnits-v.stringUnits {
		v.error(node, "string work limit exceeded")
		return nil, func() {}, false
	}
	v.liveBytes += 2 * length
	return make([]uint16, length), func() { v.liveBytes -= 2 * length }, true
}

func (v *typefnVM) textConcat(node *ast.Node, arg *Type) *Type {
	if !isTupleType(arg) || arg.TargetTupleType().combinedFlags & ^ElementFlagsRequired != 0 {
		return v.error(node, "concat requires a required finite tuple of string literals")
	}
	elements := v.c.getTypeArguments(arg)
	total := 0
	for _, element := range elements {
		element = v.observe(element)
		if typefnUnresolved(element) {
			v.blocked = true
			return v.c.errorType
		}
		part, ok := v.stringUnitsOf(node, element)
		if !ok {
			return v.c.errorType
		}
		if len(part) > v.limits.StringUnits-total {
			return v.error(node, "string work limit exceeded")
		}
		total += len(part)
	}
	buffer, release, ok := v.textBuffer(node, total)
	if !ok {
		return v.c.errorType
	}
	defer release()
	offset := 0
	for _, element := range elements {
		part, _ := v.stringUnitsOf(node, v.observe(element))
		if !v.stringWork(node, len(part)) {
			return v.c.errorType
		}
		offset += copy(buffer[offset:], part)
	}
	return v.textResult(node, buffer)
}

func (v *typefnVM) textSearchIntrinsic(node *ast.Node, name string, text []uint16, args []*Type) *Type {
	needle, ok := v.stringUnitsOf(node, args[1])
	if !ok {
		return v.c.errorType
	}
	if name == "text.startsWith" || name == "text.endsWith" {
		if len(needle) > len(text) {
			return v.boolType(false)
		}
		start := 0
		if name == "text.endsWith" {
			start = len(text) - len(needle)
		}
		for i, unit := range needle {
			if !v.stringWork(node, 1) {
				return v.c.errorType
			}
			if text[start+i] != unit {
				return v.boolType(false)
			}
		}
		return v.boolType(true)
	}
	start := int64(0)
	if name == "text.indexOf" && len(args) == 3 {
		start, ok = v.intrinsicOffset(node, args[2])
		if !ok {
			return v.c.errorType
		}
		if start > int64(len(text)) {
			return v.error(node, "text offset exceeds length")
		}
	}
	if (name == "text.split" || name == "text.replaceAll") && len(needle) == 0 {
		return v.error(node, "split/replaceAll requires a nonempty delimiter")
	}
	table, release, ok := v.textSearcher(node, needle)
	if !ok {
		return v.c.errorType
	}
	defer release()
	if name == "text.indexOf" || name == "text.includes" {
		index := v.textIndex(node, text, needle, table, int(start))
		if v.failed {
			return v.c.errorType
		}
		if name == "text.includes" {
			return v.boolType(index >= 0)
		}
		return v.c.getNumberLiteralType(jsnum.Number(index))
	}
	var replacement []uint16
	if name == "text.replaceAll" {
		replacement, ok = v.stringUnitsOf(node, args[2])
		if !ok {
			return v.c.errorType
		}
	}
	// Count first to bound allocations before constructing output. Two linear
	// passes avoid unaccounted growing buffers, even for many empty split parts.
	matches := 0
	for offset := 0; ; {
		index := v.textIndex(node, text, needle, table, offset)
		if v.failed {
			return v.c.errorType
		}
		if index < 0 {
			break
		}
		matches++
		offset = index + len(needle)
	}
	if name == "text.split" {
		if !v.allocate(node, matches+2) {
			return v.c.errorType
		}
		if matches+1 > (v.limits.LiveBytes-v.liveBytes)/8 {
			return v.error(node, "live VM allocation limit exceeded")
		}
		size := (matches + 1) * 8
		v.liveBytes += size
		defer func() { v.liveBytes -= size }()
		values := make([]*Type, matches+1)
		offset := 0
		for i := range values {
			end := len(text)
			if i < matches {
				end = v.textIndex(node, text, needle, table, offset)
			}
			if v.failed {
				return v.c.errorType
			}
			values[i] = v.textResult(node, text[offset:end])
			if v.failed {
				return v.c.errorType
			}
			offset = end + len(needle)
		}
		return v.c.createTupleType(values)
	}
	retained := len(text) - matches*len(needle)
	if matches != 0 && len(replacement) > (v.limits.StringUnits-retained)/matches {
		return v.error(node, "string work limit exceeded")
	}
	length := retained + matches*len(replacement)
	buffer, free, ok := v.textBuffer(node, length)
	if !ok {
		return v.c.errorType
	}
	defer free()
	offset, out := 0, 0
	for i := 0; i < matches; i++ {
		end := v.textIndex(node, text, needle, table, offset)
		if v.failed {
			return v.c.errorType
		}
		if !v.stringWork(node, end-offset+len(replacement)) {
			return v.c.errorType
		}
		out += copy(buffer[out:], text[offset:end])
		out += copy(buffer[out:], replacement)
		offset = end + len(needle)
	}
	if !v.stringWork(node, len(text)-offset) {
		return v.c.errorType
	}
	copy(buffer[out:], text[offset:])
	return v.textResult(node, buffer)
}
