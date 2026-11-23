package evaluator

import (
	"monkey-go/ast"
	"monkey-go/object"
)

func evalStatement(node ast.Statement, env *object.Environment) object.Object {
	switch node := node.(type) {
	case *ast.BlockStatement:
		return evalBlockStatement(node, env)
	case *ast.ExpressionStatement:
		return eval(node.Expression, env)
	case *ast.LetStatement:
		return evalLetStatement(node, env)
	default:
		return evalControlStatement(node, env)
	}
}

func evalControlStatement(node ast.Statement, env *object.Environment) object.Object {
	switch node := node.(type) {
	case *ast.ForInStatement:
		return evalForInStatement(node, env)
	case *ast.ReturnStatement:
		return evalReturnStatement(node, env)
	case *ast.BreakStatement:
		return &loopControl{keyword: "break", span: node.Span()}
	case *ast.ContinueStatement:
		return &loopControl{keyword: "continue", span: node.Span()}
	}
	return nil
}

func evalExpression(node ast.Expression, env *object.Environment) object.Object {
	switch node := node.(type) {
	case *ast.PrefixExpression:
		return evalPrefixExpressionNode(node, env)
	case *ast.InfixExpression:
		return evalInfixExpressionNode(node, env)
	case *ast.IfExpression:
		return evalIfExpression(node, env)
	case *ast.Identifier:
		return evalIdentifier(node, env)
	case *ast.CallExpression:
		return evalCallExpression(node, env)
	case *ast.IndexExpression:
		return evalIndexExpressionNode(node, env)
	default:
		return evalLiteral(node, env)
	}
}

func evalLiteral(node ast.Expression, env *object.Environment) object.Object {
	switch node := node.(type) {
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.Boolean:
		return nativeBoolToBooleanObject(node.Value)
	case *ast.StringLiteral:
		return &object.String{Value: node.Value}
	case *ast.FunctionLiteral:
		return &object.Function{Parameters: node.Parameters, Env: env, Body: node.Body}
	case *ast.ArrayLiteral:
		return evalArrayLiteral(node, env)
	case *ast.HashMapLiteral:
		return evalHashMapLiteral(node, env)
	}
	return nil
}

func evalReturnStatement(node *ast.ReturnStatement, env *object.Environment) object.Object {
	value := eval(node.ReturnValue, env)
	if isEvaluationExit(value) {
		return value
	}
	return &object.ReturnValue{Value: value}
}

func evalPrefixExpressionNode(node *ast.PrefixExpression, env *object.Environment) object.Object {
	right := eval(node.Right, env)
	if isEvaluationExit(right) {
		return right
	}
	return evalPrefixExpression(node.Operator, right)
}

func evalLetStatement(node *ast.LetStatement, env *object.Environment) object.Object {
	value := eval(node.Value, env)
	if isEvaluationExit(value) {
		return value
	}
	env.Set(node.Name.Value, value)
	return value
}

func evalCallExpression(node *ast.CallExpression, env *object.Environment) object.Object {
	function := eval(node.Function, env)
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
	left := eval(node.Left, env)
	if isEvaluationExit(left) {
		return left
	}
	index := eval(node.Index, env)
	if isEvaluationExit(index) {
		return index
	}
	return evalIndexExpression(left, index)
}
