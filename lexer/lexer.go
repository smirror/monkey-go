package lexer

import (
	"monkey-go/token"
)

type Lexer struct {
	input        []rune
	position     int  // 入力における現在の位置
	readPosition int  // 入力における次の位置
	r            rune // 現在見ている文字
}

func New(input string) *Lexer {
	ir := []rune(input)
	l := &Lexer{input: ir}
	l.readRune()
	return l
}

func (l *Lexer) readRune() {
	if l.readPosition >= len(l.input) { // input = Null or 終端に達した場合
		l.r = 0
	} else {
		l.r = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition += 1
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhitespace()

	switch l.r {
	case '=':
		if l.peekRune() == '=' {
			r := l.r
			l.readRune()
			literal := string(r) + string(l.r)
			tok = token.Token{Type: token.EQ, Literal: literal}
		} else {
			tok = newToken(token.ASSIGN, l.r)
		}
	case '+':
		tok = newToken(token.PLUS, l.r)
	case '-':
		tok = newToken(token.MINUS, l.r)
	case '!':
		if l.peekRune() == '=' {
			r := l.r
			l.readRune()
			literal := string(r) + string(l.r)
			tok = token.Token{Type: token.NOT_EQ, Literal: literal}
		} else {
			tok = newToken(token.BANG, l.r)
		}
	case '/':
		tok = newToken(token.SLASH, l.r)
	case '*':
		tok = newToken(token.ASTERISK, l.r)
	case '<':
		tok = newToken(token.LT, l.r)
	case '>':
		tok = newToken(token.GT, l.r)
	case ';':
		tok = newToken(token.SEMICOLON, l.r)
	case ',':
		tok = newToken(token.COMMA, l.r)
	case '(':
		tok = newToken(token.LPAREN, l.r)
	case ')':
		tok = newToken(token.RPAREN, l.r)
	case '{':
		tok = newToken(token.LBRACE, l.r)
	case '}':
		tok = newToken(token.RBRACE, l.r)
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if isLetter(l.r) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			return tok
		} else if isDigit(l.r) {
			tok.Type = token.INT
			tok.Literal = l.readNumber()
			return tok
		} else {
			tok = newToken(token.ILLEGAL, l.r)
		}
	}

	l.readRune()
	return tok
}

func (l *Lexer) skipWhitespace() {
	for l.r == ' ' || l.r == '\t' || l.r == '\n' || l.r == '\r' {
		l.readRune()
	}
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.r) {
		l.readRune()
	}
	return string(l.input[position:l.position])
}

func isLetter(r rune) bool {
	return 'a' <= r && r <= 'z' || 'A' <= r && r <= 'Z' || r == '_'
}

func (l *Lexer) readNumber() string {
	position := l.position
	for isDigit(l.r) {
		l.readRune()
	}
	return string(l.input[position:l.position])
}

func isDigit(r rune) bool {
	return '0' <= r && r <= '9'
}

func newToken(tokenType token.TokenType, r rune) token.Token {
	return token.Token{Type: tokenType, Literal: string(r)}
}

func (l *Lexer) peekRune() byte {
	if l.readPosition >= len(l.input) {
		return 0
	} else {
		return byte(l.input[l.readPosition])
	}
}
