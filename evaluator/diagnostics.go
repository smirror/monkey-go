package evaluator

import (
	"monkey-go/ast"
	"monkey-go/diagnostic"
	"monkey-go/object"
	"monkey-go/token"
)

func withErrorSpan(value object.Object, span token.Span) object.Object {
	err, ok := value.(*object.Error)
	if !ok {
		return value
	}
	if err.Span.Start.Line > 0 && err.Kind != "" {
		return value
	}
	copy := *err
	if copy.Span.Start.Line == 0 {
		copy.Span = span
	}
	if copy.Kind == "" {
		copy.Kind = diagnostic.Runtime
	}
	return &copy
}

func errorSpan(node ast.Node) token.Span {
	switch node := node.(type) {
	case *ast.InfixExpression:
		return node.Token.Span
	case *ast.PrefixExpression:
		return node.Token.Span
	default:
		return node.Span()
	}
}

func withCallFrame(value object.Object, node *ast.CallExpression) object.Object {
	err, ok := value.(*object.Error)
	if !ok {
		return value
	}
	name := "<anonymous>"
	if identifier, ok := node.Function.(*ast.Identifier); ok {
		name = identifier.Value
	}
	copy := *err
	copy.Stack = append([]diagnostic.Frame(nil), err.Stack...)
	copy.Stack = append(copy.Stack, diagnostic.Frame{Name: name, Span: node.Span()})
	return &copy
}
