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

	env := object.NewEnvironment()
	return Eval(p.ParseProgram(), env)
}

func TestArrayIndexExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`[1,2,3][0]`, 1},
		{`[1,2,3][1]`, 2},
		{`[1,2,3][2]`, 3},
		{`let i = 0; [1,2,3][i]`, 1},
		{`let myArray = [1,2,3]; let i = myArray[0]; myArray[i]`, 2},
		{`[2,3][3]`, nil},
		{`[1,2,3][-1]`, nil},
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

func TestArrayLiterals(t *testing.T) {
	input := "[1, 2*2, 3+3]"

	evaluated := testEval(input)

	result, ok := evaluated.(*object.Array)

	if !ok {
		t.Fatalf("object is not Array. got=%T (%+v)", evaluated, evaluated)
	}

	if len(result.Elements) != 3 {
		t.Fatalf("array has wrong num of elements. got=%d", len(result.Elements))
	}

	testIntegerObject(t, result.Elements[0], 1)
	testIntegerObject(t, result.Elements[1], 4)
	testIntegerObject(t, result.Elements[2], 6)
}

func TestBuiltinFunctions(t *testing.T) {
	//
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`len("")`, 0},
		{`len("four")`, 4},
		{`len("hello world")`, 11},
		{`len(1)`, "argument to `len` not suported, got INTEGER"},
		{`len("one", "two")`, "wrong number of arguments. got=2, want=1"},
	}

	for _, tt := range tests {

		obj := testEval(tt.input)

		switch expected := tt.expected.(type) {
		case int:
			testIntegerObject(t, obj, int64(expected))
		case string:
			errObj, ok := obj.(*object.Error)
			if !ok {
				t.Errorf("object is not Error. got=%T(%+v)", obj, obj)
				continue
			}
			if errObj.Message != expected {
				t.Errorf("wrong error message. expected=%q, got=%q", expected, errObj.Message)
			}
		}

	}
}

func TestStringObject(t *testing.T) {
	input := `"HELLO World"`
	obj := testEval(input)
	s, ok := obj.(*object.String)

	if !ok {
		t.Fatalf("object is  not *object.String. got=%T(%+v)", obj, obj)
	}

	if s.Value != "HELLO World" {
		t.Fatalf("object.Value wrong. got=%q, expected=%q", s.Value, "HELLO World")
	}
}

func TestStringConcat(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			`"" + ""`,
			"",
		},

		{
			`"let" + " identity"`,
			"let identity",
		},

		{
			`"A" + "B" + "C"`,
			"ABC",
		},
	}

	for _, tt := range tests {

		obj := testEval(tt.input)
		s, ok := obj.(*object.String)

		if !ok {
			t.Fatalf("object is  not *object.String. got=%T(%+v)", obj, obj)
		}

		if s.Value != tt.expected {
			t.Fatalf("object.Value wrong. got=%q, expected=%q", s.Value, tt.expected)
		}
	}

}

func TestFunctionObject(t *testing.T) {
	input := "fn(x){x+2;};"
	obj := testEval(input)

	fn, ok := obj.(*object.Function)
	if !ok {
		t.Fatalf("object is  not *object.Function. got=%T(%+v)", obj, obj)
	}

	if len(fn.Parameters) != 1 {
		t.Fatalf("function has wrong parameters. Parameters= %+v", fn.Parameters)
	}

	if fn.Parameters[0].String() != "x" {
		t.Fatalf("Parameter is not 'x'. got= %q", fn.Parameters[0])
	}

	expepectedBody := "(x + 2)"
	if fn.Body.String() != expepectedBody {
		t.Fatalf("body is not %q. got=%q", expepectedBody, fn.Body.String())
	}
}

func TestFunctionClosure(t *testing.T) {
	input := ` 
	let newAdder = fn(x){
	 fn(y) {x+y};
	 };

	 let addTwo = newAdder(2);
	 addTwo(2);
	 `
	testIntegerObject(t, testEval(input), 4)
}

func TestFunctionApplication(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{
			"let identity = fn(x){x;}; identity(5)",
			5,
		},
		{
			"let identity = fn(x){return x;}; identity(5)",
			5,
		},
		{
			"let double = fn(x){x*2;}; double(5)",
			10,
		},

		{
			"let add = fn(x,y){x+y;}; add(5,5)",
			10,
		},
		{
			"let add = fn(x,y){x+y;}; add(5+5,add(5,5))",
			20,
		},
		{
			"fn(y){y;}(5)",
			5,
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestLetStatement(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{
			input:    "let x=5;x",
			expected: 5,
		},
		{
			input:    "let x=5*5;x",
			expected: 25,
		},
		{
			input:    "let a=5; let b=a;b;",
			expected: 5,
		},
		{
			input:    "let a=5; let b=a;let c = a+b+5;c;",
			expected: 15,
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		input           string
		expectedMessage string
	}{
		{
			"5+true;",
			"type mismatch: INTEGER + BOOLEAN",
		},
		{
			"5+true;5;",
			"type mismatch: INTEGER + BOOLEAN",
		},
		{
			"-true;5;",
			"unknown operator: -BOOLEAN",
		},
		{
			"true + false;",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"5;true + false;5",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"if(10>1){true + false;}",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{`
          if(10>1){
		    if(10>1){
			   false + true;
			  }
			  return 1;
		  }
		`,
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"foobar",
			"identifier not found: foobar",
		},
		{
			`"foo" - "bar"`,
			"unknown operator: STRING - STRING",
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("no error object returned. got=%T(%+v)", evaluated, evaluated)
			continue
		}

		if errObj.Message != tt.expectedMessage {
			t.Errorf("wrong error message. expected=%q, got%q", tt.expectedMessage, errObj.Message)
		}
	}
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
