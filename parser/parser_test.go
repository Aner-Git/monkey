package parser

import (
    "fmt"
	"monkey/ast"
	lex "monkey/lexer"
	"testing"
)

func TestNewParser(t *testing.T) {
	input := ``

	l := lex.New(input)
	p := New(l)

	if p == nil {
		t.Fatalf("New Parser returned nil")
	}
}

func TestParserReportsErrors(t *testing.T) {
	input := `
    let x := 5;
    let  = 10;
    let 838383;
    `

	l := lex.New(input)
	p := New(l)

    _ = p.ParseProgram()


    if len(p.Errors()) != 6 {
		t.Fatalf("Expected errors[expected=6, got=%d]", len(p.Errors()))
    }

}


func TestReturnStatements(t *testing.T) {
	tests := []struct {
		input         string
		expectedValue interface{}
	}{
		{"return 5;", 5},
		{"return true;", true},
		{"return foobar;", "foobar"},
	}

	for _, tt := range tests {
		l := lex.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

        checkStatements(t, 1, program)

		stmt := program.Statements[0]
		returnStmt, ok := stmt.(*ast.ReturnStatement)
		if !ok {
			t.Fatalf("stmt not *ast.returnStatement. got=%T", stmt)
		}
		if returnStmt.TokenLiteral() != "return" {
			t.Fatalf("returnStmt.TokenLiteral not 'return', got %q",
				returnStmt.TokenLiteral())
		}
		if testLiteralExpression(t, returnStmt.ReturnValue, tt.expectedValue) {
			return
		}
	}
}


func TestLetStatements(t *testing.T) {

    tests := []struct {
        input string
		expectedIdentifier string
        expectedValue interface{}
	}{
		{"let x = 5;", "x", 5},
		{"let y = true;", "y", true},
		{"let foobar = 838383;", "foobar", 838383},
	}

    for _, tt := range tests{
        l := lex.New(tt.input)
        p := New(l)

        program := p.ParseProgram()
        checkParserErrors(t, p)
        checkStatements(t, 1, program)
		stmt := program.Statements[0]

        if !testLetStatement(t, stmt, tt.expectedIdentifier) {
			return
		}

        letstmt, ok := stmt.(*ast.LetStatement)
        if !ok {
            t.Fatalf("stmt is not ast.LetStatement. got=%T", letstmt)
        }

        testLiteralExpression(t, letstmt.Value, tt.expectedValue) 
    }
}

func testLetStatement(t *testing.T, s ast.Statement, name string) bool {

	if s.TokenLiteral() != "let" {
		t.Errorf("s.TokenLiteral not 'let'. [literal=%q]", s.TokenLiteral())
		return false
	}

	letStmt, ok := s.(*ast.LetStatement)
	if !ok {
		t.Errorf("s not *ast.LetStatement. [type=%T]", s)
	}

	if letStmt.Name.Value != name {
		t.Errorf("[letStmt.Name.Value=%s, expected=%s]", letStmt.Name.Value, name)
		return false
	}

	if letStmt.Name.TokenLiteral() != name {
		t.Errorf("[letStmt.Name.TokenLiteral()=%s, expected=%s]", letStmt.Name.TokenLiteral(), name)
		return false
	}

	return true
}



func TestIdentifierExpression(t *testing.T){
    input := "foobar;"

    l := lex.New(input)
    p := New(l)
    program := p.ParseProgram()

    checkParserErrors(t, p)
    checkStatements(t, 1, program)

    stmt := checkExpressionStatement(t, program)

    ident, ok := stmt.Expression.(*ast.Identifier)
    if !ok {
        t.Fatalf("exp is wrong type. [expected=*ast.Identifier, got=%T]", stmt.Expression)
    }

    if ident.Value != "foobar" {
        t.Errorf("ident.Value wrong. [expected=foobar, got=%s]", ident.Value)
    }

    if ident.TokenLiteral() != "foobar" {
        t.Errorf("ident.TokenLiteral is wrong. [expected=foobar, got=%s]", ident.TokenLiteral())
    }
}

func TestIntegerLiteralExpression(t *testing.T){
    input := "5;"

    l := lex.New(input)
    p := New(l)
    program := p.ParseProgram()
    checkParserErrors(t, p)

    checkStatements(t, 1, program)

    stmt, ok := program.Statements[0].(*ast.ExpressionStatement)

    if !ok {
        t.Fatalf("program.Statements[0] is wrong type. [expected=ast.ExpressionStatement, got=%T]", program.Statements[0])
    }

    literal, ok := stmt.Expression.(*ast.IntegerLiteral)

    if !ok {
        t.Fatalf("exp not *ast.IntegerLiteral. [expected=ast.IntegerLiteral, got=%T]", stmt.Expression) 
    }

    if literal.Value != 5 {
        t.Fatalf("Wrong Integer Literal value. [expected=5, got=%d]", literal.Value) 
    }

    if literal.Token.Literal != "5" {
        t.Fatalf("Wrong Integer TokenLiteral. [expected=\"5\", got=%q]", literal.TokenLiteral()) 
    }

}

func TestBooleanExpression(t *testing.T) {
	tests := []struct {
		input           string
		expectedBoolean bool
	}{
		{"true;", true},
		{"false;", false},
	}

	for _, tt := range tests {
		l := lex.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)


        checkStatements(t, 1, program)
        stmt := checkExpressionStatement(t, program)

		boolean, ok := stmt.Expression.(*ast.Boolean)
		if !ok {
			t.Fatalf("exp not *ast.Boolean. got=%T", stmt.Expression)
		}
		if boolean.Value != tt.expectedBoolean {
			t.Errorf("boolean.Value not %t. got=%t", tt.expectedBoolean,
				boolean.Value)
		}
	}
}

func TestParsingPrefixExpressions(t *testing.T){
    prefixTests := []struct{
        input string
        operator string
        value interface{} 
    }{
        {"!5;", "!", 5},
        {"-15;", "-", 15},
        {"!true;", "!", true},
        {"!false;", "!", false},
    }

    for _, tt := range prefixTests {
        l := lex.New(tt.input)
        p := New(l)
        program := p.ParseProgram()
        checkParserErrors(t, p)
        checkStatements(t, 1, program)
        stmt := checkExpressionStatement(t, program)

        exp, ok := stmt.Expression.(*ast.PrefixExpression)
        if !ok {
            t.Fatalf("stmt is not ast.PrefixExpression. got=%T", stmt.Expression)
        }

        if exp.Operator != tt.operator {
            t.Fatalf("Wrong exp.Operator [expected='%s', got=%s]", tt.operator, exp.Operator)
        }

        if !testLiteralExpression(t, exp.Right, tt.value){
            return
        }
    }
}

func checkExpressionStatement(t *testing.T, program *ast.Program ) *ast.ExpressionStatement{
    stmt, ok := program.Statements[0].(*ast.ExpressionStatement)

    if !ok {
        t.Fatalf("program.Statements[0] is wrong type. [expected=ast.ExpressionStatement, got=%T]", program.Statements[0])
        t.FailNow()
        return nil
    }

    return stmt
}

func checkStatements(t *testing.T, expected int, p *ast.Program ){

    if len(p.Statements) == expected {
        return
    }

    t.Fatalf("program has wrong number of statements. [expected=%d, got=%d]", expected, len(p.Statements))
    t.FailNow()
}

func checkParserErrors(t *testing.T, p *Parser){

    if !p.HasErrors() {
        return
    }

    errors := p.Errors()

    t.Errorf("parser had %d errors", len(errors))
    for _, msg := range errors{
        t.Errorf("parse error: %q", msg)
    }

    t.FailNow()
}

func testIntegerLiteral(t *testing.T, il ast.Expression, value int64) bool{

    integ, ok := il.(*ast.IntegerLiteral)
    if !ok {
        t.Errorf("il not *ast.IntegerLiteral. got=%T", il)
        return false
    }

    if integ.Value != value {
        t.Errorf("Wrong integ.Value [expected=%d, got=%d]", value, integ.Value)
        return false
    }

    if integ.TokenLiteral() != fmt.Sprintf("%d", value) {
        t.Errorf("Wrong integ.TokenLiteral [expected=%d, got=%q]", value, integ.TokenLiteral())
        return false
    }

    return true
}

func TestParsingInfixExpressions(t *testing.T){
    infixTests := []struct{
        input   string
        leftValue interface{} 
        operator string 
        rightValue  interface{}
    }{
        {"5 + 5;", 5, "+", 5},
        {"5 - 5;", 5, "-", 5},
        {"5 * 5;", 5, "*", 5},
        {"5 / 5;", 5, "/", 5},
        {"5 > 5;", 5, ">", 5},
        {"5 < 5;", 5, "<", 5},
        {"5 == 5;", 5, "==", 5},
        {"5 != 5;", 5, "!=", 5},
        {"true == true;", true, "==", true},
        {"true != false;", true, "!=", false},
        {"false == false;", false, "==", false},
    }

    for _, tt := range infixTests {
        l :=lex.New(tt.input)
        p := New(l)
        program := p.ParseProgram()
        checkParserErrors(t, p)
        checkStatements(t, 1, program)

        stmt := checkExpressionStatement(t, program)

        if !testInfixExpression(t, stmt.Expression, tt.leftValue, tt.operator, tt.rightValue){
            return
        }
    }
}


func TestOperatorPrecedenceParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			"-a * b",
			"((-a) * b)",
		},
		{
			"!-a",
			"(!(-a))",
		},
		{
			"a + b + c",
			"((a + b) + c)",
		},
		{
			"a + b - c",
			"((a + b) - c)",
		},
		{
			"a * b * c",
			"((a * b) * c)",
		},
		{
			"a * b / c",
			"((a * b) / c)",
		},
		{
			"a + b / c",
			"(a + (b / c))",
		},
		{
			"a + b * c + d / e - f",
			"(((a + (b * c)) + (d / e)) - f)",
		},
		{
			"3 + 4; -5 * 5",
			"(3 + 4)((-5) * 5)",
		},
		{
			"5 > 4 == 3 < 4",
			"((5 > 4) == (3 < 4))",
		},
		{
			"5 < 4 != 3 > 4",
			"((5 < 4) != (3 > 4))",
		},
		{
			"3 + 4 * 5 == 3 * 1 + 4 * 5",
			"((3 + (4 * 5)) == ((3 * 1) + (4 * 5)))",
		},
        {
            "true",
            "true",
        },
        {
            "false",
            "false",
        },
        {
            "3 > 5 == false",
            "((3 > 5) == false)",
        },
        {
            "3 < 5 == true",
            "((3 < 5) == true)",
        },
        {
            "1+(2+3)+4",
            "((1 + (2 + 3)) + 4)",
        },
        {
            "!(true==true)",
            "(!(true == true))",
        },
        {
            "1+(2+3)*4",
            "(1 + ((2 + 3) * 4))",
        },
        {
            "1+add(2+3)*4",
            "(1 + (add((2 + 3)) * 4))",
        },
        {
            "add(a,b,2*3,4+5,add(6,7*8))",
            "add(a,b,(2 * 3),(4 + 5),add(6,(7 * 8)))",
        },
        {
            "-add(a,b)",
            "(-add(a,b))",
        },
    }

	for _, tt := range tests {
		l := lex.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		actual := program.String()
		if actual != tt.expected {
			t.Errorf("expected=%q, got=%q", tt.expected, actual)
		}
	}
}

func testIdentifier(t *testing.T, exp ast.Expression, value string)bool{
    ident, ok := exp.(*ast.Identifier)

    if !ok {
        t.Errorf("exp not *ast.Identifier. got %T", exp)
        return false
    }

    if ident.String() != value{
        t.Errorf("Ident.Value not equal. expected=%s, got=%s", value, ident.Value)
        return false
    }

    if ident.TokenLiteral() != value{
        t.Errorf("ident.TokenLiteral not equal. expected=%s, got=%s", value, ident.TokenLiteral())
        return false
    }

    return true;
}

func testLiteralExpression(t *testing.T, exp ast.Expression, expected interface{}) bool{
    switch v:= expected.(type){
        case int:
            return testIntegerLiteral(t, exp, int64(v))
        case int64:
            return testIntegerLiteral(t, exp, v)
        case string: 
            return testIdentifier(t, exp, v)
        case bool: 
            return testBooleanLiteral(t, exp, v)
    }

    t.Errorf("type of exp not handled. got=%T", exp)
    return false
}

func testBooleanLiteral(t *testing.T, exp ast.Expression, value bool)bool{
    boolean, ok := exp.(*ast.Boolean)
 
    if !ok{
        t.Errorf("exp not *ast.Boolean. got=%T", exp)
        return false
    }

    if boolean.Value != value {
        t.Errorf("boolean.Value not matched. expected=%t, got=%t", value, boolean.Value)
        return false
    }

    if boolean.TokenLiteral() != fmt.Sprintf("%t", value){
        t.Errorf("boolean.TokenLiteral not matched. expected=%t, got=%s", value, boolean.TokenLiteral())
        return false
    }

    return true
}


func testInfixExpression(t* testing.T, exp ast.Expression, left interface{},
    operator string, right interface{})bool{

        opExp, ok := exp.(*ast.InfixExpression)

        if !ok{
            t.Errorf("exp is not ast.InfixExpression. got=%T(%s)", exp, exp)
            return false
        }

        if !testLiteralExpression(t, opExp.Left, left){
            return false
        }

        if opExp.Operator != operator{
            t.Errorf("exp.Operator did not match. expected='%s', got=%q", operator, opExp.Operator)
            return false
        }

        if !testLiteralExpression(t, opExp.Right, right){
            return false
        }
        return true
}

func TestIfExpression(t *testing.T){
    input := `if(x<y){x}`

    l := lex.New(input)
    p := New(l)

    program := p.ParseProgram()
    checkParserErrors(t,p)
    checkStatements(t, 1, program)
    stmt := checkExpressionStatement(t, program)
    
    exp, ok := stmt.Expression.(*ast.IfExpression)
    if !ok {
        t.Fatalf("exp is wrong type. [expected=*ast.IfExpression, got=%T]", stmt.Expression)
    }

    if !testInfixExpression(t, exp.Condition, "x","<", "y"){
        return
    }

    //check the consequence of the 'if'
    if ll := len(exp.Consequence.Statements); ll != 1{
        t.Errorf("consequence is not 1 statement. got=%d\n", ll)
    }

    consequence, ok := exp.Consequence.Statements[0].(*ast.ExpressionStatement)
    if !ok{
        t.Errorf("Statement[0] is not ast.ExpressionStatement. got=%T\n", consequence)
    }

    if !testIdentifier(t, consequence.Expression, "x"){
        return
    }

    if exp.Alternative != nil{
        t.Errorf("exp.Alternative was not nil. got=%+v,",exp.Alternative)
    }

}

func TestIfElseExpression(t *testing.T) {
	input := `if (x < y) { x } else { y }`

	l := lex.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain %d statements. got=%d\n",
			1, len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	exp, ok := stmt.Expression.(*ast.IfExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.IfExpression. got=%T", stmt.Expression)
	}

	if !testInfixExpression(t, exp.Condition, "x", "<", "y") {
		return
	}

	if len(exp.Consequence.Statements) != 1 {
		t.Errorf("consequence is not 1 statements. got=%d\n",
			len(exp.Consequence.Statements))
	}

	consequence, ok := exp.Consequence.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Statements[0] is not ast.ExpressionStatement. got=%T",
			exp.Consequence.Statements[0])
	}

	if !testIdentifier(t, consequence.Expression, "x") {
		return
	}

	if len(exp.Alternative.Statements) != 1 {
		t.Errorf("exp.Alternative.Statements does not contain 1 statements. got=%d\n",
			len(exp.Alternative.Statements))
	}

	alternative, ok := exp.Alternative.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Statements[0] is not ast.ExpressionStatement. got=%T",
			exp.Alternative.Statements[0])
	}

	if !testIdentifier(t, alternative.Expression, "y") {
		return
	}
}

func TestFunctionLiteralParsing(t *testing.T){
    input := `fn(x,y){x + y;}`

    l := lex.New(input)
    p := New(l)
    program := p.ParseProgram()

    checkParserErrors(t, p)
    checkStatements(t, 1, program)

    stmt := checkExpressionStatement(t, program)

    function, ok := stmt.Expression.(*ast.FunctionLiteral)
    if !ok {

		t.Fatalf("Statements[0] is not ast.FunctionLiteral. got=%T",
			stmt.Expression)
    }

    if l := len(function.Parameters);l !=2 {
		t.Fatalf("Function literal parameters wrong. exp=2, got=%d",l)
    }

    testLiteralExpression(t, function.Parameters[0], "x")
    testLiteralExpression(t, function.Parameters[1], "y")

    if l := len(function.Body.Statements); l !=1 {
		t.Fatalf("Function expected to have 1 body. exp=1, got=%d",l)
    }

    bodystmt, ok := function.Body.Statements[0].(*ast.ExpressionStatement)
    if !ok {
		t.Fatalf("Function body stmt is not ast.ExpressionStatement. got=%T",function.Body.Statements[0])
    }

    testInfixExpression(t, bodystmt.Expression, "x", "+", "y")
}


func TestFunctionParameterParsing(t *testing.T){
    tests :=[]struct{
        input string
        expectedParams []string
    }{
        {input:"fn(){};", expectedParams:[]string{}},
        {input:"fn(x){};", expectedParams:[]string{"x"}},
        {input:"fn(x,y,z){};", expectedParams:[]string{"x", "y", "z"}},
    }

    for _, tt := range tests{

        l := lex.New(tt.input)
        p := New(l)
        program := p.ParseProgram()

        checkParserErrors(t, p)
        checkStatements(t, 1, program)

        stmt := checkExpressionStatement(t, program)

        function, ok := stmt.Expression.(*ast.FunctionLiteral)
        if !ok {

            t.Fatalf("Statements[0] is not ast.FunctionLiteral. got=%T",
                stmt.Expression)
        }

        if le, lg := len(function.Parameters), len(tt.expectedParams);le != lg {
            t.Fatalf("Function literal parameters wrong. exp=%d, got=%d",le, lg)
        }

        for i, exp := range tt.expectedParams {
            testLiteralExpression(t, function.Parameters[i], exp)
        }

    }
}


func TestCallExpresssionParsing(t *testing.T){
    input := "add(1, 2*3, 4+5);"

    l := lex.New(input)
    p := New(l)
    program := p.ParseProgram()

    checkParserErrors(t, p)
    checkStatements(t, 1, program)
    stmt := checkExpressionStatement(t, program)

    exp, ok := stmt.Expression.(*ast.CallExpression)
    if !ok {

        t.Fatalf("Statements[0] is not ast.CallExpression. got=%T",
            stmt.Expression)
    }

    if !testIdentifier(t, exp.Function, "add"){
        return
    }

    if l := len(exp.Arguments); l != 3 {
            t.Fatalf("Function literal parameters wrong. exp=3, got=%d", l)
    }

    testLiteralExpression(t, exp.Arguments[0], 1)
    testInfixExpression(t, exp.Arguments[1], 2, "*", 3)
    testInfixExpression(t, exp.Arguments[2], 4, "+", 5)

}
