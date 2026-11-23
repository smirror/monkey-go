package lexer

import (
	"monkey-go/token"
	"testing"
)

func TestLoopControlKeywordsAndSourceSpans(t *testing.T) {
	l := NewWithFilename("loops.mk", "\"日本\"\nbreak continue breakfast continued")
	_ = l.NextToken()
	for _, want := range []struct {
		kind           token.TokenType
		literal        string
		offset, column int
	}{
		{token.BREAK, "break", 9, 1},
		{token.CONTINUE, "continue", 15, 7},
	} {
		got := l.NextToken()
		if got.Type != want.kind || got.Literal != want.literal {
			t.Fatalf("got %+v, want %s %q", got, want.kind, want.literal)
		}
		if got.Span.Start.Line != 2 || got.Span.Start.Offset != want.offset || got.Span.Start.Column != want.column {
			t.Fatalf("incorrect keyword source span: %+v", got.Span)
		}
	}
	for _, word := range []string{"breakfast", "continued"} {
		if got := l.NextToken(); got.Type != token.IDENT || got.Literal != word {
			t.Fatalf("keyword prefix changed identifier %q into %+v", word, got)
		}
	}
}
