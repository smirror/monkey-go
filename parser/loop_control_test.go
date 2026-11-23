package parser

import (
	"monkey-go/ast"
	"monkey-go/diagnostic"
	"monkey-go/lexer"
	"monkey-go/token"
	"strings"
	"testing"
)

func TestLoopControlStatementsHaveCompleteSpans(t *testing.T) {
	for _, keyword := range []string{"break", "continue"} {
		t.Run(keyword, func(t *testing.T) {
			source := "for i in range(2) {\n  " + keyword + ";\n}"
			program := parseLocatedProgram(t, source)
			statement := program.Statements[0].(*ast.ForInStatement).Body.Statements[0]
			wantSyntax := [2]string{keyword, keyword + ";"}
			if got := [2]string{statement.TokenLiteral(), statement.String()}; got != wantSyntax {
				t.Fatalf("incorrect loop control AST: %T %q", statement, statement.String())
			}
			span := statement.Span()
			checkSpanText(t, span, keyword+";")
			wantSource := token.Source{Name: "source.mk", Text: source}
			if *span.Source != wantSource {
				t.Fatalf("statement source = %+v, want %+v", *span.Source, wantSource)
			}
			start := strings.Index(source, keyword)
			wantSpan := token.Span{
				Source: span.Source, Start: position(start, 2, 3),
				End: position(start+len(keyword)+1, 2, 3+len(keyword)+1),
			}
			if span != wantSpan {
				t.Fatalf("statement span = %+v, want %+v", span, wantSpan)
			}
		})
	}
}

func TestLoopControlAcceptsLoopBodiesAndOptionalTerminators(t *testing.T) {
	inputs := []string{
		"for i in [] { break }",
		"for i in [] { continue\n1 }",
		"for i in [] { if (true) { continue; } break; }",
		"for i in [] { for j in [] { continue; } break; }",
		"for i in [] { let f = fn() { for j in [] { break; } }; continue; }",
		"let f = fn() { for i in [] { break; } }; for j in [] { continue; }",
	}
	for _, input := range inputs {
		t.Run(input, func(t *testing.T) { parseForTest(t, input) })
	}
}

func TestLoopControlRejectsOutsideAndCrossFunctionUse(t *testing.T) {
	templates := []string{
		"CONTROL;",
		"\"日本\";\nCONTROL;",
		"if (true) { CONTROL; }",
		"for i in [] {} CONTROL;",
		"for i in if (true) { CONTROL; } { 1; }",
		"for i in [] { let f = fn() { CONTROL; }; }",
		"let f = fn() { for i in [] {} CONTROL; };",
		"for i in [] { let f = fn() { let g = fn() { CONTROL; }; }; }",
	}
	for _, keyword := range []string{"break", "continue"} {
		for _, template := range templates {
			input := strings.ReplaceAll(template, "CONTROL", keyword)
			t.Run(input, func(t *testing.T) { checkOutsideLoopDiagnostic(t, input, keyword) })
		}
	}
}

func checkOutsideLoopDiagnostic(t *testing.T, input, keyword string) {
	t.Helper()
	p := New(lexer.NewWithFilename("invalid.mk", input))
	_ = p.ParseProgram()
	for _, report := range p.Diagnostics() {
		if report.Message != keyword+" outside loop" {
			continue
		}
		if report.Kind != diagnostic.Syntax || report.Span.Source.Name != "invalid.mk" {
			t.Fatalf("incorrect control diagnostic: %+v", report)
		}
		if report.Span.Start.Offset != strings.Index(input, keyword) {
			t.Fatalf("control diagnostic points at a different token: %+v", report.Span)
		}
		return
	}
	t.Fatalf("missing outside-loop diagnostic for %q: %v", input, p.Errors())
}

func TestLoopControlRejectsValuesLabelsAndParameterNames(t *testing.T) {
	inputs := []string{
		"for i in [] { break 1; }",
		"for i in [] { continue outer; }",
		"fn(break) {}",
		"fn(x, continue) {}",
	}
	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			p := New(lexer.New(input))
			_ = p.ParseProgram()
			if len(p.Errors()) == 0 {
				t.Fatalf("invalid control syntax was accepted: %q", input)
			}
		})
	}
}
