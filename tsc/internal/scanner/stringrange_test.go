package scanner

import "testing"

func TestDecodedStringRange(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, token, want string
		start, end        int
	}{
		{"plain", `"common"`, "mm", 2, 4},
		{"escaped operator after astral escape", `"\u{1f600}\x2b"`, `\x2b`, 2, 3},
		{"astral literal", `"a😀b"`, "😀", 1, 3},
		{"lone surrogate", `"\uD800x"`, `\uD800`, 0, 1},
		{"escaped pair", `"\ud83d\ude00x"`, `\ud83d\ude00`, 0, 2},
		{"end of input", `''`, "", 0, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			loc, ok := DecodedStringRange(tc.token, tc.start, tc.end)
			if !ok || tc.token[loc.Pos():loc.End()] != tc.want {
				t.Fatalf("range for %s[%d:%d] = %v, %t", tc.token, tc.start, tc.end, loc, ok)
			}
		})
	}
}
