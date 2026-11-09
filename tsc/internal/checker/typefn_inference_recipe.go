package checker

import "github.com/microsoft/TypeScript/tsc/internal/ast"

// Instantiated signatures are shared across calls. Default provenance is not:
// an inferred argument can have the same Type as an unused default.
type typefnCallDefaults struct {
	signature *Signature
	used      []bool
}

type typefnCallResolution struct {
	node     *ast.Node
	previous *typefnCallDefaults
	selected *typefnCallDefaults
}

func (c *Checker) beginTypefnCallResolution(node *ast.Node) int {
	state := c.getTypefnState()
	index := len(state.callResolutions)
	state.callResolutions = append(state.callResolutions, typefnCallResolution{node: node, previous: state.callDefaults[node]})
	return index
}

func (c *Checker) endTypefnCallResolution(index int) {
	state := c.typefn
	state.callResolutions[index] = typefnCallResolution{}
	state.callResolutions = state.callResolutions[:index]
}

func (c *Checker) recordTypefnCallDefaults(node *ast.Node, signature *Signature, context *InferenceContext) {
	state := c.typefn
	if state == nil || context == nil || !ast.IsCallExpression(node) {
		return
	}
	for index := len(state.callResolutions) - 1; index >= 0; index-- {
		frame := &state.callResolutions[index]
		if frame.node != node {
			continue
		}
		frame.selected = nil
		for i, inference := range context.inferences {
			if inference.fromDefault {
				if frame.selected == nil {
					frame.selected = &typefnCallDefaults{signature, make([]bool, len(context.inferences))}
				}
				frame.selected.used[i] = true
			}
		}
		return
	}
}

func (c *Checker) publishTypefnCallResolution(index int, signature *Signature, resolvedElsewhere bool) {
	if index < 0 {
		return
	}
	state := c.typefn
	frame := state.callResolutions[index]
	if resolvedElsewhere {
		// A nested resolution owns its published provenance. An error signature
		// can also be cached early without publishing anything: discard stale facts.
		if state.callDefaults[frame.node] == frame.previous {
			delete(state.callDefaults, frame.node)
		}
		return
	}
	if frame.selected == nil || frame.selected.signature != signature {
		delete(state.callDefaults, frame.node)
		return
	}
	if state.callDefaults == nil {
		state.callDefaults = map[*ast.Node]*typefnCallDefaults{}
	}
	state.callDefaults[frame.node] = frame.selected
}

func (c *Checker) typefnCallDefaultOrigin(node *ast.Node) *typefnCallDefaults {
	if c.typefn == nil {
		return nil
	}
	return c.typefn.callDefaults[node]
}

func (c *Checker) restoreTypefnCallDefaultOrigin(node *ast.Node, origin *typefnCallDefaults) {
	if origin == nil {
		if c.typefn != nil {
			delete(c.typefn.callDefaults, node)
		}
		return
	}
	state := c.getTypefnState()
	if state.callDefaults == nil {
		state.callDefaults = map[*ast.Node]*typefnCallDefaults{}
	}
	state.callDefaults[node] = origin
}
