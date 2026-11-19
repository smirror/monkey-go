package ast_test

import (
	"monkey-go/lexer"
	"monkey-go/parser"
	"testing"
)

func TestForInNodeString(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"for (x in range(2, 5)) { x }", "for (x in range(2, 5)) {x}"},
		{"for (value in [1, 2]) {}", "for (value in [1, 2]) {}"},
		{"for (value in items) { print(value); }", "for (value in items) {print(value)}"},
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
