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

	case *ast.FunctionLiteral:
		return evalFunction(v, env)

	//Expressions
	case *ast.IntegerLiteral:
		return &object.Integer{Value: v.Value}

	case *ast.StringLiteral:
		return &object.String{Value: v.Value}

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

	case *ast.CallExpression:
		function := Eval(v.Function, env)
		if isError(function) {
			return function
		}

		args := evalExpressions(v.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}

		//bind arguments, and call Body Expression
		return applyFunction(function, args)

	case *ast.ArrayLiteral:
		elements := evalExpressions(v.Elements, env)
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}

		return &object.Array{Elements: elements}

	case *ast.IndexExpression:
		left := Eval(v.Left, env)
		if isError(left) {
			return left
		}

		index := Eval(v.Index, env)
		if isError(index) {
			return index
		}

		return evalIndexExpression(left, index)
	}

	return nil
}

func evalIndexExpression(left, index object.Object) object.Object {

	arrayOb, ok := left.(*object.Array)
	if !ok {
		return newError("index operator not supported: %s", left.Type())
	}

	indxOb, ok := index.(*object.Integer)
	if !ok {
		return newError("Invalid index")
	}

	idx := indxOb.Value
	max := int64(len(arrayOb.Elements) - 1)

	if idx < 0 || idx > max {
		return NULL
	}

	return arrayOb.Elements[idx]
}

func applyFunction(fn object.Object, args []object.Object) object.Object {

	switch fn := fn.(type) {
	case *object.Function:
		extendedEnv := extendFunctionEnv(fn, args)
		evaluated := Eval(fn.Body, extendedEnv)

		return unwrapReturnValue(evaluated)

	case *object.Builtin:
		return fn.Fn(args...)
	}

	return newError("not a function: %s", fn.Type())
}

func unwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(*object.ReturnValue); ok {
		return returnValue.Value
	}
	return obj
}

func extendFunctionEnv(fn *object.Function, args []object.Object) *object.Environment {
	env := object.ExtendEnvironment(fn.Env)
	for i, param := range fn.Parameters {
		env.Set(param.Value, args[i])
	}
	return env
}

func evalExpressions(args []ast.Expression, env *object.Environment) []object.Object {
	var evaluated []object.Object
	for _, arg := range args {
		obj := Eval(arg, env)
		if isError(obj) {
			return []object.Object{obj}
		}
		evaluated = append(evaluated, obj)
	}

	return evaluated
}

func evalFunction(node *ast.FunctionLiteral, env *object.Environment) object.Object {
	return &object.Function{Parameters: node.Parameters, Body: node.Body, Env: env}
}

func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	if val, ok := env.Get(node.Value); ok {
		return val
	}

	if builtin, ok := builtins[node.Value]; ok {
		return builtin
	}

	return newError("identifier not found: " + node.Value)
}

func evalReturnStatement(node *ast.ReturnStatement, env *object.Environment) object.Object {
	v := Eval(node.ReturnValue, env)
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
	case (right.Type() == object.STRING_OBJ && left.Type() == object.STRING_OBJ):
		return evalStringInfixExpression(exp.Token.Type, left, right)
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

func evalStringInfixExpression(op token.TokenType, left, right object.Object) object.Object {
	rvalue := right.(*object.String).Value
	lvalue := left.(*object.String).Value

	switch op {
	case token.PLUS:
		return &object.String{Value: lvalue + rvalue}
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
