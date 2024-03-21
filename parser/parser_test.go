package parser

import (
	"strings"
	"testing"

	"danielmcm.com/interpreterbook/ast"
	"danielmcm.com/interpreterbook/lexer"
)

func checkParserErrors(t *testing.T, parser *Parser) {
	errors := parser.Errors()
	if len(errors) > 0 {
		errStrings := make([]string, 0, len(errors))
		for _, err := range errors {
			errStrings = append(errStrings, err.Error())
		}
		t.Fatalf("ParserProgram() returned errors: %v", strings.Join(errStrings, ". "))
	}
}

func TestLetStatements(t *testing.T) {
	input := `
		let x = 5;
		let y = 10;
		let foobar = 838383;
	`

	lexer := lexer.New(input)
	parser := New(lexer)

	program := parser.ParseProgram()
	checkParserErrors(t, parser)
	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	} else if len(program.Statements) != 3 {
		t.Fatalf("program.Statements does not contain 3 statements. Got %d", len(program.Statements))
	}

	tests := []struct {
		expectedIdentifier string
		expectedInt        int64
	}{
		{"x", 5},
		{"y", 10},
		{"foobar", 838383},
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

		if !testIntegerLiteral(t, letStatement.Value, test.expectedInt) {
			break
		}
	}
}

func TestReturnStatements(t *testing.T) {
	input := `
		return 5;
		return 10;
		return 838383;
	`

	lexer := lexer.New(input)
	parser := New(lexer)

	program := parser.ParseProgram()
	checkParserErrors(t, parser)
	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	} else if len(program.Statements) != 3 {
		t.Fatalf("program.Statements does not contain 3 statements. Got %d", len(program.Statements))
	}

	tests := []struct {
		expectedReturnValue string
		expectedInt         int64
	}{
		{"5", 5},
		{"10", 10},
		{"838383", 838383},
	}
	for i, test := range tests {
		statement := program.Statements[i]
		if statement.TokenLiteral() != "return" {
			t.Errorf("Statement TokenLiteral not 'return'. Got %q", statement.TokenLiteral())
			break
		}

		returnStatement, ok := statement.(*ast.ReturnStatement)
		if !ok {
			t.Errorf("Statement not ast.ReturnStatement. Got %T", statement)
			break
		}

		if returnStatement.ReturnValue.TokenLiteral() != test.expectedReturnValue {
			t.Errorf("Return statement expression not '%s'. Got %q.", test.expectedReturnValue, returnStatement.ReturnValue.TokenLiteral())
			break
		}

		if !testIntegerLiteral(t, returnStatement.ReturnValue, test.expectedInt) {
			break
		}
	}
}

func TestIdentifierExpressions(t *testing.T) {
	input := "foobar;"

	lexer := lexer.New(input)
	parser := New(lexer)
	program := parser.ParseProgram()

	checkParserErrors(t, parser)
	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	} else if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statements. Got %d", len(program.Statements))
	}

	statement, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Expected expression statement, got %T", program.Statements[0])
	}

	ident, ok := statement.Expression.(*ast.Identifier)
	if !ok {
		t.Fatalf("Expected identifier, got %T", statement.Expression)
	}

	if ident.TokenLiteral() != "foobar" {
		t.Fatalf("Expected identifier %s, got %s", "foobar", ident.TokenLiteral())
	}
}

func TestIntegerLiteralExpression(t *testing.T) {
	input := "5;"

	lexer := lexer.New(input)
	parser := New(lexer)
	program := parser.ParseProgram()

	checkParserErrors(t, parser)
	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	} else if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statements. Got %d", len(program.Statements))
	}

	statement, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Expected expression statement, got %T", program.Statements[0])
	}

	literal, ok := statement.Expression.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("Expected IntegerLiteral, got %T", statement.Expression)
	}

	if literal.TokenLiteral() != "5" || literal.Value != 5 {
		t.Fatalf("Expected integer %s, got %s", "foobar", literal.TokenLiteral())
	}
}

func TestPrefixExpression(t *testing.T) {
	prefixTests := []struct {
		input        string
		operator     string
		integerValue int64
	}{
		{"!5;", "!", 5},
		{"-15;", "-", 15},
	}

	for _, test := range prefixTests {
		lexer := lexer.New(test.input)
		parser := New(lexer)
		program := parser.ParseProgram()
		checkParserErrors(t, parser)
		if program == nil {
			t.Fatalf("ParseProgram() returned nil")
		} else if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain 1 statements. Got %d", len(program.Statements))
		}

		statement, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("Expected expression statement, got %T", program.Statements[0])
		}

		prefix, ok := statement.Expression.(*ast.PrefixExpression)
		if !ok {
			t.Fatalf("Expected PrefixExpression, got %T", statement.Expression)
		}
		if prefix.Operator != test.operator {
			t.Fatalf("Expected operator %s, got %s", test.operator, prefix.Operator)
		}
		if !testIntegerLiteral(t, prefix.Right, test.integerValue) {
			return
		}
	}
}

func TestParsingInfixExpressions(t *testing.T) {
	infixTests := []struct {
		input      string
		leftValue  int64
		operator   string
		rightValue int64
	}{
		{"5 + 5;", 5, "+", 5},
		{"5 - 5;", 5, "-", 5},
		{"5 * 5;", 5, "*", 5},
		{"5 / 5;", 5, "/", 5},
		{"5 > 5;", 5, ">", 5},
		{"5 < 5;", 5, "<", 5},
		{"5 == 5;", 5, "==", 5},
		{"5 != 5;", 5, "!=", 5},
	}

	for _, test := range infixTests {
		lexer := lexer.New(test.input)
		parser := New(lexer)
		program := parser.ParseProgram()
		checkParserErrors(t, parser)
		if program == nil {
			t.Fatalf("ParseProgram() returned nil")
		} else if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain 1 statements. Got %d", len(program.Statements))
		}

		statement, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("Expected expression statement, got %T", program.Statements[0])
		}

		expr, ok := statement.Expression.(*ast.InfixExpression)
		if !ok {
			t.Fatalf("Expected InfixExpression, got %T", program.Statements[0])
		}

		if !testIntegerLiteral(t, expr.Left, test.leftValue) {
			return
		}
		if expr.Operator != test.operator {
			t.Fatalf("Expected operator %s, got %s", test.operator, expr.Operator)
		}
		if !testIntegerLiteral(t, expr.Right, test.rightValue) {
			return
		}
	}
}

func testIntegerLiteral(t *testing.T, expr ast.Expression, value int64) bool {
	literal, ok := expr.(*ast.IntegerLiteral)
	if !ok {
		t.Errorf("Expected IntegerLiteral, got %T", expr)
		return false
	}

	if literal.Value != value {
		t.Errorf("Expected integer %v, got %v", value, literal.Value)
		return false
	}
	return true
}
