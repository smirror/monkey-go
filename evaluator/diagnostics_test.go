package evaluator

import (
	"monkey-go/ast"
	"monkey-go/diagnostic"
	"monkey-go/lexer"
	"monkey-go/object"
	"monkey-go/parser"
	"monkey-go/token"
	"testing"
)

func evalNamed(t *testing.T, name, source string, env *object.Environment) object.Object {
	t.Helper()
	p := parser.New(lexer.NewWithFilename(name, source))
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		t.Fatalf("parse %s: %v", name, p.Errors())
	}
	return Eval(program, env)
}

func runtimeError(t *testing.T, value object.Object) *object.Error {
	t.Helper()
	err, ok := value.(*object.Error)
	if !ok {
		t.Fatalf("expected a runtime error, got %T (%v)", value, value)
	}
	return err
}

func TestRuntimeDiagnosticPreservesDefinitionAndCallSites(t *testing.T) {
	env := object.NewEnvironment()
	definition := "let divide = fn(x) {\n    10 / x\n};"
	evalNamed(t, "divide.mk", definition, env)
	evalNamed(t, "wrapper.mk", "let wrapper = fn(x) {\n    divide(x)\n};", env)
	for i := 0; i < 2; i++ {
		err := runtimeError(t, evalNamed(t, "call.mk", "wrapper(0)", env))
		if err.Kind != diagnostic.Runtime || err.Message != "division by zero" {
			t.Fatalf("unexpected diagnostic: %+v", err)
		}
		if err.Span.Source.Name != "divide.mk" || err.Span.Source.Text != definition {
			t.Fatalf("definition source was lost: %+v", err.Span)
		}
		if err.Span.Start.Line != 2 || err.Span.Start.Column != 8 {
			t.Fatalf("expected division operator at 2:8, got %+v", err.Span)
		}
		checkDiagnosticFrames(t, err)
	}
}

func checkDiagnosticFrames(t *testing.T, err *object.Error) {
	t.Helper()
	type callSite struct {
		name, file   string
		line, column int
	}
	expected := []callSite{{"divide", "wrapper.mk", 2, 5}, {"wrapper", "call.mk", 1, 1}}
	if len(err.Stack) != len(expected) {
		t.Fatalf("want exactly two frames, got %+v", err.Stack)
	}
	for i, want := range expected {
		frame := err.Stack[i]
		if frame.Span.Source == nil {
			t.Fatalf("frame %d lost its source", i)
		}
		got := callSite{frame.Name, frame.Span.Source.Name, frame.Span.Start.Line, frame.Span.Start.Column}
		if got != want {
			t.Errorf("call site %d = %+v, want %+v", i, got, want)
		}
	}
}

func TestRuntimeStackOnlyIncludesEnteredFunctions(t *testing.T) {
	tests := []struct {
		name, source string
		frames       []string
	}{
		{"argument expression", "let f = fn(x) { x }; f(1 / 0)", nil},
		{"argument call", "let f = fn(x) { x }; let fail = fn() { 1 / 0 }; f(fail())", []string{"fail"}},
		{"missing callee", "missing(1)", nil},
		{"non-function callee", "1(2)", nil},
		{"wrong arity", "let f = fn(x) { x }; f()", nil},
		{"builtin error", "len(1)", nil},
		{"anonymous function", "fn() { 1 / 0 }()", []string{"<anonymous>"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := runtimeError(t, evalNamed(t, "case.mk", tt.source, object.NewEnvironment()))
			if len(err.Stack) != len(tt.frames) {
				t.Fatalf("stack = %+v, want %v", err.Stack, tt.frames)
			}
			for i, name := range tt.frames {
				if err.Stack[i].Name != name {
					t.Errorf("frame %d = %q, want %q", i, err.Stack[i].Name, name)
				}
			}
			if err.Span.Source.Name != "case.mk" {
				t.Fatalf("error has no source: %+v", err.Span)
			}
		})
	}
}

func TestRuntimeDiagnosticUnicodeOperator(t *testing.T) {
	err := runtimeError(t, evalNamed(t, "unicode.mk", "\"日本\" + 1", object.NewEnvironment()))
	if err.Span.Start.Offset != 9 || err.Span.Start.Column != 6 {
		t.Fatalf("byte offset and rune column were confused: %+v", err.Span)
	}
}

func TestRuntimeDiagnosticsDoNotMutateHostErrors(t *testing.T) {
	shared := &object.Error{Message: "host failure"}
	env := object.NewEnvironment()
	env.Set("host", &object.Builtin{Fn: func(args ...object.Object) object.Object { return shared }})
	evalNamed(t, "def.mk", "let f = fn() { host() };", env)
	for i := 0; i < 2; i++ {
		err := runtimeError(t, evalNamed(t, "call.mk", "f()", env))
		if err.Kind != diagnostic.Runtime || len(err.Stack) != 1 {
			t.Fatalf("incorrect host error annotation: %+v", err)
		}
	}
	if shared.Span.Source != nil || len(shared.Stack) != 0 || shared.Kind != "" {
		t.Fatalf("shared host error was mutated: %+v", shared)
	}
}

func TestRuntimeErrorWithoutSourceKeepsLegacyAPI(t *testing.T) {
	value := Eval(&ast.Identifier{Value: "missing"}, object.NewEnvironment())
	err := runtimeError(t, value)
	if err.Type() != object.ERROR_OBJ || err.Inspect() != "ERROR: identifier not found: missing" {
		t.Fatalf("legacy error API changed: %+v", err)
	}
	if got := diagnostic.Format(err.Diagnostic()); got != "runtime error: identifier not found: missing\n" {
		t.Fatalf("unexpected source-less diagnostic: %q", got)
	}
}

func TestHostErrorLocationSurvivesKindAnnotation(t *testing.T) {
	span := token.Span{
		Source: &token.Source{Name: "host.mk", Text: "failure"},
		Start:  token.Position{Line: 1, Column: 1},
		End:    token.Position{Offset: 7, Line: 1, Column: 8},
	}
	shared := &object.Error{Message: "host failure", Span: span}
	env := object.NewEnvironment()
	env.Set("host", &object.Builtin{Fn: func(args ...object.Object) object.Object { return shared }})
	err := runtimeError(t, evalNamed(t, "caller.mk", "host()", env))
	if err.Kind != diagnostic.Runtime || err.Span != span {
		t.Fatalf("host position was changed or kind missing: %+v", err)
	}
	if shared.Kind != "" {
		t.Fatalf("host-owned error was changed: %+v", shared)
	}
}
