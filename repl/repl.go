package repl

import (
	"bufio"
	"fmt"
	"io"
	lex "monkey/lexer"
	"monkey/parser"
	"monkey/evaluator"
)

const PROMPRT = " >> "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

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

        obj := evaluator.Eval(program)

        if obj == nil{
            io.WriteString(out, "\tCant evaluate the expression!\n")
            continue
        }

        io.WriteString(out,obj.Inspect() + "\n")

	}
}

func printParserErrors(out io.Writer, errors []string){

    for _, e := range(errors){
        io.WriteString(out, "\t" + e + "\n")
    }


}
