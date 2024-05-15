package evaluator

import (
	"fmt"

	"danielmcm.com/interpreterbook/object"
)

var builtins = map[string]*object.Builtin{
	"len": {
		Fn: func(args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("`len` received wrong number of arguments. expected %d, got %d", 1, len(args))
			}
			switch arg := args[0].(type) {
			case *object.String:
				return &object.Integer{Value: int64(len(arg.Value))}, nil
			default:
				return nil, fmt.Errorf("`len` argument of type %s not supported", args[0].Type())
			}
		},
	},
}
