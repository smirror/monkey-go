package evaluator

import (
	"monkey-go/ast"
	"monkey-go/object"
)

func evalReturnStatement(node *ast.ReturnStatement, env *object.Environment) object.Object {
	value := Eval(node.ReturnValue, env)
	if isEvaluationExit(value) {
		return value
	}
	return &object.ReturnValue{Value: value}
}

func evalPrefixExpressionNode(node *ast.PrefixExpression, env *object.Environment) object.Object {
	right := Eval(node.Right, env)
	if isEvaluationExit(right) {
		return right
	}
	return evalPrefixExpression(node.Operator, right)
}

func evalLetStatement(node *ast.LetStatement, env *object.Environment) object.Object {
	value := Eval(node.Value, env)
	if isEvaluationExit(value) {
		return value
	}
	env.Set(node.Name.Value, value)
	return value
}

func evalCallExpression(node *ast.CallExpression, env *object.Environment) object.Object {
	function := Eval(node.Function, env)
	if isEvaluationExit(function) {
		return function
	}
	args := evalExpressions(node.Arguments, env)
	if len(args) == 1 && isEvaluationExit(args[0]) {
		return args[0]
	}
	value := applyFunction(function, args)
	fn, entered := function.(*object.Function)
	if !entered || len(args) != len(fn.Parameters) {
		return value
	}
	return withCallFrame(value, node)
}

func evalArrayLiteral(node *ast.ArrayLiteral, env *object.Environment) object.Object {
	elements := evalExpressions(node.Elements, env)
	if len(elements) == 1 && isEvaluationExit(elements[0]) {
		return elements[0]
	}
	return &object.Array{Elements: elements}
}

func evalIndexExpressionNode(node *ast.IndexExpression, env *object.Environment) object.Object {
	left := Eval(node.Left, env)
	if isEvaluationExit(left) {
		return left
	}
	index := Eval(node.Index, env)
	if isEvaluationExit(index) {
		return index
	}
	return evalIndexExpression(left, index)
}

func evalMacroExpansion(node *ast.MacroExpansion, env *object.Environment) object.Object {
	return withCallFrame(Eval(node.Body, env), node.Call)
}

// An expanded template can return from the surrounding function even when it
// occurs inside another expression. Never consume that return as an operand.
func isEvaluationExit(value object.Object) bool {
	if value == nil {
		return false
	}
	return value.Type() == object.ERROR_OBJ || value.Type() == object.RETURN_VALUE_OBJ
}
