package checker

import (
	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"slices"
)

func typefnPublicSignature(sig *Signature) bool {
	return sig.declaration == nil || ast.GetCombinedModifierFlags(sig.declaration)&ast.ModifierFlagsNonPublicAccessibilityModifier == 0
}

func (v *typefnVM) signatureObject(node *ast.Node, value *Type) bool {
	if typefnUnresolved(value) || v.c.isGenericObjectType(value) {
		v.blocked = true
		return false
	}
	if value.flags&(TypeFlagsObject|TypeFlagsIntersection) == 0 {
		v.error(node, "signature query requires one object Type")
		return false
	}
	return true
}

func (v *typefnVM) reflectSignatures(node *ast.Node, value *Type, kind SignatureKind) *Type {
	if !v.signatureObject(node, value) {
		return v.c.errorType
	}
	signatures := v.c.getSignaturesOfType(value, kind)
	if !v.allocate(node, 2*len(signatures)+1) {
		return v.c.errorType
	}
	result := make([]*Type, 0, len(signatures))
	for _, sig := range signatures {
		if !v.step(node) {
			return v.c.errorType
		}
		if !typefnPublicSignature(sig) {
			continue
		}
		result = append(result, v.c.typefnSignatureType(sig, kind))
	}
	return v.c.createTupleType(result)
}

func (c *Checker) typefnSignatureType(sig *Signature, kind SignatureKind) *Type {
	var calls, constructs []*Signature
	if kind == SignatureKindCall {
		calls = []*Signature{sig}
	} else {
		constructs = []*Signature{sig}
	}
	return c.newTypefnObjectEx(ast.NewSymbolTable(), calls, constructs, nil)
}

func (v *typefnVM) singleSignature(node *ast.Node, value *Type) (*Signature, SignatureKind, bool) {
	value = v.observe(value)
	if !v.signatureObject(node, value) {
		return nil, 0, false
	}
	calls := v.c.getSignaturesOfType(value, SignatureKindCall)
	constructs := v.c.getSignaturesOfType(value, SignatureKindConstruct)
	if len(calls)+len(constructs) != 1 {
		v.error(node, "signature requires exactly one call or construct signature")
		return nil, 0, false
	}
	sig, kind := (*Signature)(nil), SignatureKindCall
	if len(calls) == 1 {
		sig = calls[0]
	} else {
		sig, kind = constructs[0], SignatureKindConstruct
	}
	if !typefnPublicSignature(sig) {
		v.error(node, "signature must be public")
		return nil, 0, false
	}
	return sig, kind, true
}

func (v *typefnVM) signatureList(node *ast.Node, value *Type, kind SignatureKind) ([]*Signature, bool) {
	entries, ok := v.requiredTuple(node, value)
	if !ok || !v.allocate(node, len(entries)) {
		return nil, false
	}
	result := make([]*Signature, 0, len(entries))
	for _, entry := range entries {
		if !v.step(node) {
			return nil, false
		}
		sig, actual, ok := v.singleSignature(node, entry)
		if !ok {
			return nil, false
		}
		if actual != kind {
			v.error(node, "signature category does not match the call/construct list")
			return nil, false
		}
		result = append(result, sig)
	}
	return result, true
}

func (v *typefnVM) overloadType(node *ast.Node, args []*Type) *Type {
	calls, ok := v.signatureList(node, args[0], SignatureKindCall)
	if !ok {
		return v.c.errorType
	}
	var constructs []*Signature
	if len(args) > 1 {
		constructs, ok = v.signatureList(node, args[1], SignatureKindConstruct)
		if !ok {
			return v.c.errorType
		}
	}
	if !v.allocate(node, 1) {
		return v.c.errorType
	}
	return v.c.newTypefnObjectEx(ast.NewSymbolTable(), calls, constructs, nil)
}

// Native signatures keep their own binders and lazy return/predicate evaluation.
// Follow native instantiation/composition provenance to inspect only free outer
// dependencies, never materializing a generic body's latent typefn computations.
func (c *Checker) typefnSignatureDependencies(sig *Signature, visit func(*Type) bool) bool {
	// Default/inherited class constructors can replace the return Type of their
	// target signature. It is an already-created class graph, not a lazy return
	// annotation. Include it even when target provenance exists.
	if result := sig.resolvedReturnType; result != nil && result.objectFlags&ObjectFlagsClassOrInterface != 0 && visit(result) {
		return true
	}
	if sig.target != nil {
		return c.typefnSignatureDependencies(sig.target, func(t *Type) bool {
			value := c.instantiateType(t, sig.mapper)
			return !slices.Contains(sig.TypeParameters(), value) && visit(value)
		})
	}
	if composite := sig.getComposite(); composite != nil {
		for _, part := range composite.signatures {
			if c.typefnSignatureDependencies(part, visit) {
				return true
			}
		}
		return false
	}
	if sig.declaration != nil {
		for _, parameter := range c.getOuterTypeParameters(sig.declaration, false) {
			if !slices.Contains(sig.TypeParameters(), parameter) && c.isTypeParameterPossiblyReferenced(parameter, sig.declaration) && visit(parameter) {
				return true
			}
		}
		return false
	}
	// Declarationless signatures made by the builder already have evaluated
	// parameter and return Types. Native synthetic signatures use target/composite
	// provenance above rather than expanding their return expressions here.
	for _, parameter := range sig.parameters {
		if visit(c.getTypeOfSymbol(parameter)) {
			return true
		}
	}
	if sig.ThisParameter() != nil && visit(c.getTypeOfSymbol(sig.ThisParameter())) {
		return true
	}
	if sig.resolvedReturnType != nil && visit(sig.resolvedReturnType) {
		return true
	}
	if predicate := sig.getResolvedTypePredicate(c); predicate != nil && predicate.t != nil && visit(predicate.t) {
		return true
	}
	return false
}

func (c *Checker) instantiateTypefnSignatures(signatures []*Signature, m *TypeMapper) []*Signature {
	var result []*Signature
	for i, sig := range signatures {
		// Preserve native identity when no free dependency changes. In particular,
		// an opaque generic signature must not acquire fresh binders on every replay.
		if c.typefnSignatureDependencies(sig, func(t *Type) bool { return c.instantiateType(t, m) != t }) {
			if result == nil {
				result = slices.Clone(signatures)
			}
			result[i] = c.instantiateSignature(sig, m)
		}
	}
	if result == nil {
		return signatures
	}
	return result
}
