package checker

import (
	"context"
	"strings"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/scanner"
)

// A constructor has a signature, not an ordinary alias's recovery Type. Keeping
// this distinction in the checker lets all editor clients present it consistently.
func (c *Checker) GetTypeFunctionDeclaration(symbol *ast.Symbol) *ast.Node {
	if symbol == nil {
		return nil
	}
	if symbol.Flags()&ast.SymbolFlagsAlias != 0 {
		symbol = c.resolveAlias(symbol)
	}
	if symbol == nil {
		return nil
	}
	return core.Find(symbol.Declarations(), ast.IsTypeFunctionDeclaration)
}

func (c *Checker) TypeFunctionSignature(symbol *ast.Symbol) string {
	declaration := c.GetTypeFunctionDeclaration(symbol)
	if declaration == nil {
		return ""
	}
	parameters := make([]string, 0, len(declaration.TypeParameters()))
	for _, parameter := range declaration.TypeParameters() {
		parameters = append(parameters, scanner.GetTextOfNode(parameter))
	}
	signature := "typefn " + declaration.Name().Text() + "<" + strings.Join(parameters, ", ") + ">"
	var resultKind *typefnKind
	if annotation := ast.TypeFunctionResultKind(declaration); annotation != nil {
		if slots := annotation.AsTypeFunctionKind().Slots; slots != nil {
			resultKind = typefnConstructorKind(len(slots.Nodes))
		}
	} else {
		resultKind = c.compileTypefn(declaration).resultKind
		// Incomplete source is not lowered. Keep parameter help available, but
		// do not claim a Type result for a factory whose kind is not known yet.
		if resultKind.resolved() < 0 {
			return signature
		}
	}
	return signature + ": " + typefnKindDisplay(resultKind)
}

// Queries evaluate only native applications, never an arbitrary generic body or
// a mutable local at an imagined invocation. Use the caller's cancellation token
// and omit error-recovery Types from UI. Query and diagnostic results share the
// native checker node caches and program-snapshot lifetime.
func (c *Checker) GetTypeFunctionApplicationResult(ctx context.Context, node *ast.Node) *Type {
	if node == nil || !ast.IsTypeReferenceNode(node) || ctx.Err() != nil {
		return nil
	}
	if len(ast.GetSourceFileOfNode(node).Diagnostics()) != 0 {
		return nil
	}
	if ast.FindAncestor(node, ast.IsTypeFunctionLike) != nil {
		return nil
	}
	symbol := c.resolveEntityName(node.AsTypeReferenceNode().TypeName, ast.SymbolFlagsType, true, false, nil)
	if c.GetTypeFunctionDeclaration(symbol) == nil || node.TypeArgumentList() == nil {
		return nil
	}
	saved := c.ctx
	c.ctx = ctx
	defer func() {
		c.ctx = saved
		if ctx.Err() != nil {
			c.wasCanceled = true
		}
	}()
	result := c.getTypeFromTypeNode(node)
	if c.isErrorType(result) || result.flags&TypeFlagsTypefn != 0 || ctx.Err() != nil {
		return nil
	}
	return result
}

func (c *Checker) TypeFunctionBindingKind(declaration *ast.Node) string {
	if declaration == nil {
		return "type"
	}
	fn := ast.FindAncestor(declaration.Parent, ast.IsTypeFunctionLike)
	if fn == nil {
		return "type"
	}
	p := c.compileTypefn(fn)
	return typefnKindDisplay(p.bindings[declaration.Symbol()])
}
