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
		{"let sum = 0; for (i in range(2)) { for (j in range(3)) { sum = sum + i + j; } } sum;", 9},
		{"let calls = 0; let get = fn() { calls = calls + 1; [3, 4]; }; let sum = 0; for (i in get()) { sum = sum + i; } calls * 10 + sum;", 17},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			testIntegerObject(t, testEval(tt.input), tt.want)
		})
	}
}

func TestForInReturn(t *testing.T) {
	inputs := []struct {
		input string
		want  int64
	}{
		{"let find = fn() { for (i in range(10)) { if (i == 4) { return i; } } return 99; }; find();", 4},
		{"let find = fn() { for (i in [1, 2, 3]) { for (j in [4, 5]) { if (i == 2 && j == 5) { return i + j; } } } return 99; }; find();", 7},
	}
	for _, tt := range inputs {
		t.Run(tt.input, func(t *testing.T) {
			testIntegerObject(t, testEval(tt.input), tt.want)
		})
	}
}

func TestForInErrors(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"for (i in 42) { i; }", "object is not iterable: INTEGER"},
		{"for (i in true) { i; }", "object is not iterable: BOOLEAN"},
		{"for (i in missing_values) { i; }", "identifier not found: missing_values"},
		{"for (i in range(1, 4, 0)) { i; }", "range step cannot be zero"},
		{"for (i in range(2)) { missing; }", "identifier not found: missing"},
		{"for (i in [1, true]) { i + 1; }", "type mismatch: BOOLEAN + INTEGER"},
		{"for (i in range(2)) { i; } i;", "identifier not found: i"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := testEval(tt.input)
			err, ok := got.(*object.Error)
			if !ok {
				t.Fatalf("expected error %q, got %T (%+v)", tt.want, got, got)
			}
			if err.Message != tt.want {
				t.Fatalf("expected error %q, got %q", tt.want, err.Message)
			}
		})
	}
}

func TestForInLocalBindingsAndOuterAssignments(t *testing.T) {
	input := "let i = 42; let sum = 0; for (i in range(3)) { sum = sum + i; } i + sum;"
	testIntegerObject(t, testEval(input), 45)
}

func TestForInReturnsNull(t *testing.T) {
	for _, input := range []string{
		"for (i in []) {}",
		"for (i in range(2)) {}",
	} {
		if got := testEval(input); got != NULL {
			t.Errorf("%q: expected null, got %T (%+v)", input, got, got)
		}
	}
}
