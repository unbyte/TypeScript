package checker

import "testing"

func TestTypefnRecursiveGraphValidation(t *testing.T) {
	for _, test := range []struct {
		name  string
		edges [][]typefnGraphEdge
		want  typefnGraphValidation
	}{
		{"leaf", [][]typefnGraphEdge{nil}, typefnGraphValid},
		{"bare self", [][]typefnGraphEdge{{{0, false}}}, typefnGraphUnguardedCycle},
		{"guarded self", [][]typefnGraphEdge{{{0, true}}}, typefnGraphValid},
		{"mutual aliases", [][]typefnGraphEdge{{{1, false}}, {{0, false}}}, typefnGraphUnguardedCycle},
		{"mutual guarded", [][]typefnGraphEdge{{{1, false}}, {{0, true}}}, typefnGraphValid},
		{"guard outside cycle", [][]typefnGraphEdge{{{1, true}}, {{1, false}}}, typefnGraphUnguardedCycle},
		{"unguarded subcycle in guarded SCC", [][]typefnGraphEdge{{{1, true}}, {{0, false}, {2, false}}, {{1, false}}}, typefnGraphUnguardedCycle},
		{"cross arena", [][]typefnGraphEdge{{{1, true}}}, typefnGraphInvalidEdge},
		{"negative edge", [][]typefnGraphEdge{{{-1, true}}}, typefnGraphInvalidEdge},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, _ := validateTypefnGraph(test.edges, []int{0}, func() bool { return true })
			if got != test.want {
				t.Fatalf("got %v, want %v", got, test.want)
			}
		})
	}
}
func TestTypefnRecursiveGraphDeepAndCanceled(t *testing.T) {
	const count = 100000
	edges := make([][]typefnGraphEdge, count)
	for i := 0; i < count-1; i++ {
		edges[i] = []typefnGraphEdge{{i + 1, false}}
	}
	edges[count-1] = []typefnGraphEdge{{0, true}}
	got, reachable := validateTypefnGraph(edges, []int{0}, func() bool { return true })
	if got != typefnGraphValid || len(reachable) != count {
		t.Fatalf("deep graph: %v, %d", got, len(reachable))
	}
	for _, limit := range []int{0, count, count*2 + 3} {
		steps := 0
		got, reachable = validateTypefnGraph(edges, []int{0}, func() bool { steps++; return steps <= limit })
		if got != typefnGraphCanceled || reachable != nil {
			t.Fatalf("cancellation at %d: %v", limit, got)
		}
	}
}
func TestTypefnRecursiveGraphOnlyPublishesReachableNodes(t *testing.T) {
	edges := [][]typefnGraphEdge{nil, {{1, false}}}
	got, reachable := validateTypefnGraph(edges, []int{0}, func() bool { return true })
	if got != typefnGraphValid || len(reachable) != 1 || reachable[0] != 0 {
		t.Fatal(got, reachable)
	}
	got, _ = validateTypefnGraph(edges, []int{0, 1}, func() bool { return true })
	if got != typefnGraphUnguardedCycle {
		t.Fatal("second root escaped validation")
	}
}
