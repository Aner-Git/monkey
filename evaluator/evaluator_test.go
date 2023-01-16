package evaluator

import (
	lex "monkey/lexer"
	"monkey/object"
	"monkey/parser"
	"testing"
)

func testEval(input string) object.Object {
	l := lex.New(input)
	p := parser.New(l)

	return Eval(p.ParseProgram())
}

func TestReturnStatement(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"5*5; return 10; 4 ", 10},
		{"return 0; 4 ", 0},
		{"return 2*4; 4 ", 8},
		{`
          if(10>1){
		    if(10>1){
			  return 10;
			  }
			  return 1;
		  }
		`,

			10},
		{`
          if(10>1){
		    if(10>1){
    		    if(10>1){
				     return 10;
				  }
				  return 2;
			  }
			  return 1;
		  }
		`,

			10},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestIfElseExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"if(true){10}", 10},
		{"if(false){10}", nil},
		{"if(1){10}", 10},
		{"if(1<2){10}", 10},
		{"if(1>2){10}", nil},
		{"if(1>2){10}else{20}", 20},
		{"if(1<2){10}else{20}", 10},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		integer, ok := tt.expected.(int)

		if ok {
			testIntegerObject(t, evaluated, int64(integer))
		} else {
			testNullObject(t, evaluated)
		}
	}
}

func TestBangOperator(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"!true", false},
		{"!false", true},
		{"!5", false},
		{"!!5", true},
		{"!!false", false},
		{"!!true", true},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestEvalBooleanExpression(t *testing.T) {

	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
		{"1 > 2", false},
		{"2 > 1", true},
		{"1 > 1", false},
		{"1 < 1", false},
		{"1 == 1", true},
		{"1 != 1", false},
		{"1 == 2", false},
		{"1 != 2", true},
		{"true == true", true},
		{"false == false", true},
		{"false == true", false},
		{"false != true", true},
		{"true != false", true},
		{"(1<2) == true", true},
		{"(1<2) == false", false},
		{"(1>2) == false", true},
		{"(1>2) == true", false},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func testNullObject(t *testing.T, obj object.Object) bool {

	if obj != NULL {
		t.Fatalf("object is  not *object.Null. got=%T(%+v)", obj, obj)
		return false
	}

	return true
}

func testBooleanObject(t *testing.T, obj object.Object, expected bool) bool {

	result, ok := obj.(*object.Boolean)

	if !ok {
		t.Fatalf("object is  not *object.Boolean. got=%T(%+v)", obj, obj)
		return false
	}

	if got := result.Value; got != expected {
		t.Fatalf("Boolean.Value wrong. wanted=%t , got=%t", expected, got)
		return false
	}

	return true
}

func TestEvalIntegerExpression(t *testing.T) {

	tests := []struct {
		input    string
		expected int64
	}{
		{"5", 5},
		{"1906", 1906},
		{"-5", -5},
		{"-1906", -1906},
		{"5+5+5+5-10", 10},
		{"2*2*2*2*2", 32},
		{"-50 +100 -50", 0},
		{"5*2 +10", 20},
		{"5+ 2 *10 ", 25},
		{"50/2 * 2 + 10 ", 60},
		{"2 * (2 + 10)", 24},
		{"3 * 3 * 3  + 10", 37},
		{"3 * (3 * 3)  + 10", 37},
		{"(5 + 10 * 2  + 15/3)*2 + -10", 50},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func testIntegerObject(t *testing.T, obj object.Object, expected int64) bool {

	result, ok := obj.(*object.Integer)

	if !ok {
		t.Fatalf("object is  not *object.Integer. got=%T(%+v)", obj, obj)
		return false
	}

	if got := result.Value; got != expected {
		t.Fatalf("Integer.Value wrong. wanted=%d , got=%d", expected, got)
		return false
	}

	return true
}
