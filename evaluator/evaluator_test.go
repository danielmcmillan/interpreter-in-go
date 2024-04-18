package evaluator

import (
	"testing"

	"danielmcm.com/interpreterbook/lexer"
	"danielmcm.com/interpreterbook/object"
	"danielmcm.com/interpreterbook/parser"
)

func TestEvalIntegerExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"5", 5},
		{"10", 10},
		{"-5", -5},
		{"-10", -10},
	}

	for _, test := range tests {
		result, ok := testEval(t, test.input)
		if ok {
			testIntegerObject(t, result, test.expected)
		}
	}
}

func TestEvalBooleanExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
	}

	for _, test := range tests {
		result, ok := testEval(t, test.input)
		if ok {
			testBooleanObject(t, result, test.expected)
		}
	}
}

func TestEvalBangOperator(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"!true", false},
		{"!false", true},
		{"!!false", false},
		{"!5", false},
		{"!!5", true},
	}

	for _, test := range tests {
		result, ok := testEval(t, test.input)
		if ok {
			testBooleanObject(t, result, test.expected)
		}
	}
}

func testEval(t *testing.T, input string) (object.Object, bool) {
	lexer := lexer.New(input)
	parser := parser.New(lexer)
	program := parser.ParseProgram()

	obj, err := Eval(program)
	if err != nil {
		t.Errorf("Eval failed: %s", err)
		return nil, false
	}
	return obj, true
}

func testIntegerObject(t *testing.T, obj object.Object, expected int64) bool {
	intObj, ok := obj.(*object.Integer)
	if !ok {
		t.Errorf("Expected Integer object, got %T (%+v)", obj, obj)
		return false
	}
	if intObj.Value != expected {
		t.Errorf("Expected integer value %v, got %v", expected, intObj.Value)
		return false
	}
	return true
}

func testBooleanObject(t *testing.T, obj object.Object, expected bool) bool {
	boolObj, ok := obj.(*object.Boolean)
	if !ok {
		t.Errorf("Expected Boolean object, got %T (%+v)", obj, obj)
		return false
	}
	if boolObj.Value != expected {
		t.Errorf("Expected boolean %v, got %v", expected, boolObj.Value)
		return false
	}
	return true
}
