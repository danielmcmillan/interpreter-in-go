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

func Eval(node ast.Node) (object.Object, error) {
	switch node := node.(type) {
	// Statements
	case *ast.Program:
		return evalProgram(node)
	case *ast.ExpressionStatement:
		return Eval(node.Expression)
	case *ast.BlockStatement:
		return evalStatements(node.Statements)
	case *ast.ReturnStatement:
		return evalReturnStatement(node)
	// Expressions
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}, nil
	case *ast.BooleanLiteral:
		return boolObjFromNativeBool(node.Value), nil
	case *ast.PrefixExpression:
		return evalPrefixExpression(node.Operator, node.Right)
	case *ast.InfixExpression:
		return evalInfixExpression(node.Operator, node.Left, node.Right)
	case *ast.IfExpression:
		return evalIfExpression(node)
	}
	return nil, fmt.Errorf(`can't eval node type %T (token "%s")`, node, node.TokenLiteral())
}

func evalProgram(program *ast.Program) (object.Object, error) {
	result, err := evalStatements(program.Statements)
	if err != nil {
		return nil, err
	}
	if returnObj, ok := result.(*object.ReturnValue); ok {
		return returnObj.Value, nil
	} else {
		return result, err
	}
}

func evalStatements(statements []ast.Statement) (object.Object, error) {
	var result object.Object
	for _, statement := range statements {
		obj, err := Eval(statement)
		if err != nil {
			return nil, err
		}
		if obj.Type() == object.RETURN_VALUE_OBJ {
			return obj, nil
		}
		result = obj
	}
	return result, nil
}

func evalReturnStatement(statement *ast.ReturnStatement) (object.Object, error) {
	val, err := Eval(statement.ReturnValue)
	if err != nil {
		return nil, err
	}
	return &object.ReturnValue{Value: val}, nil
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
		return nil, fmt.Errorf("operator %s not supported on %s (%s %s)", operator, right.String(), operand.Type(), operand.Inspect())
	}
	return result, nil
}

func isTruthy(cond object.Object) bool {
	switch cond := cond.(type) {
	case *object.Boolean:
		return cond.Value
	case *object.Integer:
		return true
	case *object.Null:
		return false
	default:
		return false
	}
}

func evalBangOperatorExpression(operand object.Object) (object.Object, bool) {
	return boolObjFromNativeBool(!isTruthy(operand)), true
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
		return nil, fmt.Errorf("operator %s not supported on %s (%s %s) and %s (%s %s)", operator, left.String(), leftOperand.Type(), leftOperand.Inspect(), right.String(), rightOperand.Type(), rightOperand.Inspect())
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

func evalIfExpression(expr *ast.IfExpression) (object.Object, error) {
	cond, err := Eval(expr.Condition)
	if err != nil {
		return nil, err
	}
	if isTruthy(cond) {
		return Eval(expr.Consequence)
	} else if expr.Alternative != nil {
		return Eval(expr.Alternative)
	} else {
		return NULL, nil
	}
}
