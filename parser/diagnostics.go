package parser

import (
	"fmt"
	"monkey-go/ast"
	"monkey-go/diagnostic"
	"monkey-go/token"
	"strings"
)

// Diagnostic is a syntax error paired with its source snapshot and range.
type Diagnostic = diagnostic.Diagnostic

// Diagnostics supplements Errors without changing its message-only API.
func (p *Parser) Diagnostics() []Diagnostic {
	return p.diagnostics
}

func (p *Parser) addError(message string, tok token.Token) {
	p.errors = append(p.errors, message)
	p.diagnostics = append(p.diagnostics, Diagnostic{Kind: diagnostic.Syntax, Message: message, Span: tok.Span})
}

func (p *Parser) addLexicalError(tok token.Token) {
	if tok.Type != token.ILLEGAL {
		return
	}
	message := fmt.Sprintf("illegal token %q", tok.Literal)
	if strings.HasPrefix(tok.Literal, "\"") {
		message = "unterminated string literal"
	}
	p.addError(message, tok)
}

func (p *Parser) locate(node ast.Node, start token.Span) {
	ast.SetSpan(node, token.Span{Source: start.Source, Start: start.Start, End: p.curToken.Span.End})
}

func identifier(tok token.Token) *ast.Identifier {
	id := &ast.Identifier{Token: tok, Value: tok.Literal}
	ast.SetSpan(id, tok.Span)
	return id
}
