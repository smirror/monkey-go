package parser

import (
	"monkey-go/ast"
	"monkey-go/lexer"
	"testing"
)

func TestMacroLiteralParsing(t *testing.T) {
	p := New(lexer.New(`let identity = macro(x) { quote(unquote(x)); };`))
	program := p.ParseProgram()
	if errs := p.Errors(); len(errs) != 0 {
		t.Fatalf("parser errors: %v", errs)
	}
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
	if len(literal.Parameters) != 1 || literal.Parameters[0].Value != "x" {
		t.Fatalf("unexpected macro parameters: %v", literal.Parameters)
	}
	if literal.Body == nil || len(literal.Body.Statements) != 1 {
		t.Fatalf("unexpected macro body: %+v", literal.Body)
	}
}
