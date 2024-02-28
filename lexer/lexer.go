package lexer

import (
	"danielmcm.com/interpreterbook/token"
)

type Lexer struct {
	input string
	// current position in input
	position int
	// current reading position in input
	readPosition int
	// current character
	char byte
}

func New(input string) *Lexer {
	lexer := &Lexer{input: input}
	lexer.readChar()
	return lexer
}

func (lexer *Lexer) NextToken() token.Token {
	var nextToken token.Token

	lexer.readMatching(isWhitespace)

	switch lexer.char {
	case '=':
		if lexer.peekChar() == '=' {
			lexer.readChar()
			nextToken = token.Token{Type: token.EQ, Literal: "=="}
		} else {
			nextToken = token.Token{Type: token.ASSIGN, Literal: string(lexer.char)}
		}

	case '+':
		nextToken = token.Token{Type: token.PLUS, Literal: string(lexer.char)}
	case '-':
		nextToken = token.Token{Type: token.MINUS, Literal: string(lexer.char)}
	case '!':
		if lexer.peekChar() == '=' {
			lexer.readChar()
			nextToken = token.Token{Type: token.NOT_EQ, Literal: string("!=")}
		} else {
			nextToken = token.Token{Type: token.BANG, Literal: string(lexer.char)}
		}
	case '/':
		nextToken = token.Token{Type: token.SLASH, Literal: string(lexer.char)}
	case '*':
		nextToken = token.Token{Type: token.ASTERISK, Literal: string(lexer.char)}
	case '<':
		nextToken = token.Token{Type: token.LT, Literal: string(lexer.char)}
	case '>':
		nextToken = token.Token{Type: token.GT, Literal: string(lexer.char)}
	case ';':
		nextToken = token.Token{Type: token.SEMICOLON, Literal: string(lexer.char)}
	case '(':
		nextToken = token.Token{Type: token.LPAREN, Literal: string(lexer.char)}
	case ')':
		nextToken = token.Token{Type: token.RPAREN, Literal: string(lexer.char)}
	case ',':
		nextToken = token.Token{Type: token.COMMA, Literal: string(lexer.char)}
	case '{':
		nextToken = token.Token{Type: token.LBRACE, Literal: string(lexer.char)}
	case '}':
		nextToken = token.Token{Type: token.RBRACE, Literal: string(lexer.char)}
	case 0:
		nextToken = token.Token{Type: token.EOF, Literal: ""}
	default:
		if isLetter(lexer.char) {
			identifier := lexer.readMatching(isLetter)
			return token.Token{Type: token.LookupIdentifier(identifier), Literal: identifier}
		} else if isDigit(lexer.char) {
			return token.Token{Type: token.INT, Literal: lexer.readMatching(isDigit)}
		} else {
			nextToken = token.Token{Type: token.ILLEGAL, Literal: string(lexer.char)}
		}
	}
	lexer.readChar()
	return nextToken
}

func (lexer *Lexer) readChar() {
	lexer.char = lexer.peekChar()
	lexer.position = lexer.readPosition
	lexer.readPosition += 1
}

func (lexer *Lexer) peekChar() byte {
	if lexer.readPosition >= len(lexer.input) {
		return 0
	} else {
		return lexer.input[lexer.readPosition]
	}
}

func (lexer *Lexer) readMatching(predicate func(byte) bool) string {
	position := lexer.position
	if position >= len(lexer.input) {
		return ""
	}
	for predicate(lexer.char) {
		lexer.readChar()
	}
	return lexer.input[position:lexer.position]
}

func isLetter(char byte) bool {
	return ('a' <= char && char <= 'z') || ('A' <= char && char <= 'Z') || char == '_'
}

func isDigit(char byte) bool {
	return '0' <= char && char <= '9'
}

func isWhitespace(char byte) bool {
	return char == ' ' || char == '\t' || char == '\n' || char == '\r'
}
