package ast_test

import (
	"monkey-go/lexer"
	"monkey-go/parser"
	"testing"
)

func TestForNodeString(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"for (i = 0; i < 3; i++) { i }", "for ((i = 0); (i < 3); (i++)) {i}"},
		{"for (; true;) {}", "for (; true; ) {}"},
		{"for (let i = 0; i < 3; i = i + 1) {}", "for (let i = 0; (i < 3); (i = (i + 1))) {}"},
		{"for (x in range(2, 5)) { x }", "for (x in range(2, 5)) {x}"},
		{"for (value in [1, 2]) {}", "for (value in [1, 2]) {}"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			p := parser.New(lexer.New(tt.input))
			program := p.ParseProgram()
			if len(p.Errors()) != 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}
			if got := program.String(); got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}
