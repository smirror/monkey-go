package lexer

import (
	"monkey-go/token"
	"testing"
)

func TestComparisonAndModuloTokensKeepSourceSpans(t *testing.T) {
	input := "\"あ\" <= 2 // 🙂\r\n3>=1%2"
	checkLocatedTokens(t, input, []locatedToken{
		{token.STRING, "あ", pos(0, 1, 1), pos(5, 1, 4)},
		{token.LTE, "<=", pos(6, 1, 5), pos(8, 1, 7)},
		{token.INT, "2", pos(9, 1, 8), pos(10, 1, 9)},
		{token.INT, "3", pos(20, 2, 1), pos(21, 2, 2)},
		{token.GTE, ">=", pos(21, 2, 2), pos(23, 2, 4)},
		{token.INT, "1", pos(23, 2, 4), pos(24, 2, 5)},
		{token.MODULO, "%", pos(24, 2, 5), pos(25, 2, 6)},
		{token.INT, "2", pos(25, 2, 6), pos(26, 2, 7)},
		{token.EOF, "", pos(26, 2, 7), pos(26, 2, 7)},
	})
}

func TestComparisonTokensUseLongestMatch(t *testing.T) {
	checkLocatedTokens(t, "< <= > >= <== >== %=", []locatedToken{
		{token.LT, "<", pos(0, 1, 1), pos(1, 1, 2)},
		{token.LTE, "<=", pos(2, 1, 3), pos(4, 1, 5)},
		{token.GT, ">", pos(5, 1, 6), pos(6, 1, 7)},
		{token.GTE, ">=", pos(7, 1, 8), pos(9, 1, 10)},
		{token.LTE, "<=", pos(10, 1, 11), pos(12, 1, 13)},
		{token.ASSIGN, "=", pos(12, 1, 13), pos(13, 1, 14)},
		{token.GTE, ">=", pos(14, 1, 15), pos(16, 1, 17)},
		{token.ASSIGN, "=", pos(16, 1, 17), pos(17, 1, 18)},
		{token.MODULO, "%", pos(18, 1, 19), pos(19, 1, 20)},
		{token.ASSIGN, "=", pos(19, 1, 20), pos(20, 1, 21)},
		{token.EOF, "", pos(20, 1, 21), pos(20, 1, 21)},
	})
}

func TestLessOrEqualAliasKeepsSourceSpellingAndSpan(t *testing.T) {
	checkLocatedTokens(t, "\"あ\" =< 2\n3=<4", []locatedToken{
		{token.STRING, "あ", pos(0, 1, 1), pos(5, 1, 4)},
		{token.LTE, "=<", pos(6, 1, 5), pos(8, 1, 7)},
		{token.INT, "2", pos(9, 1, 8), pos(10, 1, 9)},
		{token.INT, "3", pos(11, 2, 1), pos(12, 2, 2)},
		{token.LTE, "=<", pos(12, 2, 2), pos(14, 2, 4)},
		{token.INT, "4", pos(14, 2, 4), pos(15, 2, 5)},
		{token.EOF, "", pos(15, 2, 5), pos(15, 2, 5)},
	})
}

func TestLessOrEqualAliasKeepsAssignmentAndEqualityDistinct(t *testing.T) {
	checkLocatedTokens(t, "= == =< = < => ==< =<=", []locatedToken{
		{token.ASSIGN, "=", pos(0, 1, 1), pos(1, 1, 2)},
		{token.EQ, "==", pos(2, 1, 3), pos(4, 1, 5)},
		{token.LTE, "=<", pos(5, 1, 6), pos(7, 1, 8)},
		{token.ASSIGN, "=", pos(8, 1, 9), pos(9, 1, 10)},
		{token.LT, "<", pos(10, 1, 11), pos(11, 1, 12)},
		{token.ASSIGN, "=", pos(12, 1, 13), pos(13, 1, 14)},
		{token.GT, ">", pos(13, 1, 14), pos(14, 1, 15)},
		{token.EQ, "==", pos(15, 1, 16), pos(17, 1, 18)},
		{token.LT, "<", pos(17, 1, 18), pos(18, 1, 19)},
		{token.LTE, "=<", pos(19, 1, 20), pos(21, 1, 22)},
		{token.ASSIGN, "=", pos(21, 1, 22), pos(22, 1, 23)},
		{token.EOF, "", pos(22, 1, 23), pos(22, 1, 23)},
	})
}
