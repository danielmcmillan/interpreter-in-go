package lexer

import (
	"testing"

	"danielmcm.com/interpreterbook/token"
)

func TestNextToken(t *testing.T) {
	input := `let five = 5;
	let ten = 10;

	let add = fn(x, y) {
		x + y;
	};

	let result = add(five, ten);
	!-/*5;
	5 < 10 > 5;

	if (5 < 10) {
		return true;
	} else {
		return false;
	}

	10 == 10;
	10 != 9;

	["foo bar", "{\"abc\": \"x\ty\n1\t2\n\"}"];
	{1:2};
	`

	tests := []struct {
		tokenType token.TokenType
		literal   string
	}{
		{token.LET, "let"},
		{token.IDENT, "five"},
		{token.ASSIGN, "="},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},
		{token.LET, "let"},
		{token.IDENT, "ten"},
		{token.ASSIGN, "="},
		{token.INT, "10"},
		{token.SEMICOLON, ";"},
		{token.LET, "let"},
		{token.IDENT, "add"},
		{token.ASSIGN, "="},
		{token.FUNCTION, "fn"},
		{token.LPAREN, "("},
		{token.IDENT, "x"},
		{token.COMMA, ","},
		{token.IDENT, "y"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.IDENT, "x"},
		{token.PLUS, "+"},
		{token.IDENT, "y"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.SEMICOLON, ";"},
		{token.LET, "let"},
		{token.IDENT, "result"},
		{token.ASSIGN, "="},
		{token.IDENT, "add"},
		{token.LPAREN, "("},
		{token.IDENT, "five"},
		{token.COMMA, ","},
		{token.IDENT, "ten"},
		{token.RPAREN, ")"},
		{token.SEMICOLON, ";"},
		{token.BANG, "!"},
		{token.MINUS, "-"},
		{token.SLASH, "/"},
		{token.ASTERISK, "*"},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},
		{token.INT, "5"},
		{token.LT, "<"},
		{token.INT, "10"},
		{token.GT, ">"},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},
		{token.IF, "if"},
		{token.LPAREN, "("},
		{token.INT, "5"},
		{token.LT, "<"},
		{token.INT, "10"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.TRUE, "true"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.ELSE, "else"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.FALSE, "false"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.INT, "10"},
		{token.EQ, "=="},
		{token.INT, "10"},
		{token.SEMICOLON, ";"},
		{token.INT, "10"},
		{token.NOT_EQ, "!="},
		{token.INT, "9"},
		{token.SEMICOLON, ";"},
		{token.LBRACKET, "["},
		{token.STRING, "foo bar"},
		{token.COMMA, ","},
		{token.STRING, "{\"abc\": \"x\ty\n1\t2\n\"}"},
		{token.RBRACKET, "]"},
		{token.SEMICOLON, ";"},
		{token.LBRACE, "{"},
		{token.INT, "1"},
		{token.COLON, ":"},
		{token.INT, "2"},
		{token.RBRACE, "}"},
		{token.SEMICOLON, ";"},
		{token.EOF, ""},
	}

	lexer := New(input)

	for i, expected := range tests {
		token, err := lexer.NextToken()
		if err != nil {
			t.Fatalf("tests[%d] - received error %v", i, err)
		}

		if token.Type != expected.tokenType {
			t.Fatalf(
				"tests[%d] - wrong TokenType. Expected=%q, Actual=%q",
				i,
				expected.tokenType,
				token.Type,
			)
		}
		if token.Literal != expected.literal {
			t.Fatalf(
				"tests[%d] - wrong Literal. Expected=%q, Actual=%q",
				i,
				expected.literal,
				token.Literal,
			)
		}
	}
}

func TestEmptyInput(t *testing.T) {
	lexer := New("")
	if tok, err := lexer.NextToken(); err != nil || tok.Type != token.EOF {
		t.Fatalf("expected EOF, got %v error %v", tok, err)
	}
}

func TestErrors(t *testing.T) {
	tests := []struct {
		input   string
		pattern string
	}{
		{"\"", "unterminated string"},
		{"\"\\\"", "unterminated string"},
		{"let x = \"hello ;", "unterminated string"},
	}

	for _, test := range tests {
		lexer := New(test.input)
		lexer.NextToken()
	}
}
