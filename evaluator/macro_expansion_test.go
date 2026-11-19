package evaluator

import (
	"monkey-go/ast"
	"monkey-go/lexer"
	"monkey-go/object"
	"monkey-go/parser"
	"strings"
	"testing"
)

func parseMacroTest(t *testing.T, source string) *parser.Parser {
	t.Helper()
	p := parser.New(lexer.New(source))
	return p
}

func parseMacroProgram(t *testing.T, source string) *ast.Program {
	t.Helper()
	p := parseMacroTest(t, source)
	program := p.ParseProgram()
	if errs := p.Errors(); len(errs) != 0 {
		t.Fatalf("parser errors: %v", errs)
	}
	return program
}

func expandMacroProgram(t *testing.T, source string, env *object.Environment) ast.Node {
	t.Helper()
	program := parseMacroProgram(t, source)
	DefineMacros(program, env)
	expanded, err := ExpandMacros(program, env)
	if err != nil {
		t.Fatal(err)
	}
	return expanded
}

func TestQuoteAndUnquote(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"quote(5 + 5)", "(5 + 5)"},
		{"quote(8 + unquote(4 + 4))", "(8 + 8)"},
		{"quote(unquote(true))", "true"},
		{"quote(unquote(false))", "false"},
		{`quote(unquote("monkey"))`, "monkey"},
		{"quote(unquote(-7))", "-7"},
		{"quote(unquote(quote(1 + 2)))", "(1 + 2)"},
		{"let value = 9; quote(unquote(value));", "9"},
		{"quote(missing(1 + 2))", "missing((1 + 2))"},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got := Eval(parseMacroProgram(t, tc.input), object.NewEnvironment())
			quoteObj, ok := got.(*object.Quote)
			if !ok {
				t.Fatalf("expected Quote, got %T (%v)", got, got)
			}
			if got := quoteObj.Node.String(); got != tc.want {
				t.Fatalf("quote node = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestMacroExpansionAndDefinitionReuse(t *testing.T) {
	input := `let twice = macro(x) { quote(unquote(x) + unquote(x)); };
twice(1);
twice(7);`
	program := parseMacroProgram(t, input)
	macroEnv := object.NewEnvironment()
	DefineMacros(program, macroEnv)
	if len(program.Statements) != 2 {
		t.Fatalf("runtime statement count = %d, want 2", len(program.Statements))
	}
	expanded, err := ExpandMacros(program, macroEnv)
	if err != nil {
		t.Fatal(err)
	}
	got := Eval(expanded, object.NewEnvironment())
	integer, ok := got.(*object.Integer)
	if !ok || integer.Value != 14 {
		t.Fatalf("result = %v (%T), want integer 14", got, got)
	}
}

func TestNestedMacroExpansion(t *testing.T) {
	input := `let inc = macro(x) { quote(unquote(x) + 1); };
let twice = macro(x) { quote(unquote(x) + unquote(x)); };
twice(inc(3));`
	expanded := expandMacroProgram(t, input, object.NewEnvironment())
	got := Eval(expanded, object.NewEnvironment())
	value, ok := got.(*object.Integer)
	if !ok || value.Value != 8 {
		t.Fatalf("result = %v (%T), want integer 8", got, got)
	}
}

func TestMacroErrors(t *testing.T) {
	cases := []struct {
		source string
		want   string
	}{
		{`let twice = macro(x) { quote(unquote(x) + unquote(x)); }; twice();`, "expects 1 arguments"},
		{`let invalid = macro() { 10; }; invalid();`, "must return quote"},
		{`let empty = macro() {}; empty();`, "must return quote"},
		{`let invalid = macro() { missing; }; invalid();`, "macro invalid: identifier not found: missing"},
		{`let invalid = macro() { quote(unquote([])); }; invalid();`, "macro invalid: cannot unquote ARRAY as an expression"},
		{`let recursive = macro(x) { quote(recursive(unquote(x))); }; recursive(1);`, "exceeded 32 passes"},
	}
	for _, tc := range cases {
		t.Run(tc.want, func(t *testing.T) {
			program := parseMacroProgram(t, tc.source)
			env := object.NewEnvironment()
			DefineMacros(program, env)
			_, err := ExpandMacros(program, env)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want substring %q", err, tc.want)
			}
		})
	}
}

func TestGeneratedMacroCallsAndExplicitReturn(t *testing.T) {
	input := `let inc = macro(x) { return quote(unquote(x) + 1); };
let generated = macro(x) { quote(inc(unquote(x))); };
generated(4);`
	expanded := expandMacroProgram(t, input, object.NewEnvironment())
	testIntegerObject(t, Eval(expanded, object.NewEnvironment()), 5)
}

func TestMacroArgumentsAreSyntax(t *testing.T) {
	input := `let ignore = macro(x) { quote(42); }; ignore(missing);`
	expanded := expandMacroProgram(t, input, object.NewEnvironment())
	testIntegerObject(t, Eval(expanded, object.NewEnvironment()), 42)
}

func TestMacroEnvironmentRemainsSeparateFromRuntime(t *testing.T) {
	env := object.NewEnvironment()
	env.Set("offset", &object.Integer{Value: 3})
	input := `let add = macro(x) { quote(unquote(x) + unquote(offset)); };
let offset = 100;
add(offset);`
	expanded := expandMacroProgram(t, input, env)
	testIntegerObject(t, Eval(expanded, object.NewEnvironment()), 103)
}

func TestExpansionPreservesRuntimeCalls(t *testing.T) {
	cases := []struct {
		input string
		want  int64
	}{
		{`len([1, 2]);`, 2},
		{`fn(x) { x + 1; }(2);`, 3},
		{`let identity = fn(x) { x; }; identity(7);`, 7},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			env := object.NewEnvironment()
			env.Set("identity", &object.Integer{Value: 99})
			expanded := expandMacroProgram(t, tc.input, env)
			if expanded.String() != parseMacroProgram(t, tc.input).String() {
				t.Fatal("expansion changed a runtime call")
			}
			testIntegerObject(t, Eval(expanded, object.NewEnvironment()), tc.want)
		})
	}
}

func TestMacroRejectsNonExpressionQuote(t *testing.T) {
	env := object.NewEnvironment()
	env.Set("statements", &object.Quote{Node: &ast.Program{}})
	program := parseMacroProgram(t, `let invalid = macro() { statements; }; invalid();`)
	DefineMacros(program, env)
	_, err := ExpandMacros(program, env)
	if err == nil || err.Error() != "macro invalid returned a non-expression AST node" {
		t.Fatalf("unexpected expansion error: %v", err)
	}
}

func TestExpansionStopsEvaluatingMacrosAfterFirstError(t *testing.T) {
	calls := 0
	env := object.NewEnvironment()
	env.Set("tick", &object.Builtin{Fn: func(args ...object.Object) object.Object {
		calls++
		return &object.Integer{Value: 1}
	}})
	program := parseMacroProgram(t, `let invalid = macro() { missing; };
let later = macro() { quote(unquote(tick())); };
invalid(); later();`)
	DefineMacros(program, env)
	_, err := ExpandMacros(program, env)
	if err == nil || err.Error() != "macro invalid: identifier not found: missing" {
		t.Fatalf("unexpected expansion error: %v", err)
	}
	if calls != 0 {
		t.Fatalf("evaluated %d later macros after expansion failed", calls)
	}
}
