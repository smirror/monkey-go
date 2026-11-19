package evaluator

import (
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

func TestQuoteAndUnquote(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"quote(5 + 5)", "(5 + 5)"},
		{"quote(8 + unquote(4 + 4))", "(8 + 8)"},
		{"quote(unquote(true))", "true"},
		{"quote(unquote(quote(1 + 2)))", "(1 + 2)"},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			p := parseMacroTest(t, tc.input)
			got := Eval(p.ParseProgram(), object.NewEnvironment())
			if errs := p.Errors(); len(errs) != 0 {
				t.Fatalf("parser errors: %v", errs)
			}
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
	p := parseMacroTest(t, input)
	program := p.ParseProgram()
	if errs := p.Errors(); len(errs) != 0 {
		t.Fatalf("parser errors: %v", errs)
	}
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
	p := parseMacroTest(t, input)
	program := p.ParseProgram()
	if errs := p.Errors(); len(errs) != 0 {
		t.Fatalf("parser errors: %v", errs)
	}
	env := object.NewEnvironment()
	DefineMacros(program, env)
	expanded, err := ExpandMacros(program, env)
	if err != nil {
		t.Fatal(err)
	}
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
		{`let recursive = macro(x) { quote(recursive(unquote(x))); }; recursive(1);`, "exceeded 32 passes"},
	}
	for _, tc := range cases {
		t.Run(tc.want, func(t *testing.T) {
			p := parseMacroTest(t, tc.source)
			program := p.ParseProgram()
			if errs := p.Errors(); len(errs) != 0 {
				t.Fatalf("parser errors: %v", errs)
			}
			env := object.NewEnvironment()
			DefineMacros(program, env)
			_, err := ExpandMacros(program, env)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want substring %q", err, tc.want)
			}
		})
	}
}
