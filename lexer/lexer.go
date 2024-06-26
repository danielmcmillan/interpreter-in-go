package lexer

import (
	"bytes"
	"errors"
	"fmt"

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

var ErrLexer error = errors.New("tokenisation error")

func New(input string) *Lexer {
	lexer := &Lexer{input: input}
	lexer.readChar()
	return lexer
}

func (lexer *Lexer) NextToken() (token.Token, error) {
	var nextToken token.Token
	var err error

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
	case ':':
		nextToken = token.Token{Type: token.COLON, Literal: string(lexer.char)}
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
	case '[':
		nextToken = token.Token{Type: token.LBRACKET, Literal: string(lexer.char)}
	case ']':
		nextToken = token.Token{Type: token.RBRACKET, Literal: string(lexer.char)}
	case '"':
		nextToken, err = lexer.readString()
	case 0:
		nextToken = token.Token{Type: token.EOF, Literal: ""}
	default:
		if isLetter(lexer.char) {
			identifier := lexer.readMatching(isLetter)
			return token.Token{Type: token.LookupIdentifier(identifier), Literal: identifier}, nil
		} else if isDigit(lexer.char) {
			return token.Token{Type: token.INT, Literal: lexer.readMatching(isDigit)}, nil
		} else {
			nextToken = token.Token{Type: token.ILLEGAL, Literal: string(lexer.char)}
		}
	}
	lexer.readChar()
	return nextToken, err
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
	for predicate(lexer.char) && lexer.char != 0 {
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

func (lexer *Lexer) readString() (token.Token, error) {
	var literal bytes.Buffer
	for {
		lexer.readChar()
		char := lexer.char
		if char == '"' {
			break
		} else if char == '\\' {
			lexer.readChar()
			char = lexer.char
			switch char {
			case 'n':
				char = '\n'
			case 't':
				char = '\t'
			}
		}
		if char == 0 {
			return token.Token{}, fmt.Errorf("unterminated string literal")
		}
		literal.WriteByte(char)
	}
	return token.Token{Type: token.STRING, Literal: literal.String()}, nil
}
