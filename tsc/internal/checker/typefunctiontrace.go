package checker

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/scanner"
)

func typeFunctionLocation(node *ast.Node) string {
	f := ast.GetSourceFileOfNode(node)
	line := scanner.ComputeLineOfPosition(f.ECMALineMap(), node.Pos())
	return fmt.Sprintf("%s:%d", f.FileName(), line+1)
}

// Trace rendering only inspects values already present in a frame. In particular,
// it must not resolve a property or instantiate a deferred application.
func (c *Checker) typeFunctionDisplay(t *Type, depth int) string {
	if t == nil {
		return "<unbound>"
	}
	if s, ok := typeFunctionString(t); ok {
		if len(s) > 120 {
			s = s[:120] + "…"
		}
		return strconv.Quote(s)
	}
	if t.flags&(TypeFlagsNumberLiteral|TypeFlagsBigIntLiteral|TypeFlagsBooleanLiteral) != 0 {
		return fmt.Sprint(t.AsLiteralType().value)
	}
	if t.typeFunctionApplication() != nil {
		return "<deferred>"
	}
	if t.IsTupleType() && depth < 2 {
		var parts []string
		for i, v := range t.AsTypeReference().resolvedTypeArguments {
			if i == 8 {
				parts = append(parts, "…")
				break
			}
			parts = append(parts, c.typeFunctionDisplay(v, depth+1))
		}
		return "[" + strings.Join(parts, ", ") + "]"
	}
	if t.symbol != nil {
		return t.symbol.Name()
	}
	return c.typeFunctionKind(t)
}

func (f *typeFunctionFrame) log(node *ast.Node) {
	c := f.c
	root := c.typeFunctionExecution
	if !c.compilerOptions.TraceTypeFunctions.IsTrue() || len(root.trace) > 128 {
		return
	}
	if len(root.trace) == 128 {
		root.trace = append(root.trace, "trace truncated")
		return
	}
	var values []string
	for _, n := range node.TypeArguments() {
		if ast.IsLiteralTypeNode(n) && ast.IsStringLiteral(n.AsLiteralTypeNode().Literal) {
			values = append(values, strconv.Quote(n.AsLiteralTypeNode().Literal.Text()))
		} else if isSimpleIdentifierTypeReference(n) {
			name := n.AsTypeReferenceNode().TypeName.Text()
			for i, source := range slices.Backward(f.sources) {
				if source.symbol != nil && source.symbol.Name() == name {
					values = append(values, c.typeFunctionDisplay(f.values[i], 0))
					break
				}
			}
		}
	}
	root.trace = append(root.trace, fmt.Sprintf("%s (called at %s) %s", typeFunctionLocation(node), typeFunctionLocation(root.site), strings.Join(values, " ")))
}

type typeFunctionTraceKey struct {
	site      *ast.Node
	arguments CacheHashKey
}

func (c *Checker) publishTypeFunctionTrace(app *typeFunctionApplication) {
	root := c.typeFunctionExecution
	if len(root.trace) == 0 || (c.ctx != nil && c.ctx.Err() != nil) {
		return
	}
	host, ok := c.program.(interface{ TraceTypeFunction(event string) })
	if !ok {
		return
	}
	key := typeFunctionTraceKey{app.use, getTypeListKey(app.arguments)}
	if c.typeFunctionReportedTraces[key] {
		return
	}
	if c.typeFunctionReportedTraces == nil {
		c.typeFunctionReportedTraces = make(map[typeFunctionTraceKey]bool)
	}
	c.typeFunctionReportedTraces[key] = true
	for _, event := range root.trace {
		host.TraceTypeFunction(event)
	}
}
