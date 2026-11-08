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
	parameters       string
}

var typefnIntrinsics = map[string]typefnIntrinsic{
	"types.tuple":           {1, 1, false, "Descriptor"},
	"reflect.tuple":         {1, 1, false, "Tuple"},
	"reflect.properties":    {1, 1, false, "Object"},
	"types.signature":       {1, 1, false, "Descriptor"},
	"reflect.signature":     {1, 1, false, "Signature"},
	"types.function":        {2, 2, false, "Parameters, Result"},
	"types.constructor":     {2, 3, false, "Parameters, Result, Abstract = false"},
	"reflect.calls":         {1, 1, false, "Object"},
	"reflect.constructs":    {1, 1, false, "Object"},
	"reflect.isGeneric":     {1, 1, false, "Signature"},
	"types.overload":        {1, 2, false, "Calls, Constructs = []"},
	"reflect.indexers":      {1, 1, false, "Object"},
	"types.object":          {1, 4, false, "Properties, Indexers = [], Calls = [], Constructs = []"},
	"types.array":           {1, 2, false, "Element, Readonly = false"},
	"reflect.array":         {1, 1, false, "Array"},
	"reflect.kind":          {1, 1, false, "Value"},
	"reflect.isAny":         {1, 1, false, "Value"},
	"reflect.isNever":       {1, 1, false, "Value"},
	"reflect.isUnknown":     {1, 1, false, "Value"},
	"text.length":           {1, 1, false, "Text"},
	"text.peek":             {2, 2, false, "Text, Offset"},
	"text.slice":            {2, 3, false, "Text, Start, End = text.length<Text>"},
	"text.digit":            {1, 1, false, "Text"},
	"text.codeUnit":         {2, 2, false, "Text, Offset"},
	"text.fromCodeUnit":     {1, 1, false, "Unit"},
	"numeric.toNumber":      {1, 1, false, "Value"},
	"numeric.toBigint":      {1, 1, false, "Value"},
	"numeric.isInteger":     {1, 1, false, "Value"},
	"numeric.isSafeInteger": {1, 1, false, "Value"},
	"numeric.floor":         {1, 1, false, "Value"},
	"numeric.ceil":          {1, 1, false, "Value"},
	"numeric.trunc":         {1, 1, false, "Value"},
	"numeric.abs":           {1, 1, false, "Value"},
	"numeric.format":        {1, 1, false, "Value"},
	"numeric.parseInteger":  {1, 1, false, "Text"},
	"text.indexOf":          {2, 3, false, "Text, Needle, Start = 0"},
	"text.includes":         {2, 2, false, "Text, Needle"},
	"text.startsWith":       {2, 2, false, "Text, Prefix"},
	"text.endsWith":         {2, 2, false, "Text, Suffix"},
	"text.split":            {2, 2, false, "Text, Delimiter"},
	"text.replaceAll":       {3, 3, false, "Text, Needle, Replacement"},
	"text.concat":           {1, 1, false, "Parts"},
	"tuple.length":          {1, 1, false, "Tuple"},
	"tuple.slice":           {2, 3, false, "Tuple, Start, End = tuple.length<Tuple>"},
	"fail":                  {1, 1, true, "Message"},
	"failAt":                {2, 2, true, "Message, Offset"},
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

func (v *typefnVM) intrinsic(e *typefnExpression, args []*Type, slots []typefnValue) (result *Type) {
	// Each pure intrinsic can discover a dependency below its outer arguments.
	// Retain one frozen operation for replay; aborting operations block the call.
	defer func() {
		if v.blocked && !v.failed && !typefnIntrinsics[e.intrinsic].aborts {
			v.blocked = false
			result = v.deferExpression(e, slots)
		}
	}()
	// Constructors can publish structure containing symbolic Type leaves. Only
	// their control/metadata operands need concrete observations.
	switch e.intrinsic {
	case "types.tuple":
		return v.tupleType(e.node, args[0])
	case "types.function":
		return v.functionType(e.node, args, SignatureKindCall)
	case "types.constructor":
		return v.functionType(e.node, args, SignatureKindConstruct)
	case "types.overload":
		return v.overloadType(e.node, args)
	case "types.array":
		return v.arrayType(e.node, args)
	case "types.object", "types.signature":
		var result *Type
		if e.intrinsic == "types.object" {
			result = v.objectType(e.node, args)
		} else {
			result = v.signatureType(e.node, args[0])
		}
		if v.blocked && !v.failed && !v.c.isErrorType(result) {
			v.blocked = false
			return v.c.guardTypefn(result, []*Type{v.deferExpression(e, slots)})
		}
		return result
	}
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
	if strings.HasPrefix(e.intrinsic, "reflect.") {
		switch e.intrinsic {
		case "reflect.tuple":
			return v.reflectTuple(e.node, args[0])
		case "reflect.properties":
			return v.reflectProperties(e.node, args[0])
		case "reflect.signature":
			return v.reflectSignature(e.node, args[0])
		case "reflect.calls":
			return v.reflectSignatures(e.node, args[0], SignatureKindCall)
		case "reflect.constructs":
			return v.reflectSignatures(e.node, args[0], SignatureKindConstruct)
		case "reflect.isGeneric":
			sig, _, ok := v.singleSignature(e.node, args[0])
			if !ok {
				return v.c.errorType
			}
			return v.boolType(len(sig.TypeParameters()) != 0)
		}
		if e.intrinsic == "reflect.indexers" {
			return v.reflectIndexers(e.node, args[0])
		}
		if e.intrinsic == "reflect.array" {
			return v.reflectArray(e.node, args[0])
		}
		return v.reflectKind(e.node, e.intrinsic, args[0])
	}
	if strings.HasPrefix(e.intrinsic, "text.") {
		return v.textIntrinsic(e.node, e.intrinsic, args)
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
