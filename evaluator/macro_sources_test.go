package evaluator

import (
	"errors"
	"fmt"
	"monkey-go/ast"
	"monkey-go/diagnostic"
	"monkey-go/lexer"
	"monkey-go/object"
	"monkey-go/parser"
	"monkey-go/token"
	"reflect"
	"testing"
)

func TestMacroRuntimeErrorPreservesNestedCallSites(t *testing.T) {
	definition := "let broken = macro() {\n  let divide = fn() { 1 / 0; };\n  divide();\n};"
	env := object.NewEnvironment()
	DefineMacros(parseLocatedMacroProgram(t, "definitions.mk", definition), env)
	for i := 0; i < 2; i++ {
		expanded, err := ExpandMacros(parseLocatedMacroProgram(t, "calls.mk", "broken()"), env)
		if err != nil {
			t.Fatalf("template executed during expansion: %v", err)
		}
		failure := runtimeError(t, Eval(expanded, object.NewEnvironment()))
		report := failure.Diagnostic()
		if report.Kind != diagnostic.Runtime || report.Message != "division by zero" {
			t.Fatalf("incorrect template runtime error: %+v", report)
		}
		if sourcePoint(report.Span) != "definitions.mk:2:25" {
			t.Fatalf("template error lost its originating expression: %+v", report.Span)
		}
		checkMacroCallSites(t, report.Stack)
		report.Stack[0].Name = "changed"
		checkMacroCallSites(t, failure.Diagnostic().Stack)
	}
}

func checkMacroCallSites(t *testing.T, stack []diagnostic.Frame) {
	t.Helper()
	if len(stack) != 2 {
		t.Fatalf("expected a function and a macro frame, got %+v", stack)
	}
	want := []string{"divide at definitions.mk:3:3", "broken at calls.mk:1:1"}
	for i, frame := range stack {
		got := frame.Name + " at " + sourcePoint(frame.Span)
		if got != want[i] {
			t.Errorf("frame %d = %q, want %q", i, got, want[i])
		}
	}
}

func sourcePoint(span token.Span) string {
	if span.Source == nil {
		return "<unknown>"
	}
	return fmt.Sprintf("%s:%d:%d", span.Source.Name, span.Start.Line, span.Start.Column)
}

func TestMacroArityErrorPointsToCallWithoutEnteringDefinition(t *testing.T) {
	env := object.NewEnvironment()
	DefineMacros(parseLocatedMacroProgram(t, "definitions.mk", "let identity = macro(x) { x; };"), env)
	program := parseLocatedMacroProgram(t, "calls.mk", "identity()")
	_, err := ExpandMacros(program, env)
	report := requireMacroDiagnostic(t, err).Diagnostic()
	call := program.Statements[0].(*ast.ExpressionStatement).Expression
	want := diagnostic.Diagnostic{
		Kind: diagnostic.Macro, Span: call.Span(),
		Message: "macro identity expects 1 arguments, got 0",
	}
	if !reflect.DeepEqual(report, want) {
		t.Fatalf("incorrect call-boundary diagnostic: %+v", report)
	}
}

func TestTemplateAndArgumentSourcesSurviveRepeatedExpansion(t *testing.T) {
	env := object.NewEnvironment()
	definition := parseLocatedMacroProgram(t, "definitions.mk", "let add = macro(x) { x + 1; };")
	literal := definition.Statements[0].(*ast.LetStatement).Value.(*ast.MacroLiteral)
	addition := literal.Body.Statements[0].(*ast.ExpressionStatement).Expression.(*ast.InfixExpression)
	DefineMacros(definition, env)
	first := parseLocatedMacroProgram(t, "first.mk", "add(7)")
	call := first.Statements[0].(*ast.ExpressionStatement).Expression.(*ast.CallExpression)
	expanded, err := ExpandMacros(first, env)
	if err != nil {
		t.Fatal(err)
	}
	_, err = ExpandMacros(parseLocatedMacroProgram(t, "second.mk", "add(9)"), env)
	if err != nil {
		t.Fatal(err)
	}
	result := expanded.(*ast.Program).Statements[0].(*ast.ExpressionStatement).Expression.(*ast.MacroExpansion)
	checkTemplateOrigins(t, result, call, addition)
	testIntegerObject(t, Eval(expanded, object.NewEnvironment()), 8)
}

func checkTemplateOrigins(t *testing.T, result *ast.MacroExpansion, call *ast.CallExpression, addition *ast.InfixExpression) {
	t.Helper()
	generated := result.Body.Statements[0].(*ast.ExpressionStatement).Expression.(*ast.InfixExpression)
	if generated.Left.Span() != call.Arguments[0].Span() {
		t.Fatal("macro expansion replaced an argument's original source")
	}
	if generated.Token.Span != addition.Token.Span || generated.Right.Span() != addition.Right.Span() {
		t.Fatal("macro expansion replaced the template's original source")
	}
	if result.Call.Span() != call.Span() || addition.Left.String() != "x" {
		t.Fatal("macro expansion altered the call or reusable definition")
	}
}

func TestNestedTemplateRuntimeErrorKeepsExpansionFrames(t *testing.T) {
	definitions := "let inner = macro(x) { x / 0; };\nlet outer = macro(x) { inner(x); };"
	env := object.NewEnvironment()
	DefineMacros(parseLocatedMacroProgram(t, "definitions.mk", definitions), env)
	expanded, err := ExpandMacros(parseLocatedMacroProgram(t, "calls.mk", "outer(3)"), env)
	if err != nil {
		t.Fatal(err)
	}
	report := runtimeError(t, Eval(expanded, object.NewEnvironment())).Diagnostic()
	if len(report.Stack) != 2 {
		t.Fatalf("expected both macro frames: %+v", report.Stack)
	}
	if report.Stack[0].Name != "inner" || report.Stack[0].Span.Source.Name != "definitions.mk" {
		t.Errorf("incorrect inner template frame: %+v", report.Stack[0])
	}
	if report.Stack[1].Name != "outer" || report.Stack[1].Span.Source.Name != "calls.mk" {
		t.Errorf("incorrect outer template frame: %+v", report.Stack[1])
	}
}

func TestMacroExpansionDiagnosticStackIsNotShared(t *testing.T) {
	failure := newMacroExpansionError(parseLocatedMacroProgram(t, "call.mk", "1").Span(), "test")
	failure.report.Stack = []diagnostic.Frame{{Name: "original"}}
	report := failure.Diagnostic()
	report.Stack[0].Name = "changed"
	if failure.Diagnostic().Stack[0].Name != "original" {
		t.Fatal("macro diagnostic stack shared mutable storage")
	}
}

func parseLocatedMacroProgram(t *testing.T, name, source string) *ast.Program {
	t.Helper()
	p := parser.New(lexer.NewWithFilename(name, source))
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		t.Fatalf("parse %s: %v", name, p.Errors())
	}
	return program
}

func requireMacroDiagnostic(t *testing.T, err error) *MacroExpansionError {
	t.Helper()
	var failure *MacroExpansionError
	if !errors.As(err, &failure) {
		t.Fatalf("expected MacroExpansionError, got %T (%v)", err, err)
	}
	return failure
}
