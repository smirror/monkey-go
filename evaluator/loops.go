package evaluator

import (
	"math"
	"monkey-go/ast"
	"monkey-go/object"
	"monkey-go/token"
)

// Each loop gets an enclosed scope. An unadorned i = 0 initializer declares
// i in that scope; assignments elsewhere still follow Environment.Reassign.
func evalForStatement(stmt *ast.ForStatement, outer *object.Environment) object.Object {
	env := object.NewEnclosedEnvironment(outer)
	if initResult := evalOptionalForInitializer(stmt.Init, env); isError(initResult) {
		return initResult
	}

	for {
		result, again := evalForIteration(stmt, env)
		if !again {
			return result
		}
	}
}

func evalOptionalForInitializer(init ast.Statement, env *object.Environment) object.Object {
	if init == nil {
		return NULL
	}
	return evalForInitializer(init, env)
}

func evalForIteration(stmt *ast.ForStatement, env *object.Environment) (object.Object, bool) {
	condition := Eval(stmt.Condition, env)
	if isError(condition) {
		return condition, false
	}
	if !isTruthy(condition) {
		return NULL, false
	}

	if result := Eval(stmt.Body, env); isLoopExit(result) {
		return result, false
	}
	postResult := evalOptionalForPost(stmt.Post, env)
	if isError(postResult) {
		return postResult, false
	}
	return nil, true
}

func evalOptionalForPost(post ast.Expression, env *object.Environment) object.Object {
	if post == nil {
		return NULL
	}
	return Eval(post, env)
}

func evalForInitializer(init ast.Statement, env *object.Environment) object.Object {
	exprStmt, ok := init.(*ast.ExpressionStatement)
	if !ok {
		return Eval(init, env)
	}
	assign, ok := exprStmt.Expression.(*ast.InfixExpression)
	if !ok || assign.Operator != token.ASSIGN {
		return Eval(init, env)
	}
	name, ok := assign.Left.(*ast.Identifier)
	if !ok {
		return newError("left side of assignment must be an identifier")
	}
	value := Eval(assign.Right, env)
	if isError(value) {
		return value
	}
	env.Set(name.Value, value)
	return value
}

// The iterable is evaluated once in the outer scope, before the loop variable
// shadows any identifier with the same name.
func evalForInStatement(stmt *ast.ForInStatement, outer *object.Environment) object.Object {
	value := Eval(stmt.Iterable, outer)
	if isError(value) {
		return value
	}
	if value == nil {
		return newError("object is not iterable: <nil>")
	}
	iterable, ok := value.(object.Iterable)
	if !ok {
		return newError("object is not iterable: %s", value.Type())
	}

	env := object.NewEnclosedEnvironment(outer)
	iter := iterable.Iterator()
	for {
		element, more := iter.Next()
		if !more {
			return NULL
		}
		env.Set(stmt.Variable.Value, element)
		result := Eval(stmt.Body, env)
		if isLoopExit(result) {
			return result
		}
	}
}

func isLoopExit(value object.Object) bool {
	if value == nil {
		return false
	}
	return isError(value) || value.Type() == object.RETURN_VALUE_OBJ
}

func evalPostfixExpression(expr *ast.PostfixExpression, env *object.Environment) object.Object {
	if expr.Operator != token.INCREMENT {
		return newError("unknown postfix operator: %s", expr.Operator)
	}
	ident, ok := expr.Left.(*ast.Identifier)
	if !ok {
		return newError("left side of ++ must be an identifier")
	}
	current, found := env.Get(ident.Value)
	if !found {
		return newError("identifier not found: " + ident.Value)
	}
	integer, ok := current.(*object.Integer)
	if !ok {
		return newError("increment requires INTEGER, got %s", current.Type())
	}
	if integer.Value == math.MaxInt64 {
		return newError("integer overflow in increment: %s++", ident.Value)
	}
	_, _ = env.Reassign(ident.Value, &object.Integer{Value: integer.Value + 1})
	return current // Postfix returns the original value.
}
