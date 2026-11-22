package parser

import (
	"monkey-go/ast"
	"monkey-go/lexer"
	"monkey-go/token"
	"testing"
	"unicode/utf8"
)

type diagnosticCase struct {
	input   string
	message string
	start   token.Position
	end     token.Position
}

func TestSyntaxDiagnosticsIdentifyTheOffendingToken(t *testing.T) {
	tests := []diagnosticCase{
		{"let = 1", "expected next token to be IDENT, got = instead.", position(4, 1, 5), position(5, 1, 6)},
		{"let x = @;", "illegal token \"@\"", position(8, 1, 9), position(9, 1, 10)},
		{"fn(@) {}", "illegal token \"@\"", position(3, 1, 4), position(4, 1, 5)},
		{"\"あ\"\nfor i range(2) { i }", "expected next token to be IN, got IDENT instead.", position(12, 2, 7), position(17, 2, 12)},
		{"for i in [1] i", "expected next token to be {, got IDENT instead.", position(13, 1, 14), position(14, 1, 15)},
		{"for i in @ {}", "illegal token \"@\"", position(9, 1, 10), position(10, 1, 11)},
		{"let f = fn() {\n  1", "unterminated block: expected } before EOF", position(18, 2, 4), position(18, 2, 4)},
		{"\"あ\n🙂", "unterminated string literal", position(0, 1, 1), position(9, 2, 2)},
		{"9223372036854775808", "could not parse \"9223372036854775808\" as integer", position(0, 1, 1), position(19, 1, 20)},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) { checkDiagnostic(t, tt) })
	}
}

func checkDiagnostic(t *testing.T, tt diagnosticCase) {
	t.Helper()
	p := New(lexer.NewWithFilename("errors.mk", tt.input))
	_ = p.ParseProgram()
	diagnostics := p.Diagnostics()
	if len(diagnostics) == 0 {
		t.Fatal("expected a located diagnostic")
	}
	first := diagnostics[0]
	expected := Diagnostic{
		Message: tt.message,
		Span:    token.Span{Source: first.Span.Source, Start: tt.start, End: tt.end},
	}
	if first != expected {
		t.Fatalf("diagnostic = %+v, want %+v", first, expected)
	}
	checkDiagnosticSource(t, first.Span.Source, tt.input)
	checkDiagnosticMessages(t, p.Errors(), diagnostics)
}

func checkDiagnosticSource(t *testing.T, source *token.Source, input string) {
	t.Helper()
	if source == nil {
		t.Fatal("diagnostic lost its source")
	}
	expected := token.Source{Name: "errors.mk", Text: input}
	if *source != expected {
		t.Fatalf("diagnostic source = %+v, want %+v", *source, expected)
	}
}

func checkDiagnosticMessages(t *testing.T, messages []string, diagnostics []Diagnostic) {
	t.Helper()
	if len(messages) != len(diagnostics) {
		t.Fatalf("message and diagnostic counts differ: %v vs %v", messages, diagnostics)
	}
	for i, diagnostic := range diagnostics {
		if diagnostic.Message != messages[i] {
			t.Fatalf("Errors()[%d] does not match its diagnostic", i)
		}
	}
}

func TestExpressionSpansIncludeClosingAndGroupingDelimiters(t *testing.T) {
	expressions := []string{
		"x", "123", "true", "\"あ🙂\"", "-x", "1 + 2", "(1 + 2) * 3",
		"f(2)", "[1, 2][0]", "fn(x){x}", "{}", "{\"key\": [1]}",
		"if (true) { 1 } else { 2 }", "((1))",
	}
	for _, expression := range expressions {
		t.Run(expression, func(t *testing.T) {
			program := parseLocatedProgram(t, "\n  "+expression+" // trailing comment")
			statement := program.Statements[0].(*ast.ExpressionStatement)
			span := statement.Expression.Span()
			checkSpanText(t, span, expression)
			if span.Start != position(3, 2, 3) || span.End.Column != 3+utf8.RuneCountInString(expression) {
				t.Fatalf("unexpected expression coordinates: %+v", span)
			}
		})
	}
}

func TestNestedASTSpansKeepTokensAndBodiesDistinct(t *testing.T) {
	input := "let f = fn(x) { return (x + 1); };\nfor i in [1, 2] { f(i); }"
	program := parseLocatedProgram(t, input)
	declaration := program.Statements[0].(*ast.LetStatement)
	function := declaration.Value.(*ast.FunctionLiteral)
	returned := function.Body.Statements[0].(*ast.ReturnStatement)
	addition := returned.ReturnValue.(*ast.InfixExpression)
	loop := program.Statements[1].(*ast.ForInStatement)
	checks := []struct {
		span token.Span
		text string
	}{
		{program.Span(), input},
		{declaration.Span(), "let f = fn(x) { return (x + 1); };"},
		{declaration.Name.Span(), "f"},
		{function.Span(), "fn(x) { return (x + 1); }"},
		{function.Parameters[0].Span(), "x"},
		{function.Body.Span(), "{ return (x + 1); }"},
		{returned.Span(), "return (x + 1);"},
		{addition.Span(), "(x + 1)"},
		{addition.Token.Span, "+"},
		{loop.Span(), "for i in [1, 2] { f(i); }"},
		{loop.Variable.Span(), "i"},
		{loop.Iterable.Span(), "[1, 2]"},
		{loop.Body.Span(), "{ f(i); }"},
	}
	for _, check := range checks {
		checkSpanText(t, check.span, check.text)
	}
}

func TestSourceSnapshotsSurviveLaterSubmissions(t *testing.T) {
	original := "let f = fn() { missing; };"
	first := parseLocatedProgram(t, original)
	function := first.Statements[0].(*ast.LetStatement).Value.(*ast.FunctionLiteral)
	second := parseLocatedProgram(t, "f()")
	if function.Span().Source == second.Span().Source {
		t.Fatal("separate inputs unexpectedly share a mutable source snapshot")
	}
	if function.Span().Source.Text != original {
		t.Fatal("later input replaced an earlier definition's source")
	}
	checkSpanText(t, function.Body.Statements[0].Span(), "missing;")
}

func TestInvalidListNodesHaveSafeSpans(t *testing.T) {
	for _, input := range []string{"[1,]", "f(,)", "if (;) {}", "\x00let x = 1"} {
		t.Run(input, func(t *testing.T) {
			p := New(lexer.NewWithFilename("invalid.mk", input))
			program := p.ParseProgram()
			if len(p.Diagnostics()) == 0 {
				t.Fatal("expected diagnostic for invalid input")
			}
			checkSpanText(t, program.Span(), input)
			for _, statement := range program.Statements {
				if statement.Span().Source == nil {
					t.Fatal("remaining AST node has no source")
				}
			}
		})
	}
}

func position(offset, line, column int) token.Position {
	return token.Position{Offset: offset, Line: line, Column: column}
}

func parseLocatedProgram(t *testing.T, input string) *ast.Program {
	t.Helper()
	p := New(lexer.NewWithFilename("source.mk", input))
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		t.Fatalf("unexpected parser errors: %v", p.Errors())
	}
	return program
}

func checkSpanText(t *testing.T, span token.Span, want string) {
	t.Helper()
	checkSpanBounds(t, span)
	if got := span.Source.Text[span.Start.Offset:span.End.Offset]; got != want {
		t.Fatalf("range text = %q, want %q (%+v)", got, want, span)
	}
}

func checkSpanBounds(t *testing.T, span token.Span) {
	t.Helper()
	if span.Source == nil {
		t.Fatal("range has no source snapshot")
	}
	if span.Start.Offset < 0 {
		t.Fatalf("range starts before the source: %+v", span)
	}
	if span.Start.Offset > span.End.Offset {
		t.Fatalf("range ends before it starts: %+v", span)
	}
	if span.End.Offset > len(span.Source.Text) {
		t.Fatalf("range ends after the source: %+v", span)
	}
}
