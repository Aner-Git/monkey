package evaluator

import (
	"fmt"
	"monkey/ast"
	"monkey/object"
	"monkey/token"
)

var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

func Eval(node ast.Node, env *object.Environment) object.Object {

	switch v := node.(type) {

	case *ast.Program:
		return evalProgram(v.Statements, env)

	case *ast.ExpressionStatement:
		return Eval(v.Expression, env)

	case *ast.BlockStatement:
		return evalBlockStatements(v.Statements, env)

	case *ast.ReturnStatement:
		return evalReturnStatement(v, env)

	case *ast.LetStatement:
		val := Eval(v.Value, env)
		if isError(val) {
			return val
		}
		//bind the identifier
		env.Set(v.Name.Value, val)

	//Expressions
	case *ast.IntegerLiteral:
		return &object.Integer{Value: v.Value}

	case *ast.Boolean:
		return nativeBoolToBooleanObject(v.Value)

	case *ast.PrefixExpression:
		return evalPrefixExpression(v, env)

	case *ast.InfixExpression:
		return evalInfixExpression(v, env)

	case *ast.IfExpression:
		return evalIfExpression(v, env)

	case *ast.Identifier:
		return evalIdentifier(v, env)
	}

	return nil
}

func evalIdentifier(stm *ast.Identifier, env *object.Environment) object.Object {
	val, ok := env.Get(stm.Value)
	if !ok {
		return newError("identifier not found: " + stm.Value)
	}
	return val
}

func evalReturnStatement(stm *ast.ReturnStatement, env *object.Environment) object.Object {
	v := Eval(stm.ReturnValue, env)
	if isError(v) {
		return v
	}

	return &object.ReturnValue{Value: v}
}

func evalIfExpression(exp *ast.IfExpression, env *object.Environment) object.Object {
	condition := Eval(exp.Condition, env)

	if isError(condition) {
		return condition
	}

	if isTruthy(condition) {
		return Eval(exp.Consequence, env)
	} else if exp.Alternative != nil {
		return Eval(exp.Alternative, env)
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

func evalInfixExpression(exp *ast.InfixExpression, env *object.Environment) object.Object {

	right := Eval(exp.Right, env)
	if isError(right) {
		return right
	}

	left := Eval(exp.Left, env)
	if isError(left) {
		return left
	}

	switch {
	case (right.Type() == object.INTEGER_OBJ && left.Type() == object.INTEGER_OBJ):
		return evalIntegerInfixExpression(exp.Token.Type, left, right)
	case (right.Type() == object.BOOLEAN_OBJ && left.Type() == object.BOOLEAN_OBJ):
		return evalBooleanInfixExpression(exp.Token.Type, left, right)
	case (right.Type() != left.Type()):
		return newError("type mismatch: %s %s %s", left.Type(), exp.Token.Type, right.Type())

	default:
		return newError("unknown operator: %s %s %s", left.Type(), exp.Token.Type, right.Type())
	}
}

func newError(format string, a ...interface{}) *object.Error {
	e := fmt.Sprintf(format, a...)
	return &object.Error{Message: e}
}

func evalBooleanInfixExpression(op token.TokenType, left, right object.Object) object.Object {

	switch op {
	case token.EQ:
		return nativeBoolToBooleanObject(left == right)
	case token.NOT_EQ:
		return nativeBoolToBooleanObject(left != right)

	default:
		return newError("unknown operator: %s %s %s", left.Type(), op, right.Type())
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
		return newError("unknown operator: %s %s %s", left.Type(), op, right.Type())
	}

	return &object.Integer{Value: value}
}

func evalPrefixExpression(exp *ast.PrefixExpression, env *object.Environment) object.Object {

	right := Eval(exp.Right, env)

	if isError(right) {
		return right
	}

	switch exp.Token.Type {
	case token.BANG:
		return evalBangOperatorExpression(right)

	case token.MINUS:
		return evalMinusOperatorExpression(right)
	}

	return newError("unknown operator: %s%s", exp.Token.Type, right.Type())
}

func evalMinusOperatorExpression(right object.Object) object.Object {

	if right.Type() != object.INTEGER_OBJ {

		return newError("unknown operator: -%s", right.Type())

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

func evalProgram(stmts []ast.Statement, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range stmts {
		result = Eval(statement, env)

		switch result := result.(type) {

		case *object.ReturnValue:
			return result.Value

		case *object.Error:
			return result

		}
	}

	return result
}

func evalBlockStatements(stmts []ast.Statement, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range stmts {
		result = Eval(statement, env)

		//Return or Error: bubble up
		if result != nil {
			if result.Type() == object.ERROR_OBJ || result.Type() == object.RETURN_VALUE_OBJ {
				return result
			}
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

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false

}
