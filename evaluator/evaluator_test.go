package evaluator

import (
	lex "monkey/lexer"
	"monkey/object"
	"monkey/parser"
	"testing"
)

func testEval(input string) object.Object{
        l := lex.New(input)
        p := parser.New(l)

        return Eval(p.ParseProgram());
}

func TestBangOperator(t* testing.T){
    tests := []struct {
        input string
		expected bool
	}{
		{"!true", false},
		{"!false", true},
		{"!5", false},
		{"!!5", true},
		{"!!false", false},
		{"!!true", true},
	}

    for _, tt := range tests{
        evaluated := testEval(tt.input)
        testBooleanObject(t, evaluated, tt.expected)
    }
}


func TestEvalBooleanExpression(t *testing.T){

    tests := []struct {
        input string
		expected bool
	}{
		{"true", true},
		{"false", false},
	}

    for _, tt := range tests{
        evaluated := testEval(tt.input)
        testBooleanObject(t, evaluated, tt.expected)
    }
}

func testBooleanObject(t *testing.T, obj object.Object, expected bool) bool{

    result, ok := obj.(*object.Boolean)

    if !ok {
        t.Fatalf("object is  not *object.Boolean. got=%T(%+v)", obj, obj)
        return false
    }

    if got := result.Value; got != expected {
        t.Fatalf("Boolean.Value wrong. wanted=%t , got=%t",expected, got)
        return false
    }

    return true
}

func TestEvalIntegerExpression(t *testing.T){

    tests := []struct {
        input string
		expected int64
	}{
		{"5", 5},
		{"1906", 1906},
		{"-5", -5},
		{"-1906", -1906},
        {"5+5+5+5-10",10},
        {"2*2*2*2*2",32},
        {"-50 +100 -50",0},
        {"5*2 +10",20},
        {"5+ 2 *10 ",25},
        {"50/2 * 2 + 10 ",60},
        {"2 * (2 + 10)",24},
        {"3 * 3 * 3  + 10",37},
        {"3 * (3 * 3)  + 10",37},
        {"(5 + 10 * 2  + 15/3)*2 + -10",50},
	}

    for _, tt := range tests{
        evaluated := testEval(tt.input)
        testIntegerObject(t, evaluated, tt.expected)
    }
}

func testIntegerObject(t *testing.T, obj object.Object, expected int64) bool{

    result, ok := obj.(*object.Integer)

    if !ok {
        t.Fatalf("object is  not *object.Integer. got=%T(%+v)", obj, obj)
        return false
    }

    if got := result.Value; got != expected {
        t.Fatalf("Integer.Value wrong. wanted=%d , got=%d",expected, got)
        return false
    }

    return true
}
