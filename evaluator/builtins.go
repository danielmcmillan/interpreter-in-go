package evaluator

import (
	"fmt"

	"danielmcm.com/interpreterbook/object"
)

func checkArgCount(name string, args []object.Object, expected int) error {
	if len(args) != expected {
		return fmt.Errorf("`%s` received wrong number of arguments. expected %d, got %d", name, expected, len(args))
	}
	return nil
}

func argTypeError(name string, arg object.Object) error {
	return fmt.Errorf("`%s` argument of type %s not supported", name, arg.Type())
}

var builtins = map[string]*object.Builtin{
	"len": {
		Fn: func(args ...object.Object) (object.Object, error) {
			if err := checkArgCount("len", args, 1); err != nil {
				return nil, err
			}
			switch arg := args[0].(type) {
			case *object.String:
				return &object.Integer{Value: int64(len(arg.Value))}, nil
			case *object.Array:
				return &object.Integer{Value: int64(len(arg.Elements))}, nil
			default:
				return nil, argTypeError("len", args[0])
			}
		},
	},
	"first": {
		Fn: func(args ...object.Object) (object.Object, error) {
			if err := checkArgCount("first", args, 1); err != nil {
				return nil, err
			}
			arr, ok := args[0].(*object.Array)
			if !ok {
				return nil, argTypeError("first", args[0])
			}
			if len(arr.Elements) > 0 {
				return arr.Elements[0], nil
			} else {
				return NULL, nil
			}
		},
	},
	"last": {
		Fn: func(args ...object.Object) (object.Object, error) {
			if err := checkArgCount("last", args, 1); err != nil {
				return nil, err
			}
			arr, ok := args[0].(*object.Array)
			if !ok {
				return nil, argTypeError("last", args[0])
			}
			if len(arr.Elements) > 0 {
				return arr.Elements[len(arr.Elements)-1], nil
			} else {
				return NULL, nil
			}
		},
	},
	"rest": {
		Fn: func(args ...object.Object) (object.Object, error) {
			if err := checkArgCount("rest", args, 1); err != nil {
				return nil, err
			}
			arr, ok := args[0].(*object.Array)
			if !ok {
				return nil, argTypeError("rest", args[0])
			}
			if len(arr.Elements) > 0 {
				return &object.Array{
					Elements: arr.Elements[1:],
				}, nil
			} else {
				return NULL, nil
			}
		},
	},
	"push": {
		Fn: func(args ...object.Object) (object.Object, error) {
			if err := checkArgCount("push", args, 2); err != nil {
				return nil, err
			}
			arr, ok := args[0].(*object.Array)
			if !ok {
				return nil, argTypeError("push", args[0])
			}
			elements := make([]object.Object, len(arr.Elements)+1)
			copy(elements, arr.Elements)
			elements[len(arr.Elements)] = args[1]
			return &object.Array{
				Elements: elements,
			}, nil
		},
	},
	"puts": {
		Fn: func(args ...object.Object) (object.Object, error) {
			for _, arg := range args {
				_, err := fmt.Println(arg.Inspect())
				if err != nil {
					return nil, err
				}
			}
			return NULL, nil
		},
	},
}
