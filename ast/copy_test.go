package ast_test

import (
	"monkey-go/ast"
	"reflect"
	"testing"
)

func TestModifyCopiesLiteralStorage(t *testing.T) {
	program := parseModifyProgram(t, `[name, 1, true, "text"]`)
	array := program.Statements[0].(*ast.ExpressionStatement).Expression.(*ast.ArrayLiteral)
	for _, original := range array.Elements {
		copied := ast.Modify(original, func(node ast.Node) ast.Node { return node })
		if copied == original {
			t.Fatalf("rewriting reused a mutable %T", original)
		}
		if !reflect.DeepEqual(copied, original) {
			t.Fatalf("copy lost literal values or source: %#v", copied)
		}
	}
}

func TestModifyInPlaceCallbackDoesNotChangeOriginal(t *testing.T) {
	original := parseModifyProgram(t, "1 + 1")
	rewritten := ast.Modify(original, func(node ast.Node) ast.Node {
		if integer, ok := node.(*ast.IntegerLiteral); ok {
			integer.Value = 2
			integer.Token.Literal = "2"
		}
		return node
	})
	assertNodeString(t, rewritten, "(2 + 2)")
	assertNodeString(t, original, "(1 + 1)")
}

func TestModifyCopiesParameterBindings(t *testing.T) {
	for _, source := range []string{"fn(first, second) { first; }", "macro(first, second) { first; }"} {
		original := parseModifyProgram(t, source)
		before := original.String()
		copied := ast.Modify(original, func(node ast.Node) ast.Node { return node }).(*ast.Program)
		expression := copied.Statements[0].(*ast.ExpressionStatement).Expression
		parameters := bindingParameters(expression)
		parameters[0].Value = "mutated"
		parameters[1] = &ast.Identifier{Value: "replacement"}
		assertNodeString(t, original, before)
	}
}

func TestModifyCopiesDeclarationBindings(t *testing.T) {
	original := parseModifyProgram(t, "let item = 1; for item in [1] { item; }")
	before := original.String()
	copied := ast.Modify(original, func(node ast.Node) ast.Node { return node }).(*ast.Program)
	copied.Statements[0].(*ast.LetStatement).Name.Value = "changed"
	copied.Statements[1].(*ast.ForInStatement).Variable.Value = "changed"
	assertNodeString(t, original, before)
}

func bindingParameters(node ast.Expression) []*ast.Identifier {
	switch node := node.(type) {
	case *ast.FunctionLiteral:
		return node.Parameters
	case *ast.MacroLiteral:
		return node.Parameters
	}
	return nil
}
