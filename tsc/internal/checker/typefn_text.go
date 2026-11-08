package checker

import (
	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/jsnum"
	"github.com/microsoft/TypeScript/tsc/internal/stringutil"
	"strings"
)

func (v *typefnVM) stringWork(node *ast.Node, units int) bool {
	if v.failed {
		return false
	}
	if units > v.limits.StringUnits-v.stringUnits {
		v.error(node, "string work limit exceeded")
		return false
	}
	v.stringUnits += units
	if v.c.isCanceled() {
		v.failed = true
		return false
	}
	return true
}

// Decoded operands are root-local and bounded. Repeated cursor peeks do not
// repeatedly scan the same string. Native strings use WTF-8 for lone surrogates;
// Go's utf16.Encode([]rune(s)) would silently destroy those code units.
func (v *typefnVM) stringUnitsOf(node *ast.Node, value *Type) ([]uint16, bool) {
	if value.flags != TypeFlagsStringLiteral {
		v.error(node, "text operand requires one unbranded string literal")
		return nil, false
	}
	s := value.AsLiteralType().value.(string)
	if units, ok := v.textUnits[s]; ok {
		return units, true
	}
	count := 0
	for i := 0; i < len(s); {
		r, width := stringutil.DecodeJSStringRune(s[i:])
		i += width
		n := 1
		if r > 0xffff {
			n = 2
		}
		if !v.stringWork(node, n) {
			return nil, false
		}
		count += n
	}
	if count > (v.limits.LiveBytes-v.liveBytes)/2 {
		v.error(node, "live VM allocation limit exceeded")
		return nil, false
	}
	v.liveBytes += count * 2
	units := make([]uint16, 0, count)
	for i := 0; i < len(s); {
		if len(units)%1024 == 0 && !v.stringWork(node, 0) {
			return nil, false
		}
		r, width := stringutil.DecodeJSStringRune(s[i:])
		i += width
		if r > 0xffff {
			high, low := stringutil.CodePointToSurrogatePair(r)
			units = append(units, uint16(high), uint16(low))
		} else {
			units = append(units, uint16(r))
		}
	}
	if v.textUnits == nil {
		v.textUnits = make(map[string][]uint16)
	}
	v.textUnits[s] = units
	return units, true
}

func (v *typefnVM) textResult(node *ast.Node, units []uint16) *Type {
	if !v.stringWork(node, len(units)) || !v.allocate(node, 1) {
		return v.c.errorType
	}
	// Reserve the maximum UTF-8 buffer size before writing. The temporary builder
	// is released after interning; retained native literals follow checker lifetime.
	if len(units) > (v.limits.LiveBytes-v.liveBytes)/3 {
		return v.error(node, "live VM allocation limit exceeded")
	}
	size := 3 * len(units)
	v.liveBytes += size
	defer func() { v.liveBytes -= size }()
	var b strings.Builder
	b.Grow(size)
	for i := 0; i < len(units); i++ {
		if i%1024 == 0 && !v.stringWork(node, 0) {
			return v.c.errorType
		}
		r := rune(units[i])
		if stringutil.IsHighSurrogate(r) && i+1 < len(units) && stringutil.IsLowSurrogate(rune(units[i+1])) {
			r = stringutil.SurrogatePairToCodePoint(r, rune(units[i+1]))
			i++
		}
		b.WriteString(stringutil.EncodeJSStringRune(r))
	}
	return v.c.getStringLiteralType(b.String())
}

func (v *typefnVM) textIntrinsic(node *ast.Node, name string, args []*Type) *Type {
	if name == "text.fromCodeUnit" {
		n, ok := v.intrinsicOffset(node, args[0])
		if !ok {
			return v.c.errorType
		}
		if n > 65535 {
			return v.error(node, "code unit must be in 0..65535")
		}
		return v.textResult(node, []uint16{uint16(n)})
	}
	units, ok := v.stringUnitsOf(node, args[0])
	if !ok {
		return v.c.errorType
	}
	if !v.stringWork(node, 1) {
		return v.c.errorType
	}
	switch name {
	case "text.length":
		return v.c.getNumberLiteralType(jsnum.Number(len(units)))
	case "text.digit":
		n := -1
		if len(units) == 1 && units[0] >= '0' && units[0] <= '9' {
			n = int(units[0] - '0')
		}
		if n < 0 {
			return v.c.getBigIntLiteralType(jsnum.NewPseudoBigInt("1", true))
		}
		return v.c.getBigIntLiteralType(jsnum.NewPseudoBigInt(string(rune('0'+n)), false))
	}
	start, ok := v.intrinsicOffset(node, args[1])
	if !ok {
		return v.c.errorType
	}
	if start > int64(len(units)) {
		return v.error(node, "text offset exceeds length")
	}
	switch name {
	case "text.peek":
		if start == int64(len(units)) {
			return v.textResult(node, nil)
		}
		return v.textResult(node, units[start:start+1])
	case "text.codeUnit":
		if start == int64(len(units)) {
			return v.error(node, "codeUnit requires an offset before the end")
		}
		return v.c.getNumberLiteralType(jsnum.Number(units[start]))
	case "text.slice":
		end := int64(len(units))
		if len(args) == 3 {
			end, ok = v.intrinsicOffset(node, args[2])
			if !ok {
				return v.c.errorType
			}
		}
		if end < start || end > int64(len(units)) {
			return v.error(node, "text slice requires start <= end <= length")
		}
		return v.textResult(node, units[start:end])
	}
	panic("unhandled text intrinsic")
}
