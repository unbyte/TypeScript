package checker

import "github.com/microsoft/TypeScript/tsc/internal/ast"

// Editor/native API queries may resolve a Type without checking its source
// element first. Keep the outer use while nested aliases instantiate DSL
// requirements. Scope cleanup stays off the ordinary native resolver path.
func (c *Checker) getTypeFromTypeNodeAtTypefnSite(node *ast.Node) *Type {
	c.currentNode = node
	defer func() { c.currentNode = nil }()
	return c.getConditionalFlowTypeOfType(c.getTypeFromTypeNodeWorker(node), node)
}

func (c *Checker) typefnFailureEpoch() uint64 {
	if c.typefn == nil {
		return 0
	}
	return c.typefn.failureEpoch
}

// A native operation may contain a complete root evaluation, or run within one.
// Check both so that unwinding the root cannot erase a child's failed status.
func (c *Checker) typefnFailedSince(epoch uint64) bool {
	return c.typefn != nil && (c.typefn.failureEpoch != epoch || c.typefn.vm != nil && c.typefn.vm.failed)
}
