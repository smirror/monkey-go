package evaluator

import (
	"monkey-go/object"
	"testing"
)

func TestMacroReturnExitsCallerFromExpressionPositions(t *testing.T) {
	tests := []string{
		`leave();`,
		`return leave();`,
		`let local = leave();`,
		`value = leave();`,
		`-leave();`,
		`leave() + tick();`,
		`1 + leave();`,
		`leave() && tick();`,
		`true && leave();`,
		`leave() || tick();`,
		`false || leave();`,
		`if (leave()) { tick(); };`,
		`leave()(tick());`,
		`sink(leave(), tick());`,
		`[leave(), tick()];`,
		`leave()[tick()];`,
		`[1][leave()];`,
		`{leave(): tick()};`,
		`{1: leave()};`,
		`for i in leave() { tick(); }`,
		`for i in range(1) { leave(); }`,
	}
	for _, expression := range tests {
		t.Run(expression, func(t *testing.T) {
			checkMacroReturn(t, expression)
		})
	}
}

func checkMacroReturn(t *testing.T, expression string) {
	t.Helper()
	calls := 0
	env := object.NewEnvironment()
	env.Set("value", &object.Integer{Value: 1})
	env.Set("tick", &object.Builtin{Fn: func(args ...object.Object) object.Object {
		calls++
		return &object.Integer{Value: 1}
	}})
	input := `let leave = macro() { return 7; };
let sink = fn(a, b) { 0; };
let caller = fn() { ` + expression + ` tick(); 99; };
caller();`
	expanded := expandMacroProgram(t, input, object.NewEnvironment())
	testIntegerObject(t, Eval(expanded, env), 7)
	if calls != 0 {
		t.Fatalf("executed %d effects after returning from a template", calls)
	}
	value, _ := env.Get("value")
	testIntegerObject(t, value, 1)
}

func TestMacroReturnRespectsGeneratedFunctionBoundary(t *testing.T) {
	input := `let make = macro(x) { fn() { return x; }; };
let caller = fn() { let generated = make(7); generated() + 1; };
caller();`
	expanded := expandMacroProgram(t, input, object.NewEnvironment())
	testIntegerObject(t, Eval(expanded, object.NewEnvironment()), 8)
}

func TestMacroReturnDoesNotBecomeLetBinding(t *testing.T) {
	input := `let leave = macro() { return 7; };
let captured = leave();`
	expanded := expandMacroProgram(t, input, object.NewEnvironment())
	env := object.NewEnvironment()
	testIntegerObject(t, Eval(expanded, env), 7)
	if _, exists := env.Get("captured"); exists {
		t.Fatal("a returning macro created a runtime binding")
	}
}

func TestQuoteNamesAreOrdinaryRuntimeFunctions(t *testing.T) {
	input := `let quote = fn(x) { x + 1; };
let unquote = fn(x) { x * 2; };
quote(unquote(3));`
	testIntegerObject(t, Eval(parseMacroProgram(t, input), object.NewEnvironment()), 7)
}

func TestRuntimeMacroLiteralsAreRejected(t *testing.T) {
	tests := []string{
		`macro() { 1; };`,
		`macro(x) { x; }(1);`,
		`let f = fn() { let local = macro() { 1; }; local(); }; f();`,
		`let make = macro() { macro() { 1; }; }; make();`,
	}
	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			expanded := expandMacroProgram(t, input, object.NewEnvironment())
			err := runtimeError(t, Eval(expanded, object.NewEnvironment()))
			if err.Message != "macro literals are only supported in top-level declarations" {
				t.Fatalf("unexpected macro literal error: %s", err.Message)
			}
		})
	}
}

func TestTemplateIdentifiersCanBeAssignmentTargets(t *testing.T) {
	input := `let identity = macro(x) { x; };
let nested = macro(x) { identity(x); };
let assign = macro(x, value) { x = value; };
let value = 1;
nested(value) = 2;
assign(value, value + 1);
value;`
	expanded := expandMacroProgram(t, input, object.NewEnvironment())
	testIntegerObject(t, Eval(expanded, object.NewEnvironment()), 3)
}

func TestTemplateAssignmentTargetsMustBeIdentifiers(t *testing.T) {
	tests := []string{
		`let target = macro() { 1 + 2; }; target() = 3;`,
		`let target = macro() { let scratch = 1; scratch; }; target() = 3;`,
		`let target = macro() { return 1; }; target() = 3;`,
		`let target = macro() {}; target() = 3;`,
	}
	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			expanded := expandMacroProgram(t, input, object.NewEnvironment())
			env := object.NewEnvironment()
			err := runtimeError(t, Eval(expanded, env))
			if err.Message != "left side of assignment must be an identifier" {
				t.Fatalf("unexpected assignment error: %s", err.Message)
			}
			if _, exists := env.Get("scratch"); exists {
				t.Fatal("invalid assignment target executed the template")
			}
		})
	}
}
