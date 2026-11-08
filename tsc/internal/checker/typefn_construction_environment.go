package checker

import (
	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"slices"
)

type typefnConstructionEnvironment struct {
	leaves     []*Type
	leafIndex  map[*Type]int
	parameters [][]int
	owners     []int
	bytes      int
}

func (p *typefnConstructionEnvironment) reserve(v *typefnVM, node *ast.Node, size int) bool {
	if !v.reserveValues(node, size) {
		return false
	}
	p.bytes += size
	return true
}

// All nodes in a recursive component share one leaf environment. Downstream
// environments are unions over the component DAG, never recursive expansion of
// native Types. Sorted leaf IDs make projections and parameter order deterministic.
// The caller retains and releases the plan's live storage through native sealing.
func (v *typefnVM) constructionEnvironment(node *ast.Node, arena *typefnConstructionArena, edges [][]typefnGraphEdge, reachable []int) *typefnConstructionEnvironment {
	plan := &typefnConstructionEnvironment{}
	// Conservative backing-storage bounds for iterative component traversal and
	// the per-node environment view. Sets/vectors are charged as entries are added.
	if !plan.reserve(v, node, 192*len(edges)) {
		return plan
	}
	for _, index := range reachable {
		if !plan.reserve(v, node, 24*len(edges[index])) {
			return plan
		}
	}
	components, owners, ok := typefnGraphComponents(edges, reachable, func() bool { return v.step(node) })
	if !ok {
		return plan
	}
	plan.owners = owners
	plan.leafIndex = map[*Type]int{}
	direct := make([][]int, len(edges))
	for _, index := range reachable {
		graph := constructionOf(arena.nodes[index])
		inputs := graph.inputs
		if graph.kind == ast.KindTypeReference {
			inputs = []*Type{arena.definitions[graph.reference]}
		}
		for _, input := range inputs {
			if !v.step(node) {
				return plan
			}
			if constructionOf(input) != nil {
				continue
			}
			id, found := plan.leafIndex[input]
			if !found {
				if !plan.reserve(v, node, 64) {
					return plan
				}
				id = len(plan.leaves)
				plan.leafIndex[input] = id
				plan.leaves = append(plan.leaves, input)
			}
			if !plan.reserve(v, node, 16) {
				return plan
			}
			direct[index] = append(direct[index], id)
		}
	}
	plan.parameters = make([][]int, len(edges))
	for i := len(components) - 1; i >= 0; i-- {
		set := map[int]bool{}
		add := func(id int) bool {
			if !v.step(node) {
				return false
			}
			if !set[id] {
				if !plan.reserve(v, node, 64) {
					return false
				}
				set[id] = true
			}
			return true
		}
		for _, index := range components[i] {
			for _, id := range direct[index] {
				if !add(id) {
					return plan
				}
			}
			for _, edge := range edges[index] {
				if !v.step(node) {
					return plan
				}
				if owners[edge.target] == i+1 {
					continue
				}
				for _, id := range plan.parameters[edge.target] {
					if !add(id) {
						return plan
					}
				}
			}
		}
		parameters := make([]int, 0, len(set))
		for id := range set {
			parameters = append(parameters, id)
		}
		slices.Sort(parameters)
		for _, index := range components[i] {
			plan.parameters[index] = parameters
		}
	}
	return plan
}
