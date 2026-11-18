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
	var input strings.Builder
	for {
		source, done := nextREPLSource(scanner, out, &input)
		if done {
			return
		}
		if source == "" {
			continue
		}
		if !evaluateREPLSource(source, env, out) {
			return
		}
	}
}

func nextREPLSource(scanner *bufio.Scanner, out io.Writer, input *strings.Builder) (string, bool) {
	prompt := PROMPT
	if input.Len() > 0 {
		prompt = CONTINUATION_PROMPT
	}
	if _, err := io.WriteString(out, prompt); err != nil {
		return "", true
	}

	if !scanner.Scan() {
		if input.Len() > 0 {
			_, _ = io.WriteString(out, "incomplete input\n")
		}
		return "", true
	}
	line := scanner.Text()
	if input.Len() == 0 && strings.EqualFold(strings.TrimSpace(line), "exit") {
		return "", true
	}
	input.WriteString(line)
	input.WriteByte('\n')
	if needsMoreInput(input.String()) {
		return "", false
	}

	source := input.String()
	input.Reset()
	return source, false
}

func evaluateREPLSource(source string, env *object.Environment, out io.Writer) bool {
	p := parser.New(lexer.New(source))
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		printParserErrors(out, p.Errors())
		return true
	}
	value := evaluator.Eval(program, env)
	if value == nil {
		return true
	}
	_, err := io.WriteString(out, value.Inspect()+"\n")
	return err == nil
}

// Keep reading when an unclosed delimiter requires another input line.
// Braces inside strings and // comments do not affect the balance.
func needsMoreInput(source string) bool {
	braces, parens, brackets := 0, 0, 0
	inString, lineComment := false, false

	for i := 0; i < len(source); i++ {
		ch := source[i]
		if lineComment {
			if ch == '\n' {
				lineComment = false
			}
			continue
		}
		if inString {
			if ch == '"' {
				inString = false
			}
			continue
		}
		if ch == '"' {
			inString = true
			continue
		}
		if ch == '/' && i+1 < len(source) && source[i+1] == '/' {
			lineComment = true
			i++
			continue
		}
		switch ch {
		case '{':
			braces++
		case '}':
			braces--
		case '(':
			parens++
		case ')':
			parens--
		case '[':
			brackets++
		case ']':
			brackets--
		}
	}
	return braces > 0 || parens > 0 || brackets > 0 || inString
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
