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
	type dependencyKey struct {
		value     *Type
		signature *Signature
	}
	seen := map[dependencyKey]bool{}
	bound := map[*Type]int{}
	var scope *Signature
	charged := 0
	defer func() { v.liveBytes -= charged }()
	var visit func(*Type) bool
	visit = func(t *Type) bool {
		key := dependencyKey{t, scope}
		if t == nil || bound[t] != 0 || seen[key] {
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
		seen[key] = true
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
			for _, kind := range []SignatureKind{SignatureKindCall, SignatureKindConstruct} {
				for _, signature := range v.c.getSignaturesOfType(t, kind) {
					if !v.step(node) {
						return true
					}
					outer := scope
					scope = signature
					for _, parameter := range signature.TypeParameters() {
						bound[parameter]++
					}
					unresolved := v.c.typefnSignatureDependencies(signature, visit)
					for _, parameter := range signature.TypeParameters() {
						bound[parameter]--
					}
					scope = outer
					if unresolved {
						return true
					}
				}
			}
			for _, info := range v.c.getIndexInfosOfType(t) {
				if visit(info.keyType) || visit(info.valueType) {
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

// Islands currently defer on symbolic captures. Constructor environments use the
// same policy, with an iterative walk over immutable capture DAGs.
func (v *typefnVM) unresolvedValue(node *ast.Node, root typefnValue) bool {
	if root.native != nil {
		return v.unresolvedGraph(node, root.native)
	}
	if root.closure != nil && len(root.closure.captures) == 0 {
		return false
	}
	work := []typefnValue{root}
	seen := map[*typefnClosure]bool{}
	for len(work) != 0 {
		value := work[len(work)-1]
		work = work[:len(work)-1]
		if !v.step(node) {
			return true
		}
		if value.native != nil {
			if v.unresolvedGraph(node, value.native) {
				return true
			}
		} else if value.closure != nil && !seen[value.closure] {
			seen[value.closure] = true
			work = append(work, value.closure.captures...)
		}
	}
	return false
}
