package evaluator

import(
    "monkey/ast"
    "monkey/object"
    "monkey/token"
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

        case *ast.PrefixExpression:
            return evalPrefixExpression(v)

        case *ast.InfixExpression:
            return evalInfixExpression(v)
    }

    return nil
}

func evalInfixExpression(exp *ast.InfixExpression) object.Object{

    right := Eval(exp.Right)
    left := Eval(exp.Left)

    switch {
        case (right.Type() == object.INTEGER_OBJ && left.Type() == object.INTEGER_OBJ):
            return evalIntegerInfixExpression(exp.Token.Type, left, right)

        default:
            return NULL
    }
}

func evalIntegerInfixExpression(op token.TokenType, left object.Object, right object.Object) object.Object{

    rvalue := right.(*object.Integer).Value
    lvalue := left.(*object.Integer).Value

    var value int64 = 0

    switch op {
        case token.PLUS:
            value = lvalue + rvalue
        case token.MINUS:
            value = lvalue - rvalue
        case token.ASTERISK:
            value = lvalue * rvalue
        case token.SLASH:
            value = lvalue / rvalue
        default:
            return NULL
    }

    return &object.Integer{Value:value}
}

func evalPrefixExpression(exp *ast.PrefixExpression) object.Object{

    right := Eval(exp.Right)

    switch exp.Token.Type {
        case token.BANG:
            return evalBangOperatorExpression(right)

        case token.MINUS:
            return evalMinusOperatorExpression(right)
    }

    return NULL
}

func evalMinusOperatorExpression(right object.Object) object.Object{

    if right.Type() != object.INTEGER_OBJ{

        return NULL

    }

    intObj, ok  := right.(*object.Integer)

    if !ok {
       panic("Failed to cast to *object.Integer") 
    }

    return &object.Integer{Value: -intObj.Value}
}

func evalBangOperatorExpression(right object.Object) object.Object{

    //use the pre defined true/false vars
    switch right {
        case TRUE:
            return FALSE
        case FALSE:
            return TRUE
        case NULL:
            return TRUE
        }

    return FALSE
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
