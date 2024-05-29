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

func Eval(node ast.Node, env *object.Environment) (object.Object, error) {
	switch node := node.(type) {
	// Statements
	case *ast.Program:
		return evalStatementsAndReturn(node.Statements, env)
	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)
	case *ast.BlockStatement:
		return evalStatements(node.Statements, env)
	case *ast.ReturnStatement:
		return evalReturnStatement(node, env)
	case *ast.LetStatement:
		return evalLetStatement(node, env)
	// Expressions
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}, nil
	case *ast.BooleanLiteral:
		return boolObjFromNativeBool(node.Value), nil
	case *ast.StringLiteral:
		return &object.String{Value: node.Value}, nil
	case *ast.Identifier:
		return evalIdentifier(node.Value, env)
	case *ast.PrefixExpression:
		return evalPrefixExpression(node.Operator, node.Right, env)
	case *ast.InfixExpression:
		return evalInfixExpression(node.Operator, node.Left, node.Right, env)
	case *ast.IfExpression:
		return evalIfExpression(node, env)
	case *ast.FunctionLiteral:
		return evalFunctionLiteral(node, env)
	case *ast.CallExpression:
		return evalCallExpression(node, env)
	case *ast.ArrayExpression:
		return evalArrayExpression(node, env)
	case *ast.IndexExpression:
		return evalIndexExpression(node, env)
	}
	return nil, fmt.Errorf(`can't eval node type %T (%s)`, node, node.String())
}

func evalStatementsAndReturn(statements []ast.Statement, env *object.Environment) (object.Object, error) {
	result, err := evalStatements(statements, env)
	if err != nil {
		return nil, err
	}
	if returnObj, ok := result.(*object.ReturnValue); ok {
		return returnObj.Value, nil
	} else {
		return result, nil
	}
}

func evalStatements(statements []ast.Statement, env *object.Environment) (object.Object, error) {
	var result object.Object = NULL
	for _, statement := range statements {
		obj, err := Eval(statement, env)
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

func evalReturnStatement(statement *ast.ReturnStatement, env *object.Environment) (object.Object, error) {
	val, err := Eval(statement.ReturnValue, env)
	if err != nil {
		return nil, err
	}
	return &object.ReturnValue{Value: val}, nil
}

func evalLetStatement(statement *ast.LetStatement, env *object.Environment) (object.Object, error) {
	val, err := Eval(statement.Value, env)
	if err != nil {
		return nil, err
	}
	env.Set(statement.Name.Value, val)
	return NULL, nil
}

func boolObjFromNativeBool(value bool) *object.Boolean {
	if value {
		return TRUE
	} else {
		return FALSE
	}
}

func evalIdentifier(ident string, env *object.Environment) (object.Object, error) {
	if val, ok := env.Get(ident); ok {
		return val, nil
	}
	if val, ok := builtins[ident]; ok {
		return val, nil
	}
	return nil, fmt.Errorf("identifier not found: %s", ident)
}

func evalPrefixExpression(operator string, right ast.Expression, env *object.Environment) (object.Object, error) {
	operand, err := Eval(right, env)
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

func evalInfixExpression(operator string, left ast.Expression, right ast.Expression, env *object.Environment) (object.Object, error) {
	leftOperand, err := Eval(left, env)
	if err != nil {
		return nil, err
	}
	rightOperand, err := Eval(right, env)
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
	} else if leftOperand.Type() == object.STRING_OBJ && rightOperand.Type() == object.STRING_OBJ {
		result, ok = evalStringInfixExpression(operator, leftOperand, rightOperand)
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

func evalStringInfixExpression(operator string, left object.Object, right object.Object) (object.Object, bool) {
	leftString, leftOk := left.(*object.String)
	rightString, rightOk := right.(*object.String)
	if !leftOk || !rightOk {
		return nil, false
	}
	switch operator {
	case "+":
		return &object.String{Value: leftString.Value + rightString.Value}, true
	case "<":
		return boolObjFromNativeBool(leftString.Value < rightString.Value), true
	case ">":
		return boolObjFromNativeBool(leftString.Value > rightString.Value), true
	case "==":
		return boolObjFromNativeBool(leftString.Value == rightString.Value), true
	case "!=":
		return boolObjFromNativeBool(leftString.Value != rightString.Value), true
	default:
		return nil, false
	}
}

func evalIfExpression(expr *ast.IfExpression, env *object.Environment) (object.Object, error) {
	cond, err := Eval(expr.Condition, env)
	if err != nil {
		return nil, err
	}
	if isTruthy(cond) {
		return Eval(expr.Consequence, env)
	} else if expr.Alternative != nil {
		return Eval(expr.Alternative, env)
	} else {
		return NULL, nil
	}
}

func evalFunctionLiteral(expr *ast.FunctionLiteral, env *object.Environment) (object.Object, error) {
	obj := &object.Function{
		Parameters: make([]string, len(expr.Parameters)),
		Body:       expr.Body,
		Env:        env,
	}
	for i, param := range expr.Parameters {
		obj.Parameters[i] = param.Value
	}
	return obj, nil
}

func evalExpressions(expressions []ast.Expression, env *object.Environment) ([]object.Object, error) {
	objects := make([]object.Object, len(expressions))
	for i, expr := range expressions {
		obj, err := Eval(expr, env)
		if err != nil {
			return nil, err
		}
		objects[i] = obj
	}
	return objects, nil
}

func evalCallExpression(expr *ast.CallExpression, env *object.Environment) (object.Object, error) {
	called, err := Eval(expr.Function, env)
	if err != nil {
		return nil, err
	}
	// Evaluate arguments
	args, err := evalExpressions(expr.Arguments, env)
	if err != nil {
		return nil, err
	}
	switch fn := called.(type) {
	case *object.Function:
		if len(fn.Parameters) != len(expr.Arguments) {
			return nil, fmt.Errorf("function with %d parameters called with %d arguments", len(fn.Parameters), len(expr.Arguments))
		}
		fnEnv := object.NewEnclosedEnvironment(fn.Env)
		for i, param := range fn.Parameters {
			fnEnv.Set(param, args[i])
		}
		return evalStatementsAndReturn(fn.Body.Statements, fnEnv)
	case *object.Builtin:
		return fn.Fn(args...)
	default:
		return nil, fmt.Errorf("not a function: %s", expr.Function.String())
	}
}

func evalArrayExpression(expr *ast.ArrayExpression, env *object.Environment) (object.Object, error) {
	elements, err := evalExpressions(expr.Elements, env)
	if err != nil {
		return nil, err
	}
	obj := &object.Array{
		Elements: elements,
	}
	return obj, nil
}

func evalIndexExpression(expr *ast.IndexExpression, env *object.Environment) (object.Object, error) {
	arrObj, err := Eval(expr.Array, env)
	if err != nil {
		return nil, err
	}
	array, ok := arrObj.(*object.Array)
	if !ok {
		return nil, fmt.Errorf("not an array: %s", expr.Array.String())
	}
	indexObj, err := Eval(expr.Index, env)
	if err != nil {
		return nil, err
	}
	index, ok := indexObj.(*object.Integer)
	if !ok {
		return nil, fmt.Errorf("index must be an integer: %s", expr.Index.String())
	}
	if index.Value < 0 || index.Value >= int64(len(array.Elements)) {
		return NULL, nil
	}
	return array.Elements[index.Value], nil
}
