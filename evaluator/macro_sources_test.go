package evaluator

import (
	"errors"
	"monkey-go/ast"
	"monkey-go/diagnostic"
	"monkey-go/lexer"
	"monkey-go/object"
	"monkey-go/parser"
	"monkey-go/token"
	"strings"
	"testing"
)

func TestMacroEvaluationErrorPreservesNestedCallSites(t *testing.T) {
	definition := "let broken = macro() {\n  let divide = fn() { 1 / 0; };\n  divide();\n};"
	env := object.NewEnvironment()
	DefineMacros(parseLocatedMacroProgram(t, "definitions.mk", definition), env)
	for i := 0; i < 2; i++ {
		_, err := ExpandMacros(parseLocatedMacroProgram(t, "calls.mk", "broken()"), env)
		failure := requireMacroDiagnostic(t, err)
		report := failure.Diagnostic()
		if report.Kind != diagnostic.Macro || failure.Error() != "macro broken: division by zero" {
			t.Fatalf("incorrect macro error: %+v", report)
		}
		if report.Span.Source.Name != "definitions.mk" || report.Span.Start.Line != 2 || report.Span.Start.Column != 25 {
			t.Fatalf("macro error lost its originating expression: %+v", report.Span)
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
	if stack[0].Name != "divide" || stack[0].Span.Source.Name != "definitions.mk" || stack[0].Span.Start.Line != 3 {
		t.Errorf("incorrect inner frame: %+v", stack[0])
	}
	if stack[1].Name != "broken" || stack[1].Span.Source.Name != "calls.mk" || stack[1].Span.Start.Column != 1 {
		t.Errorf("incorrect macro call frame: %+v", stack[1])
	}
}

func TestMacroArityErrorPointsToCallWithoutEnteringDefinition(t *testing.T) {
	env := object.NewEnvironment()
	DefineMacros(parseLocatedMacroProgram(t, "definitions.mk", "let identity = macro(x) { quote(unquote(x)); };"), env)
	program := parseLocatedMacroProgram(t, "calls.mk", "identity()")
	_, err := ExpandMacros(program, env)
	report := requireMacroDiagnostic(t, err).Diagnostic()
	call := program.Statements[0].(*ast.ExpressionStatement).Expression
	if report.Kind != diagnostic.Macro || report.Span != call.Span() || len(report.Stack) != 0 {
		t.Fatalf("incorrect call-boundary diagnostic: %+v", report)
	}
}

func TestComputedUnquoteValuesRetainOrigin(t *testing.T) {
	for _, source := range []string{"quote(unquote(2 + 3))", "quote(unquote(true))", "quote(unquote(\"日本\"))"} {
		t.Run(source, func(t *testing.T) {
			program := parseLocatedMacroProgram(t, "unquote.mk", source)
			quoteCall := program.Statements[0].(*ast.ExpressionStatement).Expression.(*ast.CallExpression)
			origin := quoteCall.Arguments[0].Span()
			result := Eval(program, object.NewEnvironment()).(*object.Quote)
			if result.Node.Span() != origin || generatedTokenSpan(t, result.Node) != origin {
				t.Fatalf("generated literal lost its unquote origin: %+v", result.Node.Span())
			}
		})
	}
}

func generatedTokenSpan(t *testing.T, node ast.Node) token.Span {
	t.Helper()
	switch literal := node.(type) {
	case *ast.IntegerLiteral:
		return literal.Token.Span
	case *ast.Boolean:
		return literal.Token.Span
	case *ast.StringLiteral:
		return literal.Token.Span
	default:
		t.Fatalf("unexpected generated literal %T", node)
		return token.Span{}
	}
}

func TestUnquoteErrorsKeepTheMostSpecificLocation(t *testing.T) {
	tests := []struct{ source, marked string }{
		{"quote(unquote())", "unquote"},
		{"quote(unquote([]))", "unquote"},
		{"quote(unquote(1 / 0))", "/"},
	}
	for _, tt := range tests {
		t.Run(tt.source, func(t *testing.T) {
			value := Eval(parseLocatedMacroProgram(t, "unquote.mk", tt.source), object.NewEnvironment())
			err := runtimeError(t, value)
			if err.Span.Source.Name != "unquote.mk" || err.Span.Start.Offset != strings.Index(tt.source, tt.marked) {
				t.Fatalf("incorrect unquote failure location: %+v", err.Span)
			}
		})
	}
}

func TestUnquotedArgumentsKeepTheirOwnSourceAcrossExpansions(t *testing.T) {
	env := object.NewEnvironment()
	DefineMacros(parseLocatedMacroProgram(t, "definitions.mk", "let identity = macro(x) { quote(unquote(x)); };"), env)
	first := parseLocatedMacroProgram(t, "first.mk", "identity(7)")
	call := first.Statements[0].(*ast.ExpressionStatement).Expression.(*ast.CallExpression)
	origin := call.Arguments[0].Span()
	expanded, err := ExpandMacros(first, env)
	if err != nil {
		t.Fatal(err)
	}
	_, err = ExpandMacros(parseLocatedMacroProgram(t, "second.mk", "identity(9)"), env)
	if err != nil {
		t.Fatal(err)
	}
	value := expanded.(*ast.Program).Statements[0].(*ast.ExpressionStatement).Expression
	if value.Span() != origin || call.Arguments[0].Span() != origin {
		t.Fatal("macro expansion replaced an argument's original source")
	}
	testIntegerObject(t, Eval(expanded, object.NewEnvironment()), 7)
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
