package evaluator

import (
	"monkey-go/ast"
	"monkey-go/lexer"
	"monkey-go/object"
	"monkey-go/parser"
	"strings"
	"testing"
)

func parseMacroTest(t *testing.T, source string) *parser.Parser {
	t.Helper()
	return parser.New(lexer.New(source))
}

func parseMacroProgram(t *testing.T, source string) *ast.Program {
	t.Helper()
	p := parseMacroTest(t, source)
	program := p.ParseProgram()
	if errs := p.Errors(); len(errs) != 0 {
		t.Fatalf("parser errors: %v", errs)
	}
	return program
}

func expandMacroProgram(t *testing.T, source string, env *object.Environment) ast.Node {
	t.Helper()
	program := parseMacroProgram(t, source)
	DefineMacros(program, env)
	expanded, err := ExpandMacros(program, env)
	if err != nil {
		t.Fatal(err)
	}
	return expanded
}

func TestMacroExpansionAndDefinitionReuse(t *testing.T) {
	env := object.NewEnvironment()
	definition := parseMacroProgram(t, `let twice = macro(x) { x + x; };`)
	DefineMacros(definition, env)
	if len(definition.Statements) != 0 {
		t.Fatal("macro definition was left in the runtime program")
	}
	for _, value := range []struct {
		source string
		want   int64
	}{{"twice(1)", 2}, {"twice(7)", 14}} {
		expanded := expandMacroProgram(t, value.source, env)
		testIntegerObject(t, Eval(expanded, object.NewEnvironment()), value.want)
	}
	macro, _ := env.Get("twice")
	if macro.(*object.Macro).Body.String() != "(x + x)" {
		t.Fatal("expansion modified the stored definition")
	}
}

func TestNestedMacroExpansion(t *testing.T) {
	input := `let inc = macro(x) { x + 1; };
let twice = macro(x) { x + x; };
twice(inc(3));`
	expanded := expandMacroProgram(t, input, object.NewEnvironment())
	testIntegerObject(t, Eval(expanded, object.NewEnvironment()), 8)
}

func TestMacroErrors(t *testing.T) {
	cases := []struct {
		source string
		want   string
	}{
		{`let twice = macro(x) { x + x; }; twice();`, "expects 1 arguments, got 0"},
		{`let identity = macro(x) { x; }; identity(1, 2);`, "expects 1 arguments, got 2"},
		{`let recursive = macro(x) { recursive(x); }; recursive(1);`, "exceeded 32 nested expansions"},
		{`let a = macro() { b(); }; let b = macro() { a(); }; a();`, "exceeded 32 nested expansions"},
	}
	for _, tc := range cases {
		t.Run(tc.source, func(t *testing.T) {
			program := parseMacroProgram(t, tc.source)
			env := object.NewEnvironment()
			DefineMacros(program, env)
			_, err := ExpandMacros(program, env)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want substring %q", err, tc.want)
			}
		})
	}
}

func TestGeneratedMacroCallsAndExplicitReturn(t *testing.T) {
	input := `let inc = macro(x) { return x + 1; };
let generated = macro(x) { inc(x); };
generated(4); 99;`
	expanded := expandMacroProgram(t, input, object.NewEnvironment())
	testIntegerObject(t, Eval(expanded, object.NewEnvironment()), 5)
}

func TestMacroArgumentsAreSyntax(t *testing.T) {
	cases := []string{
		`let ignore = macro(x) { 42; }; ignore(missing);`,
		`let required = macro(x) { x; }; let ignore = macro(x) { 42; }; ignore(required());`,
		`let recursive = macro() { recursive(); }; let ignore = macro(x) { 42; }; ignore(recursive());`,
	}
	for _, source := range cases {
		t.Run(source, func(t *testing.T) {
			expanded := expandMacroProgram(t, source, object.NewEnvironment())
			testIntegerObject(t, Eval(expanded, object.NewEnvironment()), 42)
		})
	}
}

func TestMacroEnvironmentRemainsSeparateFromRuntime(t *testing.T) {
	env := object.NewEnvironment()
	env.Set("offset", &object.Integer{Value: 3})
	input := `let add = macro(x) { x + offset; };
let offset = 100;
add(7);`
	expanded := expandMacroProgram(t, input, env)
	testIntegerObject(t, Eval(expanded, object.NewEnvironment()), 107)
}

func TestExpansionPreservesRuntimeCalls(t *testing.T) {
	cases := []struct {
		input string
		want  int64
	}{
		{`len([1, 2]);`, 2},
		{`fn(x) { x + 1; }(2);`, 3},
		{`let identity = fn(x) { x; }; identity(7);`, 7},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			env := object.NewEnvironment()
			env.Set("identity", &object.Integer{Value: 99})
			expanded := expandMacroProgram(t, tc.input, env)
			if expanded.String() != parseMacroProgram(t, tc.input).String() {
				t.Fatal("expansion changed a runtime call")
			}
			testIntegerObject(t, Eval(expanded, object.NewEnvironment()), tc.want)
		})
	}
}

func TestExpansionNeverExecutesTemplateBody(t *testing.T) {
	calls := 0
	env := object.NewEnvironment()
	env.Set("tick", &object.Builtin{Fn: func(args ...object.Object) object.Object {
		calls++
		return &object.Integer{Value: 1}
	}})
	expanded := expandMacroProgram(t, `let effect = macro() { tick(); 7; }; effect();`, env)
	if calls != 0 {
		t.Fatal("expansion executed the macro body")
	}
	testIntegerObject(t, Eval(expanded, env), 7)
	if calls != 1 {
		t.Fatalf("runtime called tick %d times, want 1", calls)
	}
}

func TestMacroExpansionDepthBoundary(t *testing.T) {
	for _, depth := range []int{32, 33} {
		input := `let identity = macro(x) { x; }; `
		input += strings.Repeat("identity(", depth) + "7" + strings.Repeat(")", depth)
		program := parseMacroProgram(t, input)
		env := object.NewEnvironment()
		DefineMacros(program, env)
		expanded, err := ExpandMacros(program, env)
		if depth == 33 {
			if err == nil {
				t.Fatal("accepted 33 nested expansions")
			}
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		testIntegerObject(t, Eval(expanded, object.NewEnvironment()), 7)
	}
}

func TestIndependentMacrosDoNotShareExpansionDepth(t *testing.T) {
	input := `let identity = macro(x) { x; }; ` + strings.Repeat("identity(7);", 64)
	expanded := expandMacroProgram(t, input, object.NewEnvironment())
	testIntegerObject(t, Eval(expanded, object.NewEnvironment()), 7)
}
