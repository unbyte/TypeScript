package checker

import (
	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"strings"
)

// Editor metadata comes from the same registry as definition checking. Returning
// copies prevents language service clients from changing the compiler's prelude.
type TypeFunctionIntrinsicInfo struct {
	Name       string
	Signature  string
	Parameters []string
}

func typefnIntrinsicInfo(name string, info typefnIntrinsic) TypeFunctionIntrinsicInfo {
	suffix := ": type"
	if info.aborts {
		suffix = " (aborts)"
	}
	return TypeFunctionIntrinsicInfo{name, name + "<" + info.parameters + ">" + suffix, strings.Split(info.parameters, ", ")}
}

func TypeFunctionIntrinsics() []TypeFunctionIntrinsicInfo {
	result := make([]TypeFunctionIntrinsicInfo, 0, len(typefnIntrinsics))
	for name, info := range typefnIntrinsics {
		result = append(result, typefnIntrinsicInfo(name, info))
	}
	return result
}

// Only the callee name of a DSL expression denotes a prelude intrinsic. Native
// namespaces with identical spellings, including inside islands, remain native.
func TypeFunctionIntrinsicAtLocation(node *ast.Node) *TypeFunctionIntrinsicInfo {
	if node == nil {
		return nil
	}
	for node.Parent != nil && ast.IsQualifiedName(node.Parent) {
		node = node.Parent
	}
	parent := node.Parent
	if parent == nil || parent.Name() != node || (parent.Kind != ast.KindTypeFunctionCall && parent.Kind != ast.KindTypeFunctionName) {
		return nil
	}
	name := typefnEntityName(node)
	if info, ok := typefnIntrinsics[name]; ok {
		result := typefnIntrinsicInfo(name, info)
		return &result
	}
	return nil
}
