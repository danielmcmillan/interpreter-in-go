package object

import (
	"bytes"
	"fmt"
	"strings"

	"danielmcm.com/interpreterbook/ast"
)

type ObjectType string

const (
	NULL_OBJ         = "NULL"
	INTEGER_OBJ      = "INTEGER"
	BOOLEAN_OBJ      = "BOOLEAN"
	STRING_OBJ       = "STRING"
	RETURN_VALUE_OBJ = "RETURN_VALUE"
	FUNCTION_OBJ     = "FUNCTION"
)

type Object interface {
	Type() ObjectType
	Inspect() string
}

type Environment struct {
	store map[string]Object
	outer *Environment
}

func NewEnvironment() *Environment {
	store := make(map[string]Object)
	return &Environment{store: store, outer: nil}
}

func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

func (env *Environment) Get(name string) (Object, bool) {
	val, ok := env.store[name]
	if !ok && env.outer != nil {
		return env.outer.Get(name)
	}
	return val, ok
}

func (env *Environment) Set(name string, val Object) Object {
	env.store[name] = val
	return val
}

type Null struct{}

func (null *Null) Type() ObjectType {
	return NULL_OBJ
}
func (null *Null) Inspect() string {
	return "null"
}

type Integer struct {
	Value int64
}

func (integer *Integer) Type() ObjectType {
	return INTEGER_OBJ
}
func (integer *Integer) Inspect() string {
	return fmt.Sprintf("%d", integer.Value)
}

type Boolean struct {
	Value bool
}

func (boolean *Boolean) Type() ObjectType {
	return BOOLEAN_OBJ
}
func (boolean *Boolean) Inspect() string {
	if boolean.Value {
		return "true"
	} else {
		return "false"
	}
}

type String struct {
	Value string
}

func (str *String) Type() ObjectType {
	return STRING_OBJ
}
func (str *String) Inspect() string {
	return str.Value
}

type ReturnValue struct {
	Value Object
}

func (returnValue *ReturnValue) Type() ObjectType {
	return RETURN_VALUE_OBJ
}
func (returnValue *ReturnValue) Inspect() string {
	return returnValue.Value.Inspect()
}

type Function struct {
	Parameters []string
	Body       *ast.BlockStatement
	Env        *Environment
}

func (fn *Function) Type() ObjectType {
	return FUNCTION_OBJ
}
func (fn *Function) Inspect() string {
	var out bytes.Buffer
	out.WriteString("fn(")
	out.WriteString(strings.Join(fn.Parameters, ", "))
	out.WriteString(") ")
	out.WriteString(fn.Body.String())
	return out.String()
}
