package object

import "fmt"

type ObjectType string

const (
	NULL_OBJ    = "NULL"
	INTEGER_OBJ = "INTEGER"
	BOOLEAN_OBJ = "BOOLEAN"
)

type Object interface {
	Type() ObjectType
	Inspect() string
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
