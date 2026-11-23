package evaluator

import (
	"monkey-go/lexer"
	"monkey-go/object"
	"monkey-go/parser"
	"testing"
)

const minInt64Expression = "(-9223372036854775807 - 1)"
const maxInt64Expression = "9223372036854775807"

func TestIntegerModuloSignsAndBounds(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{"0 % 3", 0},
		{"5 % 3", 2},
		{"-5 % 3", -2},
		{"5 % -3", 2},
		{"-5 % -3", -2},
		{"6 % 3", 0},
		{"-6 % -3", 0},
		{minInt64Expression + " % -1", 0},
		{minInt64Expression + " % 1", 0},
		{minInt64Expression + " % 3", -2},
		{minInt64Expression + " % " + maxInt64Expression, -1},
		{maxInt64Expression + " % " + minInt64Expression, 9223372036854775807},
		{minInt64Expression + " % " + minInt64Expression, 0},
		{maxInt64Expression + " % -3", 1},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			testIntegerObject(t, evalOperatorInput(t, tt.input), tt.want)
		})
	}
}

func TestIntegerInclusiveComparisons(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"1 <= 2", true},
		{"2 <= 2", true},
		{"3 <= 2", false},
		{"1 >= 2", false},
		{"2 >= 2", true},
		{"3 >= 2", true},
		{"-5 <= -3", true},
		{"-3 >= -5", true},
		{maxInt64Expression + " >= " + minInt64Expression, true},
		{minInt64Expression + " <= " + maxInt64Expression, true},
		{minInt64Expression + " >= " + minInt64Expression, true},
		{maxInt64Expression + " <= " + maxInt64Expression, true},
		{maxInt64Expression + " <= " + minInt64Expression, false},
		{minInt64Expression + " >= " + maxInt64Expression, false},
		{"1 + 2 <= 7 % 4 == true", true},
		{"10 % 4 >= 2 && 3 <= 3", true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			testBooleanObject(t, evalOperatorInput(t, tt.input), tt.want)
		})
	}
}

func TestModuloAssociativityAndOperandEvaluation(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{"20 % 6 * 3", 6},
		{"20 / 3 % 4", 2},
		{"20 % 6 / 2", 1},
		{"20 % 6 % 3", 2},
		{"17 % 5 + 2 * 3", 8},
		{"2 + 3 % 2", 3},
		{"let n = 0; let next = fn() { n = n + 1; n; }; let r = next() % next(); n * 10 + r;", 21},
		{"let n = 0; let next = fn() { n = n + 1; n; }; let r = next() <= next(); if (r) { n } else { 99 }", 2},
		{"let n = 0; let next = fn() { n = n + 1; n; }; let r = next() >= next(); if (r) { 99 } else { n }", 2},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			testIntegerObject(t, evalOperatorInput(t, tt.input), tt.want)
		})
	}
}

func TestComparisonAndModuloInForInLoops(t *testing.T) {
	input := `let total = 0;
for i in range(-5, 6) {
    if (i % 2 == 0 && i >= -2 && i <= 4) { total = total + i; }
}
total;`
	testIntegerObject(t, evalOperatorInput(t, input), 4)
	checkOperatorError(t, "for i in [2, 0, 3] { 10 % i; }", "modulo by zero")
}

func TestModuloPreservesShortCircuitEvaluation(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"false && 1 % 0 == 0", false},
		{"true || 1 % 0 == 0", true},
		{"0 >= 1 && 1 % 0 == 0", false},
		{"2 <= 2 || 1 % 0 == 0", true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			testBooleanObject(t, evalOperatorInput(t, tt.input), tt.want)
		})
	}
	input := "let n = 0; let f = fn() { n = n + 1; 5; }; false && f() % 2 == 1; true || f() >= 0; n;"
	testIntegerObject(t, evalOperatorInput(t, input), 0)
}

func TestModuloAndComparisonErrors(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"1 % 0", "modulo by zero"},
		{"0 % 0", "modulo by zero"},
		{minInt64Expression + " % 0", "modulo by zero"},
		{"true && 1 % 0 == 0", "modulo by zero"},
		{"false || 1 % 0 == 0", "modulo by zero"},
		{"missing % 0", "identifier not found: missing"},
		{"1 % (1 / 0)", "division by zero"},
		{"true % 2", "type mismatch: BOOLEAN % INTEGER"},
		{"1 % true", "type mismatch: INTEGER % BOOLEAN"},
		{"true % false", "unknown operator: BOOLEAN % BOOLEAN"},
		{"\"a\" % \"b\"", "unknown operator: STRING % STRING"},
		{"[1] % [2]", "unknown operator: ARRAY % ARRAY"},
		{"fn() {}() % 2", "type mismatch: NULL % INTEGER"},
		{"true <= false", "unknown operator: BOOLEAN <= BOOLEAN"},
		{"true >= false", "unknown operator: BOOLEAN >= BOOLEAN"},
		{"\"a\" <= \"b\"", "unknown operator: STRING <= STRING"},
		{"\"a\" >= \"b\"", "unknown operator: STRING >= STRING"},
		{"1 <= true", "type mismatch: INTEGER <= BOOLEAN"},
		{"false >= 1", "type mismatch: BOOLEAN >= INTEGER"},
		{"1 <= 2 <= 3", "type mismatch: BOOLEAN <= INTEGER"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) { checkOperatorError(t, tt.input, tt.want) })
	}
}

func evalOperatorInput(t *testing.T, input string) object.Object {
	t.Helper()
	p := parser.New(lexer.New(input))
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		t.Fatalf("unexpected parser errors: %v", p.Errors())
	}
	return Eval(program, object.NewEnvironment())
}

func checkOperatorError(t *testing.T, input, want string) {
	t.Helper()
	got := evalOperatorInput(t, input)
	err, ok := got.(*object.Error)
	if !ok {
		t.Fatalf("expected Monkey error %q, got %T (%+v)", want, got, got)
	}
	if err.Message != want {
		t.Fatalf("error message = %q, want %q", err.Message, want)
	}
}
