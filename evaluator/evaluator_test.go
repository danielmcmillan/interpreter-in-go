package evaluator

import (
	"strings"
	"testing"

	"danielmcm.com/interpreterbook/lexer"
	"danielmcm.com/interpreterbook/object"
	"danielmcm.com/interpreterbook/parser"
)

func TestEvalEmptyProgram(t *testing.T) {
	result, ok := testEval(t, "")
	if ok && result != NULL {
		t.Fatalf("expected empty program to evaluate to NULL, got %v", result)
	}
}

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
		{`"a"=="a"`, true},
		{`"a"!="a"`, false},
		{`"a"=="b"`, false},
		{`"a"!="b"`, true},
		{`"a"<"b"`, true},
		{`"a">"b"`, false},
	}

	for _, test := range tests {
		result, ok := testEval(t, test.input)
		if ok {
			testBooleanObject(t, result, test.expected)
		}
	}
}

func TestEvalStringExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello"`, "hello"},
		{`"hello" + " " + "world"`, "hello world"},
	}

	for _, test := range tests {
		result, ok := testEval(t, test.input)
		if ok {
			testStringObject(t, result, test.expected)
		}
	}
}

func TestEvalArrayExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected []interface{}
	}{
		{`["hello", 1, true]`, []interface{}{"hello", 1, true}},
		{`[]`, []interface{}{}},
	}

	for _, test := range tests {
		result, ok := testEval(t, test.input)
		if !ok {
			continue
		}
		testArrayObject(t, result, test.expected)
	}
}

func TestEvalHashExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected map[interface{}]interface{}
	}{
		{`let t = true; {"a": 1, 2: "b", t: {4: [5]}}`, map[interface{}]interface{}{
			"a": 1,
			2:   "b",
			true: map[interface{}]interface{}{
				4: []interface{}{5},
			},
		}},
		{`{}`, map[interface{}]interface{}{}},
	}

	for _, test := range tests {
		result, ok := testEval(t, test.input)
		if !ok {
			continue
		}
		testHashObject(t, result, test.expected)
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

func TestReturnStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"return 5", 5},
		{"1; 2; return 3; return 4; 5", 3},
		{"1; if (true) { if (5) { 1; return 2; }; 3; } return 4;", 2},
		{"1; if (false) { return 1 } else { return 2 } return 3;", 2},
		{"(fn() {1; (fn() { return 2; })(); return 5; 6;})();", 5},
	}

	for _, test := range tests {
		result, ok := testEval(t, test.input)
		if ok {
			testIntegerObject(t, result, test.expected)
		}
	}
}

func TestLetStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let a = 5; a", 5},
		{"let a = 2; let b = a*3; a*b;", 12},
		{"let a = 2; let b = a == 2; if(b) {a} else {0}", 2},
	}

	for _, test := range tests {
		result, ok := testEval(t, test.input)
		if ok {
			testIntegerObject(t, result, test.expected)
		}
	}
}

func TestFunctionObject(t *testing.T) {
	input := "fn(x) { x + 2 }"
	result, ok := testEval(t, input)
	if !ok {
		return
	}
	fn, ok := result.(*object.Function)
	if !ok {
		t.Fatalf("expected function object, got %T (%+v)", fn, fn)
	}
	if len(fn.Parameters) != 1 || fn.Parameters[0] != "x" {
		t.Fatalf("expected 1 parameter x, got %+v", fn.Parameters)
	}
	expectedBody := "{ (x + 2); }"
	if fn.Body.String() != expectedBody {
		t.Fatalf("expected function body %q, got %q", expectedBody, fn.Body.String())
	}
}

func TestFunctionCall(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"(fn() {5})()", 5},
		{"let id = fn(x) { x }; id(5);", 5},
		{"let id = fn(x) { return x }; id(5);", 5},
		{"(fn(x) {x() * 2})(fn() {3})", 6},
		{"let mul = fn(x, y) { x * y }; mul(mul(mul(1, 2), 3), mul(2, 5));", 60},
		{"(fn() {let x = 5; fn() {x}})()()", 5},
		{"let adder = fn(x){fn(y){x+y}}; let aa = adder(3); let ab = adder(5); aa(-3) + ab(-5)", 0},
	}
	for _, test := range tests {
		result, ok := testEval(t, test.input)
		if ok {
			testIntegerObject(t, result, test.expected)
		}
	}
}

func TestArrayIndex(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"[123][0]", 123},
		{"[][0]", nil},
		{"let a = [1,2,3]; a[10/5]", 3},
		{"let i = 2/2; [1, [2, 3]][i]", []interface{}{2, 3}},
	}
	for _, test := range tests {
		result, ok := testEval(t, test.input)
		if ok {
			testObject(t, result, test.expected)
		}
	}
}

func TestHashIndex(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"{0: 123}[0]", 123},
		{"{}[0]", nil},
		{"let a = {\"x\": \"y\"}; a[\"x\"]", "y"},
		{"{false: 9}[false]", 9},
	}
	for _, test := range tests {
		result, ok := testEval(t, test.input)
		if ok {
			testObject(t, result, test.expected)
		}
	}
}

func TestBuiltinFunctions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`len("")`, 0},
		{`len("hello world")`, 11},
		{`len([1, 2, 3])`, 3},
		{`first([1, 2, 3])`, 1},
		{`first([])`, nil},
		{`rest([1, 2, 3])`, []interface{}{2, 3}},
		{`rest([])`, nil},
		{`last([1, 2, 3])`, 3},
		{`last([])`, nil},
		{`push([1, 2], 3)`, []interface{}{1, 2, 3}},
		{`push([], 2)`, []interface{}{2}},
		{`let x = [1]; push(x, 2); x`, []interface{}{1}},
		{`puts("hey")`, nil},
	}
	for _, test := range tests {
		result, err := runEval(test.input)
		if err != nil {
			t.Errorf("Eval failed: %s", err)
		} else {
			testObject(t, result, test.expected)
		}
	}
}

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		input   string
		pattern string
	}{
		{"true-false", "- not supported"},
		{"5+true; 5-false", "+ not supported"},
		{"-true", "- not supported"},
		{"foobar", "identifier not found: foobar"},
		{"true()", "not a function: true"},
		{"(fn() {0})(5)", "called with 1 argument"},
		{"(fn(x) {x})()", "called with 0 argument"},
		{`"5"-"4"`, "- not supported"},
		{`"5"+4`, "+ not supported"},
		{`len(1)`, "not supported"},
		{`len("a", "b")`, "number of arguments"},
		{`{fn(){}: 0}`, "hash key must be string, integer or boolean"},
		{`5["x"]`, "not an array or hash: 5"},
		{`"a"[true]`, "not an array or hash: \"a\""},
		{`true[0]`, "not an array or hash: true"},
		{`[5][true]`, "array index must be an integer"},
		{`{1:2}[fn(){}]`, "hash index must be string, integer or boolean"},
	}

	for _, test := range tests {
		result, err := runEval(test.input)
		if err == nil || !strings.Contains(err.Error(), test.pattern) {
			t.Errorf("expected Eval(\"%v\") to return error matching \"%v\", got \"%v\", result = %+v\n", test.input, test.pattern, err, result)
		}
	}
}

func runEval(input string) (object.Object, error) {
	lexer := lexer.New(input)
	parser := parser.New(lexer)
	program := parser.ParseProgram()
	return Eval(program, object.NewEnvironment())
}

func testEval(t *testing.T, input string) (object.Object, bool) {
	obj, err := runEval(input)
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

func testStringObject(t *testing.T, obj object.Object, expected string) bool {
	strObj, ok := obj.(*object.String)
	if !ok {
		t.Errorf("Expected String object, got %T (%+v)", obj, obj)
		return false
	}
	if strObj.Value != expected {
		t.Errorf("Expected boolean %v, got %v", expected, strObj.Value)
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

func testObject(t *testing.T, obj object.Object, expected interface{}) bool {
	switch expected := expected.(type) {
	case int:
		return testIntegerObject(t, obj, int64(expected))
	case bool:
		return testBooleanObject(t, obj, expected)
	case string:
		return testStringObject(t, obj, expected)
	case []interface{}:
		return testArrayObject(t, obj, expected)
	case map[interface{}]interface{}:
		return testHashObject(t, obj, expected)
	default:
		return false
	}
}

func testArrayObject(t *testing.T, obj object.Object, expected []interface{}) bool {
	arr, ok := obj.(*object.Array)
	if !ok {
		t.Errorf("expected Array object, got %s", obj.Type())
		return false
	}
	if len(arr.Elements) != len(expected) {
		t.Errorf("expected array to have %d elements, got %d", len(expected), len(arr.Elements))
		return false
	}
	for i, elem := range arr.Elements {
		testObject(t, elem, expected[i])
		return false
	}
	return true
}

func testHashObject(t *testing.T, obj object.Object, expected map[interface{}]interface{}) bool {
	hash, ok := obj.(*object.Hash)
	if !ok {
		t.Errorf("expected Hash object, got %s", obj.Type())
		return false
	}
	if len(hash.Entries) != len(expected) {
		t.Errorf("expected hash to have %d entries, got %d", len(expected), len(hash.Entries))
		return false
	}
	for expectedKey, expectedVal := range expected {
		var key object.HashKey
		switch expectedKey := expectedKey.(type) {
		case string:
			key = object.HashKeyFromString(expectedKey)
		case int:
			key = object.HashKeyFromInt(int64(expectedKey))
		case bool:
			key = object.HashKeyFromBool(expectedKey)
		default:
			t.Fatalf("invalid expected hash key %v", expectedKey)
			return false
		}

		val, ok := hash.Entries[key]
		if !ok {
			t.Errorf("hash missing expected key %v", expectedKey)
			return false
		}
		if !testObject(t, val, expectedVal) {
			return false
		}
	}
	return true
}
