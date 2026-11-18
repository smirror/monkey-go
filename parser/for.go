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
	var init ast.Statement

	// Initialization is a let-statement, an expression, or empty.
	if !p.curTokenIs(token.SEMICOLON) {
		if p.curTokenIs(token.LET) {
			init = p.parseLetStatement()
			if init == nil {
				return nil
			}
			// parseLetStatement consumes the first semicolon itself.
			if !p.curTokenIs(token.SEMICOLON) {
				p.errors = append(p.errors, "expected ';' after for initializer")
				return nil
			}
		} else {
			firstToken := p.curToken
			expr := p.parseExpression(LOWEST)
			if expr == nil || !p.expectPeek(token.SEMICOLON) {
				return nil
			}
			init = &ast.ExpressionStatement{Token: firstToken, Expression: expr}
		}
	}

	p.nextToken() // Start of the condition, after the first ';'.
	if p.curTokenIs(token.SEMICOLON) {
		p.errors = append(p.errors, "for condition must not be empty")
		return nil
	}
	condition := p.parseExpression(LOWEST)
	if condition == nil || !p.expectPeek(token.SEMICOLON) {
		return nil
	}

	p.nextToken() // Start of the optional post expression.
	var post ast.Expression
	if !p.curTokenIs(token.RPAREN) {
		post = p.parseExpression(LOWEST)
		if post == nil || !p.expectPeek(token.RPAREN) {
			return nil
		}
	}
	if !p.expectPeek(token.LBRACE) {
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
