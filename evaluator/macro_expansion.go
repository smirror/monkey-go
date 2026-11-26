package evaluator

import (
	"monkey-go/ast"
	"monkey-go/object"
)

// DefineMacros registers top-level macro declarations and removes them from
// the runtime program. Definitions persist when the same environment is reused.
func DefineMacros(program *ast.Program, env *object.Environment) {
	remaining := make([]ast.Statement, 0, len(program.Statements))
	for _, statement := range program.Statements {
		declaration, ok := statement.(*ast.LetStatement)
		if !ok || declaration.Value == nil {
			remaining = append(remaining, statement)
			continue
		}
		literal, ok := declaration.Value.(*ast.MacroLiteral)
		if !ok {
			remaining = append(remaining, statement)
			continue
		}
		env.Set(declaration.Name.Value, &object.Macro{
			Parameters: literal.Parameters,
			Body:       literal.Body,
		})
	}
	program.Statements = remaining
}

// ExpandMacros replaces macro calls with syntax templates before runtime.
// Templates expand from the outside in, so unused arguments are never visited.
// A depth limit prevents unbounded self-expanding macros.
func ExpandMacros(program ast.Node, env *object.Environment) (ast.Node, error) {
	expander := &macroExpander{env: env}
	result := expander.expand(program, make(macroScope), 0)
	if expander.failure != nil {
		return nil, expander.failure
	}
	return result, nil
}

type macroExpander struct {
	env     *object.Environment
	failure error
}

func (e *macroExpander) expand(program ast.Node, scope macroScope, depth int) ast.Node {
	walker := macroSyntaxWalker{
		scope: scope,
		rewrite: func(current ast.Node, scope macroScope) (ast.Node, bool) {
			if e.failure != nil {
				return current, false
			}
			call, ok := current.(*ast.CallExpression)
			if !ok {
				return current, true
			}
			name, macro := lookupMacro(call, e.env, scope)
			if macro == nil {
				return current, true
			}
			return e.expandCall(name, call, macro, scope, depth), false
		},
	}
	return walker.walk(program)
}

func (e *macroExpander) expandCall(name string, call *ast.CallExpression, macro *object.Macro, scope macroScope, depth int) ast.Node {
	const maxDepth = 32
	if depth >= maxDepth {
		e.failure = newMacroExpansionError(call.Span(), "macro expansion exceeded %d nested expansions", maxDepth)
		return call
	}
	replacement, err := expandMacroCall(name, call, macro)
	if err != nil {
		e.failure = err
		return call
	}
	// Process generated bindings before moving to the next caller statement.
	return e.expand(replacement, scope, depth+1)
}

func lookupMacro(call *ast.CallExpression, env *object.Environment, scope macroScope) (string, *object.Macro) {
	identifier, ok := call.Function.(*ast.Identifier)
	if !ok || scope[identifier.Value] {
		return "", nil
	}
	value, _ := env.Get(identifier.Value)
	macro, _ := value.(*object.Macro)
	return identifier.Value, macro
}

func expandMacroCall(name string, call *ast.CallExpression, macro *object.Macro) (ast.Expression, error) {
	if len(call.Arguments) != len(macro.Parameters) {
		return nil, newMacroExpansionError(call.Span(), "macro %s expects %d arguments, got %d", name, len(macro.Parameters), len(call.Arguments))
	}
	expanded := &ast.MacroExpansion{
		Body: substituteMacroBody(macro, call.Arguments),
		Call: cloneMacroSyntax(call).(*ast.CallExpression),
	}
	ast.SetSpan(expanded, call.Span())
	return expanded, nil
}
