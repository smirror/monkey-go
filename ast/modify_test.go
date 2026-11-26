package ast_test

import (
	"fmt"
	"monkey-go/ast"
	"monkey-go/lexer"
	"monkey-go/parser"
	"monkey-go/token"
	"reflect"
	"strconv"
	"testing"
)

func TestModifyTraversesChildrenWithoutChangingOriginal(t *testing.T) {
	cases := []struct {
		source string
		want   string
	}{
		{"1", "2"},
		{"-1", "-2"},
		{"1 + 1", "2 + 2"},
		{"let value = 1;", "let value = 2;"},
		{"return 1;", "return 2;"},
		{"if (1 < 1) { 1; } else { 1; }", "if (2 < 2) { 2; } else { 2; }"},
		{"if (true) { 1; }", "if (true) { 2; }"},
		{"fn(x) { 1; }", "fn(x) { 2; }"},
		{"macro(x) { 1 + x; }", "macro(x) { 2 + x; }"},
		{"call(1, 1)", "call(2, 2)"},
		{"fn(x) { 1; }(1)", "fn(x) { 2; }(2)"},
		{"[1, 1]", "[2, 2]"},
		{"[1][1]", "[2][2]"},
		{"{1: 1}", "{2: 2}"},
		{"for item in [1] { item + 1; }", "for item in [2] { item + 2; }"},
		{"true; false; \"text\"; name;", "true; false; \"text\"; name;"},
	}
	for _, tc := range cases {
		t.Run(tc.source, func(t *testing.T) {
			original := parseModifyProgram(t, tc.source)
			before := original.String()
			rewritten := ast.Modify(original, replaceOneWith(2))
			want := parseModifyProgram(t, tc.want).String()
			assertNodeString(t, rewritten, want)
			assertNodeString(t, original, before)
		})
	}
}

func TestModifyVisitsChildrenBeforeParents(t *testing.T) {
	original := parseModifyProgram(t, "1 + 1")
	var visits []string
	rewritten := ast.Modify(original, func(node ast.Node) ast.Node {
		visits = append(visits, fmt.Sprintf("%T: %s", node, node.String()))
		return replaceOneWith(2)(node)
	})
	want := []string{
		"*ast.IntegerLiteral: 1",
		"*ast.IntegerLiteral: 1",
		"*ast.InfixExpression: (2 + 2)",
		"*ast.ExpressionStatement: (2 + 2)",
		"*ast.Program: (2 + 2)",
	}
	if !reflect.DeepEqual(visits, want) {
		t.Fatalf("visits = %v, want %v", visits, want)
	}
	assertNodeString(t, rewritten, "(2 + 2)")
}

func TestModifyPreservesBindingNames(t *testing.T) {
	original := parseModifyProgram(t, `
let name = fn(name) { name; };
for name in name { name; }
macro(name) { name; };
`)
	rewritten := ast.Modify(original, func(node ast.Node) ast.Node {
		identifier, ok := node.(*ast.Identifier)
		if !ok {
			return node
		}
		return &ast.Identifier{Token: identifier.Token, Value: "replaced"}
	})
	want := parseModifyProgram(t, `
let name = fn(name) { replaced; };
for name in replaced { replaced; }
macro(name) { replaced; };
`)
	assertNodeString(t, rewritten, want.String())
}

func TestModifyCopiesCollectionStorage(t *testing.T) {
	original := parseModifyProgram(t, "fn() { call([1], {1: 1}); }")
	before := original.String()
	rewritten := ast.Modify(original, func(node ast.Node) ast.Node { return node }).(*ast.Program)
	function := rewritten.Statements[0].(*ast.ExpressionStatement).Expression.(*ast.FunctionLiteral)
	call := function.Body.Statements[0].(*ast.ExpressionStatement).Expression.(*ast.CallExpression)
	array := call.Arguments[0].(*ast.ArrayLiteral)
	array.Elements[0] = integerNode(2)
	assertNodeString(t, original, before)

	hash := call.Arguments[1].(*ast.HashMapLiteral)
	for key := range hash.Pairs {
		hash.Pairs[key] = integerNode(2)
	}
	assertNodeString(t, original, before)
	call.Arguments[0] = integerNode(3)
	assertNodeString(t, original, before)
	function.Body.Statements[0] = &ast.ReturnStatement{ReturnValue: integerNode(4)}
	assertNodeString(t, original, before)
	rewritten.Statements[0] = &ast.ExpressionStatement{Expression: integerNode(5)}
	assertNodeString(t, original, before)
}

func TestModifyCanReuseMacroDefinition(t *testing.T) {
	definition := parseModifyProgram(t, "macro(x) { 1 + x; }")
	before := definition.String()
	first := ast.Modify(definition, replaceOneWith(2))
	second := ast.Modify(definition, replaceOneWith(3))
	assertNodeString(t, first, "macro(x) {(2 + x)}")
	assertNodeString(t, second, "macro(x) {(3 + x)}")
	assertNodeString(t, definition, before)
}

func TestModifyPreservesMissingChildren(t *testing.T) {
	nodes := []ast.Node{
		&ast.Program{Statements: []ast.Statement{nil}},
		&ast.ExpressionStatement{},
		&ast.LetStatement{},
		&ast.ReturnStatement{},
		&ast.ForInStatement{},
		&ast.BlockStatement{Statements: []ast.Statement{nil}},
		&ast.PrefixExpression{},
		&ast.InfixExpression{},
		&ast.IfExpression{},
		&ast.FunctionLiteral{},
		&ast.MacroLiteral{},
		&ast.CallExpression{Arguments: []ast.Expression{nil}},
		&ast.ArrayLiteral{Elements: []ast.Expression{nil}},
		&ast.IndexExpression{},
		&ast.MacroExpansion{},
		&ast.HashMapLiteral{Pairs: map[ast.Expression]ast.Expression{}},
	}
	for _, node := range nodes {
		t.Run(fmt.Sprintf("%T", node), func(t *testing.T) {
			visits := 0
			rewritten := ast.Modify(node, func(current ast.Node) ast.Node {
				visits++
				return current
			})
			if !reflect.DeepEqual(rewritten, node) {
				t.Fatalf("missing children changed: %#v", rewritten)
			}
			if visits != 1 {
				t.Fatalf("modifier called %d times, want only the parent", visits)
			}
		})
	}
}

func TestModifyNilDoesNotCallModifier(t *testing.T) {
	rewritten := ast.Modify(nil, func(node ast.Node) ast.Node {
		t.Fatal("modifier called for a nil node")
		return node
	})
	if rewritten != nil {
		t.Fatalf("Modify(nil) = %v, want nil", rewritten)
	}
}

func TestMacroLiteralString(t *testing.T) {
	cases := []struct {
		source string
		want   string
	}{
		{"macro() {}", "macro() {}"},
		{"macro(x, y) { x + y; }", "macro(x, y) {(x + y)}"},
	}
	for _, tc := range cases {
		program := parseModifyProgram(t, tc.source)
		macro := program.Statements[0].(*ast.ExpressionStatement).Expression.(*ast.MacroLiteral)
		if macro.TokenLiteral() != "macro" {
			t.Fatalf("macro token = %q", macro.TokenLiteral())
		}
		assertNodeString(t, macro, tc.want)
	}
}

func parseModifyProgram(t *testing.T, source string) *ast.Program {
	t.Helper()
	p := parser.New(lexer.New(source))
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors for %q: %v", source, p.Errors())
	}
	return program
}

func replaceOneWith(value int64) ast.Modifier {
	return func(node ast.Node) ast.Node {
		integer, ok := node.(*ast.IntegerLiteral)
		if !ok || integer.Value != 1 {
			return node
		}
		return integerNode(value)
	}
}

func integerNode(value int64) *ast.IntegerLiteral {
	return &ast.IntegerLiteral{
		Token: token.Token{Type: token.INT, Literal: strconv.FormatInt(value, 10)},
		Value: value,
	}
}

func assertNodeString(t *testing.T, node ast.Node, want string) {
	t.Helper()
	if got := node.String(); got != want {
		t.Fatalf("node = %q, want %q", got, want)
	}
}
