package repl

import (
	"bufio"
	"io"
	"monkey-go/evaluator"
	"monkey-go/lexer"
	"monkey-go/object"
	"monkey-go/parser"
	"strings"
)

const PROMPT = ">> "

const CONTINUATION_PROMPT = ".. "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	env := object.NewEnvironment()
	macroEnv := object.NewEnvironment()
	var input strings.Builder
	for {
		source, done := nextREPLSource(scanner, out, &input)
		if done {
			return
		}
		if source == "" {
			continue
		}
		if !evaluateREPLSource(source, env, macroEnv, out) {
			return
		}
	}
}

func evaluateREPLSource(source string, env, macroEnv *object.Environment, out io.Writer) bool {
	p := parser.New(lexer.New(source))
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		printParserErrors(out, p.Errors())
		return true
	}
	evaluator.DefineMacros(program, macroEnv)
	expanded, expansionErr := evaluator.ExpandMacros(program, macroEnv)
	if expansionErr != nil {
		_, err := io.WriteString(out, "MACRO ERROR: "+expansionErr.Error()+"\n")
		return err == nil
	}
	value := evaluator.Eval(expanded, env)
	if value == nil {
		return true
	}
	_, err := io.WriteString(out, value.Inspect()+"\n")
	return err == nil
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
