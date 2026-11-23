package evaluator

import (
	"monkey-go/ast"
	"monkey-go/diagnostic"
	"monkey-go/object"
	"reflect"
	"testing"
)

func TestMacroExpansionPreservesComparisonAndModuloInLoops(t *testing.T) {
	input := `let eligible = macro(x) {
    quote(unquote(x) % 2 == 0 && unquote(x) >= -2 && unquote(x) <= 4);
};
let total = 0;
for i in range(-5, 6) {
    if (eligible(i)) { total = total + i; }
}
total;`
	expanded := expandMacroProgram(t, input, object.NewEnvironment())
	testIntegerObject(t, Eval(expanded, object.NewEnvironment()), 4)
}

func TestUnquoteComputesComparisonAndModulo(t *testing.T) {
	input := `let remainder = macro() {
    quote(unquote(if (2 <= 2 && 3 >= 3) { -5 % 3; } else { 99; }));
};
remainder();`
	expanded := expandMacroProgram(t, input, object.NewEnvironment())
	testIntegerObject(t, Eval(expanded, object.NewEnvironment()), -2)
}

func TestExpandedModuloDiagnosticUsesArgumentSource(t *testing.T) {
	env := object.NewEnvironment()
	definition := "let identity = macro(x) { quote(unquote(x)); };"
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
	if got := failure.Diagnostic(); !reflect.DeepEqual(got, want) {
		t.Fatalf("expanded modulo diagnostic = %+v, want %+v", got, want)
	}
}
