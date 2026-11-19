package repl

import (
	"bytes"
	"io"
	"monkey-go/object"
	"strings"
	"testing"
)

func TestMacrosPersistBetweenREPLInputs(t *testing.T) {
	source := `let twice = macro(x) { quote(unquote(x) + unquote(x)); };
twice(4);
exit
`
	var out bytes.Buffer
	Start(strings.NewReader(source), &out)
	if !strings.Contains(out.String(), "8\n") {
		t.Fatalf("macro evaluation missing result 8: %q", out.String())
	}
}

func TestREPLContinuesAfterMacroError(t *testing.T) {
	source := `let twice = macro(x) { quote(unquote(x) + unquote(x)); };
twice();
twice(4);
exit
`
	var out bytes.Buffer
	Start(strings.NewReader(source), &out)
	want := PROMPT + PROMPT + "MACRO ERROR: macro twice expects 1 arguments, got 0\n" + PROMPT + "8\n" + PROMPT
	if got := out.String(); got != want {
		t.Fatalf("REPL output = %q, want %q", got, want)
	}
}

func TestMacroErrorStopsOnWriteFailure(t *testing.T) {
	source := `let invalid = macro() { 10; }; invalid();`
	if evaluateREPLSource(source, object.NewEnvironment(), object.NewEnvironment(), failingMacroWriter{}) {
		t.Fatal("REPL continued after failing to write a macro error")
	}
}

type failingMacroWriter struct{}

func (failingMacroWriter) Write([]byte) (int, error) {
	return 0, io.ErrClosedPipe
}
