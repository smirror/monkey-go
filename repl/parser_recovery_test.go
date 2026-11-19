package repl

import (
	"bytes"
	"strings"
	"testing"
)

func TestREPLContinuesAfterParserError(t *testing.T) {
	var output bytes.Buffer
	Start(strings.NewReader("let = 1\n40 + 2\nexit\n"), &output)
	if !strings.Contains(output.String(), "parser errors:") {
		t.Fatalf("expected a parser error, got %q", output.String())
	}
	if !strings.Contains(output.String(), "42\n") {
		t.Fatalf("REPL did not evaluate the following valid input: %q", output.String())
	}
}

func TestREPLReturnWithoutSemicolonDoesNotHang(t *testing.T) {
	var output bytes.Buffer
	Start(strings.NewReader("return 42\nexit\n"), &output)
	if !strings.Contains(output.String(), "42\n") {
		t.Fatalf("expected return result, got %q", output.String())
	}
}
