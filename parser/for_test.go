package parser

import (
	"monkey-go/ast"
	"monkey-go/lexer"
	"strings"
	"testing"
)

func parseForTest(t *testing.T, input string) *ast.Program {
	t.Helper()
	p := New(lexer.New(input))
	program := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("parser errors: %v", errs)
	}
	return program
}

func TestForInParsing(t *testing.T) {
	tests := []struct {
		input       string
		variable    string
		iterableAST string
	}{
		{"for (i in range(2, 10)) { print(i); }", "i", "range(2, 10)"},
		{"for (i in range(10)) { print(i) }", "i", "range(10)"},
		{"for (item in [10, 20, 30]) { print(item) }", "item", "[10, 20, 30]"},
		{"for (item in values) { print(item) }", "item", "values"},
		{"for (i in getItems(2)) { i }", "i", "getItems(2)"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			program := parseForTest(t, tt.input)
			if len(program.Statements) != 1 {
				t.Fatalf("got %d statements, want 1", len(program.Statements))
			}
			stmt, ok := program.Statements[0].(*ast.ForInStatement)
			if !ok {
				t.Fatalf("got %T, want *ast.ForInStatement", program.Statements[0])
			}
			if stmt.Variable.Value != tt.variable || stmt.Iterable.String() != tt.iterableAST {
				t.Fatalf("variable=%q iterable=%q", stmt.Variable.Value, stmt.Iterable.String())
			}
			if len(stmt.Body.Statements) != 1 {
				t.Fatalf("body statement count = %d", len(stmt.Body.Statements))
			}
		})
	}
}

func TestClassicForParsing(t *testing.T) {
	tests := []struct {
		input    string
		initAST  string
		condAST  string
		postAST  string
	}{
		{"for (i = 0; i < 10; i++) { print(i) }", "(i = 0)", "(i < 10)", "(i++)"},
		{"for (let i = 0; i < 10; i = i + 1) { print(i); }", "let i = 0", "(i < 10)", "(i = (i + 1))"},
		{"for (; i < 10; i++) { i }", "", "(i < 10)", "(i++)"},
		{"for (i = 0; i < 10;) { i }", "(i = 0)", "(i < 10)", ""},
		{"for (i = 0; i < 10 && active; i++) { i }", "(i = 0)", "((i < 10) && active)", "(i++)"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			program := parseForTest(t, tt.input)
			if len(program.Statements) != 1 {
				t.Fatalf("got %d statements, want 1", len(program.Statements))
			}
			stmt, ok := program.Statements[0].(*ast.ForStatement)
			if !ok {
				t.Fatalf("got %T, want *ast.ForStatement", program.Statements[0])
			}
			init := ""
			if stmt.Init != nil {
				init = strings.TrimSuffix(stmt.Init.String(), ";")
			}
			post := ""
			if stmt.Post != nil {
				post = stmt.Post.String()
			}
			if init != tt.initAST || stmt.Condition.String() != tt.condAST || post != tt.postAST {
				t.Fatalf("init=%q cond=%q post=%q", init, stmt.Condition.String(), post)
			}
			if len(stmt.Body.Statements) != 1 {
				t.Fatalf("body statements=%d, want 1", len(stmt.Body.Statements))
			}
		})
	}
}

func TestNestedForParsing(t *testing.T) {
	program := parseForTest(t, "for (i in range(2)) { for (j = 0; j < 3; j++) { print(i + j) } }")
	outer, ok := program.Statements[0].(*ast.ForInStatement)
	if !ok {
		t.Fatalf("outer statement is %T", program.Statements[0])
	}
	if _, ok := outer.Body.Statements[0].(*ast.ForStatement); !ok {
		t.Fatalf("inner statement is %T", outer.Body.Statements[0])
	}
}

func TestInvalidForSyntax(t *testing.T) {
	tests := []string{
		"for (i = 0, i < 10, i++) { i }",
		"for (i = 0; ; i++) { i }",
		"for (i in ) { i }",
		"for (i in range(3)) print(i)",
		"for (i = 0; i < 3; i++ { i }",
	}
	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			p := New(lexer.New(input))
			_ = p.ParseProgram()
			if len(p.Errors()) == 0 {
				t.Fatalf("expected parser error for %q", input)
			}
		})
	}
}
