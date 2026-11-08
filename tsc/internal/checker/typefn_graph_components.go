package checker

// Components are returned in source-to-sink order: every inter-component edge
// points to a later component. Guarded edges participate too, because native
// lexical dependencies flow through all structural positions. Inputs must have
// passed graph validation. Zero in componentOf denotes an unreachable node;
// otherwise its value is the component index plus one.
func typefnGraphComponents(edges [][]typefnGraphEdge, reachable []int, step func() bool) (components [][]int, componentOf []int, ok bool) {
	type frame struct{ node, next int }
	seen := make([]bool, len(edges))
	var order []int
	var stack []frame
	for _, root := range reachable {
		if seen[root] {
			continue
		}
		seen[root] = true
		stack = append(stack, frame{node: root})
		for len(stack) > 0 {
			if !step() {
				return nil, nil, false
			}
			top := &stack[len(stack)-1]
			if top.next == len(edges[top.node]) {
				order = append(order, top.node)
				stack = stack[:len(stack)-1]
				continue
			}
			next := edges[top.node][top.next].target
			top.next++
			if !seen[next] {
				seen[next] = true
				stack = append(stack, frame{node: next})
			}
		}
	}
	reverse := make([][]int, len(edges))
	for _, node := range reachable {
		if !step() {
			return nil, nil, false
		}
		for _, edge := range edges[node] {
			if !step() {
				return nil, nil, false
			}
			reverse[edge.target] = append(reverse[edge.target], node)
		}
	}
	componentOf = make([]int, len(edges))
	var work []int
	for i := len(order) - 1; i >= 0; i-- {
		root := order[i]
		if componentOf[root] != 0 {
			continue
		}
		id := len(components) + 1
		componentOf[root] = id
		work = append(work, root)
		var members []int
		for len(work) > 0 {
			if !step() {
				return nil, nil, false
			}
			node := work[len(work)-1]
			work = work[:len(work)-1]
			members = append(members, node)
			for _, previous := range reverse[node] {
				if !step() {
					return nil, nil, false
				}
				if componentOf[previous] == 0 {
					componentOf[previous] = id
					work = append(work, previous)
				}
			}
		}
		components = append(components, members)
	}
	return components, componentOf, true
}
