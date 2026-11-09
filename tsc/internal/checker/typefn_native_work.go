package checker

import "slices"

// Native relations, substitution and normalization may do substantial work inside
// a single VM opcode. Poll internally, while keeping native projects off this path.
// These visits are not yet a complete quota: remaining adapters and primitive
// internals still need accounting before nativeWork overrides can open.
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

type typefnUnionSortCanceled struct{}

// Only the normalizer's private temporary slice is sorted here. The sentinel
// unwinds the library sort, not any native resolution/mapper frames. Its caller
// discards the incomplete set and releases the buffer through its normal defer.
func (c *Checker) sortTypefnUnion(types []*Type) (complete bool) {
	defer func() {
		if value := recover(); value != nil {
			if _, canceled := value.(typefnUnionSortCanceled); !canceled {
				panic(value)
			}
		}
	}()
	slices.SortStableFunc(types, func(a, b *Type) int {
		if !c.pollTypefnNativeWork() {
			panic(typefnUnionSortCanceled{})
		}
		return CompareTypes(a, b)
	})
	return true
}
