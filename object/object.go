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
	ARRAY_OBJ        = "ARRAY"
	HASH_OBJ         = "HASH"
	RETURN_VALUE_OBJ = "RETURN_VALUE"
	FUNCTION_OBJ     = "FUNCTION"
	BUILTIN_OBJ      = "BUILTIN"
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

type Array struct {
	Elements []Object
}

func (arr *Array) Type() ObjectType {
	return ARRAY_OBJ
}
func (arr *Array) Inspect() string {
	var out bytes.Buffer
	out.WriteString("[")
	for i, elem := range arr.Elements {
		if i > 0 {
			out.WriteString(", ")
		}
		out.WriteString(elem.Inspect())
	}
	out.WriteString("]")
	return out.String()
}

type HashKey struct {
	Type ObjectType
	str  string
	num  int64
}

func HashKeyFromString(str string) HashKey {
	return HashKey{Type: STRING_OBJ, str: str}
}
func HashKeyFromInt(integer int64) HashKey {
	return HashKey{Type: INTEGER_OBJ, num: integer}
}
func HashKeyFromBool(boolean bool) HashKey {
	key := HashKey{Type: BOOLEAN_OBJ}
	if boolean {
		key.num = 1
	}
	return key
}
func HashKeyFromObject(obj Object) (HashKey, bool) {
	switch obj := obj.(type) {
	case *String:
		return HashKeyFromString(obj.Value), true
	case *Integer:
		return HashKeyFromInt(obj.Value), true
	case *Boolean:
		return HashKeyFromBool(obj.Value), true
	default:
		return HashKey{}, false
	}
}

func (key *HashKey) AsString() (string, bool) {
	if key.Type == STRING_OBJ {
		return key.str, true
	}
	return "", false
}

func (key *HashKey) AsInteger() (int64, bool) {
	if key.Type == INTEGER_OBJ {
		return key.num, true
	}
	return 0, false
}

func (key *HashKey) AsBoolean() (bool, bool) {
	if key.Type == BOOLEAN_OBJ {
		if key.num == 1 {
			return true, true
		} else {
			return false, true
		}
	}
	return false, false
}

type Hash struct {
	Entries map[HashKey]Object
}

func (hash *Hash) Type() ObjectType {
	return HASH_OBJ
}
func (hash *Hash) Inspect() string {
	var out bytes.Buffer
	out.WriteString("{")
	first := true
	for key, elem := range hash.Entries {
		if !first {
			out.WriteString(", ")
		}
		if str, ok := key.AsString(); ok {
			out.WriteString("\"")
			out.WriteString(str)
			out.WriteString("\"")
		} else if num, ok := key.AsInteger(); ok {
			out.WriteString(fmt.Sprintf("%d", num))
		} else if bool, ok := key.AsBoolean(); ok {
			if bool {
				out.WriteString("true")
			} else {
				out.WriteString("false")
			}
		}
		out.WriteString(": ")
		out.WriteString(elem.Inspect())
		first = false
	}
	out.WriteString("}")
	return out.String()
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

type BuiltinFunction func(args ...Object) (Object, error)

type Builtin struct {
	Fn BuiltinFunction
}

func (b *Builtin) Type() ObjectType {
	return BUILTIN_OBJ
}
func (b *Builtin) Inspect() string {
	return "builtin function"
}
