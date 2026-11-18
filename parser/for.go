package parser

import (
	"fmt"
	"monkey-go/ast"
	"monkey-go/token"
)

// parseForStatement distinguishes for-in from the semicolon-separated classic form.
func (p *Parser) parseForStatement() ast.Statement {
	forToken := p.curToken
	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	p.nextToken() // First token after '('.

	if p.curTokenIs(token.IDENT) && p.peekTokenIs(token.IN) {
		return p.parseForInStatement(forToken)
	}
	return p.parseClassicForStatement(forToken)
}

func (p *Parser) parseForInStatement(forToken token.Token) ast.Statement {
	variable := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	p.nextToken() // 'in'
	p.nextToken() // Start of the iterable expression.

	iterable := p.parseExpression(LOWEST)
	if iterable == nil || !p.expectPeek(token.RPAREN) || !p.expectPeek(token.LBRACE) {
		return nil
	}

	return &ast.ForInStatement{
		Token:    forToken,
		Variable: variable,
		Iterable: iterable,
		Body:     p.parseBlockStatement(),
	}
}

func (p *Parser) parseClassicForStatement(forToken token.Token) ast.Statement {
	init, ok := p.parseForInitializer()
	if !ok {
		return nil
	}
	condition, ok := p.parseForCondition()
	if !ok {
		return nil
	}
	post, ok := p.parseForPost()
	if !ok || !p.expectPeek(token.LBRACE) {
		return nil
	}
	return &ast.ForStatement{
		Token:     forToken,
		Init:      init,
		Condition: condition,
		Post:      post,
		Body:      p.parseBlockStatement(),
	}
}

func (p *Parser) parseForInitializer() (ast.Statement, bool) {
	if p.curTokenIs(token.SEMICOLON) {
		return nil, true
	}
	if p.curTokenIs(token.LET) {
		return p.parseForLetInitializer()
	}

	firstToken := p.curToken
	expr := p.parseExpression(LOWEST)
	if expr == nil || !p.expectPeek(token.SEMICOLON) {
		return nil, false
	}
	return &ast.ExpressionStatement{Token: firstToken, Expression: expr}, true
}

func (p *Parser) parseForLetInitializer() (ast.Statement, bool) {
	letStmt := p.parseLetStatement()
	if letStmt == nil {
		return nil, false
	}
	if !p.curTokenIs(token.SEMICOLON) {
		p.errors = append(p.errors, "expected ';' after for initializer")
		return nil, false
	}
	return letStmt, true
}

func (p *Parser) parseForCondition() (ast.Expression, bool) {
	p.nextToken() // After initializer ';'.
	if p.curTokenIs(token.SEMICOLON) {
		p.errors = append(p.errors, "for condition must not be empty")
		return nil, false
	}
	condition := p.parseExpression(LOWEST)
	if condition == nil || !p.expectPeek(token.SEMICOLON) {
		return nil, false
	}
	return condition, true
}

func (p *Parser) parseForPost() (ast.Expression, bool) {
	p.nextToken() // After condition ';'.
	if p.curTokenIs(token.RPAREN) {
		return nil, true
	}
	post := p.parseExpression(LOWEST)
	if post == nil || !p.expectPeek(token.RPAREN) {
		return nil, false
	}
	return post, true
}

func (p *Parser) parsePostfixExpression(left ast.Expression) ast.Expression {
	if left == nil {
		p.errors = append(p.errors, fmt.Sprintf("missing operand for %s", p.curToken.Literal))
		return nil
	}
	return &ast.PostfixExpression{
		Token:    p.curToken,
		Left:     left,
		Operator: p.curToken.Literal,
	}
}
