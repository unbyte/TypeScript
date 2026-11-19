package checker

import (
	"fmt"

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
	case ast.KindTypeFunctionAssignment:
		a := node.AsTypeFunctionAssignment()
		symbol := c.resolveName(a.Target, a.Target.Text(), ast.SymbolFlagsTypeParameter, nil, false, false)
		if symbol == nil {
			c.error(a.Target, diagnostics.Type_function_Colon_0, "assignment requires a bound local")
		} else if decl := ast.GetDeclarationOfKind(symbol, ast.KindTypeParameter); decl == nil || !(ast.IsTypeFunctionBinding(decl.Parent) || ast.IsTypeFunctionFor(decl.Parent)) {
			c.error(a.Target, diagnostics.Type_function_Colon_0, "only a calculation local may be rebound")
		}
	case ast.KindTypeFunctionCallStatement:
		operand := node.AsTypeFunctionCallStatement().Operand
		if !ast.IsTypeReferenceNode(operand) || operand.TypeArgumentList() == nil {
			c.error(operand, diagnostics.Type_function_Colon_0, "expected an explicit type function call")
		}
	case ast.KindTypeFunctionMatch:
		arms := node.AsTypeFunctionMatch().Arms.Nodes
		for i, arm := range arms {
			if arm.AsTypeFunctionArm().Pattern == nil && i+1 < len(arms) {
				c.error(arm, diagnostics.Type_function_Colon_0, "the else arm must be last")
			}
		}
	case ast.KindTypeFunctionArm:
		if pattern := node.AsTypeFunctionArm().Pattern; pattern != nil {
			seen := make(map[string]bool)
			for _, capture := range typeFunctionPatternCaptures(pattern) {
				name := capture.Name().Text()
				if seen[name] {
					c.error(capture, diagnostics.Type_function_Colon_0, "duplicate pattern capture "+name)
				}
				seen[name] = true
			}
		}
	case ast.KindTypeFunctionDeclaration:
		return
	case ast.KindTypeReference:
		if typeFunctionHole(node) && c.typeFunctionConstructorArity(node.Parent) > 0 {
			return
		}
		name := node.AsTypeReferenceNode().TypeName
		if ast.IsIdentifier(name) && ast.IsTypeFunctionCallStatement(node.Parent) && (name.Text() == "log" || name.Text() == "fail") {
			if name.Text() == "fail" && (len(node.TypeArguments()) < 1 || len(node.TypeArguments()) > 2) {
				c.error(node, diagnostics.Type_function_Colon_0, "fail requires a message and an optional span")
			}
			if name.Text() == "log" {
				for i, a := range node.TypeArguments() {
					if i == 0 && ast.IsLiteralTypeNode(a) && ast.IsStringLiteral(a.AsLiteralTypeNode().Literal) {
						continue
					}
					if !isSimpleIdentifierTypeReference(a) {
						c.error(a, diagnostics.Type_function_Colon_0, "log operands must name a binding")
					}
				}
			}
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
			symbol := c.getSymbolFromTypeReference(node)
			if typeFunctionConstructorParameter(symbol) != 0 && node.TypeArgumentList() == nil {
				c.error(node, diagnostics.Type_function_Colon_0, "a constructor input must be applied or passed with explicit holes")
			}
			if decl := ast.GetDeclarationOfKind(symbol, ast.KindTypeFunctionDeclaration); decl != nil {
				if node.TypeArgumentList() == nil || len(node.TypeArguments()) != len(decl.TypeParameters()) {
					c.error(node, diagnostics.Type_function_Colon_0, fmt.Sprintf("%s requires %d explicit type arguments", symbol.Name(), len(decl.TypeParameters())))
				}
			}
		}
	}
	node.ForEachChild(visit)
}
