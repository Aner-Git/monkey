package repl

import (
	"bufio"
	"fmt"
	"io"
	"monkey/evaluator"
	lex "monkey/lexer"
	"monkey/object"
	"monkey/parser"
)

const PROMPRT = " >> "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	env := object.NewEnvironment()

	for {
		fmt.Printf(PROMPRT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		l := lex.New(line)

		p := parser.New(l)
		program := p.ParseProgram()

		if p.HasErrors() {
			printParserErrors(out, p.Errors())
			continue
		}

		obj := evaluator.Eval(program, env)

		if obj == nil {
			//	io.WriteString(out, "\tCant evaluate the expression!\n")
			continue
		}

		io.WriteString(out, obj.Inspect()+"\n")

	}
}

func printParserErrors(out io.Writer, errors []string) {

	for _, e := range errors {
		io.WriteString(out, "\t"+e+"\n")
	}

}
