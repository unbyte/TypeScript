package checker

import (
	"math"
	"math/big"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/jsnum"
)

func (v *typefnVM) expression(e *typefnExpression, slots []typefnValue) *Type {
	return v.nativeValue(e.node, v.valueExpression(e, slots))
}

func (v *typefnVM) valueExpression(e *typefnExpression, slots []typefnValue) typefnValue {
	if !v.step(e.node) {
		return typefnValue{native: v.c.errorType}
	}
	var value typefnValue
	if e.slot >= 0 {
		if e.node.Kind == ast.KindTypeFunctionCall {
			value = v.callHelper(e, slots)
		} else {
			value = slots[e.slot]
		}
	} else if e.callee != nil {
		args := make([]typefnValue, len(e.args))
		for i, arg := range e.args {
			args[i] = v.resultExpression(arg, slots)
			if v.failed || v.blocked {
				return typefnValue{native: v.c.errorType}
			}
		}
		value = v.callValue(e.node, e.callee.declaration, args, nil)
	} else if e.nativeContext != nil {
		value.native = v.mapNative(e.nativeContext)
	} else if e.valueSymbol != nil {
		value.native = v.mapNative(v.c.getTypeOfSymbol(e.valueSymbol))
	} else if e.constructorAlias != nil {
		closure := v.constructorAlias(e.constructorAlias)
		if v.failed {
			return typefnValue{native: v.c.errorType}
		}
		// A native alias caches its defining environment. A DSL use selects a
		// mapped immutable view without mutating that declaration's cache.
		if v.nativeMapper != nil {
			closure = v.c.instantiateTypefnClosure(closure, v.nativeMapper)
		}
		if v.failed {
			return typefnValue{native: v.c.errorType}
		}
		if e.node.Kind != ast.KindTypeFunctionCall {
			value.closure = closure
		} else {
			args := make([]*Type, len(e.args))
			for i, arg := range e.args {
				args[i] = v.typeArgument(arg, slots)
				if v.failed || v.blocked {
					return typefnValue{native: v.c.errorType}
				}
			}
			value.native = v.applyConstructor(e.node, closure, args)
		}
	} else if e.nativeConstructor != nil {
		value.closure = v.captureNativeConstructor(e.nativeConstructor)
	} else if e.helper != nil {
		if e.helper.local {
			value = typefnValue{closure: v.captureHelper(e.helper, slots)}
		} else {
			value = typefnValue{closure: v.captureHelper(e.helper, nil)}
		}
	} else {
		value = typefnValue{native: v.expressionWorker(e, slots)}
	}
	// Keep native failures and reached checks even when a result is discarded,
	// absorbed by normalization or projected away from the returned shape.
	if value.native != nil {
		if v.c.isErrorType(value.native) && !v.blocked {
			v.failed = true
		}
		value.native = v.observe(value.native)
	}
	return value
}

func (v *typefnVM) expressionWorker(e *typefnExpression, slots []typefnValue) *Type {
	node := e.node
	if node.Kind == ast.KindTypeFunctionTemplate {
		return v.templateExpression(e, slots)
	}
	data := node.AsTypeFunctionExpression()
	switch node.Kind {
	case ast.KindTypeFunctionNativeType:
		return v.nativeIsland(e, slots)
	case ast.KindTypeFunctionLiteral:
		value := v.c.getTypeFromTypeNode(data.NativeType)
		if value.flags == TypeFlagsNumberLiteral {
			n := float64(value.AsLiteralType().value.(jsnum.Number))
			if math.IsInf(n, 0) || math.IsNaN(n) {
				return v.error(node, "number literal must be finite")
			}
		}
		if value.flags == TypeFlagsBigIntLiteral {
			n, _ := new(big.Int).SetString(value.AsLiteralType().value.(jsnum.PseudoBigInt).String(), 10)
			if n.BitLen() > v.limits.BigintBits {
				return v.error(node, "bigint magnitude limit exceeded")
			}
		}
		return value
	case ast.KindTypeFunctionName:
		return v.apply(node, e.symbol, nil)
	case ast.KindTypeFunctionBinary:
		left := v.expression(e.args[0], slots)
		if v.failed || v.blocked {
			return v.c.errorType
		}
		op := data.OperatorToken.Kind
		if op == ast.KindAmpersandAmpersandToken || op == ast.KindBarBarToken {
			truth, ok := v.boolean(node, left)
			if !ok {
				return v.c.errorType
			}
			if truth == (op == ast.KindBarBarToken) {
				return left
			}
			right := v.expression(e.args[1], slots)
			if v.failed || v.blocked {
				return v.c.errorType
			}
			if _, ok := v.boolean(node, right); !ok {
				return v.c.errorType
			}
			return right
		}
		right := v.expression(e.args[1], slots)
		if v.failed || v.blocked {
			return v.c.errorType
		}
		return v.binary(node, op, left, right)
	case ast.KindTypeFunctionUnary:
		value := v.expression(e.args[0], slots)
		if v.failed || v.blocked {
			return v.c.errorType
		}
		return v.unary(node, data.OperatorToken.Kind, value)
	case ast.KindTypeFunctionTuple:
		return v.tupleExpression(e, slots)
	case ast.KindTypeFunctionObject:
		properties := []typefnConstructionProperty{}
		values := []*Type{}
		names := map[string]bool{}
		for index, field := range data.Operands.Nodes {
			fieldData := field.AsTypeFunctionExpression()
			name := ""
			var nameType *Type
			if e.keys[index] == nil {
				// Text is the scanner's decoded/canonical value, not source spelling.
				name = field.Name().Text()
				if ast.IsNumericLiteral(field.Name()) {
					number := jsnum.FromString(name)
					if math.IsInf(float64(number), 0) || math.IsNaN(float64(number)) {
						return v.error(field, "number property key must be finite")
					}
					name = number.String()
				}
			}
			if e.keys[index] != nil {
				key := v.expression(e.keys[index], slots)
				if v.failed {
					return v.c.errorType
				}
				if v.rejectConstruction(field, key) {
					return v.c.errorType
				}
				if typefnUnresolved(key) {
					return v.deferExpression(e, slots)
				}
				if key.flags == TypeFlagsStringLiteral {
					name = key.AsLiteralType().value.(string)
				} else if key.flags == TypeFlagsNumberLiteral {
					number := key.AsLiteralType().value.(jsnum.Number)
					if math.IsInf(float64(number), 0) || math.IsNaN(float64(number)) {
						return v.error(field, "number property key must be finite")
					}
					name = number.String()
				} else if key.flags == TypeFlagsUniqueESSymbol {
					name, nameType = getPropertyNameFromType(key), key
				} else {
					return v.error(field, "computed key requires one string/number literal or unique symbol")
				}
			}
			if names[name] {
				if nameType != nil {
					return v.error(field, "duplicate object field for the same unique symbol")
				}
				return v.error(field, "duplicate object field "+name)
			}
			value := v.expression(e.args[index], slots)
			if v.failed || v.blocked {
				return v.c.errorType
			}
			properties = append(properties, typefnConstructionProperty{name: name, key: nameType, optional: fieldData.Optional, readonly: fieldData.Readonly})
			values = append(values, value)
			names[name] = true
		}
		return v.constructionObject(node, properties, values)

	case ast.KindTypeFunctionCall, ast.KindTypeFunctionIndex:
		if e.intrinsic == "types.recursive" || e.intrinsic == "types.recursiveGroup" {
			return v.recursiveConstruction(e, slots)
		}
		args := make([]*Type, len(e.args))
		for i, arg := range e.args {
			args[i] = v.expression(arg, slots)
			if v.failed || v.blocked {
				return v.c.errorType
			}
		}
		if node.Kind == ast.KindTypeFunctionIndex {
			if constructionOf(args[0]) != nil {
				return v.constructionIndex(node, args[0], args[1])
			}
			if v.rejectConstruction(node, args[1]) {
				return v.c.errorType
			}
			return v.c.getIndexedAccessTypeEx(args[0], args[1], AccessFlagsNone, node, nil)
		}
		if e.intrinsic != "" {
			return v.intrinsic(e, args, slots)
		}
		return v.apply(node, e.symbol, args)
	}
	return v.error(node, "expression is not implemented yet")
}
