package parser

import (
	"monkey-go/ast"
	"monkey-go/lexer"
	"testing"
)

func TestMacroLiteralParsing(t *testing.T) {
	literal := parseMacroDeclaration(t, `let identity = macro(x) { x; };`)
	if literal.TokenLiteral() != "macro" {
		t.Fatalf("macro token = %q, want macro", literal.TokenLiteral())
	}
	if len(literal.Parameters) != 1 {
		t.Fatalf("parameter count = %d, want 1", len(literal.Parameters))
	}
	testIdentifier(t, literal.Parameters[0], "x")
	if literal.Body == nil || len(literal.Body.Statements) != 1 {
		t.Fatalf("unexpected macro body: %+v", literal.Body)
	}
	if got := literal.Body.String(); got != "x" {
		t.Fatalf("macro body = %q, want x", got)
	}
}

func TestMacroTemplateBodies(t *testing.T) {
	tests := []struct {
		input      string
		statements []string
	}{
		{
			`let remember = macro(x) { let saved = x; saved + 1; };`,
			[]string{"let saved = x;", "saved + 1;"},
		},
		{
			`let choose = macro(condition, yes, no) { if (condition) { yes; } else { no; }; };`,
			[]string{"if (condition) { yes; } else { no; };"},
		},
		{
			`let accumulate = macro(values) {
    let positive = fn(value) { if (value > 0) { value; } else { 0; }; };
    let total = 0;
    for value in values { total = total + positive(value); }
    return total;
};`,
			[]string{
				"let positive = fn(value) { if (value > 0) { value; } else { 0; }; };",
				"let total = 0;",
				"for value in values { total = total + positive(value); }",
				"return total;",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			literal := parseMacroDeclaration(t, tt.input)
			if len(literal.Body.Statements) != len(tt.statements) {
				t.Fatalf("body has %d statements, want %d", len(literal.Body.Statements), len(tt.statements))
			}
			for i, statement := range literal.Body.Statements {
				checkSpanText(t, statement.Span(), tt.statements[i])
			}
		})
	}
}

func parseMacroDeclaration(t *testing.T, input string) *ast.MacroLiteral {
	t.Helper()
	p := New(lexer.New(input))
	program := p.ParseProgram()
	checkParserErrors(t, p)
	if len(program.Statements) != 1 {
		t.Fatalf("statement count = %d, want 1", len(program.Statements))
	}
	declaration, ok := program.Statements[0].(*ast.LetStatement)
	if !ok {
		t.Fatalf("expected let statement, got %T", program.Statements[0])
	}
	literal, ok := declaration.Value.(*ast.MacroLiteral)
	if !ok {
		t.Fatalf("expected macro literal, got %T", declaration.Value)
	}
	return literal
}

func TestMacroParameterParsing(t *testing.T) {
	tests := []struct {
		input      string
		parameters []string
	}{
		{`let identity = macro() {};`, nil},
		{`let identity = macro(x, y, z) {};`, []string{"x", "y", "z"}},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			literal := parseMacroDeclaration(t, tt.input)
			if len(literal.Parameters) != len(tt.parameters) {
				t.Fatalf("parameter count = %d, want %d", len(literal.Parameters), len(tt.parameters))
			}
			for i, parameter := range tt.parameters {
				testIdentifier(t, literal.Parameters[i], parameter)
			}
			if len(literal.Body.Statements) != 0 {
				t.Fatalf("expected an empty macro body, got %s", literal.Body.String())
			}
		})
	}
}

func TestCallableLiteralSyntaxErrors(t *testing.T) {
	tests := []struct {
		suffix string
		want   string
	}{
		{" x { x }", "expected next token to be (, got IDENT instead."},
		{"(x { x }", "expected next token to be ), got { instead."},
		{"(x) x", "expected next token to be {, got IDENT instead."},
	}
	for _, keyword := range []string{"fn", "macro"} {
		for _, tt := range tests {
			t.Run(keyword+tt.suffix, func(t *testing.T) {
				p := New(lexer.New(keyword + tt.suffix))
				p.ParseProgram()
				if len(p.Errors()) == 0 {
					t.Fatal("expected a parser error")
				}
				if got := p.Errors()[0]; got != tt.want {
					t.Fatalf("first error = %q, want %q", got, tt.want)
				}
			})
		}
	}
}

func TestCallArgumentLists(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"f();", "f()"},
		{"f(g(1), [2, 3]);", "f(g(1), [2, 3])"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			p := New(lexer.New(tt.input))
			program := p.ParseProgram()
			checkParserErrors(t, p)
			if got := program.String(); got != tt.want {
				t.Fatalf("call AST = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCallArgumentListRequiresClosingParenthesis(t *testing.T) {
	p := New(lexer.New("f(1;"))
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Fatal("expected a parser error")
	}
	want := "expected next token to be ), got ; instead."
	if got := p.Errors()[0]; got != want {
		t.Fatalf("first error = %q, want %q", got, want)
	}
}
