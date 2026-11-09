package checker

// A native type parameter may explicitly promise this exact deferred bound.
// Following that declaration is a scoped assignability proof, not evaluation of
// the bound and not a claim that the parameter is identical to its constraint.
// Concrete arguments still instantiate and validate the original computation.
func (c *Checker) typefnHasDeclaredConstraint(source, target *Type, relation *Relation) bool {
	if relation == c.identityRelation || target.flags&TypeFlagsTypefn == 0 || source.flags&TypeFlagsTypeParameter == 0 {
		return false
	}
	failureEpoch := c.typefnFailureEpoch()
	seen := map[*Type]bool{}
	for source.flags&TypeFlagsTypeParameter != 0 && !seen[source] {
		if !c.pollTypefnNativeWork() {
			return false
		}
		seen[source] = true
		constraint := c.getConstraintOfTypeParameter(source)
		if c.typefnFailedSince(failureEpoch) {
			return false
		}
		if constraint == target {
			return true
		}
		if constraint == nil {
			return false
		}
		source = constraint
	}
	return false
}
