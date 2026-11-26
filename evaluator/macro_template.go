package evaluator

import (
	"monkey-go/ast"
	"monkey-go/object"
)

// substituteMacroBody only replaces references to the macro's parameters.
// Argument syntax is copied without visiting it again as part of the template.
func substituteMacroBody(macro *object.Macro, arguments []ast.Expression) *ast.BlockStatement {
	bindings := make(map[string]ast.Expression, len(arguments))
	for i, parameter := range macro.Parameters {
		bindings[parameter.Value] = arguments[i]
	}
	result := rewriteMacroSyntax(macro.Body, func(node ast.Node, scope macroScope) (ast.Node, bool) {
		identifier, ok := node.(*ast.Identifier)
		if !ok || scope[identifier.Value] {
			return node, true
		}
		replacement, found := bindings[identifier.Value]
		if !found {
			return node, true
		}
		return cloneMacroSyntax(replacement), false
	})
	return result.(*ast.BlockStatement)
}

func cloneMacroSyntax(node ast.Node) ast.Node {
	return ast.Rewrite(node, func(current ast.Node) (ast.Node, bool) {
		return current, true
	})
}
