package parser

import (
	"fmt"
	"monkey/ast"
	lxr "monkey/lexer"
	"monkey/token"
	"strconv"
)

const (
	_ int = iota
	LOWEST
	EQUALS      //==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      //-X !X
	CALL        // myfoo(X)
)

var precedences = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	token.LPAREN:   CALL,
}

type Parser struct {
	l      *lxr.Lexer
	errors []string

	curToken  token.Token
	peekToken token.Token

	//For parsing prefix operators
	prefixParseFns map[token.TokenType]prefixParseFn

	//For parsing infix operators
	infixParseFns map[token.TokenType]infixParseFn
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

func New(l *lxr.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}
	/*
	   Parsing protocol for the parsing functions - prefix or infix -:
	   Start with curToken being the type of token you're associated with
	   and return with curToken being the last token that's part of your
	   expression type. Never advance the token too far
	*/
	//Map the prefix/infix operators
	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.TRUE, p.parseBooleanExpression)
	p.registerPrefix(token.FALSE, p.parseBooleanExpression)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.IF, p.parseIfExpression)
	p.registerPrefix(token.FUNCTION, p.parseFunctionLiteral)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)

	//Read two tokens - sets curToken and peekToken
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}

		p.nextToken()

	}

	return program
}

// Parse a statement
func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

//Return Statement
func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}

	p.nextToken()

	stmt.ReturnValue = p.parseExpression(LOWEST)

	for p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

//Let  Statement
func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken}

	//Identifier is expected, found?
	if !p.expectedPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectedPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken()

	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {

	stmt := &ast.ExpressionStatement{Token: p.curToken}

	stmt.Expression = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) HasErrors() bool {
	return len(p.errors) != 0
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("Mismatch token[expected='%s', got='%s']", t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

/*
 * The function validates our expectation of the next token (peek):
 * If the expectation is validated we move to this token,
 * If NOT the program syntax is invalid and we report and error
 */
func (p *Parser) expectedPeek(t token.TokenType) bool {
	if !p.peekTokenIs(t) {
		p.peekError(t)
		return false
	}

	p.nextToken()
	return true
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for token `%s` found", t)
	p.errors = append(p.errors, msg)
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]

	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}

	leftExp := prefix()

	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()

		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {

	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)

	if err != nil {
		msg := fmt.Sprintf("could not parse %q as interger", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value

	return lit
}

func (p *Parser) parseFunctionLiteral() ast.Expression {
	expression := &ast.FunctionLiteral{
		Token:      p.curToken, //"fn"
		Parameters: []*ast.Identifier{},
	}

	if !p.expectedPeek(token.LPAREN) {
		return nil
	}

	//parse fun parameters: fn(x,y){..}
	if !p.peekTokenIs(token.RPAREN) {
		//parse parameters x,y...
		p.nextToken()
		for {

			iden := p.parseIdentifier()

			expression.Parameters = append(expression.Parameters, iden.(*ast.Identifier))
			if !p.peekTokenIs(token.COMMA) {
				break //we are done
			}
			//skip token.COMMA and to the next parameter
			p.nextToken()
			p.nextToken()
		}
	}

	//check right paren in fn(...)
	if !p.expectedPeek(token.RPAREN) {
		return nil
	}

	if !p.expectedPeek(token.LBRACE) {
		return nil
	}

	expression.Body = p.parseBlockStatement()

	return expression
}

func (p *Parser) parseIfExpression() ast.Expression {

	expression := &ast.IfExpression{
		Token: p.curToken, //"if"
	}

	if !p.expectedPeek(token.LPAREN) {
		return nil
	}

	//move current token to the first token in the Parentences: If(...)
	p.nextToken()

	expression.Condition = p.parseExpression(LOWEST)

	if !p.expectedPeek(token.RPAREN) {
		return nil
	}

	//Expecting '{'. validate and move to the consequence left brace: If(<cond>){...}
	if !p.expectedPeek(token.LBRACE) {
		return nil
	}

	expression.Consequence = p.parseBlockStatement()

	//do we have an else: if(<cond>){...}else{...}
	if !p.peekTokenIs(token.ELSE) {
		return expression
	}

	//move to else
	p.nextToken()
	//Expecting '{'. validate and move to the alternative left brace: If(<cond>){...}else{...}
	if !p.expectedPeek(token.LBRACE) {
		return nil
	}

	expression.Alternative = p.parseBlockStatement()

	return expression
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {

	block := &ast.BlockStatement{
		Token: p.curToken, //"{"
	}

	p.nextToken()

	//parse all the statements in the block
	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}

		p.nextToken()
	}

	return block
}

func (p *Parser) parseGroupedExpression() ast.Expression {

	p.nextToken()

	exp := p.parseExpression(LOWEST)

	if !p.expectedPeek(token.RPAREN) {
		return nil
	}

	return exp
}

func (p *Parser) parseBooleanExpression() ast.Expression {

	return &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{Token: p.curToken, Operator: p.curToken.Literal}

	p.nextToken()

	expression.Right = p.parseExpression(PREFIX)

	return expression
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseCallArguments()

	return exp
}

func (p *Parser) parseCallArguments() []ast.Expression {
	args := []ast.Expression{}
	//parse call arguments: foo(x,y)
	if !p.peekTokenIs(token.RPAREN) {
		//parse arguments x,y...
		p.nextToken()
		for {

			args = append(args, p.parseExpression(LOWEST))

			if !p.peekTokenIs(token.COMMA) {
				break //we are done
			}
			//skip token.COMMA and to the next argument
			p.nextToken()
			p.nextToken()
		}
	}

	//check right paren in foo(...)
	if !p.expectedPeek(token.RPAREN) {
		return nil
	}

	return args
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}
