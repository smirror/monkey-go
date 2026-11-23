package repl

import (
	"bytes"
	"strings"
	"testing"
)

func TestREPLMultilineWhileWithBreakAndContinue(t *testing.T) {
	input := "let n = 0;\nwhile n < 5 {\n  n = n + 1;\n  if (n == 2) { continue; }\n  if (n == 3) { break; }\n}\nn;\nexit\n"
	var out, errOut bytes.Buffer
	StartWithErrorWriter(strings.NewReader(input), &out, &errOut)
	if errOut.Len() != 0 {
		t.Fatalf("valid while loop produced diagnostics: %q", errOut.String())
	}
	if !strings.Contains(out.String(), "3\n") || !strings.Contains(out.String(), CONTINUATION_PROMPT) {
		t.Fatalf("multiline while did not complete: %q", out.String())
	}
}

func TestREPLWhileErrorsKeepTheirSourceAndAllowRecovery(t *testing.T) {
	for _, tt := range []struct{ input, location string }{
		{"while 1 / 0 {}\n42\nexit\n", "<repl:1>:1:9: runtime error: division by zero"},
		{"let fail = fn() {\n  while true {\n    1 / 0;\n  }\n};\nfail()\n42\nexit\n", "<repl:1>:3:7: runtime error: division by zero"},
	} {
		var out, errOut bytes.Buffer
		StartWithErrorWriter(strings.NewReader(tt.input), &out, &errOut)
		if !strings.Contains(errOut.String(), tt.location) {
			t.Errorf("missing original while error %q: %q", tt.location, errOut.String())
		}
		if strings.Contains(out.String(), "error") || !strings.Contains(out.String(), "42\n") {
			t.Errorf("REPL failed to separate while diagnostics and resume: %q", out.String())
		}
	}
}

func TestREPLUnfinishedWhileBodyReportsEOF(t *testing.T) {
	var out, errOut bytes.Buffer
	StartWithErrorWriter(strings.NewReader("while true {\n  continue"), &out, &errOut)
	if !strings.Contains(errOut.String(), "<repl:1>:3:1: syntax error: unterminated block") {
		t.Fatalf("unfinished while was not diagnosed at EOF: %q", errOut.String())
	}
	if strings.Contains(errOut.String(), "outside loop") || strings.Contains(out.String(), "error") {
		t.Fatalf("unfinished while lost its context or stream: out=%q err=%q", out.String(), errOut.String())
	}
}
