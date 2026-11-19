package checker

import (
	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/diagnostics"
	"github.com/microsoft/TypeScript/tsc/internal/typefunction"
)

func (c *Checker) checkTypeFunctionBlock(node *ast.Node, loopDepth int) {
	visit := func(child *ast.Node) bool { c.checkTypeFunctionBlock(child, loopDepth); return false }
	switch node.Kind {
	case ast.KindTypeFunctionFor, ast.KindTypeFunctionWhile:
		loopDepth++
	case ast.KindBreakStatement, ast.KindContinueStatement:
		if loopDepth == 0 {
			c.error(node, diagnostics.Type_function_Colon_0, "loop control requires an enclosing loop")
		}
		return
	case ast.KindTypeReference:
		name := node.AsTypeReferenceNode().TypeName
		if ast.IsIdentifier(name) && ast.IsTypeFunctionCallStatement(node.Parent) && (name.Text() == "log" || name.Text() == "fail") {
			for _, arg := range node.TypeArguments() {
				visit(arg)
			}
			return
		}
		if name := c.typeFunctionBuiltinName(node); name != "" {
			op, ok := typefunction.Lookup(name)
			if !ok {
				c.error(node, diagnostics.Type_function_Colon_0, "unknown builtin "+name)
			} else if len(node.TypeArguments()) < op.MinArguments || len(node.TypeArguments()) > op.MaxArguments {
				c.error(node, diagnostics.Type_function_Colon_0, "incorrect number of arguments to "+name)
			}
		} else {
			c.getSymbolFromTypeReference(node)
		}
	}
	node.ForEachChild(visit)
}
