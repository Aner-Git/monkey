package evaluator

import "monkey/object"

var builtins = map[string]*object.Builtin{
	"len": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			if args[0].Type() != object.STRING_OBJ {
				return newError("argument to `len` not suported, got %s", args[0].Type())
			}

			strObj, _ := args[0].(*object.String)
			return &object.Integer{Value: int64(len(strObj.Value))}
		},
	},
}
