package ast_test

import (
	"fmt"
	"monkey-go/ast"
	"reflect"
	"testing"
)

func TestRewriteVisitsParentsBeforeChildren(t *testing.T) {
	original := parseModifyProgram(t, "1 + 1")
	var visits []string
	rewritten := ast.Rewrite(original, func(node ast.Node) (ast.Node, bool) {
		visits = append(visits, fmt.Sprintf("%T: %s", node, node.String()))
		return replaceOneWith(2)(node), true
	})
	want := []string{
		"*ast.Program: (1 + 1)",
		"*ast.ExpressionStatement: (1 + 1)",
		"*ast.InfixExpression: (1 + 1)",
		"*ast.IntegerLiteral: 1",
		"*ast.IntegerLiteral: 1",
	}
	if !reflect.DeepEqual(visits, want) {
		t.Fatalf("visits = %v, want %v", visits, want)
	}
	assertNodeString(t, rewritten, "(2 + 2)")
	assertNodeString(t, original, "(1 + 1)")
}

func TestRewriteDoesNotVisitSkippedReplacement(t *testing.T) {
	original := parseModifyProgram(t, "argument + 1")
	argument := parseModifyProgram(t, "call(argument, 1)")
	replacement := argument.Statements[0].(*ast.ExpressionStatement).Expression
	var identifiers []string
	rewritten := ast.Rewrite(original, func(node ast.Node) (ast.Node, bool) {
		if identifier, ok := node.(*ast.Identifier); ok {
			identifiers = append(identifiers, identifier.Value)
			return replacement, false
		}
		return replaceOneWith(2)(node), true
	})
	assertNodeString(t, rewritten, "(call(argument, 1) + 2)")
	assertNodeString(t, original, "(argument + 1)")
	if !reflect.DeepEqual(identifiers, []string{"argument"}) {
		t.Fatalf("replacement identifiers were visited: %v", identifiers)
	}
}

func TestRewriteRetainsSourceLocations(t *testing.T) {
	original := parseModifyProgram(t, "fn(x) { x + 1; }")
	rewritten := ast.Rewrite(original, func(node ast.Node) (ast.Node, bool) {
		return node, true
	})
	if !reflect.DeepEqual(original, rewritten) {
		t.Fatal("rewriting changed AST values or source locations")
	}
	if rewritten == original {
		t.Fatal("rewriting returned the original program container")
	}
}

func TestRewriteNilDoesNotCallRewriter(t *testing.T) {
	rewritten := ast.Rewrite(nil, func(node ast.Node) (ast.Node, bool) {
		t.Fatal("rewriter called for a nil node")
		return node, true
	})
	if rewritten != nil {
		t.Fatalf("Rewrite(nil) = %v, want nil", rewritten)
	}
}

func TestRewriteMacroExpansionSkipsDiagnosticMetadata(t *testing.T) {
	body := parseMacroDefinition(t, "macros.mk", "macro(x) { body(x); }").Body
	callProgram := parseModifyProgram(t, "original(argument)")
	call := callProgram.Statements[0].(*ast.ExpressionStatement).Expression.(*ast.CallExpression)
	expansion := &ast.MacroExpansion{Body: body, Call: call}
	var identifiers []string
	rewritten := ast.Rewrite(expansion, func(node ast.Node) (ast.Node, bool) {
		if identifier, ok := node.(*ast.Identifier); ok {
			identifiers = append(identifiers, identifier.Value)
		}
		return node, true
	}).(*ast.MacroExpansion)
	if !reflect.DeepEqual(identifiers, []string{"body", "x"}) {
		t.Fatalf("visited identifiers = %v, diagnostic call must stay unvisited", identifiers)
	}
	assertNodeString(t, rewritten, "body(x)")
	assertNodeString(t, rewritten.Call, "original(argument)")
}

func TestMacroExpansionString(t *testing.T) {
	cases := []struct {
		source string
		want   string
	}{
		{"macro() {}", "{}"},
		{"macro(x) { x + 1; }", "(x + 1)"},
		{"macro(x) { let value = x; value; }", "{let value = x;value}"},
		{"macro(x) { return x; }", "{return x;}"},
	}
	for _, tc := range cases {
		expansion := &ast.MacroExpansion{Body: parseMacroDefinition(t, "macros.mk", tc.source).Body}
		assertNodeString(t, expansion, tc.want)
		if expansion.TokenLiteral() != "macro" {
			t.Fatalf("expansion token = %q", expansion.TokenLiteral())
		}
	}
	assertNodeString(t, &ast.MacroExpansion{}, "{}")
}
