package scanner

import (
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/stringutil"
)

// DecodedStringRange maps UTF-16 offsets in a quoted string back to the bytes of
// its source spelling. Escape sequences use the same decoder as token scanning.
func DecodedStringRange(token string, start, end int) (core.TextRange, bool) {
	if len(token) < 2 || (token[0] != '"' && token[0] != '\'') || token[len(token)-1] != token[0] || start < 0 || end < start {
		return core.TextRange{}, false
	}
	s := NewScanner()
	s.SetText(token)
	s.ResetPos(1)
	units, first := 0, -1
	for s.pos < len(token) {
		if units == start {
			first = s.pos
		}
		if units == end {
			return core.NewTextRange(first, s.pos), first >= 0
		}
		if s.pos == len(token)-1 {
			break
		}
		if token[s.pos] == '\\' {
			decoded := s.scanEscapeSequence(EscapeSequenceScanningFlagsString)
			for len(decoded) != 0 {
				r, size := stringutil.DecodeJSStringRune(decoded)
				units++
				if r > 0xffff {
					units++
				}
				decoded = decoded[size:]
			}
		} else {
			r, size := stringutil.DecodeJSStringRune(token[s.pos:])
			s.pos += size
			units++
			if r > 0xffff {
				units++
			}
		}
	}
	return core.TextRange{}, false
}
