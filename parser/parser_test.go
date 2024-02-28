package parser

import (
	"testing"

	"danielmcm.com/interpreterbook/ast"
	"danielmcm.com/interpreterbook/lexer"
)

func TestLetStatements(t *testing.T) {
	input := `
		let x = 5;
		let y = 10;
		let foobar = 838383;
	`

	lexer := lexer.New(input)
	parser := New(lexer)

	program, err := parser.ParseProgram()
	if err != nil {
		t.Fatalf("ParserProgram() returned error: %v", err)
	} else if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	} else if len(program.Statements) != 3 {
		t.Fatalf("program.Statements does not contain 3 statements. Got %d", len(program.Statements))
	}

	tests := []struct {
		expectedIdentifier string
	}{
		{"x"},
		{"y"},
		{"foobar"},
	}
	for i, test := range tests {
		statement := program.Statements[i]
		if statement.TokenLiteral() != "let" {
			t.Errorf("Statement TokenLiteral not 'let'. Got %q", statement.TokenLiteral())
			break
		}

		letStatement, ok := statement.(*ast.LetStatement)
		if !ok {
			t.Errorf("Statement not ast.LetStatement. Got %T", statement)
			break
		}

		if letStatement.Name.Value != test.expectedIdentifier {
			t.Errorf("Let statement identifier not '%s'. Got %q.", test.expectedIdentifier, letStatement.Name.Value)
			break
		}
	}
}
