package evaluator

import (
	"monkey-go/ast"
	"monkey-go/object"
	"monkey-go/token"
	"strconv"
)

// quote retains syntax without evaluating it, except for unquote expressions.
func quote(node ast.Node, env *object.Environment) object.Object {
	var failure *object.Error
	rewritten := ast.Modify(node, func(current ast.Node) ast.Node {
		if failure != nil {
			return current
		}
		call, ok := current.(*ast.CallExpression)
		if !ok {
			return current
		}
		name, ok := call.Function.(*ast.Identifier)
		if !ok || name.Value != "unquote" {
			return current
		}
		if len(call.Arguments) != 1 {
			failure = newError("unquote expects 1 argument, got %d", len(call.Arguments))
			return current
		}
		value := Eval(call.Arguments[0], env)
		if value == nil {
			failure = newError("unquote evaluated to no value")
			return current
		}
		if errObj, ok := value.(*object.Error); ok {
			failure = errObj
			return current
		}
		replacement, ok := objectToExpression(value)
		if !ok {
			failure = newError("cannot unquote %s as an expression", value.Type())
			return current
		}
		return replacement
	})
	if failure != nil {
		return failure
	}
	return &object.Quote{Node: rewritten}
}

func objectToExpression(value object.Object) (ast.Expression, bool) {
	switch value := value.(type) {
	case *object.Integer:
		literal := strconv.FormatInt(value.Value, 10)
		return &ast.IntegerLiteral{
			Token: token.Token{Type: token.INT, Literal: literal},
			Value: value.Value,
		}, true
	case *object.Boolean:
		literal := "false"
		typeName := token.FALSE
		if value.Value {
			literal = "true"
			typeName = token.TRUE
		}
		return &ast.Boolean{
			Token: token.Token{Type: typeName, Literal: literal},
			Value: value.Value,
		}, true
	case *object.String:
		return &ast.StringLiteral{
			Token: token.Token{Type: token.STRING, Literal: value.Value},
			Value: value.Value,
		}, true
	case *object.Quote:
		expression, ok := value.Node.(ast.Expression)
		return expression, ok
	default:
		return nil, false
	}
}
