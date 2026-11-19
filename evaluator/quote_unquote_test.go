package evaluator

import (
	"monkey-go/object"
	"testing"
)

func TestQuoteErrors(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{`quote();`, "quote expects 1 argument, got 0"},
		{`quote(1, 2);`, "quote expects 1 argument, got 2"},
		{`quote(unquote());`, "unquote expects 1 argument, got 0"},
		{`quote(unquote(1, 2));`, "unquote expects 1 argument, got 2"},
		{`quote(unquote(missing));`, "identifier not found: missing"},
		{`quote(unquote(noValue()));`, "unquote evaluated to no value"},
		{`quote(unquote([]));`, "cannot unquote ARRAY as an expression"},
		{`quote(unquote(fn() { 1; }));`, "cannot unquote FUNCTION as an expression"},
		{`quote(unquote(if (false) { 1; }));`, "cannot unquote NULL as an expression"},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			env := object.NewEnvironment()
			env.Set("noValue", &object.Builtin{Fn: func(args ...object.Object) object.Object { return nil }})
			got := Eval(parseMacroProgram(t, tc.input), env)
			err, ok := got.(*object.Error)
			if !ok {
				t.Fatalf("expected error %q, got %T (%v)", tc.want, got, got)
			}
			if err.Message != tc.want {
				t.Fatalf("error = %q, want %q", err.Message, tc.want)
			}
		})
	}
}

func TestUnquotedLiteralsRemainEvaluable(t *testing.T) {
	cases := []struct {
		input string
		want  object.Object
	}{
		{`quote(unquote(-123));`, &object.Integer{Value: -123}},
		{`quote(unquote(true));`, TRUE},
		{`quote(unquote(false));`, FALSE},
		{`quote(unquote("猿"));`, &object.String{Value: "猿"}},
		{`quote(unquote(quote(1 + 2)));`, &object.Integer{Value: 3}},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			value := Eval(parseMacroProgram(t, tc.input), object.NewEnvironment())
			quoted, ok := value.(*object.Quote)
			if !ok {
				t.Fatalf("expected quoted syntax, got %T (%v)", value, value)
			}
			got := Eval(quoted.Node, object.NewEnvironment())
			if got.Type() != tc.want.Type() || got.Inspect() != tc.want.Inspect() {
				t.Fatalf("evaluated quote = %T (%v), want %T (%v)", got, got, tc.want, tc.want)
			}
		})
	}
}

func TestQuoteStopsUnquotingAfterFirstError(t *testing.T) {
	calls := 0
	env := object.NewEnvironment()
	env.Set("tick", &object.Builtin{Fn: func(args ...object.Object) object.Object {
		calls++
		return &object.Integer{Value: 1}
	}})
	input := `quote([unquote(missing), unquote(tick())]);`
	value := Eval(parseMacroProgram(t, input), env)
	err, ok := value.(*object.Error)
	if !ok || err.Message != "identifier not found: missing" {
		t.Fatalf("first unquote error was lost: %v", value)
	}
	if calls != 0 {
		t.Fatalf("evaluated %d unquotes after the first error", calls)
	}
}

func TestQuoteDoesNotEvaluateOrdinaryCalls(t *testing.T) {
	input := `quote(fn() { missing; }());`
	value := Eval(parseMacroProgram(t, input), object.NewEnvironment())
	quoted, ok := value.(*object.Quote)
	if !ok {
		t.Fatalf("expected quoted call, got %T (%v)", value, value)
	}
	got := Eval(quoted.Node, object.NewEnvironment())
	err, ok := got.(*object.Error)
	if !ok || err.Message != "identifier not found: missing" {
		t.Fatalf("ordinary function call syntax was changed: %v", got)
	}
}
