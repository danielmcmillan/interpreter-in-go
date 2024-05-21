package parser

import (
	"strings"
	"testing"

	"danielmcm.com/interpreterbook/ast"
	"danielmcm.com/interpreterbook/lexer"
)

func TestEmptyProgram(t *testing.T) {
	lexer := lexer.New("")
	parser := New(lexer)

	program := parser.ParseProgram()
	checkParserErrors(t, parser)
	checkProgramLen(t, program, 0)
}

func TestLetStatements(t *testing.T) {
	input := `
		let x = 5;
		let y = true;
		let foobar = y;
	`

	lexer := lexer.New(input)
	parser := New(lexer)

	program := parser.ParseProgram()
	checkParserErrors(t, parser)
	checkProgramLen(t, program, 3)

	tests := []struct {
		expectedIdentifier string
		expectedLiteral    interface{}
	}{
		{"x", 5},
		{"y", true},
		{"foobar", "y"},
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
		if !testLiteralExpression(t, letStatement.Value, test.expectedLiteral) {
			break
		}
	}
}

func TestReturnStatements(t *testing.T) {
	input := `
		return 5;
		return true;
		return x;
	`

	lexer := lexer.New(input)
	parser := New(lexer)

	program := parser.ParseProgram()
	checkParserErrors(t, parser)
	checkProgramLen(t, program, 3)

	tests := []struct {
		expectedReturnValue interface{}
	}{
		{5},
		{true},
		{"x"},
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

		if !testLiteralExpression(t, returnStatement.ReturnValue, test.expectedReturnValue) {
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

func TestStringLiteralExpression(t *testing.T) {
	input := `"hello"`

	lexer := lexer.New(input)
	parser := New(lexer)
	program := parser.ParseProgram()

	checkParserErrors(t, parser)
	checkProgramLen(t, program, 1)

	statement, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Expected expression statement, got %T", program.Statements[0])
	}

	testStringLiteral(t, statement.Expression, "hello")
}

func TestArrayExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected []interface{}
	}{
		{`[1, two, true]`, []interface{}{1, "two", true}},
		{`[]`, []interface{}{}},
		{`[1 + 1, 1 + 2]`, []interface{}{nil, nil}},
	}

	for _, test := range tests {
		lexer := lexer.New(test.input)
		parser := New(lexer)
		program := parser.ParseProgram()
		checkParserErrors(t, parser)
		checkProgramLen(t, program, 1)

		statement, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("Expected expression statement, got %T", program.Statements[0])
		}

		arr, ok := statement.Expression.(*ast.ArrayExpression)
		if !ok {
			t.Fatalf("Expected ArrayExpression, got %T", statement.Expression)
		}

		if len(arr.Elements) != len(test.expected) {
			t.Fatalf("Expected %d elements, got %d", len(test.expected), len(arr.Elements))
		}

		for i, elem := range arr.Elements {
			if test.expected[i] != nil && !testLiteralExpression(t, elem, test.expected[i]) {
				return
			}
		}
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
		{"-a * b", "((-a) * b);\n"},
		{"!-a", "(!(-a));\n"},
		{"a + b + c", "((a + b) + c);\n"},
		{"a + b - c", "((a + b) - c);\n"},
		{"a * b * c", "((a * b) * c);\n"},
		{"a * b / c", "((a * b) / c);\n"},
		{"a + b / c", "(a + (b / c));\n"},
		{"a + b * c + d / e - f", "(((a + (b * c)) + (d / e)) - f);\n"},
		{"3 + 4;\n -5 * 5", "(3 + 4);\n((-5) * 5);\n"},
		{"5 > 4 == 3 < 4", "((5 > 4) == (3 < 4));\n"},
		{"5 < 4 != 3 > 4", "((5 < 4) != (3 > 4));\n"},
		{"3 + 4 * 5 == 3 * 1 + 4 * 5", "((3 + (4 * 5)) == ((3 * 1) + (4 * 5)));\n"},
		{"true", "true;\n"},
		{"false", "false;\n"},
		{"3 > 5 == false", "((3 > 5) == false);\n"},
		{"3 < 5 == true", "((3 < 5) == true);\n"},
		{"1 + (2 + 3) + 4", "((1 + (2 + 3)) + 4);\n"},
		{"(5 + 5) * 2", "((5 + 5) * 2);\n"},
		{"2 / (5 + 5)", "(2 / (5 + 5));\n"},
		{"-(5+5)", "(-(5 + 5));\n"},
		{"!(true == true)", "(!(true == true));\n"},
		{"a * add(b + c, d - e) * f", "((a * add((b + c), (d - e))) * f);\n"},
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

func TestIfExpression(t *testing.T) {
	tests := []struct {
		input       string
		alternative string
	}{
		{`if (x < y) { x }`, ""},
		{`if (x < y) { x } else { y }`, "y"},
	}

	for _, test := range tests {
		lexer := lexer.New(test.input)
		parser := New(lexer)
		program := parser.ParseProgram()
		checkParserErrors(t, parser)
		checkProgramLen(t, program, 1)

		statement, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("Expected expression statement, got %T", program.Statements[0])
		}

		ifExpr, ok := statement.Expression.(*ast.IfExpression)
		if !ok {
			t.Fatalf("Expected IfExpression, got %T", statement.Expression)
		}

		if !testInfixExpression(t, ifExpr.Condition, "x", "<", "y") {
			break
		}
		if len(ifExpr.Consequence.Statements) != 1 {
			t.Fatalf("Expected Consequence to have 1 statement, got %d", len(ifExpr.Consequence.Statements))
		}

		consequence, ok := ifExpr.Consequence.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("Expected Consequence to have expression statement, got %T", ifExpr.Consequence.Statements[0])
		}

		if !testLiteralExpression(t, consequence.Expression, "x") {
			break
		}

		if len(test.alternative) == 0 {
			if ifExpr.Alternative != nil {
				t.Fatalf("Expected Alternative to be unspecified, got %+v", ifExpr.Alternative)
			}
		} else {
			alternative, ok := ifExpr.Alternative.Statements[0].(*ast.ExpressionStatement)
			if !ok {
				t.Fatalf("Expected Alternative to have expression statement, got %T", ifExpr.Consequence.Statements[0])
			}

			if !testLiteralExpression(t, alternative.Expression, test.alternative) {
				break
			}
		}
	}
}

func TestFunctionLiteral(t *testing.T) {
	tests := []struct {
		input  string
		params []string
	}{
		{`fn() { x + y }`, []string{}},
		{`fn(x) { x + y }`, []string{"x"}},
		{`fn(x, y) { x + y }`, []string{"x", "y"}},
	}

	for _, test := range tests {
		lexer := lexer.New(test.input)
		parser := New(lexer)
		program := parser.ParseProgram()
		checkParserErrors(t, parser)
		checkProgramLen(t, program, 1)

		statement, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("Expected expression statement, got %T", program.Statements[0])
		}

		fnLit, ok := statement.Expression.(*ast.FunctionLiteral)
		if !ok {
			t.Fatalf("Expected FunctionLiteral, got %T", statement.Expression)
		}

		if len(fnLit.Parameters) != len(test.params) {
			t.Fatalf("Expected %d parameters, got %d", len(test.params), len(fnLit.Parameters))
		}

		for i, param := range fnLit.Parameters {
			if !testIdentifier(t, &param, test.params[i]) {
				return
			}
		}
		if len(fnLit.Body.Statements) != 1 {
			t.Fatalf("Expected function body to have 1 statement, got %d", len(fnLit.Body.Statements))
		}

		body, ok := fnLit.Body.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("Expected function body to have expression statement, got %T", fnLit.Body.Statements[0])
		}

		if !testInfixExpression(t, body.Expression, "x", "+", "y") {
			break
		}
	}
}

func TestCallExpression(t *testing.T) {
	tests := []struct {
		input string
		fn    string
		args  [](func(*testing.T, ast.Expression) bool)
	}{
		{"f()", "f", [](func(*testing.T, ast.Expression) bool){}},
		{"f(x + y)", "f", [](func(*testing.T, ast.Expression) bool){func(t *testing.T, arg ast.Expression) bool { return testInfixExpression(t, arg, "x", "+", "y") }}},
		{"f(x, y)", "f", [](func(*testing.T, ast.Expression) bool){
			func(t *testing.T, arg ast.Expression) bool { return testLiteralExpression(t, arg, "x") },
			func(t *testing.T, arg ast.Expression) bool { return testLiteralExpression(t, arg, "y") },
		}},
	}

	for _, test := range tests {
		lexer := lexer.New(test.input)
		parser := New(lexer)
		program := parser.ParseProgram()
		checkParserErrors(t, parser)
		checkProgramLen(t, program, 1)

		statement, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("Expected expression statement, got %T", program.Statements[0])
		}

		expr, ok := statement.Expression.(*ast.CallExpression)
		if !ok {
			t.Fatalf("Expected CallExpression, got %T", statement.Expression)
		}

		if !testLiteralExpression(t, expr.Function, test.fn) {
			return
		}

		if len(expr.Arguments) != len(test.args) {
			t.Fatalf("Expected %d parameters, got %d", len(test.args), len(expr.Arguments))
		}

		for i, arg := range expr.Arguments {
			if !test.args[i](t, arg) {
				return
			}
		}
	}
}

func TestIndexExpression(t *testing.T) {
	tests := []struct {
		input string
		array string
		index interface{}
	}{
		{"a[1]", "a", 1},
		{"b[x]", "b", "x"},
	}

	for _, test := range tests {
		lexer := lexer.New(test.input)
		parser := New(lexer)
		program := parser.ParseProgram()
		checkParserErrors(t, parser)
		checkProgramLen(t, program, 1)

		statement, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("Expected expression statement, got %T", program.Statements[0])
		}

		expr, ok := statement.Expression.(*ast.IndexExpression)
		if !ok {
			t.Fatalf("Expected IndexExpression, got %T", statement.Expression)
		}

		if !testLiteralExpression(t, expr.Array, test.array) {
			return
		}
		if !testLiteralExpression(t, expr.Index, test.index) {
			return
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

func testStringLiteral(t *testing.T, expr ast.Expression, value string) bool {
	literal, ok := expr.(*ast.StringLiteral)
	if !ok {
		t.Errorf("Expected StringLiteral, got %T", expr)
		return false
	}

	if literal.Value != value {
		t.Errorf("Expected string %v, got %v", value, literal.Value)
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
