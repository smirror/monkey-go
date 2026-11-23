package evaluator

import (
	"monkey-go/ast"
	"monkey-go/diagnostic"
	"monkey-go/object"
	"strings"
	"testing"
)

func TestWhileReevaluatesConditionBeforeEveryBody(t *testing.T) {
	for _, tt := range []struct {
		input string
		want  int64
	}{
		{"let n = 0; while n < 3 { n = n + 1; } n;", 3},
		{"let n = 0; while (n < 3) { n = n + 1; } n;", 3},
		{"let n = 0; while (n = n + 1) < 4 {} n;", 4},
		{"let n = 0; while false { n = n + 1; missing; } n;", 0},
		{"let n = 0; let checks = 0; let ready = fn() { checks = checks + 1; n < 3; }; while ready() { n = n + 1; } checks * 10 + n;", 43},
	} {
		t.Run(tt.input, func(t *testing.T) {
			testIntegerObject(t, evalNamed(t, "while.mk", tt.input, object.NewEnvironment()), tt.want)
		})
	}
}

func TestWhileUsesIfTruthinessAndExistingLoopResult(t *testing.T) {
	for _, condition := range []string{"true", "0", "\"\"", "[]", "{}"} {
		input := "let ran = 0; while " + condition + " { ran = ran + 1; break; } ran;"
		testIntegerObject(t, evalNamed(t, "truthy.mk", input, object.NewEnvironment()), 1)
	}
	for _, input := range []string{
		"while false { missing; }",
		"while if (false) { true; } { missing; }",
		"while true { break; missing; }",
		"let n = 0; while n < 2 { n = n + 1; continue; missing; }",
	} {
		if value := evalNamed(t, "result.mk", input, object.NewEnvironment()); value != NULL {
			t.Fatalf("while changed the existing loop result: %T (%v)", value, value)
		}
	}
}

func TestWhileBreakAndContinue(t *testing.T) {
	for _, tt := range []struct {
		input string
		want  int64
	}{
		{"let n = 0; let sum = 0; while n < 6 { n = n + 1; if (n == 2) { continue; } if (n == 5) { break; } sum = sum + n; } sum;", 8},
		{"let checks = 0; let ready = fn() { checks = checks + 1; true; }; while ready() { break; } checks;", 1},
		{"let n = 0; let checks = 0; let ready = fn() { checks = checks + 1; n < 3; }; while ready() { n = n + 1; continue; missing; } checks;", 4},
		{"let n = 0; while n < 3 { n = n + 1; let value = if (true) { continue; }; n = 99; } n;", 3},
		{"let n = 0; while true { n = n + 1; let value = if (true) { break; }; n = 99; } n;", 1},
	} {
		t.Run(tt.input, func(t *testing.T) {
			testIntegerObject(t, evalNamed(t, "control.mk", tt.input, object.NewEnvironment()), tt.want)
		})
	}
}

func TestWhileAndForConsumeOnlyTheirOwnBodyControl(t *testing.T) {
	for _, tt := range []struct {
		input string
		want  int64
	}{
		{"let count = 0; for i in range(3) { let n = 0; while n < 3 { n = n + 1; if (n == 2) { break; } count = count + 1; } count = count + 10; } count;", 33},
		{"let count = 0; let n = 0; while n < 3 { n = n + 1; for i in range(3) { if (i == 1) { continue; } count = count + 1; } count = count + 10; } count;", 36},
		{"let count = 0; let n = 0; while n < 3 { n = n + 1; while true { count = count + 1; break; } count = count + 10; } count;", 33},
		{"let count = 0; for i in range(3) { let n = 0; while n < 3 { n = n + 1; if (n < 3) { continue; } count = count + 1; } count = count + 10; } count;", 33},
	} {
		t.Run(tt.input, func(t *testing.T) {
			testIntegerObject(t, evalNamed(t, "nested.mk", tt.input, object.NewEnvironment()), tt.want)
		})
	}
}

func TestWhileConditionControlBelongsToTheEnclosingLoop(t *testing.T) {
	for _, tt := range []struct {
		input string
		want  int64
	}{
		{"let n = 0; for i in range(3) { n = n + 1; while if (true) { break; } {} n = 99; } n;", 1},
		{"let n = 0; for i in range(3) { n = n + 1; while if (true) { continue; } {} n = 99; } n;", 3},
		{"let n = 0; while n < 3 { n = n + 1; while if (true) { continue; } {} n = 99; } n;", 3},
		{"let n = 0; while n < 3 { n = n + 1; for i in if (true) { break; } {} n = 99; } n;", 1},
	} {
		testIntegerObject(t, evalNamed(t, "condition-control.mk", tt.input, object.NewEnvironment()), tt.want)
	}
}

func TestWhileBodyScopeIsLocalAndReused(t *testing.T) {
	for _, tt := range []struct {
		input string
		want  int64
	}{
		{"let n = 0; let enabled = true; while enabled && n < 2 { let enabled = false; n = n + 1; } if (enabled) { n; } else { 99; }", 2},
		{"let n = 0; let seen = 0; while n < 3 { if (n == 0) { let local = 10; } local = local + 1; seen = local; n = n + 1; } seen;", 13},
		{"let outer = 7; let n = 0; while n < 1 { let outer = 99; n = n + 1; } outer;", 7},
	} {
		testIntegerObject(t, evalNamed(t, "scope.mk", tt.input, object.NewEnvironment()), tt.want)
	}
	for _, input := range []string{
		"while true { let local = 1; break; } local;",
		"let n = 0; while n == 0 || local { let local = false; n = n + 1; }",
	} {
		err := runtimeError(t, evalNamed(t, "scope.mk", input, object.NewEnvironment()))
		if err.Message != "identifier not found: local" {
			t.Fatalf("body binding leaked into an outer scope: %+v", err)
		}
	}
}

func TestWhilePreservesReturnsFromConditionAndBody(t *testing.T) {
	for _, input := range []string{
		"let f = fn() { while true { return 7; } return 99; }; f();",
		"let f = fn() { while if (true) { return 7; } { missing; } return 99; }; f();",
		"let f = fn() { while true { for i in [1] { return 7; } } }; f();",
		"let f = fn() { for i in [1] { while true { return 7; } } }; f();",
		"let f = fn() { while true { let value = if (true) { return 7; }; } }; f();",
	} {
		testIntegerObject(t, evalNamed(t, "return.mk", input, object.NewEnvironment()), 7)
	}
}

func TestWhilePreservesRuntimeLocationAndCallStack(t *testing.T) {
	for _, source := range []string{
		"let fail = fn() { while 1 / 0 {} };",
		"let fail = fn() { while true {\n  1 / 0;\n} };",
		"let fail = fn() { let n = 1; while 1 / n { n = n - 1; } };",
	} {
		env := object.NewEnvironment()
		evalNamed(t, "definition.mk", source, env)
		err := runtimeError(t, evalNamed(t, "call.mk", "fail();", env))
		if err.Kind != diagnostic.Runtime || err.Message != "division by zero" {
			t.Fatalf("while replaced the original error: %+v", err)
		}
		if err.Span.Source.Name != "definition.mk" || err.Span.Source.Text != source {
			t.Fatalf("while lost the definition source: %+v", err.Span)
		}
		if err.Span.Start.Offset != strings.Index(source, "/") {
			t.Fatalf("while error does not point at the failed operator: %+v", err.Span)
		}
		checkFrameNames(t, err.Stack, []string{"fail"})
		if err.Stack[0].Span.Source.Name != "call.mk" {
			t.Fatal("while error lost its call site")
		}
	}
}

func TestWhileCannotConsumeInvalidControlAcrossFunctionBoundary(t *testing.T) {
	for _, keyword := range []string{"break", "continue"} {
		for _, source := range []string{"while true { bad(); }", "while bad() {}"} {
			env := object.NewEnvironment()
			control := handBuiltControl(keyword)
			env.Set("bad", &object.Function{Env: env, Body: &ast.BlockStatement{Statements: []ast.Statement{control}}})
			err := runtimeError(t, evalNamed(t, "caller.mk", source, env))
			checkControlBoundaryError(t, err, keyword, control.Span())
			checkFrameNames(t, err.Stack, []string{"bad"})
		}
	}
}
