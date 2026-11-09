package checker

// Native relation traversal and substitution may do substantial work inside a
// single VM opcode. Poll internally, while keeping native projects off this path.
// These visits are not yet a complete quota: normalization and primitive internals
// still need accounting before nativeWork overrides can open.
func (c *Checker) pollTypefnNativeWork() bool {
	if c.typefn == nil || c.typefn.vm == nil {
		return true
	}
	v := c.typefn.vm
	if v.nativeCanceled {
		return false
	}
	v.nativeSteps++
	if (v.nativeSteps == 1 || v.nativeSteps%1024 == 0) && c.isCanceled() {
		v.nativeCanceled = true
		v.failed = true
		return false
	}
	return true
}

func (c *Checker) typefnNativeCanceled() bool {
	return c.typefn != nil && c.typefn.vm != nil && c.typefn.vm.nativeCanceled
}
