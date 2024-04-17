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
