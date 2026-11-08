package checker

import "github.com/microsoft/TypeScript/tsc/internal/ast"

// Immutable native type sources for forwarded signatures on a sealed TypeLiteral.
// They share its lexical environment. Resolving them for each native instance
// preserves opaque binders, method origins, duplicates and overload ordering.
// An intersection would instead deduplicate/reorder signatures. Instantiating the
// raw target's signatures would lose opaque handles hidden behind its environment.
type typefnConstructionMembers struct {
	calls, constructs []*ast.Node
}

func (c *Checker) resolveTypefnConstructionMembers(t *Type, plan *typefnConstructionMembers) {
	d := t.AsObjectType()
	c.setStructuredTypeMembers(t, nil, nil, nil, nil)
	var members ast.SymbolTable
	var indices []*IndexInfo
	if d.target != nil {
		members = c.createInstantiatedSymbolTable(c.getPropertiesOfObjectType(d.target), d.mapper)
		indices = c.instantiateIndexInfos(c.getIndexInfosOfType(d.target), d.mapper)
	} else {
		members = c.getMembersOfSymbol(c.getMergedSymbol(t.symbol))
		indices = c.getIndexInfosOfSymbol(t.symbol)
	}
	signatures := func(sources []*ast.Node, kind SignatureKind) []*Signature {
		var result []*Signature
		for _, source := range sources {
			value := c.instantiateType(c.getTypeFromTypeNode(source), d.mapper)
			result = append(result, c.getSignaturesOfType(value, kind)...)
		}
		return result
	}
	calls := signatures(plan.calls, SignatureKindCall)
	constructs := signatures(plan.constructs, SignatureKindConstruct)
	c.setStructuredTypeMembers(t, members, calls, constructs, indices)
}
