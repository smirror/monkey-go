package repl

import (
	"bufio"
	"io"
	"strings"
)

// Read one complete REPL input, including optional multiline constructs.
func nextREPLSource(scanner *bufio.Scanner, out io.Writer, input *strings.Builder) (string, bool, error) {
	prompt := PROMPT
	if input.Len() > 0 {
		prompt = CONTINUATION_PROMPT
	}
	if _, err := io.WriteString(out, prompt); err != nil {
		return "", true, err
	}

	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return "", true, err
		}
		// Let the parser diagnose unfinished input at EOF with a source span.
		return input.String(), true, nil
	}
	line := scanner.Text()
	if input.Len() == 0 && strings.EqualFold(strings.TrimSpace(line), "exit") {
		return "", true, nil
	}
	input.WriteString(line)
	input.WriteByte('\n')
	if needsMoreInput(input.String()) {
		return "", false, nil
	}

	source := input.String()
	input.Reset()
	return source, false, nil
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
		b.lineComment = ch != '\n' && ch != '\r'
		return false
	}
	if b.inString {
		b.inString = ch != '"'
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
