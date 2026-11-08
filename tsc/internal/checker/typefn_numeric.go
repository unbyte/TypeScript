package checker

import (
	"math"
	"math/big"
	"strings"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/jsnum"
)

func (v *typefnVM) bigintResult(node *ast.Node, n *big.Int) *Type {
	if n.BitLen() > v.limits.BigintBits {
		return v.error(node, "bigint magnitude limit exceeded")
	}
	if !v.allocate(node, 1) {
		return v.c.errorType
	}
	return v.c.getBigIntLiteralType(jsnum.NewPseudoBigInt(strings.TrimPrefix(n.String(), "-"), n.Sign() < 0))
}

func (v *typefnVM) numericIntrinsic(node *ast.Node, name string, arg *Type) *Type {
	if name == "numeric.parseInteger" {
		if arg.flags != TypeFlagsStringLiteral {
			return v.error(node, "parseInteger requires one string literal")
		}
		s := arg.AsLiteralType().value.(string)
		digits := s
		if len(digits) > 0 && (digits[0] == '+' || digits[0] == '-') {
			digits = digits[1:]
		}
		if len(digits) == 0 {
			return v.error(node, "parseInteger requires [+-]?[0-9]+")
		}
		for i := range len(digits) {
			if !v.stringWork(node, 1) {
				return v.c.errorType
			}
			if digits[i] < '0' || digits[i] > '9' {
				return v.error(node, "parseInteger requires [+-]?[0-9]+")
			}
		}
		significant := strings.TrimLeft(digits, "0")
		// A bit-count-sized integer cannot require more than this many decimal
		// digits. Check before the big.Int allocation; check its exact bits afterward.
		if float64(len(significant)) > math.Floor(float64(v.limits.BigintBits)*math.Log10(2))+1 {
			return v.error(node, "bigint magnitude limit exceeded")
		}
		if significant == "" {
			significant = "0"
		}
		n, ok := new(big.Int).SetString(significant, 10)
		if !ok {
			return v.error(node, "invalid integer")
		}
		if len(s) > 0 && s[0] == '-' {
			n.Neg(n)
		}
		return v.bigintResult(node, n)
	}
	if arg.flags == TypeFlagsBigIntLiteral {
		n, _ := new(big.Int).SetString(arg.AsLiteralType().value.(jsnum.PseudoBigInt).String(), 10)
		switch name {
		case "numeric.toNumber":
			f, accuracy := new(big.Float).SetInt(n).Float64()
			if accuracy != big.Exact || math.IsInf(f, 0) {
				return v.error(node, "bigint is not exactly representable as a finite number")
			}
			return v.c.getNumberLiteralType(jsnum.Number(f))
		case "numeric.abs":
			return v.bigintResult(node, n.Abs(n))
		case "numeric.format":
			s := n.String()
			if !v.stringWork(node, len(s)) || !v.allocate(node, 1) {
				return v.c.errorType
			}
			return v.c.getStringLiteralType(s)
		}
		return v.error(node, name+" requires a number literal")
	}
	if arg.flags != TypeFlagsNumberLiteral {
		return v.error(node, "numeric operand requires one unbranded numeric literal")
	}
	n := float64(arg.AsLiteralType().value.(jsnum.Number))
	if math.IsInf(n, 0) || math.IsNaN(n) {
		return v.error(node, "numeric operand must be finite")
	}
	switch name {
	case "numeric.toNumber":
		return v.error(node, "toNumber requires a bigint literal")
	case "numeric.toBigint":
		if math.Trunc(n) != n {
			return v.error(node, "toBigint requires an integral number")
		}
		integer, _ := new(big.Float).SetFloat64(n).Int(nil)
		return v.bigintResult(node, integer)
	case "numeric.isInteger":
		return v.boolType(math.Trunc(n) == n)
	case "numeric.isSafeInteger":
		return v.boolType(math.Trunc(n) == n && math.Abs(n) <= 9007199254740991)
	case "numeric.floor":
		n = math.Floor(n)
	case "numeric.ceil":
		n = math.Ceil(n)
	case "numeric.trunc":
		n = math.Trunc(n)
	case "numeric.abs":
		n = math.Abs(n)
	case "numeric.format":
		s := jsnum.Number(n).String()
		if !v.stringWork(node, len(s)) || !v.allocate(node, 1) {
			return v.c.errorType
		}
		return v.c.getStringLiteralType(s)
	default:
		panic("unhandled numeric intrinsic")
	}
	if n == 0 {
		n = 0
	} // canonicalize negative zero
	return v.c.getNumberLiteralType(jsnum.Number(n))
}
