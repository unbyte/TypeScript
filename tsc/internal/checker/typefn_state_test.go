package checker

import "testing"

func TestTypefnStateOwnership(t *testing.T) {
	first, types := newCompositeTestChecker()
	first.getUnionType(types)
	if first.typefn != nil {
		t.Fatal("native operations allocated DSL state")
	}
	second, _ := newCompositeTestChecker()
	state := first.getTypefnState()
	if state != first.getTypefnState() {
		t.Fatal("checker did not retain its state")
	}
	if second.typefn != nil || second.getTypefnState() == state {
		t.Fatal("state shared across checker snapshots")
	}
}
