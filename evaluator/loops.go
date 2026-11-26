package evaluator

import (
	"monkey-go/ast"
	"monkey-go/object"
)

// The iterable is evaluated once in the outer scope, before the loop variable
// shadows any identifier with the same name.
func evalForInStatement(stmt *ast.ForInStatement, outer *object.Environment) object.Object {
	value := Eval(stmt.Iterable, outer)
	if isEvaluationExit(value) {
		return value
	}
	if value == nil {
		return newError("object is not iterable: <nil>")
	}
	iterable, ok := value.(object.Iterable)
	if !ok {
		return withErrorSpan(newError("object is not iterable: %s", value.Type()), stmt.Iterable.Span())
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
		if isEvaluationExit(result) {
			return result
		}
	}
}
