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
	want := PROMPT + PROMPT + "<repl:2>:1:1: macro error: macro twice expects 1 arguments, got 0\n" +
		"  | twice();\n  | ^~~~~~~\n" + PROMPT + "8\n" + PROMPT
	if got := out.String(); got != want {
		t.Fatalf("REPL output = %q, want %q", got, want)
	}
}

func TestMacroErrorStopsOnWriteFailure(t *testing.T) {
	source := `let invalid = macro() { 10; }; invalid();`
	session := replSession{env: object.NewEnvironment(), out: io.Discard, errOut: failingMacroWriter{}}
	if session.evaluateREPLSource(source) {
		t.Fatal("REPL continued after failing to write a macro error")
	}
}

func TestMacroDiagnosticsKeepDefinitionAndCallSources(t *testing.T) {
	source := "let broken = macro() {\n  1 / 0;\n};\nbroken();\n40 + 2;\nexit\n"
	var out, errOut bytes.Buffer
	StartWithErrorWriter(strings.NewReader(source), &out, &errOut)
	if strings.Contains(out.String(), "error") || !strings.Contains(out.String(), "42\n") {
		t.Fatalf("macro failure corrupted the normal output: %q", out.String())
	}
	for _, want := range []string{
		"<repl:1>:2:5: macro error: macro broken: division by zero",
		"  |   1 / 0;\n",
		"at broken (<repl:2>:1:1)",
	} {
		if !strings.Contains(errOut.String(), want) {
			t.Errorf("macro diagnostic missing %q: %q", want, errOut.String())
		}
	}
	if strings.Contains(errOut.String(), PROMPT) || strings.Contains(errOut.String(), "42") {
		t.Fatalf("normal output leaked into macro diagnostics: %q", errOut.String())
	}
}

func TestUnquotedArgumentRuntimeErrorUsesItsSubmission(t *testing.T) {
	source := "let identity = macro(x) { quote(unquote(x)); };\nidentity(missing);\n40 + 2;\nexit\n"
	var out, errOut bytes.Buffer
	StartWithErrorWriter(strings.NewReader(source), &out, &errOut)
	want := "<repl:2>:1:10: runtime error: identifier not found: missing"
	if !strings.Contains(errOut.String(), want) {
		t.Fatalf("argument source was replaced by its macro definition: %q", errOut.String())
	}
	if !strings.Contains(out.String(), "42\n") {
		t.Fatalf("REPL did not continue after the expanded runtime error: %q", out.String())
	}
}

func TestUnlocatedMacroErrorStillUsesDiagnosticStream(t *testing.T) {
	var out, errOut bytes.Buffer
	session := replSession{out: &out, errOut: &errOut}
	if !session.printMacroError(io.ErrUnexpectedEOF) {
		t.Fatal("failed to write an unlocated macro error")
	}
	if out.Len() != 0 || errOut.String() != "macro error: unexpected EOF\n" {
		t.Fatalf("incorrect diagnostic routing: out=%q, err=%q", out.String(), errOut.String())
	}
}

type failingMacroWriter struct{}

func (failingMacroWriter) Write([]byte) (int, error) {
	return 0, io.ErrClosedPipe
}
