package lexer

import (
	"monkey-go/token"
	"unicode/utf8"
)

type Lexer struct {
	input        string
	source       *token.Source
	position     int // UTF-8 byte offset of the current rune
	readPosition int // UTF-8 byte offset of the next rune
	r            rune
	current      token.Position
	next         token.Position
}

func New(input string) *Lexer {
	return NewWithFilename("", input)
}

// NewWithFilename associates diagnostics with a source name and text snapshot.
func NewWithFilename(filename, input string) *Lexer {
	l := &Lexer{
		input:  input,
		source: &token.Source{Name: filename, Text: input},
		next:   token.Position{Line: 1, Column: 1},
	}
	l.readRune()
	return l
}

func (l *Lexer) readRune() {
	previous := l.r
	l.position = l.readPosition
	l.current = l.next
	if l.atEnd() {
		l.r = 0
		return
	}
	var size int
	l.r, size = utf8.DecodeRuneInString(l.input[l.position:])
	l.readPosition += size
	l.next = advancePosition(l.current, l.r, previous, size)
}

func (l *Lexer) atEnd() bool {
	return l.position >= len(l.input)
}

// CRLF is one newline; a lone CR or LF is also a newline.
func advancePosition(pos token.Position, current, previous rune, size int) token.Position {
	pos.Offset += size
	if current == '\r' || current == '\n' {
		if current != '\n' || previous != '\r' {
			pos.Line++
		}
		pos.Column = 1
	} else {
		pos.Column++
	}
	return pos
}

var singleCharacterTokens = map[rune]token.TokenType{
	'+': token.PLUS, '-': token.MINUS, '/': token.SLASH,
	'*': token.ASTERISK, '%': token.MODULO, ';': token.SEMICOLON,
	':': token.COLON, ',': token.COMMA, '(': token.LPAREN,
	')': token.RPAREN, '{': token.LBRACE, '}': token.RBRACE,
	'[': token.LBRACKET, ']': token.RBRACKET,
}

var twoCharacterTokens = map[rune]struct {
	second   rune
	matched  token.TokenType
	fallback token.TokenType
}{
	'=': {'=', token.EQ, token.ASSIGN},
	'!': {'=', token.NOT_EQ, token.BANG},
	'&': {'&', token.AND, token.ILLEGAL},
	'|': {'|', token.OR, token.ILLEGAL},
	'<': {'=', token.LTE, token.LT},
	'>': {'=', token.GTE, token.GT},
}

func (l *Lexer) NextToken() token.Token {
	l.skipIgnored()
	start := l.current
	tok := l.readToken()
	tok.Span = token.Span{Source: l.source, Start: start, End: l.current}
	return tok
}

func (l *Lexer) skipIgnored() {
	l.skipWhitespace()
	for l.r == '/' && l.peekRune() == '/' {
		l.skipLineComment()
		l.skipWhitespace()
	}
}

func (l *Lexer) readToken() token.Token {
	if l.atEnd() {
		return token.Token{Type: token.EOF}
	}
	if operator, ok := twoCharacterTokens[l.r]; ok {
		tok := l.makeTwoCharToken(operator.second, operator.matched, operator.fallback)
		l.readRune()
		return tok
	}
	if l.r == '"' {
		return l.readString()
	}
	if isLetter(l.r) {
		literal := l.readIdentifier()
		return token.Token{Type: token.LookupIdent(literal), Literal: literal}
	}
	if isDigit(l.r) {
		return token.Token{Type: token.INT, Literal: l.readNumber()}
	}
	kind, ok := singleCharacterTokens[l.r]
	if !ok {
		kind = token.ILLEGAL
	}
	tok := newToken(kind, l.r)
	l.readRune()
	return tok
}

func (l *Lexer) skipLineComment() {
	for !l.atEnd() && l.r != '\n' && l.r != '\r' {
		l.readRune()
	}
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
	return l.input[position:l.position]
}

func isLetter(r rune) bool {
	return 'a' <= r && r <= 'z' || 'A' <= r && r <= 'Z' || r == '_'
}

func (l *Lexer) readNumber() string {
	position := l.position
	for isDigit(l.r) {
		l.readRune()
	}
	return l.input[position:l.position]
}

func isDigit(r rune) bool {
	return '0' <= r && r <= '9'
}

func newToken(tokenType token.TokenType, r rune) token.Token {
	return token.Token{Type: tokenType, Literal: string(r)}
}

// makeTwoCharToken は2文字演算子を処理するヘルパー関数
func (l *Lexer) makeTwoCharToken(expected rune, twoCharType, oneCharType token.TokenType) token.Token {
	if l.peekRune() == expected {
		r := l.r
		l.readRune()
		return token.Token{Type: twoCharType, Literal: string(r) + string(l.r)}
	}
	return newToken(oneCharType, l.r)
}

func (l *Lexer) peekRune() rune {
	if l.readPosition >= len(l.input) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(l.input[l.readPosition:])
	return r
}

func (l *Lexer) readString() token.Token {
	position := l.position
	l.readRune()
	for !l.atEnd() && l.r != '"' {
		l.readRune()
	}
	if l.atEnd() {
		return token.Token{Type: token.ILLEGAL, Literal: l.input[position:l.position]}
	}
	literal := l.input[position+1 : l.position]
	l.readRune()
	return token.Token{Type: token.STRING, Literal: literal}
}
