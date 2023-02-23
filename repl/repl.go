package repl

import (
	"bufio"
	"fmt"
	"io"
	"monkey-go/evaluator"
	"monkey-go/lexer"
	"monkey-go/object"
	"monkey-go/parser"
	"strings"
)

const PROMPT = ">> "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	env := object.NewEnvironment()

	for {
		fmt.Fprintf(out, PROMPT) // nolint
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		// exit the REPL
		if strings.ToLower(line) == "exit" {
			break
		}

		l := lexer.New(line)
		p := parser.New(l)

		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			printParserErrors(out, p.Errors())
			continue
		}

		evaluated := evaluator.Eval(program, env)
		if evaluated != nil {
			_, err := io.WriteString(out, evaluated.Inspect()+"\n")
			if err != nil {
				return
			}
		}

	}
}

const MONKEY_FACE = `            __,__
   .--.  .-"     "-.  .--.
  / .. \/  .-. .-.  \/ .. \
 | |  '|  /   Y   \  |'  | |
 | \   \  \ 0 | 0 /  /   / |
  \ '- ,\.-"""""""-./, -' /
   ''-' /_   ^ ^   _\ '-''
       |  \._   _./  |
       \   \ '~' /   /
        '._ '-=-' _.'
           '-----'
`

func printParserErrors(out io.Writer, errors []string) {
	io.WriteString(out, MONKEY_FACE)                                       // nolint
	io.WriteString(out, "Woops! We ran into some monkey business here!\n") // nolint
	io.WriteString(out, " parser errors:\n")                               // nolint
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg+"\n") // nolint
	}
}
