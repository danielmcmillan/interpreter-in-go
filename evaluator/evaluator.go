package evaluator

import (
	"fmt"

	"danielmcm.com/interpreterbook/ast"
	"danielmcm.com/interpreterbook/object"
)

var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

type EvalError struct {
	Message string
}

func (evalError EvalError) Error() string {
	return evalError.Message
}

func Eval(node ast.Node) (object.Object, error) {
	switch node := node.(type) {
	// Statements
	case *ast.Program:
		return evalStatements(node.Statements)
	case *ast.ExpressionStatement:
		return Eval(node.Expression)
	// Expressions
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}, nil
	case *ast.BooleanLiteral:
		return boolObjFromNativeBool(node.Value), nil
	case *ast.PrefixExpression:
		return evalPrefixExpression(node.Operator, node.Right)
	}
	return nil, EvalError{Message: fmt.Sprintf(`Can't eval node type %T (token "%s")`, node, node.TokenLiteral())}
}

func evalStatements(statements []ast.Statement) (object.Object, error) {
	var result object.Object
	for _, statement := range statements {
		obj, err := Eval(statement)
		if err != nil {
			return nil, err
		}
		result = obj
	}
	return result, nil
}

func boolObjFromNativeBool(value bool) *object.Boolean {
	if value {
		return TRUE
	} else {
		return FALSE
	}
}

func evalPrefixExpression(operator string, operand ast.Expression) (object.Object, error) {
	value, err := Eval(operand)
	if err != nil {
		return nil, err
	}

	var result object.Object = nil
	ok := false
	switch operator {
	case "!":
		result, ok = evalBangOperatorExpression(value)
	case "-":
		result, ok = evalMinusPrefixOperatorExpression(value)
	}
	if !ok {
		return nil, EvalError{Message: fmt.Sprintf("prefix operator %s not supported on %s (%s %s)", operator, operand.String(), value.Type(), value.Inspect())}
	}
	return result, nil
}

func evalBangOperatorExpression(value object.Object) (object.Object, bool) {
	switch value := value.(type) {
	case *object.Boolean:
		return boolObjFromNativeBool(!value.Value), true
	case *object.Integer:
		return FALSE, true
	case *object.Null:
		return TRUE, true
	}
	return nil, false
}

func evalMinusPrefixOperatorExpression(value object.Object) (object.Object, bool) {
	intObj, ok := value.(*object.Integer)
	if !ok {
		return nil, false
	}
	return &object.Integer{Value: -intObj.Value}, true
}
