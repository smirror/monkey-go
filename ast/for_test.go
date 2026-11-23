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
		{"for x in range(2, 5) { x }", "for x in range(2, 5) {x}"},
		{"for value in [1, 2] {}", "for value in [1, 2] {}"},
		{"for value in items { print(value); }", "for value in items {print(value)}"},
		{"for i in [] { break; }", "for i in [] {break;}"},
		{"for i in [] { continue; }", "for i in [] {continue;}"},
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

func TestWhileNodeString(t *testing.T) {
	for _, tt := range []struct{ input, want string }{
		{"while ready {}", "while ready {}"},
		{"while (n < 3) { n = n + 1; }", "while (n < 3) {(n = (n + 1))}"},
		{"while true { break; }", "while true {break;}"},
		{"while false { continue; }", "while false {continue;}"},
	} {
		p := parser.New(lexer.New(tt.input))
		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			t.Fatalf("parser errors: %v", p.Errors())
		}
		if got := program.String(); got != tt.want {
			t.Errorf("got %q, want %q", got, tt.want)
		}
	}
}
