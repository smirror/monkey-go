package ast_test

import (
	"monkey-go/ast"
	"monkey-go/lexer"
	"monkey-go/parser"
	"monkey-go/token"
	"testing"
)

func TestMacroSourceLocationsSurviveASTRewriting(t *testing.T) {
	source := "macro(x) {\n  1 + x;\n}"
	macro := parseMacroDefinition(t, "macros.mk", source)
	span := macro.Span()
	assertMacroSourceSpan(t, span, source)
	rewritten := ast.Modify(macro, replaceOneWith(2)).(*ast.MacroLiteral)
	if rewritten == macro || rewritten.Body == macro.Body {
		t.Fatal("rewriting reused a mutable macro container")
	}
	if rewritten.Span() != span || rewritten.Body.Span() != macro.Body.Span() {
		t.Fatal("rewriting discarded a macro or body source location")
	}
	assertNodeString(t, macro, "macro(x) {(1 + x)}")
	assertNodeString(t, rewritten, "macro(x) {(2 + x)}")
}

func TestMacroExpansionPreservesSourceAndDiagnosticCall(t *testing.T) {
	macro := parseMacroDefinition(t, "macros.mk", "macro(x) { 1 + x; }")
	callProgram := parseModifyProgram(t, "template(1)")
	call := callProgram.Statements[0].(*ast.ExpressionStatement).Expression.(*ast.CallExpression)
	expansion := &ast.MacroExpansion{Body: macro.Body, Call: call}
	ast.SetSpan(expansion, call.Span())
	rewritten := ast.Modify(expansion, replaceOneWith(2)).(*ast.MacroExpansion)
	if rewritten.Span() != call.Span() || rewritten.Body.Span() != macro.Body.Span() {
		t.Fatal("expanded call or template body location was discarded")
	}
	assertNodeString(t, rewritten, "(2 + x)")
	assertNodeString(t, rewritten.Call, "template(1)")
	if rewritten.Call == call {
		t.Fatal("rewriting reused mutable diagnostic metadata")
	}
	rewritten.Call.Arguments[0].(*ast.IntegerLiteral).Value = 7
	if call.Arguments[0].(*ast.IntegerLiteral).Value != 1 {
		t.Fatal("rewriting shared diagnostic argument storage")
	}
}

func parseMacroDefinition(t *testing.T, filename, source string) *ast.MacroLiteral {
	t.Helper()
	p := parser.New(lexer.NewWithFilename(filename, source))
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		t.Fatalf("parse macro: %v", p.Errors())
	}
	return program.Statements[0].(*ast.ExpressionStatement).Expression.(*ast.MacroLiteral)
}

func assertMacroSourceSpan(t *testing.T, span token.Span, source string) {
	t.Helper()
	if span.Source == nil || span.Source.Name != "macros.mk" || span.Source.Text != source {
		t.Fatalf("macro lost its source: %+v", span)
	}
	if span.Start.Offset != 0 || span.End.Offset != len(source) {
		t.Fatalf("macro range does not cover its definition: %+v", span)
	}
}
