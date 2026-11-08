package checker

import (
	"fmt"
	"math"
	"strings"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/jsnum"
)

// Intrinsics are resolved only in DSL scope, before native entity lookup. Keep
// arity and effect information here so unused branches receive the same checks.
type typefnIntrinsic struct {
	minimum, maximum int
	aborts           bool
}

var typefnIntrinsics = map[string]typefnIntrinsic{
	"text.length":           {1, 1, false},
	"text.peek":             {2, 2, false},
	"text.slice":            {2, 3, false},
	"text.digit":            {1, 1, false},
	"text.codeUnit":         {2, 2, false},
	"text.fromCodeUnit":     {1, 1, false},
	"numeric.toNumber":      {1, 1, false},
	"numeric.toBigint":      {1, 1, false},
	"numeric.isInteger":     {1, 1, false},
	"numeric.isSafeInteger": {1, 1, false},
	"numeric.floor":         {1, 1, false},
	"numeric.ceil":          {1, 1, false},
	"numeric.trunc":         {1, 1, false},
	"numeric.abs":           {1, 1, false},
	"numeric.format":        {1, 1, false},
	"numeric.parseInteger":  {1, 1, false},
	"text.indexOf":          {2, 3, false},
	"text.includes":         {2, 2, false},
	"text.startsWith":       {2, 2, false},
	"text.endsWith":         {2, 2, false},
	"text.split":            {2, 2, false},
	"text.replaceAll":       {3, 3, false},
	"text.concat":           {1, 1, false},
	"tuple.length":          {1, 1, false},
	"tuple.slice":           {2, 3, false},
	"fail":                  {1, 1, true},
	"failAt":                {2, 2, true},
}

func typefnEntityName(node *ast.Node) string {
	if ast.IsIdentifier(node) {
		return node.Text()
	}
	if ast.IsQualifiedName(node) {
		q := node.AsQualifiedName()
		return typefnEntityName(q.Left) + "." + q.Right.Text()
	}
	return ""
}

func (b *typefnCompiler) intrinsic(e *typefnExpression) bool {
	name := typefnEntityName(e.node.Name())
	info, ok := typefnIntrinsics[name]
	if !ok {
		return false
	}
	e.intrinsic = name
	count := 0
	if operands := e.node.AsTypeFunctionExpression().Operands; operands != nil {
		count = len(operands.Nodes)
	}
	if e.node.Kind != ast.KindTypeFunctionCall || count < info.minimum || count > info.maximum {
		b.error(e.node, fmt.Sprintf("%s requires %d..%d Type arguments", name, info.minimum, info.maximum))
	}
	if info.aborts && (e.node.Parent == nil || e.node.Parent.Kind != ast.KindTypeFunctionExpressionStatement) {
		b.error(e.node, name+" is only permitted as an application statement")
	}
	return true
}

func (v *typefnVM) intrinsic(e *typefnExpression, args, slots []*Type) *Type {
	for _, arg := range args {
		if typefnUnresolved(arg) {
			if typefnIntrinsics[e.intrinsic].aborts {
				// Failure has no successful result to publish. The enclosing call waits.
				v.blocked = true
				return v.c.errorType
			}
			return v.deferExpression(e, slots)
		}
	}
	if strings.HasPrefix(e.intrinsic, "text.") {
		result := v.textIntrinsic(e.node, e.intrinsic, args)
		if v.blocked && !v.failed {
			v.blocked = false
			return v.deferExpression(e, slots)
		}
		return result
	}
	if strings.HasPrefix(e.intrinsic, "numeric.") {
		return v.numericIntrinsic(e.node, e.intrinsic, args[0])
	}
	if strings.HasPrefix(e.intrinsic, "tuple.") {
		return v.tupleIntrinsic(e.node, e.intrinsic, args)
	}
	switch e.intrinsic {
	case "fail", "failAt":
		if args[0].flags != TypeFlagsStringLiteral {
			return v.error(e.node, "failure message requires one string literal")
		}
		message := args[0].AsLiteralType().value.(string)
		if len(args) == 2 {
			offset, ok := v.intrinsicOffset(e.node, args[1])
			if !ok {
				return v.c.errorType
			}
			message = fmt.Sprintf("%s (UTF-16 offset %d)", message, offset)
		}
		return v.error(e.node, message)
	}
	panic("unhandled typefn intrinsic")
}

func (v *typefnVM) intrinsicOffset(node *ast.Node, value *Type) (int64, bool) {
	if value.flags == TypeFlagsNumberLiteral {
		n := float64(value.AsLiteralType().value.(jsnum.Number))
		if n >= 0 && n <= 9007199254740991 && math.Trunc(n) == n {
			return int64(n), true
		}
	}
	v.error(node, "offset requires a nonnegative safe-integer number literal")
	return 0, false
}
