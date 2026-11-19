package declarations

import (
	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/diagnostics"
)

// A consumer executes the body from the declaration file. Follow its lexical
// dependencies before stripping private helpers, including through ordinary aliases.
func (tx *DeclarationTransformer) retainTypeFunctionDependencies(root *ast.Node) {
	if tx.typeFunctionDependencies == nil {
		tx.typeFunctionDependencies = make(map[*ast.Node]bool)
	}
	seen := map[*ast.Node]bool{root: true}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		var name *ast.Node
		switch node.Kind {
		case ast.KindTypeReference:
			name = node.AsTypeReferenceNode().TypeName
		case ast.KindTypeQuery:
			name = node.AsTypeQueryNode().ExprName
		}
		if name != nil {
			for _, dep := range tx.resolver.ReferencedTypeFunctionDependency(name) {
				if seen[dep] {
					continue
				}
				seen[dep] = true
				if dep.Parent == nil || !(ast.IsSourceFile(dep.Parent) || ast.IsModuleBlock(dep.Parent)) {
					continue
				}
				source := ast.GetSourceFileOfNode(dep)
				if source.IsDeclarationFile {
					continue
				}
				if tx.isInternalDeclaration(dep, source) {
					if source != tx.state.currentSourceFile || ast.HasSyntacticModifier(dep, ast.ModifierFlagsExport) || !ast.IsExternalModule(source) {
						tx.state.addDiagnostic(createDiagnosticForNode(name, diagnostics.Type_function_Colon_0, "an exported computation depends on a declaration removed by stripInternal"))
						continue
					}
					if !tx.typeFunctionDependencies[dep] {
						tx.typeFunctionDependencies[dep] = true
						tx.state.lateMarkedStatements = append(tx.state.lateMarkedStatements, dep)
					}
				}
				if ast.IsTypeAliasDeclaration(dep) || ast.IsTypeFunctionDeclaration(dep) || ast.IsInterfaceDeclaration(dep) {
					dep.ForEachChild(visit)
				}
			}
		}
		return node.ForEachChild(visit)
	}
	root.ForEachChild(visit)
}
