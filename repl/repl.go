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
	session := replSession{env: object.NewEnvironment(), out: out, errOut: errOut}
	var input strings.Builder
	for {
		source, done, err := nextREPLSource(scanner, out, &input)
		if err != nil {
			session.reportIOError(err)
			return
		}
		if !session.evaluateREPLSource(source) || done {
			return
		}
	}
}

type replSession struct {
	env, macroEnv *object.Environment
	out, errOut   io.Writer
	submission    int
}

func (s *replSession) evaluateREPLSource(source string) bool {
	if source == "" {
		return true
	}
	s.submission++
	name := fmt.Sprintf("<repl:%d>", s.submission)
	p := parser.New(lexer.NewWithFilename(name, source))
	program := p.ParseProgram()
	if len(p.Diagnostics()) != 0 {
		return printParserErrors(s.errOut, p.Diagnostics())
	}
	if s.macroEnv == nil {
		s.macroEnv = object.NewEnvironment()
	}
	evaluator.DefineMacros(program, s.macroEnv)
	expanded, err := evaluator.ExpandMacros(program, s.macroEnv)
	if err != nil {
		return s.printMacroError(err)
	}
	return s.printValue(evaluator.Eval(expanded, s.env))
}

func (s *replSession) printMacroError(err error) bool {
	report := diagnostic.Diagnostic{Kind: diagnostic.Macro, Message: err.Error()}
	if located, ok := err.(*evaluator.MacroExpansionError); ok {
		report = located.Diagnostic()
	}
	return diagnostic.Write(s.errOut, report) == nil
}

func (s *replSession) printValue(value object.Object) bool {
	if err, ok := value.(*object.Error); ok {
		return diagnostic.Write(s.errOut, err.Diagnostic()) == nil
	}
	if value == nil {
		return true
	}
	if _, err := io.WriteString(s.out, value.Inspect()+"\n"); err != nil {
		s.reportIOError(err)
		return false
	}
	return true
}

func (s *replSession) reportIOError(err error) {
	_ = diagnostic.Write(s.errOut, diagnostic.Diagnostic{Kind: diagnostic.IO, Message: err.Error()})
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
