package evaluator

import (
	"monkey-go/ast"
	"monkey-go/diagnostic"
	"monkey-go/object"
	"monkey-go/token"
	"strings"
	"testing"
)

func TestBreakAndContinueInRangesAndArrays(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{"let sum = 0; for i in range(10) { if (i == 3) { break; } sum = sum + i; } sum;", 3},
		{"let sum = 0; for i in range(6) { if (i % 2 == 0) { continue; } sum = sum + i; } sum;", 9},
		{"let sum = 0; for i in [1, 2, 3] { if (i == 2) { break; } sum = sum + i; } sum;", 1},
		{"let sum = 0; for i in [1, 2, 3] { if (i == 2) { continue; } sum = sum + i; } sum;", 4},
		{"let sum = 0; for i in range(5, 0, -1) { if (i == 3) { break; } sum = sum + i; } sum;", 9},
		{"let n = 0; for i in range(3) { false && if (true) { break; }; true || if (true) { continue; }; n = n + 1; } n;", 3},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			value := evalNamed(t, "loops.mk", tt.input, object.NewEnvironment())
			testIntegerObject(t, value, tt.want)
		})
	}
}

func TestNestedLoopsConsumeOnlyTheirOwnControl(t *testing.T) {
	tests := []struct {
		body string
		want int64
	}{
		{"for j in range(3) { count = count + 1; break; } count = count + 10;", 33},
		{"for j in range(3) { if (j == 1) { continue; } count = count + 1; } count = count + 10;", 36},
		{"for j in range(3) { break; } continue; count = 99;", 0},
		{"for j in range(3) { continue; } count = count + 1; break;", 1},
	}
	for _, tt := range tests {
		t.Run(tt.body, func(t *testing.T) {
			input := "let count = 0; for i in range(3) { " + tt.body + " } count;"
			testIntegerObject(t, evalNamed(t, "nested.mk", input, object.NewEnvironment()), tt.want)
		})
	}
}

func TestLoopControlDoesNotChangeReturnOrErrorPropagation(t *testing.T) {
	inputs := []string{
		"let f = fn() { for i in range(4) { if (i == 2) { return 7; } continue; } return 99; }; f();",
		"let f = fn() { for i in range(3) { let x = if (true) { return 7; }; break; } return 99; }; f();",
		"let f = fn() { for i in range(3) { for j in range(2) { return 7; } continue; } }; f();",
	}
	for _, input := range inputs {
		testIntegerObject(t, evalNamed(t, "return.mk", input, object.NewEnvironment()), 7)
	}
	for _, keyword := range []string{"break", "continue"} {
		input := "for i in range(3) { 1 / 0; " + keyword + "; }"
		err := runtimeError(t, evalNamed(t, "error.mk", input, object.NewEnvironment()))
		if err.Message != "division by zero" || err.Span.Start.Offset != strings.Index(input, "/") {
			t.Fatalf("loop control replaced the earlier error: %+v", err)
		}
	}
}

func TestEmptyAndControlledLoopsKeepExistingResult(t *testing.T) {
	inputs := []string{
		"for i in [] { missing; break; }",
		"for i in range(0) { missing; continue; }",
		"for i in range(3) { break; missing; }",
		"for i in range(3) { continue; missing; }",
	}
	for _, input := range inputs {
		if value := evalNamed(t, "empty.mk", input, object.NewEnvironment()); value != NULL {
			t.Fatalf("loop returned %T (%v) instead of its existing null result", value, value)
		}
	}
}

func TestFunctionLocalLoopControlDoesNotAffectCallerLoop(t *testing.T) {
	input := `let inner = fn() {
    let n = 0;
    for j in range(3) { if (j == 1) { break; } n = n + 1; }
    n;
};
let count = 0;
for i in range(3) { count = count + inner(); }
count;`
	testIntegerObject(t, evalNamed(t, "functions.mk", input, object.NewEnvironment()), 3)
}

func TestLoopControlPropagatesThroughExpressionEvaluation(t *testing.T) {
	expressions := []string{
		"let local = CONTROL;", "saved = CONTROL;", "return CONTROL;",
		"-CONTROL;", "CONTROL + effect(0);", "0 + CONTROL;",
		"CONTROL && effect(0);", "true && CONTROL;", "false || CONTROL;",
		"if (CONTROL) { effect(0); }", "(CONTROL)(effect(0));", "effect(CONTROL);",
		"[CONTROL, effect(0)];", "(CONTROL)[effect(0)];", "[1][CONTROL];",
		"{CONTROL: effect(0)};", "{\"key\": CONTROL};",
		"for j in CONTROL { effect(0); }",
	}
	for _, keyword := range []string{"break", "continue"} {
		for _, expression := range expressions {
			t.Run(keyword+"/"+expression, func(t *testing.T) {
				checkControlExpression(t, keyword, expression)
			})
		}
	}
}

func checkControlExpression(t *testing.T, keyword, expression string) {
	t.Helper()
	control := "if (true) { " + keyword + "; }"
	body := strings.ReplaceAll(expression, "CONTROL", control)
	input := "let saved = 9; let calls = 0; let effect = fn(x) { calls = calls + 1; x; }; " +
		"for i in range(3) { " + body + " calls = calls + 100; } saved + calls;"
	testIntegerObject(t, evalNamed(t, "expressions.mk", input, object.NewEnvironment()), 9)
}

func TestPublicEvalRejectsUnconsumedHandBuiltControl(t *testing.T) {
	for _, keyword := range []string{"break", "continue"} {
		control := handBuiltControl(keyword)
		after := &ast.LetStatement{Name: &ast.Identifier{Value: "after"}, Value: &ast.IntegerLiteral{Value: 99}}
		for _, node := range []ast.Node{
			control,
			&ast.BlockStatement{Statements: []ast.Statement{control, after}},
			&ast.Program{Statements: []ast.Statement{control, after}},
		} {
			env := object.NewEnvironment()
			err := runtimeError(t, Eval(node, env))
			checkControlBoundaryError(t, err, keyword, control.Span())
			if _, exists := env.Get("after"); exists {
				t.Fatal("evaluation continued after an invalid loop control")
			}
		}
	}
}

func TestHandBuiltControlCannotCrossFunctionBoundary(t *testing.T) {
	for _, keyword := range []string{"break", "continue"} {
		t.Run(keyword, func(t *testing.T) {
			env := object.NewEnvironment()
			control := handBuiltControl(keyword)
			env.Set("bad", &object.Function{Env: env, Body: &ast.BlockStatement{Statements: []ast.Statement{control}}})
			input := "let visited = 0; for i in range(3) { visited = visited + 1; bad(); }"
			err := runtimeError(t, evalNamed(t, "caller.mk", input, env))
			checkControlBoundaryError(t, err, keyword, control.Span())
			checkFrameNames(t, err.Stack, []string{"bad"})
			if err.Stack[0].Span.Source.Name != "caller.mk" {
				t.Fatal("invalid loop control lost the function call site")
			}
			visited, _ := env.Get("visited")
			testIntegerObject(t, visited, 1)
		})
	}
}

func handBuiltControl(keyword string) ast.Statement {
	var statement ast.Statement = &ast.BreakStatement{}
	if keyword == "continue" {
		statement = &ast.ContinueStatement{}
	}
	source := &token.Source{Name: "handmade.mk", Text: "  " + keyword + ";"}
	ast.SetSpan(statement, token.Span{
		Source: source, Start: token.Position{Offset: 2, Line: 1, Column: 3},
		End: token.Position{Offset: len(source.Text), Line: 1, Column: len(source.Text) + 1},
	})
	return statement
}

func checkControlBoundaryError(t *testing.T, err *object.Error, keyword string, span token.Span) {
	t.Helper()
	if err.Kind != diagnostic.Runtime || err.Message != keyword+" outside loop" || err.Span != span {
		t.Fatalf("invalid control escaped or lost its source: %+v", err)
	}
}
