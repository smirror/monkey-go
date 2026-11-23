package parser

import (
	"monkey-go/ast"
	"monkey-go/lexer"
	"strings"
	"testing"
)

func TestWhileParsing(t *testing.T) {
	for _, tt := range []struct{ input, condition string }{
		{"while true { break; }", "true"},
		{"while (x < 5) { continue; }", "(x < 5)"},
		{"while x < 5 && ready() { x = x + 1; }", "((x < 5) && ready())"},
		{"while false {}", "false"},
	} {
		t.Run(tt.input, func(t *testing.T) {
			program := parseForTest(t, tt.input)
			if len(program.Statements) != 1 {
				t.Fatalf("expected one statement, got %d", len(program.Statements))
			}
			statement, ok := program.Statements[0].(*ast.WhileStatement)
			if !ok {
				t.Fatalf("got %T, want *ast.WhileStatement", program.Statements[0])
			}
			if statement.TokenLiteral() != "while" || statement.Condition.String() != tt.condition {
				t.Fatalf("incorrect while AST: %+v", statement)
			}
			if statement.Body == nil {
				t.Fatal("while AST has no body")
			}
		})
	}
}

func TestWhileSpansCoverConditionAndBody(t *testing.T) {
	source := "\"日本\";\nwhile n < 2 {\n    continue;\n}"
	program := parseLocatedProgram(t, source)
	statement := program.Statements[1].(*ast.WhileStatement)
	for _, tt := range []struct {
		node ast.Node
		want string
	}{
		{statement, "while n < 2 {\n    continue;\n}"},
		{statement.Condition, "n < 2"},
		{statement.Body, "{\n    continue;\n}"},
	} {
		span := tt.node.Span()
		checkSpanText(t, span, tt.want)
		if span.Source != program.Span().Source || span.Start.Line != 2 {
			t.Fatalf("incorrect while node location: %+v", span)
		}
	}
	checkSpanText(t, statement.Token.Span, "while")
	if statement.Token.Span.Start.Column != 1 || statement.Token.Span.End.Column != 6 {
		t.Fatalf("while token does not identify its keyword: %+v", statement.Token.Span)
	}
}

func TestWhileLoopContextAndFunctionBoundaries(t *testing.T) {
	valid := []string{
		"while true { for i in [] { break; } continue; }",
		"for i in [] { while true { continue; } break; }",
		"while true { while false {} break; }",
		"while true { let f = fn() { while true { break; } }; continue; }",
		"while true { while if (true) { break; } {} }",
	}
	for _, input := range valid {
		parseForTest(t, input)
	}
	for _, keyword := range []string{"break", "continue"} {
		for _, template := range []string{
			"while false {} CONTROL;",
			"while true { let f = fn() { CONTROL; }; }",
			"while if (true) { CONTROL; } {}",
		} {
			checkOutsideLoopDiagnostic(t, strings.ReplaceAll(template, "CONTROL", keyword), keyword)
		}
	}
}

func TestWhileRejectsMissingConditionOrBlock(t *testing.T) {
	for _, input := range []string{
		"while", "while () {}", "while {}", "while true", "while true 42;",
		"while true: {}", "while true {", "while (true {}", "while true {} else {}",
		"fn(while) {}", "let while = 1;",
	} {
		t.Run(input, func(t *testing.T) {
			p := New(lexer.New(input))
			_ = p.ParseProgram()
			if len(p.Diagnostics()) == 0 {
				t.Fatalf("invalid while syntax was accepted: %q", input)
			}
		})
	}
}
