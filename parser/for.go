package parser

import (
	"monkey-go/ast"
	"monkey-go/token"
)

// The only for syntax is: for identifier in iterable { ... }.
func (p *Parser) parseForStatement() ast.Statement {
	forToken := p.curToken
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	return p.parseForInStatement(forToken)
}

func (p *Parser) parseForInStatement(forToken token.Token) ast.Statement {
	variable := identifier(p.curToken)
	if !p.expectPeek(token.IN) {
		return nil
	}
	p.nextToken() // First token of the iterable expression.

	iterable := p.parseExpression(LOWEST)
	if iterable == nil {
		return nil
	}
	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	return &ast.ForInStatement{
		Token:    forToken,
		Variable: variable,
		Iterable: iterable,
		Body:     p.parseBlockStatement(),
	}
}
