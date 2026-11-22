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
	located := *err
	if located.Span.Start.Line == 0 {
		located.Span = span
	}
	if located.Kind == "" {
		located.Kind = diagnostic.Runtime
	}
	return &located
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
	traced := *err
	traced.Stack = append([]diagnostic.Frame(nil), err.Stack...)
	traced.Stack = append(traced.Stack, diagnostic.Frame{Name: name, Span: node.Span()})
	return &traced
}
