package repl

import (
	"bytes"
	"strings"
	"testing"
)

func TestMultilineArrayForInREPL(t *testing.T) {
	input := "let total = 0;\nfor (i in [1, 2, 3]) {\n  total = total + i;\n}\ntotal;\nexit\n"
	var out bytes.Buffer
	Start(strings.NewReader(input), &out)
	if got := out.String(); !strings.Contains(got, "6\n") {
		t.Fatalf("expected total 6 in REPL output, got %q", got)
	}
	if !strings.Contains(out.String(), CONTINUATION_PROMPT) {
		t.Fatalf("REPL did not prompt for continued input: %q", out.String())
	}
}

func TestMultilineForInREPL(t *testing.T) {
	input := "let total = 0;\nfor (i in range(2, 5)) {\n  total = total + i;\n} // result: 9\ntotal;\nexit\n"
	var out bytes.Buffer
	Start(strings.NewReader(input), &out)
	if got := out.String(); !strings.Contains(got, "9\n") {
		t.Fatalf("expected total 9 in REPL output, got %q", got)
	}
}

func TestInputBalanceIgnoresStringsAndComments(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"for (i in range(3)) {\n", true},
		{"for (i in range(3)) {\n i;\n}\n", false},
		{"for (i in [\n1, 2]) { }\n", false},
		{"for (i in range(2)) { // }\n", true},
		{"for (i in [\"{\", \"}\"]) { }\n", false},
		{"for (i in range(2)) {\n i;\n} // {\n", false},
	}
	for _, tt := range tests {
		if got := needsMoreInput(tt.input); got != tt.want {
			t.Errorf("needsMoreInput(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
