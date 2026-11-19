package evaluator

import (
	"math"
	"monkey-go/object"
)

// Arithmetic overflows are Monkey errors, not wrapped two's-complement values.
// This keeps results deterministic across the Interpreter and a future VM.
func integerOverflow(operator string, left, right int64) object.Object {
	return newError("integer overflow: %d %s %d", left, operator, right)
}

func evalIntegerAdd(left, right int64) object.Object {
	if right > 0 && left > math.MaxInt64-right {
		return integerOverflow("+", left, right)
	}
	if right < 0 && left < math.MinInt64-right {
		return integerOverflow("+", left, right)
	}
	return &object.Integer{Value: left + right}
}

func evalIntegerSubtract(left, right int64) object.Object {
	if right < 0 && left > math.MaxInt64+right {
		return integerOverflow("-", left, right)
	}
	if right > 0 && left < math.MinInt64+right {
		return integerOverflow("-", left, right)
	}
	return &object.Integer{Value: left - right}
}

func evalIntegerMultiply(left, right int64) object.Object {
	if left == math.MinInt64 && right == -1 {
		return integerOverflow("*", left, right)
	}
	if right == math.MinInt64 && left == -1 {
		return integerOverflow("*", left, right)
	}
	result := left * right
	if right != 0 && result/right != left {
		return integerOverflow("*", left, right)
	}
	return &object.Integer{Value: result}
}

func evalIntegerDivide(left, right int64) object.Object {
	if right == 0 {
		return newError("division by zero")
	}
	if left == math.MinInt64 && right == -1 {
		return integerOverflow("/", left, right)
	}
	return &object.Integer{Value: left / right}
}

func evalIntegerNegate(value int64) object.Object {
	if value == math.MinInt64 {
		return newError("integer overflow in negation: %d", value)
	}
	return &object.Integer{Value: -value}
}
