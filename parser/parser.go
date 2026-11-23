package parser

import (
	"fmt"
	"monkey-go/ast"
	"monkey-go/lexer"
	"monkey-go/token"
	"strconv"
)

const (
	_ int = iota
	LOWEST
	ASSIGN_PREC // =
	OR_PREC     // ||
	AND_PREC    // &&
	EQUALS      // ==
	LESSGREATER // <, <=, >, >=
	SUM         // +
	PRODUCT     // *, /, %
	PREFIX      // -X or !X
	CALL        // myFunction(X)
	INDEX       // array[index]
)

var precedencs = map[token.TokenType]int{
	token.ASSIGN:   ASSIGN_PREC, // =
	token.OR:       OR_PREC,     // ||
	token.AND:      AND_PREC,    // &&
	token.EQ:       EQUALS,      // ==
	token.NOT_EQ:   EQUALS,      // !=
	token.LT:       LESSGREATER, // <
	token.GT:       LESSGREATER, // >
	token.LTE:      LESSGREATER, // <=
	token.GTE:      LESSGREATER, // >=
	token.PLUS:     SUM,         // +
	token.MINUS:    SUM,         // -
	token.SLASH:    PRODUCT,     // /
	token.ASTERISK: PRODUCT,     // *
	token.MODULO:   PRODUCT,     // %
	token.LPAREN:   CALL,        // (
	token.LBRACKET: INDEX,       // [
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

type Parser struct {
	l *lexer.Lexer

	errors      []string
	diagnostics []Diagnostic

	curToken  token.Token
	peekToken token.Token

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parserIntegerLiteral)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.TRUE, p.parseBoolean)
	p.registerPrefix(token.FALSE, p.parseBoolean)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.IF, p.parseIfExpression)
	p.registerPrefix(token.FUNCTION, p.parseFunctionExpression)
	p.registerPrefix(token.LBRACKET, p.parseArrayLiteral)
	p.registerPrefix(token.LBRACE, p.parseHashMapLiteral)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.MODULO, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.LTE, p.parseInfixExpression)
	p.registerInfix(token.GTE, p.parseInfixExpression)
	p.registerInfix(token.AND, p.parseInfixExpression)
	p.registerInfix(token.OR, p.parseInfixExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)
	p.registerInfix(token.LBRACKET, p.parseIndexExpression)
	p.registerInfix(token.ASSIGN, p.parseAssignExpression)

	//2つトークンに読み込む。curTokenとpeekTokenの両方がセットされる。
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(t token.TokenType) {
	if p.peekTokenIs(token.ILLEGAL) {
		return // The lexer error was recorded when this token was read.
	}
	msg := fmt.Sprintf("expected next token to be %s, got %s instead.", t, p.peekToken.Type)
	p.addError(msg, p.peekToken)
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
	p.addLexicalError(p.peekToken)
}

func (p *Parser) ParseProgram() *ast.Program {
	start := p.curToken.Span
	program := &ast.Program{Statements: p.parseStatements(token.EOF)}
	p.locate(program, start)
	return program
}

// parseStatements is shared by the top level and nested blocks. Failed
// statements are omitted from the AST while parse errors remain available.
func (p *Parser) parseStatements(end token.TokenType) []ast.Statement {
	statements := []ast.Statement{}
	for !p.curTokenIs(token.EOF) && !p.curTokenIs(end) {
		if stmt := p.parseStatement(); stmt != nil {
			statements = append(statements, stmt)
		}
		p.nextToken()
	}
	return statements
}

func (p *Parser) parseStatement() ast.Statement {
	start := p.curToken.Span
	statement := p.parseStatementNode()
	p.locate(statement, start)
	return statement
}

func (p *Parser) parseStatementNode() ast.Statement {
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	case token.FOR:
		return p.parseForStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseLetStatement() ast.Statement {
	stmt := &ast.LetStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = identifier(p.curToken)

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken()

	stmt.Value = p.parseStatementValue()
	if stmt.Value == nil {
		return nil
	}

	return stmt
}

// parseStatementValue parses an expression and consumes a trailing semicolon,
// if present, without advancing past the next statement.
func (p *Parser) parseStatementValue() ast.Expression {
	value := p.parseExpression(LOWEST)
	if value == nil {
		return nil
	}
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return value
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

func (p *Parser) parseReturnStatement() ast.Statement {
	stmt := &ast.ReturnStatement{Token: p.curToken}

	p.nextToken()

	stmt.ReturnValue = p.parseStatementValue()
	if stmt.ReturnValue == nil {
		return nil
	}

	return stmt
}

func (p *Parser) parseExpressionStatement() ast.Statement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}

	stmt.Expression = p.parseStatementValue()
	if stmt.Expression == nil {
		return nil
	}

	return stmt
}

func (p *Parser) parseIdentifier() ast.Expression {
	return identifier(p.curToken)
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	if t == token.ILLEGAL {
		return // Avoid duplicating the diagnostic for an invalid token.
	}
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.addError(msg, p.curToken)
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	start := p.curToken.Span
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()
	if leftExp == nil {
		return nil
	}
	p.locate(leftExp, start)

	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()

		leftExp = infix(leftExp)
		if leftExp == nil {
			return nil
		}
		p.locate(leftExp, start)
	}

	return leftExp
}

func (p *Parser) parserIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.addError(msg, p.curToken)
		return nil
	}
	lit.Value = value
	return lit
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	operator := p.curToken
	right := p.parseRightOperand(PREFIX)
	if right == nil {
		return nil
	}
	return &ast.PrefixExpression{
		Token:    operator,
		Operator: operator.Literal,
		Right:    right,
	}
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	return p.parseInfixExpressionWithPrecedence(left, p.curPrecedence())
}

func (p *Parser) parseAssignExpression(left ast.Expression) ast.Expression {
	return p.parseInfixExpressionWithPrecedence(left, p.curPrecedence()-1)
}

func (p *Parser) parseInfixExpressionWithPrecedence(left ast.Expression, precedence int) ast.Expression {
	operator := p.curToken
	right := p.parseRightOperand(precedence)
	if right == nil {
		return nil
	}
	return &ast.InfixExpression{
		Token:    operator,
		Operator: string(operator.Type), // Normalize aliases while keeping the source token.
		Left:     left,
		Right:    right,
	}
}

func (p *Parser) parseRightOperand(precedence int) ast.Expression {
	p.nextToken()
	return p.parseExpression(precedence)
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedencs[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedencs[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	exp := p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return exp
}

func (p *Parser) parseIfExpression() ast.Expression {
	expression := &ast.IfExpression{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()
	expression.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	expression.Consequence = p.parseBlockStatement()

	if p.peekTokenIs(token.ELSE) {
		p.nextToken()

		if !p.expectPeek(token.LBRACE) {
			return nil
		}

		expression.Alternative = p.parseBlockStatement()
	}

	return expression

}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}

	p.nextToken()
	block.Statements = p.parseStatements(token.RBRACE)
	if p.curTokenIs(token.EOF) {
		p.addError("unterminated block: expected } before EOF", p.curToken)
	}
	p.locate(block, block.Token.Span)
	return block
}

func (p *Parser) parseFunctionExpression() ast.Expression {
	lit := &ast.FunctionLiteral{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	lit.Parameters = p.parseFunctionParameters()

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	lit.Body = p.parseBlockStatement()

	return lit
}

func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	var identifiers []*ast.Identifier

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return identifiers
	}

	p.nextToken()

	ident := identifier(p.curToken)
	identifiers = append(identifiers, ident)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		ident := identifier(p.curToken)
		identifiers = append(identifiers, ident)
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return identifiers
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseCallArguments()
	return exp
}

func (p *Parser) parseCallArguments() []ast.Expression {
	var args []ast.Expression

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return args
	}

	p.nextToken()
	args = append(args, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		args = append(args, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return args
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{Token: p.curToken}

	array.Elements = p.parseExpressionList(token.RBRACKET)

	return array
}

func (p *Parser) parseExpressionList(end token.TokenType) []ast.Expression {
	var list []ast.Expression

	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	p.nextToken()
	list = append(list, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(end) {
		return nil
	}

	return list
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{Token: p.curToken, Left: left}

	p.nextToken()
	exp.Index = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RBRACKET) {
		return nil
	}

	return exp
}

func (p *Parser) parseHashMapLiteral() ast.Expression {
	hashMap := &ast.HashMapLiteral{Token: p.curToken}
	hashMap.Pairs = make(map[ast.Expression]ast.Expression)

	for !p.peekTokenIs(token.RBRACE) {
		p.nextToken()
		key := p.parseExpression(LOWEST)

		if !p.expectPeek(token.COLON) {
			return nil
		}

		p.nextToken()
		value := p.parseExpression(LOWEST)

		hashMap.Pairs[key] = value

		if !p.peekTokenIs(token.RBRACE) && !p.expectPeek(token.COMMA) {
			return nil
		}
	}

	if !p.expectPeek(token.RBRACE) {
		return nil
	}

	return hashMap
}
