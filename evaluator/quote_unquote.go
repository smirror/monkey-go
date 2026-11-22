package evaluator

import (
	"monkey-go/ast"
	"monkey-go/object"
	"monkey-go/token"
	"strconv"
)

func isCallNamed(call *ast.CallExpression, name string) bool {
	identifier, ok := call.Function.(*ast.Identifier)
	return ok && identifier.Value == name
}

func evalQuoteCall(call *ast.CallExpression, env *object.Environment) object.Object {
	if len(call.Arguments) != 1 {
		return newError("quote expects 1 argument, got %d", len(call.Arguments))
	}
	return quote(call.Arguments[0], env)
}

// quote retains syntax without evaluating it, except for unquote expressions.
func quote(node ast.Node, env *object.Environment) object.Object {
	var failure *object.Error
	rewritten := ast.Modify(node, func(current ast.Node) ast.Node {
		if failure != nil {
			return current
		}
		call, ok := current.(*ast.CallExpression)
		if !ok || !isCallNamed(call, "unquote") {
			return current
		}
		replacement, err := evalUnquoteCall(call, env)
		if err != nil {
			failure = withErrorSpan(err, call.Span()).(*object.Error)
			return current
		}
		return replacement
	})
	if failure != nil {
		return failure
	}
	return &object.Quote{Node: rewritten}
}

func evalUnquoteCall(call *ast.CallExpression, env *object.Environment) (ast.Expression, *object.Error) {
	if len(call.Arguments) != 1 {
		return nil, newError("unquote expects 1 argument, got %d", len(call.Arguments))
	}
	value := Eval(call.Arguments[0], env)
	if value == nil {
		return nil, newError("unquote evaluated to no value")
	}
	if errObj, ok := value.(*object.Error); ok {
		return nil, errObj
	}
	replacement, ok := objectToExpression(value, call.Span())
	if !ok {
		return nil, newError("cannot unquote %s as an expression", value.Type())
	}
	return replacement, nil
}

func objectToExpression(value object.Object, origin token.Span) (ast.Expression, bool) {
	var expression ast.Expression
	switch value := value.(type) {
	case *object.Integer:
		literal := strconv.FormatInt(value.Value, 10)
		expression = &ast.IntegerLiteral{
			Token: token.Token{Type: token.INT, Literal: literal, Span: origin},
			Value: value.Value,
		}
	case *object.Boolean:
		literal := "false"
		typeName := token.TokenType(token.FALSE)
		if value.Value {
			literal = "true"
			typeName = token.TRUE
		}
		expression = &ast.Boolean{
			Token: token.Token{Type: typeName, Literal: literal, Span: origin},
			Value: value.Value,
		}
	case *object.String:
		expression = &ast.StringLiteral{
			Token: token.Token{Type: token.STRING, Literal: value.Value, Span: origin},
			Value: value.Value,
		}
	case *object.Quote:
		expression, ok := value.Node.(ast.Expression)
		return expression, ok
	default:
		return nil, false
	}
	ast.SetSpan(expression, origin)
	return expression, true
}
