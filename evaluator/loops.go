package evaluator

import (
	"monkey-go/ast"
	"monkey-go/object"
)

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
