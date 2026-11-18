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
	var balance inputBalance
	for i := 0; i < len(source); i++ {
		var next byte
		if i+1 < len(source) {
			next = source[i+1]
		}
		if balance.consume(source[i], next) {
			i++ // Skip the second '/' in a line-comment marker.
		}
	}
	return balance.braces > 0 || balance.parens > 0 || balance.brackets > 0 || balance.inString
}

type inputBalance struct {
	braces      int
	parens      int
	brackets    int
	inString    bool
	lineComment bool
}

// consume returns true when it consumed both bytes of a // comment marker.
func (b *inputBalance) consume(ch, next byte) bool {
	if b.lineComment {
		if ch == '\n' {
			b.lineComment = false
		}
		return false
	}
	if b.inString {
		if ch == '"' {
			b.inString = false
		}
		return false
	}
	if ch == '"' {
		b.inString = true
		return false
	}
	if ch == '/' && next == '/' {
		b.lineComment = true
		return true
	}
	b.trackDelimiter(ch)
	return false
}

func (b *inputBalance) trackDelimiter(ch byte) {
	switch ch {
	case '{':
		b.braces++
	case '}':
		b.braces--
	case '(':
		b.parens++
	case ')':
		b.parens--
	case '[':
		b.brackets++
	case ']':
		b.brackets--
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
