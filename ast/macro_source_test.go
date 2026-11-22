package ast_test

import (
	"monkey-go/ast"
	"monkey-go/lexer"
	"monkey-go/parser"
	"testing"
)

func TestMacroSourceLocationsSurviveASTRewriting(t *testing.T) {
	source := "macro(x) {\n  quote(1 + unquote(x));\n}"
	p := parser.New(lexer.NewWithFilename("macros.mk", source))
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		t.Fatalf("parse macro: %v", p.Errors())
	}
	macro := program.Statements[0].(*ast.ExpressionStatement).Expression.(*ast.MacroLiteral)
	span := macro.Span()
	if span.Source == nil || span.Source.Name != "macros.mk" || span.Source.Text != source {
		t.Fatalf("macro lost its source: %+v", span)
	}
	if span.Start.Offset != 0 || span.End.Offset != len(source) {
		t.Fatalf("macro range does not cover its definition: %+v", span)
	}
	rewritten := ast.Modify(macro, replaceOneWith(2)).(*ast.MacroLiteral)
	if rewritten == macro || rewritten.Body == macro.Body {
		t.Fatal("rewriting reused a mutable macro container")
	}
	if rewritten.Span() != span || rewritten.Body.Span() != macro.Body.Span() {
		t.Fatal("rewriting discarded a macro or body source location")
	}
	assertNodeString(t, macro, "macro(x) {quote((1 + unquote(x)))}")
	assertNodeString(t, rewritten, "macro(x) {quote((2 + unquote(x)))}")
}
