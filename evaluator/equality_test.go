package evaluator

import (
	"monkey-go/object"
	"testing"
)

type equalityCase struct {
	name, setup, left, right string
	equal                    bool
}

func TestEqualityValuesAndReferences(t *testing.T) {
	tests := []equalityCase{
		{"integer value", "", "1 + 1", "2", true},
		{"integer difference", "", "1", "2", false},
		{"integer minimum", "", minInt64Expression, minInt64Expression, true},
		{"integer bounds", "", minInt64Expression, maxInt64Expression, false},
		{"boolean value", "", "1 < 2", "true", true},
		{"boolean difference", "", "true", "false", false},
		{"empty string", "", `""`, `""`, true},
		{"computed Unicode string", "", `"日" + "本"`, `"日本"`, true},
		{"Unicode string difference", "", `"日本"`, `"日木"`, false},
		{"string case", "", `"Monkey"`, `"monkey"`, false},
		{"string normalization", "", `"é"`, `"é"`, false},
		{"integer and boolean", "", "0", "false", false},
		{"integer and string", "", "1", `"1"`, false},
		{"boolean and string", "", "true", `"true"`, false},
		{"array alias", "let a = [1]; let b = a;", "a", "b", true},
		{"separate arrays", "", "[1]", "[1]", false},
		{"empty arrays", "", "[]", "[]", false},
		{"nested array alias", "let a = [1];", "[a][0]", "a", true},
		{"hash alias", `let a = {"key": 1}; let b = a;`, "a", "b", true},
		{"separate hashes", "", `{"key": 1}`, `{"key": 1}`, false},
		{"empty hashes", "", "{}", "{}", false},
		{"array and hash", "", "[]", "{}", false},
		{"function alias", "let f = fn(x) { x }; let g = f;", "f", "g", true},
		{"separate functions", "", "fn(x) { x }", "fn(x) { x }", false},
		{"separate closures", "let make = fn() { fn(x) { x } };", "make()", "make()", false},
		{"builtin alias", "let f = len;", "f", "len", true},
		{"separate builtins", "", "first", "last", false},
		{"function and builtin", "", "fn(x) { len(x) }", "len", false},
		{"range alias", "let a = range(3); let b = a;", "a", "b", true},
		{"separate ranges", "", "range(3)", "range(3)", false},
		{"array and range", "", "[0, 1, 2]", "range(3)", false},
		{"null and integer", "", "first([])", "0", false},
		{"null and false", "", "first([])", "false", false},
		{"null and string", "", "first([])", `""`, false},
		{"null and array", "", "first([])", "[]", false},
		{"null in an array", "let a = [first([])];", "a[0]", "first([])", true},
		{"null in a hash", `let a = {"key": first([])};`, `a["key"]`, `a["missing"]`, true},
		{"null through a call", "let f = fn(x) { x };", "f(first([]))", "first([])", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkEqualityPair(t, tt)
		})
	}
}

func checkEqualityPair(t *testing.T, tt equalityCase) {
	t.Helper()
	for _, operands := range [][2]string{{tt.left, tt.right}, {tt.right, tt.left}} {
		lhs, rhs := "("+operands[0]+")", "("+operands[1]+")"
		t.Run(lhs+" == "+rhs, func(t *testing.T) {
			testBooleanObject(t, evalOperatorInput(t, tt.setup+lhs+" == "+rhs), tt.equal)
		})
		t.Run(lhs+" != "+rhs, func(t *testing.T) {
			testBooleanObject(t, evalOperatorInput(t, tt.setup+lhs+" != "+rhs), !tt.equal)
		})
	}
}

func TestEqualityUsesHostScalarValues(t *testing.T) {
	tests := []struct {
		name        string
		left, right object.Object
		equal       bool
	}{
		{"integer zero values", &object.Integer{}, &object.Integer{}, true},
		{"false zero values", &object.Boolean{}, &object.Boolean{}, true},
		{"false and canonical false", &object.Boolean{}, FALSE, true},
		{"true and canonical true", &object.Boolean{Value: true}, TRUE, true},
		{"different boolean values", &object.Boolean{}, &object.Boolean{Value: true}, false},
		{"string zero values", &object.String{}, &object.String{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := object.NewEnvironment()
			env.Set("left", tt.left)
			env.Set("right", tt.right)
			testBooleanObject(t, evalNamed(t, "host.mk", "left == right", env), tt.equal)
			testBooleanObject(t, evalNamed(t, "host.mk", "left != right", env), !tt.equal)
		})
	}
}

func TestEqualityEvaluatesOperandsOnceInOrder(t *testing.T) {
	for _, operator := range []string{"==", "!="} {
		t.Run(operator, func(t *testing.T) {
			env := object.NewEnvironment()
			input := `let order = 0;
let left = fn() { order = order * 10 + 1; "a" + "b"; };
let right = fn() { order = order * 10 + 2; "ab"; };
left() ` + operator + ` right();`
			testBooleanObject(t, evalNamed(t, "order.mk", input, env), operator == "==")
			order, _ := env.Get("order")
			testIntegerObject(t, order, 12)
		})
	}
}

func TestEqualityPreservesOperandErrors(t *testing.T) {
	for _, operator := range []string{"==", "!="} {
		t.Run(operator, func(t *testing.T) {
			env := object.NewEnvironment()
			definition := "let fail = fn() {\n    1 / 0\n};\nlet count = 0;\nlet next = fn() { count = count + 1; 1; };"
			evalNamed(t, "definition.mk", definition, env)
			checkEqualityOperandError(t, env, "fail() "+operator+" next()", 0)
			checkEqualityOperandError(t, env, "next() "+operator+" fail()", 1)
		})
	}
}

func checkEqualityOperandError(t *testing.T, env *object.Environment, input string, wantCount int64) {
	t.Helper()
	err := runtimeError(t, evalNamed(t, "equality.mk", input, env))
	if err.Message != "division by zero" {
		t.Errorf("operand error = %q, want division by zero", err.Message)
	}
	if err.Span.Source == nil {
		t.Fatal("operand error lost its source")
	}
	type location struct {
		file         string
		line, column int
	}
	got := location{err.Span.Source.Name, err.Span.Start.Line, err.Span.Start.Column}
	if want := (location{"definition.mk", 2, 7}); got != want {
		t.Errorf("operand location = %+v, want %+v", got, want)
	}
	checkFrameNames(t, err.Stack, []string{"fail"})
	count, _ := env.Get("count")
	testIntegerObject(t, count, wantCount)
}

func TestImplicitNullCompatibility(t *testing.T) {
	inputs := []string{
		"fn() {}()", "if (true) {}", "if (false) { 1 }", "print()",
		"fn() { for i in [] {} }()", "fn() { for i in [1] { i; } }()",
		"[1][-1]", "[1][1]", "[][0]", `{}["missing"]`,
		"first([])", "last([])", "rest([])",
	}
	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			program := "let absent = " + input + "; [absent, absent == fn() {}(), absent != fn() {}(), !absent, absent && missing, absent || true];"
			value := evalOperatorInput(t, program)
			array, ok := value.(*object.Array)
			if !ok || len(array.Elements) != 6 {
				t.Fatalf("expected compatibility results, got %T (%+v)", value, value)
			}
			testNullObject(t, array.Elements[0])
			for i, want := range []bool{true, false, true, false, true} {
				testBooleanObject(t, array.Elements[i+1], want)
			}
		})
	}
	value := evalOperatorInput(t, "rest([1])")
	array, ok := value.(*object.Array)
	if !ok || len(array.Elements) != 0 {
		t.Fatalf("rest of a singleton must remain an empty array, got %T (%+v)", value, value)
	}
}

func TestNullRemainsAnIdentifierAndUnhashableValue(t *testing.T) {
	testIntegerObject(t, evalOperatorInput(t, "let null = 42; null"), 42)
	checkOperatorError(t, "null", "identifier not found: null")
	checkOperatorError(t, "{first([]): 1}", "unusable as hash key: NULL")
	checkOperatorError(t, "{}[first([])]", "unusable as hash key: NULL")
	checkOperatorError(t, "len(first([]))", "argument to `len` not supported NULL")
}
