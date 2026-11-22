package repl

import (
	"bufio"
	"fmt"
	"io"
	"monkey-go/diagnostic"
	"monkey-go/evaluator"
	"monkey-go/lexer"
	"monkey-go/object"
	"monkey-go/parser"
	"strings"
)

const PROMPT = ">> "
const CONTINUATION_PROMPT = ".. "

// Start preserves the original API with a combined output stream.
func Start(in io.Reader, out io.Writer) {
	StartWithErrorWriter(in, out, out)
}

// StartWithErrorWriter sends parser, runtime and input diagnostics to errOut.
// Prompts and evaluated values use out. The print builtin retains its existing
// process-stdout behavior; builtin IO routing is a separate runtime facility.
func StartWithErrorWriter(in io.Reader, out, errOut io.Writer) {
	scanner := bufio.NewScanner(in)
	env := object.NewEnvironment()
	var input strings.Builder
	submission := 0
	for {
		source, done, err := nextREPLSource(scanner, out, &input)
		if err != nil {
			_ = diagnostic.Write(errOut, diagnostic.Diagnostic{Kind: diagnostic.IO, Message: err.Error()})
			return
		}
		if source != "" {
			submission++
			name := fmt.Sprintf("<repl:%d>", submission)
			if !evaluateREPLSource(name, source, env, out, errOut) {
				return
			}
		}
		if done {
			return
		}
	}
}

func evaluateREPLSource(name, source string, env *object.Environment, out, errOut io.Writer) bool {
	p := parser.New(lexer.NewWithFilename(name, source))
	program := p.ParseProgram()
	if len(p.Diagnostics()) != 0 {
		return printParserErrors(errOut, p.Diagnostics())
	}
	value := evaluator.Eval(program, env)
	if err, ok := value.(*object.Error); ok {
		return diagnostic.Write(errOut, err.Diagnostic()) == nil
	}
	if value == nil {
		return true
	}
	if _, err := io.WriteString(out, value.Inspect()+"\n"); err != nil {
		_ = diagnostic.Write(errOut, diagnostic.Diagnostic{Kind: diagnostic.IO, Message: err.Error()})
		return false
	}
	return true
}

func printParserErrors(out io.Writer, errors []parser.Diagnostic) bool {
	if _, err := io.WriteString(out, "parser errors:\n"); err != nil {
		return false
	}
	for _, item := range errors {
		if err := diagnostic.Write(out, item); err != nil {
			return false
		}
	}
	return true
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
