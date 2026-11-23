package parser

import "testing"

func TestComparisonAndModuloPrecedence(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"a + b <= c % d", "((a + b) <= (c % d))"},
		{"a >= b == c < d", "((a >= b) == (c < d))"},
		{"a % b * c / d", "(((a % b) * c) / d)"},
		{"a / b % c * d", "(((a / b) % c) * d)"},
		{"a % b % c", "((a % b) % c)"},
		{"a <= b < c >= d", "(((a <= b) < c) >= d)"},
		{"a = b % c <= d || e >= f && g", "(a = (((b % c) <= d) || ((e >= f) && g)))"},
		{"-a % b", "((-a) % b)"},
		{"f(7) % a[0]", "(f(7) % (a[0]))"},
		{"5 % (2 + 1)", "(5 % (2 + 1))"},
		{"1 <= 2 && 3 >= 2", "((1 <= 2) && (3 >= 2))"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			program := parseLocatedProgram(t, tt.input)
			if got := program.String(); got != tt.want {
				t.Fatalf("parsed expression = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMalformedComparisonAndModuloDiagnostics(t *testing.T) {
	tests := []diagnosticCase{
		{"1 <= ", "no prefix parse function for EOF found", position(5, 1, 6), position(5, 1, 6)},
		{"1 >= ;", "no prefix parse function for ; found", position(5, 1, 6), position(6, 1, 7)},
		{"1 % ;", "no prefix parse function for ; found", position(4, 1, 5), position(5, 1, 6)},
		{"% 2", "no prefix parse function for % found", position(0, 1, 1), position(1, 1, 2)},
		{"1 < = 2", "no prefix parse function for = found", position(4, 1, 5), position(5, 1, 6)},
		{"1 > = 2", "no prefix parse function for = found", position(4, 1, 5), position(5, 1, 6)},
		{"1 %% 2", "no prefix parse function for % found", position(3, 1, 4), position(4, 1, 5)},
		{"1 %= 2", "no prefix parse function for = found", position(3, 1, 4), position(4, 1, 5)},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) { checkDiagnostic(t, tt) })
	}
}
