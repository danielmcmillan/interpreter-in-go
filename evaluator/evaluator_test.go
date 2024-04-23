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
		{"2+3", 5},
		{"2-3", -1},
		{"2*3", 6},
		{"6/3", 2},
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
		{"2<5", true},
		{"5<2", false},
		{"2>5", false},
		{"5>2", true},
		{"2==2", true},
		{"2==3", false},
		{"2!=3", true},
		{"2!=2", false},
		{"true==true", true},
		{"true==false", false},
		{"true!=false", true},
		{"true!=true", false},
	}

	for _, test := range tests {
		result, ok := testEval(t, test.input)
		if ok {
			testBooleanObject(t, result, test.expected)
		}
	}
}

func TestEvalIfExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"if (true) {10}", 10},
		{"if (false) {10}", nil},
		{"if (0) {10}", 10},
		{"if (5>4) {10} else {9}", 10},
		{"if (5<4) {10} else {9}", 9},
	}

	for _, test := range tests {
		result, ok := testEval(t, test.input)
		if ok {
			if integer, ok := test.expected.(int); ok {
				testIntegerObject(t, result, int64(integer))
			} else {
				testNullObject(t, result)
			}
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

func testNullObject(t *testing.T, obj object.Object) bool {
	if obj != NULL {
		t.Errorf("Expected Null object, got %T (%+v)", obj, obj)
		return false
	}
	return true
}
