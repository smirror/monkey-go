package evaluator

import (
	"monkey-go/ast"
	"monkey-go/diagnostic"
	"monkey-go/object"
	"testing"
)

func TestMacroExpansionPreservesComparisonAndModuloInLoops(t *testing.T) {
	input := `let eligible = macro(x) {
    x % 2 == 0 && x >= -2 && x <= 4;
};
let total = 0;
for i in range(-5, 6) {
    if (eligible(i)) { total = total + i; }
}
total;`
	expanded := expandMacroProgram(t, input, object.NewEnvironment())
	testIntegerObject(t, Eval(expanded, object.NewEnvironment()), 4)
}

func TestTemplateComputesComparisonAndModuloAtRuntime(t *testing.T) {
	input := `let remainder = macro() {
    if (2 <= 2 && 3 >= 3) { -5 % 3; } else { 99; };
};
remainder();`
	expanded := expandMacroProgram(t, input, object.NewEnvironment())
	testIntegerObject(t, Eval(expanded, object.NewEnvironment()), -2)
}

func TestExpandedModuloDiagnosticUsesArgumentSource(t *testing.T) {
	env := object.NewEnvironment()
	definition := "let identity = macro(x) { x; };"
	DefineMacros(parseLocatedMacroProgram(t, "definitions.mk", definition), env)
	program := parseLocatedMacroProgram(t, "calls.mk", "identity(7 % 0)")
	call := program.Statements[0].(*ast.ExpressionStatement).Expression.(*ast.CallExpression)
	operator := call.Arguments[0].(*ast.InfixExpression)
	expanded, err := ExpandMacros(program, env)
	if err != nil {
		t.Fatal(err)
	}
	failure := runtimeError(t, Eval(expanded, object.NewEnvironment()))
	want := diagnostic.Diagnostic{
		Kind: diagnostic.Runtime, Message: "modulo by zero", Span: operator.Token.Span,
	}
	got := failure.Diagnostic()
	if got.Kind != want.Kind || got.Message != want.Message || got.Span != want.Span {
		t.Fatalf("expanded modulo diagnostic = %+v, want %+v", got, want)
	}
	if len(got.Stack) != 1 || got.Stack[0].Name != "identity" || got.Stack[0].Span != call.Span() {
		t.Fatalf("expanded modulo diagnostic lost macro call: %+v", got.Stack)
	}
}
