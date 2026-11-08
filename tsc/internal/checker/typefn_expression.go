package checker

import (
	"math"
	"math/big"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/jsnum"
)

func (v *typefnVM) expression(e *typefnExpression, slots []*Type) *Type {
	value := v.expressionWorker(e, slots)
	// Preserve native operation failure even if this value would be discarded,
	// absorbed by a union, or projected away by the surrounding DSL expression.
	if v.c.isErrorType(value) && !v.blocked {
		v.failed = true
	}
	return v.observe(value)
}

func (v *typefnVM) expressionWorker(e *typefnExpression, slots []*Type) *Type {
	node := e.node
	if !v.step(node) {
		return v.c.errorType
	}
	if e.slot >= 0 {
		return slots[e.slot]
	}
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
		var values []*Type
		for _, element := range e.args {
			item := element
			spread := element.node.Kind == ast.KindTypeFunctionSpread
			if spread {
				item = element.args[0]
			}
			value := v.expression(item, slots)
			if v.failed || v.blocked {
				return v.c.errorType
			}
			if spread {
				if value.flags == TypeFlagsStringLiteral {
					return v.error(node, "tuple spread requires a tuple")
				}
				elements := v.elements(node, value, false)
				if v.blocked && !v.failed {
					v.blocked = false
					return v.deferExpression(e, slots)
				}
				if v.failed {
					return v.c.errorType
				}
				values = append(values, elements...)
			} else {
				values = append(values, value)
			}
			if !v.allocate(node, 1) {
				return v.c.errorType
			}
		}
		if !v.allocate(node, len(values)+1) {
			return v.c.errorType
		}
		return v.c.createTupleType(values)
	case ast.KindTypeFunctionObject:
		members := ast.NewSymbolTable()
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
			if members.Get(name) != nil {
				if nameType != nil {
					return v.error(field, "duplicate object field for the same unique symbol")
				}
				return v.error(field, "duplicate object field "+name)
			}
			value := v.expression(e.args[index], slots)
			if v.failed || v.blocked {
				return v.c.errorType
			}
			if !v.allocate(field, 1) {
				return v.c.errorType
			}
			flags := ast.SymbolFlagsProperty
			if fieldData.Optional {
				flags |= ast.SymbolFlagsOptional
			}
			checkFlags := ast.CheckFlagsNone
			if fieldData.Readonly {
				checkFlags |= ast.CheckFlagsReadonly
			}
			symbol := v.c.newSymbolEx(flags, name, checkFlags)
			v.c.valueSymbolLinks.Get(symbol).resolvedType = v.c.addOptionalityEx(value, true, fieldData.Optional)
			v.c.valueSymbolLinks.Get(symbol).setNameType(v.c, nameType)
			members.Set(name, symbol)
		}
		if !v.allocate(node, 1) {
			return v.c.errorType
		}
		return v.c.newTypefnObject(members)
	case ast.KindTypeFunctionCall, ast.KindTypeFunctionIndex:
		args := make([]*Type, len(e.args))
		for i, arg := range e.args {
			args[i] = v.expression(arg, slots)
			if v.failed || v.blocked {
				return v.c.errorType
			}
		}
		if node.Kind == ast.KindTypeFunctionIndex {
			return v.c.getIndexedAccessTypeEx(args[0], args[1], AccessFlagsNone, node, nil)
		}
		if e.intrinsic != "" {
			return v.intrinsic(e, args, slots)
		}
		return v.apply(node, e.symbol, args)
	}
	return v.error(node, "expression is not implemented yet")
}
