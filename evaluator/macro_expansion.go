package evaluator

import (
	"fmt"
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
			Env:        env,
		})
	}
	program.Statements = remaining
}

// ExpandMacros evaluates macro calls against quoted arguments before runtime.
// The pass limit prevents unbounded self-expanding macros.
func ExpandMacros(program ast.Node, env *object.Environment) (ast.Node, error) {
	const maxPasses = 32
	for pass := 0; pass < maxPasses; pass++ {
		result, changed, failure := expandMacroPass(program, env)
		if failure != nil {
			return nil, failure
		}
		program = result
		if !changed {
			return program, nil
		}
	}
	return nil, fmt.Errorf("macro expansion exceeded %d passes", maxPasses)
}

func expandMacroPass(program ast.Node, env *object.Environment) (ast.Node, bool, error) {
	changed := false
	var failure error
	result := ast.Modify(program, func(current ast.Node) ast.Node {
		if failure != nil {
			return current
		}
		call, ok := current.(*ast.CallExpression)
		if !ok {
			return current
		}
		name, macro := lookupMacro(call, env)
		if macro == nil {
			return current
		}
		replacement, err := expandMacroCall(name, call, macro)
		if err != nil {
			failure = err
			return current
		}
		changed = true
		return replacement
	})
	return result, changed, failure
}

func lookupMacro(call *ast.CallExpression, env *object.Environment) (string, *object.Macro) {
	identifier, ok := call.Function.(*ast.Identifier)
	if !ok {
		return "", nil
	}
	value, _ := env.Get(identifier.Value)
	macro, _ := value.(*object.Macro)
	return identifier.Value, macro
}

func expandMacroCall(name string, call *ast.CallExpression, macro *object.Macro) (ast.Expression, error) {
	if len(call.Arguments) != len(macro.Parameters) {
		return nil, fmt.Errorf("macro %s expects %d arguments, got %d", name, len(macro.Parameters), len(call.Arguments))
	}
	local := object.NewEnclosedEnvironment(macro.Env)
	for i, parameter := range macro.Parameters {
		local.Set(parameter.Value, &object.Quote{Node: call.Arguments[i]})
	}
	evaluated := unwrapReturnValue(Eval(macro.Body, local))
	if errObj, ok := evaluated.(*object.Error); ok {
		return nil, fmt.Errorf("macro %s: %s", name, errObj.Message)
	}
	quoted, ok := evaluated.(*object.Quote)
	if !ok {
		return nil, fmt.Errorf("macro %s must return quote(...)", name)
	}
	expression, ok := quoted.Node.(ast.Expression)
	if !ok {
		return nil, fmt.Errorf("macro %s returned a non-expression AST node", name)
	}
	return expression, nil
}
