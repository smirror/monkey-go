package repl

import (
	"bytes"
	"strings"
	"testing"
)

func TestREPLMultilineBreakAndContinue(t *testing.T) {
	input := "let sum = 0;\nfor i in range(1, 6) {\n  if (i == 2) { continue; }\n  if (i == 4) { break; }\n  sum = sum + i;\n}\nsum;\nexit\n"
	var out, errOut bytes.Buffer
	StartWithErrorWriter(strings.NewReader(input), &out, &errOut)
	if errOut.Len() != 0 {
		t.Fatalf("valid loop produced diagnostics: %q", errOut.String())
	}
	if !strings.Contains(out.String(), "4\n") || !strings.Contains(out.String(), CONTINUATION_PROMPT) {
		t.Fatalf("multiline loop did not finish with sum 4: %q", out.String())
	}
}

func TestREPLContinuesAfterOutsideLoopControl(t *testing.T) {
	var out, errOut bytes.Buffer
	StartWithErrorWriter(strings.NewReader("break;\ncontinue;\n40 + 2\nexit\n"), &out, &errOut)
	for _, want := range []string{
		"<repl:1>:1:1: syntax error: break outside loop",
		"<repl:2>:1:1: syntax error: continue outside loop",
	} {
		if !strings.Contains(errOut.String(), want) {
			t.Errorf("missing control diagnostic %q: %q", want, errOut.String())
		}
	}
	if strings.Contains(out.String(), "error") || !strings.Contains(out.String(), "42\n") {
		t.Fatalf("REPL failed to separate diagnostics and resume: %q", out.String())
	}
}

func TestREPLUnfinishedLoopControlBodyReportsEOF(t *testing.T) {
	for _, keyword := range []string{"break", "continue"} {
		input := "for i in range(2) {\n  " + keyword
		var out, errOut bytes.Buffer
		StartWithErrorWriter(strings.NewReader(input), &out, &errOut)
		if !strings.Contains(errOut.String(), "syntax error: unterminated block") {
			t.Fatalf("unfinished loop was not diagnosed: %q", errOut.String())
		}
		if strings.Contains(errOut.String(), "outside loop") || strings.Contains(out.String(), "error") {
			t.Fatalf("unfinished loop lost its context or stream: out=%q err=%q", out.String(), errOut.String())
		}
	}
}
