package parser

import (
	"fmt"

	"danielmcm.com/interpreterbook/ast"
	"danielmcm.com/interpreterbook/lexer"
	"danielmcm.com/interpreterbook/token"
)

type Parser struct {
	lexer *lexer.Lexer

	currentToken token.Token
	peekToken    token.Token
}

type ParseError struct {
	reason string
}

func New(lexer *lexer.Lexer) *Parser {
	parser := &Parser{lexer: lexer}

	// Populate current and peek token
	parser.nextToken()
	parser.nextToken()

	return parser
}

func (parser *Parser) nextToken() {
	parser.currentToken = parser.peekToken
	parser.peekToken = parser.lexer.NextToken()
}

func (parser *Parser) ParseProgram() (*ast.Program, error) {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for parser.currentToken.Type != token.EOF {
		statement, err := parser.ParseStatement()
		if err == nil {
			program.Statements = append(program.Statements, statement)
		} else {
			return nil, err
		}

		parser.nextToken()
	}

	return program, nil
}

func (parser *Parser) ParseStatement() (ast.Statement, error) {
	switch parser.currentToken.Type {
	case token.LET:
		return parser.ParseLetStatement()
	default:
		return nil, ParseError{reason: fmt.Sprintf("unexpected token %s, expected statement", parser.currentToken.Literal)}
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

	statement.Value = parser.ParseExpression()
	for parser.currentToken.Literal != token.SEMICOLON {
		parser.nextToken()
	}

	return statement, nil
}

func (parser *Parser) ParseExpression() ast.Expression {
	return nil
}

func (parser *Parser) expectPeek(tokenType token.TokenType) error {
	if parser.peekToken.Type == tokenType {
		parser.nextToken()
		return nil
	} else {
		return ParseError{reason: fmt.Sprintf("unexpected token %s, expected %s", parser.peekToken.Literal, tokenType)}
	}
}

func (parseError ParseError) Error() string {
	return parseError.reason
}
