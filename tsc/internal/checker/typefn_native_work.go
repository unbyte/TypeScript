package checker

// Native relation traversal may do substantial work inside a single VM opcode.
// Poll within that traversal, while keeping native projects off this path. These
// visits are not yet a complete native-work quota: normalization, instantiation
// and primitive internals need their own accounting before overrides can open.
func (c *Checker) pollTypefnNativeRelation() bool {
	if c.typefn == nil || c.typefn.vm == nil {
		return true
	}
	v := c.typefn.vm
	if v.nativeCanceled {
		return false
	}
	v.nativeRelationSteps++
	if (v.nativeRelationSteps == 1 || v.nativeRelationSteps%1024 == 0) && c.isCanceled() {
		v.nativeCanceled = true
		v.failed = true
		return false
	}
	return true
}

func (c *Checker) typefnNativeCanceled() bool {
	return c.typefn != nil && c.typefn.vm != nil && c.typefn.vm.nativeCanceled
}
