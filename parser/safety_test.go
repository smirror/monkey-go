package parser

import (
	"monkey-go/ast"
	"monkey-go/lexer"
	"testing"
)

func TestStatementsWithoutSemicolonsKeepFollowingStatements(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"let and expression", "let a = 1\nlet b = 2\na + b", 3},
		{"return at EOF", "return 42", 1},
		{"return before next statement", "return 42\n10", 2},
		{"return in block", "let f = fn(x) { return x }\nf(3)", 2},
		{"let inside block", "if (true) { let x = 1\nlet y = 2\nx + y }", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New(lexer.New(tt.input))
			program := p.ParseProgram()
			if errs := p.Errors(); len(errs) > 0 {
				t.Fatalf("unexpected parser errors: %v", errs)
			}
			if got := len(program.Statements); got != tt.want {
				t.Fatalf("got %d statements, want %d", got, tt.want)
			}
		})
	}
}

func TestParserInvalidInputDoesNotAppendNilStatements(t *testing.T) {
	inputs := []string{
		"let = 1;",
		"let x = ;",
		"return ;",
		"1 + ;",
		"for x in range(3) {",
		"if (true) { 10",
		"fn(x) { return x",
		"(",
	}
	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			p := New(lexer.New(input))
			program := p.ParseProgram()
			if len(p.Errors()) == 0 {
				t.Fatalf("expected parser errors for %q", input)
			}
			for _, stmt := range program.Statements {
				if stmt == nil {
					t.Fatal("nil statement in parsed program")
				}
				if e, ok := stmt.(*ast.ExpressionStatement); ok && e.Expression == nil {
					t.Fatal("expression statement has nil expression")
				}
			}
		})
	}
}

func TestParserUnterminatedBlockReportsError(t *testing.T) {
	tests := []string{
		"if (true) { 1",
		"let f = fn() { return 1",
		"for i in range(3) { print(i);",
	}
	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			p := New(lexer.New(input))
			_ = p.ParseProgram()
			found := false
			for _, e := range p.Errors() {
				if e == "unterminated block: expected } before EOF" {
					found = true
				}
			}
			if !found {
				t.Fatalf("expected unterminated block error, got %v", p.Errors())
			}
		})
	}
}

func TestParserPreservesReturnStatementsInsideBlocks(t *testing.T) {
	p := New(lexer.New("let f = fn(x) { if (x > 0) { return x } return 0 };"))
	program := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("unexpected parser errors: %v", errs)
	}
	letStmt, ok := program.Statements[0].(*ast.LetStatement)
	if !ok {
		t.Fatalf("expected let statement, got %T", program.Statements[0])
	}
	fn, ok := letStmt.Value.(*ast.FunctionLiteral)
	if !ok {
		t.Fatalf("expected function literal, got %T", letStmt.Value)
	}
	if got := len(fn.Body.Statements); got != 2 {
		t.Fatalf("got %d statements in body, want 2", got)
	}
}
