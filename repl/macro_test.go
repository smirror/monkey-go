package repl

import (
	"bytes"
	"io"
	"monkey-go/object"
	"strings"
	"testing"
)

func TestMacrosPersistBetweenREPLInputs(t *testing.T) {
	source := `let twice = macro(x) { x + x; };
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
	source := `let twice = macro(x) { x + x; };
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
	source := `let invalid = macro(x) { x; }; invalid();`
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
		"<repl:1>:2:5: runtime error: division by zero",
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

func TestMacroArgumentRuntimeErrorUsesItsSubmission(t *testing.T) {
	source := "let identity = macro(x) { x; };\nidentity(missing);\n40 + 2;\nexit\n"
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

func TestTemplateMacroREPLBehavior(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{
			"reverse expression arguments",
			"let reverse = macro(a, b) { b - a; };\nreverse(2 + 2, 10 - 5);\n",
			"1",
		},
		{
			"multiple statements use the caller scope",
			"let remember = macro(value) { let saved = value; saved + 1; };\nremember(40);\nsaved;\n",
			"41\n" + PROMPT + "40",
		},
		{
			"free names use the caller function scope",
			"let add = macro(value) { value + offset; };\nlet run = fn(offset) { add(2); };\nrun(40);\n",
			"42",
		},
		{
			"return exits the caller function",
			"let stop = macro(value) { return value; };\nlet run = fn() { stop(7); 99; };\nrun();\n",
			"7",
		},
		{
			"unused arguments are not expanded",
			"let recursive = macro() { recursive(); };\nlet ignore = macro(value) { 42; };\nignore(recursive());\n",
			"42",
		},
		{
			"unused arguments are not evaluated",
			"let ignore = macro(value) { 42; };\nlet counter = 0;\nignore(counter = counter + 1);\ncounter;\n",
			"42\n" + PROMPT + "0",
		},
		{
			"each inserted argument is evaluated",
			"let twice = macro(value) { value + value; };\nlet counter = 0;\ntwice(counter = counter + 1);\ncounter;\n",
			"3\n" + PROMPT + "2",
		},
		{
			"if selects runtime expressions",
			"let choose = macro(test, yes, no) { if (test) { yes; } else { no; }; };\nchoose(false, missing, 8);\n",
			"8",
		},
		{
			"for runs inside the template body",
			"let sum = macro(values) { let total = 0; for value in values { total = total + value; } total; };\nsum([1, 2, 3]);\n",
			"6",
		},
		{
			"quote and unquote are ordinary function names",
			"let quote = fn(value) { value + 1; };\nlet unquote = fn(value) { value * 2; };\nquote(unquote(3));\n",
			"7",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { checkTemplateREPL(t, tt.source, tt.want) })
	}
}

func checkTemplateREPL(t *testing.T, source, want string) {
	t.Helper()
	var out, errOut bytes.Buffer
	StartWithErrorWriter(strings.NewReader(source+"exit\n"), &out, &errOut)
	if errOut.Len() != 0 {
		t.Fatalf("unexpected macro diagnostic: %s", errOut.String())
	}
	if !strings.HasSuffix(out.String(), want+"\n"+PROMPT) {
		t.Fatalf("REPL output = %q, want final result %q", out.String(), want)
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
