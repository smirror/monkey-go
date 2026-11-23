package token

import "testing"

func TestTokenTypeUniqueness(t *testing.T) {
	tokens := []struct {
		name  string
		value TokenType
	}{
		{"ILLEGAL", ILLEGAL},
		{"EOF", EOF},
		{"IDENT", IDENT},
		{"INT", INT},
		{"STRING", STRING},
		{"ASSIGN", ASSIGN},
		{"PLUS", PLUS},
		{"MINUS", MINUS},
		{"BANG", BANG},
		{"ASTERISK", ASTERISK},
		{"SLASH", SLASH},
		{"MODULO", MODULO},
		{"LT", LT},
		{"GT", GT},
		{"LTE", LTE},
		{"GTE", GTE},
		{"EQ", EQ},
		{"NOT_EQ", NOT_EQ},
		{"AND", AND},
		{"OR", OR},
		{"COMMA", COMMA},
		{"SEMICOLON", SEMICOLON},
		{"COLON", COLON},
		{"LPAREN", LPAREN},
		{"RPAREN", RPAREN},
		{"LBRACE", LBRACE},
		{"RBRACE", RBRACE},
		{"LBRACKET", LBRACKET},
		{"RBRACKET", RBRACKET},
		{"FUNCTION", FUNCTION},
		{"LET", LET},
		{"TRUE", TRUE},
		{"FALSE", FALSE},
		{"IF", IF},
		{"ELSE", ELSE},
		{"RETURN", RETURN},
		{"FOR", FOR},
		{"IN", IN},
	}

	seen := make(map[TokenType]string)
	for _, tt := range tokens {
		if tt.value == "" {
			t.Errorf("token %s has empty value", tt.name)
		}
		if prev, ok := seen[tt.value]; ok {
			t.Errorf("token %s has same value %q as %s", tt.name, tt.value, prev)
		}
		seen[tt.value] = tt.name
	}
}
