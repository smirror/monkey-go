package evaluator

import (
	"monkey-go/object"
	"testing"
)

func TestForInEvaluation(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{"let sum = 0; for (i in range(2, 10)) { sum = sum + i; } sum;", 44},
		{"let sum = 0; for (i in range(5)) { sum = sum + i; } sum;", 10},
		{"let sum = 0; for (i in range(10, 0, -2)) { sum = sum + i; } sum;", 30},
		{"let sum = 0; for (i in range(1, 10, 3)) { sum = sum + i; } sum;", 12},
		{"let sum = 0; for (i in range(0)) { sum = sum + i; } sum;", 0},
		{"let sum = 0; for (i in range(5, 1)) { sum = sum + i; } sum;", 0},
		{"let sum = 0; for (item in [10, 20, 30]) { sum = sum + item; } sum;", 60},
		{"let xs = [3, 4]; let sum = 0; for (item in xs) { sum = sum + item; } sum;", 7},
		{"let sum = 0; for (item in []) { sum = sum + item; } sum;", 0},
		{"let f = fn() { [3, 4]; }; let sum = 0; for (i in f()) { sum = sum + i; } sum;", 7},
		{"let i = 100; for (i in range(3)) { i = 999; } i;", 100},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			testIntegerObject(t, testEval(tt.input), tt.want)
		})
	}
}

func TestClassicForEvaluation(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{"let sum = 0; for (i = 0; i < 10; i++) { sum = sum + i; } sum;", 45},
		{"let sum = 0; for (let i = 0; i < 4; i++) { sum = sum + i; } sum;", 6},
		{"let sum = 0; for (i = 0; i < 4; i = i + 1) { sum = sum + i; } sum;", 6},
		{"let i = 0; for (; i < 4; i++) { } i;", 4},
		{"let sum = 0; for (i = 0; i < 4;) { sum = sum + i; i = i + 1; } sum;", 6},
		{"let sum = 0; for (i = 0; i < 0; i++) { sum = sum + 1; } sum;", 0},
		{"let i = 100; for (i = 0; i < 3; i++) { } i;", 100},
		{"let sum = 0; for (i = 0; i < 3 && true; i++) { sum = sum + i; } sum;", 3},
		{"let sum = 0; for (i = 0; i < 2; i++) { for (j = 0; j < 3; j++) { sum = sum + 1; } } sum;", 6},
		{"let sum = 0; for (i in range(2)) { for (j = 0; j < 3; j++) { sum = sum + 1; } } sum;", 6},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			testIntegerObject(t, testEval(tt.input), tt.want)
		})
	}
}

func TestForLoopReturn(t *testing.T) {
	inputs := []string{
		"let find = fn() { for (i in range(10)) { if (i == 4) { return i; } } return 99; }; find();",
		"let find = fn() { for (i = 0; i < 10; i++) { if (i == 4) { return i; } } return 99; }; find();",
	}
	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			testIntegerObject(t, testEval(input), 4)
		})
	}
}

func TestForLoopErrors(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"for (i in 42) { i; }", "object is not iterable: INTEGER"},
		{"for (i in true) { i; }", "object is not iterable: BOOLEAN"},
		{"for (i in missing_values) { i; }", "identifier not found: missing_values"},
		{"for (i in range(1, 4, 0)) { i; }", "range step cannot be zero"},
		{"for (i = 0; missing_condition; i++) { i; }", "identifier not found: missing_condition"},
		{"for (i = 0; i < 3; missing++) { i; }", "identifier not found: missing"},
		{"for (i = 0; i < 3; i++) { missing; }", "identifier not found: missing"},
		{"let i = true; i++;", "increment requires INTEGER, got BOOLEAN"},
		{"42++;", "left side of ++ must be an identifier"},
		{"missing++;", "identifier not found: missing"},
		{"let i = 9223372036854775807; i++;", "integer overflow in increment: i++"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := testEval(tt.input)
			err, ok := got.(*object.Error)
			if !ok || err.Message != tt.want {
				t.Fatalf("expected error %q, got %T (%+v)", tt.want, got, got)
			}
		})
	}
}

func TestPostfixIncrementValue(t *testing.T) {
	testIntegerObject(t, testEval("let i = 3; let old = i++; old;"), 3)
	testIntegerObject(t, testEval("let i = 3; i++; i;"), 4)
}

func TestLoopLocalBindingsAndOuterAssignments(t *testing.T) {
	input := "let i = 42; let sum = 0; for (i = 0; i < 3; i++) { sum = sum + i; } i + sum;"
	testIntegerObject(t, testEval(input), 45)
	input = "let i = 42; let sum = 0; for (i in range(3)) { sum = sum + i; } i + sum;"
	testIntegerObject(t, testEval(input), 45)
}

func TestForExpressionReturnsNull(t *testing.T) {
	for _, input := range []string{
		"for (i in []) {}",
		"for (i = 0; i < 2; i++) {}",
	} {
		result := testEval(input)
		if result != NULL {
			t.Errorf("%q: expected null, got %T (%+v)", input, result, result)
		}
	}
}
