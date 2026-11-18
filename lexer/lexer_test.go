package lexer

import (
	"testing"

	"monkey-go/token"
)

func TestNextToken(t *testing.T) {
	input := `
let five = 5;
let ten = 10;

let add = fn(x,y){
	x+y;
};

let result = add(five, ten);
!-/*5;
5 < 10 > 5;

if (5<10){
	return true;
} else{
	return false;
}

10 == 10;
10 != 9;
"foobar"
"foo bar"
[1, 2];
{"foo": "bar"}
`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.LET, "let"},
		{token.IDENT, "five"},
		{token.ASSIGN, "="},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},

		{token.LET, "let"},
		{token.IDENT, "ten"},
		{token.ASSIGN, "="},
		{token.INT, "10"},
		{token.SEMICOLON, ";"},

		{token.LET, "let"},
		{token.IDENT, "add"},
		{token.ASSIGN, "="},
		{token.FUNCTION, "fn"},
		{token.LPAREN, "("},
		{token.IDENT, "x"},
		{token.COMMA, ","},
		{token.IDENT, "y"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.IDENT, "x"},
		{token.PLUS, "+"},
		{token.IDENT, "y"},
		{token.SEMICOLON, ";"},

		{token.RBRACE, "}"},
		{token.SEMICOLON, ";"},
		{token.LET, "let"},
		{token.IDENT, "result"},
		{token.ASSIGN, "="},
		{token.IDENT, "add"},
		{token.LPAREN, "("},
		{token.IDENT, "five"},
		{token.COMMA, ","},
		{token.IDENT, "ten"},
		{token.RPAREN, ")"},
		{token.SEMICOLON, ";"},

		{token.BANG, "!"},
		{token.MINUS, "-"},
		{token.SLASH, "/"},
		{token.ASTERISK, "*"},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},

		{token.INT, "5"},
		{token.LT, "<"},
		{token.INT, "10"},
		{token.GT, ">"},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},

		{token.IF, "if"},
		{token.LPAREN, "("},
		{token.INT, "5"},
		{token.LT, "<"},
		{token.INT, "10"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.TRUE, "true"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.ELSE, "else"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.FALSE, "false"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},

		{token.INT, "10"},
		{token.EQ, "=="},
		{token.INT, "10"},
		{token.SEMICOLON, ";"},

		{token.INT, "10"},
		{token.NOT_EQ, "!="},
		{token.INT, "9"},
		{token.SEMICOLON, ";"},

		{token.STRING, "foobar"},
		{token.STRING, "foo bar"},

		{token.LBRACKET, "["},
		{token.INT, "1"},
		{token.COMMA, ","},
		{token.INT, "2"},
		{token.RBRACKET, "]"},
		{token.SEMICOLON, ";"},

		{token.LBRACE, "{"},
		{token.STRING, "foo"},
		{token.COLON, ":"},
		{token.STRING, "bar"},
		{token.RBRACE, "}"},

		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q", i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q", i, tt.expectedLiteral, tok.Literal)
		}

	}
}

func TestLogicalOperators(t *testing.T) {
	input := `true && false;
true || false;`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.TRUE, "true"},
		{token.AND, "&&"},
		{token.FALSE, "false"},
		{token.SEMICOLON, ";"},

		{token.TRUE, "true"},
		{token.OR, "||"},
		{token.FALSE, "false"},
		{token.SEMICOLON, ";"},

		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q", i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestInvalidLogicalOperators(t *testing.T) {
	tests := []struct {
		input string
		want  []token.Token
	}{
		{"&", []token.Token{{Type: token.ILLEGAL, Literal: "&"}, {Type: token.EOF, Literal: ""}}},
		{"|", []token.Token{{Type: token.ILLEGAL, Literal: "|"}, {Type: token.EOF, Literal: ""}}},
		{"&Ħ", []token.Token{{Type: token.ILLEGAL, Literal: "&"}, {Type: token.ILLEGAL, Literal: "Ħ"}, {Type: token.EOF, Literal: ""}}},
		{"|ż", []token.Token{{Type: token.ILLEGAL, Literal: "|"}, {Type: token.ILLEGAL, Literal: "ż"}, {Type: token.EOF, Literal: ""}}},
		{"&Ħ&&", []token.Token{{Type: token.ILLEGAL, Literal: "&"}, {Type: token.ILLEGAL, Literal: "Ħ"}, {Type: token.AND, Literal: "&&"}, {Type: token.EOF, Literal: ""}}},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			for i, want := range tt.want {
				got := l.NextToken()
				if got.Type != want.Type || got.Literal != want.Literal {
					t.Fatalf("token[%d] got=%+v, want=%+v", i, got, want)
				}
			}
		})
	}
}

func TestForLoopTokens(t *testing.T) {
	input := `for (i in range(2, 10)) { i++; }
for (i = 0; i < 10; i++) {}`
	tests := []struct {
		kind token.TokenType
		text string
	}{
		{token.FOR, "for"}, {token.LPAREN, "("}, {token.IDENT, "i"},
		{token.IN, "in"}, {token.IDENT, "range"}, {token.LPAREN, "("},
		{token.INT, "2"}, {token.COMMA, ","}, {token.INT, "10"},
		{token.RPAREN, ")"}, {token.RPAREN, ")"}, {token.LBRACE, "{"},
		{token.IDENT, "i"}, {token.INCREMENT, "++"}, {token.SEMICOLON, ";"},
		{token.RBRACE, "}"}, {token.FOR, "for"}, {token.LPAREN, "("},
		{token.IDENT, "i"}, {token.ASSIGN, "="}, {token.INT, "0"},
		{token.SEMICOLON, ";"}, {token.IDENT, "i"}, {token.LT, "<"},
		{token.INT, "10"}, {token.SEMICOLON, ";"}, {token.IDENT, "i"},
		{token.INCREMENT, "++"}, {token.RPAREN, ")"}, {token.LBRACE, "{"},
		{token.RBRACE, "}"}, {token.EOF, ""},
	}
	l := New(input)
	for i, want := range tests {
		got := l.NextToken()
		if got.Type != want.kind || got.Literal != want.text {
			t.Fatalf("token[%d]: got=%+v, want=%+v", i, got, want)
		}
	}
}
