package evaluator

import (
	"monkey-go/ast"
	"monkey-go/object"
	"monkey-go/token"
)

// Loop control is internal to evaluation and is consumed by the closest loop.
// Public Eval and function calls reject signals that reach their boundary.
type loopControl struct {
	keyword string
	span    token.Span
}

func (c *loopControl) Type() object.ObjectType { return "LOOP_CONTROL" }
func (c *loopControl) Inspect() string         { return c.keyword }

// The iterable is evaluated once in the outer scope, before the loop variable
// shadows any identifier with the same name.
func evalForInStatement(stmt *ast.ForInStatement, outer *object.Environment) object.Object {
	value := eval(stmt.Iterable, outer)
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
		result, exit := evalLoopBody(stmt.Body, env)
		if exit {
			return result
		}
	}
}

// Each condition uses the outer scope. The body reuses a loop-local scope,
// matching for-in, so body declarations cannot shadow names in the condition.
func evalWhileStatement(stmt *ast.WhileStatement, outer *object.Environment) object.Object {
	env := object.NewEnclosedEnvironment(outer)
	for {
		condition := eval(stmt.Condition, outer)
		if isEvaluationExit(condition) {
			return condition
		}
		if !isTruthy(condition) {
			return NULL
		}
		if result, exit := evalLoopBody(stmt.Body, env); exit {
			return result
		}
	}
}

// A continue finishes the current body; a break finishes this loop. A return
// or error propagates unchanged, while the loop ignores ordinary body values.
func evalLoopBody(body *ast.BlockStatement, env *object.Environment) (object.Object, bool) {
	result := eval(body, env)
	if control, ok := result.(*loopControl); ok {
		return NULL, control.keyword == "break"
	}
	return result, isLoopExit(result)
}

func isEvaluationExit(value object.Object) bool {
	if _, ok := value.(*loopControl); ok {
		return true
	}
	return isLoopExit(value)
}

func isLoopExit(value object.Object) bool {
	if value == nil {
		return false
	}
	return isError(value) || value.Type() == object.RETURN_VALUE_OBJ
}
