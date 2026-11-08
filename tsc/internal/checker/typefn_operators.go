package checker

import (
	"math"
	"math/big"
	"strings"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/jsnum"
)

func (v *typefnVM) binary(node *ast.Node, op ast.Kind, a, b *Type) *Type {
	a, b = v.observe(a), v.observe(b)
	if op != ast.KindBarToken && op != ast.KindAmpersandToken &&
		!(a == b && (op == ast.KindEqualsEqualsToken || op == ast.KindExclamationEqualsToken)) &&
		(typefnUnresolved(a) || typefnUnresolved(b)) {
		return v.c.newTypefnTerm(&typefnTerm{kind: typefnBinaryTerm, node: node, operator: op, args: []*Type{a, b}})
	}
	switch op {
	case ast.KindBarToken:
		return v.c.getUnionType([]*Type{a, b})
	case ast.KindAmpersandToken:
		return v.c.getIntersectionType([]*Type{a, b})
	case ast.KindEqualsEqualsToken, ast.KindExclamationEqualsToken, ast.KindIsKeyword:
		if a != b && (v.unresolvedGraph(node, a) || v.unresolvedGraph(node, b)) {
			if v.failed {
				return v.c.errorType
			}
			return v.c.newTypefnTerm(&typefnTerm{kind: typefnBinaryTerm, node: node, operator: op, args: []*Type{a, b}})
		}
		result := a == b
		if !result {
			if op == ast.KindIsKeyword {
				result = v.c.isTypeAssignableTo(a, b)
			} else {
				result = v.c.isTypeIdenticalTo(a, b)
			}
		}
		if op == ast.KindExclamationEqualsToken {
			result = !result
		}
		return v.boolType(result)
	}
	compare := func(result int) *Type {
		switch op {
		case ast.KindLessThanToken:
			return v.boolType(result < 0)
		case ast.KindLessThanEqualsToken:
			return v.boolType(result <= 0)
		case ast.KindGreaterThanToken:
			return v.boolType(result > 0)
		case ast.KindGreaterThanEqualsToken:
			return v.boolType(result >= 0)
		}
		return nil
	}
	if a.flags == TypeFlagsNumberLiteral && b.flags == TypeFlagsNumberLiteral {
		left, right := float64(a.AsLiteralType().value.(jsnum.Number)), float64(b.AsLiteralType().value.(jsnum.Number))
		if math.IsInf(left, 0) || math.IsInf(right, 0) || math.IsNaN(left) || math.IsNaN(right) {
			return v.error(node, "number operands must be finite")
		}
		cmp := 0
		if left < right {
			cmp = -1
		} else if left > right {
			cmp = 1
		}
		if result := compare(cmp); result != nil {
			return result
		}
		var result float64
		switch op {
		case ast.KindPlusToken:
			result = left + right
		case ast.KindMinusToken:
			result = left - right
		case ast.KindAsteriskToken:
			result = left * right
		case ast.KindSlashToken:
			if right == 0 {
				return v.error(node, "division by zero")
			}
			result = left / right
		case ast.KindPercentToken:
			if right == 0 {
				return v.error(node, "remainder by zero")
			}
			result = math.Mod(left, right)
		default:
			return v.error(node, "unsupported scalar operator")
		}
		if math.IsNaN(result) || math.IsInf(result, 0) {
			return v.error(node, "number arithmetic must produce a finite value")
		}
		if result == 0 {
			result = 0
		}
		return v.c.getNumberLiteralType(jsnum.Number(result))
	}
	if a.flags == TypeFlagsBigIntLiteral && b.flags == TypeFlagsBigIntLiteral {
		left, _ := new(big.Int).SetString(a.AsLiteralType().value.(jsnum.PseudoBigInt).String(), 10)
		right, _ := new(big.Int).SetString(b.AsLiteralType().value.(jsnum.PseudoBigInt).String(), 10)
		if left.BitLen() > v.limits.BigintBits || right.BitLen() > v.limits.BigintBits {
			return v.error(node, "bigint magnitude limit exceeded")
		}
		if op == ast.KindAsteriskToken && left.Sign() != 0 && right.Sign() != 0 && left.BitLen()-1 > v.limits.BigintBits-right.BitLen() {
			return v.error(node, "bigint magnitude limit exceeded")
		}
		if result := compare(left.Cmp(right)); result != nil {
			return result
		}
		result := new(big.Int)
		switch op {
		case ast.KindPlusToken:
			result.Add(left, right)
		case ast.KindMinusToken:
			result.Sub(left, right)
		case ast.KindAsteriskToken:
			result.Mul(left, right)
		case ast.KindSlashToken:
			if right.Sign() == 0 {
				return v.error(node, "division by zero")
			}
			result.Quo(left, right)
		case ast.KindPercentToken:
			if right.Sign() == 0 {
				return v.error(node, "remainder by zero")
			}
			result.Rem(left, right)
		default:
			return v.error(node, "unsupported bigint operator")
		}
		if result.BitLen() > v.limits.BigintBits {
			return v.error(node, "bigint magnitude limit exceeded")
		}
		text := result.String()
		return v.c.getBigIntLiteralType(jsnum.NewPseudoBigInt(strings.TrimPrefix(text, "-"), result.Sign() < 0))
	}
	return v.error(node, "scalar operator requires single unbranded literals of the same numeric family")
}

func (v *typefnVM) unary(node *ast.Node, op ast.Kind, value *Type) *Type {
	value = v.observe(value)
	if op != ast.KindKeyOfKeyword && typefnUnresolved(value) {
		return v.c.newTypefnTerm(&typefnTerm{kind: typefnUnaryTerm, node: node, operator: op, args: []*Type{value}})
	}
	switch op {
	case ast.KindKeyOfKeyword:
		return v.c.getIndexType(value)
	case ast.KindExclamationToken:
		b, ok := v.boolean(node, value)
		if !ok {
			return v.c.errorType
		}
		return v.boolType(!b)
	case ast.KindPlusToken, ast.KindMinusToken:
		zero := v.c.getNumberLiteralType(0)
		if value.flags == TypeFlagsBigIntLiteral {
			zero = v.c.getBigIntLiteralType(jsnum.PseudoBigInt{})
		}
		return v.binary(node, op, zero, value)
	}
	return v.error(node, "unsupported unary operator")
}
