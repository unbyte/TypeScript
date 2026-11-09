package checker

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
