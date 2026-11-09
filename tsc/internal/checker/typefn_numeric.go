package checker

import (
	"math"
	"math/big"
	"strings"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/jsnum"
)

// Native aliases and parameters can supply bigint literals without executing a
// DSL literal expression. Every numeric consumer must therefore validate its
// input here, before allocating an arbitrary-precision representation.
func (v *typefnVM) bigintOperand(node *ast.Node, value *Type) *big.Int {
	return v.bigintDecimal(node, value.AsLiteralType().value.(jsnum.PseudoBigInt).String())
}

// The decimal-length check bounds SetString's input allocation even when the
// exact bit count is not known yet. The final bit check handles the boundary
// decimal digit, including values such as 255 and 256 under an eight-bit limit.
func (v *typefnVM) bigintDecimal(node *ast.Node, text string) *big.Int {
	if v.failed || v.c.isCanceled() {
		v.failed = true
		return nil
	}
	negative := strings.HasPrefix(text, "-")
	digits := strings.TrimLeft(strings.TrimPrefix(strings.TrimPrefix(text, "-"), "+"), "0")
	if float64(len(digits)) > math.Floor(float64(v.limits.BigintBits)*math.Log10(2))+1 {
		v.error(node, "bigint magnitude limit exceeded")
		return nil
	}
	if digits == "" {
		digits = "0"
	}
	n, ok := new(big.Int).SetString(digits, 10)
	if !ok {
		v.error(node, "invalid integer")
		return nil
	}
	if v.c.isCanceled() {
		v.failed = true
		return nil
	}
	if n.BitLen() > v.limits.BigintBits {
		v.error(node, "bigint magnitude limit exceeded")
		return nil
	}
	if negative {
		n.Neg(n)
	}
	return n
}

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
		n := v.bigintDecimal(node, s)
		if n == nil {
			return v.c.errorType
		}
		return v.bigintResult(node, n)
	}
	if arg.flags == TypeFlagsBigIntLiteral {
		n := v.bigintOperand(node, arg)
		if n == nil {
			return v.c.errorType
		}
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
