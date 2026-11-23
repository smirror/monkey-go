// Package diagnostic formats structured language errors without evaluating code.
package diagnostic

import (
	"fmt"
	"io"
	"monkey-go/token"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Kind string

const (
	Syntax  Kind = "syntax"
	Runtime Kind = "runtime"
	IO      Kind = "io"
)

// Frame identifies an entered function's call site, not its declaration.
type Frame struct {
	Name string
	Span token.Span
}

// Diagnostic holds data independently of its terminal representation.
// Stack is ordered from the innermost call to the outermost call.
type Diagnostic struct {
	Kind    Kind
	Message string
	Span    token.Span
	Stack   []Frame
}

// Write renders one diagnostic and reports an output failure to its caller.
func Write(out io.Writer, d Diagnostic) error {
	_, err := io.WriteString(out, Format(d))
	return err
}

// Format uses one-based rune columns. ASCII excerpts use carets; Unicode
// excerpts mark the range inline without relying on terminal cell widths.
func Format(d Diagnostic) string {
	var out strings.Builder
	kind := string(d.Kind)
	if kind != "" {
		kind += " "
	}
	if where := location(d.Span); where != "" {
		out.WriteString(where + ": ")
	}
	out.WriteString(kind + "error: " + displayText(d.Message) + "\n")
	out.WriteString(excerpt(d.Span))
	for _, frame := range d.Stack {
		out.WriteString("  at " + displayText(frame.Name))
		if where := location(frame.Span); where != "" {
			out.WriteString(" (" + where + ")")
		}
		out.WriteByte('\n')
	}
	return out.String()
}

func location(span token.Span) string {
	if span.Start.Line < 1 || span.Start.Column < 1 {
		return ""
	}
	name := "<input>"
	if span.Source != nil && span.Source.Name != "" {
		name = displayText(span.Source.Name)
	}
	return fmt.Sprintf("%s:%d:%d", name, span.Start.Line, span.Start.Column)
}

func excerpt(span token.Span) string {
	line, start, end, ok := sourceLine(span)
	if !ok {
		return ""
	}
	if !isASCII(line) {
		return "  | " + displayText(line[:start]) + "⟦" + displayText(line[start:end]) +
			"⟧" + displayText(line[end:]) + "\n"
	}
	indent := len(displayText(line[:start]))
	marked := len(displayText(line[start:end]))
	if marked < 1 {
		marked = 1 // EOF still needs a visible caret.
	}
	return "  | " + displayText(line) + "\n  | " + strings.Repeat(" ", indent) +
		"^" + strings.Repeat("~", marked-1) + "\n"
}

func isASCII(text string) bool {
	for i := 0; i < len(text); i++ {
		if text[i] >= utf8.RuneSelf {
			return false
		}
	}
	return true
}

// sourceLine clips a multi-line span to its first source line. Invalid or absent
// metadata from a hand-built AST is harmless: only the message will be shown.
func sourceLine(span token.Span) (string, int, int, bool) {
	if !validSourceStart(span) {
		return "", 0, 0, false
	}
	text := span.Source.Text
	start := span.Start.Offset
	lineStart, lineEnd := lineBounds(text, start)
	end := clipOffset(span.End.Offset, start, lineEnd)
	if !runeBoundary(text, start) || !runeBoundary(text, end) {
		return "", 0, 0, false
	}
	return text[lineStart:lineEnd], start - lineStart, end - lineStart, true
}

func validSourceStart(span token.Span) bool {
	if span.Source == nil {
		return false
	}
	if span.Start.Line < 1 || span.Start.Column < 1 {
		return false
	}
	return span.Start.Offset >= 0 && span.Start.Offset <= len(span.Source.Text)
}

func lineBounds(text string, start int) (int, int) {
	lineStart := strings.LastIndexAny(text[:start], "\r\n") + 1
	lineEnd := len(text)
	if next := strings.IndexAny(text[start:], "\r\n"); next >= 0 {
		lineEnd = start + next
	}
	return lineStart, lineEnd
}

func clipOffset(offset, start, end int) int {
	if offset < start {
		return start
	}
	if offset > end {
		return end
	}
	return offset
}

func runeBoundary(text string, offset int) bool {
	return offset == len(text) || utf8.RuneStart(text[offset])
}

func displayText(text string) string {
	var out strings.Builder
	for _, ch := range text {
		switch {
		case ch == '\t':
			out.WriteString("    ")
		case unicode.IsControl(ch):
			escaped := strconv.QuoteRuneToASCII(ch)
			out.WriteString(escaped[1 : len(escaped)-1])
		default:
			out.WriteRune(ch)
		}
	}
	return out.String()
}
