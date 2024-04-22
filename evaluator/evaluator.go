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
	case *ast.InfixExpression:
		return evalInfixExpression(node.Operator, node.Left, node.Right)
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

func evalPrefixExpression(operator string, right ast.Expression) (object.Object, error) {
	operand, err := Eval(right)
	if err != nil {
		return nil, err
	}

	var result object.Object = nil
	ok := false
	switch operator {
	case "!":
		result, ok = evalBangOperatorExpression(operand)
	case "-":
		result, ok = evalMinusPrefixOperatorExpression(operand)
	}
	if !ok {
		return nil, EvalError{Message: fmt.Sprintf("operator %s not supported on %s (%s %s)", operator, right.String(), operand.Type(), operand.Inspect())}
	}
	return result, nil
}

func evalBangOperatorExpression(operand object.Object) (object.Object, bool) {
	switch value := operand.(type) {
	case *object.Boolean:
		return boolObjFromNativeBool(!value.Value), true
	case *object.Integer:
		return FALSE, true
	case *object.Null:
		return TRUE, true
	}
	return nil, false
}

func evalMinusPrefixOperatorExpression(operand object.Object) (object.Object, bool) {
	intObj, ok := operand.(*object.Integer)
	if !ok {
		return nil, false
	}
	return &object.Integer{Value: -intObj.Value}, true
}

func evalInfixExpression(operator string, left ast.Expression, right ast.Expression) (object.Object, error) {
	leftOperand, err := Eval(left)
	if err != nil {
		return nil, err
	}
	rightOperand, err := Eval(right)
	if err != nil {
		return nil, err
	}
	var result object.Object = nil
	ok := false

	if leftOperand.Type() == object.INTEGER_OBJ && rightOperand.Type() == object.INTEGER_OBJ {
		result, ok, err = evalIntegerInfixExpression(operator, leftOperand, rightOperand)
		if err != nil {
			return nil, err
		}
	} else if leftOperand.Type() == object.BOOLEAN_OBJ && rightOperand.Type() == object.BOOLEAN_OBJ {
		result, ok = evalBooleanInfixExpression(operator, leftOperand, rightOperand)
	}
	if !ok {
		return nil, EvalError{Message: fmt.Sprintf("operator %s not supported on %s (%s %s) and %s (%s %s)", operator, left.String(), leftOperand.Type(), leftOperand.Inspect(), right.String(), rightOperand.Type(), rightOperand.Inspect())}
	}
	return result, nil
}

func evalIntegerInfixExpression(operator string, left object.Object, right object.Object) (object.Object, bool, error) {
	leftInt, leftOk := left.(*object.Integer)
	rightInt, rightOk := right.(*object.Integer)
	if !leftOk || !rightOk {
		return nil, false, nil
	}
	switch operator {
	case "+":
		return &object.Integer{Value: leftInt.Value + rightInt.Value}, true, nil
	case "-":
		return &object.Integer{Value: leftInt.Value - rightInt.Value}, true, nil
	case "*":
		return &object.Integer{Value: leftInt.Value * rightInt.Value}, true, nil
	case "/":
		if rightInt.Value == 0 {
			return nil, true, fmt.Errorf("cannot divide by 0")
		}
		return &object.Integer{Value: leftInt.Value / rightInt.Value}, true, nil
	case "<":
		return boolObjFromNativeBool(leftInt.Value < rightInt.Value), true, nil
	case ">":
		return boolObjFromNativeBool(leftInt.Value > rightInt.Value), true, nil
	case "==":
		return boolObjFromNativeBool(leftInt.Value == rightInt.Value), true, nil
	case "!=":
		return boolObjFromNativeBool(leftInt.Value != rightInt.Value), true, nil
	default:
		return nil, false, nil
	}
}

func evalBooleanInfixExpression(operator string, left object.Object, right object.Object) (object.Object, bool) {
	switch operator {
	case "==":
		return boolObjFromNativeBool(left == right), true
	case "!=":
		return boolObjFromNativeBool(left != right), true
	default:
		return nil, false
	}
}
