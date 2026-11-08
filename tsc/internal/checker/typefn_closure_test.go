package checker

import "testing"

func TestTypefnClosureSnapshotAndInterning(t *testing.T) {
	c, types := newCompositeTestChecker()
	program := &typefnProgram{}
	captures := []typefnValue{{native: types[0]}}
	first := c.internTypefnClosure(program, captures)
	captures[0].native = types[1]
	second := c.internTypefnClosure(program, captures)
	if first == second || first.captures[0].native != types[0] {
		t.Fatal("capture mutation changed a published closure")
	}
	if c.internTypefnClosure(program, []typefnValue{{native: types[0]}}) != first {
		t.Fatal("equivalent snapshots did not reuse constructor identity")
	}
	if c.internTypefnClosure(&typefnProgram{}, []typefnValue{{native: types[0]}}) == first {
		t.Fatal("different helper code shared constructor identity")
	}
	nested := c.internTypefnClosure(program, []typefnValue{{closure: first}})
	if nested == first || nested.captures[0].closure != first {
		t.Fatal("constructor captures were confused with native Types")
	}
	other, _ := newCompositeTestChecker()
	if other.typefn != nil {
		t.Fatal("closure interning escaped its checker snapshot")
	}
}
