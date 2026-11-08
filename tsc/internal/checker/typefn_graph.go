package checker

// Recursive construction uses an arena separate from native Types and caches.
// A structural edge guards recursion; aliases, unions and intersections do not.
// Validation removes guarded edges and requires the remaining graph to be a DAG.
// This is stronger than checking that each SCC has at least one guarded edge:
// an SCC can also contain an independently unguarded cycle.
type typefnGraphEdge struct {
	target  int
	guarded bool
}

type typefnGraphValidation uint8

const (
	typefnGraphValid typefnGraphValidation = iota
	typefnGraphInvalidEdge
	typefnGraphUnguardedCycle
	typefnGraphCanceled
)

// All roots/edges use arena-local indices. The caller supplies a work/cancellation
// gate so long chains never depend on the Go stack or evade the invocation budget.
// Unreachable scratch nodes are ignored and cannot become published roots.
func validateTypefnGraph(edges [][]typefnGraphEdge, roots []int, step func() bool) (typefnGraphValidation, []int) {
	seen := make([]bool, len(edges))
	work := append([]int(nil), roots...)
	var reachable []int
	for len(work) > 0 {
		if !step() {
			return typefnGraphCanceled, nil
		}
		index := work[len(work)-1]
		work = work[:len(work)-1]
		if index < 0 || index >= len(edges) {
			return typefnGraphInvalidEdge, nil
		}
		if seen[index] {
			continue
		}
		seen[index] = true
		reachable = append(reachable, index)
		for _, edge := range edges[index] {
			if !step() {
				return typefnGraphCanceled, nil
			}
			work = append(work, edge.target)
		}
	}
	type frame struct{ node, next int }
	color := make([]uint8, len(edges))
	var stack []frame
	for _, root := range reachable {
		if color[root] != 0 {
			continue
		}
		color[root] = 1
		stack = append(stack, frame{node: root})
		for len(stack) > 0 {
			if !step() {
				return typefnGraphCanceled, nil
			}
			top := &stack[len(stack)-1]
			if top.next == len(edges[top.node]) {
				color[top.node] = 2
				stack = stack[:len(stack)-1]
				continue
			}
			edge := edges[top.node][top.next]
			top.next++
			if edge.guarded {
				continue
			}
			switch color[edge.target] {
			case 1:
				return typefnGraphUnguardedCycle, nil
			case 0:
				color[edge.target] = 1
				stack = append(stack, frame{node: edge.target})
			}
		}
	}
	return typefnGraphValid, reachable
}
