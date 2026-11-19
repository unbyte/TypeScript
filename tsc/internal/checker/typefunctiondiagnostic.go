package checker

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/diagnostics"
	"github.com/microsoft/TypeScript/tsc/internal/scanner"
	"github.com/microsoft/TypeScript/tsc/internal/stringutil"
)

func (c *Checker) failTypeFunctionAtSpan(node *ast.Node, message string, span *Type) {
	root := c.typeFunctionExecution
	if root.failed {
		return
	}
	source, ok := typeFunctionString(c.getTypeOfPropertyOfType(span, "source"))
	start, startOK := typeFunctionOffset(c.getTypeOfPropertyOfType(span, "start"))
	end, endOK := typeFunctionOffset(c.getTypeOfPropertyOfType(span, "end"))
	pos := typeFunctionTextPosition{text: stringutil.CombineSurrogatePairs(source)}
	if !ok || !startOK || !endOK || end < start || !pos.seek(start) || !pos.seek(end) {
		c.failTypeFunction(node, "fail requires a valid decoded input span")
		return
	}
	// Literal types are interned. Equal arguments at different positions cannot
	// identify an occurrence, so keep the decoded offsets on the root in that case.
	var literal *ast.Node
	for _, arg := range root.site.TypeArguments() {
		if ast.IsLiteralTypeNode(arg) {
			n := arg.AsLiteralTypeNode().Literal
			if ast.IsStringLiteral(n) && n.Text() == source {
				if literal != nil {
					literal = nil
					break
				}
				literal = n
			}
		}
	}
	if literal != nil {
		file := ast.GetSourceFileOfNode(literal)
		offset := scanner.GetTokenPosOfNode(literal, file, false)
		if loc, ok := scanner.DecodedStringRange(file.Text()[offset:literal.End()], start, end); ok {
			root.failed = true
			d := ast.NewDiagnostic(file, core.NewTextRange(offset+loc.Pos(), offset+loc.End()), diagnostics.Type_function_Colon_0, message)
			d.AddRelatedInfo(createDiagnosticForNode(root.site, diagnostics.Type_function_Colon_0, "while evaluating this application"))
			root.addFrames(d, node, message)
			c.addDiagnostic(d)
			return
		}
	}
	c.failTypeFunction(node, fmt.Sprintf("%s (decoded UTF-16 offsets %d..%d)", message, start, end))
}

func (root *typeFunctionExecution) addFrames(d *ast.Diagnostic, node *ast.Node, message string) {
	if node != nil && node != root.site {
		d.AddRelatedInfo(createDiagnosticForNode(node, diagnostics.Type_function_Colon_0, message))
	}
	count := 0
	for i := len(root.stack) - 1; i >= 1 && count < 6; i-- {
		site := root.stack[i].use
		if site != nil && site != node && site != root.site {
			d.AddRelatedInfo(createDiagnosticForNode(site, diagnostics.Type_function_Colon_0, "called here"))
			count++
		}
	}
}
