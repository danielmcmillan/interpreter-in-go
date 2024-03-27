package parser

import (
	"strings"
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

	program := parser.ParseProgram()
	checkParserErrors(t, parser)
	checkProgramLen(t, program, 3)

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

		if !testIdentifier(t, letStatement.Name, test.expectedIdentifier) {
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
	checkProgramLen(t, program, 3)

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
	checkProgramLen(t, program, 1)

	statement, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Expected expression statement, got %T", program.Statements[0])
	}

	testIdentifier(t, statement.Expression, "foobar")
}

func TestIntegerLiteralExpression(t *testing.T) {
	input := "5;"

	lexer := lexer.New(input)
	parser := New(lexer)
	program := parser.ParseProgram()

	checkParserErrors(t, parser)
	checkProgramLen(t, program, 1)

	statement, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Expected expression statement, got %T", program.Statements[0])
	}

	testIntegerLiteral(t, statement.Expression, 5)
}

func TestBooleanLiteralExpression(t *testing.T) {
	input := "true;false;"

	lexer := lexer.New(input)
	parser := New(lexer)
	program := parser.ParseProgram()

	checkParserErrors(t, parser)
	checkProgramLen(t, program, 2)

	for i, value := range []bool{true, false} {
		statement, ok := program.Statements[i].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("Expected expression statement, got %T", program.Statements[0])
		}

		testBoolean(t, statement.Expression, value)
	}
}

func TestParsingPrefixExpression(t *testing.T) {
	prefixTests := []struct {
		input    string
		operator string
		value    interface{}
	}{
		{"!5;", "!", 5},
		{"-15;", "-", 15},
		{"!false", "!", false},
	}

	for _, test := range prefixTests {
		lexer := lexer.New(test.input)
		parser := New(lexer)
		program := parser.ParseProgram()
		checkParserErrors(t, parser)
		checkProgramLen(t, program, 1)

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
		if !testLiteralExpression(t, prefix.Right, test.value) {
			return
		}
	}
}

func TestParsingInfixExpressions(t *testing.T) {
	infixTests := []struct {
		input      string
		leftValue  interface{}
		operator   string
		rightValue interface{}
	}{
		{"5 + 5;", 5, "+", 5},
		{"5 - 5;", 5, "-", 5},
		{"5 * 5;", 5, "*", 5},
		{"5 / 5;", 5, "/", 5},
		{"5 > 5;", 5, ">", 5},
		{"5 < 5;", 5, "<", 5},
		{"5 == 5;", 5, "==", 5},
		{"5 != 5;", 5, "!=", 5},
		{"true == true", true, "==", true},
		{"false != true", false, "!=", true},
	}

	for _, test := range infixTests {
		lexer := lexer.New(test.input)
		parser := New(lexer)
		program := parser.ParseProgram()
		checkParserErrors(t, parser)
		checkProgramLen(t, program, 1)

		statement, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("Expected expression statement, got %T", program.Statements[0])
		}

		if !testInfixExpression(t, statement.Expression, test.leftValue, test.operator, test.rightValue) {
			break
		}
	}
}

func TestOperatorPrecedenceParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"-a * b", "((-a) * b)"},
		{"!-a", "(!(-a))"},
		{"a + b + c", "((a + b) + c)"},
		{"a + b - c", "((a + b) - c)"},
		{"a * b * c", "((a * b) * c)"},
		{"a * b / c", "((a * b) / c)"},
		{"a + b / c", "(a + (b / c))"},
		{"a + b * c + d / e - f", "(((a + (b * c)) + (d / e)) - f)"},
		{"3 + 4; -5 * 5", "(3 + 4)((-5) * 5)"},
		{"5 > 4 == 3 < 4", "((5 > 4) == (3 < 4))"},
		{"5 < 4 != 3 > 4", "((5 < 4) != (3 > 4))"},
		{"3 + 4 * 5 == 3 * 1 + 4 * 5", "((3 + (4 * 5)) == ((3 * 1) + (4 * 5)))"},
		{"true", "true"},
		{"false", "false"},
		{"3 > 5 == false", "((3 > 5) == false)"},
		{"3 < 5 == true", "((3 < 5) == true)"},
	}

	for _, test := range tests {
		lexer := lexer.New(test.input)
		parser := New(lexer)
		program := parser.ParseProgram()
		checkParserErrors(t, parser)
		checkProgram(t, program)
		actual := program.String()

		if test.expected != actual {
			t.Fatalf("Expected %s, got %s\n", test.expected, actual)
		}
	}
}

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

func checkProgram(t *testing.T, program *ast.Program) {
	if program == nil {
		t.Fatalf("program is nil")
	}
}

func checkProgramLen(t *testing.T, program *ast.Program, length int) {
	checkProgram(t, program)
	if len(program.Statements) != length {
		t.Fatalf("Expected program.Statements to have %d statements, got %d", length, len(program.Statements))
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

func testIdentifier(t *testing.T, expr ast.Expression, value string) bool {
	ident, ok := expr.(*ast.Identifier)
	if !ok {
		t.Errorf("Expected Identifier, got %T\n", expr)
		return false
	}

	if ident.Value != value {
		t.Errorf("Expected identifier %v, got %v", value, ident.Value)
		return false
	}
	return true
}

func testBoolean(t *testing.T, expr ast.Expression, value bool) bool {
	boolExpr, ok := expr.(*ast.BooleanLiteral)
	if !ok {
		t.Errorf("Expected BooleanLiteral, got %T\n", expr)
		return false
	}

	if boolExpr.Value != value {
		t.Errorf("Expected bool %v, got %v", value, boolExpr.Value)
		return false
	}
	return true
}

func testLiteralExpression(t *testing.T, expr ast.Expression, expected interface{}) bool {
	switch value := expected.(type) {
	case int:
		return testIntegerLiteral(t, expr, int64(value))
	case int64:
		return testIntegerLiteral(t, expr, value)
	case string:
		return testIdentifier(t, expr, value)
	case bool:
		return testBoolean(t, expr, value)
	default:
		return false
	}
}

func testInfixExpression(t *testing.T, expr ast.Expression, left interface{}, operator string, right interface{}) bool {
	infix, ok := expr.(*ast.InfixExpression)
	if !ok {
		t.Errorf("Expected InfixExpresion, got %T", expr)
		return false
	}

	if !testLiteralExpression(t, infix.Left, left) {
		return false
	}
	if infix.Operator != operator {
		t.Errorf("Expected operator %s, got %s", operator, infix.Operator)
		return false
	}
	if !testLiteralExpression(t, infix.Right, right) {
		return false
	}

	return true
}
