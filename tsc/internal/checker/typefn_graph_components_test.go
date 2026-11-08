package checker

import (
	"math/rand/v2"
	"testing"
)

func TestTypefnGraphComponentsAgainstReachability(t *testing.T) {
	random := rand.New(rand.NewPCG(1, 2))
	for sample := 0; sample < 200; sample++ {
		count := 1 + random.IntN(20)
		edges := make([][]typefnGraphEdge, count)
		roots := make([]int, count)
		closure := make([][]bool, count)
		for i := range count {
			roots[i] = i
			closure[i] = make([]bool, count)
			closure[i][i] = true
			for j := range count {
				if random.IntN(6) == 0 {
					edges[i] = append(edges[i], typefnGraphEdge{target: j, guarded: random.IntN(2) == 0})
					closure[i][j] = true
				}
			}
		}
		// Independent transitive-closure oracle; no DFS/component implementation.
		for k := range count {
			for i := range count {
				for j := range count {
					closure[i][j] = closure[i][j] || closure[i][k] && closure[k][j]
				}
			}
		}
		groups, ids, ok := typefnGraphComponents(edges, roots, func() bool { return true })
		if !ok {
			t.Fatal("unexpected cancellation")
		}
		seen := make([]bool, count)
		for i, group := range groups {
			for _, node := range group {
				if seen[node] || ids[node] != i+1 {
					t.Fatal("invalid component membership")
				}
				seen[node] = true
			}
		}
		for i := range count {
			if !seen[i] {
				t.Fatal("lost reachable node")
			}
			for j := range count {
				if (ids[i] == ids[j]) != (closure[i][j] && closure[j][i]) {
					t.Fatalf("sample %d: incorrect component for %d, %d", sample, i, j)
				}
			}
			for _, edge := range edges[i] {
				if ids[i] > ids[edge.target] {
					t.Fatal("dependency order is not topological")
				}
			}
		}
	}
}
func TestTypefnGraphComponentsDeepAndCanceled(t *testing.T) {
	const count = 100000
	edges := make([][]typefnGraphEdge, count+1)
	roots := make([]int, count)
	for i := range count {
		roots[i] = i
		if i+1 < count {
			edges[i] = []typefnGraphEdge{{target: i + 1}}
		}
	}
	groups, ids, ok := typefnGraphComponents(edges, roots, func() bool { return true })
	if !ok || len(groups) != count || ids[count] != 0 {
		t.Fatal("deep traversal or unreachable node")
	}
	for _, limit := range []int{0, count, 3 * count, 5 * count} {
		steps := 0
		groups, ids, ok = typefnGraphComponents(edges, roots, func() bool { steps++; return steps <= limit })
		if ok || groups != nil || ids != nil {
			t.Fatalf("partial component output after cancellation at %d", limit)
		}
	}
}
