package repl

import (
	"bytes"
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
