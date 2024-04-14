package parser

import (
	"fmt"
	"strconv"

	"danielmcm.com/interpreterbook/ast"
	"danielmcm.com/interpreterbook/lexer"
	"danielmcm.com/interpreterbook/token"
)

type Parser struct {
	lexer *lexer.Lexer

	currentToken token.Token
	peekToken    token.Token
	errors       []error

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

type (
	prefixParseFn func() (ast.Expression, error)
	infixParseFn  func(ast.Expression) (ast.Expression, error)
)

type ParseError struct {
	reason string
}

const (
	_ int = iota
	LOWEST
	EQUALS
	LESSGREATER
	SUM
	PRODUCT
	PREFIX
	CALL
)

var precedences = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.ASTERISK: PRODUCT,
	token.SLASH:    PRODUCT,
	token.LPAREN:   CALL,
}

func getPrecedence(tokenType token.TokenType) int {
	if precedence, ok := precedences[tokenType]; ok {
		return precedence
	}
	return LOWEST
}

func New(lexer *lexer.Lexer) *Parser {
	parser := &Parser{lexer: lexer}

	parser.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	parser.infixParseFns = make(map[token.TokenType]infixParseFn)
	parser.registerPrefix(token.IDENT, parser.parseIdentifier)
	parser.registerPrefix(token.INT, parser.parseIntegerLiteral)
	parser.registerPrefix(token.TRUE, parser.parseBooleanLiteral)
	parser.registerPrefix(token.FALSE, parser.parseBooleanLiteral)
	parser.registerPrefix(token.BANG, parser.parsePrefixExpression)
	parser.registerPrefix(token.MINUS, parser.parsePrefixExpression)
	parser.registerPrefix(token.IF, parser.parseIfExpression)
	parser.registerPrefix(token.FUNCTION, parser.parseFunctionalLiteral)
	parser.registerInfix(token.EQ, parser.parseInfixExpression)
	parser.registerInfix(token.NOT_EQ, parser.parseInfixExpression)
	parser.registerInfix(token.LT, parser.parseInfixExpression)
	parser.registerInfix(token.GT, parser.parseInfixExpression)
	parser.registerInfix(token.PLUS, parser.parseInfixExpression)
	parser.registerInfix(token.MINUS, parser.parseInfixExpression)
	parser.registerInfix(token.ASTERISK, parser.parseInfixExpression)
	parser.registerInfix(token.SLASH, parser.parseInfixExpression)
	parser.registerPrefix(token.LPAREN, parser.parseGroupedExpression)
	parser.registerInfix(token.LPAREN, parser.parseCallExpression)

	// Populate current and peek token
	parser.nextToken()
	parser.nextToken()

	return parser
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (parser *Parser) nextToken() {
	parser.currentToken = parser.peekToken
	parser.peekToken = parser.lexer.NextToken()
}

func (parser *Parser) Errors() []error {
	return parser.errors
}

func (parser *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for !parser.currentTokenIs(token.EOF) {
		statement, err := parser.ParseStatement()
		if err == nil {
			program.Statements = append(program.Statements, statement)
		} else {
			parser.errors = append(parser.errors, err)
		}

		parser.nextToken()
	}

	return program
}

func (parser *Parser) ParseStatement() (ast.Statement, error) {
	switch parser.currentToken.Type {
	case token.LET:
		return parser.ParseLetStatement()
	case token.RETURN:
		return parser.ParseReturnStatement()
	default:
		return parser.ParseExpressionStatement()
	}
}

func (parser *Parser) ParseLetStatement() (*ast.LetStatement, error) {
	statement := &ast.LetStatement{Token: parser.currentToken}

	if err := parser.expectPeek(token.IDENT); err != nil {
		return nil, err
	}

	statement.Name = &ast.Identifier{Token: parser.currentToken, Value: parser.currentToken.Literal}

	if err := parser.expectPeek(token.ASSIGN); err != nil {
		return nil, err
	}
	parser.nextToken()

	expr, err := parser.ParseExpression(LOWEST)
	if err != nil {
		return nil, err
	}
	statement.Value = expr

	if parser.peekTokenIs(token.SEMICOLON) {
		parser.nextToken()
	}
	return statement, nil
}

func (parser *Parser) ParseReturnStatement() (*ast.ReturnStatement, error) {
	statement := &ast.ReturnStatement{Token: parser.currentToken}
	parser.nextToken()

	expr, err := parser.ParseExpression(LOWEST)
	if err != nil {
		return nil, err
	}
	statement.ReturnValue = expr

	if parser.peekTokenIs(token.SEMICOLON) {
		parser.nextToken()
	}
	return statement, nil
}

func (parser *Parser) ParseExpressionStatement() (*ast.ExpressionStatement, error) {
	statement := &ast.ExpressionStatement{Token: parser.currentToken}
	expr, err := parser.ParseExpression(LOWEST)
	if err != nil {
		return nil, err
	}
	statement.Expression = expr

	for parser.peekTokenIs(token.SEMICOLON) {
		parser.nextToken()
	}
	return statement, nil
}

func (parser *Parser) ParseExpression(precedence int) (ast.Expression, error) {
	prefix := parser.prefixParseFns[parser.currentToken.Type]

	if prefix == nil {
		return nil, ParseError{reason: fmt.Sprintf("expected expression, got token %q", parser.currentToken.Literal)}
	}

	leftExpr, err := prefix()
	if err != nil {
		return nil, err
	}

	for !parser.peekTokenIs(token.SEMICOLON) && precedence < getPrecedence(parser.peekToken.Type) {
		infix := parser.infixParseFns[parser.peekToken.Type]
		if infix == nil {
			return nil, ParseError{reason: fmt.Sprintf("cannot parse infix expression for operator %q", parser.peekToken.Literal)}
		}
		parser.nextToken()
		infixExpr, err := infix(leftExpr)
		if err != nil {
			return nil, err
		}
		leftExpr = infixExpr
	}
	return leftExpr, nil
}

func (parser *Parser) parseIdentifier() (ast.Expression, error) {
	return &ast.Identifier{Token: parser.currentToken, Value: parser.currentToken.Literal}, nil
}

func (parser *Parser) parseIntegerLiteral() (ast.Expression, error) {
	value, err := strconv.ParseInt(parser.currentToken.Literal, 10, 64)
	if err != nil {
		return nil, err
	}
	return &ast.IntegerLiteral{Token: parser.currentToken, Value: value}, nil
}

func (parser *Parser) parseBooleanLiteral() (ast.Expression, error) {
	return &ast.BooleanLiteral{
		Token: parser.currentToken,
		Value: parser.currentTokenIs(token.TRUE),
	}, nil
}

func (parser *Parser) parsePrefixExpression() (ast.Expression, error) {
	expr := &ast.PrefixExpression{
		Token:    parser.currentToken,
		Operator: parser.currentToken.Literal,
	}
	parser.nextToken()
	right, err := parser.ParseExpression(PREFIX)
	if err != nil {
		return nil, err
	}
	expr.Right = right
	return expr, nil
}

func (parser *Parser) parseInfixExpression(left ast.Expression) (ast.Expression, error) {
	expr := &ast.InfixExpression{
		Token:    parser.currentToken,
		Operator: parser.currentToken.Literal,
		Left:     left,
	}
	precedence := getPrecedence(parser.currentToken.Type)
	parser.nextToken()

	right, err := parser.ParseExpression(precedence)
	if err != nil {
		return nil, err
	}
	expr.Right = right
	return expr, nil
}

func (parser *Parser) parseGroupedExpression() (ast.Expression, error) {
	parser.nextToken()

	expr, err := parser.ParseExpression(LOWEST)
	if err != nil {
		return nil, err
	}
	if err = parser.expectPeek(token.RPAREN); err != nil {
		return nil, err
	}

	return expr, nil
}

func (parser *Parser) parseBlockStatement() (*ast.BlockStatement, error) {
	block := &ast.BlockStatement{Token: parser.currentToken}
	block.Statements = []ast.Statement{}
	parser.nextToken()

	for !parser.currentTokenIs(token.RBRACE) && !parser.currentTokenIs(token.EOF) {
		statement, err := parser.ParseStatement()
		if err == nil {
			block.Statements = append(block.Statements, statement)
		} else {
			return nil, err
		}

		parser.nextToken()
	}

	return block, nil
}

func (parser *Parser) parseIfExpression() (ast.Expression, error) {
	expr := &ast.IfExpression{Token: parser.currentToken}

	if err := parser.expectPeek(token.LPAREN); err != nil {
		return nil, err
	}
	parser.nextToken()

	condExpr, err := parser.ParseExpression(LOWEST)
	if err != nil {
		return nil, err
	}
	expr.Condition = condExpr
	if err := parser.expectPeek(token.RPAREN); err != nil {
		return nil, err
	}

	if err := parser.expectPeek(token.LBRACE); err != nil {
		return nil, err
	}
	consequence, err := parser.parseBlockStatement()
	if err != nil {
		return nil, err
	}
	expr.Consequence = consequence

	if parser.peekTokenIs(token.ELSE) {
		parser.nextToken()
		if err := parser.expectPeek(token.LBRACE); err != nil {
			return nil, err
		}
		alternative, err := parser.parseBlockStatement()
		if err != nil {
			return nil, err
		}
		expr.Alternative = alternative
	}

	return expr, nil
}

func (parser *Parser) parseFunctionalLiteral() (ast.Expression, error) {
	expr := &ast.FunctionLiteral{Token: parser.currentToken, Parameters: make([]ast.Identifier, 0)}

	if err := parser.expectPeek(token.LPAREN); err != nil {
		return nil, err
	}
	for parser.currentTokenIs(token.COMMA) || !parser.peekTokenIs(token.RPAREN) {
		if err := parser.expectPeek(token.IDENT); err != nil {
			return nil, err
		}
		expr.Parameters = append(expr.Parameters, ast.Identifier{Token: parser.currentToken, Value: parser.currentToken.Literal})
		if parser.peekTokenIs(token.COMMA) {
			parser.nextToken()
		} else {
			break
		}
	}
	if err := parser.expectPeek(token.RPAREN); err != nil {
		return nil, err
	}

	if err := parser.expectPeek(token.LBRACE); err != nil {
		return nil, err
	}
	body, err := parser.parseBlockStatement()
	if err != nil {
		return nil, err
	}
	expr.Body = body

	return expr, nil
}

func (parser *Parser) parseCallExpression(function ast.Expression) (ast.Expression, error) {
	expr := &ast.CallExpression{Token: parser.currentToken, Function: function, Arguments: make([]ast.Expression, 0)}
	for parser.currentTokenIs(token.COMMA) || !parser.peekTokenIs(token.RPAREN) {
		parser.nextToken()
		arg, err := parser.ParseExpression(LOWEST)
		if err != nil {
			return nil, err
		}
		expr.Arguments = append(expr.Arguments, arg)
		if parser.peekTokenIs(token.COMMA) {
			parser.nextToken()
		} else {
			break
		}
	}
	if err := parser.expectPeek(token.RPAREN); err != nil {
		return nil, err
	}
	return expr, nil
}

func (parser *Parser) currentTokenIs(tokenType token.TokenType) bool {
	return parser.currentToken.Type == tokenType
}

func (parser *Parser) peekTokenIs(tokenType token.TokenType) bool {
	return parser.peekToken.Type == tokenType
}

func (parser *Parser) expectPeek(tokenType token.TokenType) error {
	if parser.peekTokenIs(tokenType) {
		parser.nextToken()
		return nil
	} else {
		return ParseError{reason: fmt.Sprintf("unexpected token %s, expected %s", parser.peekToken.Literal, tokenType)}
	}
}

func (parseError ParseError) Error() string {
	return parseError.reason
}
