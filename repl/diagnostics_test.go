package repl

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestREPLSeparatesDiagnosticsAndContinues(t *testing.T) {
	var out, errOut bytes.Buffer
	input := "let = 1\n1 / 0\n40 + 2\nexit\n"
	StartWithErrorWriter(strings.NewReader(input), &out, &errOut)
	if !strings.Contains(out.String(), "42\n") || strings.Contains(out.String(), "error") {
		t.Fatalf("incorrect normal output: %q", out.String())
	}
	for _, want := range []string{
		"<repl:1>:1:5: syntax error:",
		"<repl:2>:1:3: runtime error: division by zero",
		"  | 1 / 0\n  |   ^\n",
	} {
		if !strings.Contains(errOut.String(), want) {
			t.Errorf("diagnostics missing %q: %q", want, errOut.String())
		}
	}
	if strings.Contains(errOut.String(), "42") || strings.Contains(errOut.String(), PROMPT) {
		t.Errorf("normal output leaked into diagnostics: %q", errOut.String())
	}
}

func TestREPLDiagnosticUsesEarlierMultilineDefinition(t *testing.T) {
	input := "let fail = fn() {\n  let s = \"日本\";\n  1 / 0\n};\nfail()\n40 + 2\nexit\n"
	var out, errOut bytes.Buffer
	StartWithErrorWriter(strings.NewReader(input), &out, &errOut)
	for _, want := range []string{
		"<repl:1>:3:5: runtime error: division by zero",
		"  |   1 / 0\n  |     ^\n",
		"at fail (<repl:2>:1:1)",
	} {
		if !strings.Contains(errOut.String(), want) {
			t.Errorf("diagnostics missing %q: %q", want, errOut.String())
		}
	}
	if !strings.Contains(out.String(), "42\n") {
		t.Fatalf("REPL did not continue: %q", out.String())
	}
}

func TestREPLMarksErrorInlineOnUnicodeSourceLine(t *testing.T) {
	var out, errOut bytes.Buffer
	StartWithErrorWriter(strings.NewReader("\"日本\" + 1\nexit\n"), &out, &errOut)
	want := "<repl:1>:1:6: runtime error: type mismatch: STRING + INTEGER\n" +
		"  | \"日本\" ⟦+⟧ 1\n"
	if got := errOut.String(); got != want {
		t.Fatalf("got diagnostic %q, want %q", got, want)
	}
}

func TestREPLLessOrEqualAliasUsesOriginalOperatorSpan(t *testing.T) {
	var out, errOut bytes.Buffer
	StartWithErrorWriter(strings.NewReader("1 =< true\nexit\n"), &out, &errOut)
	want := "<repl:1>:1:3: runtime error: type mismatch: INTEGER <= BOOLEAN\n" +
		"  | 1 =< true\n  |   ^~\n"
	if got := errOut.String(); got != want {
		t.Fatalf("got diagnostic %q, want %q", got, want)
	}
}

func TestREPLIncompleteInputIsDiagnosedAtEOF(t *testing.T) {
	for _, input := range []string{"let f = fn() {\n", "\"日本\n", "[1,\n"} {
		t.Run(input, func(t *testing.T) {
			var out, errOut bytes.Buffer
			StartWithErrorWriter(strings.NewReader(input), &out, &errOut)
			if !strings.Contains(errOut.String(), "syntax error:") || !strings.Contains(errOut.String(), "<repl:1>:") {
				t.Fatalf("unfinished input lacks source diagnostic: %q", errOut.String())
			}
			if strings.Contains(out.String(), "error") || strings.Contains(out.String(), "incomplete") {
				t.Fatalf("input error leaked into normal output: %q", out.String())
			}
		})
	}
}

type failedReader struct{}

func (failedReader) Read([]byte) (int, error) { return 0, errors.New("reader failed") }

type failedWriter struct{}

func (failedWriter) Write([]byte) (int, error) { return 0, errors.New("writer failed") }

func TestREPLReadFailureIsAnInputDiagnostic(t *testing.T) {
	var out, errOut bytes.Buffer
	StartWithErrorWriter(failedReader{}, &out, &errOut)
	if errOut.String() != "io error: reader failed\n" {
		t.Fatalf("incorrect read error: %q", errOut.String())
	}
}

func TestREPLStopsWhenDiagnosticWriterFails(t *testing.T) {
	var out bytes.Buffer
	StartWithErrorWriter(strings.NewReader("1 / 0\n42\n"), &out, failedWriter{})
	if strings.Contains(out.String(), "42") {
		t.Fatalf("REPL continued after output failed: %q", out.String())
	}
}

func TestREPLReportsPromptOutputFailure(t *testing.T) {
	var errOut bytes.Buffer
	StartWithErrorWriter(strings.NewReader("42\n"), failedWriter{}, &errOut)
	if errOut.String() != "io error: writer failed\n" {
		t.Fatalf("incorrect prompt error: %q", errOut.String())
	}
}

type failAfterPrompt struct {
	writes int
}

func (w *failAfterPrompt) Write(data []byte) (int, error) {
	w.writes++
	if w.writes == 1 {
		return len(data), nil
	}
	return 0, errors.New("result output failed")
}

func TestREPLReportsValueOutputFailureAndStops(t *testing.T) {
	var errOut bytes.Buffer
	out := &failAfterPrompt{}
	StartWithErrorWriter(strings.NewReader("42\n1 / 0\n"), out, &errOut)
	if errOut.String() != "io error: result output failed\n" {
		t.Fatalf("incorrect output failure diagnostic: %q", errOut.String())
	}
	if out.writes != 2 {
		t.Fatalf("REPL continued after result output failure: %d writes", out.writes)
	}
}

func TestREPLNormalEOFIsSilent(t *testing.T) {
	var errOut bytes.Buffer
	StartWithErrorWriter(strings.NewReader(""), io.Discard, &errOut)
	if errOut.Len() != 0 {
		t.Fatalf("normal EOF was reported as an error: %q", errOut.String())
	}
}
