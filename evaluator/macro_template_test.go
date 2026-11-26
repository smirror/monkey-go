package evaluator

import (
	"monkey-go/ast"
	"monkey-go/object"
	"testing"
)

func TestMacroTemplates(t *testing.T) {
	cases := []struct {
		name   string
		source string
		want   int64
	}{
		{"reverse", `let reverse = macro(a, b) { b - a; }; reverse(2 + 2, 10 - 5);`, 1},
		{"arithmetic", `let twice = macro(x) { x + x; }; twice(3 + 1);`, 8},
		{"multiple statements", `let next = macro(x) { let value = x; value + 1; }; next(6);`, 7},
		{"caller bindings", `let bind = macro(x) { let value = x; }; bind(6); value;`, 6},
		{"runtime conditional", `let unless = macro(condition, body) { if (!condition) { body; } }; let calls = 0; unless(true, calls = calls + 1); calls;`, 0},
		{"repeated effects", `let calls = 0; let tick = fn() { calls = calls + 1; calls; }; let twice = macro(x) { x + x; }; let value = twice(tick()); value * 10 + calls;`, 32},
		{"simultaneous substitution", `let x = 7; let y = 12; let sub = macro(x, y) { x - y; }; sub(y, x);`, 5},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			expanded := expandMacroProgram(t, tc.source, object.NewEnvironment())
			testIntegerObject(t, Eval(expanded, object.NewEnvironment()), tc.want)
		})
	}
}

func TestEmptyMacroTemplate(t *testing.T) {
	expanded := expandMacroProgram(t, `let empty = macro() {}; empty();`, object.NewEnvironment())
	if got := Eval(expanded, object.NewEnvironment()); got != NULL {
		t.Fatalf("empty template = %v, want null", got)
	}
}

func TestTemplateParametersRespectBindings(t *testing.T) {
	cases := []struct {
		name   string
		source string
		want   int64
	}{
		{"let initializer", `let m = macro(x) { let x = x + 1; x; }; m(4);`, 5},
		{"function parameter", `let m = macro(x) { fn(x) { x + 1; }(5) + x; }; m(10);`, 16},
		{"function local", `let m = macro(x) { fn() { let x = 3; x; }() + x; }; m(10);`, 13},
		{"recursive function", `let m = macro(f) { let f = fn(n) { if (n == 0) { 0; } else { f(n - 1) + 1; } }; f(3); }; m(999);`, 3},
		{"loop iterable", `let m = macro(x) { let sum = 0; for x in x { sum = sum + x; } sum; }; m([1, 2, 3]);`, 6},
		{"loop variable", `let m = macro(x) { for x in [1] {} x; }; m(7);`, 7},
		{"loop local", `let m = macro(x) { for i in [1] { let x = 10; } x; }; m(7);`, 7},
		{"conditional alternatives", `let m = macro(x) { if (false) { let x = 10; x; } else { x; } }; m(7);`, 7},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			expanded := expandMacroProgram(t, tc.source, object.NewEnvironment())
			testIntegerObject(t, Eval(expanded, object.NewEnvironment()), tc.want)
		})
	}
}

func TestRuntimeBindingsShadowMacroNames(t *testing.T) {
	cases := []string{
		`let m = macro() { 99; }; let m = fn() { 7; }; m();`,
		`let m = macro() { 99; }; let f = fn(m) { m(); }; f(fn() { 7; });`,
		`let m = macro() { 99; }; let f = fn() { let m = fn() { 7; }; m(); }; f();`,
		`let m = macro() { 99; }; let value = 0; for m in [fn() { 7; }] { value = m(); } value;`,
		`let m = macro() { 99; }; let define = macro() { let m = fn() { 7; }; }; define(); m();`,
		`let m = macro() { 99; }; let define = macro() { let m = fn() { 7; }; }; let nested = macro() { define(); }; nested(); m();`,
	}
	for _, source := range cases {
		t.Run(source, func(t *testing.T) {
			expanded := expandMacroProgram(t, source, object.NewEnvironment())
			testIntegerObject(t, Eval(expanded, object.NewEnvironment()), 7)
		})
	}
}

func TestGeneratedBindingsApplyBeforeLaterMacroArityChecks(t *testing.T) {
	input := `let define = macro() { let inc = fn(x) { x * 2; }; };
let inc = macro() { 1; };
define(); inc(2);`
	expanded := expandMacroProgram(t, input, object.NewEnvironment())
	testIntegerObject(t, Eval(expanded, object.NewEnvironment()), 4)
}

func TestNestedMacroParametersAreNotOuterPlaceholders(t *testing.T) {
	program := parseMacroProgram(t, `let outer = macro(x, y) { macro(x) { x + y; }; }; outer(7, 9);`)
	env := object.NewEnvironment()
	DefineMacros(program, env)
	expanded, err := ExpandMacros(program, env)
	if err != nil {
		t.Fatal(err)
	}
	invocation := expanded.(*ast.Program).Statements[0].(*ast.ExpressionStatement).Expression.(*ast.MacroExpansion)
	nested := invocation.Body.Statements[0].(*ast.ExpressionStatement).Expression.(*ast.MacroLiteral)
	if nested.String() != "macro(x) {(x + 9)}" {
		t.Fatalf("nested parameter binding was replaced: %s", nested.String())
	}
}

func TestTemplateExpansionCopiesArgumentsAndDefinitions(t *testing.T) {
	env := object.NewEnvironment()
	program := parseMacroProgram(t, `let twice = macro(x) { x + x; }; twice(7);`)
	DefineMacros(program, env)
	original := program.String()
	call := program.Statements[0].(*ast.ExpressionStatement).Expression.(*ast.CallExpression)
	argument := call.Arguments[0].(*ast.IntegerLiteral)
	expanded, err := ExpandMacros(program, env)
	if err != nil {
		t.Fatal(err)
	}
	invocation := expanded.(*ast.Program).Statements[0].(*ast.ExpressionStatement).Expression.(*ast.MacroExpansion)
	sum := invocation.Body.Statements[0].(*ast.ExpressionStatement).Expression.(*ast.InfixExpression)
	sum.Left.(*ast.IntegerLiteral).Value = 100
	if sum.Right.(*ast.IntegerLiteral).Value != 7 {
		t.Fatal("expanded placeholders shared mutable syntax with each other")
	}
	if argument.Value != 7 {
		t.Fatal("expanded placeholder shared mutable syntax with the original argument")
	}
	if program.String() != original {
		t.Fatal("expansion modified the original program")
	}
	next := expandMacroProgram(t, "twice(2)", env)
	testIntegerObject(t, Eval(next, object.NewEnvironment()), 4)
}
