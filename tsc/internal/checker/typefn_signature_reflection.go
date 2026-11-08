package checker

import (
	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/jsnum"
)

func (v *typefnVM) reflectSignature(node *ast.Node, value *Type) *Type {
	sig, kind, ok := v.singleSignature(node, value)
	if !ok {
		return v.c.errorType
	}
	if len(sig.TypeParameters()) != 0 {
		return v.error(node, "reflect.signature does not expose native generic binders; forward the original signature")
	}
	if !v.allocate(node, len(sig.parameters)+5) {
		return v.c.errorType
	}
	parameters := make([]*Type, 0, len(sig.parameters))
	for i, parameter := range sig.parameters {
		if !v.step(node) {
			return v.c.errorType
		}
		mode := "required"
		switch {
		case signatureHasRestParameter(sig) && i == len(sig.parameters)-1:
			mode = "rest"
		case parameter.CheckFlags()&ast.CheckFlagsOptionalParameter != 0 || parameter.Flags()&ast.SymbolFlagsOptional != 0:
			mode = "optional"
		case parameter.ValueDeclaration() != nil && v.c.isOptionalParameter(parameter.ValueDeclaration()):
			mode = "optional"
		}
		parameters = append(parameters, v.record(node,
			typefnRecordField{"name", v.c.getStringLiteralType(parameter.Name())},
			typefnRecordField{"value", v.c.getTypeOfSymbol(parameter)},
			typefnRecordField{"mode", v.c.getStringLiteralType(mode)}))
		if v.failed {
			return v.c.errorType
		}
	}
	var this, predicate []*Type
	if sig.ThisParameter() != nil {
		this = []*Type{v.c.getTypeOfSymbol(sig.ThisParameter())}
	}
	if pred := v.c.getTypePredicateOfSignature(sig); pred != nil {
		predicateKind := "is"
		if pred.kind == TypePredicateKindAssertsIdentifier || pred.kind == TypePredicateKindAssertsThis {
			predicateKind = "asserts"
		}
		target := v.c.getNumberLiteralType(jsnum.Number(pred.parameterIndex))
		if pred.kind == TypePredicateKindThis || pred.kind == TypePredicateKindAssertsThis {
			target = v.c.getStringLiteralType("this")
		}
		var narrowed []*Type
		if pred.t != nil {
			narrowed = []*Type{pred.t}
		}
		predicate = []*Type{v.record(node,
			typefnRecordField{"kind", v.c.getStringLiteralType(predicateKind)},
			typefnRecordField{"target", target},
			typefnRecordField{"narrowed", v.c.createTupleType(narrowed)})}
	}
	kindName := "call"
	if kind == SignatureKindConstruct {
		kindName = "construct"
	}
	return v.record(node,
		typefnRecordField{"kind", v.c.getStringLiteralType(kindName)},
		typefnRecordField{"parameters", v.c.createTupleType(parameters)},
		typefnRecordField{"this", v.c.createTupleType(this)},
		typefnRecordField{"returns", v.c.getReturnTypeOfSignature(sig)},
		typefnRecordField{"abstract", v.boolType(sig.flags&SignatureFlagsAbstract != 0)},
		typefnRecordField{"predicate", v.c.createTupleType(predicate)})
}
