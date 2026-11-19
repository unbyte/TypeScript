package checker

import (
	"math"
	"math/big"
	"regexp"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/jsnum"
	"github.com/microsoft/TypeScript/tsc/internal/typefunction"
)

func (c *Checker) typeFunctionBoolean(b bool) *Type {
	if b {
		return c.regularTrueType
	}
	return c.regularFalseType
}

func typeFunctionString(t *Type) (string, bool) {
	if t != nil && t.flags&TypeFlagsStringLiteral != 0 {
		return t.AsLiteralType().value.(string), true
	}
	return "", false
}

func typeFunctionNumber(t *Type) (float64, bool) {
	if t != nil && t.flags&TypeFlagsNumberLiteral != 0 {
		return float64(t.AsLiteralType().value.(jsnum.Number)), true
	}
	return 0, false
}

func typeFunctionInteger(t *Type) (*big.Int, bool) {
	if t == nil || t.flags&TypeFlagsBigIntLiteral == 0 {
		return nil, false
	}
	return new(big.Int).SetString(t.AsLiteralType().value.(jsnum.PseudoBigInt).String(), 10)
}

var (
	typeFunctionDecimal = regexp.MustCompile("^[+-]?[0-9]+([.][0-9]+)?([eE][+-]?[0-9]+)?$")
	typeFunctionWhole   = regexp.MustCompile("^[+-]?[0-9]+$")
)

func (c *Checker) runTypeFunctionBuiltin(node *ast.Node, name string, args []*Type) *Type {
	op, ok := typefunction.Lookup(name)
	if !ok || len(args) < op.MinArguments || len(args) > op.MaxArguments {
		return c.failTypeFunction(node, "invalid builtin application "+name)
	}
	first := args[0]
	// Charge work before parsing digits or allocating a larger numeric result.
	for _, arg := range args {
		size := 0
		if v, ok := typeFunctionString(arg); ok {
			size = len(v)
		} else if arg.flags&TypeFlagsBigIntLiteral != 0 {
			size = len(arg.AsLiteralType().value.(jsnum.PseudoBigInt).String())
		} else if arg.IsTupleType() {
			size = len(c.getTypeArguments(arg))
		} else if arg.flags&TypeFlagsUnionOrIntersection != 0 {
			size = len(arg.Types())
		}
		if !c.typeFunctionCharge(node, size) {
			return c.anyType
		}
	}
	switch name {
	case "type.kind":
		return c.getStringLiteralType(c.typeFunctionKind(first))
	case "type.isIdentical":
		return c.typeFunctionBoolean(c.isTypeIdenticalTo(first, args[1]))
	case "type.isAssignableTo":
		return c.typeFunctionBoolean(c.isTypeAssignableTo(first, args[1]))
	case "type.baseOfLiteral":
		return c.getBaseTypeOfLiteralType(first)
	case "tuple.count":
		if first.flags&TypeFlagsNever != 0 {
			return c.failTypeFunction(node, "tuple.count requires a tuple")
		}
		var counts []*Type
		for _, part := range first.Distributed() {
			if !part.IsTupleType() {
				return c.failTypeFunction(node, "tuple.count requires a tuple")
			}
			counts = append(counts, c.getTypeOfPropertyOfType(part, "length"))
		}
		return c.getUnionType(counts)
	case "intersection.parts":
		if first.flags&TypeFlagsIntersection != 0 {
			return c.createTupleType(first.Types())
		}
		return c.createTupleType([]*Type{first})
	case "intersection.fromTuple":
		if first.IsTupleType() {
			return c.getIntersectionType(c.getTypeArguments(first))
		}
	}
	group, operation, _ := strings.Cut(name, ".")
	switch group {
	case "number", "bigint":
		if v := c.typeFunctionArithmetic(group, operation, args); v != nil {
			return v
		}
	case "fn":
		if v := c.typeFunctionCallable(node, operation, args); v != nil {
			return v
		}
	case "string":
		if v := c.typeFunctionText(node, operation, args); v != nil {
			return v
		}
	default:
		if v := c.typeFunctionStructure(node, name, args); v != nil {
			return v
		}
	}
	return c.failTypeFunction(node, "invalid operands for "+name)
}

func (c *Checker) typeFunctionArithmetic(group, op string, args []*Type) *Type {
	var right *Type
	if len(args) > 1 {
		right = args[1]
	}
	if group == "number" {
		a, ok := typeFunctionNumber(args[0])
		var b float64
		switch op {
		case "fromDecimal":
			s, valid := typeFunctionString(args[0])
			if !valid || !typeFunctionDecimal.MatchString(s) {
				return nil
			}
			a, _ = strconv.ParseFloat(s, 64)
		case "fromBigInt":
			i, valid := typeFunctionInteger(args[0])
			if !valid || i.BitLen() > 53 {
				return nil
			}
			a, _ = new(big.Float).SetInt(i).Float64()
		default:
			if !ok {
				return nil
			}
			if right != nil {
				b, ok = typeFunctionNumber(right)
				if !ok {
					return nil
				}
			}
			switch op {
			case "add":
				a += b
			case "subtract":
				a -= b
			case "multiply":
				a *= b
			case "divide":
				a /= b
			case "remainder":
				a = math.Mod(a, b)
			case "negate":
				a = -a
			case "trunc":
				a = math.Trunc(a)
			case "lessThan":
				return c.typeFunctionBoolean(a < b)
			}
		}
		if math.IsInf(a, 0) || math.IsNaN(a) {
			return nil
		}
		return c.getNumberLiteralType(jsnum.Number(a))
	}
	var a *big.Int
	switch op {
	case "fromDecimal":
		s, ok := typeFunctionString(args[0])
		if !ok || !typeFunctionWhole.MatchString(s) {
			return nil
		}
		a, _ = new(big.Int).SetString(s, 10)
	case "fromNumber":
		n, ok := typeFunctionNumber(args[0])
		if !ok || math.Trunc(n) != n || math.Abs(n) > 9007199254740991 {
			return nil
		}
		a, _ = big.NewFloat(n).Int(nil)
	default:
		var ok bool
		a, ok = typeFunctionInteger(args[0])
		if !ok {
			return nil
		}
		var b *big.Int
		if right != nil {
			b, ok = typeFunctionInteger(right)
			if !ok {
				return nil
			}
		}
		switch op {
		case "add":
			a.Add(a, b)
		case "subtract":
			a.Sub(a, b)
		case "multiply":
			a.Mul(a, b)
		case "divide":
			if b.Sign() == 0 {
				return nil
			}
			a.Quo(a, b)
		case "remainder":
			if b.Sign() == 0 {
				return nil
			}
			a.Rem(a, b)
		case "negate":
			a.Neg(a)
		case "lessThan":
			return c.typeFunctionBoolean(a.Cmp(b) < 0)
		}
	}
	return c.getBigIntLiteralType(jsnum.NewPseudoBigInt(new(big.Int).Abs(a).String(), a.Sign() < 0))
}

func (c *Checker) typeFunctionKind(t *Type) string {
	switch {
	case t.flags&TypeFlagsEnumLiteral != 0:
		if t.flags&TypeFlagsUnion != 0 {
			return "enum"
		}
		return "enumMember"
	case t == c.booleanType:
		return "boolean"
	case t.IsTupleType():
		return "tuple"
	case c.isArrayType(t):
		return "array"
	}
	switch t.flags {
	case TypeFlagsAny:
		return "any"
	case TypeFlagsUnknown:
		return "unknown"
	case TypeFlagsNever:
		return "never"
	case TypeFlagsVoid:
		return "void"
	case TypeFlagsNull:
		return "null"
	case TypeFlagsUndefined:
		return "undefined"
	case TypeFlagsString:
		return "string"
	case TypeFlagsNumber:
		return "number"
	case TypeFlagsBooleanLiteral:
		return "booleanLiteral"
	case TypeFlagsStringLiteral:
		return "stringLiteral"
	case TypeFlagsNumberLiteral:
		return "numberLiteral"
	case TypeFlagsBigIntLiteral:
		return "bigintLiteral"
	case TypeFlagsBigInt:
		return "bigint"
	case TypeFlagsESSymbol:
		return "symbol"
	case TypeFlagsUniqueESSymbol:
		return "uniqueSymbol"
	case TypeFlagsStringMapping:
		return "stringMapping"
	case TypeFlagsTemplateLiteral:
		return "templateLiteral"
	case TypeFlagsNonPrimitive:
		return "nonPrimitive"
	case TypeFlagsUnion:
		return "union"
	case TypeFlagsIntersection:
		return "intersection"
	case TypeFlagsObject:
		return "object"
	}
	return "other"
}
