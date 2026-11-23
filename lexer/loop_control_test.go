package lexer

import (
	"monkey-go/token"
	"testing"
)

func TestLoopControlKeywordsAndSourceSpans(t *testing.T) {
	input := "\"日本\"\nbreak continue breakfast continued"
	checkLocatedTokens(t, input, []locatedToken{
		{token.STRING, "日本", pos(0, 1, 1), pos(8, 1, 5)},
		{token.BREAK, "break", pos(9, 2, 1), pos(14, 2, 6)},
		{token.CONTINUE, "continue", pos(15, 2, 7), pos(23, 2, 15)},
		{token.IDENT, "breakfast", pos(24, 2, 16), pos(33, 2, 25)},
		{token.IDENT, "continued", pos(34, 2, 26), pos(43, 2, 35)},
		{token.EOF, "", pos(43, 2, 35), pos(43, 2, 35)},
	})
}
