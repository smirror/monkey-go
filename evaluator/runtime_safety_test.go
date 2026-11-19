package evaluator

import (
	"monkey-go/object"
	"testing"
)

func TestRuntimeBoundaryErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"divide by zero", "1 / 0", "division by zero"},
		{"divide zero by zero", "0 / 0", "division by zero"},
		{"division in loop", "for i in range(3) { 10 / (2 - i); }", "division by zero"},
		{"add overflow", "9223372036854775807 + 1", "integer overflow: 9223372036854775807 + 1"},
		{"subtract overflow", "(-9223372036854775807 - 1) - 1", "integer overflow: -9223372036854775808 - 1"},
		{"multiply overflow", "3037000500 * 3037000500", "integer overflow: 3037000500 * 3037000500"},
		{"multiply min by minus one", "(-9223372036854775807 - 1) * -1", "integer overflow: -9223372036854775808 * -1"},
		{"divide min by minus one", "(-9223372036854775807 - 1) / -1", "integer overflow: -9223372036854775808 / -1"},
		{"negate min", "-(-9223372036854775807 - 1)", "integer overflow in negation: -9223372036854775808"},
		{"missing argument", "fn(a, b) { a + b }(1)", "wrong number of arguments. got=1, want=2"},
		{"extra argument", "fn(a) { a }(1, 2)", "wrong number of arguments. got=2, want=1"},
		{"extra argument to zero-arg function", "fn() { 42 }(1)", "wrong number of arguments. got=1, want=0"},
		{"invalid array index type", "[1][true]", "index operator not supported: ARRAY"},
		{"null in arithmetic", "let f = fn() {}; f() + 1", "type mismatch: NULL + INTEGER"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertMonkeyError(t, tt.input, tt.want)
		})
	}
}

func assertMonkeyError(t *testing.T, input, want string) {
	t.Helper()
	got := testEval(input)
	err, ok := got.(*object.Error)
	if !ok {
		t.Fatalf("expected a Monkey error %q, got %T (%+v)", want, got, got)
	}
	if err.Message != want {
		t.Fatalf("expected Monkey error %q, got %q", want, err.Message)
	}
}

func TestEmptyBlockReturnsNullObject(t *testing.T) {
	inputs := []string{
		"fn() {}()",
		"if (true) {}",
		"let f = fn() {}; [f()][0]",
	}
	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			if result := testEval(input); result != NULL {
				t.Fatalf("expected NULL, got %T (%+v)", result, result)
			}
		})
	}
}

func TestIntegerOperationsAtBounds(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{"9223372036854775807 - 1 + 1", 9223372036854775807},
		{"(-9223372036854775807 - 1) + 1", -9223372036854775807},
		{"(-9223372036854775807 - 1) / 1", -9223372036854775807 - 1},
		{"0 * 9223372036854775807", 0},
		{"10 / -2", -5},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			testIntegerObject(t, testEval(tt.input), tt.want)
		})
	}
}
