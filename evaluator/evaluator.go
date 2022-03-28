package evaluator

import(
    "monkey/ast"
    "monkey/object"
)

var (

    NULL = &object.Null{}
    TRUE = &object.Boolean{Value: true}
    FALSE = &object.Boolean{Value: false}

)

func Eval(node ast.Node) object.Object{


    switch  v := node.(type){

        case *ast.Program:
            return evalStatements(v.Statements)

        case *ast.ExpressionStatement:
            return Eval(v.Expression)

        //Expressions
        case *ast.IntegerLiteral:
            return &object.Integer{Value: v.Value}

        case *ast.Boolean:
            return nativeBoolToBooleanObject(v.Value)
    }

    return nil
}

func evalStatements(stmts []ast.Statement) object.Object {
    var result object.Object

    for _, statement := range stmts{
        result = Eval(statement)
    }

    return result
}

func nativeBoolToBooleanObject(in bool) *object.Boolean {
    if in {
        return TRUE
    }

    return FALSE
}
