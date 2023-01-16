package evaluator

import (
	"monkey/ast"
	"monkey/object"
	"monkey/token"
)

var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

func Eval(node ast.Node) object.Object {

	switch v := node.(type) {

	case *ast.Program:
		return evalProgram(v.Statements)

	case *ast.ExpressionStatement:
		return Eval(v.Expression)

	case *ast.BlockStatement:
		return evalBlockStatements(v.Statements)

	case *ast.ReturnStatement:
		return evalReturnStatement(v)

	//Expressions
	case *ast.IntegerLiteral:
		return &object.Integer{Value: v.Value}

	case *ast.Boolean:
		return nativeBoolToBooleanObject(v.Value)

	case *ast.PrefixExpression:
		return evalPrefixExpression(v)

	case *ast.InfixExpression:
		return evalInfixExpression(v)

	case *ast.IfExpression:
		return evalIfExpression(v)
	}

	return nil
}

func evalReturnStatement(stm *ast.ReturnStatement) object.Object {
	v := Eval(stm.ReturnValue)
	return &object.ReturnValue{Value: v}
}

func evalIfExpression(exp *ast.IfExpression) object.Object {
	condition := Eval(exp.Condition)

	if isTruthy(condition) {
		return Eval(exp.Consequence)
	} else if exp.Alternative != nil {
		return Eval(exp.Alternative)
	}

	return NULL

}

func isTruthy(o object.Object) bool {

	//null and false are false, all other are true
	switch o {
	case NULL, FALSE:
		return false
	}

	return true
}

func evalInfixExpression(exp *ast.InfixExpression) object.Object {

	right := Eval(exp.Right)
	left := Eval(exp.Left)

	switch {
	case (right.Type() == object.INTEGER_OBJ && left.Type() == object.INTEGER_OBJ):
		return evalIntegerInfixExpression(exp.Token.Type, left, right)
	case (right.Type() == object.BOOLEAN_OBJ && left.Type() == object.BOOLEAN_OBJ):
		return evalBooleanInfixExpression(exp.Token.Type, left, right)

	default:
		return NULL
	}
}

func evalBooleanInfixExpression(op token.TokenType, left, right object.Object) object.Object {

	switch op {
	case token.EQ:
		return nativeBoolToBooleanObject(left == right)
	case token.NOT_EQ:
		return nativeBoolToBooleanObject(left != right)

	default:
		return NULL
	}
}

func evalIntegerInfixExpression(op token.TokenType, left, right object.Object) object.Object {

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
	case token.LT:
		return nativeBoolToBooleanObject(lvalue < rvalue)
	case token.GT:
		return nativeBoolToBooleanObject(lvalue > rvalue)
	case token.EQ:
		return nativeBoolToBooleanObject(lvalue == rvalue)
	case token.NOT_EQ:
		return nativeBoolToBooleanObject(lvalue != rvalue)
	default:
		return NULL
	}

	return &object.Integer{Value: value}
}

func evalPrefixExpression(exp *ast.PrefixExpression) object.Object {

	right := Eval(exp.Right)

	switch exp.Token.Type {
	case token.BANG:
		return evalBangOperatorExpression(right)

	case token.MINUS:
		return evalMinusOperatorExpression(right)
	}

	return NULL
}

func evalMinusOperatorExpression(right object.Object) object.Object {

	if right.Type() != object.INTEGER_OBJ {

		return NULL

	}

	intObj, ok := right.(*object.Integer)

	if !ok {
		panic("Failed to cast to *object.Integer")
	}

	return &object.Integer{Value: -intObj.Value}
}

func evalBangOperatorExpression(right object.Object) object.Object {

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

func evalProgram(stmts []ast.Statement) object.Object {
	var result object.Object

	for _, statement := range stmts {
		result = Eval(statement)

		if rv, ok := result.(*object.ReturnValue); ok {
			return rv.Value
		}
	}

	return result
}

func evalBlockStatements(stmts []ast.Statement) object.Object {
	var result object.Object

	for _, statement := range stmts {
		result = Eval(statement)

		//keep returning bubbling up
		if result != nil && result.Type() == object.RETURN_VALUE_OBJ {
			return result
		}
	}

	return result
}

func nativeBoolToBooleanObject(in bool) *object.Boolean {
	if in {
		return TRUE
	}

	return FALSE
}
