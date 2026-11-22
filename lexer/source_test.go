package lexer

import (
	"monkey-go/token"
	"testing"
)

type locatedToken struct {
	kind    token.TokenType
	literal string
	start   token.Position
	end     token.Position
}

func TestTokenSpansUseByteOffsetsAndRuneColumns(t *testing.T) {
	input := "let s = \"あ🙂\";\r\ns + 1\r@\n"
	want := []locatedToken{
		{token.LET, "let", pos(0, 1, 1), pos(3, 1, 4)},
		{token.IDENT, "s", pos(4, 1, 5), pos(5, 1, 6)},
		{token.ASSIGN, "=", pos(6, 1, 7), pos(7, 1, 8)},
		{token.STRING, "あ🙂", pos(8, 1, 9), pos(17, 1, 13)},
		{token.SEMICOLON, ";", pos(17, 1, 13), pos(18, 1, 14)},
		{token.IDENT, "s", pos(20, 2, 1), pos(21, 2, 2)},
		{token.PLUS, "+", pos(22, 2, 3), pos(23, 2, 4)},
		{token.INT, "1", pos(24, 2, 5), pos(25, 2, 6)},
		{token.ILLEGAL, "@", pos(26, 3, 1), pos(27, 3, 2)},
		{token.EOF, "", pos(28, 4, 1), pos(28, 4, 1)},
	}
	checkLocatedTokens(t, input, want)
}

func TestSkippedCommentsAndMultilineStringsKeepPositions(t *testing.T) {
	tests := []struct {
		input string
		want  []locatedToken
	}{
		{"// あ🙂\r\n\tlet", []locatedToken{
			{token.LET, "let", pos(13, 2, 2), pos(16, 2, 5)},
			{token.EOF, "", pos(16, 2, 5), pos(16, 2, 5)},
		}},
		{"// あ🙂", []locatedToken{
			{token.EOF, "", pos(10, 1, 6), pos(10, 1, 6)},
		}},
		{"\"あ\n🙂\" @", []locatedToken{
			{token.STRING, "あ\n🙂", pos(0, 1, 1), pos(10, 2, 3)},
			{token.ILLEGAL, "@", pos(11, 2, 4), pos(12, 2, 5)},
			{token.EOF, "", pos(12, 2, 5), pos(12, 2, 5)},
		}},
		{"", []locatedToken{
			{token.EOF, "", pos(0, 1, 1), pos(0, 1, 1)},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) { checkLocatedTokens(t, tt.input, tt.want) })
	}
}

func TestInvalidInputDoesNotMasqueradeAsEOF(t *testing.T) {
	tests := []struct {
		input string
		want  []locatedToken
	}{
		{"a\x00b", []locatedToken{
			{token.IDENT, "a", pos(0, 1, 1), pos(1, 1, 2)},
			{token.ILLEGAL, "\x00", pos(1, 1, 2), pos(2, 1, 3)},
			{token.IDENT, "b", pos(2, 1, 3), pos(3, 1, 4)},
			{token.EOF, "", pos(3, 1, 4), pos(3, 1, 4)},
		}},
		{"\xff+1", []locatedToken{
			{token.ILLEGAL, "�", pos(0, 1, 1), pos(1, 1, 2)},
			{token.PLUS, "+", pos(1, 1, 2), pos(2, 1, 3)},
			{token.INT, "1", pos(2, 1, 3), pos(3, 1, 4)},
			{token.EOF, "", pos(3, 1, 4), pos(3, 1, 4)},
		}},
		{"\"あ\n🙂", []locatedToken{
			{token.ILLEGAL, "\"あ\n🙂", pos(0, 1, 1), pos(9, 2, 2)},
			{token.EOF, "", pos(9, 2, 2), pos(9, 2, 2)},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) { checkLocatedTokens(t, tt.input, tt.want) })
	}
}

func TestTwoCharacterOperatorsIncludeBothRunes(t *testing.T) {
	checkLocatedTokens(t, "== != && ||", []locatedToken{
		{token.EQ, "==", pos(0, 1, 1), pos(2, 1, 3)},
		{token.NOT_EQ, "!=", pos(3, 1, 4), pos(5, 1, 6)},
		{token.AND, "&&", pos(6, 1, 7), pos(8, 1, 9)},
		{token.OR, "||", pos(9, 1, 10), pos(11, 1, 12)},
		{token.EOF, "", pos(11, 1, 12), pos(11, 1, 12)},
	})
}

func pos(offset, line, column int) token.Position {
	return token.Position{Offset: offset, Line: line, Column: column}
}

func checkLocatedTokens(t *testing.T, input string, want []locatedToken) {
	t.Helper()
	l := NewWithFilename("source.mk", input)
	var source *token.Source
	for i, expected := range want {
		got := l.NextToken()
		if source == nil {
			source = got.Span.Source
		}
		checkSource(t, got.Span.Source, source, input)
		if got.Type != expected.kind || got.Literal != expected.literal {
			t.Fatalf("token[%d]: got %s %q, want %s %q", i, got.Type, got.Literal, expected.kind, expected.literal)
		}
		if got.Span.Start != expected.start || got.Span.End != expected.end {
			t.Fatalf("token[%d]: range = %+v, want %+v..%+v", i, got.Span, expected.start, expected.end)
		}
	}
	firstEOF := l.NextToken()
	if nextEOF := l.NextToken(); nextEOF != firstEOF || firstEOF.Type != token.EOF {
		t.Fatalf("EOF must have a stable location: first=%+v, next=%+v", firstEOF, nextEOF)
	}
}

func checkSource(t *testing.T, got, want *token.Source, input string) {
	t.Helper()
	if got == nil || got != want || got.Name != "source.mk" || got.Text != input {
		t.Fatalf("missing or changed source snapshot: %+v", got)
	}
}
