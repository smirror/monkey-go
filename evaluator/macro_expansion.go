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
			Body: literal.Body,
			Env: env,
		})
	}
	program.Statements = remaining
}

// ExpandMacros evaluates macro calls against quoted arguments before runtime.
// The pass limit prevents unbounded self-expanding macros.
func ExpandMacros(program ast.Node, env *object.Environment) (ast.Node, error) {
	const maxPasses = 32
	for pass := 0; pass < maxPasses; pass++ {
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
			ident, ok := call.Function.(*ast.Identifier)
			if !ok {
				return current
			}
			value, found := env.Get(ident.Value)
			if !found {
				return current
			}
			macro, ok := value.(*object.Macro)
			if !ok {
				return current
			}
			if len(call.Arguments) != len(macro.Parameters) {
				failure = fmt.Errorf("macro %s expects %d arguments, got %d", ident.Value, len(macro.Parameters), len(call.Arguments))
				return current
			}
			local := object.NewEnclosedEnvironment(macro.Env)
			for i, parameter := range macro.Parameters {
				local.Set(parameter.Value, &object.Quote{Node: call.Arguments[i]})
			}
			evaluated := Eval(macro.Body, local)
			if wrapped, ok := evaluated.(*object.ReturnValue); ok {
				evaluated = wrapped.Value
			}
			if errObj, ok := evaluated.(*object.Error); ok {
				failure = fmt.Errorf("macro %s: %s", ident.Value, errObj.Message)
				return current
			}
			quoted, ok := evaluated.(*object.Quote)
			if !ok {
				failure = fmt.Errorf("macro %s must return quote(...)", ident.Value)
				return current
			}
			if _, ok := quoted.Node.(ast.Expression); !ok {
				failure = fmt.Errorf("macro %s returned a non-expression AST node", ident.Value)
				return current
			}
			changed = true
			return quoted.Node
		})
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
