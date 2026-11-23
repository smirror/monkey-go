package parser

import (
	"fmt"
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
		Body:     p.parseLoopBody(),
	}
}

func (p *Parser) parseLoopBody() *ast.BlockStatement {
	p.loopDepth++
	defer func() { p.loopDepth-- }()
	return p.parseBlockStatement()
}

func (p *Parser) parseLoopControlStatement() ast.Statement {
	keyword := p.curToken
	if p.loopDepth == 0 {
		p.addError(keyword.Literal+" outside loop", keyword)
	}
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	} else if !p.loopControlEndsHere() {
		p.addError(fmt.Sprintf("%s does not take a value or label", keyword.Literal), p.peekToken)
	}
	if keyword.Type == token.BREAK {
		return &ast.BreakStatement{Token: keyword}
	}
	return &ast.ContinueStatement{Token: keyword}
}

func (p *Parser) loopControlEndsHere() bool {
	return p.peekTokenIs(token.RBRACE) || p.peekTokenIs(token.EOF) ||
		p.peekToken.Span.Start.Line > p.curToken.Span.End.Line
}
