package object

import (
	"monkey-go/ast"
	"testing"
)

func TestMacroObjects(t *testing.T) {
	parameter := &ast.Identifier{Value: "input"}
	tests := []struct {
		name    string
		value   Object
		kind    ObjectType
		display string
	}{
		{
			name:    "quoted syntax",
			value:   &Quote{Node: &ast.InfixExpression{Left: parameter, Operator: "+", Right: parameter}},
			kind:    QUOTE_OBJ,
			display: "QUOTE((input + input))",
		},
		{
			name: "macro definition",
			value: &Macro{
				Parameters: []*ast.Identifier{parameter},
				Body: &ast.BlockStatement{Statements: []ast.Statement{
					&ast.ExpressionStatement{Expression: parameter},
				}},
				Env: NewEnvironment(),
			},
			kind:    MACRO_OBJ,
			display: "macro(1 parameters) {input}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.value.Type(); got != tt.kind {
				t.Errorf("Type() = %q, want %q", got, tt.kind)
			}
			if got := tt.value.Inspect(); got != tt.display {
				t.Errorf("Inspect() = %q, want %q", got, tt.display)
			}
		})
	}
}
