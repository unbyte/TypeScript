package checker

import "github.com/microsoft/TypeScript/tsc/internal/ast"

// Observe dependency summaries, not every latent property of a native graph.
// Native references expose their arguments; anonymous native Types expose their
// referenced lexical binders and mapper. Only VM-built objects need a member walk:
// their members were already evaluated when the object was constructed.
func (v *typefnVM) unresolvedGraph(node *ast.Node, root *Type) bool {
	// Scalar-heavy parsers should not allocate a graph traversal for each digit
	// comparison. Their complete dependency check is already the outer-Type check.
	if root.flags&(TypeFlagsObject|TypeFlagsUnionOrIntersection) == 0 {
		return typefnUnresolved(root)
	}
	seen := map[*Type]bool{}
	charged := 0
	defer func() { v.liveBytes -= charged }()
	var visit func(*Type) bool
	visit = func(t *Type) bool {
		if t == nil || seen[t] {
			return false
		}
		if !v.step(node) {
			return true
		}
		if 32 > v.limits.LiveBytes-v.liveBytes {
			v.error(node, "live VM allocation limit exceeded")
			return true
		}
		v.liveBytes += 32
		charged += 32
		seen[t] = true
		t = v.observe(t)
		if t.flags&TypeFlagsTypeParameter != 0 && t.AsTypeParameter().isThisType {
			return false
		}
		if typefnUnresolved(t) {
			return true
		}
		if t.flags&TypeFlagsUnionOrIntersection != 0 {
			for _, part := range t.Types() {
				if visit(part) {
					return true
				}
			}
			return false
		}
		if t.flags&TypeFlagsObject == 0 {
			return false
		}
		if v.c.typefn != nil && v.c.typefn.objects[t] {
			for _, property := range v.c.getPropertiesOfType(t) {
				if visit(v.c.getTypeOfSymbol(property)) {
					return true
				}
			}
			return false
		}
		if t.objectFlags&ObjectFlagsReference != 0 {
			for _, argument := range v.c.getTypeArguments(t) {
				if visit(argument) {
					return true
				}
			}
			return false
		}
		if t.symbol != nil {
			for _, declaration := range t.symbol.Declarations() {
				for _, parameter := range v.c.getOuterTypeParameters(declaration, false) {
					if !v.c.isTypeParameterPossiblyReferenced(parameter, declaration) {
						continue
					}
					value := v.c.instantiateType(parameter, t.Mapper())
					if visit(value) {
						return true
					}
				}
			}
		}
		return false
	}
	return visit(root)
}
